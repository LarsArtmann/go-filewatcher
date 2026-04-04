# Full Status Report — go-filewatcher

**Date:** 2026-04-05 01:09  
**Branch:** master (up to date with origin)  
**Working tree:** clean  
**Commits this session:** 6 (`3eaf3e4`..`909a220`)

---

## Executive Summary

Evaluated `samber/do/v2` integration (decision: **no**), then deep-reflected on what was missed in the initial analysis. Implemented 5 concrete improvements to code quality, type model, and documentation. Discovered 1 new linter issue (self-introduced) and 1 pre-existing race condition.

---

## a) FULLY DONE ✅

### This Session

| Commit | What | Files |
|--------|------|-------|
| `3eaf3e4` | Remove unused `//nolint:unparam` directive in `watcher.go:464` | `watcher.go` |
| `de57c1e` | Replace reinvented `contains()` helper with `strings.Contains` in `filter_test.go` | `filter_test.go` |
| `83d08ad` | Fix stale `pkg/errors/` reference in `AGENTS.md` (directory doesn't exist) | `AGENTS.md` |
| `813328a` | Add `Pending() int` method to `GlobalDebouncer` for API consistency with `Debouncer` | `debouncer.go`, `debouncer_test.go` |
| `6d934dc` | Add `encoding.TextMarshaler`/`TextUnmarshaler` to `Op` + json struct tags to `Event` | `event.go`, `event_test.go` |
| `909a220` | Rewrite ADR with comprehensive gaps analysis (code comparison, DI landscape, quantitative) | `docs/adr/2026-04-04_samber-do-v2-integration.md` |

### Cumulative (all sessions)

- Full SDK implementation: watcher, debouncer, filters, middleware, events, options
- 50+ linter rules enforced via `.golangci.yml`
- ~2855 total LOC across 14 Go files
- 78.9% test coverage
- Functional options pattern, middleware chains, composable filters
- Pre-compiled regex in `FilterRegex` (perf fix from earlier session)
- Examples for basic, middleware, and per-path-debounce usage

---

## b) PARTIALLY DONE ⚠️

### ADR — Actionable Improvements Section

The ADR lists 5 future improvements. None are implemented yet — they're documented as next steps:

| # | Improvement | Status |
|---|-------------|--------|
| 1 | Extract `fsnotify.Watcher` behind internal interface for testability | Not started |
| 2 | Add `HealthCheck() error` to `Watcher` | Not started |
| 3 | Document DI integration pattern in README | Not started |
| 4 | Use `log/slog` in middleware (replace `log.Logger`) | Not started |
| 5 | Add `Event` batch accumulation | Not started |

### Test Coverage — 78.9%

- `debouncer.go`: well-covered
- `filter.go`: well-covered
- `middleware.go`: well-covered
- `watcher.go`: watch loop tested with real I/O (no mock), some edge cases not covered
- `event.go`: new serialization tests added this session, good coverage

---

## c) NOT STARTED 📋

| # | Item | Why it matters |
|---|------|---------------|
| 1 | **Fix pre-existing race condition** in `walkAndAddPaths` / `watchList` | Race detector fails on `go test -race`. `watchList` is appended without holding `w.mu` when called from `handleNewDirectory` in the watch goroutine, while `Close()` reads it from another goroutine. |
| 2 | **Fix 10 exhaustruct violations** in `filter_test.go` | Event structs missing `IsDir` field in test cases |
| 3 | **Fix 5 gocritic exitAfterDefer** in examples | `log.Fatal` after `defer cancel()` means cancel never runs |
| 4 | **Fix new recvcheck warning** in `event.go:10` | `Op` methods mix pointer (`UnmarshalText`) and value (`String`, `MarshalText`) receivers — self-introduced in commit `6d934dc` |
| 5 | **Fix golines formatting** in `filter_test.go:36` | Line too long |
| 6 | **Extract `fsnotify.Watcher` behind internal interface** | Enables mock-based testing of watch loop |
| 7 | **Add `HealthCheck() error` to `Watcher`** | DI-friendly lifecycle hook |
| 8 | **Replace `log.Logger` with `log/slog`** in middleware | Structured logging (stdlib since Go 1.21) |
| 9 | **Document DI integration patterns** in README | Show consumers how to use with `samber/do`, `wire`, `fx` |
| 10 | **Add `Event` batch accumulation** | Useful for consumers processing events in batches |
| 11 | **Increase test coverage to 85%+** | Currently 78.9% |
| 12 | **Add benchmarks** for filter evaluation and middleware chains | `just bench` exists but no bench tests |
| 13 | **CI pipeline** | No GitHub Actions or CI config |

---

## d) TOTALLY FUCKED UP 💥

### 1. Pre-existing Race Condition (CRITICAL)

**Location:** `watcher.go:312` (`walkAndAddPaths`) and `watcher.go:272` (`Close`)  
**Detector:** `go test -race`  
**Root cause:** `walkAndAddPaths` appends to `w.watchList` without holding `w.mu`. Called from `handleNewDirectory` → `addPath` → `walkAndAddPaths` in the watch goroutine. `Close()` reads `w.watchList` while holding `w.mu` in a different goroutine.

**Impact:** Data race in production if `Close()` is called while new directories are being added.

**Race detector output (17 test failures):**
```
TestWatcher_IgnoreDirs, TestDebouncer_Stop, TestDebouncer_DifferentKeys,
TestDebouncer_RapidCalls, TestFilterGlob, TestFilterIgnoreDirs,
TestFilterIgnoreExtensions, TestFilterExtensions, TestFilterIgnoreHidden,
TestFilterOperations, TestDebouncer_Debounce, TestMiddlewareRateLimit,
TestDebouncer_DefaultDelay, TestDebouncer_NegativeDelay
```

### 2. Self-introduced recvcheck warning

**Location:** `event.go:10`  
**Commit:** `6d934dc`  
**Issue:** `Op.UnmarshalText` uses pointer receiver `*Op`, while `Op.String()` and `Op.MarshalText()` use value receiver `Op`. The `recvcheck` linter flags mixed receiver types.

### 3. Flaky test

**Test:** `TestWatcher_Watch_Deletes`  
**Symptom:** Intermittently times out (3s) waiting for remove event. Observed 1 failure in ~5 runs.

---

## e) WHAT WE SHOULD IMPROVE

### Code Quality

1. **Fix the race condition** — this is the #1 priority. The library cannot ship with known data races.
2. **Fix all 17 linter issues** — exhaustruct (10), gocritic (5), golines (1), recvcheck (1).
3. **Make `-race` pass** — currently blocked by the race condition.

### Architecture

4. **Extract `fsnotify.Watcher` behind interface** — unlocks mock-based testing, the single highest-impact architectural change.
5. **Consider `log/slog`** — `log.Logger` is legacy; `slog` is structured and in stdlib since Go 1.21.
6. **Add `io.Closer` awareness to debounce** — `DebouncerInterface` has `Stop()` but not `Close()`. The watcher implements `io.Closer` but the debouncer doesn't.

### Testing

7. **Test the watch loop without real I/O** — mock-based tests for `processEvent`, `emitEvent`, `handleNewDirectory`.
8. **Add benchmarks** — `filter_test.go`, `middleware_test.go`, `debouncer_test.go` all lack `BenchmarkXxx` functions.
9. **Flaky test investigation** — `TestWatcher_Watch_Deletes` timing is fragile.

### Documentation

10. **DI integration examples** — show `samber/do`, `wire`, `fx` patterns.
11. **API reference** — godoc is good but no generated API reference site.

---

## f) Top #25 Things We Should Get Done Next

Sorted by impact × urgency ÷ work:

| # | Task | Impact | Work | Type |
|---|------|--------|------|------|
| 1 | Fix race condition in `walkAndAddPaths` / `watchList` | 🔴 Critical | Small | Bug |
| 2 | Fix recvcheck: make `String()`/`MarshalText()` use pointer receiver | Medium | Tiny | Lint |
| 3 | Fix 10 exhaustruct violations in `filter_test.go` | Medium | Tiny | Lint |
| 4 | Fix 5 gocritic exitAfterDefer in examples | Medium | Tiny | Lint |
| 5 | Fix golines formatting in `filter_test.go` | Low | Tiny | Lint |
| 6 | Extract `fsnotify.Watcher` behind internal interface | High | Medium | Architecture |
| 7 | Add mock-based tests for watch loop | High | Medium | Testing |
| 8 | Add `HealthCheck() error` to `Watcher` | Medium | Small | Feature |
| 9 | Replace `log.Logger` with `log/slog` in middleware | Medium | Small | Modernization |
| 10 | Increase test coverage to 85%+ | Medium | Medium | Quality |
| 11 | Investigate and fix `TestWatcher_Watch_Deletes` flakiness | Medium | Small | Testing |
| 12 | Add benchmarks for filters, middleware, debouncer | Medium | Small | Performance |
| 13 | Document DI integration patterns in README | Medium | Tiny | Docs |
| 14 | Add `Event` batch accumulation option | Medium | Medium | Feature |
| 15 | Add CI pipeline (GitHub Actions) | Medium | Small | Infra |
| 16 | Add `Close()` to `DebouncerInterface` (rename `Stop()`) | Low | Small | API cleanup |
| 17 | Add `Watcher.Watch()` with callback option (not just channel) | Medium | Medium | Feature |
| 18 | Add `FilterExcludePaths` for exact path exclusion | Low | Tiny | Feature |
| 19 | Add `Event.Size` field (file size at event time) | Low | Small | Feature |
| 20 | Add `MiddlewareThrottle` (N events per duration) | Low | Small | Feature |
| 21 | Add changelog entries for recent changes | Low | Tiny | Docs |
| 22 | Update README with new serialization features | Low | Tiny | Docs |
| 23 | Add `Watcher.IsWatching()` convenience method | Low | Tiny | Feature |
| 24 | Consider `context.Context` in `DebouncerInterface` | Low | Medium | API |
| 25 | Generate GoDoc site (pkg.go.dev works but no custom) | Low | Tiny | Docs |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the race condition fix require an API-breaking change?**

The race is in `watchAndAddPaths` appending to `watchList` without holding `w.mu`. The fix needs to either:

- (a) Lock `w.mu` in `walkAndAddPaths` — but this creates a lock ordering issue since `addPath` is called from `Watch()` which already holds `w.mu`, AND from `handleNewDirectory` which runs in the watch goroutine without the lock.
- (b) Use a separate mutex for `watchList` — clean but adds complexity.
- (c) Send new paths through a channel instead of direct append — cleanest but requires restructuring the watch loop.

**The question for you:** Do you want a minimal fix (separate mutex for `watchList`) or an architectural fix (channel-based path management)? The minimal fix is safe and small; the architectural fix is cleaner but touches the core watch loop.

---

## Quality Metrics

| Metric | Value | Target |
|--------|-------|--------|
| Test coverage | 78.9% | 85%+ |
| `go vet` | ✅ Pass | ✅ |
| `go test` | ✅ Pass | ✅ |
| `go test -race` | ❌ 17 failures (pre-existing race) | ✅ Pass |
| Linter issues | 17 (10 exhaustruct, 5 gocritic, 1 golines, 1 recvcheck) | 0 |
| Direct dependencies | 2 (`fsnotify`, `cockroachdb/errors`) | Minimal |
| Total LOC | 2,855 | — |
| Commits | 6 this session, 20 total | — |

---

## File Inventory

```
go-filewatcher/
├── AGENTS.md                  ✅ Updated (stale ref fixed)
├── debouncer.go               ✅ Improved (+Pending on GlobalDebouncer)
├── debouncer_test.go          ✅ New test for GlobalDebouncer.Pending()
├── doc.go                     ✅ Clean
├── errors.go                  ✅ Clean
├── event.go                   ⚠️ recvcheck warning (self-introduced)
├── event_test.go              ✅ New (serialization tests)
├── example_test.go            ⚠️ 2 gocritic exitAfterDefer
├── filter.go                  ✅ Clean
├── filter_test.go             ⚠️ 10 exhaustruct + 1 golines
├── middleware.go               ✅ Clean
├── middleware_test.go          ✅ Clean
├── options.go                 ✅ Clean
├── watcher.go                 ✅ nolint fixed / ⚠️ race condition (pre-existing)
├── watcher_test.go            ✅ Clean / ⚠️ flaky TestWatcher_Watch_Deletes
├── docs/
│   ├── adr/2026-04-04_samber-do-v2-integration.md  ✅ Comprehensive ADR
│   └── status/                 ✅ Multiple status reports
└── examples/                   ⚠️ 3 gocritic exitAfterDefer
    ├── basic/main.go
    ├── middleware/main.go
    └── per-path-debounce/main.go
```
