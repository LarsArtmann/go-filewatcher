//nolint:varnamelen // idiomatic short names: w (watcher), t (testing)
package filewatcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Poll Loop Case-Awareness Tests (T7) ---

func TestPollWalkDir_UsesCanonicalKeys_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "MyDir", "SubFolder")
	mkdirErr := os.MkdirAll(subDir, 0o750)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	snapshot := make(map[string]fileState)
	watcher.pollWalkDir(tmpDir, snapshot)

	if len(snapshot) == 0 {
		t.Fatal("expected non-empty snapshot")
	}

	for key := range snapshot {
		if key != strings.ToLower(key) {
			t.Errorf("snapshot key %q should be lowercased (canonical) on case-insensitive", key)
		}
	}
}

func TestPollWalkDir_UsesCanonicalKeys_CaseSensitive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "MyDir", "SubFolder")
	mkdirErr := os.MkdirAll(subDir, 0o750)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	snapshot := make(map[string]fileState)
	watcher.pollWalkDir(tmpDir, snapshot)

	foundOriginal := false

	for key := range snapshot {
		if strings.Contains(key, "MyDir") {
			foundOriginal = true
		}
	}

	if !foundOriginal {
		t.Errorf("expected to find 'MyDir' (original case) in snapshot keys on case-sensitive, got: %v", snapshot)
	}
}

func TestPollWalkDir_PreservesOriginalPathInState(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "TestFile.txt")
	writeErr := os.WriteFile(testFile, []byte("content"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	snapshot := make(map[string]fileState)
	watcher.pollWalkDir(tmpDir, snapshot)

	found := false

	for _, state := range snapshot {
		if state.path == testFile {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected fileState.path to preserve original path %q", testFile)
	}
}

func TestPollDetectChanges_NoPhantomEventsOnCaseInsensitive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "File.txt")
	writeErr := os.WriteFile(testFile, []byte("initial"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb),
		WithCaseSensitivity(CaseInsensitive),
		WithPolling(true),
		WithPollInterval(100*time.Millisecond),
	)

	ctx := setupTestContext(t, 5*time.Second)

	events, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	time.Sleep(250 * time.Millisecond)

	newFile := filepath.Join(tmpDir, "file.txt")
	renameErr := os.Rename(testFile, newFile)
	if renameErr != nil {
		t.Fatal(renameErr)
	}

	var eventCount int

	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}

			base := filepath.Base(event.Path)
			if base == "File.txt" || base == "file.txt" {
				eventCount++
			}
		case <-timer.C:
			if eventCount > 1 {
				t.Errorf("expected at most 1 event for case-only rename on case-insensitive, got %d", eventCount)
			}

			return
		}
	}
}

// --- Gitignore Case-Awareness Tests (T8) ---

func TestGitignore_AncestorPrefix_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreDir := filepath.Join(tmpDir, "Project")
	mkdirErr := os.MkdirAll(gitignoreDir, 0o750)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignorePath := filepath.Join(gitignoreDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("*.tmp\n"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	watcher.loadGitignoreForDir(gitignoreDir)

	checkedPath := filepath.Join(strings.ToLower(tmpDir), "project", "data", "file.tmp")
	if !watcher.shouldSkipByGitignore(checkedPath) {
		t.Errorf("shouldSkipByGitignore(%q) = false, want true (case-insensitive ancestor match)", checkedPath)
	}
}

func TestGitignore_AncestorPrefix_CaseSensitiveDoesNotMatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreDir := filepath.Join(tmpDir, "Project")
	mkdirErr := os.MkdirAll(gitignoreDir, 0o750)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignorePath := filepath.Join(gitignoreDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("*.tmp\n"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	watcher.loadGitignoreForDir(gitignoreDir)

	checkedPath := filepath.Join(tmpDir, "project", "data", "file.tmp")
	if watcher.shouldSkipByGitignore(checkedPath) {
		t.Errorf("shouldSkipByGitignore(%q) = true, want false (case-sensitive, different case)", checkedPath)
	}
}

func TestGitignore_RuleMatches_CaseInsensitiveDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("secret\n"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	watcher.loadGitignoreForDir(tmpDir)

	secretPath := filepath.Join(tmpDir, "secret")
	if !watcher.shouldSkipByGitignore(secretPath) {
		t.Errorf("shouldSkipByGitignore(%q) = false, want true (pattern match)", secretPath)
	}

	normalPath := filepath.Join(tmpDir, "normal.go")
	if watcher.shouldSkipByGitignore(normalPath) {
		t.Errorf("shouldSkipByGitignore(%q) = true, want false (no pattern match)", normalPath)
	}
}
