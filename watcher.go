//nolint:varnamelen // Idiomatic short names: p (path), w (watcher)
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

// WatcherStateFlags holds state booleans as bit flags for memory efficiency.
// 4 bools (4 bytes) → 1 byte with 4 bit flags.
type WatcherStateFlags byte

const (
	flagClosed WatcherStateFlags = 1 << iota
	flagWatching
)

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
//
// Thread-safety guarantees:
//   - New() is not safe for concurrent use during creation
//   - Watch() is safe to call concurrently with Close()
//   - Add(), Remove(), WatchList(), Stats(), IsClosed() are safe for concurrent use
//   - Close() is safe to call multiple times and concurrently with other methods
//   - The event channel returned by Watch() is closed when the watcher stops
//   - All callbacks (errorHandler, onAdd) may be called concurrently
type Watcher struct {
	fswatcher *fsnotify.Watcher

	// Configuration
	paths           []string
	filters         []Filter
	middleware      []Middleware
	recursive       bool
	globalDebounce  time.Duration
	perPathDebounce time.Duration
	skipDotDirs     bool
	bufferSize      int
	onAdd           func(path string) // callback when a path is added
	ignoreDirNames  []string          // user-configured dir names to skip during walk
	errorHandler    ErrorHandler      // callback for errors during event processing

	// Internal state
	mu        sync.RWMutex
	state     WatcherStateFlags // bit flags: closed, watching
	watchList []string          // tracked paths currently being watched

	// Debouncer (initialized based on config)
	debounceInterface DebouncerInterface
}

// Compile-time interface check: Watcher implements io.Closer.
var _ io.Closer = (*Watcher)(nil)

// IsClosed reports if the watcher has been closed.
// This is safe to call concurrently with other methods.
func (w *Watcher) IsClosed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.state&flagClosed != 0
}

// DebouncerInterface is the interface for debouncer implementations.
type DebouncerInterface interface {
	Debounce(key DebounceKey, fn func())
	Stop()
	Flush()
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
		return nil, fmt.Errorf("%w: at least one path must be provided", ErrNoPaths)
	}

	// Validate all paths exist
	for _, p := range paths {
		abs, resolveErr := filepath.Abs(p)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolving path %q during validation: %w", p, resolveErr)
		}

		info, statErr := os.Stat(abs)
		if statErr != nil {
			return nil, fmt.Errorf("%w: path %q (resolved: %q)", ErrPathNotFound, p, abs)
		}

		if !info.IsDir() {
			return nil, fmt.Errorf("%w: path %q must be a directory", ErrPathNotDir, p)
		}
	}

	fswatcher, fsErr := fsnotify.NewWatcher()
	if fsErr != nil {
		return nil, fmt.Errorf("creating fsnotify watcher: %w", fsErr)
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
		state:             0,
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

	if w.state&flagClosed != 0 {
		return nil, fmt.Errorf("%w: cannot start watch on closed watcher", ErrWatcherClosed)
	}

	if w.state&flagWatching != 0 {
		return nil, fmt.Errorf("%w: watcher is already running", ErrWatcherRunning)
	}

	// Add initial paths to the fsnotify watcher
	for _, p := range w.paths {
		addErr := w.addPath(RootPath(p))
		if addErr != nil {
			return nil, fmt.Errorf("adding watch path %q during Watch(): %w", p, addErr)
		}
	}

	eventCh := make(chan Event, w.bufferSize)

	w.state |= flagWatching
	go w.watchLoop(ctx, eventCh)

	return eventCh, nil
}

// Add adds a new path to the watcher. The path must be an existing directory.
// This method is safe for concurrent use with other methods.
func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state&flagClosed != 0 {
		return fmt.Errorf("%w: cannot add path to closed watcher", ErrWatcherClosed)
	}

	abs, resolveErr := filepath.Abs(path)
	if resolveErr != nil {
		return fmt.Errorf("resolving path %q in Add(): %w", path, resolveErr)
	}

	pathErr := w.addPath(RootPath(abs))
	if pathErr != nil {
		return fmt.Errorf("adding resolved path %q to watcher: %w", abs, pathErr)
	}

	w.watchList = append(w.watchList, abs)

	return nil
}

// Remove removes a path from the watcher. The watcher stops monitoring
// this path and all its subdirectories (if recursive).
// This method is safe for concurrent use with other methods.
func (w *Watcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.state&flagClosed != 0 {
		return fmt.Errorf("%w: cannot remove path from closed watcher", ErrWatcherClosed)
	}

	abs, resolveErr := filepath.Abs(path)
	if resolveErr != nil {
		return fmt.Errorf("resolving path %q in Remove(): %w", path, resolveErr)
	}

	removeErr := w.fswatcher.Remove(abs)
	if removeErr != nil {
		return fmt.Errorf("removing watch path %q from fsnotify: %w", abs, removeErr)
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
// This method is safe for concurrent use with other methods.
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
// This method is safe for concurrent use with other methods.
func (w *Watcher) Stats() Stats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return Stats{
		WatchCount: len(w.watchList),
		IsWatching: w.state&flagWatching != 0,
		IsClosed:   w.state&flagClosed != 0,
	}
}

// Close stops the watcher and releases all resources.
// It is safe to call Close multiple times.
func (w *Watcher) Close() error {
	w.mu.Lock()

	if w.state&flagClosed != 0 {
		w.mu.Unlock()

		return nil
	}

	w.state |= flagClosed
	w.state &^= flagWatching
	w.watchList = w.watchList[:0]

	w.mu.Unlock()

	// Close fsnotify watcher FIRST - this causes watchLoop to exit
	// and ensures no new events will be processed.
	err := w.fswatcher.Close()
	if err != nil {
		return fmt.Errorf("closing fsnotify watcher: %w", err)
	}

	// Stop the debouncer AFTER closing fsnotify to wait for any
	// in-flight debounced callbacks to complete before we return.
	if w.debounceInterface != nil {
		w.debounceInterface.Stop()
	}

	return nil
}
