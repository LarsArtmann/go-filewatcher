package filewatcher

import (
	"fmt"
	"path/filepath"

	id "github.com/larsartmann/go-branded-id"
)

// Brand types for compile-time type safety.

// EventPathBrand is the brand for event file/directory paths.
type EventPathBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (EventPathBrand) Name() string { return "EventPath" }

// RootPathBrand is the brand for root directory paths during filesystem walking.
type RootPathBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (RootPathBrand) Name() string { return "RootPath" }

// DebounceKeyBrand is the brand for debouncer keys (typically file paths).
type DebounceKeyBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (DebounceKeyBrand) Name() string { return "DebounceKey" }

// LogSubstringBrand is the brand for log substring assertions in tests.
type LogSubstringBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (LogSubstringBrand) Name() string { return "LogSubstring" }

// TempDirBrand is the brand for temporary directory paths in tests.
type TempDirBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (TempDirBrand) Name() string { return "TempDir" }

// OpStringBrand is the brand for operation names (e.g., "fsnotify", "middleware").
type OpStringBrand struct{}

// Name implements id.Brand for debugging/introspection.
func (OpStringBrand) Name() string { return "OpString" }

// EventPath is a branded type for event file/directory paths.
// It prevents accidentally passing event paths where other path types are expected.
type EventPath struct {
	id id.ID[EventPathBrand, string]
}

// NewEventPath creates a new EventPath from a string.
func NewEventPath(path string) EventPath {
	return EventPath{id: id.NewID[EventPathBrand](path)}
}

// Get returns the underlying string value.
func (ep EventPath) Get() string {
	return ep.id.Get()
}

// IsZero returns true if the EventPath is empty.
func (ep EventPath) IsZero() bool {
	return ep.id.IsZero()
}

// String returns the string representation.
func (ep EventPath) String() string {
	return ep.id.String()
}

// Equal returns true if the two EventPaths are equal.
func (ep EventPath) Equal(other EventPath) bool {
	return ep.id.Equal(other.id)
}

// Compare returns -1 if ep < other, 0 if equal, 1 if greater.
// Returns error if values cannot be compared.
func (ep EventPath) Compare(other EventPath) (int, error) {
	cmp, err := ep.id.Compare(other.id)
	if err != nil {
		return 0, fmt.Errorf("EventPath.Compare: %w", err)
	}

	return cmp, nil
}

// Base returns the last element of the path.
// Example: EventPath("/home/user/file.go").Base() returns "file.go".
func (ep EventPath) Base() string {
	return filepath.Base(ep.Get())
}

// Dir returns all but the last element of the path.
// Example: EventPath("/home/user/file.go").Dir() returns EventPath("/home/user").
func (ep EventPath) Dir() EventPath {
	return NewEventPath(filepath.Dir(ep.Get()))
}

// Ext returns the file extension of the path.
// Example: EventPath("/home/user/file.go").Ext() returns ".go".
func (ep EventPath) Ext() string {
	return filepath.Ext(ep.Get())
}

// Join appends the given elements to the path.
// Example: EventPath("/home/user").Join("docs", "readme.md") returns EventPath("/home/user/docs/readme.md").
func (ep EventPath) Join(elem ...string) EventPath {
	all := make([]string, 0, len(elem)+1)
	all = append(all, ep.Get())
	all = append(all, elem...)

	return NewEventPath(filepath.Join(all...))
}

// RootPath is a branded type for root directory paths during filesystem walking.
// It prevents accidentally passing event paths or other paths where root paths are expected.
type RootPath struct {
	id id.ID[RootPathBrand, string]
}

// NewRootPath creates a new RootPath from a string.
func NewRootPath(path string) RootPath {
	return RootPath{id: id.NewID[RootPathBrand](path)}
}

// Get returns the underlying string value.
func (rp RootPath) Get() string {
	return rp.id.Get()
}

// IsZero returns true if the RootPath is empty.
func (rp RootPath) IsZero() bool {
	return rp.id.IsZero()
}

// String returns the string representation.
func (rp RootPath) String() string {
	return rp.id.String()
}

// Equal returns true if the two RootPaths are equal.
func (rp RootPath) Equal(other RootPath) bool {
	return rp.id.Equal(other.id)
}

// DebounceKey is a branded type for debouncer keys (typically file paths).
// It ensures debounce keys are not mixed with other path-like strings.
type DebounceKey struct {
	id id.ID[DebounceKeyBrand, string]
}

// NewDebounceKey creates a new DebounceKey from a string.
func NewDebounceKey(key string) DebounceKey {
	return DebounceKey{id: id.NewID[DebounceKeyBrand](key)}
}

// Get returns the underlying string value.
func (dk DebounceKey) Get() string {
	return dk.id.Get()
}

// IsZero returns true if the DebounceKey is empty.
func (dk DebounceKey) IsZero() bool {
	return dk.id.IsZero()
}

// String returns the string representation.
func (dk DebounceKey) String() string {
	return dk.id.String()
}

// Equal returns true if the two DebounceKeys are equal.
func (dk DebounceKey) Equal(other DebounceKey) bool {
	return dk.id.Equal(other.id)
}

// LogSubstring is a branded type for log substring assertions in tests.
type LogSubstring struct {
	id id.ID[LogSubstringBrand, string]
}

// NewLogSubstring creates a new LogSubstring from a string.
func NewLogSubstring(s string) LogSubstring {
	return LogSubstring{id: id.NewID[LogSubstringBrand](s)}
}

// Get returns the underlying string value.
func (ls LogSubstring) Get() string {
	return ls.id.Get()
}

// String returns the string representation of LogSubstring.
func (ls LogSubstring) String() string {
	return ls.id.String()
}

// TempDir is a branded type for temporary directory paths in tests.
type TempDir struct {
	id id.ID[TempDirBrand, string]
}

// NewTempDir creates a new TempDir from a string.
func NewTempDir(path string) TempDir {
	return TempDir{id: id.NewID[TempDirBrand](path)}
}

// Get returns the underlying string value.
func (td TempDir) Get() string {
	return td.id.Get()
}

// String returns the string representation.
func (td TempDir) String() string {
	return td.id.String()
}

// OpString is a branded type for operation names (e.g., "fsnotify", "middleware").
type OpString struct {
	id id.ID[OpStringBrand, string]
}

// NewOpString creates a new OpString from a string.
func NewOpString(op string) OpString {
	return OpString{id: id.NewID[OpStringBrand](op)}
}

// Get returns the underlying string value.
func (os OpString) Get() string {
	return os.id.Get()
}

// String returns the string representation.
func (os OpString) String() string {
	return os.id.String()
}
