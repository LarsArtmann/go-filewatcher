package filewatcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew_NoPaths(t *testing.T) {
	t.Parallel()

	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
	if !errors.Is(err, ErrNoPaths) {
		t.Errorf("expected ErrNoPaths, got %v", err)
	}
}

func TestNew_NonexistentPath(t *testing.T) {
	t.Parallel()

	_, err := New([]string{"/nonexistent/path/that/does/not/exist"})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !errors.Is(err, ErrPathNotFound) {
		t.Errorf("expected ErrPathNotFound, got %v", err)
	}
}

func TestNew_FilePath(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := New([]string{tmpFile})
	if err == nil {
		t.Fatal("expected error for file path")
	}
	if !errors.Is(err, ErrPathNotDir) {
		t.Errorf("expected ErrPathNotDir, got %v", err)
	}
}

func TestNew_ValidPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = w.Close() }()

	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestWatcher_Close_Twice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}

func TestWatcher_Watch_AfterClose(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	_ = w.Close()

	_, err = w.Watch(context.Background())
	if err == nil {
		t.Fatal("expected error when watching closed watcher")
	}
	if !errors.Is(err, ErrWatcherClosed) {
		t.Errorf("expected ErrWatcherClosed, got %v", err)
	}
}

func TestWatcher_Watch_DetectsWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithExtensions(".go"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package test"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Path != testFile {
			t.Errorf("expected path %s, got %s", testFile, event.Path)
		}
		if event.Op != Write && event.Op != Create {
			t.Errorf("expected Write or Create, got %s", event.Op)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestWatcher_Watch_FiltersExtensions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithExtensions(".go"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}

	goFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(goFile, []byte("package test"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Path != goFile {
			t.Errorf("expected go file event, got %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for .go file event")
	}
}

func TestWatcher_Watch_ContextCancellation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	cancel()

	_, ok := <-events
	if ok {
		t.Error("expected channel to be closed after context cancellation")
	}
}

func TestWatcher_Watch_Deletes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "todelete.go")
	if err := os.WriteFile(testFile, []byte("package test"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Wait for create event
	select {
	case <-events:
	case <-time.After(2 * time.Second):
	}

	if err := os.Remove(testFile); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Op != Remove {
			t.Errorf("expected Remove, got %s", event.Op)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for remove event")
	}
}

func TestWatcher_Watch_WithMiddleware(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var processed atomic.Int32

	w, err := New([]string{tmpDir},
		WithMiddleware(func(next Handler) Handler {
			return func(ctx context.Context, event Event) error {
				processed.Add(1)
				return next(ctx, event)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-events:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	if got := processed.Load(); got != 1 {
		t.Errorf("expected middleware to be called once, got %d", got)
	}
}

func TestWatcher_Watch_WithDebounce(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir},
		WithDebounce(100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")

	for i := range 5 {
		if err := os.WriteFile(testFile, []byte("test"+string(rune('0'+i))), 0o600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	var eventCount atomic.Int32
	timeout := time.After(2 * time.Second)

collect:
	for {
		select {
		case <-events:
			eventCount.Add(1)
		case <-timeout:
			break collect
		}
	}

	if got := eventCount.Load(); got != 1 {
		t.Errorf("expected 1 debounced event from 5 rapid writes, got %d", got)
	}
}

func TestWatcher_Watch_WithPerPathDebounce(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir},
		WithPerPathDebounce(50*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")

	if err := os.WriteFile(file1, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	var paths []string
	timeout := time.After(2 * time.Second)

collect:
	for {
		select {
		case event := <-events:
			paths = append(paths, event.Path)
			if len(paths) == 2 {
				break collect
			}
		case <-timeout:
			break collect
		}
	}

	if len(paths) != 2 {
		t.Errorf(
			"expected 2 events for different files with per-path debounce, got %d: %v",
			len(paths),
			paths,
		)
	}
}

func TestWatcher_Watch_NewDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithRecursive(true))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	newDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(newDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Wait for directory creation event
	select {
	case <-events:
	case <-time.After(2 * time.Second):
	}

	// Now create a file in the new subdirectory
	nestedFile := filepath.Join(newDir, "nested.txt")
	if err := os.WriteFile(nestedFile, []byte("nested"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Path != nestedFile {
			t.Errorf("expected event for nested file %s, got %s", nestedFile, event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for nested file event - new directory may not be watched")
	}
}

func TestWatcher_Watch_ErrorHandler(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var errorReceived atomic.Pointer[error]
	w, err := New([]string{tmpDir},
		WithErrorHandler(func(err error) {
			errorReceived.Store(&err)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	_ = w

	// Error handler is set; just verify construction works
	if errorReceived.Load() != nil {
		t.Error("expected no error yet")
	}
}

func TestWatcher_Add(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	newDir := t.TempDir()
	if err := w.Add(newDir); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	testFile := filepath.Join(newDir, "added.txt")
	if err := os.WriteFile(testFile, []byte("added"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Path != testFile {
			t.Errorf("expected event from added dir, got %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event from added directory")
	}
}

func TestWatcher_IgnoreDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir},
		WithIgnoreDirs("vendor"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.Mkdir(vendorDir, 0o750); err != nil {
		t.Fatal(err)
	}

	vendorFile := filepath.Join(vendorDir, "lib.go")
	if err := os.WriteFile(vendorFile, []byte("package vendor"), 0o600); err != nil {
		t.Fatal(err)
	}

	normalFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(normalFile, []byte("package main"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Path == vendorFile {
			t.Error("vendor file should have been filtered")
		}
		if event.Path != normalFile {
			t.Errorf("expected normal file event, got %s", event.Path)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}
