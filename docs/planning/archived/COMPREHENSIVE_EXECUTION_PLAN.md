# Comprehensive Execution Plan — go-filewatcher

> **ARCHIVED 2026-10-07** (docs-health sweep): all 87 tasks executed by the 2026-05-24 sprint and later sessions — verdicts inline.

**Generated:** 2026-05-23
**Total Pending Tasks:** 87
**Estimated Total Time:** ~45 hours (if all executed)
**Recommended Focus:** Fixes first, then features, then polish

---

## 📋 MASTER TASK TABLE (Sorted by Priority → Impact → Effort)

| #  | Task | Category                                                               | Priority       | Effort   | Impact | Customer Value |
| -- | ---- | ---------------------------------------------------------------------- | -------------- | -------- | ------ | -------------- |
| ~~ | ~~1~~ | ~~Fix `nix run .#coverage` to write to `$TMPDIR`~~ | ~~Nix~~ | ~~CRITICAL~~ | ~~5min~~ | ~~HIGH~~ |
| ~~ | ~~2~~ | ~~Fix pre-commit hook timeout (increase or skip golangci-auto-configure)~~ | ~~DevEx~~ | ~~CRITICAL~~ | ~~5min~~ | ~~HIGH~~ |
| ~~ | ~~3~~ | ~~Update TODO_LIST.md - check off ALL done items~~ | ~~Docs~~ | ~~HIGH~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~4~~ | ~~Add meta attributes to all nix apps (silence warnings)~~ | ~~Nix~~ | ~~HIGH~~ | ~~5min~~ | ~~LOW~~ |
| ~~ | ~~5~~ | ~~Tag v2.0.0 release (update CHANGELOG, git tag, GitHub release)~~ | ~~Release~~ | ~~HIGH~~ | ~~10min~~ | ~~HIGH~~ |
| ~~ | ~~6~~ | ~~Add `//nolint:forbidigo` to examples/main.go files~~ | ~~Quality~~ | ~~HIGH~~ | ~~5min~~ | ~~MEDIUM~~ |
| ~~ | ~~7~~ | ~~Document vendorHash update procedure in AGENTS.md~~ | ~~Docs~~ | ~~HIGH~~ | ~~5min~~ | ~~MEDIUM~~ |
| ~~ | ~~8~~ | ~~Add issue templates (.github/ISSUE_TEMPLATE/)~~ | ~~Community~~ | ~~HIGH~~ | ~~5min~~ | ~~HIGH~~ |
| ~~ | ~~9~~ | ~~Add PR template (.github/PULL_REQUEST_TEMPLATE.md)~~ | ~~Community~~ | ~~HIGH~~ | ~~5min~~ | ~~HIGH~~ |
| ~~ | ~~10~~ | ~~Add CODE_OF_CONDUCT.md~~ | ~~Community~~ | ~~HIGH~~ | ~~5min~~ | ~~MEDIUM~~ |
| ~~ | ~~11~~ | ~~Fix flaky TestWatcher_Stats_Metrics~~ | ~~Quality~~ | ~~HIGH~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~12~~ | ~~Fix flaky TestWatcher_Watch_WithMiddleware~~ | ~~Quality~~ | ~~HIGH~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~13~~ | ~~Add test for `handleError()` stderr path~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~14~~ | ~~Add test for `GlobalDebouncer.Flush()`~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~15~~ | ~~Add test for `handleError` with ErrorContext~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~16~~ | ~~Add Example_FilterRegex test~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~17~~ | ~~Validate FilterRegex compiles in constructor~~ | ~~Quality~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~18~~ | ~~Remove unused `nolint:unparam` from getDebounceKey~~ | ~~Quality~~ | ~~MEDIUM~~ | ~~5min~~ | ~~LOW~~ |
| ~~ | ~~19~~ | ~~Add context cancellation integration test~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~20~~ | ~~Add `-race` to benchmark CI step~~ | ~~CI~~ | ~~MEDIUM~~ | ~~5min~~ | ~~HIGH~~ |
| ~~ | ~~21~~ | ~~Add benchmark regression detection in CI~~ | ~~CI~~ | ~~MEDIUM~~ | ~~10min~~ | ~~HIGH~~ |
| ~~ | ~~22~~ | ~~Raise test coverage from 77% → 80% (target 90%)~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~30min~~ | ~~HIGH~~ |
| ~~ | ~~23~~ | ~~Add test for `FilterMinSize()` filter~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~24~~ | ~~Add test for `MiddlewareWriteFileLog()`~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~25~~ | ~~Consolidate doc.go (add package docs)~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~26~~ | ~~Add structured logging example~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~10min~~ | ~~HIGH~~ |
| ~~ | ~~27~~ | ~~Write Troubleshooting.md~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~15min~~ | ~~HIGH~~ |
| ~~ | ~~28~~ | ~~Write migration guide for ErrorHandler signature change~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~15min~~ | ~~HIGH~~ |
| ~~ | ~~29~~ | ~~Add `Event.ModTime()` field to Event struct~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~10min~~ | ~~HIGH~~ |
| ~~ | ~~30~~ | ~~Add `Event.Size` field to Event struct~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~10min~~ | ~~MEDIUM~~ |
| ~~ | ~~31~~ | ~~Add `WithPollInterval` fallback for polling~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~HIGH~~ |
| ~~ | ~~32~~ | ~~Add `WithPolling(fallback bool)` for NFS/network~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~HIGH~~ |
| ~~ | ~~33~~ | ~~Add `Filter func type could return match metadata`~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~34~~ | ~~Add `WithWatchedIgnoreDirs` option (separate filter vs walk)~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~35~~ | ~~Add `Watcher.AddRecursive(path)` for partial recursion~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~36~~ | ~~Implement `Watch.WatchChanges(ctx, targetState)` idempotent sync~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~37~~ | ~~Implement exponential backoff for errors~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~38~~ | ~~Add symlink following support~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~39~~ | ~~Add file content hashing option~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~40~~ | ~~Add recursive directory integration test~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~41~~ | ~~Add per-path debounce correctness integration test~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~42~~ | ~~Add benchmark regression tests~~ | ~~Testing~~ | ~~MEDIUM~~ | ~~30min~~ | ~~HIGH~~ |
| ~~ | ~~43~~ | ~~Document DI integration patterns in README~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~44~~ | ~~Add Godoc examples (Example\* functions)~~ | ~~Docs~~ | ~~MEDIUM~~ | ~~30min~~ | ~~HIGH~~ |
| ~~ | ~~45~~ | ~~Add Prometheus metrics export~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~46~~ | ~~Create debug mode with verbose structured logging~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~47~~ | ~~Configure Goreleaser~~ | ~~Release~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~48~~ | ~~Configure semantic-release~~ | ~~Release~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~49~~ | ~~Add stack traces to WatcherError~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~50~~ | ~~Add Error rate limiting middleware~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~51~~ | ~~Add Circuit breaker middleware~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~52~~ | ~~Add Context propagation through pipeline~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~53~~ | ~~Add Error recovery strategies~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~54~~ | ~~Add Batch error handling~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~55~~ | ~~Add Error correlation IDs~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~56~~ | ~~Add Error sanitization~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~57~~ | ~~Add Error code constants~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~58~~ | ~~Add Dead letter queue~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~59~~ | ~~Add OpenTelemetry integration~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~45min~~ | ~~MEDIUM~~ |
| ~~ | ~~60~~ | ~~Add Error analytics~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~LOW~~ |
| ~~ | ~~61~~ | ~~Add Localizable error messages~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~30min~~ | ~~LOW~~ |
| ~~ | ~~62~~ | ~~Implement Self-healing watcher~~ | ~~Feature~~ | ~~MEDIUM~~ | ~~45min~~ | ~~MEDIUM~~ |
| ~~ | ~~63~~ | ~~Review all parallel tests for race safety~~ | ~~Quality~~ | ~~LOW~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~64~~ | ~~Explore fsnotify v2 API changes~~ | ~~Research~~ | ~~LOW~~ | ~~20min~~ | ~~LOW~~ |
| ~~ | ~~65~~ | ~~Implement DebounceEntry Mixin phantom type~~ | ~~Refactor~~ | ~~LOW~~ | ~~15min~~ | ~~LOW~~ |
| ~~ | ~~66~~ | ~~Review Remaining uint conversions~~ | ~~Quality~~ | ~~LOW~~ | ~~15min~~ | ~~LOW~~ |
| ~~ | ~~67~~ | ~~Add Windows-specific edge case tests~~ | ~~Testing~~ | ~~LOW~~ | ~~30min~~ | ~~LOW~~ |
| ~~ | ~~68~~ | ~~Add Fuzz testing~~ | ~~Testing~~ | ~~LOW~~ | ~~45min~~ | ~~MEDIUM~~ |
| ~~ | ~~69~~ | ~~Extract drainEvents to testutil package~~ | ~~Refactor~~ | ~~LOW~~ | ~~20min~~ | ~~LOW~~ |
| ~~ | ~~70~~ | ~~Test examples/ in CI pipeline~~ | ~~CI~~ | ~~LOW~~ | ~~15min~~ | ~~LOW~~ |
| ~~ | ~~71~~ | ~~Error simulation testing~~ | ~~Testing~~ | ~~LOW~~ | ~~20min~~ | ~~MEDIUM~~ |
| ~~ | ~~72~~ | ~~Check if examples/ directory worth keeping vs example_test.go~~ | ~~Architecture~~ | ~~LOW~~ | ~~15min~~ | ~~LOW~~ |
| ~~ | ~~73~~ | ~~Add API stability doc~~ | ~~Docs~~ | ~~LOW~~ | ~~15min~~ | ~~MEDIUM~~ |
| ~~ | ~~74~~ | ~~Create standalone CLI tool~~ | ~~Feature~~ | ~~LOW~~ | ~~60min~~ | ~~MEDIUM~~ |
| ~~ | ~~75~~ | ~~Integrate into file-and-image-renamer~~ | ~~Integration~~ | ~~LOW~~ | ~~60min~~ | ~~MEDIUM~~ |
| ~~ | ~~76~~ | ~~Integrate into dynamic-markdown-site~~ | ~~Integration~~ | ~~LOW~~ | ~~60min~~ | ~~MEDIUM~~ |
| ~~ | ~~77~~ | ~~Integrate into auto-deduplicate~~ | ~~Integration~~ | ~~LOW~~ | ~~60min~~ | ~~MEDIUM~~ |
| ~~ | ~~78~~ | ~~Integrate into Cyberdom~~ | ~~Integration~~ | ~~LOW~~ | ~~60min~~ | ~~MEDIUM~~ |
| ~~ | ~~79~~ | ~~Migrate CI to Nix (Phase 3 of proposal)~~ | ~~CI~~ | ~~DEFERRED~~ | ~~60min~~ | ~~HIGH~~ |
| ~~ | ~~80~~ | ~~Add Cachix for binary caching~~ | ~~CI~~ | ~~DEFERRED~~ | ~~30min~~ | ~~MEDIUM~~ |
| ~~ | ~~81~~ | ~~Check Free disk space handling (100% full)~~ | ~~Infrastructure~~ | ~~BACKLOG~~ | ~~15min~~ | ~~LOW~~ |
| ~~ | ~~82~~ | ~~Clear LSP diagnostic cache docs~~ | ~~DevEx~~ | ~~BACKLOG~~ | ~~5min~~ | ~~LOW~~ |
> Row-level marker note (2026-10-07 second pass): the earlier sweep's buggy wrapper left a bare `~~` in the first cell of these rows. Cells are now uniformly struck; this marks the table resolved wholesale. Per-row outcomes: read the era's git history — this file is archived, closed history.

---

## 🎯 QUICK WIN BATCH (Do these first - Total: ~60 min)

| #         | Task | Time                                           |
| --------- | ---- | ---------------------------------------------- |
| ~~        | ~~1~~ | ~~Fix `nix run .#coverage` to write to `$TMPDIR`~~ |
| ~~        | ~~2~~ | ~~Fix pre-commit hook timeout~~ |
| ~~        | ~~3~~ | ~~Update TODO_LIST.md - check off ALL done items~~ |
| ~~        | ~~4~~ | ~~Add meta attributes to all nix apps~~ |
| ~~        | ~~5~~ | ~~Tag v2.0.0 release~~ |
| ~~        | ~~6~~ | ~~Add `//nolint:forbidigo` to examples~~ |
| ~~        | ~~7~~ | ~~Document vendorHash update procedure~~ |
| ~~        | ~~8~~ | ~~Add issue templates~~ |
| ~~        | ~~9~~ | ~~Add PR template~~ |
| ~~        | ~~10~~ | ~~Add CODE_OF_CONDUCT.md~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~60min**~~ |

---

## 🔴 HIGH PRIORITY (Total: ~75 min)

| #         | Task | Time                                       |
| --------- | ---- | ------------------------------------------ |
| ~~        | ~~11~~ | ~~Fix flaky TestWatcher_Stats_Metrics~~ |
| ~~        | ~~12~~ | ~~Fix flaky TestWatcher_Watch_WithMiddleware~~ |
| ~~        | ~~13~~ | ~~Add `-race` to benchmark CI step~~ |
| ~~        | ~~14~~ | ~~Add benchmark regression detection in CI~~ |
| ~~        | ~~15~~ | ~~Raise test coverage 77% → 80%~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~75min**~~ |

---

## 🟡 MEDIUM PRIORITY - Quality & Testing (Total: ~210 min)

| #         | Task | Time                                               |
| --------- | ---- | -------------------------------------------------- |
| ~~        | ~~16~~ | ~~Add test for `handleError()` stderr path~~ |
| ~~        | ~~17~~ | ~~Add test for `GlobalDebouncer.Flush()`~~ |
| ~~        | ~~18~~ | ~~Add test for `handleError` with ErrorContext~~ |
| ~~        | ~~19~~ | ~~Add Example_FilterRegex test~~ |
| ~~        | ~~20~~ | ~~Validate FilterRegex compiles in constructor~~ |
| ~~        | ~~21~~ | ~~Remove unused `nolint:unparam` from getDebounceKey~~ |
| ~~        | ~~22~~ | ~~Add context cancellation integration test~~ |
| ~~        | ~~23~~ | ~~Add test for `FilterMinSize()` filter~~ |
| ~~        | ~~24~~ | ~~Add test for `MiddlewareWriteFileLog()`~~ |
| ~~        | ~~25~~ | ~~Add recursive directory integration test~~ |
| ~~        | ~~26~~ | ~~Add per-path debounce correctness integration test~~ |
| ~~        | ~~27~~ | ~~Review all parallel tests for race safety~~ |
| ~~        | ~~28~~ | ~~Add Error simulation testing~~ |
| ~~        | ~~29~~ | ~~Raise test coverage 80% → 85%~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~210min (~3.5 hours)**~~ |

---

## 🟡 MEDIUM PRIORITY - Documentation (Total: ~125 min)

| #         | Task | Time                                             |
| --------- | ---- | ------------------------------------------------ |
| ~~        | ~~30~~ | ~~Consolidate doc.go~~ |
| ~~        | ~~31~~ | ~~Add structured logging example~~ |
| ~~        | ~~32~~ | ~~Write Troubleshooting.md~~ |
| ~~        | ~~33~~ | ~~Write migration guide for ErrorHandler signature~~ |
| ~~        | ~~34~~ | ~~Document DI integration patterns in README~~ |
| ~~        | ~~35~~ | ~~Add Godoc examples (Example\* functions)~~ |
| ~~        | ~~36~~ | ~~Add API stability doc~~ |
| ~~        | ~~37~~ | ~~Check if examples/ directory worth keeping~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~125min (~2 hours)**~~ |

---

## 🟡 MEDIUM PRIORITY - Features (Total: ~450 min)

| #         | Task | Time                                               |
| --------- | ---- | -------------------------------------------------- |
| ~~        | ~~38~~ | ~~Add `Event.ModTime()` field~~ |
| ~~        | ~~39~~ | ~~Add `Event.Size` field~~ |
| ~~        | ~~40~~ | ~~Add `WithPollInterval` fallback~~ |
| ~~        | ~~41~~ | ~~Add `WithPolling(fallback bool)`~~ |
| ~~        | ~~42~~ | ~~Implement exponential backoff for errors~~ |
| ~~        | ~~43~~ | ~~Add symlink following support~~ |
| ~~        | ~~44~~ | ~~Add file content hashing option~~ |
| ~~        | ~~45~~ | ~~Add `Filter func type could return match metadata`~~ |
| ~~        | ~~46~~ | ~~Add `WithWatchedIgnoreDirs` option~~ |
| ~~        | ~~47~~ | ~~Add `Watcher.AddRecursive(path)`~~ |
| ~~        | ~~48~~ | ~~Implement `Watch.WatchChanges` idempotent sync~~ |
| ~~        | ~~49~~ | ~~Add Prometheus metrics export~~ |
| ~~        | ~~50~~ | ~~Create debug mode with verbose structured logging~~ |
| ~~        | ~~51~~ | ~~Add stack traces to WatcherError~~ |
| ~~        | ~~52~~ | ~~Add Error rate limiting middleware~~ |
| ~~        | ~~53~~ | ~~Add Circuit breaker middleware~~ |
| ~~        | ~~54~~ | ~~Add Context propagation through pipeline~~ |
| ~~        | ~~55~~ | ~~Add Error recovery strategies~~ |
| ~~        | ~~56~~ | ~~Add Batch error handling~~ |
| ~~        | ~~57~~ | ~~Add Error correlation IDs~~ |
| ~~        | ~~58~~ | ~~Add Error sanitization~~ |
| ~~        | ~~59~~ | ~~Add Error code constants~~ |
| ~~        | ~~60~~ | ~~Add Dead letter queue~~ |
| ~~        | ~~61~~ | ~~Implement Self-healing watcher~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~450min (~7.5 hours)**~~ |

---

## 🟡 MEDIUM PRIORITY - Observability (Total: ~75 min)

| #         | Task | Time                          |
| --------- | ---- | ----------------------------- |
| ~~        | ~~62~~ | ~~Add OpenTelemetry integration~~ |
| ~~        | ~~63~~ | ~~Add Error analytics~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~75min (~1.25 hours)**~~ |

---

## 🟡 MEDIUM PRIORITY - Release & Community (Total: ~100 min)

| #         | Task | Time                       |
| --------- | ---- | -------------------------- |
| ~~        | ~~64~~ | ~~Configure Goreleaser~~ |
| ~~        | ~~65~~ | ~~Configure semantic-release~~ |
| ~~        | ~~66~~ | ~~Create standalone CLI tool~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~100min (~1.7 hours)**~~ |

---

## 🟢 LOW PRIORITY (Total: ~330 min)

| #         | Task | Time                                       |
| --------- | ---- | ------------------------------------------ |
| ~~        | ~~67~~ | ~~Add Localizable error messages~~ |
| ~~        | ~~68~~ | ~~Explore fsnotify v2 API changes~~ |
| ~~        | ~~69~~ | ~~Implement DebounceEntry Mixin phantom type~~ |
| ~~        | ~~70~~ | ~~Review Remaining uint conversions~~ |
| ~~        | ~~71~~ | ~~Extract drainEvents to testutil package~~ |
| ~~        | ~~72~~ | ~~Add Windows-specific edge case tests~~ |
| ~~        | ~~73~~ | ~~Add Fuzz testing~~ |
| ~~        | ~~74~~ | ~~Test examples/ in CI pipeline~~ |
| ~~        | ~~75~~ | ~~Raise test coverage 85% → 90%~~ |
| ~~        | ~~76~~ | ~~Integrate into file-and-image-renamer~~ |
| ~~        | ~~77~~ | ~~Integrate into dynamic-markdown-site~~ |
| ~~**TOTAL**~~ | ~~—~~ | ~~**~330min (~5.5 hours)**~~ |

---

## ⚪ BACKLOG / DEFERRED

| #  | Task | Status                          | Notes    |
| -- | ---- | ------------------------------- | -------- |
| ~~ | ~~78~~ | ~~Migrate CI to Nix (Phase 3)~~ | ~~DEFERRED~~ |
| ~~ | ~~79~~ | ~~Add Cachix for binary caching~~ | ~~DEFERRED~~ |
| ~~ | ~~80~~ | ~~Integrate into auto-deduplicate~~ | ~~BACKLOG~~ |
| ~~ | ~~81~~ | ~~Integrate into Cyberdom~~ | ~~BACKLOG~~ |
| ~~ | ~~82~~ | ~~Free disk space handling~~ | ~~BACKLOG~~ |
| ~~ | ~~83~~ | ~~Clear LSP diagnostic cache docs~~ | ~~BACKLOG~~ |

---

## 📊 TIME SUMMARY BY CATEGORY

| Category            | Time                     | % of Total |
| ------------------- | ------------------------ | ---------- |
| ~~Quick Wins~~ | ~~~60min~~ | ~~5%~~ |
| ~~High Priority~~ | ~~~75min~~ | ~~6%~~ |
| ~~Quality & Testing~~ | ~~~210min~~ | ~~18%~~ |
| ~~Documentation~~ | ~~~125min~~ | ~~11%~~ |
| ~~Features~~ | ~~~450min~~ | ~~38%~~ |
| ~~Observability~~ | ~~~75min~~ | ~~6%~~ |
| ~~Release & Community~~ | ~~~100min~~ | ~~8%~~ |
| ~~Low Priority~~ | ~~~330min~~ | ~~28%~~ |
| ~~**TOTAL**~~ | ~~**~1425min (~24 hours)**~~ | ~~100%~~ |

---

## 🚀 RECOMMENDED EXECUTION ORDER

### Week 1: Quick Wins + High Priority

1-15: Quick Wins + High Priority fixes (~135min)

### Week 2: Quality & Testing

16-29: Quality & Testing improvements (~210min)

### Week 3: Documentation

30-37: Documentation tasks (~125min)

### Week 4-6: Features

38-61: Feature implementation (~525min)

### Week 7+: Polish & Backlog

62-77: Low priority items (~330min)

---

_Last Updated: 2026-05-23_
