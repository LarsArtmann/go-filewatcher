# Comprehensive Status Report — Post-Audit Fix Cycle

**Date:** 2026-05-03 02:30
**Session:** Self-audit fixes + type simplification + test coverage
**Commits this session:** 14 (d2e0d99..685f26c)

---

## a) FULLY DONE

| #  | Item                                                                                        | Commit  |
| -- | ------------------------------------------------------------------------------------------- | ------- |
| 1  | Relicense to MIT (LICENSE, README, flake.nix)                                               | d2e0d99 |
| 2  | Comprehensive 90-task execution plan                                                        | 0eb802a |
| 3  | Fix 5 critical issues (double-append, error swallowing, toolchain, test pollution)          | e0fe750 |
| 4  | Code quality improvements across 12 areas                                                   | a1a86cc |
| 5  | 15 new coverage tests (rename, multi-dir, non-recursive, concurrent, buffer, state, errors) | 87031f3 |
| 6  | CHANGELOG with v0.1.0/v0.2.0/[Unreleased], CONTRIBUTING.md, AGENTS.md updates               | 6607da3 |
| 7  | CI benchmark job + Dependabot config                                                        | c1f9cb4 |
| 8  | WatchOnce, MiddlewareThrottle, FilterIgnoreGlobs, WithIgnorePatterns                        | 1602096 |
| 9  | Fix WatchOnce nil-error wrapping                                                            | f750a73 |
| 10 | Replace 3 rate limiters with golang.org/x/time/rate                                         | a53bdfe |
| 11 | Add golang.org/x/time to go.mod + depguard                                                  | 5e77799 |
| 12 | Simplify phantom types (remove go-branded-id)                                               | a4d6f4c |
| 13 | Tests for WatchOnce, MiddlewareThrottle, FilterIgnoreGlobs, WithIgnorePatterns              | dede9c3 |
| 14 | TODO_LIST.md update (20+ items marked done)                                                 | 685f26c |

## b) PARTIALLY DONE

| Item                    | Status                                    | Remaining                                                                 |
| ----------------------- | ----------------------------------------- | ------------------------------------------------------------------------- |
| Testing helpers cleanup | `testing_helpers.go` → `_test.go` renamed | `testing_helpers.go` still exists (empty/ghost file warning from gopls)   |
| LSP diagnostic warnings | 16 warnings remain                        | All in test files (noinlineerr, modernize, gosec, goconst) — non-blocking |

## c) NOT STARTED

| #  | Item                                           | Priority |
| -- | ---------------------------------------------- | -------- |
~~| 1  | `WithPolling(fallback)` for NFS/network mounts | Medium   |~~ 114
~~| 2  | Exponential backoff for errors                 | Medium   |~~ 114
~~| 3  | Symlink following support                      | Medium   |~~ 114
~~| 4  | `Event.ModTime()` field                        | Medium   |~~ 114
~~| 5  | File content hashing option                    | Medium   |~~ 114
~~| 6  | Prometheus metrics export                      | Medium   |~~ 114
~~| 7  | OpenTelemetry integration                      | Medium   |~~ 114
~~| 8  | Self-healing watcher                           | Medium   |~~ 114
~~| 9  | Circuit breaker middleware                     | Medium   |~~ 114
~~| 10 | Goreleaser configuration                       | Medium   |~~ 114
~~| 11 | CLI tool                                       | Medium   |~~ 114
~~| 12 | Fuzz testing                                   | Backlog  |~~ 114

## d) TOTALLY FUCKED UP

| Item                              | What happened                                                          | Resolution                                   |
| --------------------------------- | ---------------------------------------------------------------------- | -------------------------------------------- |
| `golang.org/x/time` not in go.mod | Previous session committed middleware rewrite but forgot `go mod tidy` | Fixed in 5e77799                             |
| Depguard blocked new import       | golang.org/x/time wasn't in allow-list                                 | Fixed in 5e77799                             |
| `testing_helpers.go` ghost file   | Old file still exists alongside `_test.go` version                     | Lint passes, but gopls warns — needs cleanup |

## e) WHAT WE SHOULD IMPROVE

~~1. **Ghost `testing_helpers.go`** — Delete it (it's been renamed to `_test.go`)~~ 114
~~2. **LSP warnings in tests** — Fix noinlineerr, modernize, gosec warnings in `watcher_coverage_test.go`~~ 114
~~3. **Flaky tests** — `TestWatcher_Stats_Metrics` and `TestWatcher_Watch_WithMiddleware` remain timing-sensitive~~ 114
~~4. **Watcher struct size** — 24 fields is large; reconsider sub-structs if the library grows~~ 114
~~5. **Event.Path is `string`** — Still not using `EventPath` phantom type internally; filters and middleware all use raw strings~~ 114
~~6. **Pre-commit hook not executable** — Every commit warns about this~~ 114
~~7. **Version tagging** — v0.1.0 and v2.0.0 releases still not tagged~~ 114

## f) Top 25 Things to Do Next

| #  | Item                                                          | Impact | Effort |
| -- | ------------------------------------------------------------- | ------ | ------ |
~~| 1  | Tag v0.2.0 release                                            | High   | Low    |~~ 114
~~| 2  | Delete ghost `testing_helpers.go`                             | Medium | Low    |~~ 114
~~| 3  | Fix pre-commit hook permissions                               | Medium | Low    |~~ 114
~~| 4  | Fix LSP warnings in test files                                | Medium | Low    |~~ 114
~~| 5  | Add `WithPolling(fallback bool)` for NFS                      | High   | Medium |~~ 114
~~| 6  | Add exponential backoff for errors                            | High   | Medium |~~ 114
~~| 7  | Add `Event.ModTime` field                                     | Medium | Low    |~~ 114
~~| 8  | Add self-healing watcher (re-add lost paths)                  | High   | High   |~~ 114
~~| 9  | Add symlink following support                                 | Medium | Medium |~~ 114
~~| 10 | Add circuit breaker middleware                                | Medium | Medium |~~ 114
~~| 11 | Prometheus/OpenTelemetry integration                          | High   | Medium |~~ 114
~~| 12 | Goreleaser configuration                                      | Medium | Medium |~~ 114
~~| 13 | Standalone CLI tool                                           | High   | High   |~~ 114
~~| 14 | Address flaky tests (Stats, Middleware)                       | Medium | Medium |~~ 114
~~| 15 | Fuzz testing for event parsing                                | Medium | Medium |~~ 114
~~| 16 | Windows-specific edge case tests                              | Medium | Medium |~~ 114
~~| 17 | File content hashing option                                   | Medium | Medium |~~ 114
~~| 18 | Document DI integration patterns                              | Low    | Low    |~~ 114
~~| 19 | Add `CODE_OF_CONDUCT.md`                                      | Low    | Low    |~~ 114
~~| 20 | Add PR template                                               | Low    | Low    |~~ 114
~~| 21 | Write Troubleshooting.md                                      | Medium | Medium |~~ 114
~~| 22 | Add `FilterGeneratedCodeFull` test for edge cases             | Medium | Low    |~~ 114
~~| 23 | Consider `WatchChanges(ctx, targetState)` for idempotent sync | Medium | High   |~~ 114
~~| 24 | Dead letter queue for dropped events                          | Medium | Medium |~~ 114
~~| 25 | Error correlation IDs                                         | Low    | Medium |~~ 114

## g) Top Question I Cannot Figure Out Myself

**Should `Event.Path` be changed from `string` to `EventPath` phantom type?**

This is a **breaking API change** that would affect:

- All consumers of the library (every `event.Path` reference)
- All filter functions (every `Filter func(Event) bool` callback)
- All middleware functions
- JSON marshaling/unmarshaling
- Example code

The `EventPath` type exists and has domain methods (Base, Dir, Ext, Join), but it's currently only used via `Event.GetPath()`. Changing `Path` field to `EventPath` would require `.Get()` calls everywhere or adding `String() string` for implicit conversion.

**Recommendation:** Defer to v2.0.0 breaking change. Current `GetPath()` provides an opt-in path.

---

## Metrics

| Metric               | Value                         |
| -------------------- | ----------------------------- |
| Commits this session | 14                            |
| Files modified       | 25+                           |
| Lines removed (net)  | ~200                          |
| Dependencies added   | 1 (golang.org/x/time)         |
| Dependencies removed | 1 (go-branded-id)             |
| New tests added      | 21 (15 coverage + 6 targeted) |
| Linter issues        | 0                             |
| Build status         | Clean                         |
| Test status          | 100% passing with -race       |
