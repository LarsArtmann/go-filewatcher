package filewatcher

import (
	"testing"
	"time"
)

func TestWatcher_Reset_AfterClose(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithFilter(func(_ Event) bool { return true }))
	if err != nil {
		t.Fatal(err)
	}

	closeErr := w.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	resetErr := w.Reset()
	if resetErr != nil {
		t.Fatalf("Reset() failed: %v", resetErr)
	}

	if w.IsClosed() {
		t.Error("watcher should not be closed after Reset()")
	}

	if w.IsWatching() {
		t.Error("watcher should not be watching after Reset()")
	}

	if len(w.WatchList()) != 0 {
		t.Error("watchList should be empty after Reset()")
	}
}

func TestWatcher_Reset_PreservesConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir},
		WithFilter(func(_ Event) bool { return true }),
		WithRecursive(true),
		WithIgnoreDirs("ignored"),
	)
	if err != nil {
		t.Fatal(err)
	}

	filterCount := len(w.filters)
	ignoreCount := len(w.ignoreDirNames)

	closeErr := w.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	resetErr := w.Reset()
	if resetErr != nil {
		t.Fatalf("Reset() failed: %v", resetErr)
	}

	if len(w.filters) != filterCount {
		t.Errorf("filters not preserved: got %d, want %d", len(w.filters), filterCount)
	}

	if !w.recursive {
		t.Error("recursive flag not preserved")
	}

	if len(w.ignoreDirNames) != ignoreCount {
		t.Errorf("ignoreDirNames not preserved: got %d, want %d", len(w.ignoreDirNames), ignoreCount)
	}
}

func TestWatcher_Reset_WhileRunning(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)
	_, watchErr := w.Watch(ctx)
	if watchErr != nil {
		t.Fatalf("Watch() failed: %v (may be ENOSPC - that's OK for this test)", watchErr)
	}

	resetErr := w.Reset()
	if resetErr == nil {
		t.Error("expected error when resetting while running")
	}
}
