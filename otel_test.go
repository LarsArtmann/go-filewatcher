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

	middleware := OTelMiddleware(func(_, _ string) OTelSpan {
		return span
	})

	handler := middleware(noopHandler())

	err := handler(context.Background(), testWriteEvent("/tmp/file.go"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !span.ended.Load() {
		t.Error("expected span to be ended")
	}

	if span.status != "ok" {
		t.Errorf("expected status=ok, got %q", span.status)
	}

	if span.attrs["filewatcher.path"] != "/tmp/file.go" {
		t.Errorf("expected path attribute, got %q", span.attrs["filewatcher.path"])
	}

	if span.attrs["filewatcher.op"] != "WRITE" {
		t.Errorf("expected op attribute, got %q", span.attrs["filewatcher.op"])
	}
}

func TestOTelMiddleware_Error(t *testing.T) {
	t.Parallel()

	span := &testSpan{}

	middleware := OTelMiddleware(func(_, _ string) OTelSpan {
		return span
	})

	handler := middleware(func(_ context.Context, _ Event) error {
		return errTest
	})

	err := handler(context.Background(), testWriteEvent("/test"))
	if err == nil {
		t.Fatal("expected error")
	}

	if span.status != "error" {
		t.Errorf("expected status=error, got %q", span.status)
	}

	if span.attrs["filewatcher.error"] == "" {
		t.Error("expected filewatcher.error attribute")
	}
}

func TestOTelMiddleware_NilSpan(t *testing.T) {
	t.Parallel()

	// startSpan returns nil; middleware should pass through
	otelMw := OTelMiddleware(func(_, _ string) OTelSpan {
		return nil
	})

	called := false

	handler := otelMw(func(_ context.Context, _ Event) error {
		called = true

		return nil
	})

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

	otelMW := OTelMiddleware(func(_, _ string) OTelSpan {
		return span
	})

	customErr := errors.New("custom error") //nolint:err113

	handler := otelMW(func(_ context.Context, _ Event) error {
		return customErr
	})

	err := handler(context.Background(), testWriteEvent("/test"))
	if !errors.Is(err, customErr) {
		t.Errorf("expected error to be propagated, got %v", err)
	}
}
