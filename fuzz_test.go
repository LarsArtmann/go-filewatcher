package filewatcher

import (
	"testing"
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
