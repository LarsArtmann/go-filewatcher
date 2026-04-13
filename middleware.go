//nolint:varnamelen // Idiomatic short names: cf (cached file)
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
					err = fmt.Errorf(
						"recovered from panic in event handler: %v\n%s",
						r,
						debug.Stack(),
					)
				}
			}()

			return next(ctx, event)
		}
	}
}

// MiddlewareRateLimit returns a middleware that limits the rate of event
// processing to at most one event per minInterval.
func MiddlewareRateLimit(minInterval time.Duration) Middleware {
	var lastEvent atomic.Int64 // stores UnixNano for atomic operations

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			now := time.Now().UnixNano()

			last := lastEvent.Load()
			if now-last < minInterval.Nanoseconds() {
				return nil
			}

			if lastEvent.CompareAndSwap(last, now) {
				return next(ctx, event)
			}

			return nil
		}
	}
}

// rateLimiter implements a simple token bucket rate limiter.
type rateLimiter struct {
	maxEvents int
	window    time.Duration
	events    []time.Time
	mu        sync.Mutex
}

// allow checks if a new event is allowed under the rate limit.
func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove events outside the window
	validEvents := make([]time.Time, 0, len(rl.events))
	for _, t := range rl.events {
		if t.After(cutoff) {
			validEvents = append(validEvents, t)
		}
	}

	rl.events = validEvents

	// Check if we can add a new event
	if len(rl.events) < rl.maxEvents {
		rl.events = append(rl.events, now)

		return true
	}

	return false
}

// MiddlewareRateLimitWindow returns a middleware that limits the rate of event
// processing to maxEvents per window duration using a sliding window algorithm.
func MiddlewareRateLimitWindow(maxEvents int, window time.Duration) Middleware {
	if maxEvents <= 0 {
		maxEvents = 1
	}

	if window <= 0 {
		window = time.Second
	}

	limiter := &rateLimiter{
		maxEvents: maxEvents,
		window:    window,
		events:    make([]time.Time, 0, maxEvents),
		mu:        sync.Mutex{},
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			if !limiter.allow() {
				return nil
			}

			return next(ctx, event)
		}
	}
}

// MiddlewareFilter returns a middleware that drops events that do not
// match the given filter.
func MiddlewareFilter(f Filter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			if !f(event) {
				return nil
			}

			return next(ctx, event)
		}
	}
}

// MiddlewareOnError returns a middleware that calls handler when the
// downstream handler returns an error.
func MiddlewareOnError(handler func(event Event, err error)) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			err := next(ctx, event)
			if err != nil {
				handler(event, err)

				return err
			}

			return nil
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
				//nolint:gosec // filePath is user-provided, intentional design for log file location
				cf.f, writeErr = os.OpenFile(
					filePath,
					os.O_APPEND|os.O_CREATE|os.O_WRONLY,
					logFilePermission,
				)
			}

			if writeErr == nil && cf.f != nil {
				_, _ = fmt.Fprintf(
					cf.f,
					"%s %s %s\n",
					event.Timestamp.Format(time.RFC3339),
					event.Op.String(),
					event.Path,
				)
			}
			cf.mu.Unlock()

			return next(ctx, event)
		}
	}
}
