package filewatcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// Middleware wraps an event handler for cross-cutting concerns.
// Middleware is applied in reverse order (last added runs first),
// matching the go-cqrs-lite convention.
type Middleware func(Handler) Handler

// Handler processes a file event.
type Handler func(ctx context.Context, event Event) error

// MiddlewareLogging returns a middleware that logs all events to the
// provided logger. If logger is nil, it uses log.Default().
func MiddlewareLogging(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.Default()
	}
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			logger.Printf("filewatcher: %s %s", event.Op.String(), event.Path)
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
	var lastEvent int64 // stores UnixNano for atomic operations
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			now := time.Now().UnixNano()
			last := atomic.LoadInt64(&lastEvent)
			if now-last < minInterval.Nanoseconds() {
				return nil
			}
			if atomic.CompareAndSwapInt64(&lastEvent, last, now) {
				return next(ctx, event)
			}
			return nil
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
			if err := next(ctx, event); err != nil {
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
func MiddlewareWriteFileLog(filePath string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			//nolint:gosec // filePath is user-provided, intentional design for log file location
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
			if err == nil {
				_, _ = fmt.Fprintf(
					f,
					"%s %s %s\n",
					event.Timestamp.Format(time.RFC3339),
					event.Op.String(),
					event.Path,
				)
				_ = f.Close()
			}
			return next(ctx, event)
		}
	}
}
