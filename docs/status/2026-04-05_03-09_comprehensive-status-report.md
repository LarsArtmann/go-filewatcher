# go-filewatcher — Comprehensive Status Report

**Date:** 2026-04-05 03:09  
**Branch:** `master` (1 commit ahead of origin)  
**Go:** 1.26.0 (Nix) | **Total LOC:** 3,047 (17 files)  
**Coverage:** 78.9% | **Tests:** 50+ (all pass without `-race`)  
**Linters:** 57 enabled + 4 formatters

---

## Executive Summary

The library is **functionally complete** for v0.1.0 but has **one critical data race** that causes 17 test failures under `go test -race`. There are also 17 linter issues remaining. No CI/CD pipeline exists. No release has been tagged.

**Ship-blocking:** Race condition in `watcher.go:312`.

---

## A) FULLY DONE ✅

| Item | Details |
|------|---------|
| Core watcher | `New()`, `Watch()`, `Close()`, `Add()`, `Remove()`, `WatchList()`, `Stats()` — all working |
| Event system | `Op` type (Create/Write/Remove/Rename), `Event` struct with JSON tags, `MarshalText`/`UnmarshalText` on Op |
| 12 functional options | `WithDebounce`, `WithPerPathDebounce`, `WithFilter`, `WithExtensions`, `WithIgnoreDirs`, `WithIgnoreHidden`, `WithRecursive`, `WithMiddleware`, `WithErrorHandler`, `WithSkipDotDirs`, `WithBuffer`, `WithOnAdd` |
| 11 filter constructors | `FilterExtensions`, `FilterIgnoreExtensions`, `FilterIgnoreDirs`, `FilterIgnoreHidden`, `FilterOperations`, `FilterNotOperations`, `FilterGlob`, `FilterRegex`, `FilterMinSize`, `FilterAnd`, `FilterOr`, `FilterNot` |
| 7 middleware | `MiddlewareLogging`, `MiddlewareRecovery`, `MiddlewareRateLimit`, `MiddlewareFilter`, `MiddlewareOnError`, `MiddlewareMetrics`, `MiddlewareWriteFileLog` |
| 2 debounce strategies | `Debouncer` (per-key) and `GlobalDebouncer` (all-events) — both with `Debounce()`, `Flush()`, `Stop()`, `Pending()` |
| 5 sentinel errors | `ErrWatcherClosed`, `ErrNoPaths`, `ErrPathNotFound`, `ErrPathNotDir`, `ErrWatcherRunning` |
| 4 critical bugs fixed | MiddlewareRateLimit data race, Debouncer.Flush() lying, double Watch() guard, Add() RLock mutation |
| 3 medium bugs fixed | error swallowing in middleware, SkipDotDirs, DebouncerInterface type |
| 3 runnable examples | `examples/basic/`, `examples/middleware/`, `examples/per-path-debounce/` |
| Op serialization | `MarshalText`/`UnmarshalText` on `Op`, JSON tags on `Event` |
| GlobalDebouncer.Pending() | API consistency with per-key `Debouncer` |
| ADR: samber/do/v2 | Evaluated and rejected — documented in `docs/adr/` |
| Regex pre-compilation | `FilterRegex` compiles pattern once at construction |
| Justfile | 20+ recipes including `check`, `ci`, `lint-fix`, `bench`, `build-all` |
| CHANGELOG.md | Unreleased section with all features listed |
| AGENTS.md | Agent onboarding guide with commands, conventions, gotchas |

**Total exported API surface:** ~55 symbols (13 types + 47 functions/methods + 5 errors + 1 var)

---

## B) PARTIALLY DONE 🟡

| Item | What's Done | What's Missing |
|------|-------------|----------------|
| Linter compliance | 57 linters configured, most pass | **17 issues remain:** 10 exhaustruct (all in `filter_test.go` — missing `IsDir` field), 5 gocritic (`exitAfterDefer`), 1 golines (line too long), 1 recvcheck (`Op` mixed receivers) |
| Test coverage | 78.9%, 50+ tests | Target 85%+; missing edge-case tests for error paths, concurrent access, and `handleNewDirectory` |
| README | Installation, quick start, options table, filters, middleware, event types | No advanced usage, no architecture diagram, no migration guide, no DI integration docs |
| Documentation | 10 status reports + 1 ADR | No CONTRIBUTING.md, CODEOWNERS, or ROADMAP.md |

---

## C) NOT STARTED ❌

| Item | Priority | Notes |
|------|----------|-------|
| **Fix race condition in `watcher.go:312`** | 🔴 CRITICAL | `walkAndAddPaths` appends to `watchList` without lock. Called from `handleNewDirectory` in watch goroutine. 17 test failures with `-race` |
| GitHub Actions CI | High | No pipeline exists — all quality checks are manual |
| Tag v0.1.0 release | High | Blocked on race fix + linter cleanup |
| Goreleaser | Medium | Cross-compilation config exists in justfile but no goreleaser |
| Benchmarks | Medium | `just bench` exists but no benchmark functions written |
| Stress tests (10k+ files) | Medium | Not tested under load |
| CONTRIBUTING.md | Medium | No contribution guidelines |
| Dependabot / Renovate | Low | No automated dependency updates |
| `context.Context` in DebouncerInterface | Low | No context support in debouncer |
| `WithPollInterval` fallback | Low | No polling fallback for platforms without fsnotify |
| `FilterModifiedSince(t)` | Low | Time-based filter not implemented |
| `MiddlewareThrottle` | Low | N events per duration middleware not implemented |
| `Watcher.IsWatching()` | Low | Convenience method not added |
| `Event.Size` field | Low | File size not tracked |
| `Errors() <-chan error` | Low | Alternative error channel not implemented |
| Extract `fsnotify.Watcher` behind interface | Low | ADR recommendation — enables mock testing |
| Push to origin | Low | 1 commit ahead of origin |

---

## D) TOTALLY FUCKED UP 💥

### 1. Data Race in `walkAndAddPaths` — `watcher.go:312`

**Severity:** 🔴 CRITICAL — ship-blocker

```
w.watchList = append(w.watchList, root)  // NO LOCK HELD
```

**Root cause:** `handleNewDirectory()` (called from `watchLoop` goroutine) calls `addPath()` → `walkAndAddPaths()` without holding `w.mu`. Meanwhile `WatchList()` and `Stats()` read `watchList` under `RLock` concurrently.

**Impact:** 17 of 50+ tests fail under `go test -race`. The library CANNOT be used in production until fixed.

**Fix approach:** Acquire `w.mu.Lock()` in `walkAndAddPaths` before appending, or protect `watchList` mutation in `handleNewDirectory` before calling `addPath`.

### 2. `go test -race` Status: 17 FAILURES

Tests failing (all due to the same root cause):
- `TestWatcher_IgnoreDirs`
- `TestWatcher_Watch_ErrorHandler`
- `TestNew_FilePath`
- `TestEvent_JSON`
- `TestFilterGlob`
- `TestDebouncer_Debounce`
- `TestDebouncer_RapidCalls`
- `TestFilterExtensions`
- `TestFilterIgnoreHidden`
- And 8 more (transitive — race in one goroutine poisons others via `t.Parallel()`)

### 3. `Op` Mixed Receivers — `event.go`

```go
func (op Op) String() string           // value receiver
func (op Op) MarshalText() ([]byte, error) // value receiver  
func (op *Op) UnmarshalText(text []byte) error // pointer receiver ← MISMATCH
```

`recvcheck` linter flags this. Self-introduced in commit `6d934dc`. Should be all value receivers (since `Op` is `int`, copying is cheap).

### 4. Flaky Test: `TestWatcher_Watch_Deletes`

Intermittent timeout waiting for Remove event. Race between file removal and fsnotify event delivery on macOS. Not reliably reproducible.

### 5. `cockroachdb/errors` Pulls 8 Transitive Deps

Including `sentry-go`, `gogo/protobuf`, `kr/pretty`, `kr/text`, `logtags`, `redact`, `pkg/errors`, `rogpeppe/go-internal`. For a library with 4 sentinel errors and ~10 `Wrap` calls, this is massive overkill. Stdlib `fmt.Errorf("%w", err)` does the same.

---

## E) WHAT WE SHOULD IMPROVE 📈

### Architecture

1. **Fix the race condition** — This is THE priority. No other work matters until the library is race-free.
2. **Remove `cockroachdb/errors`** — Replace with stdlib `errors.New` + `fmt.Errorf`. Eliminates 8 transitive deps. This is a utility library — minimal deps is the feature.
3. **Extract `fsnotify.Watcher` behind interface** — Enables proper unit testing without filesystem dependencies.
4. **Split `watcher.go`** (548 lines) — Consider splitting into `watcher.go` (core), `watcher_ops.go` (Add/Remove/WatchList/Stats), and `watcher_internal.go` (walkAndAddPaths, handleNewDirectory, convertEvent).

### Testing

5. **Target 85%+ coverage** — Currently 78.9%. Missing: error paths, concurrent access patterns, `handleNewDirectory`, `Remove()`.
6. **Add benchmarks** — Measure debouncer throughput, filter chain latency, middleware overhead.
7. **Stress tests** — 10k+ files, rapid create/delete cycles, concurrent Add/Remove.
8. **Race test in CI** — `go test -race` must be a hard gate.

### Developer Experience

9. **CI/CD pipeline** — GitHub Actions with `just ci` + `go test -race`. No excuses.
10. **Tag v0.1.0** — After race fix + linter cleanup. Ship it.
11. **CONTRIBUTING.md** — How to contribute, development setup, PR requirements.
12. **README expansion** — Architecture diagram, advanced usage, DI integration patterns.

### Code Quality

13. **Fix all 17 linter issues** — Mostly mechanical (exhaustruct `IsDir: false`, receiver consistency, line length).
14. **Fix `convertEvent` losing combined ops** — fsnotify `Create|Write` becomes just `Create`. Should either emit multiple events or add `Ops []Op` bitmask.
15. **Fix `shouldSkipDir` disconnect** — `WithIgnoreDirs("build")` filters events but doesn't prevent directory walking, wasting kernel file descriptors.
16. **Cache file handle in `MiddlewareWriteFileLog`** — Opens file per event → potential fd exhaustion under burst.

---

## F) TOP 25 THINGS TO DO NEXT (Prioritized)

| # | Task | Priority | Effort | Blocked By |
|---|------|----------|--------|------------|
| 1 | Fix race condition in `walkAndAddPaths` (`watcher.go:312`) | 🔴 Critical | Small | Nothing |
| 2 | Fix `Op` mixed receivers (`event.go`) — use all value receivers | 🔴 Critical | Trivial | Nothing |
| 3 | Fix 10 exhaustruct issues in `filter_test.go` — add `IsDir: false` | High | Trivial | Nothing |
| 4 | Fix 5 gocritic `exitAfterDefer` issues | High | Small | Nothing |
| 5 | Fix 1 golines issue (line too long) | High | Trivial | Nothing |
| 6 | Run `go test -race ./...` and confirm 0 failures | 🔴 Critical | — | #1 |
| 7 | Decide: remove `cockroachdb/errors` or keep? | High | Decision | User input |
| 8 | Remove `cockroachdb/errors` (if decided) — replace with stdlib | High | Medium | #7 |
| 9 | Tag v0.1.0 release | High | Trivial | #1-6 |
| 10 | GitHub Actions CI pipeline (`just ci` + `go test -race`) | High | Medium | Nothing |
| 11 | Push to origin | Medium | Trivial | Nothing |
| 12 | Fix `TestWatcher_Watch_Deletes` flakiness | Medium | Medium | Nothing |
| 13 | Raise test coverage to 85%+ | Medium | Medium | #1 |
| 14 | Add benchmarks (debouncer, filters, middleware) | Medium | Medium | Nothing |
| 15 | Add stress tests (10k+ files) | Medium | Medium | #1 |
| 16 | Extract `fsnotify.Watcher` behind interface | Medium | Medium | Nothing |
| 17 | Fix `convertEvent` combined ops (emit multiple or bitmask) | Medium | Small | Nothing |
| 18 | Fix `shouldSkipDir` to respect `WithIgnoreDirs` during walking | Medium | Small | Nothing |
| 19 | Split `watcher.go` into focused files | Medium | Small | #1 |
| 20 | Add CONTRIBUTING.md + CODEOWNERS | Low | Small | Nothing |
| 21 | Goreleaser configuration | Low | Medium | #9 |
| 22 | Replace `log.Logger` with `log/slog` | Low | Medium | Nothing |
| 23 | Cache file handle in `MiddlewareWriteFileLog` | Low | Small | Nothing |
| 24 | Add `Errors() <-chan error` method | Low | Small | Nothing |
| 25 | Validate in real projects (file-and-image-renamer, dynamic-markdown-site) | Low | Medium | #1, #9 |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF 🤔

**Should we remove `cockroachdb/errors`?**

- **For removal:** 4 sentinel errors + ~10 `Wrap` calls. Stdlib does this. Eliminates 8 transitive deps (including `sentry-go`). This is a utility library — minimal deps is THE selling point.
- **For keeping:** `go-cqrs-lite` uses `cockroachdb/errors`. If you want consistency across your ecosystem, keeping it makes sense. Stack traces ARE useful for debugging.

**I need your decision.** This affects the module's dependency footprint permanently.
