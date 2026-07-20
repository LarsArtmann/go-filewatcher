// filter-generated demonstrates how to use go-filewatcher with gogenfilter
// to automatically exclude auto-generated Go code files from file watching events.
//
// This is useful when you want to watch for changes in your source code but
// skip generated files from tools like sqlc, templ, protobuf, mockgen, etc.
//
// Run this example:
//
//	go run ./examples/filter-generated
//
// Then in another terminal, create some files:
//
//	touch /tmp/watchtest/main.go              # Will be detected
//	touch /tmp/watchtest/models.go            # Will be ignored (sqlc)
//	touch /tmp/watchtest/page_templ.go        # Will be ignored (templ)
//	touch /tmp/watchtest/user.pb.go           # Will be ignored (protobuf)
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/gogenfilter/v3"
	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

const (
	debounceDelay = 100 * time.Millisecond
	watchTimeout  = 2 * time.Second
	filePerms     = 0o600
	dirPerms      = 0o750
)

func main() {
	watchDir, err := os.MkdirTemp("", "filewatcher-example-*")
	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll(watchDir)
	if err != nil {
		log.Printf("Failed to cleanup watch dir: %v", err)
	}

	// Create subdirectories to demonstrate filtering
	dirs := []string{"db", "api", "web", "mocks"}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(watchDir, dir), dirPerms)
		if err != nil {
			log.Printf("Failed to create directory %s: %v", dir, err)

			return
		}
	}

	log.Println("Watching directory:", watchDir)
	log.Println()

	// Example 1: Filter specific generator types
	demonstrateSpecificFilters(watchDir)

	// Example 2: Filter all generated code types
	demonstrateAllFilters(watchDir)

	// Example 3: Using the detector directly
	demonstrateDetector(watchDir)
}

// collectEvents collects events from a channel until context is done.
func collectEvents(ctx context.Context, events <-chan filewatcher.Event) []string {
	var receivedEvents []string

	eventDone := make(chan struct{})

	go func() {
		for event := range events {
			receivedEvents = append(receivedEvents, filepath.Base(event.Path))
		}

		close(eventDone)
	}()

	<-ctx.Done()
	<-eventDone

	return receivedEvents
}

// startFilteredWatch creates a watcher over watchDir, starts a timeout-bounded
// Watch, and registers cleanup for both. Centralizes the setup boilerplate so
// each demonstration can focus on its specific configuration.
func startFilteredWatch(watchDir string, opts ...filewatcher.Option) (context.Context, <-chan filewatcher.Event) {
	watcher, err := filewatcher.New(
		[]string{watchDir},
		opts...,
	)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	defer func() { _ = watcher.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), watchTimeout)
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		//nolint:gocritic // log.Fatalf exits, cancel() runs via defer on success path
		log.Fatalf("Failed to watch: %v", err)
	}

	return ctx, events
}

// printTriggeredEvents logs each event name under a "Files that triggered events:" header.
// Centralizes the post-collection reporting used by every demonstration.
func printTriggeredEvents(receivedEvents []string) {
	log.Println("Files that triggered events:")

	for _, name := range receivedEvents {
		log.Printf("  - %s\n", name)
	}

	log.Println()
}

// collectAndReport waits for file events to settle, prints which files triggered
// events, then logs each summary line. Centralizes the collect-then-report flow
// shared by the demonstration functions.
func collectAndReport(ctx context.Context, events <-chan filewatcher.Event, summary ...string) {
	receivedEvents := collectEvents(ctx, events)
	printTriggeredEvents(receivedEvents)

	for _, line := range summary {
		log.Println(line)
	}
}

// demonstrateSpecificFilters shows filtering specific generator types.
func demonstrateSpecificFilters(watchDir string) {
	log.Println("=== Example 1: Filter Specific Generator Types ===")
	log.Println("Filtering: sqlc and protobuf files only")
	log.Println()

	// Create watcher that filters sqlc and protobuf files
	ctx, events := startFilteredWatch(
		watchDir,
		filewatcher.WithFilter(filewatcher.FilterGeneratedCode(
			gogenfilter.FilterSQLC,
			gogenfilter.FilterProtobuf,
		)),
		filewatcher.WithDebounce(debounceDelay),
	)

	// Create test files
	createTestFile(watchDir, "main.go", "package main")
	createTestFile(watchDir, "db/models.go", "package db")             // sqlc - filtered
	createTestFile(watchDir, "api/user.pb.go", "package api")          // protobuf - filtered
	createTestFile(watchDir, "web/page_templ.go", "package web")       // templ - NOT filtered
	createTestFile(watchDir, "mocks/service_mock.go", "package mocks") // mockgen - NOT filtered

	collectAndReport(
		ctx, events,
		"Filtered (no events):",
		"  - models.go (sqlc)",
		"  - user.pb.go (protobuf)",
		"",
	)
}

// demonstrateAllFilters shows filtering all generator types.
func demonstrateAllFilters(watchDir string) {
	log.Println("=== Example 2: Filter All Generated Code ===")
	log.Println()

	// Create watcher that filters ALL generated code types
	ctx, events := startFilteredWatch(
		watchDir,
		filewatcher.WithFilter(filewatcher.FilterGeneratedCode()), // Defaults to FilterAll
		filewatcher.WithDebounce(debounceDelay),
	)

	// Create more test files
	createTestFile(watchDir, "regular.go", "package main")
	createTestFile(watchDir, "handlers.go", "package main")

	collectAndReport(ctx, events, "All generated files are filtered!", "")
}

// demonstrateDetector shows using the detector directly.
func demonstrateDetector(watchDir string) {
	log.Println("=== Example 3: Direct Detector Usage ===")
	log.Println()

	// Create a detector for specific generator types
	detector := filewatcher.NewGeneratedCodeDetector(
		gogenfilter.FilterSQLC,
		gogenfilter.FilterTempl,
		gogenfilter.FilterProtobuf,
	)

	// Test various file paths
	testFiles := []string{
		filepath.Join(watchDir, "db/models.go"),
		filepath.Join(watchDir, "web/page_templ.go"),
		filepath.Join(watchDir, "api/user.pb.go"),
		filepath.Join(watchDir, "main.go"),
		filepath.Join(watchDir, "utils.go"),
	}

	log.Println("Checking files with detector:")

	for _, path := range testFiles {
		isGenerated := detector.IsGenerated(path)
		reason := detector.GetReason(path)

		status := "regular"
		if isGenerated {
			status = string(reason)
		}

		log.Printf("  - %s: %s\n", filepath.Base(path), status)
	}

	log.Println()
}

// createTestFile creates a test file with the given content.
func createTestFile(dir, filename, content string) {
	path := filepath.Join(dir, filename)

	err := os.MkdirAll(filepath.Dir(path), dirPerms)
	if err != nil {
		log.Printf("Failed to create directory for %s: %v", filename, err)

		return
	}

	err = os.WriteFile(path, []byte(content), filePerms)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filename, err)

		return
	}
}
