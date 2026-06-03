package filewatcher

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

type testSpan struct {
	ended     atomic.Bool
	status    string
	desc      string
	attrs     map[string]string
	path      string
	op        string
	startFunc func(path, op string) OTelSpan
}

func (s *testSpan) End() {
	s.ended.Store(true)
}

func (s *testSpan) SetStatus(code, desc string) {
	s.status = code
	s.desc = desc
}

func (s *testSpan) SetAttributes(attrs ...Attribute) {
	if s.attrs == nil {
		s.attrs = make(map[string]string)
	}

	for _, attr := range attrs {
		s.attrs[attr.Key] = attr.Value
	}
}

// assertSpanStatus fails the test if span.status != want.
func assertSpanStatus(t *testing.T, span *testSpan, want string) {
	t.Helper()

	if span.status != want {
		t.Errorf("expected status=%s, got %q", want, span.status)
	}
}

// assertSpanAttr fails the test if span.attrs[key] != want.
func assertSpanAttr(t *testing.T, span *testSpan, key, want string) {
	t.Helper()

	if got := span.attrs[key]; got != want {
		t.Errorf("expected %s=%q, got %q", key, want, got)
	}
}

func TestOTelMiddleware_NilStartFunc(t *testing.T) {
	t.Parallel()

	mw := OTelMiddleware(nil)

	var called atomic.Int32

	handler := mw(func(_ context.Context, _ Event) error {
		called.Add(1)

		return nil
	})

	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if called.Load() != 1 {
		t.Errorf("expected handler called once, got %d", called.Load())
	}
}

func TestOTelMiddleware_Success(t *testing.T) {
	t.Parallel()

	span := &testSpan{}

	middleware := OTelMiddleware(fixedOTelStart(span))

	handler := middleware(noopHandler())

	err := handler(context.Background(), testWriteEvent("/tmp/file.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !span.ended.Load() {
		t.Error("expected span to be ended")
	}

	assertSpanStatus(t, span, "ok")

	assertSpanAttr(t, span, "filewatcher.path", "/tmp/file.go")
	assertSpanAttr(t, span, "filewatcher.op", "WRITE")
}

func TestOTelMiddleware_Error(t *testing.T) {
	t.Parallel()

	span := &testSpan{}

	middleware := OTelMiddleware(fixedOTelStart(span))

	handler := middleware(errReturningHandler())

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error")
	}

	assertSpanStatus(t, span, "error")

	if span.attrs["filewatcher.error"] == "" {
		t.Error("expected filewatcher.error attribute")
	}
}

func TestOTelMiddleware_NilSpan(t *testing.T) {
	t.Parallel()

	// startSpan returns nil; middleware should pass through
	otelMw := OTelMiddleware(fixedOTelStart(nil))

	called := false

	handler := otelMw(calledFlagHandler(&called))

	err := handler(context.Background(), testWriteEvent("/test"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !called {
		t.Error("expected handler to be called even when startSpan returns nil")
	}
}

func TestOTelMiddleware_PropagatesError(t *testing.T) {
	t.Parallel()

	span := &testSpan{}

	otelMW := OTelMiddleware(fixedOTelStart(span))

	customErr := errors.New("custom error") //nolint:err113

	handler := otelMW(handlerReturning(customErr))

	err := handler(context.Background(), testWriteEvent("/test"))
	if !errors.Is(err, customErr) {
		t.Errorf("expected error to be propagated, got %v", err)
	}
}
