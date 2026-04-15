//nolint:testpackage,varnamelen,exhaustruct // Tests need internal access; idiomatic short names; partial struct initialization acceptable
package filewatcher

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

	err := os.WriteFile(tmpFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = New([]string{tmpFile})
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

	for i := 1; i <= 2; i++ {
		err := w.Close()
		if err != nil {
			t.Fatalf("close attempt %d failed: %v", i, err)
		}
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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := createTestFile(t, TempDir(tmpDir), "test.go", "package test")

	event := waitForEventOrFail(t, events, 3*time.Second)
	assertEventPath(t, event, testFile)

	if event.Op != Write && event.Op != Create {
		t.Errorf("expected Write or Create, got %s", event.Op.String())
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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	txtFile := filepath.Join(tmpDir, "test.txt")

	err = os.WriteFile(txtFile, []byte("text"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	goFile := filepath.Join(tmpDir, "test.go")

	err = os.WriteFile(goFile, []byte("package test"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventMatchingOrTimeout(t, events, 3*time.Second,
		func(event Event) {
			if event.Path != goFile {
				t.Errorf("expected go file event, got %s", event.Path)
			}
		},
		"timed out waiting for .go file event",
	)
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

	ctx := setupTestContext(t, 10*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := createTestFile(t, TempDir(tmpDir), "todelete.go", "package test")

	// Drain all events from file creation/write
	for waitForEventOrTimeout(t, events, 500*time.Millisecond) {
	}

	err = os.Remove(testFile)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventMatchingOrTimeout(t, events, 5*time.Second,
		func(event Event) {
			if event.Op != Remove {
				t.Errorf("expected Remove, got %s", event.Op.String())
			}
		},
		"timed out waiting for remove event",
	)
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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	_ = createTestFile(t, TempDir(tmpDir), "test.txt", "test")

	receiveEventOrTimeout(t, events, 3*time.Second)

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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")

	for i := range 5 {
		err := os.WriteFile(testFile, []byte("test"+string(rune('0'+i))), 0o600)
		if err != nil {
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

	// Close watcher explicitly after collecting events to avoid race
	// between debouncer callbacks and channel closure.
	_ = w.Close()

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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	file1 := filepath.Join(tmpDir, "a.txt")
	file2 := filepath.Join(tmpDir, "b.txt")

	err = os.WriteFile(file1, []byte("a"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file2, []byte("b"), 0o600)
	if err != nil {
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

	// Close watcher explicitly after collecting events to avoid race.
	_ = w.Close()

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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	newDir := filepath.Join(tmpDir, "subdir")

	err = os.Mkdir(newDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for directory creation event
	select {
	case <-events:
	case <-time.After(2 * time.Second):
	}

	// Now create a file in the new subdirectory
	nestedFile := filepath.Join(newDir, "nested.txt")

	err = os.WriteFile(nestedFile, []byte("nested"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventMatchingOrTimeout(t, events, 3*time.Second,
		func(event Event) {
			if event.Path != nestedFile {
				t.Errorf("expected event for nested file %s, got %s", nestedFile, event.Path)
			}
		},
		"timed out waiting for nested file event - new directory may not be watched",
	)
}

func TestWatcher_Watch_ErrorHandler(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var errorReceived atomic.Pointer[error]

	w, err := New([]string{tmpDir},
		WithErrorHandler(func(ctx ErrorContext, err error) {
			_ = ctx

			errorReceived.Store(&err)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	newDir := t.TempDir()

	err = w.Add(newDir)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	testFile := createTestFile(t, TempDir(newDir), "added.txt", "added")

	receiveEventMatchingOrTimeout(t, events, 3*time.Second,
		func(event Event) {
			if event.Path != testFile {
				t.Errorf("expected event from added dir, got %s", event.Path)
			}
		},
		"timed out waiting for event from added directory",
	)
}

func TestWatcher_Remove(t *testing.T) {
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

	err = w.Remove(tmpDir)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_ = createTestFile(t, TempDir(tmpDir), "after-remove.txt", "test")

	select {
	case event := <-events:
		t.Errorf("expected no events after Remove(), got %v", event)
	case <-time.After(500 * time.Millisecond):
	}
}

func TestWatcher_Remove_ClosedWatcher(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	_ = w.Close()

	err = w.Remove(tmpDir)
	if err == nil {
		t.Fatal("expected error when removing from closed watcher")
	}
}

func TestWatcher_WatchList(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	_, err = w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	list := w.WatchList()
	if len(list) == 0 {
		t.Fatal("expected non-empty watch list after Watch()")
	}

	found := slices.Contains(list, tmpDir)
	if !found {
		t.Errorf("expected %q in watch list, got %v", tmpDir, list)
	}
}

func TestWatcher_WatchList_IsCopy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx := setupTestContext(t, 5*time.Second)

	_, err = w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	list1 := w.WatchList()
	list2 := w.WatchList()

	if len(list1) != len(list2) {
		t.Fatal("WatchList should return consistent length")
	}

	if &list1[0] == &list2[0] {
		t.Error("WatchList should return a copy, not the same slice")
	}
}

func TestWatcher_Stats(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	stats := w.Stats()
	if stats.IsClosed {
		t.Error("expected watcher not to be closed")
	}

	if stats.IsWatching {
		t.Error("expected watcher not to be watching before Watch()")
	}

	if stats.WatchCount != 0 {
		t.Errorf("expected 0 watch count before Watch(), got %d", stats.WatchCount)
	}

	ctx := setupTestContext(t, 5*time.Second)

	_, err = w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	stats = w.Stats()
	if stats.IsClosed {
		t.Error("expected watcher not to be closed")
	}

	if !stats.IsWatching {
		t.Error("expected watcher to be watching after Watch()")
	}

	if stats.WatchCount == 0 {
		t.Error("expected non-zero watch count after Watch()")
	}
}

func TestWatcher_Stats_Metrics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir},
		WithExtensions(".go"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	// Verify initial metrics are zero
	stats := w.Stats()
	if stats.EventsProcessed != 0 {
		t.Errorf("expected 0 events processed initially, got %d", stats.EventsProcessed)
	}

	if stats.EventsFilteredOut != 0 {
		t.Errorf("expected 0 events filtered initially, got %d", stats.EventsFilteredOut)
	}

	if stats.ErrorsEncountered != 0 {
		t.Errorf("expected 0 errors initially, got %d", stats.ErrorsEncountered)
	}

	if stats.Uptime != 0 {
		t.Error("expected 0 uptime before Watch()")
	}

	// Start watching
	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Create a .go file (should pass filter)
	testFile := filepath.Join(tmpDir, "test.go")

	err = os.WriteFile(testFile, []byte("package main"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for event
	event := waitForEventOrFail(t, events, 2*time.Second)
	assertEventPath(t, event, testFile)

	// Create a .txt file (should be filtered)
	txtFile := filepath.Join(tmpDir, "test.txt")

	err = os.WriteFile(txtFile, []byte("text"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Wait a bit for potential events
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats = w.Stats()

	// Should have processed the .go file
	if stats.EventsProcessed != 1 {
		t.Errorf("expected 1 event processed, got %d", stats.EventsProcessed)
	}

	// Should have filtered the .txt file
	if stats.EventsFilteredOut == 0 {
		t.Error("expected some events to be filtered out")
	}

	// Should have uptime
	if stats.Uptime == 0 {
		t.Error("expected non-zero uptime after Watch()")
	}

	// No errors expected
	if stats.ErrorsEncountered != 0 {
		t.Errorf("expected 0 errors, got %d", stats.ErrorsEncountered)
	}

	// Drain remaining events
	for {
		select {
		case <-events:
		default:
			return
		}
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

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	vendorDir := filepath.Join(tmpDir, "vendor")

	err = os.Mkdir(vendorDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	vendorFile := filepath.Join(vendorDir, "lib.go")

	err = os.WriteFile(vendorFile, []byte("package vendor"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	normalFile := filepath.Join(tmpDir, "main.go")

	err = os.WriteFile(normalFile, []byte("package main"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	receiveEventMatchingOrTimeout(t, events, 3*time.Second,
		func(event Event) {
			if event.Path == vendorFile {
				t.Error("vendor file should have been filtered")
			}

			if event.Path != normalFile {
				t.Errorf("expected normal file event, got %s", event.Path)
			}
		},
		"timed out waiting for event",
	)
}

//nolint:paralleltest // Not parallel: captures os.Stderr, which is a global resource.
func TestWatcher_handleError_Default(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	old := os.Stderr
	r, w2, _ := os.Pipe()
	os.Stderr = w2

	//nolint:err113 // test-only error for stderr validation
	w.handleError(ErrorContext{Operation: "test"}, errors.New("test stderr error"))

	_ = w2.Close()
	os.Stderr = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "test stderr error") {
		t.Errorf("expected error on stderr, got %q", buf.String())
	}
}

func TestWatcher_Watch_DoubleWatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	_, err = w.Watch(ctx)
	if err != nil {
		t.Fatalf("first Watch failed: %v", err)
	}

	_, err = w.Watch(ctx)
	if err == nil {
		t.Fatal("expected error when watching already running watcher")
	}

	if !errors.Is(err, ErrWatcherRunning) {
		t.Errorf("expected ErrWatcherRunning, got %v", err)
	}
}

func TestWatcher_Add_ClosedWatcher(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	_ = w.Close()

	if !w.IsClosed() {
		t.Error("expected IsClosed() to return true after Close()")
	}

	err = w.Add(t.TempDir())
	if err == nil {
		t.Fatal("expected error when adding to closed watcher")
	}
}

// TestWatcher_FullLifecycle is an integration test for the complete
// Watch → Event → Close lifecycle.
//
//nolint:all // Integration test - complexity acceptable for comprehensive coverage
func TestWatcher_FullLifecycle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create watcher with filters and middleware
	var eventCount atomic.Int32

	w, err := New([]string{tmpDir},
		WithExtensions(".go"),
		WithDebounce(50*time.Millisecond),
		WithMiddleware(func(next Handler) Handler {
			return func(ctx context.Context, event Event) error {
				eventCount.Add(1)

				return next(ctx, event)
			}
		}),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Verify initial state
	if w.IsClosed() {
		t.Error("expected watcher not to be closed initially")
	}

	if w.IsWatching() {
		t.Error("expected watcher not to be watching before Watch()")
	}

	stats := w.Stats()
	if stats.IsClosed || stats.IsWatching || stats.WatchCount != 0 {
		t.Errorf("unexpected initial stats: %+v", stats)
	}

	// Start watching
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Verify watching state
	if !w.IsWatching() {
		t.Error("expected watcher to be watching after Watch()")
	}

	stats = w.Stats()
	if !stats.IsWatching {
		t.Error("expected stats.IsWatching to be true")
	}

	// Create a test file to trigger an event
	testFile := filepath.Join(tmpDir, "test.go")

	err = os.WriteFile(testFile, []byte("package main"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Wait for event with timeout
	select {
	case event := <-events:
		assertEventPath(t, event, testFile)

		if event.Op != Create && event.Op != Write {
			t.Errorf("expected Create or Write op, got %s", event.Op)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	// Verify event was processed by middleware
	if eventCount.Load() != 1 {
		t.Errorf("expected middleware to process 1 event, got %d", eventCount.Load())
	}

	// Close the watcher
	err = w.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify closed state
	if !w.IsClosed() {
		t.Error("expected watcher to be closed after Close()")
	}

	if w.IsWatching() {
		t.Error("expected watcher not to be watching after Close()")
	}

	stats = w.Stats()
	if !stats.IsClosed || stats.IsWatching {
		t.Errorf("unexpected final stats: %+v", stats)
	}

	// Verify channel is closed
	assertChannelClosed(t, events, time.Second, "event channel")

	// Verify WatchList is empty after close
	if len(w.WatchList()) != 0 {
		t.Errorf("expected empty watch list after close, got %v", w.WatchList())
	}
}

func TestWatcher_Errors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	// Get errors channel
	errorsCh := w.Errors()
	if errorsCh == nil {
		t.Fatal("expected non-nil errors channel")
	}

	// Verify we can receive from the channel
	select {
	case <-errorsCh:
		// No errors expected yet
	default:
		// Channel is empty, which is expected
	}
}

//nolint:paralleltest // Not parallel: uses os.Pipe which has global effects
func TestWatcher_Errors_ReceivesErrors(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	// Get errors channel BEFORE starting watch
	errorsCh := w.Errors()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Trigger an error by sending to a closed watcher (simulated via handleError)
	//nolint:err113 // test-only error
	testErr := errors.New("test error from handler")
	w.handleError(ErrorContext{Operation: "test_op", Path: tmpDir}, testErr)

	// Wait for error on channel
	select {
	case err := <-errorsCh:
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, testErr) {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for error on channel")
	}

	_ = w.Close()

	// Verify channel is closed
	assertChannelClosed(t, errorsCh, time.Second, "errors channel")
}
