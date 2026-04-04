// Example: Middleware chain
// Demonstrates logging, recovery, and metrics middleware.
package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var createCount, writeCount, removeCount atomic.Int64

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithExtensions(".go"),
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareRecovery(),
			filewatcher.MiddlewareLogging(nil),
			filewatcher.MiddlewareMetrics(func(op filewatcher.Op) {
				switch op {
				case filewatcher.Create:
					createCount.Add(1)
				case filewatcher.Write:
					writeCount.Add(1)
				case filewatcher.Remove:
					removeCount.Add(1)
				case filewatcher.Rename:
					// Rename operations tracked separately
				}
			}),
		),
	)
	if err != nil {
		_ = watcher.Close()
		cancel()
		log.Fatal(err)
	}

	events, err := watcher.Watch(ctx)
	if err != nil {
		_ = watcher.Close()
		cancel()
		log.Fatal(err)
	}

	fmt.Println("Watching with middleware: logging + metrics")
	fmt.Println("Press Ctrl+C or wait 10s to exit")

	counter := 0
	for range events {
		counter++
		if counter >= 10 {
			break
		}
	}

	fmt.Printf("\nFinal counts - Create: %d, Write: %d, Remove: %d\n",
		createCount.Load(), writeCount.Load(), removeCount.Load())

	_ = watcher.Close()
}
