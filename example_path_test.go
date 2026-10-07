//go:build !windows

package filewatcher_test

import (
	"fmt"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

// ExampleEventPath demonstrates phantom type usage for type-safe paths.
// The Windows-shaped variant lives in example_path_windows_test.go.
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
	// Directory: /home/user/project
}
