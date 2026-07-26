package filewatcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Test sentinel errors for error simulation (package-level to satisfy err113).
var (
	errFakeENOSPC      = errors.New("no space left on device")
	errFakePermDenied  = errors.New("permission denied")
	errFakeBackend     = errors.New("simulated backend error")
	errFakeHandlerFail = errors.New("handler failure")
)

// cancelAndDrain cancels the watch context and drains the events channel
// until it closes. Centralizes the test cleanup pattern.
func cancelAndDrain(cancel context.CancelFunc, events <-chan Event) {
	cancel()

	for range events {
	}
}

// newTestWatcherWithFake creates a watcher backed by a fakeBackend instead of
// a real fsnotify watcher. The watcher is auto-cleaned via t.Cleanup.
func newTestWatcherWithFake(
	t *testing.T,
	tmpDir string,
	fake *fakeBackend,
	opts ...Option,
) *Watcher {
	t.Helper()

	opts = append(opts, withBackend(fake))

	return newTestWatcher(t, tmpDir, opts...)
}

// --- Self-heal tests with injected Add failures ---

func TestSelfHeal_HealsFailedPathAfterRetry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	// Simulate ENOSPC: all Add calls fail on first attempt.
	fake.addFn = func(_ string, _ int) error {
		return errFakeENOSPC
	}

	watcher := newTestWatcherWithFake(t, tmpDir, fake, WithSelfHeal(50*time.Millisecond))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// All paths should be in failedPaths since Add always fails.
	waitForCondition(t, time.Second, "failedPaths populated", func() bool {
		return watcher.failedPathCount() > 0
	})

	initialFailed := watcher.failedPathCount()
	if initialFailed == 0 {
		t.Fatal("expected failed paths after ENOSPC")
	}

	// Verify WatchErrors counter was incremented.
	if got := watcher.Stats().WatchErrors; got == 0 {
		t.Error("expected WatchErrors > 0")
	}

	// Now make Add succeed — self-heal should re-register paths.
	fake.mu.Lock()
	fake.addFn = nil // nil = always succeed
	fake.mu.Unlock()

	waitForCondition(t, 2*time.Second, "failedPaths healed to zero", func() bool {
		return watcher.failedPathCount() == 0
	})

	cancelAndDrain(cancel, events)
}

func TestSelfHeal_AbandonsPermanentError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	// Return a permanent error (wraps ErrPathNotDir which is CategoryPermanent).
	permanentErr := fmt.Errorf("%w: simulated permanent failure", ErrPathNotDir)

	fake.addFn = func(_ string, _ int) error {
		return permanentErr
	}

	watcher := newTestWatcherWithFake(t, tmpDir, fake, WithSelfHeal(50*time.Millisecond))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Initially, paths are in failedPaths.
	waitForCondition(t, time.Second, "failedPaths populated", func() bool {
		return watcher.failedPathCount() > 0
	})

	// After self-heal runs, permanent errors should be abandoned (removed from
	// failedPaths because they will never succeed).
	waitForCondition(t, 2*time.Second, "permanent paths abandoned", func() bool {
		return watcher.failedPathCount() == 0
	})

	cancelAndDrain(cancel, events)
}

// --- Error channel propagation tests ---

func TestErrorChannel_PropagatesBackendErrorToHandler(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	type capturedError struct {
		ctx ErrorContext
		err error
	}

	errReceived := make(chan capturedError, 1)

	watcher := newTestWatcherWithFake(t, tmpDir, fake, WithErrorHandler(func(ctx ErrorContext, err error) {
		select {
		case errReceived <- capturedError{ctx, err}:
		default:
		}
	}))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Inject an error via the fake backend's error channel.
	fake.sendError(errFakeBackend)

	select {
	case captured := <-errReceived:
		if !errors.Is(captured.err, errFakeBackend) {
			t.Errorf("error handler got %v, want wrap of %v", captured.err, errFakeBackend)
		}

		if captured.ctx.Operation != operationFsnotify {
			t.Errorf("operation = %q, want %q", captured.ctx.Operation, operationFsnotify)
		}

		if !captured.ctx.Retryable {
			t.Error("expected Retryable=true for fsnotify errors")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error handler to receive backend error")
	}

	cancelAndDrain(cancel, events)
}

func TestErrorChannel_PropagatesToErrorsChannel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	errCh := watcher.Errors()

	fake.sendError(errFakePermDenied)

	select {
	case received := <-errCh:
		if !errors.Is(received, errFakePermDenied) {
			t.Errorf("Errors() got %v, want wrap of %v", received, errFakePermDenied)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error on Errors() channel")
	}

	cancelAndDrain(cancel, events)
}

// --- Full pipeline event flow with fake backend ---

func TestEventPipeline_EventsFlowThroughFakeBackend(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Inject a Create event for a .go file.
	testFile := filepath.Join(tmpDir, "test.go")

	fake.sendEvent(fsnotify.Event{
		Name: testFile,
		Op:   fsnotify.Create,
	})

	select {
	case event := <-events:
		if event.Path != testFile {
			t.Errorf("event.Path = %q, want %q", event.Path, testFile)
		}

		if event.Op != Create {
			t.Errorf("event.Op = %v, want %v", event.Op, Create)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event from fake backend")
	}

	cancelAndDrain(cancel, events)
}

func TestEventPipeline_MiddlewareProcessesFakeEvents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	var eventCount atomic.Int32

	watcher := newTestWatcherWithFake(
		t, tmpDir, fake,
		WithMiddleware(MiddlewareMetrics(func(_ Op) {
			eventCount.Add(1)
		})),
	)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Inject multiple events.
	for i := range 3 {
		fake.sendEvent(fsnotify.Event{
			Name: filepath.Join(tmpDir, "file"+string(rune('A'+i))+".go"),
			Op:   fsnotify.Create,
		})
	}

	waitForCondition(t, 2*time.Second, "3 events processed by middleware", func() bool {
		return eventCount.Load() >= 3
	})

	cancelAndDrain(cancel, events)
}

// --- Closed backend graceful shutdown ---

func TestClosedBackend_WatchLoopStopsGracefully(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Close the fake backend's channels — simulates fsnotify watcher closing.
	_ = fake.Close()

	// Watch loop should exit and events channel should close.
	select {
	case _, ok := <-events:
		if ok {
			// Got an event, then channel should close next.
			for range events {
			}
		}
		// Channel closed — expected.
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: events channel did not close after backend closed")
	}
}

// --- Circuit breaker through the pipeline ---
// CircuitBreaker is tested at the middleware level because the watcher
// pipeline wraps each middleware with wrapHandlerWithNilReturn, which absorbs
// errors from inner layers. This means circuit-breaker failures from inner
// middleware never propagate to the breaker. Testing at the middleware level
// verifies the state machine directly.

func TestCircuitBreaker_DropsEventsAfterFailures(t *testing.T) {
	t.Parallel()

	var handlerCalls atomic.Int32

	failHandler := func(_ context.Context, _ Event) error {
		handlerCalls.Add(1)

		return errFakeHandlerFail
	}

	breaker := MiddlewareCircuitBreaker(2, 50*time.Millisecond)
	wrapped := breaker(failHandler)

	// First 2 calls pass through and return errors.
	for range 2 {
		_ = wrapped(context.Background(), testWriteEvent("/test1.go"))
	}

	// Circuit should now be open — next call is dropped (nil return, no handler call).
	before := handlerCalls.Load()

	_ = wrapped(context.Background(), testWriteEvent("/test2.go"))

	after := handlerCalls.Load()

	if after != before {
		t.Errorf("circuit should be open: handler called %d times, expected no new calls", after-before)
	}

	// Wait for reset timeout — circuit goes half-open.
	time.Sleep(60 * time.Millisecond)

	// Half-open: one event passes through.
	_ = wrapped(context.Background(), testWriteEvent("/test3.go"))

	if got := handlerCalls.Load(); got != before+1 {
		t.Errorf("half-open: expected 1 new handler call, got %d", got-before)
	}
}

func TestErrorRecovery_StrategyReceivesEventAndError(t *testing.T) {
	t.Parallel()

	var (
		capturedEvent Event
		capturedErr   error
	)

	strategy := func(event Event, err error) error {
		capturedEvent = event
		capturedErr = err

		return nil // recovered
	}

	mw := MiddlewareErrorRecovery(strategy)
	handler := mw(errReturningHandler())

	testEvt := testWriteEvent("/recover.go")

	err := handler(context.Background(), testEvt)
	if err != nil {
		t.Errorf("expected recovery to suppress error, got %v", err)
	}

	if capturedEvent.Path != testEvt.Path {
		t.Errorf("strategy received event.Path = %q, want %q", capturedEvent.Path, testEvt.Path)
	}

	if capturedErr == nil {
		t.Error("strategy received nil error, expected non-nil")
	}
}

// --- Watch errors counter verification ---

func TestWatchErrors_CounterIncrementsOnAddFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	fake.addFn = func(_ string, _ int) error {
		return errFakeENOSPC
	}

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	waitForCondition(t, time.Second, "WatchErrors counter > 0", func() bool {
		return watcher.Stats().WatchErrors > 0
	})

	cancelAndDrain(cancel, events)
}

// --- Multiple error injection types ---

func TestFakeBackend_AddFailsSpecificPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fake := newFakeBackend()

	// Create a subdirectory to have multiple paths.
	subDir := filepath.Join(tmpDir, "sub")

	err := os.MkdirAll(subDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	// Only fail Add for the subdirectory path.
	fake.addFn = func(path string, _ int) error {
		if path == subDir {
			return errFakePermDenied
		}

		return nil
	}

	watcher := newTestWatcherWithFake(t, tmpDir, fake)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	events, watchErr := watcher.Watch(ctx)
	if watchErr != nil {
		t.Fatalf("Watch failed: %v", watchErr)
	}

	waitForCondition(t, time.Second, "subDir in failedPaths", func() bool {
		watcher.mu.RLock()
		defer watcher.mu.RUnlock()

		_, failed := watcher.failedPaths[subDir]

		return failed
	})

	// Root path should NOT be in failedPaths.
	watcher.mu.RLock()

	_, rootFailed := watcher.failedPaths[tmpDir]

	watcher.mu.RUnlock()

	if rootFailed {
		t.Error("root path should not be in failedPaths")
	}

	cancelAndDrain(cancel, events)
}
