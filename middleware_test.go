package filewatcher

import (
	"context"
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
	_ = handler(context.Background(), Event{Path: "test.go", Op: Write})

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

	err := wrapped(context.Background(), Event{Path: "test.go", Op: Write})
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

	err := wrapped(context.Background(), Event{Path: "test.go", Op: Write})
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

	_ = handler(context.Background(), Event{Op: Write})
	_ = handler(context.Background(), Event{Op: Write})
	_ = handler(context.Background(), Event{Op: Write})

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 call due to rate limiting, got %d", got)
	}

	time.Sleep(150 * time.Millisecond)

	_ = handler(context.Background(), Event{Op: Write})

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

	_ = handler(context.Background(), Event{Path: "test.txt", Op: Write})
	_ = handler(context.Background(), Event{Path: "test.go", Op: Write})

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

	err := handler(context.Background(), Event{Path: "test.go", Op: Write})
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

	_ = handler(context.Background(), Event{Op: Write})
	_ = handler(context.Background(), Event{Op: Write})
	_ = handler(context.Background(), Event{Op: Create})

	if metrics[Write] != 2 {
		t.Errorf("expected 2 Write metrics, got %d", metrics[Write])
	}
	if metrics[Create] != 1 {
		t.Errorf("expected 1 Create metric, got %d", metrics[Create])
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

	_ = handler(context.Background(), Event{Op: Write})

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
