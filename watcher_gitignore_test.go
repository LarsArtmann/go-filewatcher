package filewatcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitignore_SkipsIgnoredDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	buildDir := filepath.Join(tmpDir, "build")
	mkdirErr := os.MkdirAll(buildDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignoreContent := "build/\n"
	writeErr := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), testFilePermission)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New([]string{tmpDir}, WithGitignore(true))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	walkErr := w.walkDirFunc(buildDir, &dirEntry{name: "build", isDir: true}, nil)
	w.mu.Unlock()

	if walkErr == nil {
		t.Error("expected SkipDir for gitignored directory")
	}
}

func TestGitignore_DoesNotSkipNonIgnored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	mkdirErr := os.MkdirAll(srcDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignoreContent := "build/\n"
	writeErr := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), testFilePermission)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New([]string{tmpDir}, WithGitignore(true))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	walkErr := w.walkDirFunc(srcDir, &dirEntry{name: "src", isDir: true}, nil)
	w.mu.Unlock()

	if walkErr != nil {
		t.Errorf("expected no error for non-ignored directory, got %v", walkErr)
	}
}

func TestGitignore_Disabled(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	buildDir := filepath.Join(tmpDir, "my-output")
	mkdirErr := os.MkdirAll(buildDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignoreContent := "my-output/\n"
	writeErr := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), testFilePermission)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New([]string{tmpDir}, WithGitignore(false))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.mu.Lock()

	walkErr := w.walkDirFunc(buildDir, &dirEntry{name: "my-output", isDir: true}, nil)
	w.mu.Unlock()

	if walkErr != nil {
		t.Errorf("expected no error when gitignore disabled, got %v", walkErr)
	}
}

func TestGitignore_NoGitignoreFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	w, err := New([]string{tmpDir}, WithGitignore(true))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	if w.shouldSkipByGitignore(filepath.Join(tmpDir, "anything")) {
		t.Error("should not skip when no .gitignore file exists")
	}
}

// dirEntry is a minimal os.DirEntry implementation for testing.
type dirEntry struct {
	name  string
	isDir bool
}

func (d *dirEntry) Name() string               { return d.name }
func (d *dirEntry) IsDir() bool                { return d.isDir }
func (d *dirEntry) Type() os.FileMode          { return os.ModeDir }
func (d *dirEntry) Info() (os.FileInfo, error) { return nil, nil }
