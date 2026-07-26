# Agent Guide: go-filewatcher

**Go 1.26.5** | `github.com/larsartmann/go-filewatcher/v2` | **MIT License**

> **Companion docs:** [FEATURES.md](./FEATURES.md) (feature inventory) ·
> [ROADMAP.md](./ROADMAP.md) (long-term direction) · [TODO_LIST.md](./TODO_LIST.md)
> (actionable work) · [CHANGELOG.md](./CHANGELOG.md) (release history).
> This file is for enduring, hard-to-discover-from-code context only.

---

## Critical Commands

```bash
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
nix run .#lint-tests     # Run linter (--tests explicit)
nix run .#bench          # Run benchmarks
nix run .#bench-baseline # Capture benchmark baseline (run from repo root)
nix run .#bench-diff     # Diff benchmarks vs baseline (hermetic benchstat)
nix run .#coverage       # Generate coverage report
nix run .#fmt            # Format Go code (writes — run from repo root)
nix run .#tidy           # Run go mod tidy (writes — run from repo root)
nix run .                # Default = check

# Nix quality gates
nix flake check          # Run all checks (build, test, lint, fmt, vet, examples-build)
nix build .              # Validate reproducible build
nix fmt                  # Format .nix files

# Inside dev shell (aliases are set automatically):
check       # nix run .#check
ci          # nix run .#ci
lint        # nix run .#lint
lint-fix    # nix run .#lint-fix
test        # nix run .#test
```

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
```

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

| File                   | Responsibility                                                                              |
| ---------------------- | ------------------------------------------------------------------------------------------- |
| `watcher.go`           | Public API: New, Watch, Add, AddRecursive, Remove, Reset, WatchList, Stats                  |
| `backend.go`           | watchBackend interface + fsnotifyBackend adapter (test seam for fake backend injection)     |
| `watcher_internal.go`  | Event processing: watchLoop, middleware, emitEvent, debugLog, handleError                   |
| `watcher_walk.go`      | Directory walking: addPath, walkAndAddPaths, addBatch, symlink resolution, budget detection |
| `watcher_gitignore.go` | .gitignore loading and matching: gitignoreCache, shouldSkipByGitignore                      |
| `watcher_selfheal.go`  | Self-healing: selfHealLoop, attemptSelfHeal, failed path tracking                           |
| `watcher_poll.go`      | Polling mode: pollLoop for NFS/FUSE environments                                            |
| `filter.go`            | All Filter functions + FilterWithMeta and combinators                                       |
| `filter_gogen.go`      | Generated-code detection filter (gogenfilter v3 integration)                                |
| `middleware.go`        | All Middleware functions (circuit breaker, error batch, correlation, exponential backoff)   |
| `metrics.go`           | PrometheusCollector, StatsFunc, CounterMetric, GaugeMetric                                  |
| `otel.go`              | OTelMiddleware, OTelSpan interface                                                          |
| `debouncer.go`         | Debouncer + GlobalDebouncer                                                                 |
| `event.go`             | Op type, Event type, JSON/Text marshaling                                                   |
| `errors.go`            | Sentinel errors, ErrorCode, ErrorCategory, WatcherError                                     |
| `options.go`           | Functional options (WithGitignore, WithExcludePaths, WithMaxWatches, etc.)                  |
| `phantom_types.go`     | Compile-time phantom types (EventPath, RootPath, DebounceKey, OpString, etc.)               |

### Examples (`examples/`)

Separate Go programs demonstrating usage. Each subdirectory is a standalone
`main` package that imports the library. Shared helpers (including
`demo.MustWatch`) live in `examples/demo/`. Build with `go run ./examples/<name>`
or `go build ./examples/...`. Part of the same Go module (no separate go.mod).

### Website (`website/`)

Separate Astro + Starlight documentation site with its own `flake.nix`, deployed to
Firebase Hosting at `filewatcher.lars.software`. Not part of the Go module — has its
own `package.json` and Node toolchain. Build with `cd website && nix run .#build`.

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

### 11. Graceful ENOSPC Handling

`fswatcher.Add()` errors (including ENOSPC) do not abort the entire walk. Instead:

- The error is logged via `handleError()`
- The `watchErrors` atomic counter is incremented
- Walking continues to add remaining directories
- `Stats.WatchErrors` tracks how many paths failed to add
- This allows the watcher to start in degraded mode instead of failing entirely

### 12. Inotify Budget Awareness

- `maxWatches` is auto-detected from `/proc/sys/fs/inotify/max_user_watches` on Linux
- Override with `WithMaxWatches(n)`
- When budget is exhausted, directories are skipped silently
- `Stats.WatchLimit` and `Stats.WatchBudgetUsed` track budget usage

### 13. .gitignore-Aware Walking

- Enabled by default (`WithGitignore(true)`)
- Loads `.gitignore` files during directory walking
- Directories matching gitignore patterns are skipped (not added to inotify)
- Uses `github.com/sabhiram/go-gitignore` (zero transitive deps)
- gitignore cache is stored per-directory for hierarchical matching

### 14. Batched Watch Registration

- Directories are collected during walk and added in batches of 1000
- `runtime.Gosched()` is called between batches to yield to event processing
- Reduces startup latency for large directory trees

### 15. Path-Level Exclusions

- `WithExcludePaths(paths...)` excludes absolute paths during walk
- Prefix matching: excluding `/home/user/forks` skips all subdirectories too
- Walk-time only: does not affect event filtering

### 16. Remove() Cleans Up Subdirectories

`Remove(path)` removes all subdirectory watches under the given path, not just
the top-level directory. This prevents watch leaks.

### 17. Reset() Method

`Reset()` clears runtime state while preserving configuration (filters,
middleware, debounce, options). Allows re-calling `Watch()` after `Close()`
without rebuilding from scratch.

---

## Key Patterns

| Pattern              | Where                                                                             |
| -------------------- | --------------------------------------------------------------------------------- |
| Functional Options   | `options.go` — `type Option func(*Watcher)`                                       |
| Middleware Chain     | `middleware.go` — applied in **reverse** order                                    |
| Filter Composition   | `filter.go` — `FilterAnd()`, `FilterOr()`                                         |
| `resolve*Defaults`   | `middleware.go` — see [Default-guard convention](#default-guard-convention) below |
| `baseDebouncer.stop` | `debouncer.go` — lock/markStopped/cleanup/unlock/wait in one place                |
| Backend Abstraction  | `backend.go` — `watchBackend` interface; `withBackend()` injects fakes            |
| `newTestWatcher`     | `testing_helpers_test.go:432` — standard `New + cleanup` for all tests            |

### Default-guard convention

Every middleware that accepts a tunable (duration, count, threshold) must
substitute a **named const** when the caller passes a non-positive value — never
a magic literal. The decision of _where_ the defaulting lives has one rule:

> **Shared defaulting → `resolve*Defaults` helper. Unique defaulting → inline guard.**

- **Two or more functions share the same defaulting** → extract a `resolve*Defaults`
  helper. This is the DRY path: one named const, one guard, one test target.
- **Exactly one function uses the default** → keep an inline `if x <= 0` guard with
  a named const. A helper for a single caller is indirection without benefit.

**Worked example** (`middleware.go`):

```go
// SHARED — two middlewares (SlidingWindowRateLimit + ErrorRateLimit) reuse the
// same window default and the same non-positive guard. Extract.
const defaultRateLimitWindow = time.Second

func resolveRateLimitDefaults(maxValue, defaultMax int, window time.Duration) (int, time.Duration) {
    if maxValue <= 0 {
        maxValue = defaultMax
    }
    if window <= 0 {
        window = defaultRateLimitWindow
    }
    return maxValue, window
}

func MiddlewareSlidingWindowRateLimit(maxEvents int, window time.Duration) Middleware {
    maxEvents, window = resolveRateLimitDefaults(maxEvents, defaultSlidingWindowEvents, window)
    // ...
}

// UNIQUE — only MiddlewareThrottle uses this default. Inline guard, named const.
const defaultThrottleEvents = 100

func MiddlewareThrottle(maxEvents, burst int) Middleware {
    if maxEvents <= 0 {
        maxEvents = defaultThrottleEvents
    }
    // ...
}
```

The three shared helpers are `resolveRateLimitDefaults`, `resolveBatchDefaults`,
and `resolveMaxFailures`. Each has direct table-driven coverage, and
`TestMiddlewareDefaultConsts_AllUsed` guards the whole inventory: if a default
const loses its call site in a refactor, the test fails before linters notice.

### Middleware Resource Cleanup

`Middleware` is `func(Handler) Handler` — a function type with no lifecycle hook.
Middleware that hold resources (e.g., file handles) cannot close them automatically.
The watcher provides `WithCleanup(fn func() error)` to register cleanup functions
called on `Close()` (after all goroutines/channels are torn down) and cleared on `Reset()`.

Pattern: factory returns `(Middleware, func() error)`; caller pairs them:

```go
mw, closeLog := NewFileLogMiddleware("audit.log")
watcher, _ := New(paths, WithMiddleware(mw), WithCleanup(closeLog))
defer watcher.Close() // closeLog is called automatically
```

`MiddlewareWriteFileLog` wraps `NewFileLogMiddleware` for backward compat but
does NOT close the file — long-lived watchers should use `NewFileLogMiddleware`.

---

## CI Workflows

| Workflow               | Purpose                                                                     |
| ---------------------- | --------------------------------------------------------------------------- |
| `ci.yml`               | Test (race + coverage), lint, examples-build, benchmark                     |
| `commitlint.yml`       | Validates PR commit subjects follow conventional-commit format              |
| `docs-consistency.yml` | Checks README.md vs API_STABILITY.md deprecation claims don't drift         |
| `release-please.yml`   | Opens release PRs from conventional commits (automates CHANGELOG + version) |
| `release.yml`          | Triggered on `v*` tags — creates GitHub Release                             |

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
github.com/fsnotify/fsnotify          # Core file watching (v1.10.1)
github.com/LarsArtmann/gogenfilter/v3 # Generated code detection (v3.2.0, local replace)
github.com/sabhiram/go-gitignore      # .gitignore pattern matching (zero transitive deps)
golang.org/x/time/rate                # rate.Limiter for MiddlewareThrottle
```

### gogenfilter v3 API

The library depends on `github.com/LarsArtmann/gogenfilter/v3` (currently v3.2.0).
The v3 API differs from v0/v2 in these ways (relevant when upgrading consumers):

- `NewFilter` returns `(*Filter, error)` — must handle error
- `WithFilterOptions` returns `(FilterConfig, error)` — must handle error
- `Enabled()` / `Disabled()` removed — auto-enables when configured
- `ShouldFilter` renamed to `Filter` — `f.Filter(path)` returns `(bool, error)`
- Generators: `FilterOapi`, `FilterDeepcopy`, `FilterWire`, `FilterMoq`

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

### Pre-existing Linter Warning

`watcher_coverage_test.go:1` has an unused `modernize` nolint directive — do not fix (unrelated to current work).
