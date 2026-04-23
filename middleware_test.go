package filewatcher

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

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
	assertLogContains(t, content, LogSubstring("filewatcher event"))
	assertLogContains(t, content, LogSubstring("WRITE"))
	assertLogContains(t, content, LogSubstring("/tmp/test.go"))
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
	if !strings.Contains(err.Error(), "panic in handler") {
		t.Errorf("expected error to contain 'panic in handler', got: %v", err)
	}
}

func TestMiddlewareRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	var called bool
	normalHandler := func(_ context.Context, _ Event) error {
		called = true

		return nil
	}

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

	var capturedEvent *Event
	var capturedErr error

	onError := func(event Event, err error) {
		capturedEvent = &event
		capturedErr = err
	}

	mw := MiddlewareOnError(onError)
	testErr := errors.New("test error")
	errorHandler := func(_ context.Context, _ Event) error {
		return testErr
	}

	testHandler := mw(errorHandler)

	err := testHandler(context.Background(), testWriteEvent("/tmp/test.go"))

	if err != testErr {
		t.Errorf("expected error to be passed through, got %v", err)
	}
	if capturedEvent == nil {
		t.Error("expected onError to be called with event")
	}
	if capturedErr != testErr {
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
	errorHandler := func(_ context.Context, _ Event) error {
		return errors.New("test error")
	}
	testHandler := mw(errorHandler)

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
	assertLogContains(t, content, LogSubstring("WRITE"))
	assertLogContains(t, content, LogSubstring("/tmp/test.go"))
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

	flush := func(events []Event) error {
		batched = append(batched, events...)

		return nil
	}

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

	if len(batched) != 3 {
		t.Errorf("expected 3 batched events, got %d", len(batched))
	}
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
	if !errors.Is(err, testErr) {
		t.Errorf("expected flush error, got %v", err)
	}
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
	select {
	case batched := <-done:
		if len(batched) != 2 {
			t.Errorf("expected 2 batched events from timer flush, got %d", len(batched))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for timer flush")
	}
}

func TestMiddlewareBatch_DefaultValues(t *testing.T) {
	t.Parallel()

	var batched []Event

	flush := func(events []Event) error {
		batched = append(batched, events...)

		return nil
	}

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
	event := Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
	ctx := context.Background()

	b.ResetTimer()

	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}
