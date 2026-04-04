// Example: Per-path debouncing
// Each file path is debounced independently.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithPerPathDebounce(500*time.Millisecond),
		filewatcher.WithFilter(filewatcher.FilterExtensions(".go")),
	)
	if err != nil {
		cancel()
		_ = watcher.Close()
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
		fmt.Printf("[%s] %s: %s\n", event.Timestamp.Format("15:04:05.000"), event.Op, event.Path)
	}

	_ = watcher.Close()
}
