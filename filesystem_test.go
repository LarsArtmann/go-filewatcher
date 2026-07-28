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
