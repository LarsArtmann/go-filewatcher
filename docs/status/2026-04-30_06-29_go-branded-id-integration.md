# Comprehensive Status Report: go-branded-id Integration

**Date:** 2026-04-30 06:29  
**Branch:** master  
**Commit base:** `0199ea7` (refactor(tests): extract shared test helper functions for DRY principle)

---

## a) FULLY DONE

### go-branded-id Integration (18 files changed, +316 -90 lines)

| Item                    | Status | Detail                                                                          |
| ----------------------- | ------ | ------------------------------------------------------------------------------- |
| Dependency added        | DONE   | `github.com/larsartmann/go-branded-id` in go.mod                                |
| Phantom types rewritten | DONE   | All 6 types now use `id.ID[Brand, string]` wrapper                              |
| Brand `Name()` methods  | DONE   | All 6 brands have `Name()` for debugging                                        |
| Source files updated    | DONE   | `watcher.go`, `watcher_internal.go`, `watcher_walk.go`, `event.go`, `errors.go` |
| Test files updated      | DONE   | All 7 test files updated with `New*()` constructors                             |
| Linter config updated   | DONE   | `.golangci.yml` depguard allows new dependency                                  |
| AGENTS.md updated       | DONE   | Dependencies section + branded types docs + known issues                        |
| Lint passes             | DONE   | 0 issues from golangci-lint                                                     |
| Format passes           | DONE   | go fmt + golines clean                                                          |
| Tests pass              | DONE   | All non-flaky tests pass with `-race`                                           |

### What the integration provides:

**Before (type aliases — mixable):**

```go
type EventPath string
type RootPath string
// EventPath("foo") assignable to RootPath("foo") — NO compile error
```

**After (branded types — compile-time distinct):**

```go
type EventPath struct { id id.ID[EventPathBrand, string] }
type RootPath struct { id id.ID[RootPathBrand, string] }
// Cannot mix — compiler catches type mismatches
```

### Branded types implemented:

| Type           | Brand               | Methods exposed                                           | Used in                               |
| -------------- | ------------------- | --------------------------------------------------------- | ------------------------------------- |
| `EventPath`    | `EventPathBrand`    | Get, IsZero, String, Equal, Compare, Base, Dir, Ext, Join | `event.go`, filters                   |
| `RootPath`     | `RootPathBrand`     | Get, IsZero, String, Equal                                | `watcher.go`, `watcher_walk.go`       |
| `DebounceKey`  | `DebounceKeyBrand`  | Get, IsZero, String, Equal                                | `debouncer.go`, `watcher_internal.go` |
| `LogSubstring` | `LogSubstringBrand` | Get, String                                               | `testing_helpers.go`                  |
| `TempDir`      | `TempDirBrand`      | Get, String                                               | `testing_helpers.go`                  |
| `OpString`     | `OpStringBrand`     | Get, String                                               | `errors.go`                           |

---

## b) PARTIALLY DONE

| Item                         | What's missing                                                                                                                                                                   |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| EventPath serialization      | The underlying `id.ID` has `MarshalJSON`, `UnmarshalJSON`, `MarshalText`, `UnmarshalText`, `MarshalBinary`, `Scan`, `Value` — we could expose these through the wrapper for free |
| RootPath/DebounceKey Compare | Underlying id supports `Compare()`, we only exposed it on EventPath                                                                                                              |
| `Or()` method                | go-branded-id provides `Or()` for default values — not exposed on any wrapper                                                                                                    |
| `Reset()` method             | go-branded-id provides `Reset()` — not exposed on any wrapper                                                                                                                    |

---

## c) NOT STARTED

| #   | Item                                                                                                      | Impact | Effort |
| --- | --------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 1   | Expose serialization methods (JSON/Text/Binary/SQL) on branded types                                      | High   | Low    |
| 2   | Add `Or()` and `Reset()` to wrappers where useful                                                         | Medium | Low    |
| 3   | Compile-time type safety tests (verify types can't be mixed)                                              | High   | Low    |
| 4   | Fix the 2 pre-existing flaky tests                                                                        | High   | Medium |
| 5   | Consider removing test-only brands (LogSubstring, TempDir) — they add complexity for marginal safety gain | Medium | Low    |
| 6   | Evaluate if OpString wrapper adds value over plain string in WatcherError                                 | Low    | Low    |

---

## d) TOTALLY FUCKED UP / PRE-EXISTING ISSUES

### Pre-existing Flaky Tests (NOT caused by our changes)

Both confirmed to fail identically on `master` without our changes:

| Test                               | Root cause                                                          | Impact                  |
| ---------------------------------- | ------------------------------------------------------------------- | ----------------------- |
| `TestWatcher_Stats_Metrics`        | Filesystem write coalescing produces 2 events instead of expected 1 | Intermittent CI failure |
| `TestWatcher_Watch_WithMiddleware` | Same timing issue — middleware called 2x instead of 1x              | Intermittent CI failure |

**These need a design fix:** The tests assume a single file write produces exactly one event, but filesystems often coalesce/duplicate events. Should use `assert.Eventually` or drain-and-count pattern.

### Design Debt from Our Changes

1. **Wrapper boilerplate is massive** — `phantom_types.go` went from 63 lines to 238 lines. Each type repeats Get/IsZero/String/Equal. Could be code-generated or use a generic base.
2. **LogSubstring/TempDir brands are test-only** — Adding branded types for test helpers is arguably over-engineering. These never cross API boundaries.
3. **No `comparable` constraint on wrapper types** — The underlying `id.ID` is `comparable` but our wrappers are too (struct with comparable field). However, we don't enforce or advertise this.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Reduce wrapper boilerplate** — Each branded type repeats 4-6 methods that just delegate to `id`. Consider a generic helper or code generation.
2. **Test-only brands should stay simple** — LogSubstring and TempDir don't benefit from branded types since they never cross package boundaries. Revert those to simple type aliases.
3. **Fix flaky tests** — These will cause CI failures forever if not addressed.

### Type Model Improvements

4. **Event.Path should become EventPath** — Currently `Event.Path` is still `string`. If we're serious about type safety, it should be `EventPath`. This is a breaking API change.
5. **WatcherError.Op could stay string** — OpString adds complexity for marginal benefit. It's only used internally for error messages.
6. **Consider if DebounceKey should wrap EventPath** — Semantically, debounce keys ARE file paths. Could use `EventPath` directly instead of a separate brand.

### Missing Leverage from go-branded-id

7. **Serialization delegation** — The library gives us JSON/Text/Binary/SQL for free. We should expose it.
8. **`Or()` for defaults** — Useful for EventPath (empty path → default path).
9. **`Ptr()` / `FromPtr()`** — Useful for optional path fields.

---

## f) Top 25 Things to Get Done Next

Sorted by impact × effort (highest first):

| #   | Task                                                                            | Impact | Effort | Category       |
| --- | ------------------------------------------------------------------------------- | ------ | ------ | -------------- |
| 1   | **Fix 2 flaky tests** (Stats_Metrics, WithMiddleware)                           | High   | Medium | Bug fix        |
| 2   | **Revert LogSubstring/TempDir to simple type aliases**                          | Medium | Low    | Simplification |
| 3   | **Add compile-time type safety test** (verify types can't mix)                  | High   | Low    | Testing        |
| 4   | **Expose JSON/Text serialization on EventPath**                                 | High   | Low    | Feature        |
| 5   | **Expose SQL Scan/Value on EventPath**                                          | High   | Low    | Feature        |
| 6   | **Expose `Or()` on EventPath**                                                  | Medium | Low    | Feature        |
| 7   | **Consider removing OpString wrapper** (use plain string)                       | Medium | Low    | Simplification |
| 8   | **Update file organization table in AGENTS.md** to include `phantom_types.go`   | Low    | Low    | Docs           |
| 9   | **Add `Compare()` to RootPath**                                                 | Low    | Low    | Feature        |
| 10  | **Add `Compare()` to DebounceKey**                                              | Low    | Low    | Feature        |
| 11  | **Evaluate: should DebounceKey just be EventPath?**                             | Medium | Low    | Architecture   |
| 12  | **Evaluate: should Event.Path become EventPath?** (breaking)                    | High   | High   | Architecture   |
| 13  | **Reduce phantom_types.go boilerplate** with generic helper                     | Medium | Medium | Refactor       |
| 14  | **Expose `Reset()` on EventPath**                                               | Low    | Low    | Feature        |
| 15  | **Expose `Ptr()`/`FromPtr()` on EventPath**                                     | Low    | Low    | Feature        |
| 16  | **Add benchmarks for branded type operations**                                  | Low    | Low    | Testing        |
| 17  | **Document the wrapper pattern in phantom_types.go**                            | Low    | Low    | Docs           |
| 18  | **Consider exposing branded types to external consumers**                       | Medium | Medium | API            |
| 19  | **Evaluate go-branded-id for use in other projects** (go-project-meta etc.)     | Medium | Low    | Cross-project  |
| 20  | **Add example_test.go for branded types**                                       | Low    | Low    | Docs           |
| 21  | **Expose Binary serialization on EventPath**                                    | Low    | Low    | Feature        |
| 22  | **Add `IsZero()` check to Event.GetPath()** — return zero value for empty paths | Low    | Low    | Feature        |
| 23  | **Investigate if `fmt.Stringer` interface is sufficient for all logging**       | Low    | Low    | Architecture   |
| 24  | **Consider if filters should accept EventPath instead of string paths**         | Medium | High   | Architecture   |
| 25  | **Add CHANGELOG.md entry for go-branded-id integration**                        | Low    | Low    | Docs           |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should Event.Path remain `string` or become `EventPath`?**

This is the single most impactful architectural decision:

- **Keeping `string`:** No breaking change. EventPath is only available via `GetPath()`. Users who want type safety opt in.
- **Changing to `EventPath`:** True end-to-end type safety. But it's a breaking API change — every consumer must update. It also means JSON deserialization becomes more complex (EventPath wrapper needs UnmarshalJSON).

This is a product/UX decision that only you can make. The technical implementation is straightforward either way.

---

## Test Results Summary

```
Lint:  0 issues
Tests: ALL PASS (excluding 2 pre-existing flaky tests)
  - Flaky: TestWatcher_Stats_Metrics (pre-existing)
  - Flaky: TestWatcher_Watch_WithMiddleware (pre-existing)
  - All other tests: PASS with -race flag
```
