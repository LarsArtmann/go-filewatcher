package filewatcher_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/larsartmann/go-filewatcher/v2"
)

// runExampleWatcher creates a watcher on the current directory with the given
// options, invokes fn with it, then closes it. Centralises the error-check and
// cleanup boilerplate shared by every Example function so each one can focus
// on the option or feature it demonstrates.
func runExampleWatcher(fn func(*filewatcher.Watcher), opts ...filewatcher.Option) {
	watcher, err := filewatcher.New([]string{"."}, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	fn(watcher)
}

// ExampleNew demonstrates creating a basic watcher with options.
func ExampleNew() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher created successfully")
		},
		filewatcher.WithExtensions(".go"),
		filewatcher.WithDebounce(500*time.Millisecond),
		filewatcher.WithIgnoreDirs("vendor", "node_modules"),
	)
	// Output: Watcher created successfully
}

// ExampleWatcher_Watch demonstrates watching for file events.
func ExampleWatcher_Watch() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		events, err := watcher.Watch(ctx)
		if err != nil {
			log.Fatal(err)
		}

		// Process events until context is cancelled
		for event := range events {
			fmt.Printf("%s: %s\n", event.Op.String(), event.Path)
		}

		fmt.Println("Watcher created and started")
	})
	// Output:
	// Watcher created and started
}

// ExampleWithMiddleware demonstrates using middleware.
func ExampleWithMiddleware() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with middleware created")
		},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareRecovery(),
			filewatcher.MiddlewareLogging(nil),
		),
	)
	// Output: Watcher with middleware created
}

// ExampleWithBuffer demonstrates using a custom buffer size.
func ExampleWithBuffer() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with custom buffer created")
		},
		filewatcher.WithBuffer(256),
	)
	// Output: Watcher with custom buffer created
}

// ExampleWatcher_Remove demonstrates removing a watched path.
func ExampleWatcher_Remove() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err := watcher.Watch(ctx)
		if err != nil {
			log.Fatal(err)
		}

		// Later, stop watching a specific subdirectory
		// In real usage, this would be a subdirectory of the watched path
		// _ = watcher.Remove("./some-subdirectory")

		fmt.Println("Watcher with remove capability created")
	})
	// Output: Watcher with remove capability created
}

// ExampleWatcher_WatchList demonstrates inspecting watched paths.
func ExampleWatcher_WatchList() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		// Get the list of paths currently being watched
		paths := watcher.WatchList()
		fmt.Printf("Watching %d paths.\n", len(paths))
	})
	// Output: Watching 0 paths.
}

// ExampleWatcher_Stats demonstrates getting watcher statistics.
func ExampleWatcher_Stats() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		// Get current statistics
		stats := watcher.Stats()
		fmt.Printf("Watch count: %d, Watching: %v, Closed: %v\n",
			stats.WatchCount, stats.IsWatching, stats.IsClosed)
	})
	// Output: Watch count: 0, Watching: false, Closed: false
}

// ExampleFilterExtensions demonstrates filtering by file extension.
func ExampleFilterExtensions() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher filtering .go and .md files")
		},
		filewatcher.WithExtensions(".go", ".md"),
	)
	// Output: Watcher filtering .go and .md files
}

// ExampleFilterRegex demonstrates filtering with regex patterns.
func ExampleFilterRegex() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher filtering with regex pattern")
		},
		filewatcher.WithFilter(filewatcher.FilterRegex(`_test\.go$`)),
	)
	// Output: Watcher filtering with regex pattern
}

// ExampleFilterAnd demonstrates combining filters with AND logic.
func ExampleFilterAnd() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with combined filters created")
		},
		filewatcher.WithFilter(goExcludeDirsFilter("vendor")),
	)
	// Output: Watcher with combined filters created
}

// ExampleMiddlewareRateLimit demonstrates rate limiting.
func ExampleMiddlewareRateLimit() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with rate limiting created")
		},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareRateLimit(10),
		),
	)
	// Output: Watcher with rate limiting created
}

// ExampleMiddlewareMetrics demonstrates event metrics.
func ExampleMiddlewareMetrics() {
	// Count events by operation type
	eventCounts := make(map[filewatcher.Op]int)

	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with metrics created")
		},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareMetrics(func(op filewatcher.Op) {
				eventCounts[op]++
			}),
		),
	)
	// Output: Watcher with metrics created
}

// ExampleDebouncer demonstrates using the debouncer directly.
func ExampleDebouncer() {
	// Create a debouncer with 100ms delay
	debouncer := filewatcher.NewDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	// Schedule a function - it will run after 100ms unless reset
	debouncer.Debounce(filewatcher.NewDebounceKey("key"), func() {
		fmt.Println("Debounced function executed")
	})

	// Flush to execute immediately
	debouncer.Flush()

	fmt.Printf("Pending: %d\n", debouncer.Pending())
	// Output: Debounced function executed
	// Pending: 0
}

// ExampleEvent demonstrates event structure.
func ExampleEvent() {
	// Use a fixed timestamp for deterministic output
	fixedTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	event := filewatcher.Event{
		Path:      "/path/to/file.go",
		Op:        filewatcher.Write,
		Timestamp: fixedTime,
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}

	fmt.Printf("Event: %s\n", event.String())
	fmt.Printf("Operation: %s\n", event.Op.String())
	fmt.Printf("Is directory: %v\n", event.IsDir)

	// Output:
	// Event: WRITE /path/to/file.go at 2006-01-02T15:04:05Z
	// Operation: WRITE
	// Is directory: false
}

func goExcludeDirsFilter(dirs ...string) filewatcher.Filter {
	return filewatcher.FilterAnd(
		filewatcher.FilterExtensions(".go"),
		filewatcher.FilterNot(filewatcher.FilterIgnoreDirs(dirs...)),
	)
}

// ExampleFilterOr demonstrates combining filters with OR logic.
func ExampleFilterOr() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with OR filter created")
		},
		filewatcher.WithFilter(filewatcher.FilterOr(
			filewatcher.FilterExtensions(".go"),
			filewatcher.FilterExtensions(".md"),
		)),
	)
	// Output: Watcher with OR filter created
}

// ExampleEventPath demonstrates phantom type usage for type-safe paths.
func ExampleEventPath() {
	// Create an event and extract its path as a phantom type
	event := filewatcher.Event{
		Path: "/home/user/project/main.go",
		Op:   filewatcher.Write,
	}

	path := event.GetPath()
	fmt.Printf("Base: %s\n", path.Base())
	fmt.Printf("Extension: %s\n", path.Ext())
	fmt.Printf("Directory: %s\n", path.Dir())

	// Output:
	// Base: main.go
	// Extension: .go
	// Directory: /home/user/project
}

// ExampleWithPerPathDebounce demonstrates per-path debouncing.
func ExampleWithPerPathDebounce() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with per-path debounce created")
		},
		filewatcher.WithPerPathDebounce(500*time.Millisecond),
	)
	// Output: Watcher with per-path debounce created
}

// ExampleMiddlewareDeduplicate demonstrates event deduplication.
func ExampleMiddlewareDeduplicate() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with deduplication created")
		},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareDeduplicate(100*time.Millisecond),
		),
	)
	// Output: Watcher with deduplication created
}

// ExampleFilterModifiedSince demonstrates filtering by modification time.
func ExampleFilterModifiedSince() {
	// Only process files modified in the last hour
	oneHourAgo := time.Now().Add(-time.Hour)

	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with time filter created")
		},
		filewatcher.WithFilter(filewatcher.FilterModifiedSince(oneHourAgo)),
	)
	// Output: Watcher with time filter created
}

// ExampleMiddlewareLogging_structured demonstrates structured logging with a custom logger.
func ExampleMiddlewareLogging_structured() {
	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with custom structured logger created")
		},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareLogging(logger),
		),
	)
	// Output: Watcher with custom structured logger created
}

// ExampleWatcher_Add demonstrates adding a path to an existing watcher.
func ExampleWatcher_Add() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		// Add a path to the existing watcher. Errors are returned for
		// invalid paths, permission issues, or fsnotify resource limits.
		addErr := watcher.Add("./internal")
		fmt.Println(addErr == nil)
	}, filewatcher.WithRecursive(false))
	// Output: true
}

// ExampleWatcher_Errors demonstrates receiving errors via a channel.
func ExampleWatcher_Errors() {
	runExampleWatcher(func(watcher *filewatcher.Watcher) {
		// Get the errors channel - closed when the watcher is closed.
		errCh := watcher.Errors()
		_ = errCh

		fmt.Println("errors channel initialized")
	})
	// Output: errors channel initialized
}

// ExampleWithErrorHandler demonstrates error handler callback.
func ExampleWithErrorHandler() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with error handler created")
		},
		filewatcher.WithErrorHandler(func(_ filewatcher.ErrorContext, err error) {
			// Handle errors - in production, log to your error tracking system
			_ = err
		}),
	)
	// Output: Watcher with error handler created
}

// ExampleWithDebug demonstrates debug logging.
func ExampleWithDebug() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with debug logging enabled")
		},
		filewatcher.WithDebug(nil), // nil uses slog.Default()
	)
	// Output: Watcher with debug logging enabled
}

// ExampleFilterMinSize demonstrates filtering by minimum file size.
func ExampleFilterMinSize() {
	runExampleWatcher(
		func(_ *filewatcher.Watcher) {
			fmt.Println("Watcher with min size filter created")
		},
		filewatcher.WithFilter(filewatcher.FilterMinSize(1024)),
	)
	// Output: Watcher with min size filter created
}

// ExampleOp demonstrates Op string conversion and JSON marshaling.
func ExampleOp() {
	op := filewatcher.Write
	fmt.Println(op.String())

	data, err := json.Marshal(op)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))

	// Output:
	// WRITE
	// "WRITE"
}
