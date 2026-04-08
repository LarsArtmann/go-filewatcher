// Package filewatcher provides a high-level, composable file system watcher
// built on top of fsnotify. It eliminates the boilerplate of raw fsnotify
// usage by providing sensible defaults for common patterns:
//   - Automatic recursive directory watching
//   - Configurable debounce (global or per-path)
//   - Composable event filters (extension, directory, glob)
//   - Middleware chains for cross-cutting concerns
//   - Graceful shutdown via context cancellation
//
// Design principles:
//   - Functional options for configuration
//   - Sentinel errors with fmt.Errorf wrapping
//   - No panics, explicit error handling
//   - Context as first parameter
//   - Channel-based event streaming
package filewatcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultEventBufferSize = 64 // Default capacity for the event channel buffer

// DefaultIgnoreDirs contains commonly ignored directory names.
//
//nolint:gochecknoglobals // Exported for user reference in configuration.
var DefaultIgnoreDirs = []string{
	".git", ".hg", ".svn",
	"vendor", "node_modules",
	"dist", "build", "bin", "out",
	"__pycache__", ".cache",
}

// Watcher watches file system paths for changes and emits filtered,
// debounced events through a channel.
type Watcher struct {
	fswatcher *fsnotify.Watcher

	// Configuration
	paths           []string
	filters         []Filter
	middleware      []Middleware
	recursive       bool
	globalDebounce  time.Duration
	perPathDebounce time.Duration
	errorHandler    func(error)
	skipDotDirs     bool
	bufferSize      int
	onAdd           func(path string) // callback when a path is added
	ignoreDirNames  []string          // user-configured dir names to skip during walk

	// Internal state
	mu        sync.RWMutex
	closed    bool
	watching  bool
	watchList []string // tracked paths currently being watched

	// Debouncer (initialized based on config)
	debounceInterface DebouncerInterface
}

// Compile-time interface check: Watcher implements io.Closer.
var _ io.Closer = (*Watcher)(nil)

// DebouncerInterface is the interface for debouncer implementations.
type DebouncerInterface interface {
	Debounce(key string, fn func())
	Stop()
}

// Compile-time interface checks.
var (
	_ DebouncerInterface = (*Debouncer)(nil)
	_ DebouncerInterface = (*GlobalDebouncer)(nil)
)

// New creates a new Watcher for the given paths with the specified options.
// At least one path must be provided. Paths are validated to exist.
//
// The watcher is not started until Watch() is called.
func New(paths []string, opts ...Option) (*Watcher, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("%w: no paths provided", ErrNoPaths)
	}

	// Validate all paths exist
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("resolving path %q: %w", p, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("%w: path %q (resolved: %q)", ErrPathNotFound, p, abs)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%w: path %q", ErrPathNotDir, p)
		}
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating fsnotify watcher: %w", err)
	}

	w := &Watcher{
		fswatcher:         fswatcher,
		paths:             paths,
		recursive:         true,
		filters:           nil,
		middleware:        nil,
		globalDebounce:    0,
		perPathDebounce:   0,
		errorHandler:      nil,
		skipDotDirs:       true,
		bufferSize:        defaultEventBufferSize,
		onAdd:             nil,
		ignoreDirNames:    nil,
		mu:                sync.RWMutex{},
		closed:            false,
		watching:          false,
		watchList:         make([]string, 0, len(paths)),
		debounceInterface: nil,
	}

	for _, opt := range opts {
		opt(w)
	}

	// Initialize debouncer based on configuration
	w.initDebouncer()

	return w, nil
}

// Watch starts watching the configured paths and returns a read-only channel
// of filtered, debounced events. The channel is closed when the context is
// cancelled or Close() is called.
//
// Callers should range over the returned channel to process events:
//
//	events, err := watcher.Watch(ctx)
//	for event := range events {
//	    handleEvent(event)
//	}
func (w *Watcher) Watch(ctx context.Context) (<-chan Event, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil, fmt.Errorf("%w: cannot watch", ErrWatcherClosed)
	}

	if w.watching {
		return nil, fmt.Errorf("%w: already running", ErrWatcherRunning)
	}

	// Add initial paths to the fsnotify watcher
	for _, p := range w.paths {
		pathErr := w.addPath(p)
		if pathErr != nil {
			return nil, fmt.Errorf("adding watch path %q: %w", p, pathErr)
		}
	}

	eventCh := make(chan Event, w.bufferSize)

	w.watching = true
	go w.watchLoop(ctx, eventCh)

	return eventCh, nil
}

// Add adds a new path to the watcher. The path must be an existing directory.
func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("%w: cannot add", ErrWatcherClosed)
	}

	abs, addErr := filepath.Abs(path)
	if addErr != nil {
		return fmt.Errorf("resolving path %q: %w", path, addErr)
	}

	if err := w.addPath(abs); err != nil {
		return err
	}
	w.watchList = append(w.watchList, abs)
	return nil
}

// Remove removes a path from the watcher. The watcher stops monitoring
// this path and all its subdirectories (if recursive).
func (w *Watcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("%w: cannot remove", ErrWatcherClosed)
	}

	abs, removeErr := filepath.Abs(path)
	if removeErr != nil {
		return fmt.Errorf("resolving path %q: %w", path, removeErr)
	}

	if err := w.fswatcher.Remove(abs); err != nil {
		return fmt.Errorf("removing watch path %q: %w", abs, err)
	}

	// Remove from watchList
	for i, p := range w.watchList {
		if p == abs {
			w.watchList = append(w.watchList[:i], w.watchList[i+1:]...)
			break
		}
	}

	return nil
}

// WatchList returns a copy of the list of paths currently being watched.
func (w *Watcher) WatchList() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]string, len(w.watchList))
	copy(result, w.watchList)
	return result
}

// Stats provides observability metrics for the watcher.
type Stats struct {
	WatchCount int
	IsWatching bool
	IsClosed   bool
}

// Stats returns current statistics about the watcher.
func (w *Watcher) Stats() Stats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return Stats{
		WatchCount: len(w.watchList),
		IsWatching: w.watching,
		IsClosed:   w.closed,
	}
}

// Close stops the watcher and releases all resources.
// It is safe to call Close multiple times.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	w.watching = false
	w.watchList = w.watchList[:0]

	if w.debounceInterface != nil {
		w.debounceInterface.Stop()
	}

	if err := w.fswatcher.Close(); err != nil {
		return fmt.Errorf("closing fsnotify watcher: %w", err)
	}
	return nil
}
