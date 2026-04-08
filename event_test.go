package filewatcher

import (
	"encoding/json"
	"testing"
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
		if err := op.UnmarshalText([]byte(tt.input)); err != nil {
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
	if err := json.Unmarshal(data, &decoded); err != nil {
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
