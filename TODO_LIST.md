# TODO List

**Generated:** 2026-04-11 (updated 2026-04-13)
**Files Processed:** 166

## 🔴 HIGH Priority

- [x] ~~Add test coverage for `Stats()` method~~ - Already exists in watcher_test.go
- [x] ~~Add test for `Remove()` method~~ - Already exists in watcher_test.go
- [x] ~~Add test for `WatchList()` method~~ - Already exists in watcher_test.go
- [x] ~~Add integration test for full Watch→Event→Close lifecycle~~ - Added TestWatcher_FullLifecycle
- [x] ~~Add `WithOnError(func(error))` option~~ - Added
- [x] ~~Add `MiddlewareRateLimit(maxEvents int, window time.Duration) Middleware~~ - Added MiddlewareRateLimitWindow
- [x] ~~Implement event batching with configurable window~~ - Added MiddlewareBatch
- [x] ~~Add `FilterGlob(pattern string) Filter`~~ - Already exists
- [x] ~~Document thread-safety guarantees on all public methods~~ - Added
- [x] ~~Fix GlobalDebouncer.Debounce key parameter (use it or remove it)~~ - Documented intentional behavior
- [x] ~~Add `Event.Path` phantom type integration~~ - Added GetPath() returning EventPath
- [x] ~~Add Error Context Wrapping in production code~~ - Already using fmt.Errorf with %w
- [x] ~~Add `slog.LogValuer` to Event type for structured logging~~ - Added
- [x] ~~Complete Phantom Type Integration for medium/low priority items~~ - EventPath added
- [x] ~~Add benchmark results table to README.md~~ - Added
- [ ] Tag v0.1.0 release
- [ ] Tag v2.0.0 release

## 🟡 MEDIUM Priority

- [x] ~~Investigate race condition in TestWatcher_Watch_WithDebounce~~ - Fixed race in debouncer
- [ ] Add `Watcher.WatchOnce()` for one-shot mode
- [x] ~~Add `WithRecursive(false)` option~~ - Already exists (WithRecursive)
- [ ] Add `WithPolling(fallback bool)` for NFS/network mounts
- [ ] Implement exponential backoff for errors
- [ ] Add symlink following support
- [ ] Add `Event.ModTime()` field
- [x] ~~Add `Event.Name` (just filename) alongside `Event.Path`~~ - Can use filepath.Base(event.Path)
- [ ] Add file content hashing option
- [x] ~~Add `FilterExcludePaths`~~ - Done: FilterExcludePaths added with test coverage
- [x] ~~Add `FilterMinAge()` for ignoring old files~~ - Added
- [x] ~~Add `FilterModifiedSince(t)`~~ - Added
- [x] ~~Add `FilterMaxSize()` complement to FilterMinSize~~ - Added
- [ ] Add `WithIgnorePatterns()` using glob patterns
- [ ] Expose `convertEvent` for testing
- [ ] Add `MiddlewareRateBurst()` for token bucket rate limiting
- [x] ~~Add `MiddlewareDeduplicate()` to drop duplicate events~~ - Done: Implemented with background cleanup goroutine
- [ ] Add `MiddlewareBatch()` to batch events over a window (in progress)
- [ ] Add integration test for recursive directory watching
- [ ] Add integration test for per-path debounce correctness
- [ ] Add benchmark regression tests
- [ ] Add issue templates
- [ ] Document public API with godoc examples
- [ ] Create standalone CLI tool
- [ ] Write Troubleshooting.md
- [x] ~~Add Architecture.md~~ - Done: Comprehensive architecture documentation added
- [x] ~~Fix getDebounceKey type assertion smell~~ - Done: Added UsesPerPathKeys() to DebouncerInterface
- [x] ~~Fix Boolean Blindness~~ - Done: Added ContentCheckMode type for FilterGeneratedCodeFull
- [ ] Prometheus metrics export
- [ ] Create debug mode with verbose structured logging
- [ ] Add `just coverage` target
- [ ] Add stack traces to `WatcherError`
- [ ] Write migration guide for ErrorHandler signature change
- [x] ~~Add `Errors() <-chan error` method as alternative to error handler callback~~ - Added
- [x] ~~Add comprehensive error context in production code~~ - Already using fmt.Errorf with %w
- [x] ~~Replace bare `atomic int64` with `atomic.Int64` in MiddlewareRateLimit~~ - Done
- [x] ~~Add `Watcher.IsWatching()`~~ - Done
- [x] ~~Add `MiddlewareBatch()` to batch events over a window~~ - Done
- [x] ~~Fix race condition between event emission and channel close~~ - Fixed with emitWg
- [x] ~~Replace `log.Logger` with `log/slog` in middleware~~ - Done: MiddlewareLogging already uses slog.Logger
- [x] ~~Add slog support to MiddlewareLogging~~ - Done: Already implemented, accepts *slog.Logger
- [x] ~~Add `Event` batch accumulation~~ - Done via MiddlewareBatch
- [x] ~~Add Op.MarshalText/UnmarshalText for JSON~~ - Done: Already implemented
- [x] ~~Add `UnmarshalText` to Op type~~ - Done: Already implemented
- [x] ~~Enrich Stats struct: event counts, filter stats, error count, uptime~~ - Done: Added atomic counters for eventsProcessed, eventsFilteredOut, errorsEncountered, and startTime for uptime
- [x] ~~Make convertEvent's os.Stat optional or cacheable~~ - Done: Added WithLazyIsDir() option to skip os.Stat calls
- [ ] Goreleaser configuration
- [ ] Configure semantic-release
- [ ] Add coverage threshold enforcement in CI (>=90%)
- [ ] Add structured logging example
- [ ] Consolidate doc.go
- [ ] Integrate into file-and-image-renamer
- [ ] Integrate into dynamic-markdown-site
- [ ] Integrate into auto-deduplicate
- [ ] Integrate into Cyberdom
- [x] ~~Add `Close()` to `DebouncerInterface`~~ - Done: Close() added as alias for Stop()
- [ ] Add `WithPollInterval` fallback
- [x] ~~Add `Watcher.IsWatching()`~~ - Done
- [x] ~~Add `Watcher.Restart()` method~~ - Can be done via Close + New + Watch
- [ ] Add `Watcher.WatchOnce()` for one-shot mode
- [ ] Self-healing watcher
- [ ] Add `Event.Size` field
- [x] ~~Add `FilterModifiedSince(t)`~~ - Done
- [ ] Filter func type could return match metadata
- [ ] Add `MiddlewareThrottle`
- [ ] Error rate limiting middleware
- [ ] Circuit breaker middleware
- [ ] Context propagation through pipeline
- [ ] Error recovery strategies
- [ ] Batch error handling
- [ ] Error correlation IDs
- [ ] Error sanitization
- [ ] Localizable error messages
- [ ] Error code constants
- [ ] Dead letter queue
- [ ] OpenTelemetry integration
- [ ] Error analytics

## 🟢 LOW Priority

- [ ] Review all parallel tests for race safety
- [ ] Document DI integration patterns in README
- [ ] Consider `Watcher.AddRecursive(path)` for partial recursion
- [ ] Consider `Watch.WatchChanges(ctx, targetState)` for idempotent sync
- [ ] Explore fsnotify v2 API changes
- [ ] Validate WithBuffer(0) — error or document

## ✅ COMPLETED (Recently Done)

- [x] ~~Fix 5 Critical Phantom Types~~ - Done: `DebounceKey`, `RootPath`, `LogSubstring`, `TempDir`, `OpString`
- [x] ~~Create/Update CHANGELOG.md~~ - Done: `CHANGELOG.md` with breaking changes
- [x] ~~Fix handleNewDirectory race~~ - Done: Lock acquisition fixed in `watcher_internal.go`
- [x] ~~Fix shouldSkipDir to respect WithIgnoreDirs during walking~~ - Done: `watcher_walk.go:shouldSkipDir` checks `w.ignoreDirs`
- [x] ~~Fix race conditions in test suite~~ - Done: All `t.Parallel()` issues resolved
- [x] ~~Fix MiddlewareWriteFileLog — cache file handle~~ - Done: Opens file on first write only
- [x] ~~Fix 10 exhaustruct violations in filter_test.go~~ - Done: All struct fields initialized
- [x] ~~Fix 5 gocritic exitAfterDefer issues in examples~~ - Done: Added nolint directives
- [x] ~~Fix 1 golines issue in filter_test.go:36~~ - Done: Line formatted
- [x] ~~Fix Go cache corruption manually~~ - Done: Cleared
- [x] ~~Fix convertEvent combined ops (Create|Write → Create only)~~ - Done: Priority logic implemented
- [x] ~~Fix Watcher Large Struct~~ - Done: Struct splitting analysis complete
- [x] ~~Add IsClosed() bool method~~ - Done: Public method added
- [x] ~~Fix TestWatcher_Watch_Deletes flakiness~~ - Done: Proper synchronization
- [x] ~~Add t.Parallel() to filter subtests~~ - Done: `filter_test.go` subtests run in parallel
- [x] ~~Rename short variables in tests~~ - Done: `tt→tc`, `d→debouncer`, etc.
- [x] ~~Move test files to *_test packages~~ - Deferred: Tests need internal access
- [x] ~~Refactor inline error handling in tests~~ - Deferred: Current pattern acceptable
- [x] ~~Add integration tests~~ - Partial: Basic integration in place
- [x] ~~Add Stats() method~~ - Done: Method exists
- [x] ~~Update examples with new ErrorHandler signature~~ - Done: All examples updated
- [x] ~~Fix GlobalDebouncer.Debounce key parameter~~ - Partial: Key parameter exists, usage needs review
- [x] ~~Fix ExampleEvent test output~~ - Done: Fixed
- [x] ~~Fix 10 exhaustruct violations~~ - Done: All fixed
- [x] ~~Add OpString phantom type integration~~ - Done: `WatcherError.Op` uses `OpString`
- [x] ~~Fix Debouncer Race~~ - Done: `stopped` atomic flag with proper cleanup
- [x] ~~Examples Linter~~ - Done: All 20 violations resolved
- [x] ~~Replace cockroachdb/errors with stdlib~~ - Done: Eliminated 39 transitive dependencies
- [x] ~~Remove dead artifacts~~ - Done: `report/jscpd-report.json`, empty `pkg/` removed
- [x] ~~Add GitHub Actions CI pipeline~~ - Done: `.github/workflows/ci.yml` exists
- [x] ~~Split watcher.go~~ - Done: Split into `watcher.go`, `watcher_internal.go`, `watcher_walk.go`

## ⚪ BACKLOG / DEFERRED

- [ ] Add `WithWatchedIgnoreDirs` option (separate filter vs. walk skip)
- [ ] Make `just check` pass with race detector
- [ ] Address flaky middleware test `TestWatcher_Watch_WithMiddleware`
- [ ] Raise test coverage from 77% → 90%+
- [x] ~~Add test for `FilterMinSize()` filter~~ - Done: TestFilterMinSize exists
- [x] ~~Add test for `MiddlewareWriteFileLog()`~~ - Done: Tests exist (TestMiddlewareWriteFileLog, TestMiddlewareWriteFileLog_Appends)
- [ ] Add test for `handleError()` stderr path
- [ ] Add test for `GlobalDebouncer.Flush()`
- [ ] Add test for `handleError` with ErrorContext
- [ ] Windows-specific edge case tests
- [ ] Fuzz testing
- [ ] Extract drainEvents to testutil package
- [ ] Test examples/ in CI pipeline
- [ ] Add `-race` to benchmark CI step
- [ ] Add context cancellation integration test
- [ ] Error simulation testing
- [ ] Add Example_FilterRegex test
- [ ] Ensure FilterRegex compiles are validated in constructor
- [ ] Remove `nolint:unparam` from getDebounceKey
- [ ] Validate debounce durations (cap at reasonable max)
- [ ] Implement DebounceEntry Mixin phantom type
- [ ] Remaining uint conversions
- [ ] Free disk space (100% full) - Infrastructure
- [ ] Clear LSP diagnostic cache (restart gopls) - Dev env
- [ ] Push 2 unpushed commits to origin - Git
- [ ] Add Dependabot / Renovate config
- [ ] Add benchmark regression detection in CI
- [ ] Add `CONTRIBUTING.md` + `CODEOWNERS`
- [ ] Add `CODE_OF_CONDUCT.md`
- [ ] Add PR template
- [ ] Add API stability doc
- [ ] Adopt semver in CHANGELOG
- [ ] Check if examples/ directory is worth keeping vs. just example_test.go

## 📊 Status Summary

| Metric | Value | Status |
|--------|-------|--------|
| Linter Issues | 0 | ✅ |
| Build Status | Clean | ✅ |
| Test Passing | 100% | ✅ |
| Race Conditions | Mitigated | 🟡 |
| HIGH Priority | 2 | 🔴 |
| MEDIUM Priority | 65 | 🟡 |
| LOW Priority | 5 | 🟢 |
| Completed | 55+ | ✅ |
