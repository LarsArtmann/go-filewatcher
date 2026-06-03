// Example: Per-path debouncing
// Each file path is debounced independently.
package main

import (
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
	demo "github.com/larsartmann/go-filewatcher/v2/examples/demo"
)

const debounceDelay = 500 * time.Millisecond // Delay for per-path debouncing

func main() {
	ctx, cancel := demo.WithDefaultTimeout()
	defer cancel()

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithPerPathDebounce(debounceDelay),
		filewatcher.WithFilter(filewatcher.FilterExtensions(".go")),
	)
	if err != nil {
		//nolint:gocritic // log.Fatal exits immediately, defer won't run (intentional)
		log.Fatal(err)
	}

	events, err := watcher.Watch(ctx)
	if err != nil {
		_ = watcher.Close()

		cancel()
		log.Fatal(err)
	}

	log.Println("Per-path debounce: each file debounced independently")
	log.Println("Edit multiple files quickly - each will trigger separately after 500ms")
	log.Println("Watching for .go file changes...")

	for event := range events {
		demo.PrintEvent(event)
	}

	_ = watcher.Close()
}
