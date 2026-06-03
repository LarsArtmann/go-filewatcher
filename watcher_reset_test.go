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

	watcher, err := New([]string{tmpDir},
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

	watcher, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

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
