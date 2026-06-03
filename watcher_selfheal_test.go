package filewatcher

import (
	"testing"
	"time"
)

func TestSelfHeal_DisabledByDefault(t *testing.T) {
	t.Parallel()

	watcher, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, "selfHealInterval", watcher.selfHealInterval, time.Duration(0))

	_ = watcher.Close()
}

func TestSelfHeal_EnabledViaOption(t *testing.T) {
	t.Parallel()

	watcher, err := New(
		[]string{t.TempDir()},
		WithSelfHeal(100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}

	if watcher.selfHealInterval != 100*time.Millisecond {
		t.Errorf("selfHealInterval = %v, want 100ms", watcher.selfHealInterval)
	}

	_ = watcher.Close()
}

func TestSelfHeal_ZeroIntervalIgnored(t *testing.T) {
	t.Parallel()

	watcher, err := New(
		[]string{t.TempDir()},
		WithSelfHeal(0), // explicitly disabled
	)
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, "selfHealInterval (zero ignored)", watcher.selfHealInterval, time.Duration(0))

	_ = watcher.Close()
}

func TestSelfHeal_TracksFailedPaths(t *testing.T) {
	t.Parallel()

	watcher, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	// Manually add non-existent paths
	watcher.failedPaths["/nonexistent/path"] = struct{}{}
	watcher.failedPaths["/another/missing"] = struct{}{}

	if got := watcher.failedPathCount(); got != 2 {
		t.Errorf("failedPathCount = %d, want 2", got)
	}

	// Remove one
	watcher.removeFailedPath("/nonexistent/path")

	assertEqual(t, "failedPathCount (after remove)", watcher.failedPathCount(), 1)
}

func TestSelfHeal_NoOpWhenEmpty(t *testing.T) {
	t.Parallel()

	watcher, err := New([]string{t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	// attemptSelfHeal should be a no-op when no paths have failed
	watcher.attemptSelfHeal()

	assertEqual(t, "failedPathCount (no-op)", watcher.failedPathCount(), 0)
}
