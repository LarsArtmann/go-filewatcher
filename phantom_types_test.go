//nolint:testpackage // Tests need internal access for phantom types
package filewatcher

import (
	"testing"
)

func TestLogSubstring_String(t *testing.T) {
	t.Parallel()

	ls := NewLogSubstring("hello world")
	if ls.String() != "hello world" {
		t.Errorf("LogSubstring.String() = %q, want %q", ls.String(), "hello world")
	}
}

func TestEventPath_String(t *testing.T) {
	t.Parallel()

	ep := NewEventPath("/tmp/test.go")
	if ep.String() != "/tmp/test.go" {
		t.Errorf("EventPath.String() = %q, want %q", ep.String(), "/tmp/test.go")
	}
}

// pathTestCase is a reusable test case for path-based tests.
type pathTestCase struct {
	input EventPath
	want  string
}

// runPathTests executes a table-driven test for EventPath methods.
func runPathTests(t *testing.T, tests []pathTestCase, fn func(EventPath) string) {
	t.Helper()

	for _, tt := range tests {
		if got := fn(tt.input); got != tt.want {
			t.Errorf("%q = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventPath_Base(t *testing.T) {
	t.Parallel()

	runPathTests(t, []pathTestCase{
		{NewEventPath("/home/user/file.go"), "file.go"},
		{NewEventPath("/home/user/"), "user"},
		{NewEventPath("file.go"), "file.go"},
	}, func(p EventPath) string { return p.Base() })
}

func TestEventPath_Dir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input EventPath
		want  EventPath
	}{
		{NewEventPath("/home/user/file.go"), NewEventPath("/home/user")},
		{NewEventPath("/home/user/"), NewEventPath("/home/user")},
		{NewEventPath("file.go"), NewEventPath(".")},
	}

	for _, tt := range tests {
		if got := tt.input.Dir(); got != tt.want {
			t.Errorf("EventPath(%q).Dir() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventPath_Ext(t *testing.T) {
	t.Parallel()

	runPathTests(t, []pathTestCase{
		{NewEventPath("/home/user/file.go"), ".go"},
		{NewEventPath("/home/user/README"), ""},
		{NewEventPath("/home/user/file.test.go"), ".go"},
	}, func(p EventPath) string { return p.Ext() })
}

func TestEventPath_Join(t *testing.T) {
	t.Parallel()

	tests := []struct {
		base  EventPath
		elems []string
		want  EventPath
	}{
		{
			NewEventPath("/home/user"),
			[]string{"docs", "readme.md"},
			NewEventPath("/home/user/docs/readme.md"),
		},
		{NewEventPath("/home/user"), []string{"file.go"}, NewEventPath("/home/user/file.go")},
		{NewEventPath("/home/user"), []string{}, NewEventPath("/home/user")},
		{
			NewEventPath("/home/user"),
			[]string{"a", "b", "c"},
			NewEventPath("/home/user/a/b/c"),
		},
	}

	for _, tt := range tests {
		if got := tt.base.Join(tt.elems...); got != tt.want {
			t.Errorf("EventPath(%q).Join(%v) = %q, want %q", tt.base, tt.elems, got, tt.want)
		}
	}
}
