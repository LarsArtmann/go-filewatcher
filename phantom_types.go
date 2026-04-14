package filewatcher

import "path/filepath"

// Phantom types for type-safe string parameters.
// These prevent accidentally passing the wrong string argument at compile time.

// DebounceKey is a phantom type for debouncer keys (typically file paths).
type DebounceKey string

// RootPath is a phantom type for root directory paths during filesystem walking.
type RootPath string

// LogSubstring is a phantom type for log substring assertions in tests.
type LogSubstring string

// String returns the string representation of LogSubstring.
func (ls LogSubstring) String() string {
	return string(ls)
}

// TempDir is a phantom type for temporary directory paths in tests.
type TempDir string

// OpString is a phantom type for operation names (e.g., "fsnotify", "middleware").
type OpString string

// EventPath is a phantom type for event file/directory paths.
// Used to distinguish event paths from other string parameters.
type EventPath string

// String returns the string representation of EventPath.
func (ep EventPath) String() string {
	return string(ep)
}

// Base returns the last element of the path.
// Example: EventPath("/home/user/file.go").Base() returns "file.go".
func (ep EventPath) Base() string {
	return filepath.Base(string(ep))
}

// Dir returns all but the last element of the path.
// Example: EventPath("/home/user/file.go").Dir() returns EventPath("/home/user").
func (ep EventPath) Dir() EventPath {
	return EventPath(filepath.Dir(string(ep)))
}

// Ext returns the file extension of the path.
// Example: EventPath("/home/user/file.go").Ext() returns ".go".
func (ep EventPath) Ext() string {
	return filepath.Ext(string(ep))
}

// Join appends the given elements to the path.
// Example: EventPath("/home/user").Join("docs", "readme.md") returns EventPath("/home/user/docs/readme.md").
func (ep EventPath) Join(elem ...string) EventPath {
	all := make([]string, 0, len(elem)+1)
	all = append(all, string(ep))
	all = append(all, elem...)

	return EventPath(filepath.Join(all...))
}
