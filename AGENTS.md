# Agent Guide for go-filewatcher

**Project**: go-filewatcher — A high-level, composable file system watcher for Go  
**Go Version**: 1.26.1  
**Module**: `github.com/larsartmann/go-filewatcher`

---

## Quick Start

```bash
# Install dependencies
go mod download

# Build
just build

# Run all checks (format, vet, lint, test)
just check

# Run tests with race detector
just test

# Run tests with coverage
just test-cover
```

---

## Essential Commands

| Command | Description |
|---------|-------------|
| `just build` | Build the project (`go build ./...`) |
| `just test` | Run tests with race detector |
| `just test-v` | Run tests with verbose output |
| `just test-cover` | Generate HTML coverage report |
| `just lint` | Run golangci-lint |
| `just lint-fix` | Run linter with auto-fix |
| `just vet` | Run `go vet ./...` |
| `just check` | Full quality gate: tidy, fmt, vet, lint, test |
| `just ci` | CI pipeline: tidy, fmt, vet, lint, test |
| `just bench` | Run benchmarks |
| `just tidy` | Tidy go.mod/go.sum |
| `just fmt` | Format code with `go fmt` |
| `just clean` | Clean build cache |

---

## Project Structure

```
go-filewatcher/
├── *.go              # Core package files (single package)
├── *_test.go         # Test files
├── justfile          # Task runner commands
├── go.mod            # Module definition
├── .golangci.yml     # Linter configuration
├── examples/         # Usage examples
│   ├── basic/        # Simple extension filter + debounce
│   ├── middleware/   # Middleware chain example
│   └── per-path-debounce/  # Per-path debounce example
├── pkg/errors/       # Custom error types (apperrors.go)
└── docs/status/      # Project status documentation
```

**Key Files**:
- `watcher.go` — Main Watcher struct and lifecycle (New, Watch, Close, Add, Remove)
- `options.go` — Functional options (WithDebounce, WithExtensions, etc.)
- `filter.go` — Event filters (FilterExtensions, FilterIgnoreDirs, etc.)
- `middleware.go` — Middleware chain (MiddlewareLogging, MiddlewareRecovery, etc.)
- `debouncer.go` — Debouncer implementations (Debouncer, GlobalDebouncer)
- `event.go` — Event types and Op constants
- `errors.go` — Sentinel errors using cockroachdb/errors

---

## Architecture & Design Patterns

### Core Design Principles

1. **Functional Options Pattern** — All configuration via `Option` funcs
2. **Middleware Chain** — Cross-cutting concerns via composable middleware
3. **Filter Composition** — AND/OR/NOT logic for event filtering
4. **Context-First** — All async operations accept `context.Context`
5. **Error Handling** — `cockroachdb/errors` for error wrapping and stack traces
6. **Channel-Based** — Event streaming via `<-chan Event`

### Watcher Lifecycle

```go
// 1. Create
w, err := filewatcher.New([]string{"./src"}, opts...)
if err != nil { log.Fatal(err) }
defer w.Close()

// 2. Start watching
events, err := w.Watch(ctx)
if err != nil { log.Fatal(err) }

// 3. Process events
for event := range events {
    // handle event
}
// Channel closes when context cancelled or watcher closed
```

### Concurrency Model

- **Thread-safe**: All public methods use `sync.RWMutex`
- **Single goroutine**: `watchLoop` runs in one goroutine, handles fsnotify events
- **Debouncing**: Timer-based, mutex-protected
- **Graceful shutdown**: Context cancellation or `Close()` stops the watcher

---

## Code Conventions

### Naming

- **Exported**: PascalCase (`Watcher`, `WithDebounce`, `FilterExtensions`)
- **Unexported**: camelCase (`watchLoop`, `addPath`, `shouldSkipDir`)
- **Interfaces**: `-er` suffix (`DebouncerInterface`)
- **Options**: `WithXxx` pattern (`WithDebounce`, `WithExtensions`)
- **Filters**: `FilterXxx` pattern (`FilterExtensions`, `FilterIgnoreDirs`)
- **Middleware**: `MiddlewareXxx` pattern (`MiddlewareLogging`, `MiddlewareRecovery`)

### Error Handling

```go
// Use cockroachdb/errors for wrapping
import "github.com/cockroachdb/errors"

// Sentinel errors (defined in errors.go)
var ErrWatcherClosed = errors.New("watcher is closed")

// Wrapping with context
return errors.Wrapf(err, "adding watch path %q", path)
return errors.WithStack(ErrNoPaths)

// Checking errors
if errors.Is(err, ErrPathNotFound) { ... }
```

### Struct Tags & Comments

```go
//nolint:gochecknoglobals // Exported for user reference
var DefaultIgnoreDirs = []string{...}

// Compile-time interface check
var _ io.Closer = (*Watcher)(nil)
```

---

## Testing Patterns

### Test Structure

```go
func TestXxx(t *testing.T) {
    t.Parallel()  // Always use parallel tests

    tmpDir := t.TempDir()  // Use temp dirs for file operations

    // Test code...
}
```

### Key Testing Patterns

- **Parallel tests**: All tests use `t.Parallel()`
- **Temp directories**: Use `t.TempDir()` for isolation
- **Race detection**: Run with `-race` flag (enabled in justfile)
- **Error checking**: Use `errors.Is()` for sentinel errors
- **Cleanup**: Use `defer` for watcher cleanup

### Running Tests

```bash
# All tests with race detection
just test

# Verbose output
just test-v

# Coverage report
just test-cover  # Generates coverage.html

# Specific test
go test -race -run TestWatcher_Watch ./...
```

---

## Linter Configuration

Uses **golangci-lint** with aggressive settings (`.golangci.yml`):

**Enabled linters** (key ones):
- `errcheck`, `errorlint`, `wrapcheck` — Error handling
- `staticcheck`, `gosimple`, `unused` — Static analysis
- `gocritic`, `revive` — Code quality
- `govet`, `ineffassign` — Correctness
- `exhaustruct` — Struct initialization
- `paralleltest` — Parallel test enforcement
- `gosec` — Security

**Run linter**:
```bash
just lint       # Check only
just lint-fix   # Auto-fix where possible
```

---

## Dependencies

```go
require (
    github.com/cockroachdb/errors v1.12.0  // Error handling with stack traces
    github.com/fsnotify/fsnotify v1.9.0     // Core file system notifications
)
```

**No other external dependencies** — keep it minimal.

---

## Common Tasks

### Adding a New Option

```go
// In options.go
func WithXxx(value SomeType) Option {
    return func(w *Watcher) {
        w.xxx = value
    }
}

// In watcher.go — add field to Watcher struct
type Watcher struct {
    // ... existing fields ...
    xxx SomeType
}
```

### Adding a New Filter

```go
// In filter.go
type Filter func(event Event) bool

func FilterXxx(params) Filter {
    return func(event Event) bool {
        // return true to keep, false to discard
    }
}
```

### Adding Middleware

```go
// In middleware.go
type Middleware func(Handler) Handler
type Handler func(ctx context.Context, event Event) error

func MiddlewareXxx() Middleware {
    return func(next Handler) Handler {
        return func(ctx context.Context, event Event) error {
            // pre-processing
            err := next(ctx, event)
            // post-processing
            return err
        }
    }
}
```

---

## Gotchas & Important Notes

1. **DefaultIgnoreDirs** is an exported global (nolint:gochecknoglobals) — users can reference it
2. **Recursive watching** is ON by default — subdirectories auto-added
3. **Dot directories** are skipped by default — use `WithSkipDotDirs(false)` to watch them
4. **Buffer size** defaults to 64 — use `WithBuffer(size)` for high-volume scenarios
5. **Debouncing**: 
   - `WithDebounce()` = global (all events coalesced)
   - `WithPerPathDebounce()` = per-path (each file debounced independently)
6. **Middleware order**: Applied in reverse (last added runs first)
7. **Event priority**: Create > Write > Remove > Rename (when multiple ops occur)
8. **Chmod events** are ignored — not mapped to any Op

---

## Related Projects

- Built on [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
- Follows conventions from [go-cqrs-lite](https://github.com/larsartmann/go-cqrs-lite)

---

## License

Proprietary — See LICENSE file.
