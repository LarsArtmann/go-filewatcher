//nolint:testpackage // Tests need internal access for phantom types
package filewatcher

import (
	"testing"
)

func TestLogSubstring_String(t *testing.T) {
	t.Parallel()

	ls := LogSubstring("hello world")
	if ls.String() != "hello world" {
		t.Errorf("LogSubstring.String() = %q, want %q", ls.String(), "hello world")
	}
}

func TestEventPath_String(t *testing.T) {
	t.Parallel()

	ep := EventPath("/tmp/test.go")
	if ep.String() != "/tmp/test.go" {
		t.Errorf("EventPath.String() = %q, want %q", ep.String(), "/tmp/test.go")
	}
}

func TestEventPath_Base(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input EventPath
		want  string
	}{
		{"/home/user/file.go", "file.go"},
		{"/home/user/", "user"},
		{"file.go", "file.go"},
	}

	for _, tt := range tests {
		if got := tt.input.Base(); got != tt.want {
			t.Errorf("EventPath(%q).Base() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventPath_Dir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input EventPath
		want  EventPath
	}{
		{"/home/user/file.go", "/home/user"},
		{"/home/user/", "/home/user"},
		{"file.go", "."},
	}

	for _, tt := range tests {
		if got := tt.input.Dir(); got != tt.want {
			t.Errorf("EventPath(%q).Dir() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventPath_Ext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input EventPath
		want  string
	}{
		{"/home/user/file.go", ".go"},
		{"/home/user/README", ""},
		{"/home/user/file.test.go", ".go"},
	}

	for _, tt := range tests {
		if got := tt.input.Ext(); got != tt.want {
			t.Errorf("EventPath(%q).Ext() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventPath_Join(t *testing.T) {
	t.Parallel()

	tests := []struct {
		base  EventPath
		elems []string
		want  EventPath
	}{
		{"/home/user", []string{"docs", "readme.md"}, "/home/user/docs/readme.md"},
		{"/home/user", []string{"file.go"}, "/home/user/file.go"},
		{"/home/user", []string{}, "/home/user"},
		{"/home/user", []string{"a", "b", "c"}, "/home/user/a/b/c"},
	}

	for _, tt := range tests {
		if got := tt.base.Join(tt.elems...); got != tt.want {
			t.Errorf("EventPath(%q).Join(%v) = %q, want %q", tt.base, tt.elems, got, tt.want)
		}
	}
}
