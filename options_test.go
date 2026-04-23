//nolint:testpackage // Tests need internal access to verify Watcher fields
package filewatcher

import (
	"errors"
	"testing"
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
		ErrorContext{ //nolint:exhaustruct // test-specific minimal fields
			Operation: "test",
			Path:      "test",
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
