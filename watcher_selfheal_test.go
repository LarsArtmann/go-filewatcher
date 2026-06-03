package filewatcher

import (
	"testing"
	"time"
)

func TestSelfHeal_DisabledByDefault(t *testing.T) {
	t.Parallel()

	w, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	if w.selfHealInterval != 0 {
		t.Errorf("selfHealInterval = %v, want 0 (disabled)", w.selfHealInterval)
	}

	_ = w.Close()
}

func TestSelfHeal_EnabledViaOption(t *testing.T) {
	t.Parallel()

	w, err := New(
		[]string{t.TempDir()},
		WithSelfHeal(100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}

	if w.selfHealInterval != 100*time.Millisecond {
		t.Errorf("selfHealInterval = %v, want 100ms", w.selfHealInterval)
	}

	_ = w.Close()
}

func TestSelfHeal_ZeroIntervalIgnored(t *testing.T) {
	t.Parallel()

	w, err := New(
		[]string{t.TempDir()},
		WithSelfHeal(0), // explicitly disabled
	)
	if err != nil {
		t.Fatal(err)
	}

	if w.selfHealInterval != 0 {
		t.Errorf("selfHealInterval = %v, want 0 (zero is ignored)", w.selfHealInterval)
	}

	_ = w.Close()
}

func TestSelfHeal_TracksFailedPaths(t *testing.T) {
	t.Parallel()

	w, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	// Manually add non-existent paths
	w.failedPaths["/nonexistent/path"] = struct{}{}
	w.failedPaths["/another/missing"] = struct{}{}

	if got := w.failedPathCount(); got != 2 {
		t.Errorf("failedPathCount = %d, want 2", got)
	}

	// Remove one
	w.removeFailedPathLocked("/nonexistent/path")

	if got := w.failedPathCount(); got != 1 {
		t.Errorf("failedPathCount = %d, want 1", got)
	}
}

func TestSelfHeal_NoOpWhenEmpty(t *testing.T) {
	t.Parallel()

	w, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	// attemptSelfHeal should be a no-op when no paths have failed
	w.attemptSelfHeal()

	if got := w.failedPathCount(); got != 0 {
		t.Errorf("failedPathCount = %d, want 0", got)
	}
}
