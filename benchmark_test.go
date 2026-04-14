//nolint:testpackage,exhaustruct,wsl,varnamelen // Benchmarks prioritize readability
package filewatcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ============================================================================
// Watcher Creation Benchmarks
// ============================================================================

func BenchmarkNew_SinglePath(b *testing.B) {
	tmpDir := b.TempDir()

	b.ResetTimer()

	for range b.N {
		w, err := New([]string{tmpDir})
		if err != nil {
			b.Fatal(err)
		}

		_ = w.Close()
	}
}

func BenchmarkNew_WithOptions(b *testing.B) {
	tmpDir := b.TempDir()
	opts := []Option{
		WithExtensions(".go", ".md"),
		WithIgnoreDirs("vendor", "node_modules"),
		WithDebounce(100 * time.Millisecond),
		WithRecursive(true),
		WithBuffer(128),
	}

	b.ResetTimer()

	for range b.N {
		w, err := New([]string{tmpDir}, opts...)
		if err != nil {
			b.Fatal(err)
		}

		_ = w.Close()
	}
}

func BenchmarkNew_WithMiddleware(b *testing.B) {
	tmpDir := b.TempDir()
	opts := []Option{
		WithMiddleware(
			MiddlewareRecovery(),
			MiddlewareMetrics(func(_ Op) {}),
		),
	}

	b.ResetTimer()

	for range b.N {
		w, err := New([]string{tmpDir}, opts...)
		if err != nil {
			b.Fatal(err)
		}

		_ = w.Close()
	}
}

// ============================================================================
// Event Conversion Benchmarks
// ============================================================================

func BenchmarkConvertEvent_Create(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Create}

	b.ResetTimer()

	for range b.N {
		_ = convertEvent(fsEvent)
	}
}

func BenchmarkConvertEvent_Write(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	b.ResetTimer()

	for range b.N {
		_ = convertEvent(fsEvent)
	}
}

func BenchmarkConvertEvent_Chmod(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Chmod}

	b.ResetTimer()

	for range b.N {
		_ = convertEvent(fsEvent)
	}
}

// ============================================================================
// Filter Pipeline Benchmarks
// ============================================================================

func BenchmarkPassesFilters_SingleFilter(b *testing.B) {
	w := &Watcher{
		filters: []Filter{FilterExtensions(".go")},
	}

	event := Event{Op: Write, Path: "/tmp/main.go"}

	b.ResetTimer()

	for range b.N {
		_ = w.passesFilters(event)
	}
}

func BenchmarkPassesFilters_MultipleFilters(b *testing.B) {
	w := &Watcher{
		filters: []Filter{
			FilterExtensions(".go"),
			FilterOperations(Write, Create),
			FilterNot(FilterIgnoreDirs("vendor")),
		},
	}

	event := Event{Op: Write, Path: "/tmp/main.go"}

	b.ResetTimer()

	for range b.N {
		_ = w.passesFilters(event)
	}
}

func BenchmarkPassesFilters_ComplexFilterChain(b *testing.B) {
	w := &Watcher{
		filters: []Filter{
			FilterAnd(
				FilterExtensions(".go", ".md"),
				FilterNot(FilterIgnoreDirs("vendor", "node_modules")),
				FilterNot(FilterIgnoreHidden()),
			),
		},
	}

	event := Event{Op: Write, Path: "/tmp/main.go"}

	b.ResetTimer()

	for range b.N {
		_ = w.passesFilters(event)
	}
}

// ============================================================================
// Middleware Pipeline Benchmarks
// ============================================================================

func BenchmarkBuildMiddlewareHandler_NoMiddleware(b *testing.B) {
	w := &Watcher{}

	b.ResetTimer()

	for range b.N {
		_ = w.buildMiddlewareHandler(func(_ Event) {})
	}
}

func BenchmarkBuildMiddlewareHandler_SingleMiddleware(b *testing.B) {
	w := &Watcher{
		middleware: []Middleware{MiddlewareRecovery()},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.buildMiddlewareHandler(func(_ Event) {})
	}
}

func BenchmarkBuildMiddlewareHandler_ThreeMiddleware(b *testing.B) {
	w := &Watcher{
		middleware: []Middleware{
			MiddlewareRecovery(),
			MiddlewareMetrics(func(_ Op) {}),
			MiddlewareRateLimit(100),
		},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.buildMiddlewareHandler(func(_ Event) {})
	}
}

// ============================================================================
// Path Management Benchmarks
// ============================================================================

func BenchmarkShouldSkipDir_DotDir(b *testing.B) {
	w := &Watcher{
		skipDotDirs:    true,
		ignoreDirNames: nil,
	}

	b.ResetTimer()

	for range b.N {
		_ = w.shouldSkipDir(".git")
	}
}

func BenchmarkShouldSkipDir_DefaultIgnore(b *testing.B) {
	w := &Watcher{
		skipDotDirs:    true,
		ignoreDirNames: nil,
	}

	b.ResetTimer()

	for range b.N {
		_ = w.shouldSkipDir("vendor")
	}
}

func BenchmarkShouldSkipDir_CustomIgnore(b *testing.B) {
	w := &Watcher{
		skipDotDirs:    true,
		ignoreDirNames: []string{"custom", "dist", "build"},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.shouldSkipDir("custom")
	}
}

func BenchmarkShouldSkipDir_Allowed(b *testing.B) {
	w := &Watcher{
		skipDotDirs:    true,
		ignoreDirNames: []string{"custom"},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.shouldSkipDir("src")
	}
}

// ============================================================================
// Watcher Stats Benchmarks
// ============================================================================

func BenchmarkStats_Empty(b *testing.B) {
	w := &Watcher{
		watchList: []string{},
		state:     0,
		mu:        sync.RWMutex{},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.Stats()
	}
}

func BenchmarkStats_WithPaths(b *testing.B) {
	paths := make([]string, 100)

	for idx := range 100 {
		paths[idx] = fmt.Sprintf("/path/to/dir%d", idx)
	}

	w := &Watcher{
		watchList: paths,
		state:     flagWatching,
		mu:        sync.RWMutex{},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.Stats()
	}
}

func BenchmarkWatchList_Copy(b *testing.B) {
	paths := make([]string, 100)

	for idx := range 100 {
		paths[idx] = fmt.Sprintf("/path/to/dir%d", idx)
	}

	w := &Watcher{
		watchList: paths,
		mu:        sync.RWMutex{},
	}

	b.ResetTimer()

	for range b.N {
		_ = w.WatchList()
	}
}

// ============================================================================
// Full Pipeline Benchmarks (Event processing)
// ============================================================================

func BenchmarkEmitEvent_NoDebounce(b *testing.B) {
	w := &Watcher{}

	event := Event{Op: Write, Path: "/tmp/test.go"}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	b.ResetTimer()

	for range b.N {
		w.emitEvent(ctx, event, eventCh)
	}
}

func BenchmarkEmitEvent_WithMiddleware(b *testing.B) {
	w := &Watcher{
		middleware: []Middleware{
			MiddlewareRecovery(),
			MiddlewareMetrics(func(_ Op) {}),
		},
	}

	event := Event{Op: Write, Path: "/tmp/test.go"}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	b.ResetTimer()

	for range b.N {
		w.emitEvent(ctx, event, eventCh)
	}
}

func BenchmarkEmitEvent_WithGlobalDebounce(b *testing.B) {
	w := &Watcher{
		debounceInterface: NewGlobalDebouncer(time.Hour), // Never fires during benchmark
	}

	defer w.debounceInterface.Stop()

	event := Event{Op: Write, Path: "/tmp/test.go"}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	b.ResetTimer()

	for range b.N {
		w.emitEvent(ctx, event, eventCh)
	}
}

func BenchmarkEmitEvent_WithPerPathDebounce(b *testing.B) {
	w := &Watcher{
		debounceInterface: NewDebouncer(time.Hour), // Never fires during benchmark
	}

	defer w.debounceInterface.Stop()

	event := Event{Op: Write, Path: "/tmp/test.go"}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	b.ResetTimer()

	for range b.N {
		w.emitEvent(ctx, event, eventCh)
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

func BenchmarkEventAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		e := Event{
			Path:      "/tmp/test.go",
			Op:        Write,
			Timestamp: time.Now(),
			IsDir:     false,
		}

		_ = e
	}
}

func BenchmarkEventString(b *testing.B) {
	e := Event{
		Path:      "/tmp/test.go",
		Op:        Write,
		Timestamp: time.Now(),
		IsDir:     false,
	}

	b.ResetTimer()

	for range b.N {
		_ = e.String()
	}
}

func BenchmarkOpString(b *testing.B) {
	b.ResetTimer()

	for range b.N {
		_ = Write.String()
	}
}
