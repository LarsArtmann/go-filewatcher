package filewatcher

import (
	"os"
	"path/filepath"
	"testing"
)

// setupGitignoreTest creates a temp directory with a .gitignore and subdirectory.
// Returns the watcher and the subdirectory path.
func setupGitignoreTest(
	t *testing.T,
	subDirName, gitignoreContent string,
	gitignoreEnabled bool,
) (*Watcher, string) {
	t.Helper()

	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, subDirName)

	mkdirErr := os.MkdirAll(subDir, 0o755) //nolint:gosec // standard temp directory permissions
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	if gitignoreContent != "" {
		writeErr := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), testFilePermission)
		if writeErr != nil {
			t.Fatal(writeErr)
		}
	}

	watcher, err := New([]string{tmpDir}, WithGitignore(gitignoreEnabled))
	if err != nil {
		t.Fatal(err)
	}

	return watcher, subDir
}

func TestGitignore_SkipsIgnoredDir(t *testing.T) {
	t.Parallel()

	watcher, buildDir := setupGitignoreTest(t, "build", "build/\n", true)

	defer func() { _ = watcher.Close() }()

	watcher.mu.Lock()

	walkErr := watcher.walkDirFunc(buildDir, &dirEntry{name: "build", isDir: true}, nil)
	watcher.mu.Unlock()

	if walkErr == nil {
		t.Error("expected SkipDir for gitignored directory")
	}
}

func TestGitignore_DoesNotSkipNonIgnored(t *testing.T) {
	t.Parallel()

	watcher, srcDir := setupGitignoreTest(t, "src", "build/\n", true)

	defer func() { _ = watcher.Close() }()

	watcher.mu.Lock()

	walkErr := watcher.walkDirFunc(srcDir, &dirEntry{name: "src", isDir: true}, nil)
	watcher.mu.Unlock()

	if walkErr != nil {
		t.Errorf("expected no error for non-ignored directory, got %v", walkErr)
	}
}

func TestGitignore_Disabled(t *testing.T) {
	t.Parallel()

	watcher, buildDir := setupGitignoreTest(t, "my-output", "my-output/\n", false)

	defer func() { _ = watcher.Close() }()

	watcher.mu.Lock()

	walkErr := watcher.walkDirFunc(buildDir, &dirEntry{name: "my-output", isDir: true}, nil)
	watcher.mu.Unlock()

	if walkErr != nil {
		t.Errorf("expected no error when gitignore disabled, got %v", walkErr)
	}
}

func TestGitignore_NoGitignoreFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	watcher, err := New([]string{tmpDir}, WithGitignore(true))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	if watcher.shouldSkipByGitignore(filepath.Join(tmpDir, "anything")) {
		t.Error("should not skip when no .gitignore file exists")
	}
}

// dirEntry is a minimal os.DirEntry implementation for testing.
type dirEntry struct {
	name  string
	isDir bool
}

func (d *dirEntry) Name() string      { return d.name }
func (d *dirEntry) IsDir() bool       { return d.isDir }
func (d *dirEntry) Type() os.FileMode { return os.ModeDir }

func (d *dirEntry) Info() (os.FileInfo, error) {
	//nolint:nilnil // test stub — callers never use the returned values
	return nil, nil
}
