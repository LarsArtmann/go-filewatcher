package filewatcher

import (
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func FuzzFilterRegex(f *testing.F) {
	// Seed corpus: valid regex patterns and paths
	seeds := []struct {
		pattern string
		path    string
	}{
		{`\.go$`, "/home/user/main.go"},
		{`\.go$`, "/home/user/readme.md"},
		{`test`, "/home/user/test_file.txt"},
		{`\.(go|rs)$`, "/home/user/main.rs"},
		{``, "/home/user/file"},
		{`^/tmp/`, "/tmp/test.go"},
		{`[a-z]+`, "/home/ABC/file.go"},
	}

	for _, s := range seeds {
		f.Add(s.pattern, s.path)
	}

	f.Fuzz(func(t *testing.T, pattern, path string) {
		t.Parallel()

		// FilterRegex with MustCompile will panic on invalid patterns
		// This is documented behavior, so we skip invalid patterns
		defer func() {
			_ = recover()
		}()

		filter := FilterRegex(pattern)

		_ = filter(testWriteEvent(path))
	})
}

func FuzzFilterExtensions(f *testing.F) {
	seeds := []struct {
		ext  string
		path string
	}{
		{".go", "/home/user/main.go"},
		{".go", "/home/user/main.rs"},
		{"", "/home/user/file"},
		{".RS", "/home/user/main.rs"},
		{".tar.gz", "/home/user/archive.tar.gz"},
	}

	for _, s := range seeds {
		f.Add(s.ext, s.path)
	}

	f.Fuzz(func(t *testing.T, ext, path string) {
		t.Parallel()

		filter := FilterExtensions(ext)

		_ = filter(testWriteEvent(path))
	})
}

func FuzzFilterIgnoreGlobs(f *testing.F) {
	seeds := []struct {
		pattern string
		path    string
	}{
		{"*.log", "/home/user/app.log"},
		{"*.log", "/home/user/main.go"},
		{".*", "/home/user/.hidden"},
		{"test_*", "/home/user/test_file.txt"},
	}

	for _, s := range seeds {
		f.Add(s.pattern, s.path)
	}

	f.Fuzz(func(t *testing.T, pattern, path string) {
		t.Parallel()

		filter := FilterIgnoreGlobs(pattern)

		_ = filter(testWriteEvent(path))
	})
}

func FuzzOpUnmarshalText(f *testing.F) {
	seeds := []string{
		"CREATE",
		"WRITE",
		"REMOVE",
		"RENAME",
		"UNKNOWN",
		"",
		"create",
		"123",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		t.Parallel()

		var op Op

		_ = op.UnmarshalText([]byte(text))
	})
}

func FuzzFilterMinSize(f *testing.F) {
	seeds := []struct {
		minSize int64
		path    string
	}{
		{0, "/home/user/file.go"},
		{1024, "/home/user/large.bin"},
		{-1, "/home/user/file.go"},
	}

	for _, s := range seeds {
		f.Add(s.minSize, s.path)
	}

	f.Fuzz(func(t *testing.T, minSize int64, path string) {
		t.Parallel()

		filter := FilterMinSize(minSize)

		_ = filter(testWriteEvent(path))
	})
}

// FuzzPathKey verifies pathKey() is robust and correct for arbitrary Unicode
// input: combining marks, emoji, zero-width joiner sequences, and invalid UTF-8.
// Invariants checked:
//  1. Never panics (graceful for any byte string).
//  2. Deterministic (same input → same key).
//  3. Idempotent (canonicalizing an already-canonical key is a no-op).
//  4. case-insensitive key == lowercased case-sensitive key (valid UTF-8 only).
//  5. NFC equivalence: pathKey(P) == pathKey(norm.NFC.String(P)) (valid UTF-8 only).
func FuzzPathKey(f *testing.F) {
	seeds := []string{
		"/home/user/file.go",
		"/home/user/café/file.go",                       // NFC
		"/home/user/cafe\u0301/file.go",                 // NFD
		"/Users/CAFÉ/FILE.GO",
		"",
		"/",
		".",
		"..",
		"/path/with/emoji/😀/file.txt",
		"/family/\U0001F468\u200D\U0001F469\u200D\U0001F467/", // ZWJ family
		"/combining/a\u0308b",                            // a with combining diaeresis
		"\\x80\\xffinvalid",                              // invalid UTF-8 escapes
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, path string) {
		t.Parallel()

		cs := &Watcher{effectiveCaseSensitivity: CaseSensitive}
		ci := &Watcher{effectiveCaseSensitivity: CaseInsensitive}

		// Invariant 1+2: never panics and is deterministic.
		keyCS := cs.pathKey(path)
		keyCI := ci.pathKey(path)

		if cs.pathKey(path) != keyCS {
			t.Errorf("pathKey not deterministic (case-sensitive) for %q", path)
		}

		if ci.pathKey(path) != keyCI {
			t.Errorf("pathKey not deterministic (case-insensitive) for %q", path)
		}

		// Invariant 3: idempotent on the canonical key.
		if cs.pathKey(keyCS) != keyCS {
			t.Errorf("pathKey not idempotent (case-sensitive): %q -> %q -> %q", path, keyCS, cs.pathKey(keyCS))
		}

		// Invariants 4+5 hold only for valid UTF-8: NFC/ToLower semantics are
		// undefined on invalid byte sequences.
		if !utf8.ValidString(path) {
			return
		}

		if keyCI != strings.ToLower(keyCS) {
			t.Errorf("case-insensitive key should be lowercased case-sensitive key: got %q, want %q",
				keyCI, strings.ToLower(keyCS))
		}

		if keyCS != cs.pathKey(norm.NFC.String(path)) {
			t.Errorf("pathKey should equal pathKey(NFC(input)): got %q, want %q",
				keyCS, cs.pathKey(norm.NFC.String(path)))
		}
	})
}
