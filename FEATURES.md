# Feature Inventory

**Last Updated:** 2026-06-18 · **Version:** v2.2.0

Honest status of every capability in go-filewatcher. Statuses:

- ✅ **DONE** — Production-ready, tested, documented
- 🟡 **PARTIALLY DONE** — Works but incomplete, rough edges, or limited docs
- 🔵 **PLANNED** — Committed to in TODO_LIST.md, not yet started
- ⚪ **WORTH CONSIDERING** — Ideas worth exploring; no commitment

---

## Core Watching

| Feature                           | Status | Notes                                                                |
| --------------------------------- | ------ | -------------------------------------------------------------------- |
| Create watcher from paths         | ✅     | `New(paths, opts...)` validates paths exist and are directories      |
| Start/stop with `context.Context` | ✅     | `Watch(ctx)` returns `<-chan Event`, channel closes on cancel/Close  |
| One-shot mode                     | ✅     | `WatchOnce(ctx)` returns the first event and closes                  |
| Recursive directory watching      | ✅     | On by default; `WithRecursive(false)` disables                       |
| Selective recursion depth         | ✅     | `AddRecursive(path, maxDepth)` — 0=flat, -1=full, N=depth-limited    |
| Dynamic path management           | ✅     | `Add`, `Remove` (subtree-aware), `WatchList`                         |
| Reset without rebuilding config   | ✅     | `Reset()` clears runtime state, preserves filters/middleware/options |
| Thread-safe concurrent access     | ✅     | All public methods documented safe-by-design; tested with `-race`    |
| Graceful close                    | ✅     | `Close()` idempotent; stops debouncer before closing channels        |

## Filtering

| Feature                         | Status | Notes                                                             |
| ------------------------------- | ------ | ----------------------------------------------------------------- |
| Extensions / IgnoreExtensions   | ✅     | Dot-prefixed                                                      |
| IgnoreDirs / ExcludePaths       | ✅     | Name-based vs absolute-path prefix matching                       |
| IgnoreHidden                    | ✅     | Dot-prefixed files/dirs                                           |
| Operations / NotOperations      | ✅     | By `Op` enum                                                      |
| Glob / Regex                    | ✅     | Filename glob, full-path regex                                    |
| MinSize / MaxSize               | ✅     | Bytes                                                             |
| MinAge / ModifiedSince          | ✅     | Time-based                                                        |
| IgnoreGlobs (patterns)          | ✅     | `WithIgnorePatterns` option                                       |
| ContentHash                     | ✅     | Filter by expected SHA-256                                        |
| Gitignore repository matcher    | ✅     | `FilterGitignore(repoRoot)` — event-time check against .gitignore |
| Generated-code detection        | ✅     | sqlc, protobuf, templ, mockgen, stringer via gogenfilter v3.2.0   |
| Filter combinators (AND/OR/NOT) | ✅     | `FilterAnd`, `FilterOr`, `FilterNot`                              |
| Metadata-returning filters      | ✅     | `FilterWithMeta`, `MatchResult`, `FilterWithMetaAnd/Or/Not`       |

## Middleware

| Feature                 | Status | Notes                                                          |
| ----------------------- | ------ | -------------------------------------------------------------- |
| Logging (slog)          | ✅     | `MiddlewareLogging(*slog.Logger)`                              |
| Panic recovery          | ✅     | `MiddlewareRecovery()`                                         |
| Filter-as-middleware    | ✅     | `MiddlewareFilter(Filter)`                                     |
| OnError handling        | ✅     | `MiddlewareOnError(func(Event, error))`                        |
| Rate limiting (fixed)   | ✅     | `MiddlewareRateLimit(maxEvents)`                               |
| Rate limiting (sliding) | ✅     | `MiddlewareSlidingWindowRateLimit(maxEvents, window)`          |
| Throttle (token bucket) | ✅     | `MiddlewareThrottle(maxEvents, burst)` via `golang.org/x/time` |
| Metrics counter         | ✅     | `MiddlewareMetrics(func(Op))`                                  |
| Deduplicate             | ✅     | `MiddlewareDeduplicate(window)`                                |
| Batch                   | ✅     | `MiddlewareBatch(window, maxSize, flush)`                      |
| Audit to file           | ✅     | `MiddlewareWriteFileLog(path)`                                 |
| Circuit breaker         | ✅     | `MiddlewareCircuitBreaker(maxFailures, resetTimeout)`          |
| Exponential backoff     | ✅     | `MiddlewareExponentialBackoff(maxFailures, initial, max)`      |
| Error rate limit        | ✅     | `MiddlewareErrorRateLimit(maxErrors, window)`                  |
| Error recovery strategy | ✅     | `MiddlewareErrorRecovery(strategy)`                            |
| Error correlation IDs   | ✅     | `MiddlewareErrorCorrelation(idGenerator)`                      |
| Error sanitization      | ✅     | `MiddlewareErrorSanitization(sanitize)`                        |
| Error batching          | ✅     | `MiddlewareErrorBatch(window, maxSize, flush)`                 |

## Debouncing

| Feature                | Status | Notes                                                               |
| ---------------------- | ------ | ------------------------------------------------------------------- |
| Global debounce        | ✅     | `WithDebounce(d)` — all events coalesced                            |
| Per-path debounce      | ✅     | `WithPerPathDebounce(d)` — independent per file                     |
| Programmatic debouncer | ✅     | `Debouncer` and `GlobalDebouncer` types with `Flush`/`Stop`/`Close` |

## Observability

| Feature                          | Status | Notes                                                            |
| -------------------------------- | ------ | ---------------------------------------------------------------- |
| `Stats()` struct                 | ✅     | Events, filters, errors, uptime, watch budget                    |
| Structured debug logging         | ✅     | `WithDebug(*slog.Logger)`                                        |
| Prometheus collector             | ✅     | `PrometheusCollector` with `StatsFunc`, `CounterMetric`, `Gauge` |
| OpenTelemetry tracing middleware | ✅     | `OTelMiddleware` with `OTelSpan` interface (zero-dep)            |
| Stack traces on errors           | ✅     | `WatcherError.Stack` via `debug.Stack()`                         |

## Resilience & Scalability

| Feature                    | Status | Notes                                                               |
| -------------------------- | ------ | ------------------------------------------------------------------- |
| Graceful ENOSPC handling   | ✅     | Add errors logged, walk continues, `Stats.WatchErrors` tracks fails |
| Inotify budget awareness   | ✅     | Auto-detected from `/proc/sys/fs/inotify/max_user_watches`          |
| Watch limit override       | ✅     | `WithMaxWatches(n)`                                                 |
| Self-healing watches       | ✅     | `WithSelfHeal(interval)` retries failed paths                       |
| Batched watch registration | ✅     | 1000 dirs/batch with `runtime.Gosched()` between batches            |
| Polling mode (NFS/FUSE)    | ✅     | `WithPolling(true)` + `WithPollInterval(d)`                         |
| Symlink following          | ✅     | `WithFollowSymlinks(true)`                                          |
| .gitignore-aware walking   | ✅     | `WithGitignore(true)` (default) skips gitignored dirs at walk time  |
| Path-level exclusions      | ✅     | `WithExcludePaths(paths...)` prefix-matches during walk             |

## Event Metadata

| Feature              | Status | Notes                                                    |
| -------------------- | ------ | -------------------------------------------------------- |
| Path, Op, Timestamp  | ✅     | Core fields                                              |
| IsDir, Size, ModTime | ✅     | Populated from `os.Stat`                                 |
| Content hash         | ✅     | `WithContentHashing()` — SHA-256, capped 10 MiB          |
| JSON marshaling      | ✅     | `Op` and `Event` implement `MarshalJSON`/`UnmarshalJSON` |
| Text marshaling      | ✅     | `Op` implements `MarshalText`/`UnmarshalText`            |
| `slog.LogValuer`     | ✅     | `Event` integrates with structured logging               |

## Type Safety (Phantom Types)

| Feature                   | Status | Notes                                                             |
| ------------------------- | ------ | ----------------------------------------------------------------- |
| `EventPath`               | ✅     | `Event.GetPath()` returns typed path with `.Base/.Dir/.Ext/.Join` |
| `RootPath`                | ✅     | Walking roots                                                     |
| `DebounceKey`             | ✅     | Debouncer keys                                                    |
| `OpString`                | ✅     | Operation names on `WatcherError`                                 |
| `LogSubstring`, `TempDir` | ✅     | Test-only assertion helpers                                       |

## Error Handling

| Feature                       | Status | Notes                                                    |
| ----------------------------- | ------ | -------------------------------------------------------- |
| Sentinel errors               | ✅     | `ErrWatcherClosed`, `ErrNoPaths`, `ErrPathNotFound`, ... |
| Typed error codes             | ✅     | `ErrorCode` constants for programmatic matching          |
| Structured `WatcherError`     | ✅     | Category (transient/permanent), op, stack trace          |
| Channel-based error stream    | ✅     | `Errors() <-chan error`                                  |
| Custom error handler callback | ✅     | `WithErrorHandler` / `WithOnError`                       |

## Developer Experience

| Feature                          | Status | Notes                                                        |
| -------------------------------- | ------ | ------------------------------------------------------------ |
| Nix flake dev shell              | ✅     | `nix develop`, `direnv allow`                                |
| Nix apps for all common commands | ✅     | `nix run .#{check,ci,test,lint,lint-fix,bench,coverage,...}` |
| GitHub Actions CI                | ✅     | Test with race + 90% threshold, lint, examples-build, bench  |
| Godoc examples                   | ✅     | 7 examples in `example_test.go`                              |
| Runnable example programs        | ✅     | `examples/{basic,middleware,per-path-debounce}`              |
| Cross-platform releases          | ✅     | `.goreleaser.yml` configuration                              |
| Issue templates                  | ✅     | Bug report + feature request                                 |

## Planned / Worth Considering

See [ROADMAP.md](./ROADMAP.md) for long-term direction and [TODO_LIST.md](./TODO_LIST.md) for committed short/mid-term work. Highlights:

| Feature                                         | Status | Notes                                        |
| ----------------------------------------------- | ------ | -------------------------------------------- |
| Windows-specific edge case tests                | 🔵     | Currently CI runs Linux only                 |
| Fuzz testing expansion                          | 🔵     | Existing fuzz tests; expand to more surfaces |
| Error simulation testing framework              | 🔵     | For testing error middleware paths           |
| Goreleaser release pipeline                     | 🔵     | Config exists; publish workflow not wired    |
| Semantic-release automation                     | ⚪     | Currently manual version bumps               |
| Localizable error messages                      | ⚪     | Sentinel errors are English-only today       |
| fsnotify v2 tracking                            | ⚪     | Monitor upstream for breaking changes        |
| `WatchChanges(ctx, targetState)` idempotent API | ⚪     | For sync-style workflows                     |
