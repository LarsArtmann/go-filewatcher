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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithExtensions(".go", ".md"),
		filewatcher.WithDebounce(300*time.Millisecond),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	events, err := watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for .go and .md file changes (10s timeout)...")

	for event := range events {
		fmt.Printf("[%s] %s: %s\n", event.Timestamp.Format("15:04:05"), event.Op, event.Path)
	}

	fmt.Println("Done.")
}
