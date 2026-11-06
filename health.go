package filewatcher

import (
	"context"
	"errors"
	"fmt"
)

// Healthy reports whether the watcher is currently functioning: it is
// running, admitting events through the configured circuit breaker, and
// holding a working watch registration for every path. It returns nil when
// every condition passes; otherwise a joined error of ALL violated
// conditions — use errors.Is to tell them apart:
//
//   - [ErrWatcherClosed]: Close was called.
//   - [ErrWatcherNotWatching]: Watch has not been called (or Reset stopped it).
//   - [ErrCircuitOpen]: the registered circuit breaker is open, dropping events.
//   - [ErrWatchFailed]: paths whose watch registration failed and has not recovered.
//
// Events dropped by throttling, dedup, or channel backpressure are throughput
// concerns, not health failures: they do not affect the result. A half-open
// circuit reports healthy (probe events are admitted); a failing probe
// re-opens the breaker and the next check reports [ErrCircuitOpen].
//
// The signature matches go-health's CheckFunc (func(ctx context.Context) error),
// so the method value wires in directly:
//
//	health.NewChecks(map[string]health.CheckFunc{"filewatcher": watcher.Healthy})
//
// ctx cancels the check eagerly: an already-cancelled ctx returns ctx.Err()
// without taking the state snapshot. The check itself is a non-blocking read.
func (w *Watcher) Healthy(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("health check cancelled: %w", ctx.Err())
	default:
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	var problems []error

	if w.state&flagClosed != 0 {
		problems = append(problems, ErrWatcherClosed)
	}

	if w.state&flagWatching == 0 {
		problems = append(problems, ErrWatcherNotWatching)
	}

	if w.breakerStateLocked() == CircuitOpen {
		problems = append(problems, ErrCircuitOpen)
	}

	if failed := len(w.failedPaths); failed > 0 {
		problems = append(problems, fmt.Errorf("%w: %d path(s) without a working watch", ErrWatchFailed, failed))
	}

	return errors.Join(problems...)
}
