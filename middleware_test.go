//nolint:varnamelen // Idiomatic short names: mw (middleware), mu (mutex)
package filewatcher

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// flushToSlice creates a flush function that appends events to a slice.
func flushToSlice(slice *[]Event) func([]Event) error {
	return func(events []Event) error {
		*slice = append(*slice, events...)

		return nil
	}
}

// assertErrorIs asserts that err wraps the target error.
func assertErrorIs(t *testing.T, err, target error, msg string) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("%s, got %v", msg, err)
	}
}

// assertBatchLen asserts the length of a batched event slice.
func assertBatchLen(t *testing.T, batched []Event, want int, msg string) {
	t.Helper()

	if len(batched) != want {
		t.Errorf("expected %d batched events%s, got %d", want, msg, len(batched))
	}
}

func TestResolveRateLimitDefaults(t *testing.T) {
	t.Parallel()

	const defaultMax = 50

	tests := []struct {
		name       string
		maxValue   int
		window     time.Duration
		wantMax    int
		wantWindow time.Duration
	}{
		{"both positive pass through", 100, 5 * time.Second, 100, 5 * time.Second},
		{"zero max uses provided default", 0, 5 * time.Second, defaultMax, 5 * time.Second},
		{"negative max uses provided default", -1, 5 * time.Second, defaultMax, 5 * time.Second},
		{"zero window uses package default", 100, 0, 100, defaultRateLimitWindow},
		{"negative window uses package default", 100, -1, 100, defaultRateLimitWindow},
		{"both non-positive use both defaults", 0, 0, defaultMax, defaultRateLimitWindow},
	}

	for _, tc := range tests {
		gotMax, gotWindow := resolveRateLimitDefaults(tc.maxValue, defaultMax, tc.window)
		if gotMax != tc.wantMax {
			t.Errorf("%s: max = %d, want %d", tc.name, gotMax, tc.wantMax)
		}

		if gotWindow != tc.wantWindow {
			t.Errorf("%s: window = %v, want %v", tc.name, gotWindow, tc.wantWindow)
		}
	}
}

func TestResolveMaxFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		maxFailures int
		want        int
	}{
		{"positive passes through", 7, 7},
		{"zero uses package default", 0, defaultMaxFailures},
		{"negative uses package default", -3, defaultMaxFailures},
	}

	for _, tc := range tests {
		if got := resolveMaxFailures(tc.maxFailures); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestResolveBatchDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		window      time.Duration
		maxSize     int
		wantWindow  time.Duration
		wantMaxSize int
	}{
		{"both positive pass through", 200 * time.Millisecond, 50, 200 * time.Millisecond, 50},
		{"zero window uses package default", 0, 50, defaultBatchWindow, 50},
		{"negative window uses package default", -1, 50, defaultBatchWindow, 50},
		{"zero maxSize uses package default", 200 * time.Millisecond, 0, 200 * time.Millisecond, defaultBatchSize},
		{"negative maxSize uses package default", 200 * time.Millisecond, -1, 200 * time.Millisecond, defaultBatchSize},
		{"both non-positive use both defaults", 0, 0, defaultBatchWindow, defaultBatchSize},
	}

	for _, tc := range tests {
		gotWindow, gotMaxSize := resolveBatchDefaults(tc.window, tc.maxSize)
		if gotWindow != tc.wantWindow {
			t.Errorf("%s: window = %v, want %v", tc.name, gotWindow, tc.wantWindow)
		}

		if gotMaxSize != tc.wantMaxSize {
			t.Errorf("%s: maxSize = %d, want %d", tc.name, gotMaxSize, tc.wantMaxSize)
		}
	}
}

func TestMiddlewareLogging(t *testing.T) {
	t.Parallel()

	var buf strings.Builder

	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)

	mw := MiddlewareLogging(logger)
	testHandler := mw(noopHandler())

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	content := buf.String()
	assertLogContains(t, content, NewLogSubstring("filewatcher event"))
	assertLogContains(t, content, NewLogSubstring("WRITE"))
	assertLogContains(t, content, NewLogSubstring("/tmp/test.go"))
}

func TestMiddlewareLogging_NilLogger(t *testing.T) {
	t.Parallel()

	mw := MiddlewareLogging(nil)
	testHandler := mw(noopHandler())

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMiddlewareRecovery(t *testing.T) {
	t.Parallel()

	panicHandler := func(_ context.Context, _ Event) error {
		panic("intentional panic")
	}

	mw := MiddlewareRecovery()
	testHandler := mw(panicHandler)

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err == nil {
		t.Error("expected error from panic recovery, got nil")
	}

	assertErrContains(t, err, "panic in handler")
}

func TestMiddlewareRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	var called bool

	normalHandler := calledFlagHandler(&called)

	mw := MiddlewareRecovery()
	testHandler := mw(normalHandler)

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestMiddlewareFilter(t *testing.T) {
	t.Parallel()

	filter := FilterExtensions(".go")
	mw := MiddlewareFilter(filter)

	var processed bool

	handler := mw(func(_ context.Context, _ Event) error {
		processed = true

		return nil
	})

	// Event matching filter should be processed
	_ = handler(context.Background(), testWriteEvent("/tmp/test.go"))

	if !processed {
		t.Error("expected event to be processed")
	}

	// Event not matching filter should be dropped
	processed = false
	_ = handler(context.Background(), testWriteEvent("/tmp/test.txt"))

	if processed {
		t.Error("expected event to be dropped")
	}
}

func TestMiddlewareOnError(t *testing.T) {
	t.Parallel()

	var (
		capturedEvent *Event
		capturedErr   error
	)

	onError := func(event Event, err error) {
		capturedEvent = &event
		capturedErr = err
	}

	mw := MiddlewareOnError(onError)    //nolint:varnamelen // idiomatic middleware abbreviation
	testErr := errors.New("test error") //nolint:err113 // test-specific dynamic error
	testHandler := mw(handlerReturning(testErr))

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))

	assertErrorIs(t, err, testErr, "expected error to be passed through")

	if capturedEvent == nil {
		t.Error("expected onError to be called with event")
	}

	if !errors.Is(capturedErr, testErr) {
		t.Errorf("expected onError to receive error %v, got %v", testErr, capturedErr)
	}
}

func TestMiddlewareOnError_NoError(t *testing.T) {
	t.Parallel()

	onErrorCalled := false
	onError := func(_ Event, _ error) {
		onErrorCalled = true
	}

	mw := MiddlewareOnError(onError)
	testHandler := mw(noopHandler())

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if onErrorCalled {
		t.Error("expected onError not to be called")
	}
}

func TestMiddlewareRateLimit(t *testing.T) {
	t.Parallel()

	mw := MiddlewareRateLimit(2)

	var processed int

	handler := mw(testHandlerFunc(&processed))

	ctx := context.Background()
	event := testWriteEvent("/tmp/test.go")

	// First 2 should be processed
	for range 2 {
		_ = handler(ctx, event)
	}

	// Third should be dropped
	_ = handler(ctx, event)

	if processed != 2 {
		t.Errorf("expected 2 events processed, got %d", processed)
	}
}

func TestMiddlewareSlidingWindowRateLimit(t *testing.T) {
	t.Parallel()

	mw := MiddlewareSlidingWindowRateLimit(2, 100*time.Millisecond)

	var processed int

	handler := mw(testHandlerFunc(&processed))

	ctx := context.Background()
	event := testWriteEvent("/tmp/test.go")

	// First 2 should be processed
	for range 2 {
		_ = handler(ctx, event)
	}

	// Third should be dropped
	_ = handler(ctx, event)

	if processed != 2 {
		t.Errorf("expected 2 events processed, got %d", processed)
	}
}

func TestMiddlewareMetrics(t *testing.T) {
	t.Parallel()

	var opCounts [4]int

	counter := func(op Op) {
		opCounts[op]++
	}

	mw := MiddlewareMetrics(counter)
	testHandler := mw(noopHandler())

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	assertOpCount(t, opCounts, Write, 1)
}

func TestMiddlewareMetrics_ErrorNotCounted(t *testing.T) {
	t.Parallel()

	var opCounts [4]int

	counter := func(op Op) {
		opCounts[op]++
	}

	mw := MiddlewareMetrics(counter)
	testHandler := mw(handlerReturning(errors.New("test error"))) //nolint:err113 // test-specific dynamic error

	_ = testHandler(context.Background(), testWriteEvent("/tmp/test.go"))

	assertOpCount(t, opCounts, Write, 0)
}

func TestMiddlewareWriteFileLog(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/events.log"

	mw := MiddlewareWriteFileLog(tmpFile)
	handler := mw(noopHandler())

	e := fixedWriteEvent("/tmp/test.go")

	err := handler(context.Background(), e)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(tmpFile) //nolint:gosec // test file from TempDir
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	assertLogContains(t, content, NewLogSubstring("WRITE"))
	assertLogContains(t, content, NewLogSubstring("/tmp/test.go"))
}

func TestMiddlewareWriteFileLog_Appends(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/events.log"

	mw := MiddlewareWriteFileLog(tmpFile)
	handler := mw(noopHandler())

	_ = handler(context.Background(), testWriteEvent("a.go"))
	_ = handler(context.Background(), testEvent("b.go", Create))

	data, err := os.ReadFile(tmpFile) //nolint:gosec // test file from TempDir
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)

	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines in log, got %d: %q", len(lines), content)
	}
}

func TestMiddlewareDeduplicate(t *testing.T) {
	t.Parallel()

	var callCount int

	mw := MiddlewareDeduplicate(100 * time.Millisecond)
	handler := mw(func(_ context.Context, _ Event) error {
		callCount++

		return nil
	})

	ctx := context.Background()
	event := testWriteEvent("/tmp/test.go")

	// Send same event multiple times rapidly
	for range 5 {
		_ = handler(ctx, event)
	}

	// Should only process once
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Different path should be processed
	otherEvent := testWriteEvent("/tmp/other.go")
	_ = handler(ctx, otherEvent)

	if callCount != 2 {
		t.Errorf("expected 2 calls after different path, got %d", callCount)
	}

	// Different operation should be processed
	createEvent := testEvent("/tmp/test.go", Create)
	_ = handler(ctx, createEvent)

	if callCount != 3 {
		t.Errorf("expected 3 calls after different op, got %d", callCount)
	}
}

func TestMiddlewareBatch_FullBatch(t *testing.T) {
	t.Parallel()

	var batched []Event

	flush := flushToSlice(&batched)

	mw := MiddlewareBatch(0, 3, flush) // use defaults for window, maxSize=3
	handler := mw(noopHandler())

	ctx := context.Background()

	// Send 3 events to fill the batch
	for i := range 3 {
		err := handler(ctx, testEvent("/tmp/file.txt", Write))
		if err != nil {
			t.Errorf("event %d: unexpected error: %v", i, err)
		}
	}

	assertBatchLen(t, batched, 3, "")
}

func TestMiddlewareBatch_FlushError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("flush error") //nolint:err113 // test-specific dynamic error

	flush := func(_ []Event) error {
		return testErr
	}

	mw := MiddlewareBatch(0, 1, flush) // maxSize=1 triggers immediate flush
	handler := mw(noopHandler())

	err := handler(context.Background(), testEvent("/tmp/file.txt", Write))
	assertErrorIs(t, err, testErr, "expected flush error")
}

func TestMiddlewareBatch_TimerFlush(t *testing.T) {
	t.Parallel()

	done := make(chan []Event, 1)

	flush := func(events []Event) error {
		done <- events

		return nil
	}

	mw := MiddlewareBatch(50*time.Millisecond, 100, flush) // short window, large maxSize
	handler := mw(noopHandler())

	ctx := context.Background()

	// Send 2 events (below maxSize, so timer is set)
	_ = handler(ctx, testEvent("/tmp/a.go", Write))
	_ = handler(ctx, testEvent("/tmp/b.go", Write))

	// Wait for timer to fire
	batched := waitForChannel(t, done, 2*time.Second, "timed out waiting for timer flush")
	assertBatchLen(t, batched, 2, " from timer flush")
}

func TestMiddlewareBatch_DefaultValues(t *testing.T) {
	t.Parallel()

	var batched []Event

	flush := flushToSlice(&batched)

	// Both window and maxSize are 0 — should use defaults
	mw := MiddlewareBatch(0, 0, flush)
	handler := mw(noopHandler())

	ctx := context.Background()
	_ = handler(ctx, testEvent("/tmp/file.txt", Write))

	// Event should pass through to next handler (not flushed yet, waiting for timer)
	// Just verify no panic or error
}

func TestMiddlewareChain(t *testing.T) {
	t.Parallel()

	var order []string

	record := func(name string) Middleware {
		return func(next Handler) Handler {
			return func(ctx context.Context, event Event) error {
				order = append(order, name)

				return next(ctx, event)
			}
		}
	}

	handler := record("first")(
		record("second")(
			record("third")(
				noopHandler(),
			),
		),
	)

	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))

	expected := []string{"first", "second", "third"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}

	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("position %d: expected %q, got %q", i, exp, order[i])
		}
	}
}

func BenchmarkMiddlewareLogging(b *testing.B) {
	logger := slog.New(slog.DiscardHandler)

	runMiddlewareBenchmark(b, func() Middleware { return MiddlewareLogging(logger) })
}

func BenchmarkMiddlewareRecovery(b *testing.B) {
	runMiddlewareBenchmark(b, MiddlewareRecovery)
}

func BenchmarkMiddlewareRateLimit(b *testing.B) {
	runMiddlewareBenchmark(b, func() Middleware { return MiddlewareRateLimit(0) })
}

func BenchmarkMiddlewareMetrics(b *testing.B) {
	runMiddlewareBenchmark(b, func() Middleware { return MiddlewareMetrics(func(_ Op) {}) })
}

func runMiddlewareBenchmark(b *testing.B, mwFunc func() Middleware) {
	b.Helper()

	handler := mwFunc()(noopHandler())
	event := Event{
		Op:        Write,
		Path:      benchmarkTestPathTestGo,
		Timestamp: time.Now(),
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}
	ctx := context.Background()

	b.ResetTimer()

	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}

func TestMiddlewareCircuitBreaker_Closed(t *testing.T) {
	t.Parallel()

	mw := MiddlewareCircuitBreaker(3, time.Second)
	handler := mw(noopHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected no error in closed circuit, got %v", err)
	}
}

func TestMiddlewareCircuitBreaker_OpensAfterFailures(t *testing.T) {
	t.Parallel()

	mw := MiddlewareCircuitBreaker(3, 100*time.Millisecond)
	handler := mw(errReturningHandler())

	// First 3 calls should pass through (and return errors)
	assertErrOnCalls(t, handler, 3, "expected error from handler")

	// Circuit should now be open - events are dropped
	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected nil when circuit is open, got %v", err)
	}
}

func TestMiddlewareCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	t.Parallel()

	mw := MiddlewareCircuitBreaker(2, 50*time.Millisecond)

	var callCount atomic.Int32

	failHandler := func(_ context.Context, _ Event) error {
		if callCount.Add(1) <= 2 {
			return errTest
		}

		return nil
	}

	handler := mw(failHandler)

	// Trigger failures to open circuit
	_ = handler(context.Background(), testWriteEvent("/test"))
	_ = handler(context.Background(), testWriteEvent("/test"))

	// Circuit is open - event dropped
	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected nil when circuit is open, got %v", err)
	}

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Circuit is now half-open - one event gets through
	err = handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected nil on recovery, got %v", err)
	}
}

func TestMiddlewareErrorRateLimit(t *testing.T) {
	t.Parallel()

	mw := MiddlewareErrorRateLimit(3, time.Second)

	var errCount atomic.Int32

	errHandler := func(_ context.Context, _ Event) error {
		errCount.Add(1)

		return errTest
	}

	handler := mw(errHandler)

	// First 2 errors pass through
	for range 2 {
		err := handler(context.Background(), testWriteEvent("/test"))
		if err == nil {
			t.Error("expected error to pass through")
		}
	}

	// Third error triggers rate limiting
	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Error("expected error to be suppressed after rate limit")
	}
}

func TestMiddlewareErrorRecovery(t *testing.T) {
	t.Parallel()

	strategy := func(_ Event, err error) error {
		return nil
	}

	mw := MiddlewareErrorRecovery(strategy)
	handler := mw(errReturningHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected recovery to suppress error, got %v", err)
	}
}

func TestMiddlewareErrorRecovery_NilStrategy(t *testing.T) {
	t.Parallel()

	mw := MiddlewareErrorRecovery(nil)
	handler := mw(errReturningHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Error("expected error to pass through with nil strategy")
	}
}

func TestMiddlewareErrorCorrelation(t *testing.T) {
	t.Parallel()

	counter := atomic.Int64{}

	mw := MiddlewareErrorCorrelation(func() string {
		return fmt.Sprintf("corr-%d", counter.Add(1))
	})
	handler := mw(errReturningHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error")
	}

	assertErrContains(t, err, "correlation-id=corr-1")
}

func TestMiddlewareErrorCorrelation_DefaultGenerator(t *testing.T) {
	t.Parallel()

	mw := MiddlewareErrorCorrelation(nil)
	handler := mw(errReturningHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error")
	}

	assertErrContains(t, err, "correlation-id=")
}

func TestMiddlewareErrorSanitization(t *testing.T) {
	t.Parallel()

	sanitize := func(msg string) string {
		return strings.ReplaceAll(msg, "/secret/", "/***REDACTED***/")
	}

	mw := MiddlewareErrorSanitization(sanitize)

	innerErr := errors.New("file changed at /secret/key.pem") //nolint:err113
	handler := mw(handlerReturning(innerErr))

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error")
	}

	sanitizedMsg := err.Error()
	if !strings.Contains(sanitizedMsg, "***REDACTED***") {
		t.Errorf("expected redacted path in error, got: %v", sanitizedMsg)
	}

	if !errors.Is(err, innerErr) {
		t.Errorf("expected errors.Is to match original error via %%w chain")
	}
}

func TestMiddlewareErrorBatch(t *testing.T) {
	t.Parallel()

	var collected atomic.Int32

	var mu sync.Mutex

	var batches [][]BatchError

	flush := func(errors []BatchError) {
		mu.Lock()

		batches = append(batches, errors)
		mu.Unlock()

		collected.Add(int32(len(errors))) //nolint:gosec
	}

	mw := MiddlewareErrorBatch(100*time.Millisecond, 3, flush)
	handler := mw(errReturningHandler())

	// Send 3 errors to trigger max size flush
	for i := range 3 {
		err := handler(context.Background(), testWriteEvent(fmt.Sprintf("/test-%d", i)))
		if err == nil {
			t.Errorf("call %d: expected error to pass through", i)
		}
	}

	// Wait for flush
	assertCount(t, &collected, 3)

	mu.Lock()
	batchCount := len(batches)
	mu.Unlock()

	if batchCount < 1 {
		t.Error("expected at least one batch flush")
	}
}

func TestCircuitState_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestMiddlewareErrorSanitization_Nil(t *testing.T) {
	t.Parallel()

	mw := MiddlewareErrorSanitization(nil)
	handler := mw(handlerReturning(errors.New("test error"))) //nolint:err113 // test-specific dynamic error

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error to pass through with nil sanitize")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestMiddlewareExponentialBackoff_DropsAfterFailures(t *testing.T) {
	t.Parallel()

	mw := MiddlewareExponentialBackoff(2, 50*time.Millisecond, 200*time.Millisecond)

	var innerCalls atomic.Int32

	errHandler := func(_ context.Context, _ Event) error {
		innerCalls.Add(1)

		return errTest
	}

	handler := mw(errHandler)

	// First 2 calls reach the inner handler (maxFailures=2)
	assertErrOnCalls(t, handler, 2, "expected error to pass through")

	// Subsequent calls should be dropped during the backoff window
	assertNilOnCalls(t, handler, 3, "expected error to be dropped during backoff")

	// The inner handler should only have been called twice
	if got := innerCalls.Load(); got != 2 {
		t.Errorf("inner handler called %d times, want 2", got)
	}
}

func TestMiddlewareExponentialBackoff_RecoversAfterBackoff(t *testing.T) {
	t.Parallel()

	mw := MiddlewareExponentialBackoff(2, 30*time.Millisecond, 100*time.Millisecond)

	var failNext atomic.Bool

	failNext.Store(true)

	handler := mw(func(_ context.Context, _ Event) error {
		if failNext.Load() {
			return errTest
		}

		return nil
	})

	// Trigger 2 failures to enter backoff
	_ = handler(context.Background(), testWriteEvent("/test"))
	_ = handler(context.Background(), testWriteEvent("/test"))

	// Now succeed and verify the backoff resets
	failNext.Store(false)

	// Wait for backoff window to expire
	time.Sleep(50 * time.Millisecond)

	// Should call inner handler again with success
	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected success after backoff recovery, got %v", err)
	}
}

// expectedMiddlewareDefaultConsts is the complete inventory of package-level
// default constants owned by middleware.go. Every entry must be both declared
// in middleware.go and referenced somewhere in the package. The test below
// guards against dead constants: if a middleware's defaulting is refactored
// away but the const is left behind, the test fails.
//
// Keep this list in sync with middleware.go. The drift check in the test will
// catch additions/removals of any "default"-family const automatically.
var expectedMiddlewareDefaultConsts = []string{
	"defaultDedupeWindow",
	"dedupeCleanupMultiplier",
	"defaultRateLimitWindow",
	"defaultSlidingWindowEvents",
	"defaultErrorRateLimit",
	"defaultBatchWindow",
	"defaultBatchSize",
	"defaultThrottleEvents",
	"circuitBreakerDefaultTimeout",
	"defaultMaxFailures",
	"defaultExponentialBackoffInitial",
	"defaultExponentialBackoffMax",
}

// TestMiddlewareDefaultConsts_AllUsed asserts every middleware default constant
// is actually referenced in the package. This is a compile-time-independent
// guard: a const that survives a refactor but loses its only call site becomes
// dead code, and this test catches it before linters (which only flag unused
// package-level consts intermittently).
func TestMiddlewareDefaultConsts_AllUsed(t *testing.T) {
	t.Parallel()

	thisDir := filepath.Dir(middlewareSourceFile(t))
	referenceCounts, declaredInMiddleware := scanPackageConsts(t, thisDir)

	// Drift check: every expected default const must be declared in middleware.go.
	for _, name := range expectedMiddlewareDefaultConsts {
		if !declaredInMiddleware[name] {
			t.Errorf("expected default const %q is not declared in middleware.go — "+
				"update expectedMiddlewareDefaultConsts", name)
		}
	}

	// Reverse drift check: any const declared in middleware.go whose name starts
	// with "default" or matches the *DefaultTimeout convention must appear in the
	// expected list. This catches new defaults added without updating the test.
	discoveredDefaults := discoverDefaultConsts(t, thisDir)
	for name := range discoveredDefaults {
		if !declaredInMiddleware[name] {
			continue
		}

		if !containsString(expectedMiddlewareDefaultConsts, name) {
			t.Errorf("default-family const %q is declared in middleware.go but missing "+
				"from expectedMiddlewareDefaultConsts — add it to keep the guard complete", name)
		}
	}

	// Core assertion: every expected default const is referenced beyond its
	// declaration (count >= 2: one declaration + at least one usage).
	for _, name := range expectedMiddlewareDefaultConsts {
		count := referenceCounts[name]
		if count < 2 {
			t.Errorf("default const %q is referenced %d time(s) in the package; "+
				"expected at least 2 (declaration + usage). If the defaulting logic "+
				"was removed, delete the const; otherwise rewire it.", name, count)
		}
	}
}

// discoverDefaultConsts returns the set of const names declared anywhere in the
// package whose name starts with "default" or contains "DefaultTimeout". These
// are the names we treat as "middleware default const" candidates for the drift
// check.
func discoverDefaultConsts(t *testing.T, dir string) map[string]bool {
	t.Helper()

	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing package dir %q: %v", dir, err)
	}

	defaults := make(map[string]bool)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.CONST {
					continue
				}

				for _, spec := range genDecl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					for _, nameIdent := range valueSpec.Names {
						nm := nameIdent.Name
						if strings.HasPrefix(nm, "default") || strings.Contains(nm, "DefaultTimeout") {
							defaults[nm] = true
						}
					}
				}
			}
		}
	}

	return defaults
}

// scanPackageConsts parses every Go file in dir and returns:
//   - referenceCounts: how many times each identifier name appears as an
//     *ast.Ident across the whole package (declaration sites count too).
//   - declaredInMiddleware: the set of const names declared in middleware.go.
func scanPackageConsts(t *testing.T, dir string) (map[string]int, map[string]bool) {
	t.Helper()

	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing package dir %q: %v", dir, err)
	}

	referenceCounts := make(map[string]int)
	declaredInMiddleware := make(map[string]bool)

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			isMiddleware := strings.HasSuffix(filename, "middleware.go")

			for _, decl := range file.Decls {
				if isMiddleware {
					recordConstNames(decl, declaredInMiddleware)
				}
			}

			ast.Inspect(file, func(n ast.Node) bool {
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}

				referenceCounts[ident.Name]++

				return true
			})
		}
	}

	return referenceCounts, declaredInMiddleware
}

func recordConstNames(decl ast.Decl, into map[string]bool) {
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok || genDecl.Tok != token.CONST {
		return
	}

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for _, nameIdent := range valueSpec.Names {
			into[nameIdent.Name] = true
		}
	}
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}

	return false
}

// middlewareSourceFile returns the path to middleware.go relative to the test
// binary's working directory. The test CWD is the package source directory.
func middlewareSourceFile(t *testing.T) string {
	t.Helper()

	const filename = "middleware.go"

	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("cannot stat %q from test CWD: %v", filename, err)
	}

	if info.IsDir() {
		t.Fatalf("%q is a directory, expected the middleware.go source file", filename)
	}

	return filename
}
