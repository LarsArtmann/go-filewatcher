# Agent Guide: go-filewatcher

**Go 1.26.2** | `github.com/larsartmann/go-filewatcher/v2` | **MIT License**

---

## Critical Commands

````bash
# Using Nix flake (recommended)
nix develop              # Enter development shell with Go and tools
direnv allow             # Auto-load environment on cd (requires direnv)

# Nix apps (run from anywhere, no need to be in dev shell)
nix run .#check          # Full quality: vet + lint + test
nix run .#ci             # Full CI: tidy + fmt + vet + lint + test
nix run .#lint-fix       # Auto-fix linter issues
nix run .#test           # Run tests with -race
nix run .#test-v         # Run tests with -race -v
nix run .#lint           # Run linter
nix run .#bench          # Run benchmarks
nix run .#coverage       # Generate coverage report
nix run .#fmt            # Format Go code
nix run .#tidy           # Run go mod tidy
nix run .                # Default = check

# Nix quality gates
nix flake check          # Run all checks (build, test, lint, fmt, vet)
nix build .              # Validate reproducible build
nix fmt                  # Format .nix files

# Inside dev shell (aliases are set automatically):
check       # nix run .#check
ci          # nix run .#ci
lint        # nix run .#lint
lint-fix    # nix run .#lint-fix
test        # nix run .#test

## Updating vendorHash

When `go.mod` or `go.sum` changes, `vendorHash` in `flake.nix` must be updated:

```bash
# 1. Update dependencies
go get github.com/some/pkg@latest
# or: go mod tidy

# 2. Update vendorHash (Nix will compute the new hash)
nix flake update

# 3. Verify everything still works
nix run .#check
````

If `nix flake update` fails with a hash mismatch, set a temporary placeholder and rebuild:

```bash
# In flake.nix, set vendorHash to an empty string temporarily:
vendorHash = "";  # Will show correct hash in error message

# Then run:
nix build .  # Error will show correct hash

# Copy the hash from the error and set it properly:
vendorHash = "sha256-XXXX...";
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

| File                  | Responsibility                                                             |
| --------------------- | -------------------------------------------------------------------------- |
| `watcher.go`          | Public API: New, Watch, Add, AddRecursive, Remove, WatchList, Stats        |
| `watcher_internal.go` | Event processing: watchLoop, middleware, emitEvent, debugLog, handleError  |
| `watcher_walk.go`     | Directory walking: addPath, walkAndAddPaths, symlink resolution            |
| `watcher_poll.go`     | Polling goroutine: pollLoop, snapshot-based change detection               |
| `filter.go`           | All Filter functions                                                       |
| `middleware.go`       | All Middleware functions (circuit breaker, error batch, correlation, etc.) |
| `debouncer.go`        | Debouncer + GlobalDebouncer                                                |
| `event.go`            | Op type, Event type, JSON/Text marshaling                                  |
| `errors.go`           | Sentinel errors, ErrorCode, ErrorCategory, WatcherError                    |
| `options.go`          | Functional options                                                         |

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

### 8. WithDebug is Active (not a stub)

`WithDebug(logger)` wires real debug logging throughout the pipeline. The `debugLog` helper checks `w.debug` and calls `w.debugLogger.Debug()`. Log calls are in `watchLoop`, `processEvent`, `emitEvent`, `handleError`, `handleNewDirectory`, and `pollLoop`.

### 9. WithPolling is Active (not a stub)

`WithPolling(true)` starts a `pollLoop` goroutine in `Watch()` that maintains a filesystem snapshot and detects new/modified/removed files at `pollInterval`. Works alongside fsnotify for NFS/FUSE environments.

### 10. Circuit Breaker States

`MiddlewareCircuitBreaker` uses three states: `CircuitClosed` → `CircuitOpen` → `CircuitHalfOpen`. In half-open, only one event passes through to test recovery.

### 8. WithDebug is Active (not a stub)

`WithDebug(logger)` wires real debug logging throughout the pipeline. The `debugLog` helper checks `w.debug` and calls `w.debugLogger.Debug()`. Log calls are in `watchLoop`, `processEvent`, `emitEvent`, `handleError`, `handleNewDirectory`, and `pollLoop`.

### 9. WithPolling is Active (not a stub)

`WithPolling(true)` starts a `pollLoop` goroutine in `Watch()` that maintains a filesystem snapshot and detects new/modified/removed files at `pollInterval`. Works alongside fsnotify for NFS/FUSE environments.

### 10. Circuit Breaker States

`MiddlewareCircuitBreaker` uses three states: `CircuitClosed` → `CircuitOpen` → `CircuitHalfOpen`. In half-open, only one event passes through to test recovery.

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

Run `nix run .#lint-fix` — it auto-fixes many issues.

---

## Dependencies

```
github.com/fsnotify/fsnotify         # Core file watching
github.com/LarsArtmann/gogenfilter  # Generated code detection (v3, local replace)
golang.org/x/time/rate              # rate.Limiter for rate limiting middleware
```

### gogenfilter v3 API

Uses `replace` directive in `go.mod` pointing to `../gogenfilter` (module path issue: v3.0.0 doesn't include `/v3` in module path).

**Breaking changes from v3:**

- `NewFilter` returns `(*Filter, error)` — must handle error
- `WithFilterOptions` returns `(FilterConfig, error)` — must handle error
- `Enabled()` / `Disabled()` removed — auto-enables when configured
- `ShouldFilter` renamed to `Filter` — `f.Filter(path)` returns `(bool, error)`
- New generators: `FilterOapi`, `FilterDeepcopy`, `FilterWire`, `FilterMoq`

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

### Pre-existing Linter Warning

`watcher_coverage_test.go:1` has an unused `modernize` nolint directive — do not fix (unrelated to current work).
