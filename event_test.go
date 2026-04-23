//nolint:testpackage,varnamelen // Tests need internal access; idiomatic short names in tests
package filewatcher

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func TestOp_MarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op   Op
		want string
	}{
		{Create, "CREATE"},
		{Write, "WRITE"},
		{Remove, "REMOVE"},
		{Rename, "RENAME"},
	}

	for _, tt := range tests {
		got, err := tt.op.MarshalText()
		if err != nil {
			t.Errorf("Op(%d).MarshalText() error: %v", tt.op, err)
		}

		if string(got) != tt.want {
			t.Errorf("Op(%d).MarshalText() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestOp_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Op
	}{
		{"CREATE", Create},
		{"WRITE", Write},
		{"REMOVE", Remove},
		{"RENAME", Rename},
	}

	for _, tt := range tests {
		var op Op

		err := op.UnmarshalText([]byte(tt.input))
		if err != nil {
			t.Errorf("UnmarshalText(%q) error: %v", tt.input, err)
		}

		if op != tt.want {
			t.Errorf("UnmarshalText(%q) = %d, want %d", tt.input, op, tt.want)
		}
	}
}

func TestOp_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()

	var op Op

	err := op.UnmarshalText([]byte("INVALID"))
	if err == nil {
		t.Error("expected error for invalid operation")
	}
}

func TestEvent_JSON(t *testing.T) {
	t.Parallel()

	event := fixedTimeEvent("/tmp/test.go", Write, 12)

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded Event

	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.Path != event.Path {
		t.Errorf("Path = %q, want %q", decoded.Path, event.Path)
	}

	if decoded.Op != event.Op {
		t.Errorf("Op = %d, want %d", decoded.Op, event.Op)
	}

	if !decoded.Timestamp.Equal(event.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", decoded.Timestamp, event.Timestamp)
	}

	if decoded.IsDir != event.IsDir {
		t.Errorf("IsDir = %v, want %v", decoded.IsDir, event.IsDir)
	}
}

func TestEvent_LogValue(t *testing.T) {
	t.Parallel()

	ts := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	event := Event{
		Path:      "/tmp/test.go",
		Op:        Write,
		Timestamp: ts,
		IsDir:     false,
	}

	val := event.LogValue()

	if val.Kind() != slog.KindGroup {
		t.Fatalf("LogValue kind = %v, want Group", val.Kind())
	}

	attrs := val.Group()
	if len(attrs) != 4 {
		t.Fatalf("LogValue group length = %d, want 4", len(attrs))
	}

	found := map[string]bool{}
	for _, attr := range attrs {
		found[attr.Key] = true
	}

	for _, key := range []string{"path", "op", "timestamp", "isDir"} {
		if !found[key] {
			t.Errorf("missing key %q in LogValue attrs", key)
		}
	}
}
