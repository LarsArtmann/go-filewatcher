//nolint:testpackage,varnamelen // Tests need internal access; idiomatic short names
package filewatcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddPath_NonRecursive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithRecursive(false))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	addErr := w.addPath(NewRootPath(tmpDir))
	w.mu.Unlock()

	if addErr != nil {
		t.Fatalf("addPath non-recursive: %v", addErr)
	}

	if len(w.watchList) != 1 {
		t.Errorf("watchList should have 1 entry for non-recursive addPath, got %d", len(w.watchList))
	}
}

func TestAddPath_Recursive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub")

	mkdirErr := os.MkdirAll(subDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	addErr := w.addPath(NewRootPath(tmpDir))
	w.mu.Unlock()

	if addErr != nil {
		t.Fatalf("addPath recursive: %v", addErr)
	}

	list := w.WatchList()
	if len(list) < 1 {
		t.Errorf("watchList should have at least root, got %d", len(list))
	}
}

func TestWalkDirFunc_WalkError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	walkErr := w.walkDirFunc("/nonexistent/path", nil, os.ErrNotExist)
	w.mu.Unlock()

	if walkErr == nil {
		t.Error("expected error for walkErr path")
	}
}

func TestWalkDirFunc_NonDirEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	filePath := filepath.Join(tmpDir, "test.txt")

	writeErr := os.WriteFile(filePath, []byte("test"), testFilePermission)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatal(readErr)
	}

	var fileEntry os.DirEntry

	for _, e := range entries {
		if !e.IsDir() {
			fileEntry = e

			break
		}
	}

	if fileEntry == nil {
		t.Fatal("no file entry found")
	}

	w.mu.Lock()

	walkErr := w.walkDirFunc(filePath, fileEntry, nil)
	w.mu.Unlock()

	if walkErr != nil {
		t.Errorf("expected nil for non-dir file entry, got %v", walkErr)
	}
}

func TestWalkDirFunc_SkipsIgnoredDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithIgnoreDirs("node_modules"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	nmDir := filepath.Join(tmpDir, "node_modules")

	mkdirErr := os.MkdirAll(nmDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatal(readErr)
	}

	var dirEntry os.DirEntry

	for _, e := range entries {
		if e.IsDir() && e.Name() == "node_modules" {
			dirEntry = e

			break
		}
	}

	if dirEntry == nil {
		t.Fatal("node_modules dir not found")
	}

	w.mu.Lock()

	walkErr := w.walkDirFunc(nmDir, dirEntry, nil)
	w.mu.Unlock()

	if walkErr == nil {
		t.Error("expected SkipDir for ignored directory")
	}
}

func TestWalkAndAddPaths_WalkError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	walkErr := w.walkAndAddPaths(NewRootPath("/nonexistent/path/that/does/not/exist"))
	w.mu.Unlock()

	if walkErr == nil {
		t.Error("expected error walking nonexistent path")
	}
}

func TestShouldSkipDir_DotDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithSkipDotDirs(true))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	tests := []struct {
		name  string
		skip  bool
		input string
	}{
		{"hidden dir", true, ".git"},
		{"hidden config", true, ".config"},
		{"normal dir", false, "src"},
		{"root dot", false, "."},
	}

	for _, tt := range tests {
		got := w.shouldSkipDir(tt.input)

		if got != tt.skip {
			t.Errorf("shouldSkipDir(%q) = %v, want %v", tt.input, got, tt.skip)
		}
	}
}
