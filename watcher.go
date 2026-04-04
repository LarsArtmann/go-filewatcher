// Package filewatcher provides a high-level, composable file system watcher
// built on top of fsnotify. It eliminates the boilerplate of raw fsnotify
// usage by providing sensible defaults for common patterns:
//   - Automatic recursive directory watching
//   - Configurable debounce (global or per-path)
//   - Composable event filters (extension, directory, glob)
//   - Middleware chains for cross-cutting concerns
//   - Graceful shutdown via context cancellation
//
// Design principles (matching go-cqrs-lite conventions):
//   - Functional options for configuration
//   - Sentinel errors with cockroachdb/errors
//   - No panics, explicit error handling
//   - Context as first parameter
//   - Channel-based event streaming
package filewatcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
)

// DefaultIgnoreDirs contains commonly ignored directory names.
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

	// Internal state
	mu     sync.RWMutex
	closed bool

	// Debouncer (initialized based on config)
	debounceInterface interface {
		Debounce(key string, fn func())
		Stop()
	}
}

// New creates a new Watcher for the given paths with the specified options.
// At least one path must be provided. Paths are validated to exist.
//
// The watcher is not started until Watch() is called.
func New(paths []string, opts ...Option) (*Watcher, error) {
	if len(paths) == 0 {
		return nil, errors.WithStack(ErrNoPaths)
	}

	// Validate all paths exist
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving path %q", p)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, errors.Wrapf(ErrPathNotFound, "path %q (resolved: %q)", p, abs)
		}
		if !info.IsDir() {
			return nil, errors.Wrapf(ErrPathNotDir, "path %q", p)
		}
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "creating fsnotify watcher")
	}

	w := &Watcher{
		fswatcher: fswatcher,
		paths:     paths,
		recursive: true,
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
		return nil, errors.WithStack(ErrWatcherClosed)
	}

	// Add initial paths to the fsnotify watcher
	for _, p := range w.paths {
		if err := w.addPath(p); err != nil {
			return nil, errors.Wrapf(err, "adding watch path %q", p)
		}
	}

	eventCh := make(chan Event, 64)

	go w.watchLoop(ctx, eventCh)

	return eventCh, nil
}

// Add adds a new path to the watcher. The path must be an existing directory.
func (w *Watcher) Add(path string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return errors.WithStack(ErrWatcherClosed)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "resolving path %q", path)
	}

	return w.addPath(abs)
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

	if w.debounceInterface != nil {
		w.debounceInterface.Stop()
	}

	return w.fswatcher.Close()
}

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
		return w.fswatcher.Add(root)
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories
		name := d.Name()
		if strings.HasPrefix(name, ".") && name != "." {
			return filepath.SkipDir
		}

		// Check against default ignore dirs
		if slices.Contains(DefaultIgnoreDirs, name) {
			return filepath.SkipDir
		}

		if addErr := w.fswatcher.Add(path); addErr != nil {
			return errors.Wrapf(addErr, "watching path %q", path)
		}

		return nil
	})
}

// watchLoop is the main event processing goroutine.
func (w *Watcher) watchLoop(ctx context.Context, eventCh chan<- Event) {
	defer close(eventCh)

	for {
		select {
		case <-ctx.Done():
			return

		case fsEvent, ok := <-w.fswatcher.Events:
			if !ok {
				return
			}
			w.processEvent(ctx, fsEvent, eventCh)

		case err, ok := <-w.fswatcher.Errors:
			if !ok {
				return
			}
			w.handleError(err)
		}
	}
}

// processEvent converts an fsnotify event, applies filters and debounce,
// and emits it to the channel.
func (w *Watcher) processEvent(ctx context.Context, fsEvent fsnotify.Event, eventCh chan<- Event) {
	event := convertEvent(fsEvent)
	if event == nil {
		return
	}

	// Apply filters
	if !w.passesFilters(*event) {
		// Still handle dynamic directory watching even if filtered
		if event.Op == Create {
			w.handleNewDirectory(fsEvent.Name)
		}
		return
	}

	// Handle dynamic directory watching
	if event.Op == Create {
		w.handleNewDirectory(fsEvent.Name)
	}

	// Build the emit function with middleware
	emit := func(e Event) {
		select {
		case eventCh <- e:
		case <-ctx.Done():
		default:
		}
	}

	// Wrap with middleware chain (applied in reverse order)
	handler := func(_ context.Context, e Event) error {
		emit(e)
		return nil
	}
	for i := len(w.middleware) - 1; i >= 0; i-- {
		mw := w.middleware[i]
		currentHandler := handler
		handler = func(ctx context.Context, e Event) error {
			return mw(func(_ context.Context, e Event) error {
				return currentHandler(ctx, e)
			})(ctx, e)
		}
	}

	// Apply debounce or emit directly
	if w.debounceInterface != nil {
		key := ""
		if _, ok := w.debounceInterface.(*Debouncer); ok {
			key = event.Path
		}
		w.debounceInterface.Debounce(key, func() {
			_ = handler(ctx, *event)
		})
	} else {
		_ = handler(ctx, *event)
	}
}

// handleNewDirectory adds newly created directories to the watcher
// when recursive mode is enabled.
func (w *Watcher) handleNewDirectory(path string) {
	if !w.recursive {
		return
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}

	w.mu.RLock()
	closed := w.closed
	w.mu.RUnlock()

	if closed {
		return
	}

	_ = w.addPath(path)
}

// passesFilters checks if an event passes all registered filters.
func (w *Watcher) passesFilters(event Event) bool {
	for _, f := range w.filters {
		if !f(event) {
			return false
		}
	}
	return true
}

// handleError dispatches errors to the configured handler or stderr.
func (w *Watcher) handleError(err error) {
	if w.errorHandler != nil {
		w.errorHandler(err)
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "filewatcher: %v\n", err)
}

// convertEvent converts an fsnotify.Event to a filewatcher.Event.
// Returns nil for operations that are not mapped (e.g., Chmod).
func convertEvent(fsEvent fsnotify.Event) *Event {
	var op Op

	switch {
	case fsEvent.Op&fsnotify.Create == fsnotify.Create:
		op = Create
	case fsEvent.Op&fsnotify.Write == fsnotify.Write:
		op = Write
	case fsEvent.Op&fsnotify.Remove == fsnotify.Remove:
		op = Remove
	case fsEvent.Op&fsnotify.Rename == fsnotify.Rename:
		op = Rename
	default:
		return nil
	}

	return &Event{
		Path:      fsEvent.Name,
		Op:        op,
		Timestamp: time.Now(),
	}
}
