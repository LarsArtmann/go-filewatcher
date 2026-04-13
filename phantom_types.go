package filewatcher

// Phantom types for type-safe string parameters.
// These prevent accidentally passing the wrong string argument at compile time.

// DebounceKey is a phantom type for debouncer keys (typically file paths).
type DebounceKey string

// RootPath is a phantom type for root directory paths during filesystem walking.
type RootPath string

// LogSubstring is a phantom type for log substring assertions in tests.
type LogSubstring string

// TempDir is a phantom type for temporary directory paths in tests.
type TempDir string

// OpString is a phantom type for operation names (e.g., "fsnotify", "middleware").
type OpString string

// WatchDirString is a phantom type for watch directory paths.
type WatchDirString string

// RootString is a phantom type for root directory paths.
type RootString string

// EventPath is a phantom type for event file/directory paths.
// Used to distinguish event paths from other string parameters.
type EventPath string

// String returns the string representation of EventPath.
func (ep EventPath) String() string {
	return string(ep)
}
