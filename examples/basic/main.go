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

const debounceDelay = 300 * time.Millisecond

func main() {
	demo.Run(func(ctx context.Context) {
		events, cleanup := demo.MustWatch(ctx, []string{"."},
			filewatcher.WithExtensions(".go", ".md"),
			filewatcher.WithDebounce(debounceDelay),
		)
		defer cleanup()

		log.Println("Watching for .go and .md file changes (10s timeout)...")

		for event := range events {
			demo.PrintEvent(event)
		}

		log.Println("Done.")
	})
}
