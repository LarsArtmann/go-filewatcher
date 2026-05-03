//nolint:varnamelen // Idiomatic short names: d (DirEntry), op (operation)
package filewatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
		err := w.fswatcher.Add(root.Get())
		if err != nil {
			return fmt.Errorf("adding watch path %q: %w", root, err)
		}

		w.watchList = append(w.watchList, root.Get())

		return nil
	}

	return w.walkAndAddPaths(root)
}

// walkAndAddPaths walks a directory tree and adds all directories to the watcher.
// Caller must hold w.mu lock.
func (w *Watcher) walkAndAddPaths(root RootPath) error {
	err := filepath.WalkDir(root.Get(), w.walkDirFunc)
	if err != nil {
		return fmt.Errorf("walking directory %q: %w", root, err)
	}
	// Track the root path (caller holds lock)
	w.watchList = append(w.watchList, root.Get())

	return nil
}

// walkDirFunc is the WalkDirFunc for adding paths during directory traversal.
func (w *Watcher) walkDirFunc(path string, d os.DirEntry, walkErr error) error {
	if walkErr != nil {
		isDir := d != nil && d.IsDir()

		return fmt.Errorf("walking directory entry %q (isDir=%v): %w", path, isDir, walkErr)
	}

	if !d.IsDir() {
		return nil
	}

	if w.shouldSkipDir(d.Name()) {
		return filepath.SkipDir
	}

	addErr := w.fswatcher.Add(path)
	if addErr != nil {
		return fmt.Errorf("watching path %q: %w", path, addErr)
	}

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
