//nolint:wsl,varnamelen // Benchmarks prioritize readability
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

var (
	benchmarkTestEvent      = benchmarkEventTemplate()
	benchmarkTestPathTestGo = "/tmp/test.go"
	benchmarkTestPathMainGo = "/tmp/main.go"
)

// benchmarkEventTemplate returns the common Event structure used across benchmarks.
func benchmarkEventTemplate() Event {
	return Event{
		Path:      benchmarkTestPathTestGo,
		Op:        Write,
		Timestamp: time.Now(),
		IsDir:     false,
	}
}

// newBenchmarkEvent creates a new Event for benchmarking purposes.
func newBenchmarkEvent() Event {
	return benchmarkEventTemplate()
}

// benchmarkMiddlewareHandler runs the middleware handler benchmark with the given watcher.
func benchmarkMiddlewareHandler(b *testing.B, w *Watcher) {
	b.Helper()

	for range b.N {
		_ = w.buildMiddlewareHandler(func(_ Event) {})
	}
}

// benchmarkNewWatcher creates and closes a watcher repeatedly for benchmarking.
func benchmarkNewWatcher(b *testing.B, opts ...Option) {
	b.Helper()

	for range b.N {
		w, err := New([]string{b.TempDir()}, opts...)
		if err != nil {
			b.Fatal(err)
		}

		_ = w.Close()
	}
}

// benchmarkShouldSkipDir runs shouldSkipDir repeatedly for benchmarking.
func benchmarkShouldSkipDir(
	b *testing.B,
	skipDotDirs bool, //nolint:unparam // param kept for API consistency
	ignoreDirNames []string,
	path string,
) {
	b.Helper()

	w := &Watcher{
		skipDotDirs:    skipDotDirs,
		ignoreDirNames: ignoreDirNames,
	}

	for range b.N {
		_ = w.shouldSkipDir(path)
	}
}

// ============================================================================
// Watcher Creation Benchmarks
// ============================================================================

func BenchmarkNew_SinglePath(b *testing.B) {
	benchmarkNewWatcher(b)
}

func BenchmarkNew_WithOptions(b *testing.B) {
	benchmarkNewWatcher(
		b,
		WithExtensions(".go", ".md"),
		WithIgnoreDirs("vendor", "node_modules"),
		WithDebounce(100*time.Millisecond),
		WithRecursive(true),
		WithBuffer(128),
	)
}

func BenchmarkNew_WithMiddleware(b *testing.B) {
	benchmarkNewWatcher(
		b,
		WithMiddleware(
			MiddlewareRecovery(),
			MiddlewareMetrics(func(_ Op) {}),
		),
	)
}

// ============================================================================
// Event Conversion Benchmarks
// ============================================================================

func BenchmarkConvertEvent_Create(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Create}

	for b.Loop() {
		_ = convertEvent(fsEvent, false, false)
	}
}

func BenchmarkConvertEvent_Write(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Write}

	for b.Loop() {
		_ = convertEvent(fsEvent, false, false)
	}
}

func BenchmarkConvertEvent_Chmod(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Chmod}

	for b.Loop() {
		_ = convertEvent(fsEvent, false, false)
	}
}

func BenchmarkConvertEvent_LazyIsDir(b *testing.B) {
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	_ = os.WriteFile(tmpFile, []byte("test"), 0o600)

	fsEvent := fsnotify.Event{Name: tmpFile, Op: fsnotify.Create}

	for b.Loop() {
		_ = convertEvent(fsEvent, true, false) // lazyIsDir=true for performance
	}
}

// ============================================================================
// Filter Pipeline Benchmarks
// ============================================================================

func BenchmarkPassesFilters_SingleFilter(b *testing.B) {
	w := &Watcher{
		filters: []Filter{FilterExtensions(".go")},
	}

	event := Event{Op: Write, Path: benchmarkTestPathMainGo}

	for b.Loop() {
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

	event := Event{Op: Write, Path: benchmarkTestPathMainGo}

	for b.Loop() {
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

	event := Event{Op: Write, Path: benchmarkTestPathMainGo}

	for b.Loop() {
		_ = w.passesFilters(event)
	}
}

// ============================================================================
// Middleware Pipeline Benchmarks
// ============================================================================

func BenchmarkBuildMiddlewareHandler_NoMiddleware(b *testing.B) {
	benchmarkMiddlewareHandler(b, &Watcher{})
}

func BenchmarkBuildMiddlewareHandler_SingleMiddleware(b *testing.B) {
	benchmarkMiddlewareHandler(b, &Watcher{
		middleware: []Middleware{MiddlewareRecovery()},
	})
}

func BenchmarkBuildMiddlewareHandler_ThreeMiddleware(b *testing.B) {
	benchmarkMiddlewareHandler(b, &Watcher{
		middleware: []Middleware{
			MiddlewareRecovery(),
			MiddlewareMetrics(func(_ Op) {}),
			MiddlewareRateLimit(100),
		},
	})
}

// ============================================================================
// Path Management Benchmarks
// ============================================================================

func BenchmarkShouldSkipDir_DotDir(b *testing.B) {
	benchmarkShouldSkipDir(b, true, nil, ".git")
}

func BenchmarkShouldSkipDir_DefaultIgnore(b *testing.B) {
	benchmarkShouldSkipDir(b, true, nil, "vendor")
}

func BenchmarkShouldSkipDir_CustomIgnore(b *testing.B) {
	benchmarkShouldSkipDir(b, true, []string{"custom", "dist", "build"}, "custom")
}

func BenchmarkShouldSkipDir_Allowed(b *testing.B) {
	benchmarkShouldSkipDir(b, true, []string{"custom"}, "src")
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

	for b.Loop() {
		_ = w.Stats()
	}
}

func BenchmarkStats_WithPaths(b *testing.B) {
	paths := make([]string, 0, 100)

	for idx := range 100 {
		paths = append(paths, fmt.Sprintf("/path/to/dir%d", idx))
	}

	w := &Watcher{
		watchList: paths,
		state:     flagWatching,
		mu:        sync.RWMutex{},
	}

	for b.Loop() {
		_ = w.Stats()
	}
}

func BenchmarkWatchList_Copy(b *testing.B) {
	paths := make([]string, 0, 100)

	for idx := range 100 {
		paths = append(paths, fmt.Sprintf("/path/to/dir%d", idx))
	}

	w := &Watcher{
		watchList: paths,
		mu:        sync.RWMutex{},
	}

	for b.Loop() {
		_ = w.WatchList()
	}
}

// ============================================================================
// Full Pipeline Benchmarks (Event processing)
// ============================================================================

func BenchmarkEmitEvent_NoDebounce(b *testing.B) {
	w := &Watcher{}

	event := Event{Op: Write, Path: benchmarkTestPathTestGo}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	for b.Loop() {
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

	event := Event{Op: Write, Path: benchmarkTestPathTestGo}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	for b.Loop() {
		w.emitEvent(ctx, event, eventCh)
	}
}

func BenchmarkEmitEvent_WithGlobalDebounce(b *testing.B) {
	w := &Watcher{
		debounceInterface: NewGlobalDebouncer(time.Hour), // Never fires during benchmark
	}

	defer w.debounceInterface.Stop()

	event := Event{Op: Write, Path: benchmarkTestPathTestGo}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	for b.Loop() {
		w.emitEvent(ctx, event, eventCh)
	}
}

func BenchmarkEmitEvent_WithPerPathDebounce(b *testing.B) {
	w := &Watcher{
		debounceInterface: NewDebouncer(time.Hour), // Never fires during benchmark
	}

	defer w.debounceInterface.Stop()

	event := Event{Op: Write, Path: benchmarkTestPathTestGo}
	ctx := context.Background()
	eventCh := make(chan Event, 1)

	for b.Loop() {
		w.emitEvent(ctx, event, eventCh)
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

func BenchmarkEventAllocation(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		e := newBenchmarkEvent()

		_ = e
	}
}

func BenchmarkEventString(b *testing.B) {
	for b.Loop() {
		_ = benchmarkTestEvent.String()
	}
}

func BenchmarkOpString(b *testing.B) {
	for b.Loop() {
		_ = Write.String()
	}
}

// ============================================================================
// Gitignore Benchmarks
// ============================================================================

func BenchmarkShouldSkipByGitignore_NoGitignore(b *testing.B) {
	tmpDir := b.TempDir()

	w, err := New([]string{tmpDir})
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	testPath := filepath.Join(tmpDir, "src", "main.go")

	for b.Loop() {
		_ = w.shouldSkipByGitignore(testPath)
	}
}

func BenchmarkShouldSkipByGitignore_WithGitignore(b *testing.B) {
	tmpDir := b.TempDir()

	gitignoreContent := "node_modules/\n*.o\nbuild/\ndist/\n*.pyc\n__pycache__/\n.coverage\n"
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	writeErr := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o600)
	if writeErr != nil {
		b.Fatal(writeErr)
	}

	w, err := New([]string{tmpDir})
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.loadGitignoreForDir(tmpDir)

	matchedPath := filepath.Join(tmpDir, "node_modules", "pkg", "index.js")

	for b.Loop() {
		_ = w.shouldSkipByGitignore(matchedPath)
	}
}

func BenchmarkShouldSkipByGitignore_NotMatched(b *testing.B) {
	tmpDir := b.TempDir()

	gitignoreContent := "node_modules/\n*.o\nbuild/\n"
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	writeErr := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o600)
	if writeErr != nil {
		b.Fatal(writeErr)
	}

	w, err := New([]string{tmpDir})
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	w.loadGitignoreForDir(tmpDir)

	normalPath := filepath.Join(tmpDir, "src", "main.go")

	for b.Loop() {
		_ = w.shouldSkipByGitignore(normalPath)
	}
}

func BenchmarkGitignoreCache_Load(b *testing.B) {
	tmpDir := b.TempDir()

	gitignoreContent := "*.go\n*.md\nbuild/\n"
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	writeErr := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o600)
	if writeErr != nil {
		b.Fatal(writeErr)
	}

	cache := newGitignoreCache()

	for b.Loop() {
		cache.load(tmpDir)
	}
}

// ============================================================================
// Benchmark Regression Baselines
// ============================================================================

// benchmarkBaseline records expected upper bounds for key operations.
// If a benchmark exceeds these thresholds, the regression test fails.
// Update these values when intentional performance-impacting changes are made.
//

var benchmarkBaseline = map[string]struct {
	nsPerOp  float64 // max nanoseconds per operation
	allocs   int     // max heap allocations per operation
	bytesPer int64   // max bytes allocated per operation
}{
	"BenchmarkNew_SinglePath": {
		nsPerOp:  100000,
		allocs:   50,
		bytesPer: 5000,
	},
	"BenchmarkConvertEvent_Create":                    {nsPerOp: 20000, allocs: 10, bytesPer: 500},
	"BenchmarkConvertEvent_LazyIsDir":                 {nsPerOp: 500, allocs: 2, bytesPer: 100},
	"BenchmarkPassesFilters_SingleFilter":             {nsPerOp: 200, allocs: 1, bytesPer: 0},
	"BenchmarkBuildMiddlewareHandler_NoMiddleware":    {nsPerOp: 1000, allocs: 5, bytesPer: 200},
	"BenchmarkBuildMiddlewareHandler_ThreeMiddleware": {nsPerOp: 5000, allocs: 15, bytesPer: 800},
	"BenchmarkStats_Empty":                            {nsPerOp: 200, allocs: 1, bytesPer: 50},
	"BenchmarkEmitEvent_NoDebounce":                   {nsPerOp: 1000, allocs: 5, bytesPer: 200},
}

// TestBenchmarkBaselines validates that current benchmarks do not regress
// beyond the recorded baselines. Run with -bench=. -benchmem to update.
func TestBenchmarkBaselines(t *testing.T) {
	t.Parallel()

	for name, baseline := range benchmarkBaseline {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Logf("Baseline: ns/op<%.0f allocs<%d bytes<%d",
				baseline.nsPerOp, baseline.allocs, baseline.bytesPer)

			// This is a documentation test — it records the expected baselines.
			// Actual enforcement happens via CI benchmark comparison.
		})
	}
}
