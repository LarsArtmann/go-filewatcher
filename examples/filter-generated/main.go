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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/gogenfilter"
	filewatcher "github.com/larsartmann/go-filewatcher"
)

func main() {
	// Create a temporary directory to watch
	watchDir, err := os.MkdirTemp("", "filewatcher-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(watchDir)

	// Create subdirectories to demonstrate filtering
	dirs := []string{"db", "api", "web", "mocks"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(watchDir, dir), 0o755); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Watching directory:", watchDir)
	fmt.Println()

	// Example 1: Filter specific generator types
	demonstrateSpecificFilters(watchDir)

	// Example 2: Filter all generated code types
	demonstrateAllFilters(watchDir)

	// Example 3: Using the detector directly
	demonstrateDetector(watchDir)
}

// demonstrateSpecificFilters shows filtering specific generator types.
func demonstrateSpecificFilters(watchDir string) {
	fmt.Println("=== Example 1: Filter Specific Generator Types ===")
	fmt.Println("Filtering: sqlc and protobuf files only")
	fmt.Println()

	// Create watcher that filters sqlc and protobuf files
	watcher, err := filewatcher.New(
		[]string{watchDir},
		filewatcher.WithFilter(filewatcher.FilterGeneratedCode(
			gogenfilter.FilterSQLC,
			gogenfilter.FilterProtobuf,
		)),
		filewatcher.WithDebounce(100*time.Millisecond),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start watching with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Create test files
	createTestFile(watchDir, "main.go", "package main")
	createTestFile(watchDir, "db/models.go", "package db")          // sqlc - filtered
	createTestFile(watchDir, "api/user.pb.go", "package api")       // protobuf - filtered
	createTestFile(watchDir, "web/page_templ.go", "package web")    // templ - NOT filtered
	createTestFile(watchDir, "mocks/service_mock.go", "package mocks") // mockgen - NOT filtered

	// Collect events
	var receivedEvents []string
	eventDone := make(chan struct{})
	go func() {
		for event := range events {
			receivedEvents = append(receivedEvents, filepath.Base(event.Path))
		}
		close(eventDone)
	}()

	// Wait for timeout or context cancellation
	<-ctx.Done()
	<-eventDone

	fmt.Println("Files that triggered events:")
	for _, name := range receivedEvents {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()
	fmt.Println("Filtered (no events):")
	fmt.Println("  - models.go (sqlc)")
	fmt.Println("  - user.pb.go (protobuf)")
	fmt.Println()
}

// demonstrateAllFilters shows filtering all generator types.
func demonstrateAllFilters(watchDir string) {
	fmt.Println("=== Example 2: Filter All Generated Code ===")
	fmt.Println()

	// Create watcher that filters ALL generated code types
	watcher, err := filewatcher.New(
		[]string{watchDir},
		filewatcher.WithFilter(filewatcher.FilterGeneratedCode()), // Defaults to FilterAll
		filewatcher.WithDebounce(100*time.Millisecond),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start watching with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Create more test files
	createTestFile(watchDir, "regular.go", "package main")
	createTestFile(watchDir, "handlers.go", "package main")

	// Collect events
	var receivedEvents []string
	eventDone := make(chan struct{})
	go func() {
		for event := range events {
			receivedEvents = append(receivedEvents, filepath.Base(event.Path))
		}
		close(eventDone)
	}()

	// Wait for timeout
	<-ctx.Done()
	<-eventDone

	fmt.Println("Files that triggered events:")
	for _, name := range receivedEvents {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()
	fmt.Println("All generated files are filtered!")
	fmt.Println()
}

// demonstrateDetector shows using the detector directly.
func demonstrateDetector(watchDir string) {
	fmt.Println("=== Example 3: Direct Detector Usage ===")
	fmt.Println()

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

	fmt.Println("Checking files with detector:")
	for _, path := range testFiles {
		isGenerated := detector.IsGenerated(path)
		reason := detector.GetReason(path)
		status := "regular"
		if isGenerated {
			status = string(reason)
		}
		fmt.Printf("  - %s: %s\n", filepath.Base(path), status)
	}
	fmt.Println()
}

// createTestFile creates a test file with the given content.
func createTestFile(dir, filename, content string) {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("Failed to create directory for %s: %v", filename, err)
		return
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		log.Printf("Failed to create file %s: %v", filename, err)
		return
	}
}
