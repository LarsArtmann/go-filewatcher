# Agent Guide: go-filewatcher

**Go 1.26.2** | `github.com/larsartmann/go-filewatcher` | **MIT License**

---

## Critical Commands

```bash
# Using Nix flake (recommended)
nix develop          # Enter development shell with Go and tools
direnv allow         # Auto-load environment on cd (requires direnv)

# Inside dev shell or with Go installed:
check       # Full quality: vet + lint + test
ci          # Full CI: tidy + fmt + vet + lint + test
lint-fix    # Auto-fix linter issues

# Direct Go commands
GOWORK=off go test -race ./...
GOWORK=off golangci-lint run ./...
```

---

## Non-Obvious Conventions

### Error Handling: Standard Library

Uses `errors` and `fmt` from the standard library:

```go
import (
    "errors"
    "fmt"
)

// Creating sentinel errors
var ErrPathNotFound = errors.New("path not found")

// Wrapping with context
return fmt.Errorf("path %q: %w", path, err)

// Checking
if errors.Is(err, ErrPathNotFound) { ... }
```

### Single Package Layout

All code in **root package** (`filewatcher`). No `internal/` or `pkg/` subdirectories — all code lives in the package root.

### File Organization

| File                  | Responsibility                                        |
| --------------------- | ----------------------------------------------------- |
| `watcher.go`          | Public API: New, Watch, Add, Remove, WatchList, Stats |
| `watcher_internal.go` | Event processing: watchLoop, middleware, emitEvent    |
| `watcher_walk.go`     | Directory walking: addPath, walkAndAddPaths           |
| `filter.go`           | All Filter functions                                  |
| `middleware.go`       | All Middleware functions                              |
| `debouncer.go`        | Debouncer + GlobalDebouncer                           |
| `event.go`            | Op type, Event type, JSON/Text marshaling             |
| `errors.go`           | Sentinel errors                                       |
| `options.go`          | Functional options                                    |

---

## Critical Gotchas

### 1. Middleware Order Is Reversed

```go
WithMiddleware(
    MiddlewareRecovery(),   // Runs LAST (innermost)
    MiddlewareLogging(nil), // Runs FIRST (outermost)
)
```

### 2. Two Debounce Modes (Different Semantics)

```go
WithDebounce(d)           // Global: ALL events → ONE callback
WithPerPathDebounce(d)    // Per-path: EACH file → separate callback
```

### 3. Strict Linter: `exhaustruct`

**All struct fields must be initialized** — no zero values allowed:

```go
// WRONG — fails lint
w := &Watcher{fswatcher: fs}

// RIGHT — all fields
w := &Watcher{
    fswatcher: fs,
    paths: paths,
    recursive: true,
    // ... every field
}
```

### 4. Required: `t.Parallel()` in All Tests

```go
func TestXxx(t *testing.T) {
    t.Parallel()  // REQUIRED (enforced by paralleltest linter)
    // ...
}
```

### 5. Event Priority (Multiple Ops)

Create > Write > Remove > Rename — highest wins.

### 6. Chmod Events Ignored

Not mapped to any Op, `convertEvent()` returns `nil`.

### 7. Exported Global with Nolint

```go
//nolint:gochecknoglobals // Intentionally exported for users
var DefaultIgnoreDirs = []string{".git", "vendor", ...}
```

Don't remove the nolint — this is intentional.

---

## Key Patterns

| Pattern            | Where                                          |
| ------------------ | ---------------------------------------------- |
| Functional Options | `options.go` — `type Option func(*Watcher)`    |
| Middleware Chain   | `middleware.go` — applied in **reverse** order |
| Filter Composition | `filter.go` — `FilterAnd()`, `FilterOr()`      |

---

## Linter Cheat Sheet

50+ linters enabled. Key ones that bite:

| Linter             | Rule                                  |
| ------------------ | ------------------------------------- |
| `exhaustruct`      | All struct fields must be initialized |
| `wrapcheck`        | All errors must be wrapped            |
| `paralleltest`     | All tests must use `t.Parallel()`     |
| `gochecknoglobals` | No globals unless `//nolint`          |
| `gci`              | Import order matters                  |

Run `just lint-fix` — it auto-fixes many issues.

---

## Dependencies

```
github.com/fsnotify/fsnotify    # Core file watching
github.com/LarsArtmann/gogenfilter  # Generated code detection
golang.org/x/time                # rate.Limiter for rate limiting middleware
```

## Named Types (phantom types)

Plain `type X string` named types for compile-time type safety on path-like strings:

| Type           | Purpose                             |
| -------------- | ----------------------------------- |
| `EventPath`    | Event file/directory paths          |
| `RootPath`     | Root directory paths during walking |
| `DebounceKey`  | Debouncer keys                      |
| `LogSubstring` | Log substring assertions (tests)    |
| `TempDir`      | Temp directory paths (tests)        |
| `OpString`     | Operation names                     |

**Usage:** Use constructor functions (e.g., `NewEventPath()`, `NewRootPath()`).

**EventPath has domain methods:** `.Base()`, `.Dir()`, `.Ext()`, `.Join()` for path operations.

---

## Known Issues

### Flaky Tests

These tests are timing-sensitive and may fail intermittently:

| Test                               | Reason                                                                                     |
| ---------------------------------- | ------------------------------------------------------------------------------------------ |
| `TestWatcher_Stats_Metrics`        | Counts `EventsProcessed` but filesystem write coalescing may produce 2 events instead of 1 |
| `TestWatcher_Watch_WithMiddleware` | Similar timing issue with middleware call counting                                         |

This is a pre-existing issue unrelated to the `go-branded-id` integration.
