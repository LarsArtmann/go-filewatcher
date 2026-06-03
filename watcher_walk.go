//nolint:varnamelen // Idiomatic short names: d (DirEntry), op (operation)
package filewatcher

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

// initDebouncer sets up the appropriate debouncer based on configuration.
func (w *Watcher) initDebouncer() {
	switch {
	case w.perPathDebounce > 0:
		w.debounceInterface = NewDebouncer(w.perPathDebounce)
	case w.globalDebounce > 0:
		w.debounceInterface = NewGlobalDebouncer(w.globalDebounce)
	}
}

// addPath adds a directory (and optionally its subdirectories) to the fsnotify watcher.
// It also appends the root path to the watchList.
func (w *Watcher) addPath(root RootPath) error {
	if !w.recursive {
		// Budget check before adding
		if w.maxWatches > 0 && len(w.watchList) >= w.maxWatches {
			return nil
		}

		err := w.fswatcher.Add(root.Get())
		if err != nil {
			w.watchErrors.Add(1)
			w.handleError(ErrorContext{
				Operation: opAddPath,
				Path:      root.Get(),
				Event:     nil,
				Retryable: true,
			}, fmt.Errorf("adding watch path %q: %w", root, err))

			return nil
		}

		w.watchList = append(w.watchList, root.Get())

		if w.onAdd != nil {
			w.onAdd(root.Get())
		}

		return nil
	}

	return w.walkAndAddPaths(root)
}

// walkAndAddPaths walks a directory tree and adds all directories to the watcher.
// Directories are collected during walking and added in batches to yield to
// event processing between batches. Caller must hold w.mu lock.
func (w *Watcher) walkAndAddPaths(root RootPath) error {
	w.walkBatch = make([]string, 0, watchBatchSize)

	err := filepath.WalkDir(root.Get(), w.walkDirFunc)

	// Flush remaining batch
	if len(w.walkBatch) > 0 {
		w.addBatch(w.walkBatch)
	}

	w.walkBatch = nil

	if err != nil {
		return fmt.Errorf("walking directory %q: %w", root, err)
	}

	// Track the root path only if it wasn't already added via addBatch.
	// filepath.WalkDir visits the root first, so it's already in watchList.
	if len(w.watchList) == 0 || w.watchList[len(w.watchList)-1] != root.Get() {
		w.watchList = append(w.watchList, root.Get())
	}

	return nil
}

// walkDirFunc is the WalkDirFunc for adding paths during directory traversal.
// When walkBatch is set, it collects paths into the batch for batched registration.
// When walkBatch is nil, it adds paths immediately (used by tests).
//
//nolint:cyclop,funlen // walk logic with multiple skip conditions
func (w *Watcher) walkDirFunc(path string, d os.DirEntry, walkErr error) error {
	if walkErr != nil {
		isDir := d != nil && d.IsDir()

		return fmt.Errorf("walking directory entry %q (isDir=%v): %w", path, isDir, walkErr)
	}

	if !d.IsDir() {
		return nil
	}

	if w.followSymlinks && d.Type()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("resolving symlink %q: %w", path, err)
		}

		info, err := os.Stat(resolved)
		if err != nil {
			return fmt.Errorf("stat resolved symlink target %q: %w", resolved, err)
		}

		if !info.IsDir() {
			return nil
		}

		return w.walkAndAddPaths(NewRootPath(resolved))
	}

	if w.shouldSkipDir(d.Name()) {
		return filepath.SkipDir
	}

	if w.shouldExcludePath(path) {
		return filepath.SkipDir
	}

	w.loadGitignoreForDir(path)

	if w.shouldSkipByGitignore(path) {
		return filepath.SkipDir
	}

	// Batched mode: collect path for later batched addition
	if w.walkBatch != nil {
		w.walkBatch = append(w.walkBatch, path)

		if len(w.walkBatch) >= watchBatchSize {
			w.addBatch(w.walkBatch)
			w.walkBatch = w.walkBatch[:0]
		}

		return nil
	}

	// Direct mode: add immediately (used by tests and depth-limited walking)
	addErr := w.fswatcher.Add(path)
	if addErr != nil {
		w.watchErrors.Add(1)
		w.handleError(ErrorContext{
			Operation: opAddPath,
			Path:      path,
			Event:     nil,
			Retryable: true,
		}, fmt.Errorf("watching path %q: %w", path, addErr))

		return nil
	}

	w.watchList = append(w.watchList, path)

	if w.onAdd != nil {
		w.onAdd(path)
	}

	return nil
}

// shouldSkipDir checks if a directory should be skipped based on ignore rules.
func (w *Watcher) shouldSkipDir(name string) bool {
	if w.skipDotDirs && strings.HasPrefix(name, ".") && name != "." {
		return true
	}

	if slices.Contains(DefaultIgnoreDirs, name) {
		return true
	}

	return slices.Contains(w.ignoreDirNames, name)
}

// shouldExcludePath checks if a path should be excluded based on absolute path matching.
// It matches exact paths and path prefixes (subtree exclusion).
func (w *Watcher) shouldExcludePath(path string) bool {
	if len(w.excludePaths) == 0 {
		return false
	}

	_, exact := w.excludePaths[path]
	if exact {
		return true
	}

	prefix := path + string(filepath.Separator)

	for excludedPath := range w.excludePaths {
		if strings.HasPrefix(excludedPath, prefix) {
			return false // path is a parent of an excluded path, don't skip it
		}

		if strings.HasPrefix(path, excludedPath+string(filepath.Separator)) {
			return true // path is under an excluded subtree
		}
	}

	return false
}

const watchBatchSize = 1000

// addBatch adds a batch of paths to the fsnotify watcher.
// Respects the maxWatches budget — stops adding when budget is exhausted.
func (w *Watcher) addBatch(paths []string) {
	for _, p := range paths {
		if w.maxWatches > 0 && len(w.watchList) >= w.maxWatches {
			w.debugLog("watch budget exhausted, skipping path",
				slog.String("path", p),
				slog.Int("max_watches", w.maxWatches),
				slog.Int("current_watches", len(w.watchList)),
			)

			continue
		}

		addErr := w.fswatcher.Add(p)
		if addErr != nil {
			w.watchErrors.Add(1)
			w.handleError(ErrorContext{
				Operation: opAddPath,
				Path:      p,
				Event:     nil,
				Retryable: true,
			}, fmt.Errorf("watching path %q: %w", p, addErr))

			continue
		}

		w.watchList = append(w.watchList, p)

		if w.onAdd != nil {
			w.onAdd(p)
		}
	}

	runtime.Gosched()
}

// detectMaxWatches reads the system inotify watch limit from /proc/sys/fs/inotify/max_user_watches.
// Returns 0 on non-Linux systems or if detection fails (meaning unlimited).
func detectMaxWatches() int {
	const procPath = "/proc/sys/fs/inotify/max_user_watches"

	data, err := os.ReadFile(procPath)
	if err != nil {
		return 0
	}

	n, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
	if parseErr != nil {
		return 0
	}

	return n
}
