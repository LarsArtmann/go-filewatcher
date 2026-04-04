package filewatcher

import (
	"testing"
	"time"
)

// runFilterTests is a helper function that runs table-driven filter tests.
func runFilterTests(t *testing.T, filterName string, f Filter, tests []struct {
	name  string
	event Event
	want  bool
},
) {
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

func TestFilterExtensions(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterExtensions", FilterExtensions(".go", ".md"), []struct {
		name  string
		event Event
		want  bool
	}{
		{"go file", Event{Path: "/tmp/main.go", Op: Write, Timestamp: time.Now()}, true},
		{"md file", Event{Path: "/tmp/readme.md", Op: Write, Timestamp: time.Now()}, true},
		{"txt file", Event{Path: "/tmp/notes.txt", Op: Write, Timestamp: time.Now()}, false},
		{
			"go file uppercase ext",
			Event{Path: "/tmp/main.GO", Op: Write, Timestamp: time.Now()},
			true,
		},
	})
}

func TestFilterIgnoreExtensions(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterIgnoreExtensions", FilterIgnoreExtensions(".log", ".tmp"), []struct {
		name  string
		event Event
		want  bool
	}{
		{"go file", Event{Path: "/tmp/main.go", Op: Write, Timestamp: time.Now()}, true},
		{"log file", Event{Path: "/tmp/app.log", Op: Write, Timestamp: time.Now()}, false},
		{"tmp file", Event{Path: "/tmp/cache.tmp", Op: Write, Timestamp: time.Now()}, false},
	})
}

func TestFilterIgnoreDirs(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterIgnoreDirs", FilterIgnoreDirs("vendor", "node_modules"), []struct {
		name  string
		event Event
		want  bool
	}{
		{"normal file", Event{Path: "/tmp/main.go", Op: Write, Timestamp: time.Now()}, true},
		{"vendor file", Event{Path: "/tmp/vendor/pkg.go", Op: Write, Timestamp: time.Now()}, false},
		{
			"nested vendor",
			Event{Path: "/tmp/pkg/vendor/lib.go", Op: Write, Timestamp: time.Now()},
			false,
		},
		{
			"node_modules",
			Event{Path: "/tmp/node_modules/index.js", Op: Write, Timestamp: time.Now()},
			false,
		},
	})
}

func TestFilterIgnoreHidden(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterIgnoreHidden", FilterIgnoreHidden(), []struct {
		name  string
		event Event
		want  bool
	}{
		{"normal file", Event{Path: "/tmp/main.go", Op: Write, Timestamp: time.Now()}, true},
		{"hidden file", Event{Path: "/tmp/.hidden", Op: Write, Timestamp: time.Now()}, false},
		{"hidden dir", Event{Path: "/tmp/.git/config", Op: Write, Timestamp: time.Now()}, false},
		{"dotfile in name", Event{Path: "/tmp/.env", Op: Write, Timestamp: time.Now()}, false},
	})
}

func TestFilterOperations(t *testing.T) {
	t.Parallel()
	runFilterTests(t, "FilterOperations", FilterOperations(Write, Create), []struct {
		name  string
		event Event
		want  bool
	}{
		{"write", Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now()}, true},
		{"create", Event{Op: Create, Path: "/tmp/test.txt", Timestamp: time.Now()}, true},
		{"remove", Event{Op: Remove, Path: "/tmp/test.txt", Timestamp: time.Now()}, false},
		{"rename", Event{Op: Rename, Path: "/tmp/test.txt", Timestamp: time.Now()}, false},
	})
}

func TestFilterNotOperations(t *testing.T) {
	t.Parallel()

	f := FilterNotOperations(Remove)

	if f(Event{Op: Remove, Path: "/tmp/test.txt", Timestamp: time.Now()}) {
		t.Error("expected Remove to be filtered out")
	}
	if !f(Event{Op: Write, Path: "/tmp/test.txt", Timestamp: time.Now()}) {
		t.Error("expected Write to pass")
	}
}

func TestFilterGlob(t *testing.T) {
	t.Parallel()

	f := FilterGlob("*.go")

	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{"go file", Event{Path: "/tmp/main.go", Op: Write, Timestamp: time.Now()}, true},
		{"txt file", Event{Path: "/tmp/readme.txt", Op: Write, Timestamp: time.Now()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := f(tt.event); got != tt.want {
				t.Errorf("FilterGlob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterAnd(t *testing.T) {
	t.Parallel()

	f := FilterAnd(
		FilterExtensions(".go"),
		FilterOperations(Write),
	)

	if !f(Event{Path: "main.go", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .go Write to pass")
	}
	if f(Event{Path: "main.go", Op: Remove, Timestamp: time.Now()}) {
		t.Error("expected .go Remove to be filtered")
	}
	if f(Event{Path: "main.txt", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .txt Write to be filtered")
	}
}

func TestFilterOr(t *testing.T) {
	t.Parallel()

	f := FilterOr(
		FilterExtensions(".go"),
		FilterExtensions(".md"),
	)

	if !f(Event{Path: "main.go", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .go to pass")
	}
	if !f(Event{Path: "readme.md", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .md to pass")
	}
	if f(Event{Path: "main.txt", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .txt to be filtered")
	}
}

func TestFilterNot(t *testing.T) {
	t.Parallel()

	f := FilterNot(FilterExtensions(".go"))

	if f(Event{Path: "main.go", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .go to be filtered after inversion")
	}
	if !f(Event{Path: "main.txt", Op: Write, Timestamp: time.Now()}) {
		t.Error("expected .txt to pass after inversion")
	}
}

func TestEvent_String(t *testing.T) {
	t.Parallel()

	e := Event{
		Path:      "/tmp/test.go",
		Op:        Write,
		Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if !contains(s, "WRITE") {
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

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
