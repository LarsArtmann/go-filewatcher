# Status Report — go-filewatcher Improvement Sprint

**Date:** 2026-04-04 19:32 CEST
**Branch:** master (6 commits ahead of origin)
**Tests:** PASSING (root package, 3.7s)
**Coverage:** ~77%
**Go:** 1.26.0 (fixed from broken 1.26.1)

---

## a) FULLY DONE

| # | What | Commit |
|---|------|--------|
| 1 | Fix go.mod version 1.26.1→1.26.0 | `4f663fd` |
| 2 | Remove pkg/errors/ template junk | `b14cef3` |
| 3 | Pre-compile regex in FilterRegex | `d5f3a40` |
| 4 | Remove FilterCustom dead alias | `f21fc03` |
| 5 | Fix justfile GOWORK=off everywhere | `f21fc03` |
| 6 | Clean test dead code (_ = w) | `f21fc03` |
| 7 | All prior implementation (8 source files, 50 tests, docs) | `ac0d50b..617678f` |

---

## b) PARTIALLY DONE

| What | Status | Blocker |
|------|--------|---------|
| Comprehensive improvement analysis | Analysis complete, execution started | Hit Go build cache corruption mid-run |

---

## c) NOT STARTED

| # | What | Impact | Work | Priority |
|---|------|--------|------|----------|
| 1 | Fix handleNewDirectory race (writes watchList without lock) | HIGH | LOW | 🔴 |
| 2 | Make shouldSkipDir respect user WithIgnoreDirs | HIGH | LOW | 🔴 |
| 3 | Replace cockroachdb/errors with stdlib errors | HIGH | MED | 🟠 |
| 4 | Split watcher.go (549 lines) into watcher.go + lifecycle.go + internal.go | MED | LOW | 🟡 |
| 5 | Add Op.MarshalText/UnmarshalText for JSON | MED | LOW | 🟡 |
| 6 | Add slog support to MiddlewareLogging | MED | LOW | 🟡 |
| 7 | Fix MiddlewareWriteFileLog (opens file on every event) | MED | LOW | 🟡 |
| 8 | Raise test coverage from 77% → 90%+ | HIGH | MED | 🔴 |
| 9 | Fix getDebounceKey type assertion smell | LOW | LOW | 🟢 |
| 10 | Remove report/ directory (jscpd-report.json) | LOW | LOW | 🟢 |
| 11 | Integrate into file-and-image-renamer | HIGH | MED | 🔴 |
| 12 | Integrate into dynamic-markdown-site | MED | LOW | 🟡 |
| 13 | Integrate into auto-deduplicate | MED | LOW | 🟡 |
| 14 | Integrate into Cyberdom | MED | LOW | 🟡 |
| 15 | Tag v0.1.0 | MED | LOW | 🟢 |
| 16 | Update README/CHANGELOG with all changes | MED | LOW | 🟡 |
| 17 | Add WithWatchedIgnoreDirs option (separate filter vs. walk skip) | HIGH | MED | 🔴 |
| 18 | Remove `nolint:unparam` from getDebounceKey | LOW | LOW | 🟢 |
| 19 | Add `Pending()` to DebouncerInterface | LOW | LOW | 🟢 |
| 20 | Validate debounce durations (cap at reasonable max) | LOW | LOW | 🟢 |
| 21 | Add Example_FilterRegex test (currently has no Output comment) | LOW | LOW | 🟢 |
| 22 | Consider removing cockroachdb/errors entirely (2 deps → 0) | MED | MED | 🟠 |
| 23 | Add `Errors() <-chan error` method as alternative to error handler callback | MED | MED | 🟡 |
| 24 | Check if examples/ directory is worth keeping vs. just example_test.go | LOW | LOW | 🟢 |
| 25 | Ensure FilterRegex compiles are validated in constructor, not at runtime | LOW | LOW | 🟢 |

---

## d) TOTALLY FUCKED UP

| What | Details |
|------|---------|
| Go build cache corruption | `go clean -cache` failed with "directory not empty". Background test jobs hung. Had to kill them. Root cause: go.mod was 1.26.1 but toolchain is 1.26.0 — fixed, but stale cache artifacts remain. May need manual `rm -rf ~/Library/Caches/go-build/` |
| TestWatcher_Watch_Deletes flakiness | Failed once during my analysis run (`timed out waiting for remove event`). Race between file removal and fsnotify event delivery. Not deterministic — passes on retry. |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture Issues

1. **cockroachdb/errors is overkill for this library** — We use 4 sentinel errors and `errors.WithStack`/`errors.Wrap`. Stdlib `errors.New` + `fmt.Errorf("...: %w", err)` does the same. Removing it eliminates 6 transitive deps (logtags, redact, sentry-go, gogo/protobuf, kr/pretty, pkg/errors). This is a utility library — minimal deps is a feature.

2. **handleNewDirectory has a race** — `watchList` is appended in `walkAndAddPaths` → `walkDirFunc` without holding `mu`. The `watchLoop` goroutine calls `handleNewDirectory` which calls `addPath` → `walkAndAddPaths` → `walkDirFunc` → `w.watchList = append(...)` — all without the lock. The `mu.RLock()` in `handleNewDirectory` only protects the `closed` check, not the `watchList` mutation.

3. **shouldSkipDir ignores user-configured dirs** — `WithIgnoreDirs` only adds a *filter*, but `shouldSkipDir` (used during directory walking) only checks `DefaultIgnoreDirs`. If a user says `WithIgnoreDirs("build")`, directories named "build" still get added to fsnotify — they're just filtered from events. This wastes kernel file descriptors.

4. **watcher.go is 549 lines** — Contains New(), Watch(), Add(), Remove(), WatchList(), Stats(), Close(), plus 10+ internal methods. Should split: `watcher.go` (public API), `lifecycle.go` (init/close/stats), `internal.go` (addPath, watchLoop, processEvent, etc).

5. **getDebounceKey uses type assertion** — Checks `w.debounceInterface.(*Debouncer)` to decide per-path vs global key. Should instead have the debouncer itself provide the key strategy, or have a boolean flag.

### Type Model Issues

6. **Op lacks JSON serialization** — No `MarshalText`/`UnmarshalText`. Users who want to serialize events to JSON (for logging, audit trails, APIs) must implement this themselves. Easy win.

7. **Event is a public struct with no constructor** — Unlike go-cqrs-lite's `Event` interface + `Core` struct pattern. For a utility library this is fine (exported struct with fields), but we could add `NewEvent(path string, op Op) Event` for validation.

8. **MiddlewareLogging uses `log.Logger`** — In Go 1.21+, `slog` is the standard. Should accept `slog.Handler` or `*slog.Logger` instead of `*log.Logger`.

### Code Quality

9. **MiddlewareWriteFileLog opens a file on every event** — `os.OpenFile` in the hot path. Should open once at middleware creation and close via a `Stop()` or finalizer.

10. **report/jscpd-report.json** — Template artifact, should be removed.

11. **examples/ directory duplicates example_test.go** — Both exist. The `examples/` dir has standalone mains; `example_test.go` has runnable godoc examples. Worth keeping both, but should verify examples actually compile.

### Missing Features for Real-World Use

12. **No `Errors() <-chan error` method** — The error handler callback is Go-1.14 style. A channel-based error stream matches the `Watch() → <-chan Event` pattern and composes better with `select`.

13. **No validation of debounce durations** — `WithDebounce(0)` silently uses default 500ms. `WithDebounce(24*time.Hour)` is accepted. Should validate and return errors, or at least document the behavior clearly.

---

## f) TOP 25 NEXT ACTIONS (Ranked by Impact × Ease)

| Rank | Action | Impact | Work | Why |
|------|--------|--------|------|-----|
| 1 | Fix handleNewDirectory race condition | HIGH | LOW | Data race on watchList — will bite in production |
| 2 | Make shouldSkipDir respect user ignore dirs | HIGH | LOW | Wastes kernel FDs, confusing behavior |
| 3 | Raise test coverage to 90%+ | HIGH | MED | Current 77% misses error paths, middleware, edge cases |
| 4 | Remove cockroachdb/errors → stdlib errors | HIGH | MED | 6 transitive deps for 4 sentinel errors is absurd |
| 5 | Integrate into file-and-image-renamer | HIGH | MED | Fixes confirmed bug, validates library in real project |
| 6 | Fix Go build cache corruption | HIGH | LOW | `rm -rf ~/Library/Caches/go-build/` needed |
| 7 | Add Op.MarshalText/UnmarshalText | MED | LOW | JSON support for audit/logging pipelines |
| 8 | Add slog support to MiddlewareLogging | MED | LOW | log.Logger is legacy; slog is standard since Go 1.21 |
| 9 | Fix MiddlewareWriteFileLog perf | MED | LOW | Opens file on every event |
| 10 | Add WithWatchedIgnoreDirs (walk-level skip) | MED | MED | Separate filter-level vs walk-level ignore |
| 11 | Split watcher.go into 3 files | MED | LOW | 549 lines is hard to navigate |
| 12 | Remove report/ template artifact | LOW | LOW | Dead file |
| 13 | Update README/CHANGELOG | MED | LOW | Reflect all improvements |
| 14 | Integrate into dynamic-markdown-site | MED | LOW | Second validation point |
| 15 | Fix getDebounceKey type assertion smell | LOW | LOW | Replace with boolean flag or interface method |
| 16 | Add `Errors() <-chan error` method | MED | MED | Better than callback for composability |
| 17 | Integrate into auto-deduplicate | MED | LOW | Collapses double abstraction |
| 18 | Validate debounce durations | LOW | LOW | Prevent footguns |
| 19 | Tag v0.1.0 | MED | LOW | After integrations pass |
| 20 | Integrate into Cyberdom | MED | LOW | Generic markdown live-reload validation |
| 21 | Clean nolint directives | LOW | LOW | Remove `nolint:unparam` from getDebounceKey |
| 22 | Add Pending() to DebouncerInterface | LOW | LOW | Completes the interface |
| 23 | Verify examples/ compile | LOW | LOW | May be broken after recent changes |
| 24 | FilterRegex already validated at compile — confirm | LOW | LOW | regexp.MustCompile panics on bad patterns |
| 25 | Add benchmark tests | LOW | LOW | Prove performance claims |

---

## g) TOP #1 QUESTION

**Should we keep cockroachdb/errors or switch to stdlib?**

Arguments for keeping: go-cqrs-lite uses it, stack traces are nice, consistency across projects.

Arguments for removing: This is a *utility library* — the entire point is minimal deps. We have 4 sentinel errors and ~10 `errors.Wrap` calls. Stdlib does all of this. cockroachdb/errors pulls in 6 transitive deps (sentry-go, gogo/protobuf, etc). Users who care about stack traces can wrap errors themselves.

**I recommend removing it.** The library should have zero transitive deps beyond fsnotify. This is a utility, not a framework.
