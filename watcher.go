//nolint:varnamelen // Idiomatic short names: p (path), w (watcher)
package filewatcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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

// DefaultIgnoreDirs returns a copy of the commonly ignored directory names.
// The returned slice is safe to modify without affecting the defaults.
//
//nolint:gochecknoglobals // Exported for user reference in configuration.
var DefaultIgnoreDirs = []string{
	".git", ".hg", ".svn",
	"vendor", "node_modules",
	"dist", "build", "bin", "out",
	"__pycache__", ".cache",
}

// DefaultIgnoreDirsCopy returns a defensive copy of DefaultIgnoreDirs
// so callers cannot mutate the global default.
func DefaultIgnoreDirsCopy() []string {
	result := make([]string, len(DefaultIgnoreDirs))
	copy(result, DefaultIgnoreDirs)

	return result
}

// Watcher watches file system paths for changes and emits filtered,
// debounced events through a channel.
//
// Thread-safety guarantees:
//   - New() is not safe for concurrent use during creation
//   - Watch() is safe to call concurrently with Close()
//   - Add(), Remove(), WatchList(), Stats(), IsClosed(), Errors() are safe for concurrent use
//   - Close() is safe to call multiple times and concurrently with other methods
//   - The event channel returned by Watch() is closed when the watcher stops
//   - The error channel returned by Errors() is closed when the watcher stops
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
	lazyIsDir       bool              // skip os.Stat calls in convertEvent for performance
	done            chan struct{}     // closed by Close() to signal shutdown to in-flight goroutines

	// Internal state
	mu        sync.RWMutex
	state     WatcherStateFlags // bit flags: closed, watching
	watchList []string          // tracked paths currently being watched
	wg        sync.WaitGroup    // tracks watchLoop goroutine for clean shutdown

	// Event channel - stored so Close() can close it after stopping debouncer
	// This prevents race between debouncer callbacks and channel close
	eventCh chan<- Event
	// closeEventChOnce ensures eventCh is closed exactly once, either by watchLoop
	// when context is cancelled, or by Close() when watcher is stopped
	closeEventChOnce sync.Once

	// Debouncer (initialized based on config)
	debounceInterface DebouncerInterface

	// Error channel - lazily initialized when Errors() is first called
	errorsMu   sync.Mutex
	errorsCh   chan error
	errorsOnce sync.Once

	// Observability metrics (atomic counters for thread-safe access)
	eventsProcessed   atomic.Uint64 // Total events that passed all filters
	eventsFilteredOut atomic.Uint64 // Events filtered out (dropped by filters)
	errorsEncountered atomic.Uint64 // Errors encountered during processing
	startTime         time.Time     // When watcher was created/started
}

// Compile-time interface check: Watcher implements io.Closer.
var _ io.Closer = (*Watcher)(nil)

// isClosed reports if the watcher has been closed (caller must hold lock).
func (w *Watcher) isClosed() bool {
	return w.state&flagClosed != 0
}

// isWatching reports if the watcher is currently running (caller must hold lock).
func (w *Watcher) isWatching() bool {
	return w.state&flagWatching != 0
}

// IsClosed reports if the watcher has been closed.
// This is safe to call concurrently with other methods.
func (w *Watcher) IsClosed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.isClosed()
}

// IsWatching reports if the watcher is currently running and watching for events.
// This is safe to call concurrently with other methods.
func (w *Watcher) IsWatching() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.isWatching()
}

// checkClosedOp returns an error if the watcher is closed.
// operation is the operation being attempted (e.g., "add path", "remove path").
// The lock must be held by the caller.
func (w *Watcher) checkClosedOp(operation string) error {
	if w.state&flagClosed != 0 {
		return fmt.Errorf("%w: cannot %s on closed watcher", ErrWatcherClosed, operation)
	}

	return nil
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
func New( //nolint:funlen // constructor with full field initialization
	paths []string,
	opts ...Option,
) (*Watcher, error) {
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
		wg:                sync.WaitGroup{},
		eventCh:           nil,
		closeEventChOnce:  sync.Once{},
		debounceInterface: nil,
		errorsCh:          nil,
		errorsMu:          sync.Mutex{},
		errorsOnce:        sync.Once{},
		eventsProcessed:   atomic.Uint64{},
		eventsFilteredOut: atomic.Uint64{},
		errorsEncountered: atomic.Uint64{},
		startTime:         time.Time{},
		lazyIsDir:         false,
		done:              make(chan struct{}),
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

	if w.isClosed() {
		return nil, fmt.Errorf("%w: cannot start watch on closed watcher", ErrWatcherClosed)
	}

	if w.isWatching() {
		return nil, fmt.Errorf("%w: watcher is already running", ErrWatcherRunning)
	}

	// Add initial paths to the fsnotify watcher
	for _, p := range w.paths {
		addErr := w.addPath(NewRootPath(p))
		if addErr != nil {
			return nil, fmt.Errorf("adding watch path %q during Watch(): %w", p, addErr)
		}
	}

	eventCh := make(chan Event, w.bufferSize)
	w.eventCh = eventCh

	w.state |= flagWatching

	// Record start time for uptime tracking
	if w.startTime.IsZero() {
		w.startTime = time.Now()
	}

	w.wg.Add(1)

	go w.watchLoop(ctx, eventCh)

	return eventCh, nil
}

// Add adds a new path to the watcher. The path must be an existing directory.
// This method is safe for concurrent use with other methods.
func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.checkClosedOp("add path")
	if err != nil {
		return err
	}

	abs, resolveErr := filepath.Abs(path)
	if resolveErr != nil {
		return fmt.Errorf("resolving path %q in Add(): %w", path, resolveErr)
	}

	pathErr := w.addPath(NewRootPath(abs))
	if pathErr != nil {
		return fmt.Errorf("adding resolved path %q to watcher: %w", abs, pathErr)
	}

	return nil
}

// Remove removes a path from the watcher. The watcher stops monitoring
// this path and all its subdirectories (if recursive).
// This method is safe for concurrent use with other methods.
func (w *Watcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.checkClosedOp("remove path")
	if err != nil {
		return err
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
	WatchCount        int
	IsWatching        bool
	IsClosed          bool
	EventsProcessed   uint64        // Total events that passed all filters
	EventsFilteredOut uint64        // Events filtered out (dropped by filters)
	ErrorsEncountered uint64        // Errors encountered during processing
	Uptime            time.Duration // Time since watcher was started
}

// Stats returns current statistics about the watcher.
// This method is safe for concurrent use with other methods.
func (w *Watcher) Stats() Stats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var uptime time.Duration
	if !w.startTime.IsZero() {
		uptime = time.Since(w.startTime)
	}

	return Stats{
		WatchCount:        len(w.watchList),
		IsWatching:        w.state&flagWatching != 0,
		IsClosed:          w.state&flagClosed != 0,
		EventsProcessed:   w.eventsProcessed.Load(),
		EventsFilteredOut: w.eventsFilteredOut.Load(),
		ErrorsEncountered: w.errorsEncountered.Load(),
		Uptime:            uptime,
	}
}

// Errors returns a receive-only channel that receives errors from the watcher.
// This provides an alternative to the error handler callback. If both are
// configured, errors are sent to the channel AND passed to the error handler.
// The channel is closed when the watcher is closed.
//
// This method is safe for concurrent use with other methods.
func (w *Watcher) Errors() <-chan error {
	w.errorsOnce.Do(func() {
		w.errorsCh = make(chan error, w.bufferSize)
	})

	return w.errorsCh
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

	// Signal in-flight goroutines to stop before closing channels.
	close(w.done)

	// Stop the debouncer FIRST - waits for all in-flight callbacks to complete.
	// This must happen before closing eventCh to prevent send-on-closed-channel.
	if w.debounceInterface != nil {
		w.debounceInterface.Stop()
	}

	// Close fsnotify watcher - this causes watchLoop to exit.
	err := w.fswatcher.Close()
	if err != nil {
		return fmt.Errorf("closing fsnotify watcher: %w", err)
	}

	// Wait for watchLoop to fully exit before closing eventCh.
	// This ensures no goroutine is mid-send when we close the channel.
	w.wg.Wait()

	// Now safe to close eventCh - watchLoop and all callbacks are done.
	// Use sync.Once to coordinate with watchLoop's defer.
	w.mu.RLock()
	ch := w.eventCh
	w.mu.RUnlock()

	if ch != nil {
		w.closeEventChOnce.Do(func() { close(ch) })
	}

	// Close the errors channel if it was created
	w.errorsMu.Lock()
	if w.errorsCh != nil {
		close(w.errorsCh)
	}
	w.errorsMu.Unlock()

	return nil
}
