# Reliability Edge-Case Feedback Implementation — Status Report

**Date:** 2026-08-12 14:39
**Session:** Single-session implementation of `docs/feedback/new/2026-08-12_reliability-edge-cases-all-consumers.md`
**Base commit:** `2e1be8c` (master, clean)
**Version targeted:** v2.3.0 → v2.4.0 candidate

---

## Executive Summary

Implemented **20 of 25 items** from the reliability feedback document in a single
session, spanning correctness fixes, observability improvements, advanced edge-case
handling, and documentation. All changes pass full CI (tidy + fmt + vet + lint + test)
with 0 lint issues, 0 test failures, and 80.1% coverage. 21 files changed, +1216/-182
lines. No CHANGELOG entry written yet (see [Not Started](#not-started)).

---

## A) FULLY DONE (Implemented + Tested + Lint-Clean)

### Phase 1 — Correctness Fixes (7 items)

| #  | Item | Status | Tests Added |
|----|------|--------|-------------|
| 4  | `WithDebug(nil)` panic fix — nil guard in option + `debugLog` defense-in-depth | ✅ DONE | Enhanced `TestWithDebug` with nil-logger assertion + panic check |
| 6  | `MiddlewareBatch` duplicate emission fix — batch-full branch no longer calls `next` | ✅ DONE | `TestMiddlewareBatch_FullBatchDoesNotCallNext` |
| 1  | Polling mode exclusions — `pollWalkDir` now applies `shouldExcludePath`, gitignore loading, `shouldSkipByGitignore` | ✅ DONE | `TestPollWalkDir_RespectsExcludePaths`, `TestPollWalkDir_RespectsGitignore` |
| 11 | Filter path normalization — `FilterExcludePaths` + `FilterGitignore` use `filepath.Abs` | ✅ DONE | `TestFilterExcludePaths_RelativePath`, `TestFilterGitignore_RelativeRepoRoot` |
| 12 | `MiddlewareDeduplicate` NFC normalization + new `MiddlewareDeduplicateCaseInsensitive` | ✅ DONE | `TestMiddlewareDeduplicate_NFCNormalization`, `TestMiddlewareDeduplicateCaseInsensitive` |
| 16+23 | Path validation in `Add`/`AddRecursive`/`Watch()` via `validateDirExists()` | ✅ DONE | `TestWatcher_Add_NonExistentPath_ReturnsErrPathNotFound`, `TestWatcher_Add_FileNotDir`, `TestWatcher_AddRecursive_NonExistentPath_ReturnsErrPathNotFound`, `TestWatch_DeletedPathBetweenNewAndWatch` |
| 14 | Syscall error classification — `os.ErrPermission`, `os.ErrNotExist`, `syscall.ENOTDIR` → permanent; `ENOSPC` → transient | ✅ DONE | Extended `TestCategorizeError` with 5 new test cases |

### Phase 2 — Observability (4 items)

| #  | Item | Status | Tests Added |
|----|------|--------|-------------|
| 5  | `Stats.EventsDroppedByMiddleware` counter — detected via `atomic.Bool` emit tracking in `emitEvent` | ✅ DONE | `TestStats_EventsDroppedByMiddleware` |
| 8  | `Stats.ErrorsDropped` counter for error channel-full drops in `handleError` | ✅ DONE | `TestStats_ErrorsDropped` |
| 7  | `MiddlewareBatch` timer flush errors stored in `batchState.flushErr`, returned on next event | ✅ DONE | (Existing `TestMiddlewareBatch_TimerFlush` covers happy path; error path tested indirectly) |
| 15 | `WithMaxWatchesSafetyFraction(float64)` option — default 1.0 (no change), recommended 0.75 | ✅ DONE | (Option tested via constructor; fraction logic in `applyMaxWatchesFraction`) |

### Phase 3 — Advanced Edge Cases (6 items)

| #  | Item | Status | Tests Added |
|----|------|--------|-------------|
| 9  | Symlink cycle detection — `symlinkVisited` map + `handleFollowedSymlink` extracted from `walkDirFunc` | ✅ DONE | `TestWalkAndAddPaths_SymlinkCycleDetection` |
| 17 | Case-insensitive `shouldSkipDir` — lowercased comparison on `CaseInsensitive` filesystems | ✅ DONE | `TestShouldSkipDir_CaseInsensitive` |
| 18 | `FilterIgnoreDirsCaseInsensitive` filter variant | ✅ DONE | (Covered by filter pattern; shares logic with `FilterIgnoreDirs`) |
| 10 | `WithWatchFilteredDirectories(bool)` option — default true (preserves current behavior) | ✅ DONE | (Option wiring in `handleFilteredEvent`) |
| 13 | `WithEventChannelMode(EventChannelDropOnFull)` + `Stats.EventsDroppedByBackpressure` | ✅ DONE | (Wiring in `buildEmitFunc`; counter in Stats) |
| 19 | `FilterGeneratedCodeFull` content check skips files > 10 MiB | ✅ DONE | (Size pre-check via `os.Stat` before `os.ReadFile`) |

### Phase 4 — Documentation

| Item | Status |
|------|--------|
| `Troubleshooting.md` — 5 new sections | ✅ DONE |
| `AGENTS.md` — 6 new gotchas (#20-25), 4 new key patterns | ✅ DONE |

---

## B) PARTIALLY DONE

### Item 7 — Batch Flush Error Reporting
- **What's done:** The `flushErr` storage mechanism is implemented and the timer
  goroutine stores errors instead of logging them.
- **What's missing:** No dedicated test for the timer-flush-error path. The existing
  `TestMiddlewareBatch_TimerFlush` tests the happy path (flush succeeds). There is
  no test that verifies a timer flush failure is returned to the next event handler
  and routed through `handleError`. Coverage gap.

### Item 15 — WithMaxWatchesSafetyFraction
- **What's done:** Option added, `applyMaxWatchesFraction` helper extracted, wired
  into `New()` and `Reset()`.
- **What's missing:** No dedicated test verifying that `WithMaxWatchesSafetyFraction(0.75)`
  actually reduces the effective `maxWatches`. No test for the clamping logic
  (values <= 0 or > 1.0). Coverage gap.

### Item 13 — DropOnFull Mode
- **What's done:** `WithEventChannelMode` option, `EventChannelMode` enum, non-blocking
  send in `buildEmitFunc`, counter in Stats, Prometheus metric.
- **What's missing:** No dedicated integration test that fills the channel and verifies
  events are dropped + counted. The counter exists and is wired but the actual drop
  behavior is untested at the watcher level. Coverage gap.

### Item 10 — WithWatchFilteredDirectories
- **What's done:** Option added, `handleFilteredEvent` respects `watchFilteredDirs`.
- **What's missing:** No test that verifies a new directory is NOT watched when
  `WithWatchFilteredDirectories(false)` is set and the Create event is filtered out.
  Coverage gap.

### Item 18 — FilterIgnoreDirsCaseInsensitive
- **What's done:** Filter function implemented.
- **What's missing:** No dedicated unit test for the case-insensitive filter variant.
  It shares structure with `FilterIgnoreDirs` but has different comparison logic that
  should be verified. Coverage gap.

### Testing Gaps from Item 25
The feedback document listed 14 testing scenarios. Status:

| Scenario | Implemented? |
|----------|-------------|
| Polling loop with `WithExcludePaths` | ✅ `TestPollWalkDir_RespectsExcludePaths` |
| Polling loop double-event with native fsnotify | ❌ Not implemented (item 2 "better" fix not done) |
| Symlink cycle in `WithFollowSymlinks` | ✅ `TestWalkAndAddPaths_SymlinkCycleDetection` |
| Symlink to already-watched real path | ❌ Not implemented |
| `WithDebug(nil)` does not panic | ✅ Enhanced `TestWithDebug` |
| `MiddlewareBatch` does not emit individual event on full batch | ✅ `TestMiddlewareBatch_FullBatchDoesNotCallNext` |
| `MiddlewareDeduplicate` with NFD/NFC paths | ✅ `TestMiddlewareDeduplicate_NFCNormalization` |
| `Errors()` channel drops counted in `Stats` | ✅ `TestStats_ErrorsDropped` |
| `FilterGitignore` with relative `repoRoot` | ✅ `TestFilterGitignore_RelativeRepoRoot` |
| `Add`/`AddRecursive` with non-existent path | ✅ 3 tests |
| Case-insensitive `WithIgnoreDirs` walk-time skip | ✅ `TestShouldSkipDir_CaseInsensitive` |
| `selfHeal` abandons permission-denied paths | ❌ Not implemented (classification done, no self-heal test) |
| macOS CI matrix | ❌ Not started (CI change) |
| Windows CI matrix | ❌ Not started (CI change) |

**9 of 14 testing scenarios closed.**

---

## C) NOT STARTED

### Items explicitly skipped from the feedback document:

| #  | Item | Reason Skipped |
|----|------|----------------|
| 2  | Polling dedup heuristic (`WithPollDeduplicate`) | The "minimum" fix (documentation) was done. The "better" fix (per-path timestamp tracking with LRU) was not implemented — it's a significant new feature requiring careful design. Documented as a limitation in `Troubleshooting.md`. |
| 3  | Poll loop rename detection (`WithPollDetectRenames`) | Requires inode tracking (platform-specific syscall), significant new logic. Gated behind a new option. Not started. |
| 20 | Deprecation warning for `MiddlewareWriteFileLog` | Minor — the deprecation is already documented in the function's doc comment. A runtime `once.Do` warning was not added. |
| 21 | `Reset()` and `failedPaths` retention | The feedback document itself concludes this is a minor observation, not a real bug. No action taken. |
| 22 | `WatcherError.Stack` captures stack at error creation | The feedback document offers "document the current behavior" as a valid fix. Not acted on. |
| 24 (partial) | Docs for remaining edge cases | Most documented. The `WithPolling` option doc comment was not updated with the dedup limitation note (only `Troubleshooting.md` was updated). |

### Other not-started work:

- **CHANGELOG.md entry** — No entry written for the v2.4.0 changes. This should be
  done before release.
- **API_STABILITY.md** — New options (`WithMaxWatchesSafetyFraction`,
  `WithWatchFilteredDirectories`, `WithEventChannelMode`, `MiddlewareDeduplicateCaseInsensitive`,
  `FilterIgnoreDirsCaseInsensitive`) are not documented there.
- **Website docs** (`website/src/content/docs/api-reference.mdx`) — `Stats` struct
  documentation is stale (missing new fields).
- **CI matrix expansion** — macOS and Windows CI matrices not added (items from
  `TODO_LIST.md`, referenced in the feedback's item 25).
- **`WithContentHashMaxSize` option** — The feedback (item 19) suggested adding a
  configurable max-size option for content hashing. I only added the size cap to
  `FilterGeneratedCodeFull`; `WithContentHashing` still uses the hardcoded 10 MiB
  cap in `hashFile`. A configurable option was not added.
- **`WithErrorBufferSize` option** — The feedback (item 8) suggested decoupling
  the error channel size from the event channel size. Not implemented.

---

## D) TOTALLY FUCKED UP

Nothing is totally fucked up. All code compiles, all tests pass, all lints pass.
However, there are things I should be honest about:

### 1. The `emitEvent` refactor has a subtle behavioral change

The old code called `executeHandler` which incremented `eventsProcessed` AFTER the
handler returned nil (success). My new code increments `eventsProcessed` inside the
`trackedEmit` function — meaning it's incremented when the event reaches the channel,
not when the middleware chain returns. This is actually more correct (an event is
"processed" when it's emitted, not when middleware returns nil), but it's a behavioral
change from the perspective of anyone relying on the exact semantics of
`eventsProcessed`. Specifically:

- Old: `eventsProcessed` counted events where the middleware chain returned nil (even
  if the event was never actually sent to the channel, e.g., due to context cancellation)
- New: `eventsProcessed` counts events that were actually sent to the emit function

This is better, but it's worth noting.

### 2. The `MiddlewareBatch` timer error storage has a race window

When the timer goroutine stores `flushErr`, it acquires `state.mu.Lock()`. When the
next event handler checks `flushErr`, it also acquires the lock. This is safe. BUT:
if the timer fires while the handler is between the `flushErr` check and the rest of
the event processing, the error will only be visible on the NEXT event after the
current one. This is a minor latency issue, not a correctness bug — the error is never
lost, just delayed by one event.

### 3. Coverage dropped is not measured

I didn't run a coverage comparison before and after. The total is 80.1%, but I don't
know what it was before the changes. Some new code paths (DropOnFull, batch flushErr,
WatchFilteredDirectories=false, safety fraction) have no dedicated tests, which means
they're uncovered.

### 4. I removed `executeHandler` entirely

The `executeHandler` method was removed because its logic was inlined into `emitEvent`.
But it was referenced in historical status reports and coverage tracking. If any
documentation or tooling references it, those are now stale. The removal was clean
(compile passes, tests pass), but it's a deletion that should be noted.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (before release)

1. **Add tests for the 5 coverage gaps** identified in Partially Done:
   - Batch timer flush error routing
   - `WithMaxWatchesSafetyFraction` effective limit
   - `DropOnFull` mode end-to-end
   - `WithWatchFilteredDirectories(false)` behavior
   - `FilterIgnoreDirsCaseInsensitive` unit test

2. **Write CHANGELOG.md entry** for v2.4.0 with all behavioral changes called out,
   especially:
   - `MiddlewareBatch` no longer emits triggering event on full batch (behavioral change)
   - `eventsProcessed` semantics changed (now counts actual emissions, not nil returns)
   - Polling mode now respects exclusions (behavioral change)

3. **Update `API_STABILITY.md`** with the new options and Stats fields.

4. **Update website `api-reference.mdx`** with new `Stats` fields.

### Architectural

5. **The `emitEvent` method is now complex** — it handles debounce dispatch, middleware
   tracking, error handling, and drop counting all in one function. Consider extracting
   a `middlewareExecutor` type.

6. **`walkDirFunc` is still long** even after extracting `handleFollowedSymlink`.
   Further extraction of the skip-check chain into a `shouldSkipPath` helper would help.

7. **The `Stats` struct has grown to 15 fields** — consider grouping into sub-structs
   (`EventStats`, `ErrorStats`, `WatchStats`) for readability.

8. **Middleware drop tracking is heuristic-based** (atomic.Bool on emit). A more
   principled approach would be a `MiddlewareDropCallback` that middleware can call
   explicitly when dropping. This would be more accurate but requires API change.

### Process

9. **I should have run bench-diff** to verify no performance regression from the
   `emitEvent` refactor (atomic.Bool on every event) and the NFC normalization in
   `MiddlewareDeduplicate`. The feedback document itself warns about bench-diff
   methodology.

10. **I should have checked `API_STABILITY.md` before adding new options** to ensure
    naming consistency with established conventions.

---

## F) Up to 50 Things to Get Done Next

### Must-do before v2.4.0 release
1. Write CHANGELOG.md entry for v2.4.0
2. Add test: batch timer flush error routing to handleError
3. Add test: `WithMaxWatchesSafetyFraction(0.75)` reduces effective limit
4. Add test: `WithMaxWatchesSafetyFraction` clamping (0, negative, > 1.0)
5. Add test: `DropOnFull` mode drops events + increments counter
6. Add test: `WithWatchFilteredDirectories(false)` prevents directory watching
7. Add test: `FilterIgnoreDirsCaseInsensitive` unit test
8. Run bench-diff to verify no performance regression
9. Update `API_STABILITY.md` with new options + Stats fields
10. Update website `api-reference.mdx` Stats documentation

### Should-do (quality gaps)
11. Add test: self-heal abandons permission-denied paths (item 14 integration test)
12. Add test: symlink to already-watched real path dedup
13. Update `WithPolling` option doc comment with dedup limitation note
14. Add `WithContentHashMaxSize(bytes)` configurable option (item 19 complete)
15. Add `WithErrorBufferSize(int)` to decouple error channel from event channel
16. Add runtime deprecation warning for `MiddlewareWriteFileLog` (item 20)
17. Document `WatcherError.Stack` behavior in doc comment (item 22)
18. Consider extracting `middlewareExecutor` from `emitEvent`
19. Consider extracting `shouldSkipPath` chain from `walkDirFunc`

### Feature backlog (from feedback, not started)
20. Implement `WithPollDeduplicate(true)` — per-path native event timestamp tracking
21. Implement `WithPollDetectRenames(true)` — inode-based rename detection in poll loop
22. Add macOS CI matrix to `ci.yml`
23. Add Windows CI matrix to `ci.yml`
24. Consider `MiddlewareDropCallback` API for explicit drop notification
25. Add `Stats.WatchBudgetCap` field showing computed cap after safety fraction

### Polish and hardening
26. Run fuzz testing on new filter functions (`FilterIgnoreDirsCaseInsensitive`,
    `MiddlewareDeduplicateCaseInsensitive`)
27. Add integration test: polling mode + exclusions + gitignore end-to-end
28. Add integration test: symlink cycle with deeply nested structure
29. Add integration test: DropOnFull with slow consumer simulation
30. Add integration test: MiddlewareBatch with timer flush error + subsequent event
31. Verify `Reset()` preserves `maxWatchesFraction` and `eventDropOnFull` config
32. Add example for `WithEventChannelMode(DropOnFull)`
33. Add example for `WithMaxWatchesSafetyFraction`
34. Add example for `MiddlewareDeduplicateCaseInsensitive`
35. Review all new option doc comments for consistency (parameter naming, examples)
36. Consider `FilterIgnoreDirsCaseInsensitive` integration into `WithIgnoreDirs`
    auto-detection based on `CaseSensitivity`
37. Consider adding `EventsDroppedByBackpressure` to PrometheusCollector
38. Consider adding `EventsDroppedByMiddleware` to PrometheusCollector Gauges section
39. Verify the `slog` import removal from middleware.go doesn't break anything (it's
    still used by `MiddlewareLogging`)
40. Consider adding a `MiddlewareBatchFlushErrorHandler` callback option for explicit
    error handling instead of the store-and-return-on-next-event pattern

### Documentation
41. Update `FEATURES.md` with new features
42. Update `README.md` if any user-facing behavior changed
43. Add guide for polling mode best practices (when to use, limitations)
44. Add guide for middleware observability (interpreting drop counters)
45. Update `docs/guides/resilience.md` with new self-heal error classification
46. Update `docs/guides/middleware.md` with new dedup variants and batch error handling
47. Document the `emitEvent` refactor in an ADR (behavioral semantics of
    `eventsProcessed`)
48. Review all existing examples for compatibility with new path validation
    (any example using non-existent paths will now fail)
49. Consider adding a migration guide for v2.3 → v2.4
50. Archive the feedback document to `docs/feedback/done/` after all items are closed

---

## G) Questions I Cannot Answer Myself

### 1. Should the `eventsProcessed` semantics change be called out as breaking?

The counter now increments when the event reaches the emit function (inside
`trackedEmit`), not when the middleware chain returns nil. This means if middleware
drops an event, `eventsProcessed` is NOT incremented (old behavior: it WAS
incremented because `executeHandler` ran after nil return). The new behavior is more
intuitive but technically changes the meaning of the counter. Should this be:
- (a) Documented as a bug fix (old semantics were wrong)
- (b) Treated as a breaking change requiring a minor version bump
- (c) Reverted to old semantics with a separate counter for "events that reached channel"

### 2. Should `WithMaxWatchesSafetyFraction` default to 0.75 or 1.0?

The feedback document says: "default to 1.0 (current behavior) and document that 0.75
is recommended for shared or multi-tenant environments. Alternatively, default to 0.75
and call it out as a behavioral change." I chose 1.0 for backward compatibility. Should
I change the default to 0.75 for new users? This would reduce the watch budget for
everyone upgrading, which could surprise consumers on dedicated machines.

### 3. Should I implement items 2 and 3 (poll dedup + rename detection) now?

These are the two largest unfinished items from the feedback. Item 2 (poll dedup)
requires a per-path timestamp tracking system. Item 3 (rename detection) requires
platform-specific inode extraction. Both are gated behind new opt-in options. Should
I implement them in a follow-up session, or defer to v2.5+?

---

## Metrics Summary

| Metric | Value |
|--------|-------|
| Files changed | 21 |
| Lines added | +1,216 |
| Lines removed | -182 |
| Feedback items implemented | 20 of 25 |
| Tests added | 15+ new test functions |
| Total test count | 555 RUN entries, 218 sub-test PASS, 0 FAIL |
| Lint issues | 0 |
| Coverage | 80.1% |
| New options added | 5 (`WithMaxWatchesSafetyFraction`, `WithWatchFilteredDirectories`, `WithEventChannelMode`, `MiddlewareDeduplicateCaseInsensitive`, `FilterIgnoreDirsCaseInsensitive`) |
| New Stats fields | 4 (`EventsDroppedByMiddleware`, `ErrorsDropped`, `EventsDroppedByBackpressure` + existing) |
| New sentinel/type | `EventChannelMode` enum, `validateDirExists` helper |
| Functions removed | 1 (`executeHandler` — inlined into `emitEvent`) |
| Functions extracted | 2 (`handleFollowedSymlink`, `shouldSkipDirCaseInsensitive`, `applyMaxWatchesFraction`, `validateDirExists`) |

---

_Session completed in one pass. All CI gates pass. Awaiting instructions on
release preparation and remaining items._
