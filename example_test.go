package filewatcher_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/larsartmann/go-filewatcher"
)

// ExampleNew demonstrates creating a basic watcher with options.
func ExampleNew() {
	// Create a watcher for the current directory
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithExtensions(".go"),
		filewatcher.WithDebounce(500*time.Millisecond),
		filewatcher.WithIgnoreDirs("vendor", "node_modules"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher created successfully")
	// Output: Watcher created successfully
}

// ExampleWatcher_Watch demonstrates watching for file events.
func ExampleWatcher_Watch() {
	// This example shows the pattern for consuming events.
	// In real usage, you would run this in a goroutine.
	watcher, err := filewatcher.New([]string{"."})
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		//nolint:gocritic // log.Fatal exits immediately, defer won't run (intentional)
		log.Fatal(err)
	}

	// Process events until context is cancelled
	for event := range events {
		fmt.Printf("%s: %s\n", event.Op.String(), event.Path)
	}

	fmt.Println("Watcher created and started")

	// Output:
	// Watcher created and started
}

// ExampleWithFilter demonstrates using size-based filters.
func ExampleWithFilter() {
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithFilter(filewatcher.FilterMinSize(100)),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with minimum size filter created")
	// Output: Watcher with minimum size filter created
}

// ExampleWithMiddleware demonstrates using middleware.
func ExampleWithMiddleware() {
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareRecovery(),
			filewatcher.MiddlewareLogging(nil),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with middleware created")
	// Output: Watcher with middleware created
}

// ExampleWithBuffer demonstrates using a custom buffer size.
func ExampleWithBuffer() {
	// Use a larger buffer for high-traffic directories
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithBuffer(256),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with custom buffer created")
	// Output: Watcher with custom buffer created
}

// ExampleWatcher_Remove demonstrates removing a watched path.
func ExampleWatcher_Remove() {
	watcher, err := filewatcher.New([]string{"."})
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	// Start watching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = watcher.Watch(ctx)
	if err != nil {
		//nolint:gocritic // log.Fatal exits immediately, defer won't run (intentional)
		log.Fatal(err)
	}

	// Later, stop watching a specific subdirectory
	// In real usage, this would be a subdirectory of the watched path
	// _ = watcher.Remove("./some-subdirectory")

	fmt.Println("Watcher with remove capability created")
	// Output: Watcher with remove capability created
}

// ExampleWatcher_WatchList demonstrates inspecting watched paths.
func ExampleWatcher_WatchList() {
	watcher, err := filewatcher.New([]string{"."})
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	// Get the list of paths currently being watched
	paths := watcher.WatchList()
	fmt.Printf("Watching %d paths.\n", len(paths))

	// Output: Watching 0 paths.
}

// ExampleWatcher_Stats demonstrates getting watcher statistics.
func ExampleWatcher_Stats() {
	watcher, err := filewatcher.New([]string{"."})
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	// Get current statistics
	stats := watcher.Stats()
	fmt.Printf("Watch count: %d, Watching: %v, Closed: %v\n",
		stats.WatchCount, stats.IsWatching, stats.IsClosed)

	// Output: Watch count: 0, Watching: false, Closed: false
}

// ExampleFilterExtensions demonstrates filtering by file extension.
func ExampleFilterExtensions() {
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithExtensions(".go", ".md"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher filtering .go and .md files")
	// Output: Watcher filtering .go and .md files
}

// ExampleFilterRegex demonstrates filtering with regex patterns.
func ExampleFilterRegex() {
	// Only match files ending with _test.go
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithFilter(filewatcher.FilterRegex(`_test\.go$`)),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher filtering with regex pattern")
	// Output: Watcher filtering with regex pattern
}

// ExampleFilterAnd demonstrates combining filters with AND logic.
func ExampleFilterAnd() {
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithFilter(goExcludeDirsFilter("vendor")),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with combined filters created")
	// Output: Watcher with combined filters created
}

// ExampleMiddlewareRateLimit demonstrates rate limiting.
func ExampleMiddlewareRateLimit() {
	// Limit to 10 events per second
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareRateLimit(10),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with rate limiting created")
	// Output: Watcher with rate limiting created
}

// ExampleMiddlewareMetrics demonstrates event metrics.
func ExampleMiddlewareMetrics() {
	// Count events by operation type
	eventCounts := make(map[filewatcher.Op]int)

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareMetrics(func(op filewatcher.Op) {
				eventCounts[op]++
			}),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with metrics created")
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
	// Accept either .go files or .md files
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithFilter(filewatcher.FilterOr(
			filewatcher.FilterExtensions(".go"),
			filewatcher.FilterExtensions(".md"),
		)),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with OR filter created")
	// Output: Watcher with OR filter created
}

// ExampleEventPath demonstrates phantom type usage for type-safe paths.
func ExampleEventPath() {
	// Create an event and extract its path as a phantom type
	//nolint:exhaustruct // example demonstrating minimal fields
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
	// Each file is debounced independently
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithPerPathDebounce(500*time.Millisecond),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with per-path debounce created")
	// Output: Watcher with per-path debounce created
}

// ExampleMiddlewareDeduplicate demonstrates event deduplication.
func ExampleMiddlewareDeduplicate() {
	// Drop duplicate events within 100ms window
	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareDeduplicate(100*time.Millisecond),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with deduplication created")
	// Output: Watcher with deduplication created
}

// ExampleFilterModifiedSince demonstrates filtering by modification time.
func ExampleFilterModifiedSince() {
	// Only process files modified in the last hour
	oneHourAgo := time.Now().Add(-time.Hour)

	watcher, err := filewatcher.New(
		[]string{"."},
		filewatcher.WithFilter(filewatcher.FilterModifiedSince(oneHourAgo)),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = watcher.Close() }()

	fmt.Println("Watcher with time filter created")
	// Output: Watcher with time filter created
}
