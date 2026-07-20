// Example: Basic file watching
// Simplest usage with extensions filter and global debounce.
package main

import (
	"context"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
	demo "github.com/larsartmann/go-filewatcher/v2/examples/demo"
)

func main() {
	demo.Run(func(ctx context.Context) {
		watcher, err := filewatcher.New(
			[]string{"."},
			filewatcher.WithExtensions(".go", ".md"),
			filewatcher.WithDebounce(300*time.Millisecond),
		)
		if err != nil {
			log.Fatal(err)
		}

		defer func() { _ = watcher.Close() }()

		events, err := watcher.Watch(ctx)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Watching for .go and .md file changes (10s timeout)...")

		for event := range events {
			demo.PrintEvent(event)
		}

		log.Println("Done.")
	})
}
