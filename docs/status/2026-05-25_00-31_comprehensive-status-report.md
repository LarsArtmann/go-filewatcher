# Comprehensive Status Report — go-filewatcher

**Date:** 2026-05-25 00:31 CEST
**Coverage:** 87.7% (main package), 76.4% (total including examples)
**Build:** Clean (`go build`, `go vet` pass)
**Tests:** 201/201 pass with `-race`, 0 failures
**Linter:** 0 errors, 27 warnings (all pre-existing)
**Branch:** master (4 commits ahead of origin)
**Lines of code:** 10,474 total across 29 Go files

---

## A) FULLY DONE

### Sprint 1 — Broken Stub Fixes (commit `0bf2d89`)

| # | Task               | What Was Fixed                                                                                                                                                                                 | Evidence                                   |
| - | ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| 1 | Wire `WithDebug`   | Was a no-op stub: set fields but no code read them. Added `debugLog()` helper; wired into `watchLoop`, `processEvent`, `emitEvent`, `handleError`, `handleNewDirectory`, `pollLoop`, `Watch()` | `watcher_internal.go:21`, `watcher.go:256` |
| 2 | Wire `WithPolling` | Was a no-op stub: no polling goroutine existed. Created `watcher_poll.go` with snapshot-based change detection                                                                                 | `watcher_poll.go` (171 lines)              |

### Sprint 1 — New Middleware (7 functions)

| # | Function                      | Lines | What It Does                                        |
| - | ----------------------------- | ----- | --------------------------------------------------- |
| 1 | `MiddlewareCircuitBreaker`    | 75    | Closed→Open→HalfOpen state machine                  |
| 2 | `MiddlewareErrorRateLimit`    | 55    | Suppresses errors after threshold                   |
| 3 | `MiddlewareErrorRecovery`     | 12    | Strategy-based error transformation                 |
| 4 | `MiddlewareErrorBatch`        | 68    | Collects `BatchError` slices, flushes on timer/size |
| 5 | `MiddlewareErrorCorrelation`  | 17    | Injects correlation IDs                             |
| 6 | `MiddlewareErrorSanitization` | 18    | Strips sensitive paths from errors                  |
| 7 | (plus existing 10)            | —     | Total: 17 middleware functions                      |

### Sprint 1 — New Features

| # | Feature                        | What                                                                 |
| - | ------------------------------ | -------------------------------------------------------------------- |
| 1 | `AddRecursive(path, maxDepth)` | Depth-limited directory walking (0=immediate, -1=full, N>0=N levels) |
| 2 | `WithFollowSymlinks(bool)`     | Resolves symlinks via `filepath.EvalSymlinks` during walking         |
| 3 | Polling goroutine              | Snapshot-based Create/Write/Remove detection                         |
| 4 | Debug logging                  | `debugLog()` helper wired throughout pipeline                        |
| 5 | Fuzz testing                   | 5 targets in `fuzz_test.go`                                          |
| 6 | Goreleaser                     | `.goreleaser.yml` with multi-platform builds                         |

### Sprint 1 — Tests Added

- 19 new test functions (7 watcher + 12 middleware)
- 5 fuzz targets
- Total: **201 tests passing**

### Sprint 1 — Housekeeping

- Removed deprecated `git-town.toml`
- Updated `AGENTS.md` with new files, gotchas, patterns

### Sprint 2 — This Session (commit `59b19ab`)

| # | Task                            | What                                                                                                                                                                                                                                       |
| - | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | Remove `categoryStr*` constants | Eliminated redundant `categoryStrTransient`, `categoryStrPermanent`, `categoryStrUnknown` — inlined into `String()` methods. Fixed cross-file coupling where `CircuitState.String()` in `middleware.go` depended on `errors.go` constants. |

---

## B) PARTIALLY DONE

| # | Task                           | What's Done                                                                 | What's Missing                                            |
| - | ------------------------------ | --------------------------------------------------------------------------- | --------------------------------------------------------- |
~~| 1 | Coverage ≥90%                  | Main package at 87.7%. Most functions at 100%.                              | **33 functions below 100%** — see Section E for breakdown |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 2 | Error simulation testing       | Indirect tests via `handleError` calls                                      | No fault injection framework                              |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 3 | Pre-commit hook compliance     | Production code clean                                                       | 9 test files have unused `testpackage` nolint directives  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 4 | Code duplication in middleware | Identified: `MiddlewareBatch` / `MiddlewareErrorBatch` share batching logic | Not yet consolidated                                      |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

---

## C) NOT STARTED

### From TODO_LIST.md — Feature Work

| #  | Task                                             | Effort |
| -- | ------------------------------------------------ | ------ |
~~| 42 | Exponential backoff middleware                   | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 45 | Filter func return match metadata                | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 48 | `WatchChanges(ctx, targetState)` idempotent sync | 25min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 49 | Prometheus metrics export                        | 30min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 60 | Dead letter queue                                | 30min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 61 | Self-healing watcher                             | 45min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 62 | OpenTelemetry integration                        | 45min  |~~ OPEN → research contract + TODO_LIST open Q3 (WatchChanges)
~~| 63 | Error analytics                                  | 30min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 66 | Standalone CLI tool                              | 60min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 67 | Localizable error messages                       | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 68 | Explore fsnotify v2 API changes                  | 30min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 69 | DebounceEntry Mixin phantom type                 | 15min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

### From TODO_LIST.md — Infra/Quality

| #  | Task                                      | Effort |
| -- | ----------------------------------------- | ------ |
~~| 65 | Configure semantic-release                | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 71 | Extract `drainEvents` to testutil package | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 72 | Windows edge case tests                   | 30min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 74 | Test examples/ in CI                      | 15min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 78 | Migrate CI to Nix                         | 60min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 79 | Add Cachix for binary caching             | 20min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

### From TODO_LIST.md — Integration

| #  | Task                                  | Effort |
| -- | ------------------------------------- | ------ |
~~| 76 | Integrate into file-and-image-renamer | 60min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 77 | Integrate into dynamic-markdown-site  | 60min  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

---

## D) TOTALLY FUCKED UP

### Previously Broken — NOW FIXED:

1. ~~**WithPolling** — Option accepted but did nothing~~ → **FIXED** in `0bf2d89`
2. ~~**WithDebug** — Option accepted but no debug logging~~ → **FIXED** in `0bf2d89`

### Currently Fucked Up:

~~1. **`pollEmitEvent` has 0.0% coverage** — The polling code path that constructs and emits events from the poll goroutine is never directly exercised by tests. Integration tests hit `pollDetectChanges` but the final emission step is missed by the coverage tool because it runs asynchronously. This is the single biggest coverage gap.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~2. **`walkDirFunc` at 52.2% coverage** — The symlink resolution branch (`followSymlinks` path) and error branches are not tested. Only the happy path with `shouldSkipDir` is covered.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~3. **`executeHandler` at 60.0% coverage** — Error branches (middleware error, debounced execution) are not covered.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~4. **`WithWatchedIgnoreDirs` at 0.0%** — Completely untested option. Functionally identical to `WithFilter(FilterIgnoreDirs(...))`.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~5. **`MiddlewareErrorSanitization` at 66.7%** — nil `sanitize` parameter path untested.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~6. **`MiddlewareErrorBatch` at 69.7%** — Timer-based flush path untested (only max-size flush tested).~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~7. **`CircuitState.String()` at 0.0%** — Stringer method never called in tests.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~8. **`ErrorCategory.String()` at 0.0%** — Just refactored but never tested.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~9. **`event.go` `omitempty` tag** — `Event.ModTime` has `omitempty` which has no effect on `time.Time` struct fields. Should be `omitzero` (Go 1.24+) or removed.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~10. **`WatchOnce` double `%w`** — `fmt.Errorf("watchonce cancelled: %w: %w", closeErr, ctx.Err())` wraps two errors but `errors.Unwrap()` can only return one. Second error is silently dropped.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

---

## E) WHAT WE SHOULD IMPROVE

### Critical (blocks release)

~~1. **Raise coverage from 87.7% → ≥90%** — 33 functions below 100%, 8 at 0%. The biggest gaps:~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
   - `pollEmitEvent` (0.0%), `CircuitState.String()` (0.0%), `ErrorCategory.String()` (0.0%), `WithWatchedIgnoreDirs` (0.0%)
   - `walkDirFunc` (52.2%), `executeHandler` (60.0%), `AddRecursive` (61.9%)
   - `pollWalkDir` (70.0%), `WatchOnce` (71.4%), `MiddlewareErrorSanitization` (66.7%)

~~2. **Fix `WatchOnce` double `%w`** — Second error silently dropped. Use `fmt.Errorf("... %w ... %w", err1, err2)` with Go 1.20+ multi-Error wrapping or use `%v` for the second.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~3. **Fix `event.go` `omitempty` on `time.Time`** — No effect, misleading. Change to `omitzero` or remove.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

### High Impact (architecture quality)

~~4. **Extract `copyWatchList()` helper** — Watch-list copy pattern duplicated 3×: `pollSnapshot`, `pollDetectChanges`, `WatchList`. Single helper eliminates duplication and lock bugs.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~5. **Consolidate `MiddlewareBatch` / `MiddlewareErrorBatch`** — Share a generic batcher. ~74+68 lines of near-identical timer-based batching logic.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~6. **`filterFileStat` return signature** — Returns `(os.FileInfo, bool, bool)` where the two bools are `isFile` and `shouldFilter`. Named results or a result struct would prevent mixups.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~7. **`MiddlewareErrorSanitization` breaks error chains** — `fmt.Errorf("%s", sanitize(err.Error()))` loses the original error. `errors.Is`/`errors.As` break on sanitized errors.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

### Medium Impact (code quality)

~~8. **Remove `WithWatchedIgnoreDirs`** — Functionally identical to `WithFilter(FilterIgnoreDirs(...))`. Adds API surface with zero value.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~9. **`WithPolling` clobbers `WithPollInterval`** — If `WithPolling` is called after `WithPollInterval`, it overwrites to `defaultPollInterval` when `pollInterval == 0`. Option ordering dependency is a hidden bug.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~10. **`NewWatcherError` always captures stack trace** — `debug.Stack()` is expensive. Should be opt-in or conditional.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~11. **Consolidate error code mapping** — Sentinel→ErrorCode→ErrorCategory maintained in 3 separate locations (`var` block, `ErrorCode` const, `errorToCode` switch). A single registration table would eliminate all three.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~12. **Extract `filepath.Abs` helper** — Pattern repeated in `Add`, `AddRecursive`, `Remove`, `New`.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~13. **`Event.LogValue` omits `Size` and `ModTime`** — Incomplete structured logging.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~14. **Remove duplicate filter test runners** — `runFilterTests`, `runFilterTestsInline`, `runFilterTestsTable` are three variants doing the same thing.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

### Housekeeping

~~15. **Consolidate `docs/status/`** — 30+ files, many stale from April.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~16. **Remove `result` binary from git** — Committed binary flagged by `go-structure-linter`.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

~~17. **File size limits exceeded** — `middleware.go` (688), `watcher.go` (599), `watcher_test.go` (1466), `middleware_test.go` (808) all exceed the 350-line soft limit.~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by **Pareto: highest impact × lowest effort first**.

| Priority | #  | Task                                                                                                    | Effort | Impact | Rationale                                  |
| -------- | -- | ------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------ |
~~| 1        | —  | **Add tests for 0% functions** (`ErrorCategory.String`, `CircuitState.String`, `WithWatchedIgnoreDirs`) | 10min  | HIGH   | Free coverage: these are trivial functions |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 2        | 9  | **Fix `WatchOnce` double `%w`** — use `%v` for second error                                             | 5min   | HIGH   | Silent error dropping bug                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 3        | —  | **Fix `Event.ModTime` `omitempty`** → `omitzero`                                                        | 2min   | MEDIUM | Misleading tag                             |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 4        | —  | **Extract `copyWatchList()` helper** — eliminate 3× duplication                                         | 10min  | MEDIUM | DRY + lock safety                          |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 5        | —  | **Add `walkDirFunc` symlink branch tests** — raise from 52.2%                                           | 15min  | HIGH   | Largest single-function gap                |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 6        | —  | **Add `executeHandler` error branch tests** — raise from 60.0%                                          | 10min  | HIGH   | Error path coverage                        |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 7        | —  | **Add `AddRecursive` depth edge case tests** — raise from 61.9%                                         | 15min  | HIGH   | New feature needs coverage                 |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 8        | —  | **Add `pollEmitEvent` test** — raise from 0.0%                                                          | 15min  | HIGH   | Biggest polling coverage gap               |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 9        | —  | **Add `MiddlewareErrorSanitization` nil path test**                                                     | 5min   | MEDIUM | Missing nil handling test                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 10       | —  | **Add `MiddlewareErrorBatch` timer flush test**                                                         | 10min  | MEDIUM | Only max-size path tested                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 11       | 42 | **Implement exponential backoff middleware**                                                            | 20min  | HIGH   | Natural pairing with circuit breaker       |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 12       | —  | **Consolidate `MiddlewareBatch` / `MiddlewareErrorBatch`** generic batcher                              | 25min  | MEDIUM | ~142 lines of duplication                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 13       | —  | **Fix `Event.LogValue` to include `Size` and `ModTime`**                                                | 5min   | LOW    | Incomplete structured logging              |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 14       | —  | **Fix `WithPolling` option ordering bug** — don't clobber `pollInterval` if already set                 | 10min  | HIGH   | Silent configuration loss                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 15       | 65 | **Configure semantic-release**                                                                          | 20min  | MEDIUM | Goreleaser alone doesn't handle versioning |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 16       | 74 | **Test `examples/` in CI** — `go build ./examples/...`                                                  | 15min  | MEDIUM | Examples should compile in CI              |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 17       | —  | **Remove `WithWatchedIgnoreDirs`** — redundant with `WithFilter(FilterIgnoreDirs(...))`                 | 10min  | LOW    | Dead API surface                           |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 18       | —  | **Make `NewWatcherError` stack capture opt-in**                                                         | 10min  | MEDIUM | Expensive default                          |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 19       | —  | **Fix `MiddlewareErrorSanitization` error chain** — preserve `errors.Is`/`errors.As`                    | 15min  | HIGH   | Silent chain breakage                      |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 20       | 71 | **Extract `drainEvents` to testutil**                                                                   | 20min  | LOW    | Test consolidation                         |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 21       | 60 | **Dead letter queue middleware**                                                                        | 30min  | MEDIUM | Pairs with circuit breaker                 |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 22       | 45 | **Filter func return match metadata**                                                                   | 20min  | MEDIUM | Richer filter semantics                    |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 23       | 72 | **Windows-specific edge case tests**                                                                    | 30min  | MEDIUM | Cross-platform goal                        |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 24       | —  | **Consolidate `docs/status/`** — archive stale files                                                    | 15min  | LOW    | 30+ files, most stale                      |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~| 25       | 49 | **Prometheus metrics export**                                                                           | 30min  | MEDIUM | Observability integration                  |~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

---

## G) TOP QUESTION I CANNOT FIGURE OUT MYSELF

**Should `WithWatchedIgnoreDirs` be removed or kept?**

It's functionally identical to `WithFilter(FilterIgnoreDirs(...))` and has 0.0% test coverage. It adds API surface with no unique behavior. However, removing it is a breaking change for any user who imported it.

Options:

~~1. **Remove it** — breaking change, but the function has zero tests and zero unique behavior~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~2. **Keep it, add tests** — maintain API compatibility~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)
~~3. **Deprecate it** — mark with `// Deprecated:` comment, redirect to `WithFilter(FilterIgnoreDirs(...))`~~ resolved — verified shipped ≤v2.4.1 or consciously dropped (see open-item markers)

I recommend option 3 — deprecation preserves compatibility while guiding users to the more composable API.

---

## Metrics

| Metric                   | Before Sprint 1 | After Sprint 1  | After Sprint 2           | Delta (Total) |
| ------------------------ | --------------- | --------------- | ------------------------ | ------------- |
| Broken stubs             | 2               | 0               | 0                        | -2            |
| Middleware functions     | 10              | 17              | 17                       | +7            |
| Test functions           | ~65             | 201             | 201                      | +136          |
| Test coverage (main pkg) | 92.3%           | 87.6%           | 87.7%                    | -4.6%         |
| Test coverage (total)    | —               | —               | 76.4%                    | —             |
| Lint issues (prod code)  | 0               | 0               | 0                        | —             |
| Lint warnings (all)      | —               | —               | 27                       | —             |
| Lines of code            | 8,755           | 10,474          | 10,474                   | +1,719        |
| Files changed            | —               | 16              | 2                        | 18 total      |
| Dead code removed        | —               | `git-town.toml` | `categoryStr*` constants | -12 lines     |

---

_Assisted-by: Crush_
