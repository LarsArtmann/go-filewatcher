//nolint:testpackage,varnamelen,exhaustruct,err113,goconst // Tests need internal access; idiomatic short names; partial initialization acceptable; dynamic errors for testing; repeated literals in tests
package filewatcher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestWatcherError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *WatcherError
		expected string
	}{
		{
			name: "with path",
			err: &WatcherError{
				Op:       OpString("watch"),
				Path:     "/test/path",
				Err:      ErrPathNotFound,
				Category: CategoryPermanent,
			},
			expected: "watch: path \"/test/path\": path not found",
		},
		{
			name: "without path",
			err: &WatcherError{
				Op:       OpString("init"),
				Err:      ErrNoPaths,
				Category: CategoryPermanent,
			},
			expected: "init: at least one path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWatcherError_Unwrap(t *testing.T) {
	t.Parallel()

	err := &WatcherError{
		Op:       OpString("test"),
		Err:      ErrWatcherClosed,
		Category: CategoryPermanent,
	}

	if !errors.Is(err, ErrWatcherClosed) {
		t.Error("expected errors.Is to match wrapped error")
	}
}

func TestWatcherError_IsTransient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		category  ErrorCategory
		transient bool
	}{
		{"transient", CategoryTransient, true},
		{"permanent", CategoryPermanent, false},
		{"unknown", CategoryUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &WatcherError{
				Op:       OpString("test"),
				Err:      errors.New("test"),
				Category: tt.category,
			}

			if got := err.IsTransient(); got != tt.transient {
				t.Errorf("IsTransient() = %v, want %v", got, tt.transient)
			}
		})
	}
}

func TestWatcherError_IsPermanent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		category  ErrorCategory
		permanent bool
	}{
		{"transient", CategoryTransient, false},
		{"permanent", CategoryPermanent, true},
		{"unknown", CategoryUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &WatcherError{
				Op:       OpString("test"),
				Err:      errors.New("test"),
				Category: tt.category,
			}

			if got := err.IsPermanent(); got != tt.permanent {
				t.Errorf("IsPermanent() = %v, want %v", got, tt.permanent)
			}
		})
	}
}

func TestNewWatcherError(t *testing.T) {
	t.Parallel()

	err := NewWatcherError(OpString("test_op"), "/test/path", ErrPathNotFound)

	if err.Op != OpString("test_op") {
		t.Errorf("Op = %q, want %q", err.Op, OpString("test_op"))
	}

	if err.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", err.Path, "/test/path")
	}

	if !errors.Is(err.Err, ErrPathNotFound) {
		t.Error("expected Err to wrap ErrPathNotFound")
	}

	if err.Category != CategoryPermanent {
		t.Errorf("Category = %v, want CategoryPermanent", err.Category)
	}
}

func TestCategorizeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{"nil", nil, CategoryUnknown},
		{"watcher closed", ErrWatcherClosed, CategoryPermanent},
		{"no paths", ErrNoPaths, CategoryPermanent},
		{"path not found", ErrPathNotFound, CategoryPermanent},
		{"path not dir", ErrPathNotDir, CategoryPermanent},
		{"watcher running", ErrWatcherRunning, CategoryPermanent},
		{"unknown op", ErrUnknownOp, CategoryPermanent},
		{"fsnotify failed", ErrFsnotifyFailed, CategoryTransient},
		{"walk failed", ErrWalkFailed, CategoryTransient},
		{"event processing failed", ErrEventProcessingFailed, CategoryTransient},
		{"wrapped permanent", fmt.Errorf("wrapped: %w", ErrWatcherClosed), CategoryPermanent},
		{"wrapped transient", fmt.Errorf("wrapped: %w", ErrFsnotifyFailed), CategoryTransient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := categorizeError(tt.err)
			if got != tt.expected {
				t.Errorf("categorizeError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsTransientError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		transient bool
	}{
		{"nil", nil, false},
		{"transient sentinel", ErrFsnotifyFailed, true},
		{"permanent sentinel", ErrWatcherClosed, false},
		{
			"transient watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryTransient},
			true,
		},
		{
			"permanent watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryPermanent},
			false,
		},
		{
			"unknown watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryUnknown},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsTransientError(tt.err); got != tt.transient {
				t.Errorf("IsTransientError() = %v, want %v", got, tt.transient)
			}
		})
	}
}

func TestIsPermanentError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		permanent bool
	}{
		{"nil", nil, false},
		{"permanent sentinel", ErrWatcherClosed, true},
		{"transient sentinel", ErrFsnotifyFailed, false},
		{
			"permanent watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryPermanent},
			true,
		},
		{
			"transient watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryTransient},
			false,
		},
		{
			"unknown watcher error",
			&WatcherError{Err: errors.New("test"), Category: CategoryUnknown},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsPermanentError(tt.err); got != tt.permanent {
				t.Errorf("IsPermanentError() = %v, want %v", got, tt.permanent)
			}
		})
	}
}

func TestErrorContext(t *testing.T) {
	t.Parallel()

	event := &Event{Path: "/test/file.go", Op: Create}
	ctx := ErrorContext{
		Operation: "test_op",
		Path:      "/test/path",
		Event:     event,
		Retryable: true,
	}

	if ctx.Operation != "test_op" {
		t.Errorf("Operation = %q, want %q", ctx.Operation, "test_op")
	}

	if ctx.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", ctx.Path, "/test/path")
	}

	if ctx.Event != event {
		t.Error("Event pointer mismatch")
	}

	if !ctx.Retryable {
		t.Error("expected Retryable to be true")
	}
}

func TestErrorHandler_WithContext(t *testing.T) {
	t.Parallel()

	var (
		receivedCtx ErrorContext
		receivedErr error
	)

	handler := func(ctx ErrorContext, err error) {
		receivedCtx = ctx
		receivedErr = err
	}

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithErrorHandler(handler))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	testErr := errors.New("test error") //nolint:err113

	w.handleError(ErrorContext{Operation: "test", Path: "/test/path", Retryable: false}, testErr)

	if receivedCtx.Operation != "test" {
		t.Errorf("Operation = %q, want %q", receivedCtx.Operation, "test")
	}

	if receivedCtx.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", receivedCtx.Path, "/test/path")
	}

	if !errors.Is(receivedErr, testErr) {
		t.Error("Error mismatch")
	}
}

//nolint:paralleltest // Not parallel: captures os.Stderr, which is a global resource.
func TestErrorHandler_DefaultLogsToStderr(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	old := os.Stderr
	r, wPipe, _ := os.Pipe()
	os.Stderr = wPipe

	w.handleError(ErrorContext{Operation: "test"}, errors.New("test error")) //nolint:err113

	_ = wPipe.Close()
	os.Stderr = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("expected error message in stderr, got %q", output)
	}
}

//nolint:paralleltest // Not parallel: captures os.Stderr, which is a global resource.
func TestErrorHandler_DefaultWithoutPath(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	old := os.Stderr
	r, wPipe, _ := os.Pipe()
	os.Stderr = wPipe

	w.handleError(ErrorContext{Operation: "fsnotify"}, errors.New("fsnotify error")) //nolint:err113

	_ = wPipe.Close()
	os.Stderr = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "fsnotify error") {
		t.Errorf("expected error message in stderr, got %q", output)
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	errors := []struct {
		err  error
		name string
	}{
		{ErrWatcherClosed, "ErrWatcherClosed"},
		{ErrNoPaths, "ErrNoPaths"},
		{ErrPathNotFound, "ErrPathNotFound"},
		{ErrPathNotDir, "ErrPathNotDir"},
		{ErrWatcherRunning, "ErrWatcherRunning"},
		{ErrUnknownOp, "ErrUnknownOp"},
		{ErrFsnotifyFailed, "ErrFsnotifyFailed"},
		{ErrWalkFailed, "ErrWalkFailed"},
		{ErrEventProcessingFailed, "ErrEventProcessingFailed"},
		{ErrMiddlewareFailed, "ErrMiddlewareFailed"},
	}

	for _, e := range errors {
		t.Run(e.name, func(t *testing.T) {
			t.Parallel()

			if e.err == nil {
				t.Error("expected non-nil error")
			}

			if e.err.Error() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestIsTransientError_Nil(t *testing.T) {
	t.Parallel()

	if IsTransientError(nil) {
		t.Error("expected IsTransientError(nil) to be false")
	}
}

func TestIsPermanentError_Nil(t *testing.T) {
	t.Parallel()

	if IsPermanentError(nil) {
		t.Error("expected IsPermanentError(nil) to be false")
	}
}

func TestCategorizeError_Nil(t *testing.T) {
	t.Parallel()

	if categorizeError(nil) != CategoryUnknown {
		t.Error("expected categorizeError(nil) to return CategoryUnknown")
	}
}

//nolint:paralleltest // Not parallel: captures error handler state.
func TestErrorHandler_Async(t *testing.T) {
	var callCount atomic.Int32

	handler := func(ctx ErrorContext, err error) {
		_ = ctx
		_ = err

		callCount.Add(1)
	}

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithErrorHandler(handler))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	// Simulate multiple concurrent errors
	for range 10 {
		go w.errorHandler(ErrorContext{Operation: "test"}, errors.New("error")) //nolint:err113
	}

	// Wait for handlers to complete (in production they'd be async)
	// This is a best-effort test for thread safety
}
