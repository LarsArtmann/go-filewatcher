package filewatcher

import (
	"fmt"
	"time"
)

// Op represents a file system operation type.
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

// Event represents a file system change event.
type Event struct {
	// Path is the absolute path of the file or directory that changed.
	Path string
	// Op is the operation that occurred.
	Op Op
	// Timestamp is when the event was detected.
	Timestamp time.Time
	// IsDir indicates whether the event is for a directory (true) or file (false).
	// This allows consumers to distinguish between directory and file events.
	IsDir bool
}

// String returns a human-readable representation of the event.
func (e Event) String() string {
	return fmt.Sprintf("%s %s at %s", e.Op, e.Path, e.Timestamp.Format(time.RFC3339))
}
