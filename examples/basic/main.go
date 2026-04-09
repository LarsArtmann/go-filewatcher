// Example: Basic file watching
// Simplest usage with extensions filter and global debounce.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
)

const (
	exampleTimeout = 10 * time.Second       // Total runtime for the example
	debounceDelay  = 300 * time.Millisecond // Delay for coalescing rapid file events
	timeFormat     = "15:04:05"
)

func printEvent(event filewatcher.Event) {
	ts := event.Timestamp.Format(timeFormat)
	fmt.Printf("[%s] %s: %s\n", ts, event.Op.String(), event.Path)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), exampleTimeout)
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
		printEvent(event)
	}

	fmt.Println("Done.")
}
