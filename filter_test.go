//nolint:testpackage,varnamelen // Tests need internal access; idiomatic short names in tests
package filewatcher

import (
	"os"
	"strings"
	"testing"
	"time"
)

// runFilterTests is a helper function that runs table-driven filter tests.

func runFilterTests(t *testing.T, filterName string, f Filter, tests filterTests) {
	t.Helper()
	t.Run(filterName, func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				if got := f(tt.event); got != tt.want {
					t.Errorf("%s() = %v, want %v", filterName, got, tt.want)
				}
			})
		}
	})
}

func runFilterTestsInline(t *testing.T, f Filter, tests filterTests) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := f(tt.event); got != tt.want {
				t.Errorf(
					"got %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

type filterTests []struct {
	name  string
	event Event
	want  bool
}

func ignoreDirTestCases() filterTests {
	return filterTests{
		{"main.go", testWriteEvent("/tmp/main.go"), true},
		{"vendor/pkg.go", testWriteEvent("/tmp/vendor/pkg.go"), false},
		{"pkg/vendor/lib.go", testWriteEvent("/tmp/pkg/vendor/lib.go"), false},
		{"node_modules/index.js", testWriteEvent("/tmp/node_modules/index.js"), false},
	}
}

func ignoreHiddenTestCases() filterTests {
	return filterTests{
		{".hidden", testWriteEvent("/tmp/.hidden"), false},
		{".git/config", testWriteEvent("/tmp/.git/config"), false},
		{".env", testWriteEvent("/tmp/.env"), false},
		{"main.go", testWriteEvent("/tmp/main.go"), true},
	}
}

func ignoreExtTestCases() filterTests {
	return filterTests{
		{"go file", testWriteEvent("/tmp/main.go"), true},
		{"log file", testWriteEvent("/tmp/app.log"), false},
		{"tmp file", testWriteEvent("/tmp/cache.tmp"), false},
	}
}

func regexTestCases() filterTests {
	return filterTests{
		{"go file", testWriteEvent("/tmp/main.go"), true},
		{"txt file", testWriteEvent("/tmp/readme.txt"), false},
		{"go file in subdir", testWriteEvent("/tmp/pkg/helper.go"), true},
	}
}

func extensionsTestCases() filterTests {
	return filterTests{
		{"go file", testWriteEvent("/tmp/main.go"), true},
		{"md file", testWriteEvent("/tmp/readme.md"), true},
		{"txt file", testWriteEvent("/tmp/notes.txt"), false},
		{"go file uppercase ext", testWriteEvent("/tmp/main.GO"), true},
	}
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterExtensions(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterExtensions", FilterExtensions(".go", ".md"), extensionsTestCases())
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterIgnoreExtensions(t *testing.T) {
	t.Parallel()
	runFilterTests(
		t,
		"FilterIgnoreExtensions",
		FilterIgnoreExtensions(".log", ".tmp"),
		ignoreExtTestCases(),
	)
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterIgnoreDirs(t *testing.T) {
	t.Parallel()
	runFilterTests(
		t,
		"FilterIgnoreDirs",
		FilterIgnoreDirs("vendor", "node_modules"),
		ignoreDirTestCases(),
	)
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterIgnoreHidden(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterIgnoreHidden", FilterIgnoreHidden(), ignoreHiddenTestCases())
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterOperations(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterOperations", FilterOperations(Write, Create), []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"write",
			testEvent("/tmp/test.txt", Write),
			true,
		},
		{
			"create",
			testEvent("/tmp/test.txt", Create),
			true,
		},
		{
			"remove",
			testEvent("/tmp/test.txt", Remove),
			false,
		},
		{
			"rename",
			testEvent("/tmp/test.txt", Rename),
			false,
		},
	})
}

func TestFilterExcludePaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create test paths
	specificFile := tmpDir + "/specific.go"
	specificDir := tmpDir + "/excluded_dir"
	otherFile := tmpDir + "/other.go"

	filter := FilterExcludePaths(specificFile, specificDir)

	cases := []struct {
		name string
		path string
		want bool
	}{
		{"excluded file", specificFile, false},
		{"excluded dir", specificDir, false},
		{"other file", otherFile, true},
		{"nonexistent path", tmpDir + "/nonexistent", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			event := Event{Path: tc.path, Op: Write, Timestamp: time.Now(), IsDir: false}
			if got := filter(event); got != tc.want {
				t.Errorf("FilterExcludePaths(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestFilterNotOperations(t *testing.T) {
	t.Parallel()

	f := FilterNotOperations(Remove)

	if f(testEvent("/tmp/test.txt", Remove)) {
		t.Error("expected Remove to be filtered out")
	}

	if !f(testWriteEvent("/tmp/test.txt")) {
		t.Error("expected Write to pass")
	}
}

//nolint:tparallel // Subtests call t.Parallel() inside runFilterTests helper
func TestFilterGlob(t *testing.T) {
	t.Parallel()

	runFilterTests(t, "FilterGlob", FilterGlob("*.go"), []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"go file",
			testWriteEvent("/tmp/main.go"),
			true,
		},
		{
			"txt file",
			testWriteEvent("/tmp/readme.txt"),
			false,
		},
	})
}

func TestFilterAnd(t *testing.T) {
	t.Parallel()

	f := FilterAnd(
		FilterExtensions(".go"),
		FilterOperations(Write),
	)

	if !f(testWriteEvent("main.go")) {
		t.Error("expected .go Write to pass")
	}

	if f(testEvent("main.go", Remove)) {
		t.Error("expected .go Remove to be filtered")
	}

	if f(testWriteEvent("main.txt")) {
		t.Error("expected .txt Write to be filtered")
	}
}

func TestFilterOr(t *testing.T) {
	t.Parallel()

	f := FilterOr(
		FilterExtensions(".go"),
		FilterExtensions(".md"),
	)

	if !f(testWriteEvent("main.go")) {
		t.Error("expected .go to pass")
	}

	if !f(testWriteEvent("readme.md")) {
		t.Error("expected .md to pass")
	}

	if f(testWriteEvent("main.txt")) {
		t.Error("expected .txt to be filtered")
	}
}

func TestFilterNot(t *testing.T) {
	t.Parallel()

	f := FilterNot(FilterExtensions(".go"))

	if f(testWriteEvent("main.go")) {
		t.Error("expected .go to be filtered after inversion")
	}

	if !f(testWriteEvent("main.txt")) {
		t.Error("expected .txt to pass after inversion")
	}
}

func TestFilterMinSize(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	smallFile := tmpDir + "/small.txt"

	err := os.WriteFile(smallFile, []byte("hi"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	largeFile := tmpDir + "/large.txt"

	err = os.WriteFile(largeFile, make([]byte, 1000), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	f := FilterMinSize(100)

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"small file",
			testWriteEvent(smallFile),
			false,
		},
		{
			"large file",
			testWriteEvent(largeFile),
			true,
		},
		{
			"directory",
			Event{Path: tmpDir, Op: Create, Timestamp: time.Now(), IsDir: true},
			true,
		},
		{
			"nonexistent file",
			testWriteEvent("/nonexistent/file.txt"),
			false,
		},
	}

	runFilterTestsInline(t, f, tests)
}

func TestFilterRegex(t *testing.T) {
	t.Parallel()
	runFilterTestsInline(t, FilterRegex(`\.go$`), regexTestCases())
}

func TestFilterMaxSize(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	smallFile := tmpDir + "/small.txt"
	if err := os.WriteFile(smallFile, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	largeFile := tmpDir + "/large.txt"
	if err := os.WriteFile(largeFile, make([]byte, 1000), 0o600); err != nil {
		t.Fatal(err)
	}

	f := FilterMaxSize(100)

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"small file",
			testWriteEvent(smallFile),
			true,
		},
		{
			"large file",
			testWriteEvent(largeFile),
			false,
		},
		{
			"directory",
			Event{Path: tmpDir, Op: Create, Timestamp: time.Now(), IsDir: true},
			true,
		},
		{
			"nonexistent file",
			testWriteEvent("/nonexistent/file.txt"),
			false,
		},
	}

	runFilterTestsInline(t, f, tests)
}

func TestFilterMinAge(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	oldFile := tmpDir + "/old.txt"
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Set modification time to 1 hour ago
	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	recentFile := tmpDir + "/recent.txt"
	if err := os.WriteFile(recentFile, []byte("recent"), 0o600); err != nil {
		t.Fatal(err)
	}

	f := FilterMinAge(10 * time.Minute)

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"old file passes",
			testWriteEvent(oldFile),
			true,
		},
		{
			"recent file filtered",
			testWriteEvent(recentFile),
			false,
		},
		{
			"directory passes through",
			Event{Path: tmpDir, Op: Create, Timestamp: time.Now(), IsDir: true},
			true,
		},
		{
			"nonexistent file filtered",
			testWriteEvent("/nonexistent/file.txt"),
			false,
		},
	}

	runFilterTestsInline(t, f, tests)
}

func TestFilterModifiedSince(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file modified 1 minute ago
	recentFile := tmpDir + "/recent.txt"
	if err := os.WriteFile(recentFile, []byte("recent"), 0o600); err != nil {
		t.Fatal(err)
	}

	recentTime := time.Now().Add(-1 * time.Minute)
	if err := os.Chtimes(recentFile, recentTime, recentTime); err != nil {
		t.Fatal(err)
	}

	// Create a file modified 1 hour ago
	oldFile := tmpDir + "/old.txt"
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	cutoff := time.Now().Add(-5 * time.Minute)
	f := FilterModifiedSince(cutoff)

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			"recent file passes",
			testWriteEvent(recentFile),
			true,
		},
		{
			"old file filtered",
			testWriteEvent(oldFile),
			false,
		},
		{
			"directory passes through",
			Event{Path: tmpDir, Op: Create, Timestamp: time.Now(), IsDir: true},
			true,
		},
		{
			"nonexistent file filtered",
			testWriteEvent("/nonexistent/file.txt"),
			false,
		},
	}

	runFilterTestsInline(t, f, tests)
}

func TestEvent_String(t *testing.T) {
	t.Parallel()

	e := fixedTimeEvent("/tmp/test.go", Write, 0)

	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}

	if !strings.Contains(s, "WRITE") {
		t.Errorf("expected string to contain WRITE, got %q", s)
	}
}

func TestOp_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op   Op
		want string
	}{
		{Create, "CREATE"},
		{Write, "WRITE"},
		{Remove, "REMOVE"},
		{Rename, "RENAME"},
		{Op(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("Op(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func BenchmarkFilterExtensions(b *testing.B) {
	f := FilterExtensions(".go", ".md", ".txt")
	event := testWriteEvent("/tmp/main.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}

func BenchmarkFilterIgnoreDirs(b *testing.B) {
	f := FilterIgnoreDirs("vendor", "node_modules", ".git")
	event := testWriteEvent("/tmp/vendor/pkg/lib.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}

func BenchmarkFilterGlob(b *testing.B) {
	f := FilterGlob("*.go")
	event := testWriteEvent("/tmp/main.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}

func BenchmarkFilterRegex(b *testing.B) {
	f := FilterRegex(`\.go$`)
	event := testWriteEvent("/tmp/main.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}

func BenchmarkFilterAnd(b *testing.B) {
	f := FilterAnd(
		FilterExtensions(".go"),
		FilterOperations(Write),
	)
	event := testWriteEvent("/tmp/main.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}

func BenchmarkFilterOr(b *testing.B) {
	f := FilterOr(
		FilterExtensions(".go"),
		FilterExtensions(".md"),
	)
	event := testWriteEvent("/tmp/main.go")

	b.ResetTimer()

	for i := range b.N {
		f(event)

		_ = i
	}
}
