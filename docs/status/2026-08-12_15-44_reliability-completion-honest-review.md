# Reliability Edge-Case Completion — Brutally Honest Status

**Date:** 2026-08-12 15:44
**Session:** Follow-up to prior session (base commit `6a075ef`)
**Prior session report:** `docs/status/2026-08-12_14-39_reliability-edge-cases-implementation.md`
**Commits this session:** Auto-git committed prior session's work as `4bd8246` and `499e162`
**Uncommitted changes:** 4 files modified, 1 new file (this report)

---

## Executive Summary

The prior session implemented 20 of 25 items from the reliability feedback
document. This session was tasked with closing the remaining gaps. I closed all
5 coverage gaps, found and fixed 2 production bugs in the prior session's work,
optimized an allocation regression, and updated all project documentation.

**Final state:** CI clean (0 lint, 0 test failures, 82.2% coverage).

But I made mistakes. Read on.

---

## A) FULLY DONE (Implemented + Tested + Lint-Clean + Verified)

### 1. Coverage Gap Tests (5 gaps → 7 tests)

| # | Gap | Test Added | File |
|---|-----|-----------|------|
| 1 | Batch timer flush error routing | `TestMiddlewareBatch_TimerFlushErrorReturnedOnNextEvent` | `middleware_test.go` |
| 2 | `WithMaxWatchesSafetyFraction` clamping | `TestWithMaxWatchesSafetyFraction_Clamping` (6 subtests) | `options_test.go` |
| 2b | Effective limit reduction | `TestApplyMaxWatchesFraction_ReducesAutoDetectedLimit` | `options_test.go` |
| 2c | Explicit limits not affected | `TestApplyMaxWatchesFraction_DoesNotAffectExplicitLimit` | `options_test.go` |
| 3 | DropOnFull mode drops + counts | `TestDropOnFull_DropsEventsAndCounts` | `watcher_test.go` |
| 4 | `WithWatchFilteredDirectories(false)` | `TestWatchFilteredDirectories_Disabled` | `watcher_test.go` |
| 5 | `FilterIgnoreDirsCaseInsensitive` | `TestFilterIgnoreDirsCaseInsensitive` (7 subtests) | `filter_test.go` |

All tests pass with `-race`. All use `t.Parallel()`.

### 2. Production Bug Fixes (2 bugs found in prior session's code)

**Bug 1: `applyMaxWatchesFraction` applied to explicit limits**

The prior session implemented `WithMaxWatchesSafetyFraction` with documentation
saying "This only affects auto-detected limits; explicit limits set via
`WithMaxWatches(n)` are used as-is." But `applyMaxWatchesFraction()` was called
unconditionally in `New()` after the auto-detect block, so it applied to ALL
`maxWatches > 0` — including explicit values. A user setting
`WithMaxWatches(1000)` + `WithMaxWatchesSafetyFraction(0.75)` would silently get
750.

**Fix:** Added `maxWatchesExplicit bool` field to `Watcher`. `WithMaxWatches(n)`
sets it when `n > 0`. `New()` and `Reset()` only call `applyMaxWatchesFraction()`
inside the auto-detect branch (`maxWatches == 0`).

**Bug 2: `Reset()` always re-detected max watches**

`Reset()` unconditionally called `w.maxWatches = detectMaxWatches()`, wiping any
explicit `WithMaxWatches(n)` setting. After `Reset()`, the user's configuration
was lost.

**Fix:** `Reset()` now checks `!w.maxWatchesExplicit` before re-detecting.

### 3. Performance Optimization

The prior session's `emitEvent` refactor added +2 allocations per event (from
`buildEmitFunc` closure + `trackedEmit` closure + `atomic.Bool`). I inlined
`buildEmitFunc` directly into `trackedEmit`, eliminating one closure allocation.
Removed `buildEmitFunc` entirely.

**Result (verified via `go test -bench`):**
- `EmitEvent_NoDebounce`: 8 allocs → 7 allocs (-1)
- `EmitEvent_WithMiddleware`: 14 allocs → 13 allocs (-1)
- Remaining +1 alloc vs baseline is the `trackedEmit` closure itself (minimum
  cost of `atomic.Bool` middleware drop tracking)

### 4. Documentation Updates

| Document | What Changed |
|----------|-------------|
| `CHANGELOG.md` | Full v2.4.0 section: 13 Added, 9 Fixed, 2 Changed items with behavioral change callouts |
| `API_STABILITY.md` | Added `WithMaxWatchesSafetyFraction`, `WithWatchFilteredDirectories`, `WithEventChannelMode` to Evolving Features; `FilterIgnoreDirsCaseInsensitive` to Evolving Filters; `MiddlewareDeduplicateCaseInsensitive` to Evolving Middleware; `EventChannelMode` to Evolving Types |
| `FEATURES.md` | Version bumped to v2.4.0; 8 new feature rows across Filtering, Middleware, Observability, Resilience |
| `website/api-reference.mdx` | Stats struct updated (3 new fields); options count 23→26; filters and middleware lists updated |
| `AGENTS.md` | Gotcha #26 (safety fraction scope); `buildEmitFunc` references updated to reflect inlining; DropOnFull pattern updated |
| `options.go` | `WithPolling` doc comment updated with dedup limitation cross-reference |

### 5. Status Reports

| Report | Location |
|--------|---------|
| Prior session | `docs/status/2026-08-12_14-39_reliability-edge-cases-implementation.md` (pre-existing) |
| This session | `docs/status/2026-08-12_reliability-edge-cases-completion.md` (written earlier) |
| This report | `docs/status/2026-08-12_15-44_reliability-completion-honest-review.md` |

### 6. Bench-Diff

Ran `nix run .#bench-diff` against the July 29 baseline. Allocation data (deterministic):
- **+1 alloc** on `EmitEvent_NoDebounce` and `EmitEvent_WithMiddleware` (trackedEmit closure)
- **All other benchmarks**: zero allocation regression
- ns/op data showed high variance (±49%) from CPU contention — unreliable for
  comparison, as the AGENTS.md bench-diff methodology note warns

---

## B) PARTIALLY DONE

### 1. `TestWatchFilteredDirectories_Disabled` — Correct but fragile

The test verifies `WatchCount <= 1` after creating a subdirectory with
`WithWatchFilteredDirectories(false)`. This works but:
- It depends on timing (uses `waitForCondition` with a 3s timeout)
- It doesn't verify the directory was actually _attempted_ to be watched and
  rejected — it just checks the count didn't grow
- On a slow CI machine, the initial walk might not have completed yet, making
  `WatchCount <= 1` true for the wrong reason

A better test would use the fake backend to verify `Add` was NOT called for the
new directory.

### 2. `TestDropOnFull_DropsEventsAndCounts` — Indirect verification

The test writes 20 files and waits for `EventsDroppedByBackpressure > 0`. This
proves the counter works but doesn't verify the exact count or that the dropped
events are truly gone from the channel. The test is more of a "smoke test" than
a precise behavioral verification.

### 3. `TestApplyMaxWatchesFraction_ReducesAutoDetectedLimit` — Reaches into internals

The test manually sets `watcher.maxWatches` and `watcher.maxWatchesExplicit` to
simulate an auto-detected limit, then calls `applyMaxWatchesFraction()`. This
tests the helper in isolation but doesn't verify the full `New()` → auto-detect →
fraction pipeline end-to-end. The reason: we can't control what
`detectMaxWatches()` returns on the test machine.

### 4. CHANGELOG entry is comprehensive but unstructured

The CHANGELOG lists all changes but doesn't separate "behavioral changes that
may affect consumers" from "pure additive changes." Consumers scanning the
changelog for migration concerns have to read every bullet. A "Breaking" or
"Behavioral Changes" subsection at the top would help.

---

## C) NOT STARTED

### From the original feedback document (25 items):

| # | Item | Status | Reason |
|---|------|--------|--------|
| 2 | Poll dedup heuristic (`WithPollDeduplicate`) | NOT STARTED | Large feature: per-path timestamp tracking with LRU. Documented as limitation. |
| 3 | Poll loop rename detection (`WithPollDetectRenames`) | NOT STARTED | Large feature: platform-specific inode extraction. |
| 20 | Runtime deprecation warning for `MiddlewareWriteFileLog` | NOT STARTED | Doc comment already warns. Runtime warning is v3 prep. |
| 21 | `Reset()` and `failedPaths` retention | NOT STARTED | Feedback document itself concludes this is minor, not a real bug. |
| 22 | `WatcherError.Stack` behavior | NOT STARTED | Feedback document offers "document current behavior" as valid fix. Not acted on. |

### From the prior session's "should-do" list:

| # | Item | Status |
|---|------|--------|
| - | macOS CI matrix | NOT STARTED |
| - | Windows CI matrix | NOT STARTED |
| - | `WithContentHashMaxSize(bytes)` configurable option | NOT STARTED (hardcoded 10 MiB cap exists) |
| - | `WithErrorBufferSize(int)` to decouple error channel | NOT STARTED |
| - | `MiddlewareDropCallback` API for explicit drop notification | NOT STARTED |
| - | `Stats.WatchBudgetCap` field | NOT STARTED |
| - | Fuzz testing on new filter/middleware functions | NOT STARTED |
| - | `MiddlewareBatchFlushErrorHandler` callback | NOT STARTED |

### From the prior session's polish list:

| # | Item | Status |
|---|------|--------|
| - | Verify `Reset()` preserves `maxWatchesFraction` and `eventDropOnFull` config | NOT VERIFIED — I fixed Reset() for maxWatches but did not write a test that verifies ALL new config fields survive Reset() |
| - | Add examples for new options | NOT STARTED |
| - | Migration guide for v2.3 → v2.4 | NOT STARTED |
| - | Archive feedback document to `docs/feedback/done/` | NOT STARTED |
| - | Update `docs/guides/resilience.md` with new error classification | NOT STARTED |
| - | Update `docs/guides/middleware.md` with new dedup variants | NOT STARTED |

---

## D) TOTALLY FUCKED UP

### 1. I wrote the `TestWatchFilteredDirectories_Disabled` test wrong the first time

I used `MiddlewareFilter` to reject Create events, but `handleFilteredEvent`
(gated by `watchFilteredDirs`) is called from `processEvent` when a **Filter**
rejects an event — NOT when middleware drops it. Middleware runs inside
`emitEvent`, which is called AFTER `handleNewDirectory`. So the test passed
locally (the directory was watched regardless) but failed on full CI because the
watch count grew.

I should have traced the code path before writing the test. I fixed it by using
`FilterNotOperations(Create)` via `WithFilter()` instead, but I wasted a full CI
round-trip on a preventable mistake.

**Root cause:** I read the `handleFilteredEvent` code but didn't trace where it
sits in the pipeline relative to middleware. I assumed middleware rejections and
filter rejections took the same code path. They don't.

### 2. I created a struct field alignment inconsistency

When I added `maxWatchesExplicit` and `maxWatchesFraction` to the `Watcher`
struct, my first edit produced misaligned fields:
```go
maxWatches        int     // wrong alignment
maxWatchesExplicit bool    // wrong alignment
```
The `gofmt` step in CI caught and fixed it, but I should have matched the
alignment from the start. I also initially forgot to add the new fields to the
`New()` struct literal, which would have failed the `exhaustruct` linter.

**Root cause:** I used `multiedit` with 4 edits at once and didn't verify the
result before running CI. I should have viewed the result after the edit.

### 3. I removed `buildEmitFunc` without checking if it was referenced elsewhere

I deleted `buildEmitFunc` and inlined its logic. I grepped for references after
the fact and found only a comment reference. But if there had been a reference
in a test file or another production file, I would have broken the build. I
should have run `grep` BEFORE the deletion, not after.

### 4. I didn't capture a fresh benchmark baseline before my changes

The existing `bench-baseline.txt` was from July 29 — before the prior session's
changes AND my changes. The bench-diff conflates both sessions' impact. I can't
isolate my changes' performance impact from the prior session's. If I had
captured a baseline from commit `6a075ef` before starting my work, I could have
diffed cleanly.

### 5. I wrote a completion status report prematurely

I wrote `docs/status/2026-08-12_reliability-edge-cases-completion.md` claiming
"All remaining coverage gaps closed" — but then the `TestWatchFilteredDirectories_Disabled`
test failed on full CI and I had to fix it. The report was written before I had
verified the final CI pass. I should have waited.

---

## E) WHAT WE SHOULD IMPROVE

### Architectural

1. **The `emitEvent` function is 50+ lines and does too much.** It handles
   debounce dispatch, middleware drop tracking, channel send logic (blocking and
   non-blocking), error handling, and debug logging — all in one function. The
   `trackedEmit` closure alone is 20 lines. Consider extracting a
   `channelSender` type that encapsulates the send strategy.

2. **Middleware drop tracking via `atomic.Bool` is a heuristic, not a contract.**
   It detects "middleware returned nil without calling emit" — but what if a
   middleware legitimately calls emit zero times AND drops zero times? Or calls
   emit multiple times? The `atomic.Bool` only records the first call. A
   `MiddlewareDropCallback` API would be more correct but requires an API change.

3. **The `Stats` struct has 15 fields.** It's becoming a god struct. Consider
   grouping into `EventStats`, `ErrorStats`, `WatchStats` sub-structs for
   readability and to make it easier to extend without breaking every struct
   literal in every test and metrics file.

4. **The `maxWatchesExplicit` flag is a workaround.** The real problem is that
   `maxWatches` serves double duty: "user's configured limit" and "effective
   limit after fraction." Consider splitting into `maxWatchesConfigured` (what
   the user set) and `maxWatchesEffective` (what the watcher uses). This
   eliminates the bool flag entirely.

5. **The `WithPolling` dedup limitation is now documented but not solved.** The
   "better" fix (per-path native event timestamp tracking with LRU) would
   eliminate the double-fire problem. It's a real feature gap for polling users.

### Process

6. **I should trace code paths before writing tests.** The
   `TestWatchFilteredDirectories_Disabled` failure was 100% preventable. I
   should have drawn the pipeline: `processEvent → passesFilters? →
   handleFilteredEvent (here!) → emitEvent → middleware`.

7. **I should capture bench baselines before starting work** so I can isolate
   my changes' performance impact.

8. **I should verify struct field alignment and initialization immediately after
   editing**, not rely on CI to catch formatting issues.

9. **I should write status reports AFTER final CI passes**, not before.

10. **I should run `grep` for references BEFORE deleting functions**, not after.

### Testing

11. **The `TestDropOnFull` test doesn't verify exact drop counts.** It just
    checks `> 0`. A precise test would fill the channel deterministically and
    verify the exact number of drops.

12. **No test verifies `Reset()` preserves new config fields.** I added
    `maxWatchesFraction`, `maxWatchesExplicit`, `eventDropOnFull`, and
    `watchFilteredDirs` but didn't test that they survive `Reset()`.

13. **No test verifies `Reset()` does NOT re-detect when `maxWatchesExplicit`
    is true.** My fix to `Reset()` is untested at the integration level.

14. **The bench-diff is unreliable for ns/op** due to CPU contention. The
    methodology note in AGENTS.md warns about this, but I still ran it without
    killing background processes first.

---

## F) Up to 50 Things to Get Done Next

### Must-do before v2.4.0 release (if not already done)

1. ~~Write CHANGELOG.md entry for v2.4.0~~ ✅ DONE
2. ~~Add test: batch timer flush error routing~~ ✅ DONE
3. ~~Add test: `WithMaxWatchesSafetyFraction` effective limit~~ ✅ DONE
4. ~~Add test: `WithMaxWatchesSafetyFraction` clamping~~ ✅ DONE
5. ~~Add test: DropOnFull mode drops events + increments counter~~ ✅ DONE
6. ~~Add test: `WithWatchFilteredDirectories(false)`~~ ✅ DONE
7. ~~Add test: `FilterIgnoreDirsCaseInsensitive`~~ ✅ DONE
8. ~~Run bench-diff~~ ✅ DONE
9. ~~Update `API_STABILITY.md`~~ ✅ DONE
10. ~~Update website `api-reference.mdx`~~ ✅ DONE
11. Capture fresh bench baseline from current HEAD for future comparisons
12. Verify `Reset()` preserves all new config fields (`maxWatchesFraction`,
    `maxWatchesExplicit`, `eventDropOnFull`, `watchFilteredDirs`) — write a test
13. Run `nix flake check` (full flake validation, not just CI apps)

### Should-do (quality gaps)

14. Add test: `Reset()` does NOT re-detect when `maxWatchesExplicit` is true
15. Add test: self-heal abandons permission-denied paths (item 14 integration)
16. Add test: symlink to already-watched real path dedup
17. Improve `TestDropOnFull` to verify exact drop count (deterministic, not "> 0")
18. Improve `TestWatchFilteredDirectories_Disabled` to use fake backend and verify
    `Add` was NOT called (instead of relying on WatchCount timing)
19. Add `WithContentHashMaxSize(bytes)` configurable option
20. Add `WithErrorBufferSize(int)` to decouple error channel from event channel
21. Add runtime deprecation warning for `MiddlewareWriteFileLog`
22. Document `WatcherError.Stack` behavior in doc comment
23. Add `Stats.WatchBudgetCap` field showing computed cap after safety fraction
24. Fuzz test `FilterIgnoreDirsCaseInsensitive` and `MiddlewareDeduplicateCaseInsensitive`
25. Integration test: polling mode + exclusions + gitignore end-to-end
26. Integration test: DropOnFull with slow consumer simulation (deterministic)
27. Add examples for `WithEventChannelMode(DropOnFull)`
28. Add example for `WithMaxWatchesSafetyFraction`
29. Add example for `MiddlewareDeduplicateCaseInsensitive`
30. Update `docs/guides/resilience.md` with new self-heal error classification
31. Update `docs/guides/middleware.md` with new dedup variants and batch error handling
32. Review all new option doc comments for consistency
33. Consider `EventsDroppedByBackpressure` Prometheus metric (currently only
    `events_dropped_by_middleware_total` and `errors_dropped_total` are exposed)
34. Consider splitting `maxWatches` into `maxWatchesConfigured` + `maxWatchesEffective`
35. Consider `MiddlewareBatchFlushErrorHandler` callback option
36. Consider splitting `Stats` struct into sub-structs
37. Consider extracting `channelSender` type from `emitEvent`
38. Add migration guide for v2.3 → v2.4

### Feature backlog (from feedback, not started)

39. Implement `WithPollDeduplicate(true)` — per-path native event timestamp tracking
40. Implement `WithPollDetectRenames(true)` — inode-based rename detection
41. Add macOS CI matrix to `ci.yml`
42. Add Windows CI matrix to `ci.yml`
43. Consider `MiddlewareDropCallback` API for explicit drop notification
44. Integrate `FilterIgnoreDirsCaseInsensitive` into `WithIgnoreDirs` auto-detection

### Documentation and cleanup

45. Update `FEATURES.md` — verify all new features are listed (may have missed some)
46. Update `README.md` if any user-facing behavior changed
47. Add guide for polling mode best practices (when to use, limitations)
48. Add guide for middleware observability (interpreting drop counters)
49. Archive feedback document to `docs/feedback/done/` after all items closed
50. Review all existing examples for compatibility with new path validation

---

## G) Questions I Cannot Answer Myself

### 1. Should I split `maxWatches` into `maxWatchesConfigured` and `maxWatchesEffective`?

The `maxWatchesExplicit` bool flag works but is a workaround. The real problem
is that `maxWatches` serves double duty (user intent vs. effective limit). Splitting
them would be cleaner but changes the internal API (any code reading
`w.maxWatches` to check the limit would need updating). Is this worth doing for
v2.4.0, or should I keep the bool flag and defer the refactor to v3?

### 2. Should the `eventsProcessed` semantics change be called out as breaking in the CHANGELOG?

I documented it under "Changed" but some consumers may have dashboards or alerts
keyed on `EventsProcessed` values. The old semantics counted events where
middleware returned nil (even if the event was never sent to the channel). The
new semantics count events that actually reached the channel. This means
`EventsProcessed` will be lower for consumers using rate-limiting or dedup
middleware. Should I add a prominent migration note, or is "Changed" sufficient?

### 3. Should I archive the feedback document to `docs/feedback/done/` now?

Items 2, 3, 20, 21, and 22 from the feedback document are not implemented. The
feedback document itself suggests some of these are minor or optional. Should I
archive it as "done" (with a note about deferred items), keep it in `new/` until
every item is closed, or split it — archive the implemented items and create a
new feedback doc for the deferred ones?

---

## Metrics Summary

| Metric | Prior Session End | This Session End | Delta |
|--------|-------------------|------------------|-------|
| Coverage | 80.1% | 82.2% | +2.1% |
| Tests added | 15+ | +7 | +7 |
| Lint issues | 0 | 0 | — |
| Test failures | 0 | 0 | — |
| EmitEvent allocs (NoDebounce) | 8 | 7 | -1 |
| Production bugs found | 0 | 2 | +2 (both fixed) |
| CHANGELOG entries | 0 | 24 | +24 |
| Uncommitted files | — | 4 modified + 1 new | — |

---

_All CI gates pass. Two production bugs found and fixed. Coverage improved.
Ready for review._
