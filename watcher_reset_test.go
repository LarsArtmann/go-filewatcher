package filewatcher

import (
	"testing"
	"time"
)

func TestWatcher_Reset_AfterClose(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New([]string{tmpDir}, WithFilter(func(_ Event) bool { return true }))
	if err != nil {
		t.Fatal(err)
	}

	closeErr := watcher.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	resetErr := watcher.Reset()
	if resetErr != nil {
		t.Fatalf("Reset() failed: %v", resetErr)
	}

	if watcher.IsClosed() {
		t.Error("watcher should not be closed after Reset()")
	}

	if watcher.IsWatching() {
		t.Error("watcher should not be watching after Reset()")
	}

	if len(watcher.WatchList()) != 0 {
		t.Error("watchList should be empty after Reset()")
	}
}

func TestWatcher_Reset_PreservesConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New(
		[]string{tmpDir},
		WithFilter(func(_ Event) bool { return true }),
		WithRecursive(true),
		WithIgnoreDirs("ignored"),
	)
	if err != nil {
		t.Fatal(err)
	}

	filterCount := len(watcher.filters)
	ignoreCount := len(watcher.ignoreDirNames)

	closeErr := watcher.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	resetErr := watcher.Reset()
	if resetErr != nil {
		t.Fatalf("Reset() failed: %v", resetErr)
	}

	if len(watcher.filters) != filterCount {
		t.Errorf("filters not preserved: got %d, want %d", len(watcher.filters), filterCount)
	}

	if !watcher.recursive {
		t.Error("recursive flag not preserved")
	}

	if len(watcher.ignoreDirNames) != ignoreCount {
		t.Errorf("ignoreDirNames not preserved: got %d, want %d", len(watcher.ignoreDirNames), ignoreCount)
	}
}

func TestWatcher_Reset_WhileRunning(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher := newTestWatcher(t, tmpDir)

	ctx := setupTestContext(t, 5*time.Second)

	_, watchErr := watcher.Watch(ctx)
	if watchErr != nil {
		t.Fatalf("Watch() failed: %v (may be ENOSPC - that's OK for this test)", watchErr)
	}

	resetErr := watcher.Reset()
	if resetErr == nil {
		t.Error("expected error when resetting while running")
	}
}

func TestWatcher_Reset_PreservesCaseSensitivity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New(
		[]string{tmpDir},
		WithCaseSensitivity(CaseInsensitive),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify config before close
	if watcher.caseSensitivity != CaseInsensitive {
		t.Fatalf("pre-close caseSensitivity = %v, want CaseInsensitive", watcher.caseSensitivity)
	}

	closeErr := watcher.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	resetErr := watcher.Reset()
	if resetErr != nil {
		t.Fatalf("Reset() failed: %v", resetErr)
	}

	// Config should survive Reset
	if watcher.caseSensitivity != CaseInsensitive {
		t.Errorf("post-reset caseSensitivity = %v, want CaseInsensitive", watcher.caseSensitivity)
	}

	// effectiveCaseSensitivity should be re-resolved
	if watcher.effectiveCaseSensitivity != CaseInsensitive {
		t.Errorf("post-reset effectiveCaseSensitivity = %v, want CaseInsensitive", watcher.effectiveCaseSensitivity)
	}
}
