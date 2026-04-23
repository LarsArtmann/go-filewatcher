package filewatcher

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestConvertEvent_LazyIsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// With lazyIsDir=true, even a directory path should return IsDir=false
	fsEvent := fsnotify.Event{
		Name: dir,
		Op:   fsnotify.Create,
	}

	result := convertEvent(fsEvent, true)
	if result == nil {
		t.Fatal("expected non-nil event")
	}

	if result.IsDir {
		t.Error("expected IsDir=false when lazyIsDir=true, even for a directory path")
	}

	if result.Op != Create {
		t.Errorf("expected Op=Create, got %v", result.Op)
	}
}

func TestConvertEvent_NormalIsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// With lazyIsDir=false, a directory should be detected
	fsEvent := fsnotify.Event{
		Name: dir,
		Op:   fsnotify.Create,
	}

	result := convertEvent(fsEvent, false)
	if result == nil {
		t.Fatal("expected non-nil event")
	}

	if !result.IsDir {
		t.Error("expected IsDir=true when lazyIsDir=false for a directory path")
	}
}

func TestConvertEvent_ChmodIgnored(t *testing.T) {
	t.Parallel()

	fsEvent := fsnotify.Event{
		Name: "/tmp/test.go",
		Op:   fsnotify.Chmod,
	}

	result := convertEvent(fsEvent, false)
	if result != nil {
		t.Error("expected nil for Chmod event")
	}
}
