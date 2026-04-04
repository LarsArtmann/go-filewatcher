# go-filewatcher — Comprehensive SDK Review & Status Report

**Date:** 2026-04-04 07:00 CEST
**Project:** `github.com/larsartmann/go-filewatcher`
**Location:** `/Users/larsartmann/projects/go-filewatcher/`
**Reviewer:** Crush (Parakletos AI)
**Scope:** Full SDK review — architecture, correctness, concurrency, API design, ergonomics, test coverage, documentation

---

## Executive Summary

A well-architected file watcher SDK wrapping `fsnotify` with composable filters, middleware chains, and configurable debouncing. The API surface is clean and idiomatic Go. **However, the review uncovered 4 critical bugs** (data race, lying method signature, missing idempotency guard, lock-level mismatch) that must be fixed before any production use.

**Verdict:** Solid foundation with good bones. Critical bugs are all localized and fixable in under an hour total. After fixes, this is a strong v0.1.0 candidate.

---

## Quality Gates

| Gate                 | Status   | Details                                     |
| -------------------- | -------- | ------------------------------------------- |
| All tests passing    | ✅ 46/46 | `go test -race -count=1 ./...` passes clean |
| Race detector clean  | ✅       | No data races detected by `-race` flag      |
| `go vet` clean       | ✅       | No issues                                   |
| Build clean          | ✅       | `go build ./...` succeeds                   |
| Coverage             | ⚠️ 84.8% | Down from reported 86.1% — below 90% target |
| Critical bugs        | ❌ 4     | See section D                               |
| Linter config        | ✅       | `.golangci.yml` with 55+ linters            |
| Dependencies minimal | ✅       | Only `fsnotify` + `cockroachdb/errors`      |

---

## File Inventory

| File                 | Lines    | Purpose                                                       | Coverage |
| -------------------- | -------- | ------------------------------------------------------------- | -------- |
| `watcher.go`         | 433      | Core: `New()`, `Watch(ctx)→<-chan Event`, `Add()`, `Close()`  | ~85%     |
| `options.go`         | 83       | 9 functional options                                          | 100%     |
| `filter.go`          | 149      | 11 composable filters (Extensions, IgnoreDirs, Hidden, Glob…) | 80-100%  |
| `debouncer.go`       | 119      | Per-key `Debouncer` + `GlobalDebouncer`                       | 66-100%  |
| `middleware.go`      | 131      | 7 middleware (Logging, Recovery, RateLimit, Metrics…)         | 0-100%   |
| `errors.go`          | 15       | 4 sentinel errors with `cockroachdb/errors`                   | 100%     |
| `event.go`           | 51       | `Op` type (Create/Write/Remove/Rename) + `Event` struct       | 100%     |
| `doc.go`             | 61       | Package documentation with examples                           | N/A      |
| **Source total**     | **1042** |                                                               |          |
| `watcher_test.go`    | 557      | 14 integration tests (real filesystem)                        |          |
| `filter_test.go`     | 243      | 18 unit tests (table-driven)                                  |          |
| `debouncer_test.go`  | 143      | 8 unit tests (concurrent-safe)                                |          |
| `middleware_test.go` | 217      | 10 unit tests                                                 |          |
| **Test total**       | **1160** |                                                               |          |
| **Grand total**      | **2202** |                                                               |          |

---

## A) FULLY DONE

### Core Library — All Working Correctly

- **Functional options pattern** — 9 options: `WithDebounce`, `WithPerPathDebounce`, `WithFilter`, `WithExtensions`, `WithIgnoreDirs`, `WithIgnoreHidden`, `WithRecursive`, `WithMiddleware`, `WithErrorHandler`. All idiomatic Go, all self-documenting.

- **Composable filters** — `FilterExtensions`, `FilterIgnoreExtensions`, `FilterIgnoreDirs`, `FilterIgnoreHidden`, `FilterOperations`, `FilterNotOperations`, `FilterGlob`, `FilterAnd`, `FilterOr`, `FilterNot`. The AND/OR/NOT combinators are a genuinely good API design — rare to see this level of composability in Go libraries.

- **Middleware chain** — 7 middleware: `MiddlewareLogging`, `MiddlewareRecovery`, `MiddlewareRateLimit`, `MiddlewareFilter`, `MiddlewareOnError`, `MiddlewareMetrics`, `MiddlewareWriteFileLog`. Follows `go-cqrs-lite` convention (reverse-order wrapping).

- **Dual debounce modes** — `Debouncer` (per-key, file-path independent) and `GlobalDebouncer` (coalesce all events into one timer). Both thread-safe with `sync.Mutex`.

- **Channel-based streaming** — `Watch(ctx) → (<-chan Event, error)` is the correct Go pattern. Context cancellation closes the channel cleanly. 64-buffer prevents goroutine blocking.

- **Automatic recursive watching** — `walkDirFunc` walks the tree on `Watch()`. `handleNewDirectory` adds newly created subdirectories dynamically during watching.

- **Sentinel errors** — `ErrWatcherClosed`, `ErrNoPaths`, `ErrPathNotFound`, `ErrPathNotDir` with `cockroachdb/errors` for proper wrapping and stack traces.

- **Idempotent `Close()`** — Safe to call multiple times. Guarded by `closed bool` and `sync.RWMutex`.

- **Godoc** — `doc.go` has package-level docs with runnable examples for filters, middleware, and debounce modes.

- **Project infrastructure** — `README.md`, `CHANGELOG.md`, `LICENSE`, `.gitignore`, `.golangci.yml`, `AUTHORS` all present.

### Tests — 46 Tests, Race-Clean

- Integration tests use real filesystem (`t.TempDir()`), not mocks
- Table-driven filter tests with helper function
- Concurrent debouncer tests using `atomic.Int32`
- Middleware chain ordering tests
- Context cancellation tests

### Build & Quality Tooling

- `.golangci.yml` with 55+ linters including `gochecknoglobals`, `exhaustruct`, `gocritic`, `gosec`
- `go vet` clean
- `go build` clean
- Race detector clean

---

## B) PARTIALLY DONE

### Test Coverage — 84.8% (Target: 90%+)

| Function                 | Coverage | Why Missing                                                                 |
| ------------------------ | -------- | --------------------------------------------------------------------------- |
| `MiddlewareLogging`      | ~0%      | Test exists but only tests the middleware mechanics, not the logging output |
| `MiddlewareWriteFileLog` | 0%       | File-based logging middleware not tested at all                             |
| `handleError`            | 0%       | Default stderr error path never exercised                                   |
| `watchLoop`              | ~60%     | Error channel branch not covered                                            |
| `addPath`                | ~68%     | Some walk error paths not covered                                           |
| `NewGlobalDebouncer`     | ~66%     | Default delay fallback branch not covered                                   |
| `shouldSkipDir`          | ~70%     | DefaultIgnoreDirs branch partially covered                                  |

### Documentation

- `README.md` exists but could benefit from more advanced examples (custom middleware, complex filter chains)
- No `Example*` test functions for godoc (the `doc.go` has inline examples but they're not runnable)
- No `examples/` directory with standalone programs

### API Surface

- No `Remove(path)` method (can add dirs but can't remove them)
- No `WatchList() []string` method (can't inspect what's being watched)
- No `Stats()` method (no observability into event counts/uptime)

---

## C) NOT STARTED

1. **`FilterRegex(pattern string)`** — Regex-based path filtering. Common need.
2. **`WithBuffer(size int)` option** — Configurable channel buffer size (currently hardcoded to 64).
3. **`Remove(path string)` method** — Ability to stop watching a specific directory.
4. **`WatchList() []string`** — Inspect which paths are currently being watched.
5. **`Stats()` method** — Event counts, uptime, last event timestamp.
6. **Benchmarks** — No benchmark tests for debounce, filter, or middleware performance.
7. **Stress tests** — No tests with 10k+ files or rapid event bursts.
8. **`Example*` test functions** — No runnable godoc examples.
9. **`examples/` directory** — No standalone example programs.
10. **CI/CD** — No GitHub Actions, no Makefile, no justfile.
11. **Named debounce interface** — `debounceInterface` is `interface{}`, not a named Go interface.
12. **`io.Closer` compliance** — `Watcher` has `Close()` but doesn't formally implement `io.Closer`.
13. **`FilterMinSize(size int64)`** — Ignore files below a size threshold.
14. **`FilterCustom(fn func(Event) bool)`** — Escape hatch alias for complex logic.
15. **`WithOnAdd(fn func(path string))`** — Callback when a directory is added to the watcher.
16. **Go module version tag** — No version tag, no stability guarantees.

---

## D) TOTALLY FUCKED UP

### 🔴 Critical Bug #1: `MiddlewareRateLimit` — Data Race

**File:** `middleware.go:56-63`

```go
var lastEvent time.Time          // shared mutable state, no synchronization
if now.Sub(lastEvent) < minInterval {  // non-atomic read
    return nil
}
lastEvent = now                  // non-atomic write
```

The closure captures `lastEvent` by reference. If the handler is invoked concurrently (debounce timers fire from different goroutines), this is a **data race**. The race detector may not catch it because the middleware is currently called serially in `watchLoop`, but it's incorrect by construction — any user composing this middleware into a concurrent pipeline will hit it.

**Fix:** Use `sync.Mutex` or `atomic.Int64` (store `time.Now().UnixNano()`).

### 🔴 Critical Bug #2: `Debouncer.Flush()` Lies About Its Behavior

**File:** `debouncer.go:49-57`

```go
// Flush executes all pending functions immediately and clears all timers.
func (d *Debouncer) Flush() {
    // ...
    for key, timer := range d.timers {
        timer.Stop()          // CANCELS — never calls fn()
        delete(d.timers, key)
    }
}
```

The doc comment says "executes all pending functions immediately." It does **not**. It cancels them — functionally identical to `Stop()`. Any caller relying on `Flush()` to guarantee execution of pending work will lose events silently.

**Fix:** Either (a) store the `fn` closures alongside timers and call them before deleting, or (b) rename to `CancelAll()` and fix the doc comment.

### 🔴 Critical Bug #3: No Guard Against Multiple `Watch()` Calls

**File:** `watcher.go:130-150`

```go
func (w *Watcher) Watch(ctx context.Context) (<-chan Event, error) {
    w.mu.Lock()
    defer w.mu.Unlock()
    if w.closed { ... }    // Only checks closed, not "already watching"
    // ...
    go w.watchLoop(ctx, eventCh)  // Spawns a NEW goroutine each call
    return eventCh, nil
}
```

Calling `Watch()` twice creates **two `watchLoop` goroutines** reading from the same `fswatcher.Events` channel. Events are randomly split between the two goroutines. The first returned channel may never receive some events. The second goroutine's channel may leak.

**Fix:** Add `watching bool` field, set it in `Watch()`, check it, clear it in `Close()`.

### 🔴 Critical Bug #4: `Add()` Uses `RLock` but Mutates `fswatcher`

**File:** `watcher.go:153-167`

```go
func (w *Watcher) Add(path string) error {
    w.mu.RLock()          // READ lock
    defer w.mu.RUnlock()
    // ...
    return w.addPath(abs)  // Calls fswatcher.Add() — MUTATION
}
```

`addPath()` calls `w.fswatcher.Add()` which mutates the underlying fsnotify watcher. Concurrent `Close()` acquires `w.mu.Lock()` (write lock) and calls `w.fswatcher.Close()`. Since multiple `RLock`s can be held simultaneously, `Close()` will block until `Add()` completes — but the intent is wrong: `Add()` should hold a write lock since it mutates state.

**Fix:** Change `Add()` to use `w.mu.Lock()` instead of `w.mu.RLock()`.

### 🟡 Medium Bug #5: `convertEvent` Loses Combined Operations

**File:** `watcher.go:412-433`

`fsnotify` reports ops as bitmasks (e.g., `Create|Write`). The `switch` picks only the first matching case. A `Create|Write` becomes just `Create`, silently discarding `Write`. fsnotify explicitly documents that `Create` may be followed by `Write` events, and combined ops are possible.

**Fix:** Either emit multiple `Event`s for combined ops, or document the priority order (Create > Write > Remove > Rename).

### 🟡 Medium Bug #6: Middleware Errors Silently Discarded

**File:** `watcher.go:342`

```go
_ = wrapped(ctx, e)  // Error from middleware chain is dropped
```

Errors from middleware (including `MiddlewareRecovery` returning panic-recovery errors) are silently discarded. This means `MiddlewareOnError` can never detect upstream middleware failures.

**Fix:** Propagate errors through the handler chain, dispatch to `errorHandler` if non-nil.

### 🟡 Medium Bug #7: Events Silently Dropped on Full Channel

**File:** `watcher.go:311`

```go
default:
    // Silent drop when channel buffer (64) is full
```

Under heavy filesystem activity, events vanish with zero indication — no logging, no metric, no error. This violates the principle of least surprise.

**Fix:** At minimum, log a warning. Better: add a `dropped` counter exposed via `Stats()`.

### 🟡 Design Issue #8: `shouldSkipDir` Hardcodes Dot-Dir Skipping

**File:** `watcher.go:243-248`

```go
if strings.HasPrefix(name, ".") && name != "." {
    return true  // Always skips .config, .local, .vscode, etc.
}
```

Dot-directories are **always skipped during tree walking**, even if the user wants to watch them. `FilterIgnoreHidden()` is a separate, configurable event filter — but by that point, dot-dirs were already excluded from watching. Users who need to watch `.config/` or `.vscode/` cannot.

**Fix:** Add `WithSkipDotDirs(bool)` option (default: true for backward compat).

### 🟡 Design Issue #9: `debounceInterface` is `interface{}`

**File:** `watcher.go:61-64`

Using an untyped `interface{}` in a Go library that otherwise embraces strong types is inconsistent. The `getDebounceKey` method (`watcher.go:361-366`) uses a fragile type assertion against `*Debouncer` — any custom debouncer implementation would silently break per-path debouncing.

**Fix:** Extract to a named interface: `type Debouncer interface { Debounce(key string, fn func()); Stop() }`.

---

## E) WHAT WE SHOULD IMPROVE

### P0 — Before Any Use (Critical Fixes)

| #   | Task                                                | Effort | File(s)         |
| --- | --------------------------------------------------- | ------ | --------------- |
| 1   | Fix `MiddlewareRateLimit` data race                 | 5min   | `middleware.go` |
| 2   | Fix `Debouncer.Flush()` to actually execute pending | 10min  | `debouncer.go`  |
| 3   | Add `watching` guard to prevent double `Watch()`    | 5min   | `watcher.go`    |
| 4   | Change `Add()` from `RLock` to `Lock`               | 2min   | `watcher.go`    |
| 5   | Propagate middleware errors instead of discarding   | 10min  | `watcher.go`    |

### P1 — Before v0.1.0 (Important)

| #   | Task                                              | Effort | File(s)                    |
| --- | ------------------------------------------------- | ------ | -------------------------- |
| 6   | Extract `debounceInterface` to named Go interface | 10min  | `watcher.go`               |
| 7   | Add `WithBuffer(size int)` option                 | 5min   | `options.go`               |
| 8   | Add `WithSkipDotDirs(bool)` option                | 5min   | `options.go`, `watcher.go` |
| 9   | Add `Remove(path)` method                         | 15min  | `watcher.go`               |
| 10  | Add `WatchList() []string` method                 | 10min  | `watcher.go`               |
| 11  | Add `FilterRegex(pattern)` filter                 | 10min  | `filter.go`                |
| 12  | Log or count dropped events on full channel       | 10min  | `watcher.go`               |
| 13  | Raise test coverage to 90%+                       | 30min  | `*_test.go`                |
| 14  | Add benchmark tests for Debouncer                 | 20min  | `debouncer_test.go`        |
| 15  | Add `Example*` test functions for godoc           | 20min  | `*_test.go`                |
| 16  | Document combined-op priority in `convertEvent`   | 2min   | `watcher.go`               |
| 17  | Formalize `io.Closer` interface compliance        | 2min   | `watcher.go`               |

### P2 — Before v1.0 (Nice to Have)

| #   | Task                                                | Effort |
| --- | --------------------------------------------------- | ------ |
| 18  | `Stats()` method (event counts, uptime, last event) | 20min  |
| 19  | Stress test with 10k+ files                         | 30min  |
| 20  | `examples/` directory with standalone programs      | 30min  |
| 21  | Set up GitHub Actions CI                            | 20min  |
| 22  | Create justfile or Makefile                         | 10min  |
| 23  | Integrate in a real project to validate API         | 1hr    |
| 24  | `FilterMinSize(size int64)` filter                  | 10min  |
| 25  | Tag v0.1.0 after all P0/P1 items done               | 2min   |

---

## F) Top 25 Things to Do Next

| #   | Task                                                                  | Priority | Effort | Status      |
| --- | --------------------------------------------------------------------- | -------- | ------ | ----------- |
| 1   | Fix `MiddlewareRateLimit` data race (use `sync.Mutex` or atomic)      | P0       | 5min   | Not started |
| 2   | Fix `Debouncer.Flush()` — execute pending fns, not cancel them        | P0       | 10min  | Not started |
| 3   | Add `watching bool` guard against double `Watch()` calls              | P0       | 5min   | Not started |
| 4   | Change `Add()` from `RLock` to `Lock` (mutation needs write lock)     | P0       | 2min   | Not started |
| 5   | Propagate middleware errors, don't discard with `_ =`                 | P0       | 10min  | Not started |
| 6   | Extract `debounceInterface` to named interface                        | P1       | 10min  | Not started |
| 7   | Add `WithBuffer(size int)` option (configurable channel buffer)       | P1       | 5min   | Not started |
| 8   | Add `WithSkipDotDirs(bool)` option                                    | P1       | 5min   | Not started |
| 9   | Add `Remove(path)` method to stop watching a directory                | P1       | 15min  | Not started |
| 10  | Add `WatchList() []string` method                                     | P1       | 10min  | Not started |
| 11  | Add `FilterRegex(pattern string)` filter                              | P1       | 10min  | Not started |
| 12  | Log/count dropped events when channel buffer is full                  | P1       | 10min  | Not started |
| 13  | Raise coverage to 90%+ (middleware, error paths)                      | P1       | 30min  | Not started |
| 14  | Add benchmark tests for Debouncer                                     | P1       | 20min  | Not started |
| 15  | Add `Example*` test functions for godoc                               | P1       | 20min  | Not started |
| 16  | Document combined-op priority in `convertEvent`                       | P1       | 2min   | Not started |
| 17  | Formalize `io.Closer` compliance: `var _ io.Closer = (*Watcher)(nil)` | P1       | 2min   | Not started |
| 18  | Add `Stats()` method (event counts, uptime, last event)               | P2       | 20min  | Not started |
| 19  | Stress test with 10k+ files                                           | P2       | 30min  | Not started |
| 20  | Add `examples/` directory with standalone programs                    | P2       | 30min  | Not started |
| 21  | Set up GitHub Actions CI                                              | P2       | 20min  | Not started |
| 22  | Create justfile or Makefile                                           | P2       | 10min  | Not started |
| 23  | Integrate in a real project (hierarchical-errors, Kernovia, etc.)     | P2       | 1hr    | Not started |
| 24  | Add `FilterMinSize(size int64)` filter                                | P3       | 10min  | Not started |
| 25  | Tag v0.1.0 after all P0/P1 items complete                             | P3       | 2min   | Not started |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `convertEvent` emit one event per fsnotify op-bit, or pick the highest-priority op?**

fsnotify reports events as bitmasks (e.g., `Create|Write`). The current code picks the first match via `switch` order (Create > Write > Remove > Rename). This means a `Create|Write` silently loses the `Write`.

Three options:

1. **Keep current behavior** — Pick highest-priority op. Simple, but loses information.
2. **Emit multiple events** — One `Event` per set bit. More accurate, but changes the API contract (one fsnotify event → multiple channel events). Could surprise users.
3. **Add `Event.Ops` field** — Change `Op` to `Ops []Op` (or use a bitmask like fsnotify). Breaking API change but most accurate.

This is a semantic API decision that affects every consumer. It can't be made without understanding the target use cases.

---

## Dependencies

| Dependency                      | Version | Type   | Why                                 |
| ------------------------------- | ------- | ------ | ----------------------------------- |
| `github.com/fsnotify/fsnotify`  | v1.9.0  | Direct | Cross-platform file system watching |
| `github.com/cockroachdb/errors` | v1.12.0 | Direct | Error wrapping with stack traces    |

---

## Bug Severity Summary

| Severity    | Count | Items                                                                  |
| ----------- | ----- | ---------------------------------------------------------------------- |
| 🔴 Critical | 4     | Rate-limit race, Flush() lying, double-Watch, RLock mutation           |
| 🟡 Medium   | 5     | Combined ops, error swallowing, silent drops, dot-dirs, type assertion |
| Total       | 9     |                                                                        |

---

_Generated: 2026-04-04 07:00 CEST by Crush (Parakletos AI)_
