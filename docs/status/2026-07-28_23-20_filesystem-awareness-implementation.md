# Status Report: Filesystem Awareness Implementation

**Date:** 2026-07-28 23:20
**Session Goal:** Make go-filewatcher aware of underlying filesystem differences (case sensitivity, path normalization)

---

## Executive Summary

Implemented filesystem case-sensitivity awareness via `WithCaseSensitivity(mode)` with
auto-detection per platform. Fixed five real correctness bugs on case-insensitive filesystems
(Windows NTFS, macOS APFS). Added O(1) watched-path lookup as a bonus.

**However, the implementation is INCOMPLETE.** Two critical code paths were missed entirely:
the polling loop and the gitignore matcher. Both use raw string path comparisons that bypass
the new `pathKey()` normalization, leaving case-insensitivity bugs active in NFS/FUSE polling
mode and gitignore-aware walking on macOS/Windows.

---

## a) FULLY DONE

| Item | Details |
| --- | --- |
| `FilesystemCaseSensitivity` enum | `filesystem.go` — `CaseSensitivityAuto` (default), `CaseSensitive`, `CaseInsensitive` with `String()` method |
| `resolveCaseSensitivity()` | Auto → platform-aware: Windows/macOS = insensitive, Linux/BSD = sensitive |
| `pathKey()` method | Single canonical point for path comparison keys. Lowercases on case-insensitive, preserves on case-sensitive |
| `WithCaseSensitivity(mode)` option | `options.go` — public functional option, documented |
| Watch deduplication | `tryAddPath` now checks `watchListKeys` set (O(1)) instead of appending blindly — prevents duplicate watches for paths differing only in case |
| `Remove()` case-aware subtree matching | `watcher.go:519` — uses `pathKey` for prefix comparison so `Remove("/dir/MyFile")` matches `/dir/myfile` on NTFS |
| `shouldExcludePath()` case-aware | `watcher_walk.go:195` — all comparisons through `pathKey` |
| `excludePaths` normalization | Keys pre-normalized to `pathKey` form in `New()` for O(1) lookups |
| Debounce key case-awareness | `getDebounceKey()` uses `pathKey` — events for `/File.go` and `/file.go` coalesce on case-insensitive |
| Self-heal case-aware | `failedPaths` keyed by `pathKey` with original-case values; `isPathWatched` and `appendToWatchList` use `watchListKeys` |
| `appendToWatchList` dedup | Now self-deduplicates (was unconditional append before) |
| O(1) `watchListKeys` map | Replaces two O(n) `slices.Contains` calls with O(1) map lookups |
| Path normalization in `New()` | Extracted `validateAndNormalizePaths()` helper — paths stored as absolute from the start |
| `Reset()` re-resolves case sensitivity | `effectiveCaseSensitivity` re-computed on reset |
| `Close()` clears `watchListKeys` | No state leak across close/restart cycles |
| Tests: 13 new in `filesystem_test.go` | Covers enum String, resolve, pathKey (both modes), dedup, Remove, debounce, exclude — all pass with `-race` |
| `watcher_selfheal_test.go` updated | Fixed `failedPaths` type from `struct{}` to `string` |
| `FEATURES.md` updated | New rows for case-sensitivity and O(1) lookup |
| `AGENTS.md` updated | Gotchas #18 and #19, Key Patterns table, File Organization table |

---

## b) PARTIALLY DONE

| Item | What's Done | What's Missing |
| --- | --- | --- |
| **`pathKey()` coverage** | Applied to: `tryAddPath`, `Remove`, `shouldExcludePath`, `getDebounceKey`, self-heal, `watchListKeys` | **NOT** applied to: `watcher_poll.go` (2 functions), `watcher_gitignore.go` (1 function), `filter.go` (3+ functions) — see section (c) |
| **Documentation** | `AGENTS.md`, `FEATURES.md` updated | `CHANGELOG.md`, `doc.go`, `Troubleshooting.md`, `DOMAIN_LANGUAGE.md`, website — not touched |
| **Test coverage** | Unit tests for all new code paths | No example in `example_test.go`; no integration test that simulates case-insensitive filesystem behavior end-to-end |

---

## c) NOT STARTED

| # | Item | Impact | Priority |
| --- | --- | --- | --- |
| 1 | **Poll loop case-awareness** (`watcher_poll.go`) | `pollDetectChanges` uses raw path strings as map keys in `snapshot`/`current`. On case-insensitive FS, rename `File.go` → `file.go` produces false Remove+Create instead of no-op. Also `pollWalkDir` stores raw paths. | **CRITICAL** |
| 2 | **Gitignore matcher case-awareness** (`watcher_gitignore.go`) | `shouldSkipByGitignore` uses `strings.HasPrefix(path, prefix)` with raw strings. If gitignore dir and event path differ in case, prefix match fails, gitignore rules silently bypassed. | **CRITICAL** |
| 3 | **User-facing filters case-awareness** (`filter.go`) | `FilterIgnoreDirs`, `FilterExcludePaths`, `FilterGlob` compare paths directly. `FilterExtensions` already lowercases extensions, proving the pattern exists. | **HIGH** (debatable — user-facing API) |
| 4 | **CHANGELOG.md entry** | No release note for the new `WithCaseSensitivity` option | MEDIUM |
| 5 | **doc.go package doc** | No mention of case-sensitivity in package-level docs | MEDIUM |
| 6 | **example_test.go** | No runnable example for `WithCaseSensitivity` | LOW |
| 7 | **Troubleshooting.md** | No guidance for case-sensitivity issues | LOW |
| 8 | **DOMAIN_LANGUAGE.md** | No entries for `FilesystemCaseSensitivity`, `pathKey`, `CaseSensitive`/`CaseInsensitive` | LOW |
| 9 | **Website docs** | `website/src/content/docs/` not updated with case-sensitivity guide | LOW |

---

## d) TOTALLY FUCKED UP

**Nothing is totally fucked up.** All code compiles, all 0 lint issues, all tests pass with
`-race`. The architecture is sound — `pathKey()` is the right single point of normalization.

**But the coverage gaps in (c) items #1 and #2 are serious enough to call out as
"half-baked":** I added a case-sensitivity system but didn't wire it through the two subsystems
that need it most for the NFS/FUSE and macOS use cases that motivated the work. A user who
enables `WithCaseSensitivity(CaseInsensitive)` on macOS with `WithPolling(true)` (a very
common combo for Docker volumes on macOS!) would still get phantom events from case-only
renames in the poll loop.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture / Design

1. **`pathKey` should be a free function, not a method.** Currently `pathKey` is a `*Watcher`
   method, which means poll loop, gitignore, and filters can't easily call it without holding
   a watcher reference. The gitignore matcher and poll loop DO have `w *Watcher`, so this is
   not blocking, but making it a method on Watcher means the function signature implies it
   needs mutable state when it only reads one field.

2. **The zero-value problem.** `effectiveCaseSensitivity` defaults to `CaseSensitivityAuto`
   (0) if someone constructs a bare `&Watcher{}` without `New()`. `pathKey()` only lowercases
   when `== CaseInsensitive`, so on a bare struct it's case-sensitive by default. This is
   safe but inconsistent with the "auto resolves to platform default" contract. Not a
   production issue (everyone uses `New()`), but a trap for tests.

3. **No runtime detection of actual filesystem case-sensitivity.** `CaseSensitivityAuto`
   resolves by `runtime.GOOS`, but you can mount a case-sensitive filesystem on macOS
   (case-sensitive APFS) or a case-insensitive filesystem on Linux (FAT/exFAT mounts). True
   detection would require probing the filesystem or reading `statvfs`/`pathconf`. This is a
   v3 feature, not a bug, but worth noting.

4. **`watchListKeys` duplicates `watchList` as a parallel data structure.** Every add/remove
   must update both. This is a maintenance risk — any future code path that adds to
   `watchList` without updating `watchListKeys` creates silent inconsistency. Consider a
   helper method `addWatch(path)` / `removeWatch(path)` that updates both atomically.

### Testing

5. **No integration test for the actual scenario.** All tests use `withBackend(fb)` with
   fake backends. There's no test that creates files with different case on a real
   filesystem and verifies dedup. This is hard to test portably but at minimum the test
   should exist on Linux with `CaseInsensitive` mode forced.

6. **Tests create bare `&Watcher{}` structs.** `TestPathKey_*` and `TestGetDebounceKey_*`
   construct `&Watcher{effectiveCaseSensitivity: CaseSensitive}` directly. This bypasses
   `New()` initialization. If the struct layout changes, these tests won't catch missing
   initialization.

### Process

7. **I should have audited ALL path comparison sites before writing code.** I searched for
   `casefold`/`case-sensitive` keywords (found nothing), then implemented. A better approach
   would have been grepping for `strings.HasPrefix(path`, `strings.Contains(event.Path`,
   and map[ path keys to find every comparison site systematically.

---

## f) Up to 50 Things to Get Done Next

### Critical (case-sensitivity completion)

1. Wire `pathKey()` into `pollWalkDir` — use canonical keys in the snapshot map
2. Wire `pathKey()` into `pollDetectChanges` — compare canonical keys, not raw paths
3. Wire `pathKey()` into `shouldSkipByGitignore` — ancestor prefix check must be case-aware
4. Wire `pathKey()` into `gitignoreCache.load` — matchers keyed by canonical dir
5. Consider `FilterCaseInsensitive()` wrapper or built-in case-normalization in filters
6. Wire case-awareness into `FilterIgnoreDirs` (path component matching)
7. Wire case-awareness into `FilterExcludePaths` (exact path matching)
8. Wire case-awareness into `FilterGlob` (filename matching)

### Testing

9. Add integration test: create `File.go` and `file.go`, verify dedup with `CaseInsensitive`
10. Add integration test: polling mode + case-only rename, verify no phantom events
11. Add integration test: gitignore + case-insensitive, verify rules applied
12. Add test for `watchListKeys` consistency after `Add` + `Remove` cycles
13. Add fuzz test for `pathKey()` with unicode/edge-case paths
14. Add test for `Reset()` preserving `caseSensitivity` config
15. Add benchmark: `pathKey` overhead on large directory trees (case-sensitive mode)
16. Add test: `WithCaseSensitivity` called after `New()` but before `Watch()`

### Documentation

17. Add CHANGELOG.md entry for `WithCaseSensitivity`
18. Update `doc.go` package doc with case-sensitivity mention
19. Add example in `example_test.go` for `WithCaseSensitivity`
20. Update `Troubleshooting.md` with case-sensitivity guidance
21. Update `docs/DOMAIN_LANGUAGE.md` with case-sensitivity terms
22. Update website `guides/resilience.mdx` or create `guides/case-sensitivity.mdx`
23. Update website `data/features.ts` with case-sensitivity feature
24. Add API_STABILITY.md entry for `WithCaseSensitivity` (new additive option)

### Architecture / Refactoring

25. Extract `addWatch(path string)` / `removeWatch(path string)` helpers that update both
    `watchList` and `watchListKeys` atomically
26. Consider making `pathKey` a free function taking `FilesystemCaseSensitivity` as arg
27. Add `FilesystemCaseSensitivity` to `Stats()` struct for observability
28. Consider `WithCaseSensitivityDetect()` that probes the actual filesystem at runtime
29. Document the interaction between `WithFollowSymlinks` and case-sensitivity (symlink
    targets may be on a different filesystem with different case semantics)
30. Consider unicode normalization (NFC vs NFD) — macOS uses NFD, most others use NFC.
    This is a separate but related issue to case sensitivity.

### From TODO_LIST.md (pre-existing, unrelated)

31. Windows CI matrix job in `ci.yml`
32. Expand fuzz tests (FilterAnd/FilterOr/FilterNot composition)
33. Large-tree stress harness (100k directories)
34. Shrink docs-consistency exemption list
35. Resolve `.goreleaser.yml` dead-config open question
36. Resolve benchmark baseline committed-vs-gitignored question
37. Design `WatchChanges(ctx, targetState)` contract

### From ROADMAP.md (pre-existing, unrelated)

38. macOS FSEvents edge case testing
39. BSD/kqueue verification
40. v3 planning — accumulated deprecations
41. Streaming filter protocol (`(keep bool, err error)`)
42. pprof endpoints for watcher introspection
43. Zero-allocation event path investigation
44. Race-detector CI statistical gate
45. Cross-platform release artifacts via goreleaser
46. Dependency freshness SLO codification
47. Docs freshness gate auto-generation
48. Benchmark freshness CI
49. `FilterGeneratedCodeFull` documentation
50. `FilterGeneratedCodeWithFilter` documentation

---

## g) Questions (cannot figure out myself)

### Q1: Should user-facing filters (`FilterIgnoreDirs`, `FilterGlob`, `FilterExcludePaths`) automatically respect case-sensitivity, or should that be explicit?

**Context:** `FilterExtensions` already lowercases extensions, proving the pattern exists.
But `FilterIgnoreDirs("Vendor")` matching `/vendor/pkg` is a user expectation question.
If we auto-apply case-sensitivity, users on Linux might be surprised that their
case-sensitive filter silently becomes case-insensitive. Options:
- (A) Auto-apply: filters read the watcher's `effectiveCaseSensitivity` (requires passing
  watcher state into filters, which currently have no such access — `Filter` is just
  `func(Event) bool`).
- (B) Provide `FilterCaseInsensitive(inner Filter) Filter` wrapper for opt-in.
- (C) Leave filters case-sensitive always; document the limitation.

### Q2: Should we implement real filesystem probing for `CaseSensitivityAuto` instead of using `runtime.GOOS`?

**Context:** macOS can have case-sensitive APFS volumes. Linux can mount FAT/exFAT
(case-insensitive). Docker bind mounts on macOS go through a case-insensitive layer
(osxfs). `runtime.GOOS` gives the right answer 95% of the time but misses these cases.
Real probing would create a temp file with mixed case and test if both resolve to the same
inode. Is the added complexity and I/O worth it, or should this stay as a v3 feature?

### Q3: Should `Stats()` expose the effective case-sensitivity mode for observability/debugging?

**Context:** When debugging "why didn't Remove() work" or "why did I get duplicate events",
knowing the resolved case-sensitivity mode would help. Adding `CaseSensitivity string` to
`Stats()` is trivial and additive. But it adds API surface. Is this worth it, or is it
over-engineering for a setting most users will never change from auto?
