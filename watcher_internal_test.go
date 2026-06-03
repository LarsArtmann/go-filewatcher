package filewatcher

import (
	"os"
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

	result := convertEvent(fsEvent, true, false)
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

	result := convertEvent(fsEvent, false, false)
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
		Name: benchmarkTestPathTestGo,
		Op:   fsnotify.Chmod,
	}

	result := convertEvent(fsEvent, false, false)
	if result != nil {
		t.Error("expected nil for Chmod event")
	}
}

func TestConvertEvent_WithHash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	file := dir + "/test.txt"

	err := os.WriteFile(file, []byte("hello world"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	fsEvent := fsnotify.Event{Name: file, Op: fsnotify.Write}

	result := convertEvent(fsEvent, false, true)
	if result == nil {
		t.Fatal("expected non-nil event")
	}

	// SHA-256 of "hello world" is b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
	want := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if result.Hash != want {
		t.Errorf("Hash = %q, want %q", result.Hash, want)
	}
}

func TestConvertEvent_WithoutHash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	file := dir + "/test.txt"

	err := os.WriteFile(file, []byte("hello"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	fsEvent := fsnotify.Event{Name: file, Op: fsnotify.Write}

	result := convertEvent(fsEvent, false, false)
	if result == nil {
		t.Fatal("expected non-nil event")
	}

	if result.Hash != "" {
		t.Errorf("Hash = %q, want empty when computeHash=false", result.Hash)
	}
}

func TestConvertEvent_HashForDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	fsEvent := fsnotify.Event{Name: dir, Op: fsnotify.Create}

	result := convertEvent(fsEvent, false, true)
	if result == nil {
		t.Fatal("expected non-nil event")
	}

	if result.Hash != "" {
		t.Errorf("Hash for directory = %q, want empty", result.Hash)
	}
}
