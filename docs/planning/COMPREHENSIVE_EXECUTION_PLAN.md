# Comprehensive Execution Plan — go-filewatcher

**Generated:** 2026-05-23
**Total Pending Tasks:** 87
**Estimated Total Time:** ~45 hours (if all executed)
**Recommended Focus:** Fixes first, then features, then polish

---

## 📋 MASTER TASK TABLE (Sorted by Priority → Impact → Effort)

| #   | Task                                                                   | Category       | Priority | Effort | Impact | Customer Value |
| --- | ---------------------------------------------------------------------- | -------------- | -------- | ------ | ------ | -------------- |
| 1   | Fix `nix run .#coverage` to write to `$TMPDIR`                         | Nix            | CRITICAL | 5min   | HIGH   | DevEx          |
| 2   | Fix pre-commit hook timeout (increase or skip golangci-auto-configure) | DevEx          | CRITICAL | 5min   | HIGH   | DevEx          |
| 3   | Update TODO_LIST.md - check off ALL done items                         | Docs           | HIGH     | 10min  | MEDIUM | Maintenance    |
| 4   | Add meta attributes to all nix apps (silence warnings)                 | Nix            | HIGH     | 5min   | LOW    | DevEx          |
| 5   | Tag v2.0.0 release (update CHANGELOG, git tag, GitHub release)         | Release        | HIGH     | 10min  | HIGH   | Users          |
| 6   | Add `//nolint:forbidigo` to examples/main.go files                     | Quality        | HIGH     | 5min   | MEDIUM | CI             |
| 7   | Document vendorHash update procedure in AGENTS.md                      | Docs           | HIGH     | 5min   | MEDIUM | DevEx          |
| 8   | Add issue templates (.github/ISSUE_TEMPLATE/)                          | Community      | HIGH     | 5min   | HIGH   | Community      |
| 9   | Add PR template (.github/PULL_REQUEST_TEMPLATE.md)                     | Community      | HIGH     | 5min   | HIGH   | Community      |
| 10  | Add CODE_OF_CONDUCT.md                                                 | Community      | HIGH     | 5min   | MEDIUM | Community      |
| 11  | Fix flaky TestWatcher_Stats_Metrics                                    | Quality        | HIGH     | 15min  | MEDIUM | Reliability    |
| 12  | Fix flaky TestWatcher_Watch_WithMiddleware                             | Quality        | HIGH     | 15min  | MEDIUM | Reliability    |
| 13  | Add test for `handleError()` stderr path                               | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 14  | Add test for `GlobalDebouncer.Flush()`                                 | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 15  | Add test for `handleError` with ErrorContext                           | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 16  | Add Example_FilterRegex test                                           | Testing        | MEDIUM   | 10min  | MEDIUM | Docs           |
| 17  | Validate FilterRegex compiles in constructor                           | Quality        | MEDIUM   | 10min  | MEDIUM | Robustness     |
| 18  | Remove unused `nolint:unparam` from getDebounceKey                     | Quality        | MEDIUM   | 5min   | LOW    | Clean Code     |
| 19  | Add context cancellation integration test                              | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 20  | Add `-race` to benchmark CI step                                       | CI             | MEDIUM   | 5min   | HIGH   | Quality        |
| 21  | Add benchmark regression detection in CI                               | CI             | MEDIUM   | 10min  | HIGH   | Quality        |
| 22  | Raise test coverage from 77% → 80% (target 90%)                        | Testing        | MEDIUM   | 30min  | HIGH   | Quality        |
| 23  | Add test for `FilterMinSize()` filter                                  | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 24  | Add test for `MiddlewareWriteFileLog()`                                | Testing        | MEDIUM   | 10min  | MEDIUM | Coverage       |
| 25  | Consolidate doc.go (add package docs)                                  | Docs           | MEDIUM   | 10min  | MEDIUM | DX             |
| 26  | Add structured logging example                                         | Docs           | MEDIUM   | 10min  | HIGH   | DX             |
| 27  | Write Troubleshooting.md                                               | Docs           | MEDIUM   | 15min  | HIGH   | Users          |
| 28  | Write migration guide for ErrorHandler signature change                | Docs           | MEDIUM   | 15min  | HIGH   | Users          |
| 29  | Add `Event.ModTime()` field to Event struct                            | Feature        | MEDIUM   | 10min  | HIGH   | Users          |
| 30  | Add `Event.Size` field to Event struct                                 | Feature        | MEDIUM   | 10min  | MEDIUM | Users          |
| 31  | Add `WithPollInterval` fallback for polling                            | Feature        | MEDIUM   | 15min  | HIGH   | Users          |
| 32  | Add `WithPolling(fallback bool)` for NFS/network                       | Feature        | MEDIUM   | 30min  | HIGH   | Users          |
| 33  | Add `Filter func type could return match metadata`                     | Feature        | MEDIUM   | 20min  | MEDIUM | DX             |
| 34  | Add `WithWatchedIgnoreDirs` option (separate filter vs walk)           | Feature        | MEDIUM   | 15min  | MEDIUM | Users          |
| 35  | Add `Watcher.AddRecursive(path)` for partial recursion                 | Feature        | MEDIUM   | 20min  | MEDIUM | Users          |
| 36  | Implement `Watch.WatchChanges(ctx, targetState)` idempotent sync       | Feature        | MEDIUM   | 30min  | MEDIUM | Users          |
| 37  | Implement exponential backoff for errors                               | Feature        | MEDIUM   | 20min  | MEDIUM | Robustness     |
| 38  | Add symlink following support                                          | Feature        | MEDIUM   | 30min  | MEDIUM | Users          |
| 39  | Add file content hashing option                                        | Feature        | MEDIUM   | 20min  | MEDIUM | Users          |
| 40  | Add recursive directory integration test                               | Testing        | MEDIUM   | 15min  | MEDIUM | Coverage       |
| 41  | Add per-path debounce correctness integration test                     | Testing        | MEDIUM   | 15min  | MEDIUM | Coverage       |
| 42  | Add benchmark regression tests                                         | Testing        | MEDIUM   | 30min  | HIGH   | Quality        |
| 43  | Document DI integration patterns in README                             | Docs           | MEDIUM   | 15min  | MEDIUM | DX             |
| 44  | Add Godoc examples (Example\* functions)                               | Docs           | MEDIUM   | 30min  | HIGH   | DX             |
| 45  | Add Prometheus metrics export                                          | Feature        | MEDIUM   | 30min  | MEDIUM | Observability  |
| 46  | Create debug mode with verbose structured logging                      | Feature        | MEDIUM   | 20min  | MEDIUM | DX             |
| 47  | Configure Goreleaser                                                   | Release        | MEDIUM   | 20min  | MEDIUM | Release        |
| 48  | Configure semantic-release                                             | Release        | MEDIUM   | 20min  | MEDIUM | Release        |
| 49  | Add stack traces to WatcherError                                       | Feature        | MEDIUM   | 15min  | MEDIUM | Debugging      |
| 50  | Add Error rate limiting middleware                                     | Feature        | MEDIUM   | 20min  | MEDIUM | Robustness     |
| 51  | Add Circuit breaker middleware                                         | Feature        | MEDIUM   | 30min  | MEDIUM | Robustness     |
| 52  | Add Context propagation through pipeline                               | Feature        | MEDIUM   | 20min  | MEDIUM | DX             |
| 53  | Add Error recovery strategies                                          | Feature        | MEDIUM   | 20min  | MEDIUM | Robustness     |
| 54  | Add Batch error handling                                               | Feature        | MEDIUM   | 15min  | MEDIUM | Robustness     |
| 55  | Add Error correlation IDs                                              | Feature        | MEDIUM   | 15min  | MEDIUM | Observability  |
| 56  | Add Error sanitization                                                 | Feature        | MEDIUM   | 15min  | MEDIUM | Security       |
| 57  | Add Error code constants                                               | Feature        | MEDIUM   | 15min  | MEDIUM | DX             |
| 58  | Add Dead letter queue                                                  | Feature        | MEDIUM   | 30min  | MEDIUM | Robustness     |
| 59  | Add OpenTelemetry integration                                          | Feature        | MEDIUM   | 45min  | MEDIUM | Observability  |
| 60  | Add Error analytics                                                    | Feature        | MEDIUM   | 30min  | LOW    | Observability  |
| 61  | Add Localizable error messages                                         | Feature        | MEDIUM   | 30min  | LOW    | i18n           |
| 62  | Implement Self-healing watcher                                         | Feature        | MEDIUM   | 45min  | MEDIUM | Robustness     |
| 63  | Review all parallel tests for race safety                              | Quality        | LOW      | 30min  | MEDIUM | Safety         |
| 64  | Explore fsnotify v2 API changes                                        | Research       | LOW      | 20min  | LOW    | Future         |
| 65  | Implement DebounceEntry Mixin phantom type                             | Refactor       | LOW      | 15min  | LOW    | Clean Code     |
| 66  | Review Remaining uint conversions                                      | Quality        | LOW      | 15min  | LOW    | Clean Code     |
| 67  | Add Windows-specific edge case tests                                   | Testing        | LOW      | 30min  | LOW    | Coverage       |
| 68  | Add Fuzz testing                                                       | Testing        | LOW      | 45min  | MEDIUM | Quality        |
| 69  | Extract drainEvents to testutil package                                | Refactor       | LOW      | 20min  | LOW    | Clean Code     |
| 70  | Test examples/ in CI pipeline                                          | CI             | LOW      | 15min  | LOW    | CI             |
| 71  | Error simulation testing                                               | Testing        | LOW      | 20min  | MEDIUM | Coverage       |
| 72  | Check if examples/ directory worth keeping vs example_test.go          | Architecture   | LOW      | 15min  | LOW    | Architecture   |
| 73  | Add API stability doc                                                  | Docs           | LOW      | 15min  | MEDIUM | Users          |
| 74  | Create standalone CLI tool                                             | Feature        | LOW      | 60min  | MEDIUM | Users          |
| 75  | Integrate into file-and-image-renamer                                  | Integration    | LOW      | 60min  | MEDIUM | Validation     |
| 76  | Integrate into dynamic-markdown-site                                   | Integration    | LOW      | 60min  | MEDIUM | Validation     |
| 77  | Integrate into auto-deduplicate                                        | Integration    | LOW      | 60min  | MEDIUM | Validation     |
| 78  | Integrate into Cyberdom                                                | Integration    | LOW      | 60min  | MEDIUM | Validation     |
| 79  | Migrate CI to Nix (Phase 3 of proposal)                                | CI             | DEFERRED | 60min  | HIGH   | DevEx          |
| 80  | Add Cachix for binary caching                                          | CI             | DEFERRED | 30min  | MEDIUM | CI             |
| 81  | Check Free disk space handling (100% full)                             | Infrastructure | BACKLOG  | 15min  | LOW    | Robustness     |
| 82  | Clear LSP diagnostic cache docs                                        | DevEx          | BACKLOG  | 5min   | LOW    | DevEx          |

---

## 🎯 QUICK WIN BATCH (Do these first - Total: ~60 min)

| #         | Task                                           | Time       |
| --------- | ---------------------------------------------- | ---------- |
| 1         | Fix `nix run .#coverage` to write to `$TMPDIR` | 5min       |
| 2         | Fix pre-commit hook timeout                    | 5min       |
| 3         | Update TODO_LIST.md - check off ALL done items | 10min      |
| 4         | Add meta attributes to all nix apps            | 5min       |
| 5         | Tag v2.0.0 release                             | 10min      |
| 6         | Add `//nolint:forbidigo` to examples           | 5min       |
| 7         | Document vendorHash update procedure           | 5min       |
| 8         | Add issue templates                            | 5min       |
| 9         | Add PR template                                | 5min       |
| 10        | Add CODE_OF_CONDUCT.md                         | 5min       |
| **TOTAL** |                                                | **~60min** |

---

## 🔴 HIGH PRIORITY (Total: ~75 min)

| #         | Task                                       | Time       |
| --------- | ------------------------------------------ | ---------- |
| 11        | Fix flaky TestWatcher_Stats_Metrics        | 15min      |
| 12        | Fix flaky TestWatcher_Watch_WithMiddleware | 15min      |
| 13        | Add `-race` to benchmark CI step           | 5min       |
| 14        | Add benchmark regression detection in CI   | 10min      |
| 15        | Raise test coverage 77% → 80%              | 30min      |
| **TOTAL** |                                            | **~75min** |

---

## 🟡 MEDIUM PRIORITY - Quality & Testing (Total: ~210 min)

| #         | Task                                               | Time                     |
| --------- | -------------------------------------------------- | ------------------------ |
| 16        | Add test for `handleError()` stderr path           | 10min                    |
| 17        | Add test for `GlobalDebouncer.Flush()`             | 10min                    |
| 18        | Add test for `handleError` with ErrorContext       | 10min                    |
| 19        | Add Example_FilterRegex test                       | 10min                    |
| 20        | Validate FilterRegex compiles in constructor       | 10min                    |
| 21        | Remove unused `nolint:unparam` from getDebounceKey | 5min                     |
| 22        | Add context cancellation integration test          | 10min                    |
| 23        | Add test for `FilterMinSize()` filter              | 10min                    |
| 24        | Add test for `MiddlewareWriteFileLog()`            | 10min                    |
| 25        | Add recursive directory integration test           | 15min                    |
| 26        | Add per-path debounce correctness integration test | 15min                    |
| 27        | Review all parallel tests for race safety          | 30min                    |
| 28        | Add Error simulation testing                       | 20min                    |
| 29        | Raise test coverage 80% → 85%                      | 45min                    |
| **TOTAL** |                                                    | **~210min (~3.5 hours)** |

---

## 🟡 MEDIUM PRIORITY - Documentation (Total: ~125 min)

| #         | Task                                             | Time                   |
| --------- | ------------------------------------------------ | ---------------------- |
| 30        | Consolidate doc.go                               | 10min                  |
| 31        | Add structured logging example                   | 10min                  |
| 32        | Write Troubleshooting.md                         | 15min                  |
| 33        | Write migration guide for ErrorHandler signature | 15min                  |
| 34        | Document DI integration patterns in README       | 15min                  |
| 35        | Add Godoc examples (Example\* functions)         | 30min                  |
| 36        | Add API stability doc                            | 15min                  |
| 37        | Check if examples/ directory worth keeping       | 15min                  |
| **TOTAL** |                                                  | **~125min (~2 hours)** |

---

## 🟡 MEDIUM PRIORITY - Features (Total: ~450 min)

| #         | Task                                               | Time                     |
| --------- | -------------------------------------------------- | ------------------------ |
| 38        | Add `Event.ModTime()` field                        | 10min                    |
| 39        | Add `Event.Size` field                             | 10min                    |
| 40        | Add `WithPollInterval` fallback                    | 15min                    |
| 41        | Add `WithPolling(fallback bool)`                   | 30min                    |
| 42        | Implement exponential backoff for errors           | 20min                    |
| 43        | Add symlink following support                      | 30min                    |
| 44        | Add file content hashing option                    | 20min                    |
| 45        | Add `Filter func type could return match metadata` | 20min                    |
| 46        | Add `WithWatchedIgnoreDirs` option                 | 15min                    |
| 47        | Add `Watcher.AddRecursive(path)`                   | 20min                    |
| 48        | Implement `Watch.WatchChanges` idempotent sync     | 30min                    |
| 49        | Add Prometheus metrics export                      | 30min                    |
| 50        | Create debug mode with verbose structured logging  | 20min                    |
| 51        | Add stack traces to WatcherError                   | 15min                    |
| 52        | Add Error rate limiting middleware                 | 20min                    |
| 53        | Add Circuit breaker middleware                     | 30min                    |
| 54        | Add Context propagation through pipeline           | 20min                    |
| 55        | Add Error recovery strategies                      | 20min                    |
| 56        | Add Batch error handling                           | 15min                    |
| 57        | Add Error correlation IDs                          | 15min                    |
| 58        | Add Error sanitization                             | 15min                    |
| 59        | Add Error code constants                           | 15min                    |
| 60        | Add Dead letter queue                              | 30min                    |
| 61        | Implement Self-healing watcher                     | 45min                    |
| **TOTAL** |                                                    | **~450min (~7.5 hours)** |

---

## 🟡 MEDIUM PRIORITY - Observability (Total: ~75 min)

| #         | Task                          | Time                     |
| --------- | ----------------------------- | ------------------------ |
| 62        | Add OpenTelemetry integration | 45min                    |
| 63        | Add Error analytics           | 30min                    |
| **TOTAL** |                               | **~75min (~1.25 hours)** |

---

## 🟡 MEDIUM PRIORITY - Release & Community (Total: ~100 min)

| #         | Task                       | Time                     |
| --------- | -------------------------- | ------------------------ |
| 64        | Configure Goreleaser       | 20min                    |
| 65        | Configure semantic-release | 20min                    |
| 66        | Create standalone CLI tool | 60min                    |
| **TOTAL** |                            | **~100min (~1.7 hours)** |

---

## 🟢 LOW PRIORITY (Total: ~330 min)

| #         | Task                                       | Time                     |
| --------- | ------------------------------------------ | ------------------------ |
| 67        | Add Localizable error messages             | 30min                    |
| 68        | Explore fsnotify v2 API changes            | 20min                    |
| 69        | Implement DebounceEntry Mixin phantom type | 15min                    |
| 70        | Review Remaining uint conversions          | 15min                    |
| 71        | Extract drainEvents to testutil package    | 20min                    |
| 72        | Add Windows-specific edge case tests       | 30min                    |
| 73        | Add Fuzz testing                           | 45min                    |
| 74        | Test examples/ in CI pipeline              | 15min                    |
| 75        | Raise test coverage 85% → 90%              | 60min                    |
| 76        | Integrate into file-and-image-renamer      | 60min                    |
| 77        | Integrate into dynamic-markdown-site       | 60min                    |
| **TOTAL** |                                            | **~330min (~5.5 hours)** |

---

## ⚪ BACKLOG / DEFERRED

| #   | Task                            | Status   | Notes                        |
| --- | ------------------------------- | -------- | ---------------------------- |
| 78  | Migrate CI to Nix (Phase 3)     | DEFERRED | Wait until after v2.0 stable |
| 79  | Add Cachix for binary caching   | DEFERRED | Wait until CI migration      |
| 80  | Integrate into auto-deduplicate | BACKLOG  | After v2.0                   |
| 81  | Integrate into Cyberdom         | BACKLOG  | After v2.0                   |
| 82  | Free disk space handling        | BACKLOG  | Infrastructure               |
| 83  | Clear LSP diagnostic cache docs | BACKLOG  | DevEx                        |

---

## 📊 TIME SUMMARY BY CATEGORY

| Category            | Time                     | % of Total |
| ------------------- | ------------------------ | ---------- |
| Quick Wins          | ~60min                   | 5%         |
| High Priority       | ~75min                   | 6%         |
| Quality & Testing   | ~210min                  | 18%        |
| Documentation       | ~125min                  | 11%        |
| Features            | ~450min                  | 38%        |
| Observability       | ~75min                   | 6%         |
| Release & Community | ~100min                  | 8%         |
| Low Priority        | ~330min                  | 28%        |
| **TOTAL**           | **~1425min (~24 hours)** | 100%       |

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
