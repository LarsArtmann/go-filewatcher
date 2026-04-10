// Example: Basic file watching
// Simplest usage with extensions filter and global debounce.
package main

import (
	"fmt"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
	demo "github.com/larsartmann/go-filewatcher/examples/demo"
)

const debounceDelay = 300 * time.Millisecond // Delay for coalescing rapid file events

func main() {
	ctx, cancel := demo.WithDefaultTimeout()
	defer cancel()

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithExtensions(".go", ".md"),
		filewatcher.WithDebounce(debounceDelay),
	)
	if err != nil {
		//nolint:gocritic // log.Fatal exits immediately, defer won't run (intentional)
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	events, err := watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for .go and .md file changes (10s timeout)...")

	for event := range events {
		demo.PrintEvent(event)
	}

	fmt.Println("Done.")
}
