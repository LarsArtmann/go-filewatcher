//nolint:varnamelen,exhaustruct // Idiomatic short names: op (operation); partial ErrorContext initialization acceptable
package filewatcher

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

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

			w.handleError(ErrorContext{Operation: "fsnotify", Retryable: true}, err)
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

	return wrapHandlerWithNilReturn(handler)
}

// wrapHandlerWithNilReturn wraps a handler to return nil error.
func wrapHandlerWithNilReturn(handler func(context.Context, Event)) Handler {
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
	wrapped := mw(wrapHandlerWithNilReturn(handler))

	return func(ctx context.Context, e Event) {
		err := wrapped(ctx, e)
		if err != nil {
			w.handleError(
				ErrorContext{Operation: "middleware", Path: e.Path, Retryable: false},
				fmt.Errorf("middleware error: %w", err),
			)
		}
	}
}

// executeHandler runs the handler, applying debouncing if configured.
func (w *Watcher) executeHandler(ctx context.Context, event Event, handler Handler) {
	execute := func() {
		err := handler(ctx, event)
		if err != nil {
			w.handleError(
				ErrorContext{Operation: "handler", Path: event.Path, Retryable: false},
				fmt.Errorf("handler error: %w", err),
			)
		}
	}

	if w.debounceInterface == nil {
		execute()

		return
	}

	key := w.getDebounceKey(event.Path)
	w.debounceInterface.Debounce(key, execute)
}

func (w *Watcher) getDebounceKey(path string) DebounceKey {
	if _, ok := w.debounceInterface.(*Debouncer); ok {
		return DebounceKey(path)
	}

	return DebounceKey("")
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
func (w *Watcher) handleError(ctx ErrorContext, err error) {
	if w.errorHandler != nil {
		w.errorHandler(ctx, err)

		return
	}

	if ctx.Path != "" {
		_, _ = fmt.Fprintf(os.Stderr, "filewatcher: %s: %s: %v\n", ctx.Operation, ctx.Path, err)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "filewatcher: %s: %v\n", ctx.Operation, err)
	}
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

	info, err := os.Stat(fsEvent.Name)
	if err == nil {
		isDir = info.IsDir()
	}

	return &Event{
		Path:      fsEvent.Name,
		Op:        op,
		Timestamp: time.Now(),
		IsDir:     isDir,
	}
}
