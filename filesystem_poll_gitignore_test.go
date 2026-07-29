package filewatcher

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/unicode/norm"
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

	// NOTE: On Linux (case-sensitive ext4), File.txt -> file.txt is a GENUINE
	// rename — the kernel sees two different files. This test verifies the poll
	// loop's canonical-key logic, but on case-sensitive filesystems it passes
	// due to timing leniency, not because canonicalization is proven. The
	// canonicalization correctness is independently verified by
	// TestPollWalkDir_NormalizesUnicodeKeysAndPreservesOriginalPath and
	// TestPathKey_NFCEquivalenceProperty. This test becomes meaningful on
	// macOS/Windows CI where the filesystem is genuinely case-insensitive.
	if runtime.GOOS != osDarwin && runtime.GOOS != osWindows {
		t.Log("Running on case-sensitive filesystem (ext4/XFS). " +
			"This test verifies timing behavior; canonicalization " +
			"correctness is proven by pathKey tests.")
	}

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "File.txt")

	writeErr := os.WriteFile(testFile, []byte("initial"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(
		t, tmpDir, withBackend(fb),
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

// TestGitignore_ExcludesDirDuringWalk_EndToEnd is a full integration test: it
// creates a real directory tree with a .gitignore, starts a real watcher in
// case-insensitive mode, and verifies the gitignored directory is never added
// to the watch list while a sibling is. This exercises the entire
// walk -> gitignore check -> tryAddPath pipeline.
func TestGitignore_ExcludesDirDuringWalk_EndToEnd(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	writeErr := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("excluded\n"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	excluded := filepath.Join(tmpDir, "excluded")
	included := filepath.Join(tmpDir, "included")

	for _, dir := range []string{excluded, included} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
	}

	watcher := newTestWatcher(t, tmpDir, WithCaseSensitivity(CaseInsensitive))

	ctx := setupTestContext(t, 5*time.Second)

	if _, err := watcher.Watch(ctx); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	watcher.mu.RLock()
	defer watcher.mu.RUnlock()

	if _, ok := watcher.watchListKeys[watcher.pathKey(excluded)]; ok {
		t.Errorf("gitignored dir %q should NOT be in watchListKeys", excluded)
	}

	if _, ok := watcher.watchListKeys[watcher.pathKey(included)]; !ok {
		t.Errorf("non-gitignored dir %q should be in watchListKeys", included)
	}
}

// TestPollWalkDir_NormalizesUnicodeKeysAndPreservesOriginalPath proves the
// Unicode pipeline: a file stored on disk with an NFD-decomposed filename is
// keyed in the poll snapshot by its NFC-normalized form (so NFD/NFC variants
// collide and don't produce phantom events), while the original NFD path is
// preserved in fileState for accurate event emission.
func TestPollWalkDir_NormalizesUnicodeKeysAndPreservesOriginalPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// NFD-decomposed filename (e + combining acute U+0301).
	nfdName := "cafe\u0301.txt"
	nfdPath := filepath.Join(tmpDir, nfdName)

	writeErr := os.WriteFile(nfdPath, []byte("data"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	fb := newFakeBackend()
	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	snapshot := make(map[string]fileState)
	watcher.pollWalkDir(tmpDir, snapshot)

	// The snapshot key must be the NFC-normalized form.
	nfcName := norm.NFC.String(nfdName)
	nfcKey := watcher.pathKey(filepath.Join(tmpDir, nfcName))

	state, ok := snapshot[nfcKey]
	if !ok {
		keys := make([]string, 0, len(snapshot))

		for k := range snapshot {
			keys = append(keys, k)
		}

		t.Fatalf("snapshot missing NFC key %q; have keys: %v", nfcKey, keys)
	}

	// The original NFD path must be preserved for event emission.
	if state.path != nfdPath {
		t.Errorf("fileState.path = %q, want original NFD path %q", state.path, nfdPath)
	}

	// The NFC and NFD forms must produce identical keys (no phantom entries).
	if watcher.pathKey(nfdPath) != nfcKey {
		t.Errorf("pathKey(NFD) != pathKey(NFC): %q != %q", watcher.pathKey(nfdPath), nfcKey)
	}
}
