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

const (
	exampleTimeout = 10 * time.Second // Total runtime for the example
	maxEventCount  = 10               // Number of events to process before stopping
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), exampleTimeout)
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
		//nolint:gocritic // log.Fatal exits immediately, defer won't run (intentional)
		log.Fatal(err)
	}

	events, err := watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching with middleware: logging + metrics")
	fmt.Println("Press Ctrl+C or wait 10s to exit")

	counter := 0
	for range events {
		counter++
		if counter >= maxEventCount {
			break
		}
	}

	fmt.Printf("\nFinal counts - Create: %d, Write: %d, Remove: %d\n",
		createCount.Load(), writeCount.Load(), removeCount.Load())

	_ = watcher.Close()
}
