package filewatcher

import (
	"errors"
	"testing"
	"time"
)

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
