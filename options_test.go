//nolint:testpackage // Tests need internal access to verify Watcher fields
package filewatcher

import (
	"errors"
	"testing"
)

func TestWithIgnoreHidden(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	w, err := New([]string{dir}, WithIgnoreHidden())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer w.Close()

	if len(w.filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(w.filters))
	}

	// Verify the filter actually rejects hidden files
	hidden := testWriteEvent(".hidden_file")
	if w.filters[0](hidden) {
		t.Error("expected hidden file to be filtered out")
	}

	visible := testWriteEvent("visible_file")
	if !w.filters[0](visible) {
		t.Error("expected visible file to pass filter")
	}
}

func TestWithOnAdd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var addedPaths []string

	w, err := New([]string{dir}, WithOnAdd(func(path string) {
		addedPaths = append(addedPaths, path)
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer w.Close()

	if w.onAdd == nil {
		t.Fatal("expected onAdd callback to be set")
	}

	// Invoke the callback to verify it works
	w.onAdd(dir)

	if len(addedPaths) != 1 || addedPaths[0] != dir {
		t.Errorf("expected callback to receive %q, got %v", dir, addedPaths)
	}
}

func TestWithOnError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var receivedErr error

	w, err := New([]string{dir}, WithOnError(func(err error) {
		receivedErr = err
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer w.Close()

	if w.errorHandler == nil {
		t.Fatal("expected errorHandler to be set")
	}

	// Invoke the error handler to verify it delegates correctly
	testErr := errors.New("test error") //nolint:err113 // test-specific dynamic error
	w.errorHandler(ErrorContext{}, testErr)

	if !errors.Is(receivedErr, testErr) {
		t.Errorf("expected callback to receive test error, got %v", receivedErr)
	}
}

func TestWithLazyIsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	w, err := New([]string{dir}, WithLazyIsDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer w.Close()

	if !w.lazyIsDir {
		t.Error("expected lazyIsDir to be true")
	}
}
