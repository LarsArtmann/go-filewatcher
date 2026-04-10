// Example: Per-path debouncing
// Each file path is debounced independently.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
	"github.com/larsartmann/go-filewatcher/examples/shared"
)

const debounceDelay = 500 * time.Millisecond // Delay for per-path debouncing

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), shared.DefaultTimeout())
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

	fmt.Println("Per-path debounce: each file debounced independently")
	fmt.Println("Edit multiple files quickly - each will trigger separately after 500ms")
	fmt.Println("Watching for .go file changes...")

	for event := range events {
		shared.PrintEvent(event)
	}

	_ = watcher.Close()
}
