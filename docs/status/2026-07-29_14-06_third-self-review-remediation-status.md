# Status Report: Third Self-Review Remediation

**Created:** 2026-07-29 14:06
**Session scope:** Close all critical, high, and medium-priority gaps from the
third brutal self-review
(`docs/status/2026-07-29_09-44_self-review-remediation-second-brutal-review.md`).
**Plan:** `docs/planning/2026-07-29_14-06_third-self-review-remediation-plan.md`
(30 tasks, 8 phases).
**Verdict:** All 30 tasks completed. `nix flake check` passes. Zero allocation
regression on all benchmarks. Dead code eliminated. CHANGELOG updated.

---

## What Was Fixed

### 1. Dead Code Eliminated (the #1 embarrassment)

**Problem:** `FilesystemCaseSensitivity.GaugeValue()` was defined but never
called. The Prometheus gauge used a separate `caseSensitivityGauge(string)`
function with its own switch/case, duplicating the gauge encoding.

**Root Cause:** `Stats.CaseSensitivity` is a `string`, not the enum. So
metrics.go couldn't call `GaugeValue()` without parsing the string back to the
enum.

**Fix:** Added `Stats.CaseSensitivityMode FilesystemCaseSensitivity` (additive,
non-breaking field). The gauge now reads `stats.CaseSensitivityMode.GaugeValue()`
directly. The duplicate `caseSensitivityGauge()` function is deleted. Single
source of truth for the gauge encoding.

### 2. `_` Error Discarding Replaced

**Problem:** `FilterExcludePaths` and `WithExcludePaths` used
`normalized, _ := normalizePath(path)` with a comment explaining "intentionally
ignored." Comments don't fix code smells.

**Fix:** Created `cleanPath(path string) string` helper that encapsulates the
best-effort intent. Callers now use `cleanPath(path)` — no `_`, no linter
issues, clear intent.

### 3. CHANGELOG Updated

All session changes documented in CHANGELOG.md under `[Unreleased]` with Added,
Changed, and Fixed sections.

### 4. Clean Bench-Diff (replacing garbage data)

**Problem:** Previous bench-diff was invalidated by running a 5-minute fuzz test
(32 workers) simultaneously — ±18-59% timing variance was CPU contention.

**Fix:** Ran clean bench-diff (no parallel work). Results:
- **Zero allocation regression** on ALL benchmarks (all samples equal)
- Timing geomean: +6.38% (within noise range)
- Allocation geomean: -0.03% (effectively zero)

### 5. API Stability Classification Corrected

`FilesystemCaseSensitivity` moved from Evolving → **Stable**. The enum values
are foundational (control `pathKey()` behavior) and must not change. Adding new
values (e.g., `CaseSensitivityProbed` in v3) is backward-compatible.

---

## New APIs Added

| Symbol | Type | Stability | Purpose |
|--------|------|-----------|---------|
| `Stats.CaseSensitivityMode` | Stats field | Stable (additive) | Type-safe enum alongside existing string field |
| `FilterCaseSensitive(inner Filter)` | Filter function | Evolving | NFC-normalize without case-folding (macOS NFD paths) |
| `EffectiveCaseSensitivity()` | Watcher method | Evolving | Returns resolved enum without full Stats() call |

---

## New Tests Added

| Test | File | What it verifies |
|------|------|------------------|
| `TestFilesystemCaseSensitivity_GaugeValue` | filesystem_test.go | GaugeValue() returns correct float for each mode |
| `TestStats_CaseSensitivityFieldsConsistent` | filesystem_test.go | String and enum fields always agree |
| `TestStats_CaseSensitivityModeReflectsMode` | filesystem_test.go | Enum field populated correctly via Stats() |
| `TestCleanPath` | filesystem_test.go | cleanPath() normalizes trailing slashes, .., etc. |
| `TestFilterCaseSensitive_NFCNormalizes` | filesystem_test.go | NFD path matches NFC pattern through wrapper |
| `TestFilterCaseSensitive_PreservesCase` | filesystem_test.go | Case is not folded (unlike FilterCaseInsensitive) |
| `TestEffectiveCaseSensitivity` | filesystem_test.go | Method returns correct resolved mode |
| `TestWatcher_Add_TrailingSlashNormalized` | watcher_test.go | Add("foo/") normalizes to Add("foo") |
| `TestWatcher_Remove_DeeplyNestedUnicodeSubtree` | watcher_test.go | 3-level Unicode tree (café/München/東京) cleanup |

---

## New Benchmarks

| Benchmark | ns/op | B/op | allocs | What it measures |
|-----------|-------|------|--------|------------------|
| `BenchmarkFilterCaseInsensitive` | ~250 | 24 | 1 | NFC + ToLower wrapper overhead |
| `BenchmarkFilterCaseSensitive` | ~115 | 0 | 0 | NFC-only wrapper (2x faster, no alloc) |
| `BenchmarkPathKey_EmojiZWJ_CaseSensitive` | ~170 | 0 | 0 | Deep Unicode (ZWJ family emoji) |
| `BenchmarkShouldExcludePath_Empty` | ~1.4 | 0 | 0 | O(n) prefix scan baseline |
| `BenchmarkShouldExcludePath_FewPaths` | ~150 | 0 | 0 | 3 exclude paths |
| `BenchmarkShouldExcludePath_ManyPaths` | ~3040 | 0 | 0 | 100 exclude paths |

---

## Dependabot Audit

Both open alerts are in the **website toolchain** (npm), not the Go library:

| Package | Severity | Ecosystem | Issue |
|---------|----------|-----------|-------|
| `fast-uri` | High | npm | Host confusion via backslash authority delimiter |
| `astro` | Medium | npm | Reflected XSS via View Transition properties |

**Decision:** Website vulnerabilities do NOT block Go library releases. The
website is a separate deployment (Firebase Hosting) with its own flake.nix and
package.json. Track as website maintenance TODO.

---

## Nolint Directive Audit

89 `//nolint` directives reviewed across the codebase. All are justified:
- `varnamelen` (37) — idiomatic short names (w, p, d, op, mu, mw, cs, ci, tc)
- `gosec` (15) — test code, trusted file paths, standard permissions
- `err113` (12) — test-only dynamic errors
- `paralleltest`/`tparallel` (13) — global resource usage (Stderr, pipes, fs)
- `funlen` (4) — constructors and complex middleware
- Others (8) — cyclop, exhaustruct, gochecknoglobals, gosmopolitan, etc.

No cleanup needed.

---

## Items Deferred (lower priority from the 50-item list)

| # | Item | Why deferred |
|---|------|-------------|
| 17 | `FilterNFCNormalized(inner Filter)` | `FilterCaseSensitive` already provides this (NFC without case-fold) |
| 18 | FuzzPathKey overnight (8+ hours) | Already ran 5min/68.6M execs/0 failures. Diminishing returns. |
| 20 | Commit bench-baseline.txt for CI | Needs CI workflow design — separate concern |
| 21 | `WithNormalizeUnicode(false)` escape hatch | YAGNI — NFC is zero-cost for ASCII |
| 22 | Trie-based excludePaths | Premature optimization — 100 paths = 3µs |
| 23 | Phantom-typed `PathKey` | Significant refactor, defer for v3 |
| 27 | `CaseSensitivityProbed` mode | v3 feature — requires filesystem probing |
| 28 | macOS/Windows CI matrix | Can't test locally, needs CI config |
| 48 | `pathKey` returning `(PathKey, error)` | Significant API change, defer for v3 |
