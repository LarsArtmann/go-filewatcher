package filewatcher

import (
	"errors"
	"testing"
	"time"
)

func TestWithIgnoreHidden(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithIgnoreHidden())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

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

	watcher, err := New([]string{dir}, WithOnAdd(func(path string) {
		addedPaths = append(addedPaths, path)
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

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

	watcher, err := New([]string{dir}, WithOnError(func(err error) {
		receivedErr = err
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

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

	watcher, err := New([]string{dir}, WithLazyIsDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	if !watcher.lazyIsDir {
		t.Error("expected lazyIsDir to be true")
	}
}

func TestWithPollInterval(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithPollInterval(5*time.Second))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	if watcher.pollInterval != 5*time.Second {
		t.Errorf("expected pollInterval 5s, got %v", watcher.pollInterval)
	}
}

func TestWithPolling(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithPolling(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	if !watcher.polling {
		t.Error("expected polling to be true")
	}

	if watcher.pollInterval != 2*time.Second {
		t.Errorf("expected default pollInterval 2s, got %v", watcher.pollInterval)
	}
}

func TestWithPolling_False(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithPolling(false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	if watcher.polling {
		t.Error("expected polling to be false")
	}
}

func TestWithDebug(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithDebug(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	if !watcher.debug {
		t.Error("expected debug to be true")
	}
}

func TestWithWatchedIgnoreDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	watcher, err := New([]string{dir}, WithWatchedIgnoreDirs("node_modules", ".cache"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	filterCount := len(watcher.filters)
	if filterCount == 0 {
		t.Error("expected at least one filter to be added")
	}
}
