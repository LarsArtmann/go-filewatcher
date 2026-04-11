//nolint:testpackage // Tests need internal access to unexported symbols
package filewatcher

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestMiddlewareLogging(t *testing.T) {
	t.Parallel()

	var called atomic.Int32

	mw := func(next Handler) Handler {
		return func(ctx context.Context, event Event) error {
			called.Add(1)

			return next(ctx, event)
		}
	}

	handler := mw(noopHandler())
	_ = handler(
		context.Background(),
		testWriteEvent("test.go"),
	)

	assertCount(t, &called, 1)
}

func TestMiddlewareRecovery(t *testing.T) {
	t.Parallel()

	recovery := MiddlewareRecovery()

	panicHandler := func(_ context.Context, _ Event) error {
		panic("test panic")
	}

	wrapped := recovery(panicHandler)

	err := wrapped(
		context.Background(),
		testWriteEvent("test.go"),
	)
	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
}

func TestMiddlewareRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	recovery := MiddlewareRecovery()

	wrapped := recovery(noopHandler())

	err := wrapped(
		context.Background(),
		testWriteEvent("test.go"),
	)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMiddlewareRateLimit(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	rateLimit := MiddlewareRateLimit(100 * time.Millisecond)

	handler := rateLimit(countHandler(&count))

	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))
	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))
	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))

	assertCount(t, &count, 1)

	time.Sleep(150 * time.Millisecond)

	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))

	assertCount(t, &count, 2)
}

func TestMiddlewareFilter(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	mw := MiddlewareFilter(FilterExtensions(".go"))

	handler := mw(countHandler(&count))

	_ = handler(context.Background(), testWriteEvent("test.txt"))
	_ = handler(context.Background(), testWriteEvent("test.go"))

	assertCount(t, &count, 1)
}

func TestMiddlewareOnError(t *testing.T) {
	t.Parallel()

	var (
		gotEvent Event
		gotErr   error
	)

	mw := MiddlewareOnError(func(event Event, err error) {
		gotEvent = event
		gotErr = err
	})

	errHandler := func(_ context.Context, _ Event) error {
		return context.DeadlineExceeded
	}

	handler := mw(errHandler)

	err := handler(context.Background(), testWriteEvent("test.go"))
	if err == nil {
		t.Fatal("expected error to propagate")
	}

	if gotEvent.Path != "test.go" {
		t.Errorf("expected error handler to receive event path, got %q", gotEvent.Path)
	}

	if gotErr == nil {
		t.Error("expected error handler to receive error")
	}
}

func TestMiddlewareMetrics(t *testing.T) {
	t.Parallel()

	metrics := make(map[Op]int)
	mw := MiddlewareMetrics(func(op Op) { metrics[op]++ })

	handler := mw(noopHandler())

	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))
	_ = handler(context.Background(), testEvent("/tmp/test.txt", Write))
	_ = handler(context.Background(), testEvent("/tmp/test.txt", Create))

	if metrics[Write] != 2 {
		t.Errorf("expected 2 Write metrics, got %d", metrics[Write])
	}

	if metrics[Create] != 1 {
		t.Errorf("expected 1 Create metric, got %d", metrics[Create])
	}
}

func TestMiddlewareLogging_NilLogger(t *testing.T) {
	t.Parallel()

	mw := MiddlewareLogging(nil)

	handler := mw(noopHandler())

	err := handler(context.Background(), testWriteEvent("test.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMiddlewareWriteFileLog(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/events.log"

	mw := MiddlewareWriteFileLog(tmpFile)

	handler := mw(noopHandler())

	ts := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	e := Event{Path: "/tmp/test.go", Op: Write, Timestamp: ts, IsDir: false}

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
