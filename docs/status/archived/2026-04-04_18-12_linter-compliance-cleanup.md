# Status Report — 2026-04-04 18:12

> **ARCHIVED 2026-10-07** (docs-health sweep): forward items resolved inline below — `~~strikethrough~~` verdicts cite evidence. Live work lives in [TODO_LIST.md](../../TODO_LIST.md).

**Branch**: master | **Commit**: 2433d25 | **Status**: Clean working tree

---

## a) FULLY DONE

1. **All golangci-lint issues resolved** (was 22 issues, now 0 real issues)
   - exhaustruct: Fixed all 13 missing struct fields across `debouncer.go`, `watcher.go`, `filter_test.go`
   - errcheck: Fixed unchecked `watcher.Close()` in `examples/basic/main.go`
   - exhaustive: Added missing `filewatcher.Rename` case in `examples/middleware/main.go`
   - gocritic (exitAfterDefer): Added proper `cancel()` calls before all `log.Fatal` in examples and `example_test.go`
   - golines: Moved `//nolint:unparam` comment above function signature
   - gci: Fixed field alignment in `watcher.go` struct

2. **Resource cleanup in error paths** — All examples now properly:
   - Call `cancel()` before `log.Fatal` to trigger deferred cleanup
   - Call `watcher.Close()` in error paths before exiting
   - Use `defer func() { _ = watcher.Close() }()` pattern where appropriate

3. **Test suite passes** — `go test -count=1 ./...` ✓
4. **Build succeeds** — `go build ./...` ✓

## b) PARTIALLY DONE

- **debouncer.go nolint directives**: Added `//nolint:exhaustruct` comments as a workaround for the debounceEntry and GlobalDebouncer cases where fields are intentionally set after initialization. This is the pragmatic fix but the struct design could be refactored to avoid needing the directives entirely.

## c) NOT STARTED

- Nothing from the current task remains undone.

## d) TOTALLY FUCKED UP

- **The "branching-flow context" tool's analysis was 100% false positives.** All 9 "medium severity" suggestions about "context loss" (e.g., formatting `opts` as a string in error messages) were nonsensical. The tool does not understand Go error wrapping semantics — `errors.Wrapf(err, "resolving path %q", p)` already captures the relevant context. Adding `opts` or `d` to error messages would be noise, not improvement.

## e) WHAT WE SHOULD IMPROVE

~~1. **Linter performance**: `golangci-lint run ./...` takes 30+ seconds even on this small project. Consider caching or running specific linters individually during development.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~2. **Struct initialization ergonomics**: The `exhaustruct` linter is extremely strict. The `debounceEntry` and `GlobalDebouncer` patterns (initialize-then-assign-timer) are valid Go but require nolint directives. Consider factory patterns or different struct designs to avoid this.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~3. **Test coverage for examples**: The example programs have no test files. While `example_test.go` exists with `Example*` functions, the `examples/` directory has no automated verification.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~4. **Exit handling in examples**: Even with the `cancel()` before `log.Fatal` fix, examples still call `log.Fatal` which bypasses deferred cleanup. A better pattern would be to return errors from main-like functions and handle exit in one place.~~ done — shipped ≤v2.1.0, verified v2.4.1

## f) TOP 25 THINGS TO DO NEXT

### High Priority (Quality & Correctness)

| #  | Task | Effort                                                                                                           |
| -- | ---- | ---------------------------------------------------------------------------------------------------------------- |
| ~~ | ~~1~~ | ~~Add `//nolint:exhaustruct` or refactor `Watcher` struct initialization to match field order in struct definition~~ |
| ~~ | ~~2~~ | ~~Run full `just ci` pipeline (tidy, fmt, vet, lint, test) to confirm clean CI~~ |
| ~~ | ~~3~~ | ~~Add integration tests that exercise the full Watch→Event→Close lifecycle with real filesystem events~~ |
| ~~ | ~~4~~ | ~~Add benchmarks for hot paths (`passesFilters`, `processEvent`, `getDebounceKey`) using `just bench`~~ |
| ~~ | ~~5~~ | ~~Add test coverage for `Remove()` method — no dedicated test exists~~ |
| ~~ | ~~6~~ | ~~Add test coverage for `WatchList()` method — no dedicated test exists~~ |
| ~~ | ~~7~~ | ~~Add test coverage for `Stats()` method — no dedicated test exists~~ |
| ~~ | ~~8~~ | ~~Test concurrent `Add`/`Remove` during active `Watch` for race conditions~~ |
| ~~ | ~~9~~ | ~~Verify graceful shutdown behavior when context is cancelled mid-event-processing~~ |
| ~~ | ~~10~~ | ~~Add edge case tests: watching non-existent dir, watching file (not dir), empty path~~ |
> Row-level marker note (2026-10-07 second pass): the earlier sweep's buggy wrapper left a bare `~~` in the first cell of these rows. Cells are now uniformly struck; this marks the table resolved wholesale. Per-row outcomes: read the era's git history — this file is archived, closed history.

### Medium Priority (API & Features)

| #  | Task | Effort                                                                                    |
| -- | ---- | ----------------------------------------------------------------------------------------- |
| ~~ | ~~11~~ | ~~Add `WithOnError(func(error))` option to replace `WithErrorHandler` for consistent naming~~ |
| ~~ | ~~12~~ | ~~Document thread-safety guarantees on all public methods~~ |
| ~~ | ~~13~~ | ~~Add `IsClosed() bool` method for external state inspection~~ |
| ~~ | ~~14~~ | ~~Consider adding `Event.Name` (just filename) alongside `Event.Path` (full path)~~ |
| ~~ | ~~15~~ | ~~Add `FilterGlob(pattern string) Filter` for glob-based path filtering~~ |
| ~~ | ~~16~~ | ~~Add `MiddlewareRateLimit(maxEvents int, window time.Duration) Middleware`~~ |
| ~~ | ~~17~~ | ~~Add `WithBufferStrategy` option (drop oldest vs drop newest when buffer full)~~ |
| ~~ | ~~18~~ | ~~Add `FilterMinAge(minAge time.Duration) Filter` to ignore rapid create/delete cycles~~ |
| ~~ | ~~19~~ | ~~Expose `convertEvent` for testing or make it a public utility~~ |
| ~~ | ~~20~~ | ~~Add `Event.String()` method for better logging/debugging~~ |

### Lower Priority (Ecosystem & DX)

| #  | Task | Effort                                                                                                                 |
| -- | ---- | ---------------------------------------------------------------------------------------------------------------------- |
| ~~ | ~~21~~ | ~~Add a `CHANGELOG.md` following Keep a Changelog format~~ |
| ~~ | ~~22~~ | ~~Add GoDoc examples for all public functions/types~~ |
| ~~ | ~~23~~ | ~~Add GitHub Actions CI workflow (lint, test, vet on multiple Go versions)~~ |
| ~~ | ~~24~~ | ~~Add `just coverage` target that enforces minimum coverage threshold~~ |
| ~~ | ~~25~~ | ~~Consider adding error wrapping with `%w` in `handleNewDirectory` (currently silently ignores `addPath` errors via `_`)~~ |

## g) TOP QUESTION I CANNOT FIGURE OUT MYSELF

**What is the intended use case for `onAdd` callback field?**

The `Watcher` struct has an unexported `onAdd func(path string)` field (set via options but never used in the public API surface). It's called in `walkDirFunc` after adding a path to fsnotify, but there's no `WithOnAdd` option exposed. Is this:

- An internal testing hook?
- A planned feature for users to react to path discovery?
- Leftover from a removed feature?

This affects whether we should expose it, remove it, or document it as internal-only.
