//nolint:varnamelen // idiomatic short names: w (watcher), fb (fakeBackend)
package filewatcher

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFilesystemCaseSensitivity_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode FilesystemCaseSensitivity
		want string
	}{
		{CaseSensitive, "case-sensitive"},
		{CaseInsensitive, "case-insensitive"},
		{CaseSensitivityAuto, "auto"},
		{FilesystemCaseSensitivity(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestResolveCaseSensitivity_AutoResolvesByPlatform(t *testing.T) {
	t.Parallel()

	result := resolveCaseSensitivity(CaseSensitivityAuto)

	switch runtime.GOOS {
	case "windows", "darwin":
		assertEqual(t, "auto on "+runtime.GOOS, result, CaseInsensitive)
	default:
		assertEqual(t, "auto on "+runtime.GOOS, result, CaseSensitive)
	}
}

func TestResolveCaseSensitivity_ExplicitModesArePreserved(t *testing.T) {
	t.Parallel()

	assertEqual(t, "explicit sensitive", resolveCaseSensitivity(CaseSensitive), CaseSensitive)
	assertEqual(t, "explicit insensitive", resolveCaseSensitivity(CaseInsensitive), CaseInsensitive)
}

func TestPathKey_CaseSensitivePreservesCase(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	path := "/home/user/MyDir/File.GO"

	if got := w.pathKey(path); got != path {
		t.Errorf("pathKey(%q) = %q, want %q (case-sensitive should preserve case)", path, got, path)
	}
}

func TestPathKey_CaseInsensitiveLowercases(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseInsensitive}

	path := "/home/user/MyDir/File.GO"
	want := strings.ToLower(path)

	if got := w.pathKey(path); got != want {
		t.Errorf("pathKey(%q) = %q, want %q (case-insensitive should lowercase)", path, got, want)
	}
}

func TestTryAddPath_DeduplicatesCaseInsensitive(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))
	// effectiveCaseSensitivity is resolved in New(); verify it was set
	assertEqual(t, "effectiveCaseSensitivity", watcher.effectiveCaseSensitivity, CaseInsensitive)

	// Add a path with mixed case
	watcher.tryAddPath(filepath.Join(tmpDir, "MyDir"))
	// Try to add same path with different case — should be deduplicated
	watcher.tryAddPath(filepath.Join(tmpDir, "mydir"))

	list := watcher.WatchList()
	if len(list) != 1 {
		t.Errorf("expected 1 watched path (case-insensitive dedup), got %d: %v", len(list), list)
	}
}

func TestTryAddPath_KeepsDistinctCaseSensitive(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseSensitive))

	upper := filepath.Join(tmpDir, "MyDir")
	lower := filepath.Join(tmpDir, "mydir")

	watcher.tryAddPath(upper)
	watcher.tryAddPath(lower)

	list := watcher.WatchList()
	if len(list) != 2 {
		t.Errorf("expected 2 watched paths (case-sensitive keeps distinct), got %d: %v", len(list), list)
	}
}

func TestRemove_CaseAwareRemovesSubtree(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	watcher := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))

	// Simulate multiple watched paths under a root
	sub1 := filepath.Join(tmpDir, "Project")
	sub2 := filepath.Join(sub1, "src")

	watcher.tryAddPath(sub1)
	watcher.tryAddPath(sub2)

	// Remove with different case — should remove all subtrees
	err := watcher.Remove(strings.ToLower(sub1))
	if err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	list := watcher.WatchList()
	if len(list) != 0 {
		t.Errorf("expected 0 paths after case-insensitive Remove, got %d: %v", len(list), list)
	}
}

func TestGetDebounceKey_CaseInsensitiveCollapsesCase(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseInsensitive}

	key1 := w.getDebounceKey("/path/to/File.go")
	key2 := w.getDebounceKey("/path/to/file.GO")

	if key1 != key2 {
		t.Errorf("debounce keys should be equal on case-insensitive: %q vs %q", key1, key2)
	}
}

func TestGetDebounceKey_CaseSensitiveKeepsDistinct(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	key1 := w.getDebounceKey("/path/to/File.go")
	key2 := w.getDebounceKey("/path/to/file.GO")

	if key1 == key2 {
		t.Errorf("debounce keys should differ on case-sensitive: both %q", key1)
	}
}

func TestShouldExcludePath_CaseInsensitive(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	excluded := filepath.Join(tmpDir, "Build")
	watcher := newTestWatcher(
		t, tmpDir, withBackend(fb),
		WithCaseSensitivity(CaseInsensitive),
		WithExcludePaths(excluded),
	)

	// Path with different case should be excluded
	other := filepath.Join(tmpDir, "build", "output")
	if !watcher.shouldExcludePath(other) {
		t.Errorf("shouldExcludePath(%q) = false, want true (case-insensitive match)", other)
	}
}

func TestShouldExcludePath_CaseSensitiveDoesNotMatchDifferentCase(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	excluded := filepath.Join(tmpDir, "Build")
	watcher := newTestWatcher(
		t, tmpDir, withBackend(fb),
		WithCaseSensitivity(CaseSensitive),
		WithExcludePaths(excluded),
	)

	// Path with different case should NOT be excluded on case-sensitive
	other := filepath.Join(tmpDir, "build", "output")
	if watcher.shouldExcludePath(other) {
		t.Errorf("shouldExcludePath(%q) = true, want false (case-sensitive, different case)", other)
	}
}

// --- NFC Normalization Tests ---

func TestPathKey_NFCNormalization_CaseSensitive(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	nfc := "/home/user/café/file.go"       // NFC: é = U+00E9 (2 bytes)
	nfd := "/home/user/cafe\u0301/file.go" // NFD: e + combining acute = U+0065 U+0301 (3 bytes)

	keyNFC := w.pathKey(nfc)
	keyNFD := w.pathKey(nfd)

	if keyNFC != keyNFD {
		t.Errorf("pathKey should normalize NFC/NFD on case-sensitive: NFC=%q (len %d) vs NFD=%q (len %d)",
			keyNFC, len(keyNFC), keyNFD, len(keyNFD))
	}

	// Verify the key is in NFC form (not NFD)
	if keyNFC != nfc {
		t.Errorf("pathKey should produce NFC output: got %q, want %q", keyNFC, nfc)
	}
}

func TestPathKey_NFCNormalization_CaseInsensitive(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseInsensitive}

	nfc := "/home/user/München/File.GO"       // NFC
	nfd := "/home/user/Mu\u0308nchen/File.go" // NFD (ü decomposed)

	keyNFC := w.pathKey(nfc)
	keyNFD := w.pathKey(nfd)

	if keyNFC != keyNFD {
		t.Errorf("pathKey should normalize NFC/NFD + lowercase: NFC=%q vs NFD=%q", keyNFC, keyNFD)
	}
}

func TestPathKey_NFCNormalization_ASCIIUnchanged(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	ascii := "/home/user/project/src/main.go"
	key := w.pathKey(ascii)

	// ASCII paths must be byte-identical after NFC normalization
	if key != ascii {
		t.Errorf("pathKey should not modify ASCII paths: got %q, want %q", key, ascii)
	}
}

func TestPathKey_NFCNormalization_Idempotent(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	nfc := "/data/café.txt"
	once := w.pathKey(nfc)
	twice := w.pathKey(once)

	if once != twice {
		t.Errorf("pathKey should be idempotent on NFC input: once=%q, twice=%q", once, twice)
	}
}

func TestExcludePath_NFCMatchesNFD(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	// Exclude path in NFC form
	nfcExcluded := filepath.Join(tmpDir, "café")
	nfdPath := filepath.Join(tmpDir, "cafe\u0301")

	watcher := newTestWatcher(
		t, tmpDir, withBackend(fb),
		WithCaseSensitivity(CaseSensitive),
		WithExcludePaths(nfcExcluded),
	)

	// NFD form of the same path should be excluded (NFC normalization)
	if !watcher.shouldExcludePath(nfdPath) {
		t.Errorf("shouldExcludePath(%q) = false, want true (NFC exclude should match NFD path)", nfdPath)
	}
}

func TestDebounceKey_NFCNormalization(t *testing.T) {
	t.Parallel()

	w := &Watcher{effectiveCaseSensitivity: CaseSensitive}

	nfc := "/project/over/café.go"
	nfd := "/project/over/cafe\u0301.go"

	keyNFC := w.getDebounceKey(nfc)
	keyNFD := w.getDebounceKey(nfd)

	if keyNFC != keyNFD {
		t.Errorf("debounce keys should be equal after NFC normalization: NFC=%q vs NFD=%q", keyNFC, keyNFD)
	}
}

func TestStats_CaseSensitivityReflectsMode(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	w := newTestWatcher(t, tmpDir, withBackend(fb), WithCaseSensitivity(CaseInsensitive))
	stats := w.Stats()

	if stats.CaseSensitivity != "case-insensitive" {
		t.Errorf("Stats().CaseSensitivity = %q, want %q", stats.CaseSensitivity, "case-insensitive")
	}
}

func TestStats_CaseSensitivityReflectsAuto(t *testing.T) {
	t.Parallel()

	fb := newFakeBackend()
	tmpDir := t.TempDir()

	w := newTestWatcher(t, tmpDir, withBackend(fb)) // Auto default
	stats := w.Stats()

	expected := resolveCaseSensitivity(CaseSensitivityAuto).String()
	if stats.CaseSensitivity != expected {
		t.Errorf("Stats().CaseSensitivity = %q, want %q (auto-resolved)", stats.CaseSensitivity, expected)
	}
}

func TestFilterCaseInsensitive_LowercasesPath(t *testing.T) {
	t.Parallel()

	// Inner filter that matches an exact (lowercase) path
	inner := FilterRegex(`^/project/src/main\.go$`)
	wrapped := FilterCaseInsensitive(inner)

	// Mixed-case path should match when wrapped
	event := Event{Path: "/Project/Src/Main.go", Op: Write}
	if !wrapped(event) {
		t.Errorf("FilterCaseInsensitive should allow mixed-case path to match lowercase pattern")
	}
}

func TestFilterCaseInsensitive_NFCNormalizes(t *testing.T) {
	t.Parallel()

	// Inner filter with NFC path pattern
	inner := FilterRegex(`^/data/café\.txt$`)
	wrapped := FilterCaseInsensitive(inner)

	// NFD path should match NFC pattern through the wrapper
	nfdEvent := Event{Path: "/data/cafe\u0301.txt", Op: Write}
	if !wrapped(nfdEvent) {
		t.Errorf("FilterCaseInsensitive should normalize NFD path to match NFC pattern")
	}
}
