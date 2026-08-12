package filewatcher

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRequireNonNegativeDuration(t *testing.T) {
	t.Parallel()

	t.Run("negative duration panics with option name", func(t *testing.T) {
		t.Parallel()

		const optionName = "WithDebounce"

		var (
			recovered any
			didPanic  bool
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					didPanic = true
					recovered = r
				}
			}()

			requireNonNegativeDuration(optionName, -1*time.Millisecond)
		}()

		if !didPanic {
			t.Fatal("expected panic for negative duration, got none")
		}

		msg, ok := recovered.(string)
		if !ok {
			t.Fatalf("expected panic value of type string, got %T", recovered)
		}

		if !strings.Contains(msg, optionName) {
			t.Errorf("expected panic message to contain %q, got %q", optionName, msg)
		}
	})

	t.Run("zero and positive durations do not panic", func(t *testing.T) {
		t.Parallel()

		requireNonNegativeDuration("WithDebounce", 0)
		requireNonNegativeDuration("WithPerPathDebounce", 5*time.Second)
	})
}

func TestWithIgnoreHidden(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithIgnoreHidden())

	if len(watcher.filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(watcher.filters))
	}

	hidden := testWriteEvent(".hidden_file")
	if watcher.filters[0](hidden) {
		t.Error("expected hidden file to be filtered out")
	}

	visible := testWriteEvent("visible_file")
	if !watcher.filters[0](visible) {
		t.Error("expected visible file to pass filter")
	}
}

func TestWithOnAdd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var addedPaths []string

	watcher := newTestWatcher(t, dir, WithOnAdd(func(path string) {
		addedPaths = append(addedPaths, path)
	}))

	if watcher.onAdd == nil {
		t.Fatal("expected onAdd callback to be set")
	}

	watcher.onAdd(dir)

	if len(addedPaths) != 1 || addedPaths[0] != dir {
		t.Errorf("expected callback to receive %q, got %v", dir, addedPaths)
	}
}

func TestWithOnError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var receivedErr error

	watcher := newTestWatcher(t, dir, WithOnError(func(err error) {
		receivedErr = err
	}))

	if watcher.errorHandler == nil {
		t.Fatal("expected errorHandler to be set")
	}

	testErr := errors.New("test error") //nolint:err113 // test-specific dynamic error
	watcher.errorHandler(
		ErrorContext{
			Operation: "test operation",
			Path:      "test path",
		},
		testErr,
	)

	if !errors.Is(receivedErr, testErr) {
		t.Errorf("expected callback to receive test error, got %v", receivedErr)
	}
}

func TestWithLazyIsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithLazyIsDir())

	if !watcher.lazyIsDir {
		t.Error("expected lazyIsDir to be true")
	}
}

func TestWithPollInterval(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithPollInterval(5*time.Second))

	assertEqual(t, "pollInterval", watcher.pollInterval, 5*time.Second)
}

func TestWithPolling(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithPolling(true))

	if !watcher.polling {
		t.Error("expected polling to be true")
	}

	assertEqual(t, "default pollInterval", watcher.pollInterval, 2*time.Second)
}

func TestWithPolling_False(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithPolling(false))

	if watcher.polling {
		t.Error("expected polling to be false")
	}
}

func TestWithDebug(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithDebug(nil))

	if !watcher.debug {
		t.Error("expected debug to be true")
	}

	// WithDebug(nil) must use slog.Default(), not store nil.
	if watcher.debugLogger == nil {
		t.Fatal("expected debugLogger to be slog.Default() when nil is passed, got nil")
	}

	// Calling debugLog must not panic.
	didPanic := true

	func() {
		defer func() { _ = recover() }()

		watcher.debugLog("test debug message", "key", "value")

		didPanic = false
	}()

	if didPanic {
		t.Fatal("expected debugLog not to panic when logger was nil at construction")
	}
}

func TestWithWatchedIgnoreDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher := newTestWatcher(t, dir, WithWatchedIgnoreDirs("node_modules", ".cache"))

	filterCount := len(watcher.filters)
	if filterCount == 0 {
		t.Error("expected at least one filter to be added")
	}
}

func TestWithMaxWatchesSafetyFraction_Clamping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"valid 0.75", 0.75, 0.75},
		{"valid 0.5", 0.5, 0.5},
		{"valid 1.0", 1.0, 1.0},
		{"zero clamped to 1.0", 0, 1.0},
		{"negative clamped to 1.0", -0.5, 1.0},
		{"above 1.0 clamped to 1.0", 1.5, 1.0},
	}

	for _, tc := range tests { //nolint:varnamelen // tc is conventional
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			watcher := newTestWatcher(t, dir, WithMaxWatchesSafetyFraction(tc.input))

			if watcher.maxWatchesFraction != tc.want {
				t.Errorf("maxWatchesFraction = %v, want %v", watcher.maxWatchesFraction, tc.want)
			}
		})
	}
}

func TestApplyMaxWatchesFraction_ReducesAutoDetectedLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	watcher := newTestWatcher(t, dir, WithMaxWatchesSafetyFraction(0.75))

	// Simulate a known auto-detected limit by setting it directly,
	// then applying the fraction.
	const simulatedLimit = 10000

	watcher.maxWatches = simulatedLimit
	watcher.maxWatchesExplicit = false
	watcher.applyMaxWatchesFraction()

	expected := int(float64(simulatedLimit) * 0.75) // 7500
	if watcher.maxWatches != expected {
		t.Errorf("maxWatches = %d, want %d (after 0.75 fraction of %d)",
			watcher.maxWatches, expected, simulatedLimit)
	}
}

func TestApplyMaxWatchesFraction_DoesNotAffectExplicitLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	const explicitLimit = 1000

	watcher := newTestWatcher(t, dir,
		WithMaxWatches(explicitLimit),
		WithMaxWatchesSafetyFraction(0.75),
	)

	if watcher.maxWatches != explicitLimit {
		t.Errorf("explicit maxWatches = %d, want %d (fraction must not apply to explicit limits)",
			watcher.maxWatches, explicitLimit)
	}

	if !watcher.maxWatchesExplicit {
		t.Error("maxWatchesExplicit should be true when WithMaxWatches(n>0) is used")
	}
}

func TestWithContentHashMaxSize(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New([]string{tmpDir}, WithContentHashMaxSize(100))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	if watcher.contentHashMaxSize != 100 {
		t.Errorf("contentHashMaxSize = %d, want 100", watcher.contentHashMaxSize)
	}
}

func TestWithContentHashMaxSize_DefaultEnablesHashing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// WithContentHashing() alone should set the default max size
	watcher, err := New([]string{tmpDir}, WithContentHashing())
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	if watcher.contentHashMaxSize != defaultContentHashMaxSize {
		t.Errorf("contentHashMaxSize = %d, want %d (default)",
			watcher.contentHashMaxSize, defaultContentHashMaxSize)
	}
}

func TestWithErrorBufferSize(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New([]string{tmpDir},
		WithBuffer(10),
		WithErrorBufferSize(100),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	errCh := watcher.Errors()
	if cap(errCh) != 100 {
		t.Errorf("error channel capacity = %d, want 100", cap(errCh))
	}
}

func TestWithErrorBufferSize_DefaultsToBuffer(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New([]string{tmpDir}, WithBuffer(50))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	errCh := watcher.Errors()
	if cap(errCh) != 50 {
		t.Errorf("error channel capacity = %d, want 50 (should default to bufferSize)", cap(errCh))
	}
}
