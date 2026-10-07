//go:build windows

package filewatcher_test

import (
	"fmt"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

// ExampleEventPath demonstrates phantom type usage for type-safe paths with
// Windows filepath semantics (backslash separators from filepath.Dir).
func ExampleEventPath() {
	// Create an event and extract its path as a phantom type
	event := filewatcher.Event{
		Path: "/home/user/project/main.go",
		Op:   filewatcher.Write,
	}

	path := event.GetPath()
	fmt.Printf("Base: %s\n", path.Base())
	fmt.Printf("Extension: %s\n", path.Ext())
	fmt.Printf("Directory: %s\n", path.Dir())

	// Output:
	// Base: main.go
	// Extension: .go
	// Directory: \home\user\project
}
