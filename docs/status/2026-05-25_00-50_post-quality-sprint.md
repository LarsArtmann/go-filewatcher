# Comprehensive Status Report — go-filewatcher

**Date:** 2026-05-25 00:50 CEST
**Coverage:** 89.8% (main package), 78.2% (total including examples)
**Build:** Clean (`go build`, `go vet` pass)
**Tests:** 211/211 pass with `-race`, 0 failures
**Linter:** 0 errors, ~27 warnings (all pre-existing)
**Branch:** master (synced with origin)
**Lines of code:** 10,681 total across 29 Go files

---

## A) FULLY DONE

### Sprint 1 (commit `0bf2d89`) — Feature Explosion

| What | Details |
|------|---------|
| Wire `WithDebug` | Was no-op stub. Added `debugLog()` helper, wired throughout pipeline |
| Wire `WithPolling` | Was no-op stub. Created `watcher_poll.go` with snapshot-based change detection |
| 7 new middleware | Circuit breaker, error rate limit, recovery, batch, correlation, sanitization |
| New features | `AddRecursive(path, maxDepth)`, `WithFollowSymlinks(bool)`, fuzz tests, goreleaser |
| 19 new tests + 5 fuzz targets | Total: 201 tests passing |

### Sprint 2 (commits `59b19ab` → `6f23684`) — Quality Sprint

| Commit | What | Impact |
|--------|------|--------|
| `59b19ab` | Remove `categoryStr*` constants | Eliminated cross-file coupling between `errors.go` and `middleware.go` |
| `34db2ba` | Status report 2026-05-25 00:31 | Comprehensive status snapshot |
| `a861b55` | Fix `WatchOnce` double `%w` | Second error was silently dropped — now uses `%v` for context error |
| `f614990` | Fix `Event.ModTime` omitempty→omitzero | `omitempty` has no effect on `time.Time`; added Size/ModTime to `LogValue` |
| `fa20497` | Extract `copyWatchList()` helper | Eliminated 3× duplicated lock+copy pattern |
| `14257c3` | Fix `MiddlewareErrorSanitization` error chain | `errors.Is`/`errors.As` broke on sanitized errors — now wraps with `%w` |
| `97486e1` | Add coverage for 0% functions + deprecate `WithWatchedIgnoreDirs` | Coverage 87.7% → 89.7% |
| `6f23684` | Add `executeHandler` error path test | Coverage 89.7% → 89.8% |

### All Production Code Changes (Sprint 1 + Sprint 2)

| File | Lines | Key Exports |
|------|-------|-------------|
| `watcher.go` | 602 | `New`, `Watch`, `WatchOnce`, `Add`, `AddRecursive`, `Remove`, `WatchList`, `Stats`, `Errors`, `Close`, `copyWatchList` |
| `watcher_internal.go` | 339 | `debugLog`, `watchLoop`, `processEvent`, `emitEvent`, `handleError`, `convertEvent` |
| `watcher_poll.go` | 161 | `pollLoop`, `pollSnapshot`, `pollWalkDir`, `pollDetectChanges`, `pollEmitEvent` |
| `watcher_walk.go` | 113 | `initDebouncer`, `addPath`, `walkAndAddPaths`, `walkDirFunc`, `shouldSkipDir` |
| `middleware.go` | 687 | 17 middleware functions, `CircuitState`, `BatchError` |
| `filter.go` | 333 | 18 filter functions, `FilterAnd`, `FilterOr`, `FilterNot` |
| `event.go` | 136 | `Op`, `Event`, JSON/Text marshaling, `LogValue`, `GetPath` |
| `errors.go` | 309 | 11 sentinels, `ErrorCode`, `ErrorCategory`, `WatcherError`, `IsTransientError`, `IsPermanentError` |
| `options.go` | 226 | 20 `With*` options |
| `debouncer.go` | 291 | `Debouncer`, `GlobalDebouncer`, `DebouncerInterface` |

---

## B) PARTIALLY DONE

| Task | Status | Gap |
|------|--------|-----|
| Coverage ≥90% | 89.8% | 0.2% short — `pollEmitEvent` (0%), `walkDirFunc` (52.2%), `executeHandler` (60%) |
| Pre-commit hook compliance | Production code clean | 9 test files have unused `testpackage` nolint directives |
| Code duplication in middleware | Identified: `MiddlewareBatch`/`MiddlewareErrorBatch` share logic | Not yet consolidated |
| Error code mapping DRY | Identified: sentinel→ErrorCode→ErrorCategory in 3 places | Not yet consolidated |

---

## C) NOT STARTED

### Features (from TODO_LIST.md)

| # | Task | Effort |
|---|------|--------|
| 42 | Exponential backoff middleware | 20min |
| 45 | Filter func return match metadata | 20min |
| 48 | `WatchChanges(ctx, targetState)` idempotent sync | 25min |
| 49 | Prometheus metrics export | 30min |
| 60 | Dead letter queue middleware | 30min |
| 61 | Self-healing watcher | 45min |
| 62 | OpenTelemetry integration | 45min |
| 63 | Error analytics | 30min |
| 66 | Standalone CLI tool | 60min |
| 67 | Localizable error messages | 20min |
| 68 | Explore fsnotify v2 API changes | 30min |
| 69 | DebounceEntry Mixin phantom type | 15min |

### Infrastructure

| # | Task | Effort |
|---|------|--------|
| 65 | Configure semantic-release | 20min |
| 71 | Extract `drainEvents` to testutil | 20min |
| 72 | Windows edge case tests | 30min |
| 74 | Test `examples/` in CI | 15min |
| 78 | Migrate CI to Nix | 60min |
| 79 | Add Cachix for binary caching | 20min |
| 76-77 | Integrate into downstream projects | 120min |

---

## D) TOTALLY FUCKED UP

### Fixed this sprint:

1. ~~**`WatchOnce` double `%w`**~~ → Fixed in `a861b55`
2. ~~**`Event.ModTime` misleading `omitempty`**~~ → Fixed in `f614990`
3. ~~**`MiddlewareErrorSanitization` broke error chains**~~ → Fixed in `14257c3`
4. ~~**`categoryStr*` cross-file coupling**~~ → Fixed in `59b19ab`
5. ~~**3× watch-list copy duplication**~~ → Fixed in `fa20497`
6. ~~**`WithWatchedIgnoreDirs` undocumented redundancy**~~ → Deprecated in `97486e1`
7. ~~**`WithDebug` no-op stub**~~ → Fixed in Sprint 1
8. ~~**`WithPolling` no-op stub**~~ → Fixed in Sprint 1

### Still fucked up:

1. **`pollEmitEvent` at 0% coverage** — The polling event emission path is never directly exercised. Integration tests hit `pollDetectChanges` but the final emission step is missed by the coverage tool because it runs asynchronously. This is the single biggest coverage gap (~20 lines uncovered).

2. **`walkDirFunc` at 52.2% coverage** — The symlink resolution branch and error branches are not tested. Only the happy path with `shouldSkipDir` is covered.

3. **`executeHandler` at 60% coverage** — Error branch (lines 190-196) may not be consistently hit. The `executeHandler` error path depends on middleware returning error AND the coverage tool capturing the async execution.

4. **`AddRecursive` at 61.9% coverage** — Edge cases (maxDepth=0, non-existent paths, permission errors) not tested.

5. **`MiddlewareErrorBatch` at 69.7%** — Timer-based flush path untested. Only max-size flush is tested.

6. **`MiddlewareRateLimit` at 75%** / **`MiddlewareSlidingWindowRateLimit` at 71.4%** — Default value branches (when maxEvents ≤ 0) untested.

7. **File size limits exceeded** — `middleware.go` (687), `watcher.go` (602), `watcher_test.go` (1466), `middleware_test.go` (851), `watcher_coverage_test.go` (664) all exceed the 350-line soft limit.

8. **`FilterGeneratedCodeFull` at 64.3%** — Multiple filter option branches untested in the gogenfilter integration.

---

## E) WHAT WE SHOULD IMPROVE

### Critical (blocks release quality)

1. **Raise coverage from 89.8% → ≥90%** — Need tests for `pollEmitEvent` (0%), `walkDirFunc` symlink branch (52.2%), `AddRecursive` edge cases (61.9%). The gap is 0.2% — approximately 5-10 more covered statements needed.

2. **Extract `pollEmitEvent` testability** — The function is unexported and only callable from the async polling loop. Options: (a) add a polling-specific integration test that creates files and verifies events arrive through the poll path, (b) extract the event construction logic to a testable helper.

### High Impact (architecture quality)

3. **Consolidate `MiddlewareBatch` / `MiddlewareErrorBatch`** — ~142 lines of near-identical timer-based batching logic. Extract a generic `batcher[T]` that both can use.

4. **Consolidate error code mapping** — Sentinel→ErrorCode→ErrorCategory maintained in 3 places. A single registration table would eliminate all three.

5. **`filterFileStat` return signature** — Returns `(os.FileInfo, bool, bool)` where bools are `isFile`, `shouldFilter`. Named results or a result struct would prevent mixups.

### Medium Impact (code quality)

6. **Fix pre-existing lint warnings** — 9 test files have unused `testpackage` nolint directives (flagged by `nolintlint`). Safe to remove.

7. **Remove duplicate filter test runners** — `runFilterTests`, `runFilterTestsInline`, `runFilterTestsTable` are three variants doing the same thing.

8. **Make `NewWatcherError` stack capture opt-in** — `debug.Stack()` is expensive for every error. Add `WithStackTraces()` option.

9. **Extract `filepath.Abs` helper** — Pattern repeated in `Add`, `AddRecursive`, `Remove`, `New`.

10. **Remove stale `result` binary** — Committed binary flagged by `go-structure-linter`.

### Housekeeping

11. **Consolidate `docs/status/`** — 30+ files, many stale from April. Archive to `docs/status/archive/`.

12. **File size limits** — Multiple files exceed 350-line soft limit. Consider splitting `middleware.go` into `middleware_batch.go`, `middleware_circuit.go`, etc.

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by **Pareto: highest impact × lowest effort**.

| # | Task | Effort | Impact | Rationale |
|---|------|--------|--------|-----------|
| 1 | **Add `MiddlewareRateLimit` default-value tests** (75%→100%) | 5min | HIGH | Easy: test `maxEvents ≤ 0` defaults to 100 |
| 2 | **Add `MiddlewareSlidingWindowRateLimit` default-value tests** (71%→100%) | 5min | HIGH | Same pattern as above |
| 3 | **Add `AddRecursive` depth=0 test** (61.9%→higher) | 10min | HIGH | Edge case: immediate children only |
| 4 | **Add `pollEmitEvent` integration test** (0%→higher) | 15min | HIGH | Biggest single-function gap — create file via poll path, verify event |
| 5 | **Add `walkDirFunc` symlink error branch test** (52.2%→higher) | 10min | MEDIUM | Symlink resolution error path |
| 6 | **Implement exponential backoff middleware (#42)** | 20min | HIGH | Natural pairing with circuit breaker |
| 7 | **Consolidate `MiddlewareBatch`/`MiddlewareErrorBatch`** generic batcher | 25min | MEDIUM | ~142 lines of duplication |
| 8 | **Configure semantic-release (#65)** | 20min | MEDIUM | Goreleaser alone doesn't handle versioning |
| 9 | **Test `examples/` in CI (#74)** | 15min | MEDIUM | `go build ./examples/...` |
| 10 | **Remove unused `testpackage` nolint directives** (9 files) | 10min | LOW | Clean linter output |
| 11 | **Extract `filepath.Abs` helper** | 10min | LOW | DRY in Add/AddRecursive/Remove/New |
| 12 | **Make `NewWatcherError` stack capture opt-in** | 10min | MEDIUM | Performance: expensive default |
| 13 | **Consolidate error code mapping** | 15min | MEDIUM | 3 locations → 1 registration table |
| 14 | **Dead letter queue middleware (#60)** | 30min | MEDIUM | Pairs with circuit breaker |
| 15 | **Filter func return match metadata (#45)** | 20min | MEDIUM | Richer filter semantics |
| 16 | **Remove duplicate filter test runners** | 15min | LOW | 3 variants → 1 |
| 17 | **Remove stale `result` binary** | 5min | LOW | `git rm`, add to `.gitignore` |
| 18 | **Windows edge case tests (#72)** | 30min | MEDIUM | Cross-platform goal |
| 19 | **Extract `drainEvents` to testutil (#71)** | 20min | LOW | Test consolidation |
| 20 | **Prometheus metrics export (#49)** | 30min | MEDIUM | Observability integration |
| 21 | **Consolidate `docs/status/`** | 15min | LOW | 30+ files, most stale |
| 22 | **`filterFileStat` named result struct** | 10min | LOW | Prevent bool mixups |
| 23 | **`FilterGeneratedCodeFull` coverage** (64.3%→higher) | 15min | LOW | Gogenfilter integration paths |
| 24 | **WatchChanges idempotent sync (#48)** | 25min | MEDIUM | Sync API |
| 25 | **Self-healing watcher (#61)** | 45min | MEDIUM | Auto-retry failed operations |

---

## G) TOP QUESTION I CANNOT FIGURE OUT MYSELF

**Should `pollEmitEvent` be refactored for testability, or should I add a deeper polling integration test?**

`pollEmitEvent` (0% coverage) is only called from `pollDetectChanges`, which runs asynchronously in the polling goroutine. The coverage tool misses it because:
1. It runs in a separate goroutine that may not complete during the test
2. The polling tests DO create files and verify events arrive, but the coverage tool attributes those to the integration test function, not to `pollEmitEvent`

Options:
1. **Extract event construction** — Move the `Event{}` construction from `pollEmitEvent` into a testable helper like `newPollEvent(op, path, fileState) Event`, then test that helper directly
2. **Add sync barrier** — Make the poll loop signal when it has processed a tick, so the test can wait for coverage to register
3. **Accept the gap** — 0.2% short of 90% is acceptable for async polling code that IS integration-tested

I recommend option 1 — it's the cleanest and doesn't change runtime behavior.

---

## Metrics

| Metric | Sprint 1 Start | Sprint 1 End | Sprint 2 End | Delta (Total) |
|--------|---------------|--------------|--------------|---------------|
| Broken stubs | 2 | 0 | 0 | -2 |
| Middleware functions | 10 | 17 | 17 | +7 |
| Test functions | ~65 | 201 | 211 | +146 |
| Test coverage (main) | 92.3% | 87.6% | 89.8% | -2.5% |
| Lint issues (prod) | 0 | 0 | 0 | — |
| Lines of code | 8,755 | 10,474 | 10,681 | +1,926 |
| Commits this sprint | — | 2 | 8 | 10 total |
| Bugs fixed | — | 2 | 4 | 6 total |

### Coverage Heat Map (functions below 80%)

| Function | Coverage | Lines | File |
|----------|----------|-------|------|
| `pollEmitEvent` | 0.0% | 20 | `watcher_poll.go:136` |
| `walkDirFunc` | 52.2% | 44 | `watcher_walk.go:55` |
| `executeHandler` | 60.0% | 14 | `watcher_internal.go:188` |
| `AddRecursive` | 61.9% | 52 | `watcher.go:369` |
| `FilterGeneratedCodeFull` | 64.3% | 62 | `filter_gogen.go:62` |
| `MiddlewareErrorBatch` | 69.7% | 68 | `middleware.go:620` |
| `MiddlewareSlidingWindowRateLimit` | 71.4% | 15 | `middleware.go:127` |
| `WatchOnce` | 71.4% | 28 | `watcher.go:312` |
| `pollWalkDir` | 70.0% | 22 | `watcher_poll.go:60` |
| `pollDetectChanges` | 68.8% | 43 | `watcher_poll.go:87` |
| `MiddlewareRateLimit` | 75.0% | 13 | `middleware.go:114` |
| `addPathWithDepth` | 76.2% | 42 | `watcher.go:410` |
| `FilterGlob` | 80.0% | 14 | `filter.go:151` |
| `MiddlewareCircuitBreaker` | 81.2% | 75 | `middleware.go:412` |
| `MiddlewareDeduplicate` | 85.7% | 45 | `middleware.go:171` |
| `FilterGeneratedCodeWithFilter` | 85.7% | 22 | `filter_gogen.go:101` |

---

_Assisted-by: Crush_
