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
func (w *Watcher) addPath(root string) error {
	if !w.recursive {
		if err := w.fswatcher.Add(root); err != nil {
			return fmt.Errorf("adding watch path %q: %w", root, err)
		}
		return nil
	}

	return w.walkAndAddPaths(root)
}

// walkAndAddPaths walks a directory tree and adds all directories to the watcher.
func (w *Watcher) walkAndAddPaths(root string) error {
	if err := filepath.WalkDir(root, w.walkDirFunc); err != nil {
		return fmt.Errorf("walking directory %q: %w", root, err)
	}
	// Track the root path
	w.watchList = append(w.watchList, root)
	return nil
}

// walkDirFunc is the WalkDirFunc for adding paths during directory traversal.
func (w *Watcher) walkDirFunc(path string, d os.DirEntry, err error) error {
	if err != nil {
		return fmt.Errorf("walking path %q: %w", path, err)
	}

	if !d.IsDir() {
		return nil
	}

	if w.shouldSkipDir(d.Name()) {
		return filepath.SkipDir
	}

	if addErr := w.fswatcher.Add(path); addErr != nil {
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
