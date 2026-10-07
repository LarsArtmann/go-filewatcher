//nolint:varnamelen,err113 // idiomatic short names; dynamic errors for testing
package filewatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// --- WithContentHashMaxSize(0): semantics pin ---------------------------------
//
// Documented contract (options.go): WithContentHashMaxSize(0) disables content
// hashing entirely. These tests pin that behavior plus the option-order trap
// where MaxSize(0) before WithContentHashing() is re-substituted with the
// default, so a future refactor cannot silently change either.

func TestContentHashMaxSize_ZeroDisablesHashing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	target := filepath.Join(tmpDir, "watched.txt")
	if err := os.WriteFile(target, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}

	w := newTestWatcher(t, tmpDir, WithContentHashing(), WithContentHashMaxSize(0))

	if w.contentHashMaxSize != 0 {
		t.Fatalf("contentHashMaxSize = %d, want 0 (disabled)", w.contentHashMaxSize)
	}

	got := convertEvent(fsnotify.Event{Name: target, Op: fsnotify.Create}, false, w.contentHashMaxSize)
	if got == nil {
		t.Fatal("convertEvent returned nil for Create")
	}

	if got.Hash != "" {
		t.Fatalf("Hash = %q, want empty when hashing is disabled", got.Hash)
	}
}

func TestContentHashMaxSize_DefaultAppliesWhenHashingEnabled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir, WithContentHashing())

	if w.contentHashMaxSize != defaultContentHashMaxSize {
		t.Fatalf("contentHashMaxSize = %d, want default %d", w.contentHashMaxSize, defaultContentHashMaxSize)
	}
}

func TestContentHashMaxSize_PositiveLimitHashesSmallFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	target := filepath.Join(tmpDir, "small.txt")
	if err := os.WriteFile(target, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}

	w := newTestWatcher(t, tmpDir, WithContentHashing(), WithContentHashMaxSize(1024))

	got := convertEvent(fsnotify.Event{Name: target, Op: fsnotify.Write}, false, w.contentHashMaxSize)
	if got == nil {
		t.Fatal("convertEvent returned nil for Write")
	}

	if len(got.Hash) != 64 {
		t.Fatalf("Hash length = %d, want 64 hex chars (sha256)", len(got.Hash))
	}

	if strings.ToUpper(got.Hash) == got.Hash {
		t.Fatal("Hash must be lowercase hex")
	}
}

func TestContentHashMaxSize_OrderTrapZeroBeforeHashing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Pins CURRENT behavior: WithContentHashing() re-substitutes the default
	// when it sees a zero size, so declaring MaxSize(0) first re-enables
	// hashing. Supported order is WithContentHashing() then MaxSize(0).
	w := newTestWatcher(t, tmpDir, WithContentHashMaxSize(0), WithContentHashing())

	if w.contentHashMaxSize != defaultContentHashMaxSize {
		t.Fatalf(
			"contentHashMaxSize = %d, want default %d (order trap: MaxSize(0) first is overridden)",
			w.contentHashMaxSize,
			defaultContentHashMaxSize,
		)
	}
}

// --- Watch budget @ fraction 1.0 ----------------------------------------------
//
// Contract: WithMaxWatchesSafetyFraction(1.0) (the default) must leave the
// effective cap identical to the detected limit — the fraction exists to give
// headroom, never to shrink the default budget.

func TestApplyMaxWatchesFraction_OneMeansCapEqualsDetected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	w.mu.Lock()
	defer w.mu.Unlock()

	w.maxWatches = 1000
	w.maxWatchesFraction = 1.0
	w.applyMaxWatchesFraction()

	if w.maxWatches != 1000 {
		t.Fatalf("maxWatches = %d, want 1000 (fraction 1.0 is a no-op)", w.maxWatches)
	}
}

func TestApplyMaxWatchesFraction_ReducesBelowOne(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	w.mu.Lock()
	defer w.mu.Unlock()

	w.maxWatches = 1000
	w.maxWatchesFraction = 0.75
	w.applyMaxWatchesFraction()

	if w.maxWatches != 750 {
		t.Fatalf("maxWatches = %d, want 750 (1000 * 0.75)", w.maxWatches)
	}
}

func TestApplyMaxWatchesFraction_ZeroIsNoOp(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	w.mu.Lock()
	defer w.mu.Unlock()

	w.maxWatches = 1000
	w.maxWatchesFraction = 0
	w.applyMaxWatchesFraction()

	if w.maxWatches != 1000 {
		t.Fatalf("maxWatches = %d, want 1000 (fraction 0 is a no-op, current behavior)", w.maxWatches)
	}
}

func TestMaxWatchesFraction_DefaultWatcherCapEqualsDetected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w := newTestWatcher(t, tmpDir)

	w.mu.Lock()
	detected := w.maxWatchesDetected
	effectiveCap := w.maxWatches
	fraction := w.maxWatchesFraction
	w.mu.Unlock()

	if fraction != 1.0 {
		t.Fatalf("default fraction = %v, want 1.0", fraction)
	}

	if detected > 0 && effectiveCap != detected {
		t.Fatalf("cap = %d, want == detected %d when fraction is 1.0", effectiveCap, detected)
	}
}

// --- ErrorContext.Event population --------------------------------------------
//
// Contract: contexts reaching the user's ErrorHandler carry the event being
// processed when one exists (handler and middleware errors), and carry
// Event == nil with a populated Path when no event is involved (fsnotify and
// add-path failures).

func TestMiddlewareErrorContextCarriesEvent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	handler, gotCtx, gotErr := newErrorHandlerCallback()

	failingMW := func(next Handler) Handler {
		return func(_ context.Context, _ Event) error {
			return errors.New("boom")
		}
	}

	w := newTestWatcher(t, tmpDir, WithErrorHandler(handler), WithMiddleware(failingMW))

	event := Event{
		Path:      filepath.Join(tmpDir, "triggered.go"),
		Op:        Write,
		Timestamp: time.Now(),
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
		Hash:      "",
	}

	wrapped := w.wrapWithMiddleware(func(_ context.Context, _ Event) {}, failingMW)
	wrapped(context.Background(), event)

	if *gotErr == nil {
		t.Fatal("error handler was not invoked")
	}

	if gotCtx.Operation != "middleware" {
		t.Fatalf("Operation = %q, want %q", gotCtx.Operation, "middleware")
	}

	if gotCtx.Event == nil {
		t.Fatal("Event = nil, want the event being processed")
	}

	if gotCtx.Event.Path != event.Path {
		t.Fatalf("Event.Path = %q, want %q", gotCtx.Event.Path, event.Path)
	}

	if gotCtx.Path != event.Path {
		t.Fatalf("Path = %q, want %q", gotCtx.Path, event.Path)
	}

	if gotCtx.Retryable {
		t.Fatal("Retryable = true, want false for middleware errors")
	}
}

func TestErrorHandlerContract_EventBearingContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	handler, gotCtx, gotErr := newErrorHandlerCallback()
	w := newTestWatcher(t, tmpDir, WithErrorHandler(handler))

	event := Event{
		Path:      filepath.Join(tmpDir, "x.go"),
		Op:        Create,
		Timestamp: time.Now(),
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
		Hash:      "",
	}

	w.handleError(ErrorContext{
		Operation: "handler",
		Path:      event.Path,
		Event:     &event,
		Retryable: false,
	}, errors.New("handler failed"))

	if *gotErr == nil {
		t.Fatal("error handler was not invoked")
	}

	if gotCtx.Event == nil || gotCtx.Event.Path != event.Path {
		t.Fatalf("Event = %v, want event with path %q", gotCtx.Event, event.Path)
	}
}

func TestErrorHandlerContract_NilEventContextKeepsPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	handler, gotCtx, gotErr := newErrorHandlerCallback()
	w := newTestWatcher(t, tmpDir, WithErrorHandler(handler))

	path := filepath.Join(tmpDir, "some", "dir")
	w.handleError(ErrorContext{
		Operation: opAddPath,
		Path:      path,
		Event:     nil,
		Retryable: true,
	}, errors.New("watch registration failed"))

	if *gotErr == nil {
		t.Fatal("error handler was not invoked")
	}

	if gotCtx.Event != nil {
		t.Fatalf("Event = %v, want nil when no event is involved", gotCtx.Event)
	}

	if gotCtx.Path != path {
		t.Fatalf("Path = %q, want %q", gotCtx.Path, path)
	}

	if !gotCtx.Retryable {
		t.Fatal("Retryable = false, want true for add-path errors")
	}
}
