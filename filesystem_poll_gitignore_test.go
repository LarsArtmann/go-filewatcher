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

	// Create a subdirectory with mixed case
	subDir := filepath.Join(tmpDir, "MyDir", "SubFolder")
	mkdirErr := os.MkdirAll(subDir, 0o755)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	snapshot := make(map[string]fileState)
	w.pollWalkDir(tmpDir, snapshot)

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
	mkdirErr := os.MkdirAll(subDir, 0o755)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	snapshot := make(map[string]fileState)
	w.pollWalkDir(tmpDir, snapshot)

	// On case-sensitive, paths should be preserved (not lowercased)
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
	writeErr := os.WriteFile(testFile, []byte("content"), 0o644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	snapshot := make(map[string]fileState)
	w.pollWalkDir(tmpDir, snapshot)

	// The state should contain the original path, not the lowercased key
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

	// Create a file
	testFile := filepath.Join(tmpDir, "File.txt")
	writeErr := os.WriteFile(testFile, []byte("initial"), 0o644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb),
		WithCaseSensitivity(CaseInsensitive),
		WithPolling(true),
		WithPollInterval(100*time.Millisecond),
	)

	ctx := setupTestContext(t, 5*time.Second)

	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Wait for initial snapshot
	time.Sleep(250 * time.Millisecond)

	// Rename the file with different case only (case-insensitive rename)
	// On Linux (case-sensitive FS), this creates a genuinely different file.
	// But with CaseInsensitive mode, the poll loop keys should be the same.
	// We verify no phantom Create+Remove burst happens for the SAME logical file.
	newFile := filepath.Join(tmpDir, "file.txt")
	renameErr := os.Rename(testFile, newFile)
	if renameErr != nil {
		t.Fatal(renameErr)
	}

	// Collect events for a short window
	var eventCount int
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			// Only count events for our test file (ignore the initial-snapshot events)
			base := filepath.Base(event.Path)
			if base == "File.txt" || base == "file.txt" {
				eventCount++
			}
		case <-timer.C:
			// On case-insensitive, the canonical keys are identical, so
			// there should be at most 1 event (a Write for the modification),
			// not a Create+Remove pair.
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

	// Create a .gitignore in a mixed-case directory
	gitignoreDir := filepath.Join(tmpDir, "Project")
	mkdirErr := os.MkdirAll(gitignoreDir, 0o755)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignorePath := filepath.Join(gitignoreDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("*.tmp\n"), 0o644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	// Load the gitignore from the original-case directory
	w.loadGitignoreForDir(gitignoreDir)

	// Check a path with DIFFERENT case — should still match the ancestor prefix
	checkedPath := filepath.Join(strings.ToLower(tmpDir), "project", "data", "file.tmp")
	if !w.shouldSkipByGitignore(checkedPath) {
		t.Errorf("shouldSkipByGitignore(%q) = false, want true (case-insensitive ancestor match)", checkedPath)
	}
}

func TestGitignore_AncestorPrefix_CaseSensitiveDoesNotMatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreDir := filepath.Join(tmpDir, "Project")
	mkdirErr := os.MkdirAll(gitignoreDir, 0o755)
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	gitignorePath := filepath.Join(gitignoreDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("*.tmp\n"), 0o644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	w.loadGitignoreForDir(gitignoreDir)

	// Path with different case should NOT match ancestor on case-sensitive
	checkedPath := filepath.Join(tmpDir, "project", "data", "file.tmp")
	if w.shouldSkipByGitignore(checkedPath) {
		t.Errorf("shouldSkipByGitignore(%q) = true, want false (case-sensitive, different case)", checkedPath)
	}
}

func TestGitignore_RuleMatches_CaseInsensitiveDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// .gitignore at root with a pattern
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	writeErr := os.WriteFile(gitignorePath, []byte("secret\n"), 0o644)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	// Load gitignore for tmpDir (the canonical key is lowercased)
	w.loadGitignoreForDir(tmpDir)

	// A file matching the pattern should be skipped
	secretPath := filepath.Join(tmpDir, "secret")
	if !w.shouldSkipByGitignore(secretPath) {
		t.Errorf("shouldSkipByGitignore(%q) = false, want true (pattern match)", secretPath)
	}

	// A non-matching file should not be skipped
	normalPath := filepath.Join(tmpDir, "normal.go")
	if w.shouldSkipByGitignore(normalPath) {
		t.Errorf("shouldSkipByGitignore(%q) = true, want false (no pattern match)", normalPath)
	}
}
