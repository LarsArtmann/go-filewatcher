package filewatcher

import (
	"encoding"
	"fmt"
	"time"
)

// Op represents a file system operation type.
//
//nolint:recvcheck // UnmarshalText/UnmarshalJSON must have pointer receiver to modify the receiver.
type Op int

const (
	// Create indicates a file or directory was created.
	Create Op = iota + 1
	// Write indicates a file was modified.
	Write
	// Remove indicates a file or directory was removed.
	Remove
	// Rename indicates a file or directory was renamed.
	Rename
)

// Compile-time interface check: Op implements encoding.TextMarshaler and
// encoding.TextUnmarshaler for JSON, XML, YAML, and other serialization.
var (
	_ encoding.TextMarshaler   = (*Op)(nil)
	_ encoding.TextUnmarshaler = (*Op)(nil)
)

// String returns a human-readable representation of the operation.
func (op Op) String() string {
	switch op {
	case Create:
		return "CREATE"
	case Write:
		return "WRITE"
	case Remove:
		return "REMOVE"
	case Rename:
		return "RENAME"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", op)
	}
}

// MarshalText implements encoding.TextMarshaler.
func (op Op) MarshalText() ([]byte, error) {
	return []byte(op.String()), nil
}

// MarshalJSON implements json.Marshaler.
func (op Op) MarshalJSON() ([]byte, error) {
	return []byte(`"` + op.String() + `"`), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (op *Op) UnmarshalText(text []byte) error {
	switch string(text) {
	case "CREATE":
		*op = Create
	case "WRITE":
		*op = Write
	case "REMOVE":
		*op = Remove
	case "RENAME":
		*op = Rename
	default:
		return fmt.Errorf("unknown operation: %q", string(text))
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (op *Op) UnmarshalJSON(data []byte) error {
	// Remove surrounding quotes from JSON string
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	return op.UnmarshalText([]byte(str))
}

// Event represents a file system change event.
type Event struct {
	// Path is the absolute path of the file or directory that changed.
	Path string `json:"path"`
	// Op is the operation that occurred.
	Op Op `json:"op"`
	// Timestamp is when the event was detected.
	Timestamp time.Time `json:"timestamp"`
	// IsDir indicates whether the event is for a directory (true) or file (false).
	IsDir bool `json:"is_dir"`
}

// String returns a human-readable representation of the event.
func (e Event) String() string {
	return fmt.Sprintf("%s %s at %s", e.Op, e.Path, e.Timestamp.Format(time.RFC3339))
}
