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

	handler := mw(func(_ context.Context, _ Event) error { return nil })
	_ = handler(
		context.Background(),
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)

	if got := called.Load(); got != 1 {
		t.Errorf("expected middleware to be called, got %d", got)
	}
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
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
}

func TestMiddlewareRecovery_NoPanic(t *testing.T) {
	t.Parallel()

	recovery := MiddlewareRecovery()

	normalHandler := func(_ context.Context, _ Event) error {
		return nil
	}

	wrapped := recovery(normalHandler)

	err := wrapped(
		context.Background(),
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMiddlewareRateLimit(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	rateLimit := MiddlewareRateLimit(100 * time.Millisecond)

	handler := rateLimit(func(_ context.Context, _ Event) error {
		count.Add(1)
		return nil
	})

	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call due to rate limiting, got %d", got)
	}

	time.Sleep(150 * time.Millisecond)

	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)

	if got := count.Load(); got != 2 {
		t.Errorf("expected 2 calls after rate limit window, got %d", got)
	}
}

func TestMiddlewareFilter(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	mw := MiddlewareFilter(FilterExtensions(".go"))

	handler := mw(func(_ context.Context, _ Event) error {
		count.Add(1)
		return nil
	})

	_ = handler(
		context.Background(),
		Event{Path: "test.txt", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call (only .go file), got %d", got)
	}
}

func TestMiddlewareOnError(t *testing.T) {
	t.Parallel()

	var gotEvent Event
	var gotErr error

	mw := MiddlewareOnError(func(event Event, err error) {
		gotEvent = event
		gotErr = err
	})

	errHandler := func(_ context.Context, _ Event) error {
		return context.DeadlineExceeded
	}

	handler := mw(errHandler)

	err := handler(
		context.Background(),
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
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

	successHandler := func(_ context.Context, _ Event) error { return nil }
	handler := mw(successHandler)

	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Op: Create, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)

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

	handler := mw(func(_ context.Context, _ Event) error { return nil })

	err := handler(
		context.Background(),
		Event{Path: "test.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMiddlewareWriteFileLog(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/events.log"

	mw := MiddlewareWriteFileLog(tmpFile)

	handler := mw(func(_ context.Context, _ Event) error { return nil })

	ts := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	err := handler(
		context.Background(),
		Event{Path: "/tmp/test.go", Op: Write, Timestamp: ts, IsDir: false},
	)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(tmpFile) //nolint:gosec // test file from TempDir
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "WRITE") {
		t.Errorf("expected log to contain WRITE, got %q", content)
	}
	if !strings.Contains(content, "/tmp/test.go") {
		t.Errorf("expected log to contain file path, got %q", content)
	}
}

func TestMiddlewareWriteFileLog_Appends(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/events.log"

	mw := MiddlewareWriteFileLog(tmpFile)
	handler := mw(func(_ context.Context, _ Event) error { return nil })

	_ = handler(
		context.Background(),
		Event{Path: "a.go", Op: Write, Timestamp: time.Now(), IsDir: false},
	)
	_ = handler(
		context.Background(),
		Event{Path: "b.go", Op: Create, Timestamp: time.Now(), IsDir: false},
	)

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
				func(_ context.Context, _ Event) error { return nil },
			),
		),
	)

	_ = handler(
		context.Background(),
		Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now(), IsDir: false},
	)

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
	mw := MiddlewareLogging(logger)
	handler := mw(func(_ context.Context, _ Event) error { return nil })
	event := Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}

func BenchmarkMiddlewareRecovery(b *testing.B) {
	mw := MiddlewareRecovery()
	handler := mw(func(_ context.Context, _ Event) error { return nil })
	event := Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}

func BenchmarkMiddlewareRateLimit(b *testing.B) {
	mw := MiddlewareRateLimit(0) // no limit
	handler := mw(func(_ context.Context, _ Event) error { return nil })
	event := Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}

func BenchmarkMiddlewareMetrics(b *testing.B) {
	mw := MiddlewareMetrics(func(_ Op) {})
	handler := mw(func(_ context.Context, _ Event) error { return nil })
	event := Event{Op: Write, Path: "/tmp/test.go", Timestamp: time.Now(), IsDir: false}
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		_ = handler(ctx, event)
		_ = i
	}
}
