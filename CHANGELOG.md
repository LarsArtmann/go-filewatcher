# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Core watcher: `New()`, `Watch(ctx)→<-chan Event`, `Add()`, `Remove()`, `WatchList()`, `Stats()`, `Close()`
- 9 functional options: debounce, per-path debounce, filter, extensions, ignore dirs, ignore hidden, recursive, middleware, error handler
- 13 composable filters: Extensions, IgnoreExtensions, IgnoreDirs, IgnoreHidden, Operations, NotOperations, Glob, Regex, MinSize, And, Or, Not
- 7 middleware: Logging, Recovery, RateLimit, Filter, OnError, Metrics, WriteFileLog
- Per-key `Debouncer` and `GlobalDebouncer`
- 4 sentinel errors with stdlib `errors` + `fmt.Errorf`
- Channel-based event streaming with context cancellation
- Automatic recursive directory watching with dynamic new-dir detection
- `MiddlewareLogging` now accepts `*slog.Logger` for structured logging
- `MiddlewareWriteFileLog` caches file handle for performance
- Benchmarks for filters, debouncers, and middleware
- GitHub Actions CI workflow (build, test with race, lint, coverage)
- JSON marshaling for `Op` type

### Changed

- Replaced `cockroachdb/errors` with stdlib `errors`/`fmt.Errorf` (eliminated 39 transitive dependencies)
- `MiddlewareLogging` changed from `*log.Logger` to `*slog.Logger`
- `shouldSkipDir` now respects user-configured `WithIgnoreDirs` during directory walking
- Split `watcher.go` into `watcher.go`, `watcher_internal.go`, `watcher_walk.go` for readability

### Fixed

- Flaky `TestWatcher_Watch_Deletes` — properly drains all create/write events before testing remove

### Removed

- `cockroachdb/errors` dependency (only `fsnotify` remains as direct dependency)
- Dead artifacts: `report/jscpd-report.json`, empty `pkg/` directory
