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
	"io"
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
		fswatcher:         fswatcher,
		paths:             paths,
		recursive:         true,
		filters:           nil,
		middleware:        nil,
		globalDebounce:    0,
		perPathDebounce:   0,
		errorHandler:      nil,
		skipDotDirs:       true,
		bufferSize:        64,
		onAdd:             nil,
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
		return nil, errors.WithStack(ErrWatcherClosed)
	}

	if w.watching {
		return nil, errors.WithStack(ErrWatcherRunning)
	}

	// Add initial paths to the fsnotify watcher
	for _, p := range w.paths {
		if err := w.addPath(p); err != nil {
			return nil, errors.Wrapf(err, "adding watch path %q", p)
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
		return errors.WithStack(ErrWatcherClosed)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "resolving path %q", path)
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
		return errors.WithStack(ErrWatcherClosed)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "resolving path %q", path)
	}

	if err := w.fswatcher.Remove(abs); err != nil {
		return errors.Wrapf(err, "removing watch path %q", abs)
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
		return errors.Wrap(err, "closing fsnotify watcher")
	}
	return nil
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
		if err := w.fswatcher.Add(root); err != nil {
			return errors.Wrapf(err, "adding watch path %q", root)
		}
		return nil
	}

	return w.walkAndAddPaths(root)
}

// walkAndAddPaths walks a directory tree and adds all directories to the watcher.
func (w *Watcher) walkAndAddPaths(root string) error {
	if err := filepath.WalkDir(root, w.walkDirFunc); err != nil {
		return errors.Wrapf(err, "walking directory %q", root)
	}
	// Track the root path
	w.watchList = append(w.watchList, root)
	return nil
}

// walkDirFunc is the WalkDirFunc for adding paths during directory traversal.
func (w *Watcher) walkDirFunc(path string, d os.DirEntry, err error) error {
	if err != nil {
		return errors.Wrapf(err, "walking path %q", path)
	}

	if !d.IsDir() {
		return nil
	}

	if w.shouldSkipDir(d.Name()) {
		return filepath.SkipDir
	}

	if addErr := w.fswatcher.Add(path); addErr != nil {
		return errors.Wrapf(addErr, "watching path %q", path)
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
	return slices.Contains(DefaultIgnoreDirs, name)
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

	if !w.passesFilters(*event) {
		w.handleFilteredEvent(fsEvent, *event)
		return
	}

	w.handleNewDirectory(fsEvent.Name)
	w.emitEvent(ctx, *event, eventCh)
}

// handleFilteredEvent processes events that don't pass filters.
func (w *Watcher) handleFilteredEvent(fsEvent fsnotify.Event, event Event) {
	if event.Op == Create {
		w.handleNewDirectory(fsEvent.Name)
	}
}

// emitEvent handles the actual event emission with middleware and debouncing.
func (w *Watcher) emitEvent(ctx context.Context, event Event, eventCh chan<- Event) {
	emit := w.buildEmitFunc(ctx, eventCh)
	handler := w.buildMiddlewareHandler(emit)
	w.executeHandler(ctx, event, handler)
}

// buildEmitFunc creates the emit function for sending events.
func (w *Watcher) buildEmitFunc(ctx context.Context, eventCh chan<- Event) func(Event) {
	return func(e Event) {
		select {
		case eventCh <- e:
		case <-ctx.Done():
		default:
		}
	}
}

// buildMiddlewareHandler creates the handler chain with all middleware applied.
func (w *Watcher) buildMiddlewareHandler(emit func(Event)) Handler {
	handler := func(_ context.Context, e Event) {
		emit(e)
	}

	for i := len(w.middleware) - 1; i >= 0; i-- {
		handler = w.wrapWithMiddleware(handler, w.middleware[i])
	}

	return func(ctx context.Context, e Event) error {
		handler(ctx, e)
		return nil
	}
}

// wrapWithMiddleware wraps a handler function with a middleware.
func (w *Watcher) wrapWithMiddleware(
	handler func(context.Context, Event),
	mw Middleware,
) func(context.Context, Event) {
	wrapped := mw(func(ctx context.Context, e Event) error {
		handler(ctx, e)
		return nil
	})
	return func(ctx context.Context, e Event) {
		if err := wrapped(ctx, e); err != nil {
			w.handleError(err)
		}
	}
}

// executeHandler runs the handler, applying debouncing if configured.
func (w *Watcher) executeHandler(ctx context.Context, event Event, handler Handler) {
	execute := func() {
		if err := handler(ctx, event); err != nil {
			w.handleError(err)
		}
	}

	if w.debounceInterface == nil {
		execute()
		return
	}

	key := w.getDebounceKey(event.Path)
	w.debounceInterface.Debounce(key, execute)
}

//nolint:unparam // path param needed for future extensibility
func (w *Watcher) getDebounceKey(path string) string {
	if _, ok := w.debounceInterface.(*Debouncer); ok {
		return path
	}
	return ""
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
//
// Priority of combined operations: Create > Write > Remove > Rename.
// This ensures the most meaningful operation is reported when multiple
// operations occur simultaneously.
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

	// Check if path is a directory. For Remove events, the file may already
	// be gone, so we ignore stat errors in that case.
	isDir := false
	if info, err := os.Stat(fsEvent.Name); err == nil {
		isDir = info.IsDir()
	}

	return &Event{
		Path:      fsEvent.Name,
		Op:        op,
		Timestamp: time.Now(),
		IsDir:     isDir,
	}
}
