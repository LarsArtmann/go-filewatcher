package filewatcher

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// pollHealthy polls watcher health until it returns nil or the timeout
// elapses. The initial directory walk is asynchronous, so a freshly started
// watcher becomes healthy only once watching has actually begun.
func pollHealthy(t *testing.T, w *Watcher, timeout time.Duration) error {
	t.Helper()

	deadline := time.Now().Add(timeout)

	var err error

	for time.Now().Before(deadline) {
		if err = w.Healthy(context.Background()); err == nil {
			return nil
		}

		time.Sleep(2 * time.Millisecond)
	}

	return err
}

func TestWatcherHealthy_NotWatchingBeforeWatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	err := w.Healthy(context.Background())
	if err == nil {
		t.Fatal("expected error for watcher that never started, got nil")
	}

	if !errors.Is(err, ErrWatcherNotWatching) {
		t.Errorf("expected ErrWatcherNotWatching, got %v", err)
	}

	if errors.Is(err, ErrWatcherClosed) {
		t.Errorf("unexpected ErrWatcherClosed for watcher that never started: %v", err)
	}
}

func TestWatcherHealthy_JoinsClosedAndNotWatching(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	if _, err := w.Watch(context.Background()); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	err := w.Healthy(context.Background())
	if err == nil {
		t.Fatal("expected error for closed watcher, got nil")
	}

	if !errors.Is(err, ErrWatcherClosed) {
		t.Errorf("expected ErrWatcherClosed, got %v", err)
	}

	if !errors.Is(err, ErrWatcherNotWatching) {
		t.Errorf("expected joined ErrWatcherNotWatching, got %v", err)
	}
}

func TestWatcherHealthy_HealthyWatcherReturnsNil(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	if _, err := w.Watch(context.Background()); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	if err := pollHealthy(t, w, time.Second); err != nil {
		t.Fatalf("expected healthy watcher after Watch, got %v", err)
	}

	stats := w.Stats()
	if stats.FailedPaths != 0 {
		t.Errorf("expected 0 failed paths, got %d", stats.FailedPaths)
	}

	if stats.CircuitBreakerState != CircuitNone {
		t.Errorf("expected CircuitNone without a registered breaker, got %v", stats.CircuitBreakerState)
	}
}

func TestWatcherHealthy_CircuitOpenUnhealthyThenRecovers(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	w := newTestWatcher(t, tmpDir, WithCircuitBreaker(cb))

	if _, err := w.Watch(context.Background()); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	if err := pollHealthy(t, w, time.Second); err != nil {
		t.Fatalf("expected healthy watcher before tripping breaker, got %v", err)
	}

	// Trip the breaker: one failure at maxFailures=1 opens the circuit.
	handler := cb.Middleware()(errReturningHandler())
	_ = handler(context.Background(), testWriteEvent("/trip"))

	if got := cb.State(); got != CircuitOpen {
		t.Fatalf("expected CircuitOpen after failure, got %v", got)
	}

	if got := w.Stats().CircuitBreakerState; got != CircuitOpen {
		t.Errorf("expected Stats.CircuitBreakerState=CircuitOpen, got %v", got)
	}

	err := w.Healthy(context.Background())
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen while breaker is open, got %v", err)
	}

	// After the reset timeout the breaker is half-open: probe events are
	// admitted, so the pipeline functions and health returns nil.
	deadline := time.Now().Add(time.Second)

	for cb.State() != CircuitHalfOpen && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}

	if got := cb.State(); got != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen after reset timeout, got %v", got)
	}

	if err := w.Healthy(context.Background()); err != nil {
		t.Errorf("expected healthy in half-open state, got %v", err)
	}
}

func TestWatcherHealthy_FailedPathsUnhealthy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	if _, err := w.Watch(context.Background()); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	if err := pollHealthy(t, w, time.Second); err != nil {
		t.Fatalf("expected healthy watcher before injecting failure, got %v", err)
	}

	const failedPath = "/gone/missing"

	w.mu.Lock()
	w.failedPaths[w.pathKey(failedPath)] = failedPath
	w.mu.Unlock()

	err := w.Healthy(context.Background())
	if !errors.Is(err, ErrWatchFailed) {
		t.Errorf("expected ErrWatchFailed with failed registration, got %v", err)
	}

	if got := w.Stats().FailedPaths; got != 1 {
		t.Errorf("expected Stats.FailedPaths=1, got %d", got)
	}

	w.mu.Lock()
	delete(w.failedPaths, w.pathKey(failedPath))
	w.mu.Unlock()

	if err := w.Healthy(context.Background()); err != nil {
		t.Errorf("expected healthy after failure cleared, got %v", err)
	}
}

func TestWatcherHealthy_CancelledContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := w.Healthy(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled for cancelled ctx, got %v", err)
	}
}

func TestWatcherHealthy_MatchesCheckFuncShape(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	// Compile-time proof that the method value satisfies go-health's
	// CheckFunc signature: func(ctx context.Context) error.
	var check func(ctx context.Context) error = w.Healthy

	if err := check(context.Background()); !errors.Is(err, ErrWatcherNotWatching) {
		t.Errorf("expected ErrWatcherNotWatching via CheckFunc-shaped call, got %v", err)
	}
}

func TestWithCircuitBreaker_RegistersHandleAndMiddleware(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cb := NewCircuitBreaker(3, time.Second)
	w := newTestWatcher(t, tmpDir, WithCircuitBreaker(cb))

	if w.circuitBreaker != cb {
		t.Fatal("expected WithCircuitBreaker to register the handle")
	}

	if len(w.middleware) != 1 {
		t.Fatalf("expected breaker middleware in pipeline, got %d middleware", len(w.middleware))
	}

	if got := w.Stats().CircuitBreakerState; got != CircuitClosed {
		t.Errorf("expected Stats.CircuitBreakerState=CircuitClosed, got %v", got)
	}
}

func TestWithCircuitBreaker_NilIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir, WithCircuitBreaker(nil))

	if w.circuitBreaker != nil {
		t.Fatal("expected nil breaker to be ignored")
	}

	if len(w.middleware) != 0 {
		t.Fatalf("expected no middleware for nil breaker, got %d", len(w.middleware))
	}

	if got := w.Stats().CircuitBreakerState; got != CircuitNone {
		t.Errorf("expected Stats.CircuitBreakerState=CircuitNone, got %v", got)
	}
}

func TestCircuitBreaker_StateTransitionsWithoutMutation(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(2, 25*time.Millisecond)
	handler := cb.Middleware()(errReturningHandler())

	ctx := context.Background()

	// Two failures at maxFailures=2 open the circuit.
	_ = handler(ctx, testWriteEvent("/a"))
	_ = handler(ctx, testWriteEvent("/b"))

	if got := cb.State(); got != CircuitOpen {
		t.Fatalf("expected CircuitOpen after two failures, got %v", got)
	}

	// State() is read-only: the open→half-open transition stays on the event
	// path, so an immediate second read still reports open.
	if got := cb.State(); got != CircuitOpen {
		t.Errorf("expected State() to be non-mutating, got %v", got)
	}

	// After the reset timeout State() reports half-open; the next event is
	// admitted (probe), fails, and re-opens the circuit.
	deadline := time.Now().Add(time.Second)

	for cb.State() != CircuitHalfOpen && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}

	if got := handler(ctx, testWriteEvent("/c")); got == nil {
		t.Error("expected probe event to pass through and return the handler error")
	}

	if got := cb.State(); got != CircuitOpen {
		t.Errorf("expected CircuitOpen after failed probe, got %v", got)
	}
}

func TestCircuitBreaker_SuccessClosesCircuit(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(1, 25*time.Millisecond)

	var fail atomic.Bool

	failing := func(_ context.Context, _ Event) error {
		if fail.Load() {
			return errTest
		}

		return nil
	}

	handler := cb.Middleware()(failing)
	ctx := context.Background()

	fail.Store(true)
	_ = handler(ctx, testWriteEvent("/a"))

	if got := cb.State(); got != CircuitOpen {
		t.Fatalf("expected CircuitOpen after failure, got %v", got)
	}

	fail.Store(false)

	deadline := time.Now().Add(time.Second)

	for cb.State() != CircuitHalfOpen && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}

	if err := handler(ctx, testWriteEvent("/b")); err != nil {
		t.Errorf("expected successful probe to pass through, got %v", err)
	}

	if got := cb.State(); got != CircuitClosed {
		t.Errorf("expected CircuitClosed after successful probe, got %v", got)
	}
}
