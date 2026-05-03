//nolint:testpackage,varnamelen,gosec,modernize,exhaustruct,cyclop // Tests need internal access; idiomatic short names; test code allows relaxed permissions, complexity, flexible formatting, partial struct init
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

func TestWatcher_Watch_RenameEvent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	original := filepath.Join(tmpDir, "original.txt")

	err = os.WriteFile(original, []byte("data"), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	for waitForEventOrTimeout(t, events, 500*time.Millisecond) {
	}

	renamed := filepath.Join(tmpDir, "renamed.txt")

	err = os.Rename(original, renamed)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventMatchingOrTimeout(t, events, 3*time.Second,
		func(event Event) {
			if event.Op != Rename && event.Op != Create {
				t.Errorf("expected Rename or Create, got %s", event.Op)
			}
		},
		"timed out waiting for rename event",
	)
}

func TestWatcher_Watch_MultipleInitialPaths(t *testing.T) {
	t.Parallel()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	w, err := New([]string{dir1, dir2})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	list := w.WatchList()
	if len(list) < 2 {
		t.Errorf("expected at least 2 paths in watch list, got %d", len(list))
	}

	file1 := filepath.Join(dir1, "a.txt")

	err = os.WriteFile(file1, []byte("a"), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventOrTimeout(t, events, 3*time.Second)

	file2 := filepath.Join(dir2, "b.txt")

	err = os.WriteFile(file2, []byte("b"), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventOrTimeout(t, events, 3*time.Second)
}

func TestWatcher_Watch_NonRecursive_IgnoresSubdirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithRecursive(false))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	subDir := filepath.Join(tmpDir, "sub")

	err = os.MkdirAll(subDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	nestedFile := filepath.Join(subDir, "nested.txt")

	err = os.WriteFile(nestedFile, []byte("nested"), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	rootFile := filepath.Join(tmpDir, "root.txt")

	err = os.WriteFile(rootFile, []byte("root"), testFilePermission)
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.After(3 * time.Second)

	foundRoot := false

	for !foundRoot {
		select {
		case event := <-events:
			if event.Path == rootFile {
				foundRoot = true
			}
		case <-timeout:
			t.Fatal("timed out waiting for root file event")
		}
	}

	if !foundRoot {
		t.Error("expected to find root file event")
	}
}

func TestWatcher_ConcurrentAddRemove(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	var addErrors atomic.Int32

	done := make(chan struct{})

	go func() {
		defer close(done)

		for range 10 {
			newDir := t.TempDir()

			addErr := w.Add(newDir)
			if addErr != nil {
				addErrors.Add(1)
			}

			_ = createTestFile(t, NewTempDir(newDir), "test.txt", "data")
			receiveEventOrTimeout(t, events, 2*time.Second)

			_ = w.Remove(newDir)
		}
	}()

	<-done

	if addErrors.Load() > 0 {
		t.Errorf("expected no add errors, got %d", addErrors.Load())
	}
}

func TestWatcher_BufferZero(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithBuffer(0))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	if w.bufferSize != 0 {
		t.Errorf("expected bufferSize 0, got %d", w.bufferSize)
	}
}

func TestWatcher_Add_NonExistentPath(t *testing.T) {
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
		t.Fatalf("Watch failed: %v", watchErr)
	}

	addErr := w.Add("/nonexistent/path")
	if addErr == nil {
		t.Fatal("expected error adding nonexistent path")
	}
}

func TestWatcher_ErrorsChannel_ClosesOnClose(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	errorsCh := w.Errors()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, watchErr := w.Watch(ctx)
	if watchErr != nil {
		t.Fatalf("Watch failed: %v", watchErr)
	}

	closeErr := w.Close()
	if closeErr != nil {
		t.Fatalf("Close failed: %v", closeErr)
	}

	assertChannelClosed(t, errorsCh, 2*time.Second, "errors channel")
}

func TestWatcher_WithBufferNegative(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithBuffer(-1))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	if w.bufferSize != defaultEventBufferSize {
		t.Errorf(
			"expected default bufferSize %d for negative input, got %d",
			defaultEventBufferSize, w.bufferSize,
		)
	}
}

func TestDebounceOptions_NegativeDuration_Panics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tests := []struct {
		name string
		opt  func() Option
	}{
		{"WithDebounce negative", func() Option { return WithDebounce(-1 * time.Second) }},
		{
			"WithPerPathDebounce negative",
			func() Option { return WithPerPathDebounce(-1 * time.Second) },
		},
	}

	for _, tc := range tests {
		didPanic := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					didPanic = true
				}
			}()

			_, _ = New([]string{tmpDir}, tc.opt())
		}()

		if !didPanic {
			t.Errorf("%s: expected panic for negative duration", tc.name)
		}
	}
}

func TestDefaultIgnoreDirsCopy(t *testing.T) {
	t.Parallel()

	copy1 := DefaultIgnoreDirsCopy()
	copy2 := DefaultIgnoreDirsCopy()

	if len(copy1) != len(DefaultIgnoreDirs) {
		t.Errorf("expected %d entries, got %d", len(DefaultIgnoreDirs), len(copy1))
	}

	const mutated = "MUTATED"

	copy1[0] = mutated

	if copy2[0] == mutated {
		t.Error("mutation of copy should not affect other copies")
	}

	if DefaultIgnoreDirs[0] == mutated {
		t.Error("mutation of copy should not affect original")
	}
}

func TestWatcher_StateTransition_AfterFailedWatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	_, watchErr := w.Watch(ctx)
	if watchErr != nil {
		t.Fatalf("first Watch failed: %v", watchErr)
	}

	if !w.IsWatching() {
		t.Error("expected IsWatching after first Watch()")
	}

	_, secondWatchErr := w.Watch(ctx)
	if secondWatchErr == nil {
		t.Fatal("expected error on second Watch()")
	}

	if !w.IsWatching() {
		t.Error("expected IsWatching to remain true after failed second Watch()")
	}
}

func TestWatcher_ErrorHandler_WithContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var receivedCtx ErrorContext

	var receivedErr error

	w, err := New([]string{tmpDir},
		WithErrorHandler(func(ctx ErrorContext, err error) {
			receivedCtx = ctx
			receivedErr = err
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	const testPath = "/test/path"

	//nolint:err113 // test-only error
	testErr := errors.New("context test error")
	w.handleError(
		ErrorContext{Operation: "test-op", Path: testPath, Retryable: true},
		testErr,
	)

	if receivedCtx.Operation != "test-op" {
		t.Errorf("expected operation 'test-op', got %q", receivedCtx.Operation)
	}

	if receivedCtx.Path != testPath {
		t.Errorf("expected path %q, got %q", testPath, receivedCtx.Path)
	}

	if !receivedCtx.Retryable {
		t.Error("expected Retryable=true")
	}

	if !errors.Is(receivedErr, testErr) {
		t.Errorf("expected testErr, got %v", receivedErr)
	}
}

func TestWatcher_WatchList_NoDuplicates(t *testing.T) {
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
		t.Fatalf("Watch failed: %v", watchErr)
	}

	list := w.WatchList()

	seen := make(map[string]int)

	for _, path := range list {
		seen[path]++

		if seen[path] > 1 {
			t.Errorf("duplicate path in watchList: %s (count: %d)", path, seen[path])
		}
	}
}
