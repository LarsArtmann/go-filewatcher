# File Organization


| File                   | Responsibility                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------------------ |
| `watcher.go`           | Public API: New, Watch, Add, AddRecursive, Remove, Reset, WatchList, Stats                       |
| `backend.go`           | watchBackend interface + fsnotifyBackend adapter (test seam for fake backend injection)          |
| `watcher_internal.go`  | Event processing: watchLoop, middleware, emitEvent, debugLog, handleError                        |
| `watcher_walk.go`      | Directory walking: addPath, walkAndAddPaths, addBatch, symlink resolution, budget detection      |
| `watcher_gitignore.go` | .gitignore loading and matching: gitignoreCache, shouldSkipByGitignore                           |
| `watcher_selfheal.go`  | Self-healing: selfHealLoop, attemptSelfHeal, failed path tracking                                |
| `watcher_poll.go`      | Polling mode: pollLoop for NFS/FUSE environments                                                 |
| `filesystem.go`        | Filesystem case-sensitivity: FilesystemCaseSensitivity enum, pathKey(), resolveCaseSensitivity() |
| `filter.go`            | All Filter functions + FilterWithMeta and combinators                                            |
| `filter_gogen.go`      | Generated-code detection filter (gogenfilter v3 integration)                                     |
| `middleware.go`        | All Middleware functions (circuit breaker, error batch, correlation, exponential backoff)        |
| `metrics.go`           | PrometheusCollector, StatsFunc, CounterMetric, GaugeMetric                                       |
| `otel.go`              | OTelMiddleware, OTelSpan interface                                                               |
| `debouncer.go`         | Debouncer + GlobalDebouncer                                                                      |
| `event.go`             | Op type, Event type, JSON/Text marshaling                                                        |
| `errors.go`            | Sentinel errors, ErrorCode, ErrorCategory, WatcherError                                          |
| `options.go`           | Functional options (WithGitignore, WithExcludePaths, WithMaxWatches, etc.)                       |
| `phantom_types.go`     | Compile-time phantom types (EventPath, RootPath, DebounceKey, OpString, etc.)                    |

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
