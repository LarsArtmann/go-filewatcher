package filewatcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const logFilePermission = 0o600 // rw------- (owner read/write only) for audit log files

// defaultDedupeWindow is the default time window for deduplicating events.
const defaultDedupeWindow = 100 * time.Millisecond

// dedupeCleanupMultiplier is the multiplier for cleanup ticker interval.
const dedupeCleanupMultiplier = 2

// Middleware wraps an event handler for cross-cutting concerns.
// Middleware is applied in reverse order (last added runs first),
// matching the go-cqrs-lite convention.
type Middleware func(Handler) Handler

// Handler processes a file event.
type Handler func(ctx context.Context, event Event) error

// MiddlewareLogging returns a middleware that logs all events to the
// provided slog logger. If logger is nil, it uses slog.Default().
func MiddlewareLogging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			logger.Info("filewatcher event",
				slog.String("op", event.Op.String()),
				slog.String("path", event.Path),
			)

			return next(ctx, event)
		}
	}
}

// MiddlewareRecovery returns a middleware that recovers from panics in
// downstream handlers, logging the panic value and stack trace.
func MiddlewareRecovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					//nolint:err113 // panic value and stack are inherently dynamic
					err = fmt.Errorf("panic in handler: %v\n%s", r, debug.Stack())
				}
			}()

			return next(ctx, event)
		}
	}
}

// MiddlewareFilter returns a middleware that applies a Filter to events.
// Events that don't pass the filter are dropped.
func MiddlewareFilter(filter Filter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			if !filter(event) {
				return nil
			}

			return next(ctx, event)
		}
	}
}

// MiddlewareOnError returns a middleware that calls the provided callback
// when an error occurs in downstream handlers.
func MiddlewareOnError(onError func(event Event, err error)) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			err := next(ctx, event)
			if err != nil {
				onError(event, err)
			}

			return err
		}
	}
}

// MiddlewareRateLimit returns a middleware that limits the rate of events.
// It allows maxEvents events per second. Events exceeding the limit are dropped.
func MiddlewareRateLimit(maxEvents int) Middleware {
	if maxEvents <= 0 {
		maxEvents = 100
	}

	var (
		count  atomic.Int64
		reset  = time.NewTicker(time.Second)
		closed atomic.Bool
	)

	// Stop ticker when middleware is no longer needed
	// Note: In production, you'd want a way to stop this
	go func() {
		for range reset.C {
			if closed.Load() {
				reset.Stop()
				return
			}
			count.Store(0)
		}
	}()

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			current := count.Add(1)
			if current > int64(maxEvents) {
				return nil
			}

			return next(ctx, event)
		}
	}
}

// MiddlewareSlidingWindowRateLimit returns a middleware that uses a sliding
// window algorithm for rate limiting. This is more accurate than the simple
// counter approach but has slightly more overhead.
func MiddlewareSlidingWindowRateLimit(maxEvents int, window time.Duration) Middleware {
	if maxEvents <= 0 {
		maxEvents = 100
	}

	if window <= 0 {
		window = time.Second
	}

	type windowState struct {
		mu     sync.Mutex
		events []time.Time
	}

	state := &windowState{
		mu:     sync.Mutex{},
		events: nil,
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			now := time.Now()
			cutoff := now.Add(-window)

			state.mu.Lock()

			// Remove events outside the window
			var newEvents []time.Time
			for _, t := range state.events {
				if t.After(cutoff) {
					newEvents = append(newEvents, t)
				}
			}
			state.events = newEvents

			// Check if we're over the limit
			if len(state.events) >= maxEvents {
				state.mu.Unlock()
				return nil
			}

			// Add this event
			state.events = append(state.events, now)
			state.mu.Unlock()

			return next(ctx, event)
		}
	}
}

// MiddlewareMetrics returns a middleware that counts processed events.
// The counter function is called with the event operation type after
// each successful event processing.
func MiddlewareMetrics(counter func(op Op)) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			err := next(ctx, event)
			if err == nil {
				counter(event.Op)
			}

			return err
		}
	}
}

// dedupeKey uniquely identifies an event for deduplication.
type dedupeKey struct {
	path string
	op   Op
}

// MiddlewareDeduplicate returns a middleware that drops duplicate events
// for the same file path and operation within a time window.
// This is useful for reducing noise from rapid successive file operations.
//
// Example: A file saved twice in quick succession generates two events,
// but only the first is processed.
func MiddlewareDeduplicate(window time.Duration) Middleware {
	if window <= 0 {
		window = defaultDedupeWindow
	}

	type seenEntry struct {
		timestamp time.Time
	}

	var (
		mu   sync.Mutex
		seen = make(map[dedupeKey]seenEntry)
	)

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(window * dedupeCleanupMultiplier)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for key, entry := range seen {
				if now.Sub(entry.timestamp) > window {
					delete(seen, key)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			key := dedupeKey{path: event.Path, op: event.Op}

			mu.Lock()
			entry, exists := seen[key]
			if exists && time.Since(entry.timestamp) < window {
				mu.Unlock()
				// Duplicate detected, drop this event
				return nil
			}

			seen[key] = seenEntry{timestamp: time.Now()}
			mu.Unlock()

			return next(ctx, event)
		}
	}
}

// defaultBatchWindow is the default time window for batching events.
const defaultBatchWindow = 100 * time.Millisecond

// defaultBatchSize is the default maximum number of events in a batch.
const defaultBatchSize = 100

// MiddlewareBatch returns a middleware that batches events over a window
// and emits them all at once. The flush function is called with all batched
// events when the window expires or the batch reaches max size.
//
// The flush function receives the batched events and should process them.
// If it returns an error, the error is passed to the next handler.
// If it returns nil, processing continues normally.
//
//nolint:funlen // Complex middleware requiring inline logic
func MiddlewareBatch(window time.Duration, maxSize int, flush func([]Event) error) Middleware {
	if window <= 0 {
		window = defaultBatchWindow
	}

	if maxSize <= 0 {
		maxSize = defaultBatchSize
	}

	type batchState struct {
		mu     sync.Mutex
		events []Event
		timer  *time.Timer
	}

	state := &batchState{
		events: make([]Event, 0, maxSize),
		mu:     sync.Mutex{},
		timer:  nil,
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			state.mu.Lock()

			state.events = append(state.events, event)

			// If batch is full, flush immediately
			if len(state.events) >= maxSize {
				events := state.events
				state.events = make([]Event, 0, maxSize)

				if state.timer != nil {
					state.timer.Stop()
					state.timer = nil
				}

				state.mu.Unlock()

				err := flush(events)
				if err != nil {
					return err
				}

				return next(ctx, event)
			}

			// Start or reset timer
			if state.timer == nil {
				state.timer = time.AfterFunc(window, func() {
					state.mu.Lock()
					events := state.events
					state.events = make([]Event, 0, maxSize)
					state.timer = nil
					state.mu.Unlock()

					if len(events) > 0 {
						_ = flush(events)
					}
				})
			}

			state.mu.Unlock()

			return next(ctx, event)
		}
	}
}

// MiddlewareWriteFileLog returns a middleware that appends event logs
// to a file at the given path. This is useful for audit trails.
// The file handle is cached for the lifetime of the middleware.
func MiddlewareWriteFileLog(filePath string) Middleware {
	type cachedFile struct {
		mu sync.Mutex
		f  *os.File
	}

	//nolint:exhaustruct // f is lazily initialized on first write
	cf := &cachedFile{}

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			cf.mu.Lock()

			var writeErr error

			if cf.f == nil {
				cf.f, writeErr = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePermission)
			}

			if writeErr == nil && cf.f != nil {
				_, writeErr = fmt.Fprintf(cf.f, "[%s] %s: %s\n",
					event.Timestamp.Format(time.RFC3339),
					event.Op,
					event.Path,
				)
			}

			cf.mu.Unlock()

			err := next(ctx, event)
			if err != nil {
				return err
			}

			return writeErr
		}
	}
}
