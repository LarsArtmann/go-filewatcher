// Example: Per-path debouncing
// Each file path is debounced independently.
package main

import (
	"context"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
	demo "github.com/larsartmann/go-filewatcher/v2/examples/demo"
)

const debounceDelay = 500 * time.Millisecond

func main() {
	demo.Run(func(ctx context.Context) {
		watcher, err := filewatcher.New(
			[]string{"."},
			filewatcher.WithPerPathDebounce(debounceDelay),
			filewatcher.WithFilter(filewatcher.FilterExtensions(".go")),
		)
		if err != nil {
			log.Fatal(err)
		}

		defer func() { _ = watcher.Close() }()

		events, err := watcher.Watch(ctx)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Per-path debounce: each file debounced independently")
		log.Println("Edit multiple files quickly - each will trigger separately after 500ms")
		log.Println("Watching for .go file changes...")

		for event := range events {
			demo.PrintEvent(event)
		}
	})
}
