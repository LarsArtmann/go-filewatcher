# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed

- Relicensed from Proprietary to MIT

### Fixed

- `Add()` no longer double-appends to `WatchList()` in recursive mode
- `MiddlewareBatch` timer-triggered flush errors now logged via `slog.Error` instead of silently dropped
- `handleNewDirectory` now propagates `addPath` errors to the error handler
- `MiddlewareSlidingWindowRateLimit` uses in-place slice compaction instead of per-event allocation

### Added

- `DefaultIgnoreDirsCopy()` function for safe access without mutation risk
- Debounce option validation: panics on negative durations
- `Errors() <-chan error` method for channel-based error consumption
- `IsWatching()` and `IsClosed()` state inspection methods
- `WithLazyIsDir()` option to skip `os.Stat` calls for performance
- `WithOnAdd()` callback option for path tracking
- `WithOnError()` simplified error callback option
- `FilterMaxSize()`, `FilterMinAge()`, `FilterModifiedSince()` filters
- `MiddlewareDeduplicate()`, `MiddlewareBatch()`, `MiddlewareSlidingWindowRateLimit()`
- `FilterGeneratedCode()`, `FilterGeneratedCodeFull()` via gogenfilter integration
- Compile-time phantom types for `EventPath`, `RootPath`, `DebounceKey`, `OpString`
- `Event.GetPath()` returning phantom-typed `EventPath`
- `slog.LogValuer` on `Event` for structured logging
- 15 new tests covering rename events, multi-directory init, concurrent ops, state transitions

### Changed

- Replaced hand-rolled `Op.MarshalJSON` with `json.Marshal` for robustness
- Modernized `errors.As` to Go 1.26 `AsType` pattern
- `testing_helpers.go` renamed to `testing_helpers_test.go` (no longer ships to consumers)
- `flake.nix` Go version aligned to 1.26 (was 1.24)
- `FilterExcludePaths` no longer calls `filepath.Abs` per event
- `WithBuffer(0)` now allowed with documented caveat

### Removed

- 306 lines of test-only code from production binary

## [0.2.0] - 2026-04-23

### Added

- `go-branded-id` integration for compile-time phantom type safety
- `FilterGeneratedCode()` and `FilterGeneratedCodeFull()` via gogenfilter v0.2.0
- `OpString`, `LogSubstring`, `TempDir`, `DebounceKey`, `RootPath` phantom types
- Extracted shared test helper functions for DRY
- Benchmark suite migrated to `b.Loop()` pattern

### Changed

- Migrated to gogenfilter v0.2.0 API
- Updated flake.lock to nixpkgs eb3b085

### Fixed

- Data race between `Close()` and `buildEmitFunc`
- Data race between `Close()` and debouncer callbacks
- fsnotify assertion tests tolerant of duplicate events

## [0.1.0] - 2026-04-04

### Added

- Core watcher: `New()`, `Watch(ctx)→<-chan Event`, `Add()`, `Remove()`, `WatchList()`, `Stats()`, `Close()`
- 14 functional options: debounce, per-path debounce, filter, extensions, ignore dirs, ignore hidden, recursive, middleware, error handler, skip dot dirs, buffer, on add, on error, lazy is dir
- 13 composable filters: Extensions, IgnoreExtensions, IgnoreDirs, ExcludePaths, IgnoreHidden, Operations, NotOperations, Glob, Regex, MinSize, MaxSize, MinAge, ModifiedSince
- Filter combinators: `FilterAnd`, `FilterOr`, `FilterNot`
- 10 middleware: Logging, Recovery, Filter, OnError, RateLimit, SlidingWindowRateLimit, Metrics, Deduplicate, Batch, WriteFileLog
- Per-key `Debouncer` and `GlobalDebouncer` with Flush/Pending/Stop
- 10 sentinel errors with structured `WatcherError` (transient/permanent categorization)
- `Errors() <-chan error` for channel-based error consumption
- `IsWatching()` and `IsClosed()` state inspection
- Channel-based event streaming with context cancellation
- Automatic recursive directory watching with dynamic new-dir detection
- `MiddlewareLogging` accepts `*slog.Logger` for structured logging
- `slog.LogValuer` on `Event` type
- JSON marshaling for `Op` and `Event` types
- Benchmarks for creation, filters, middleware, debounce, full pipeline
- GitHub Actions CI (test with race + 90% threshold, lint with 90+ rules)
- Nix flake dev shell for 4 platforms
- Comprehensive documentation: README, ARCHITECTURE.md, MIGRATION.md, examples

### Changed

- Replaced `cockroachdb/errors` with stdlib (eliminated 39 transitive dependencies)
- Split `watcher.go` into `watcher.go`, `watcher_internal.go`, `watcher_walk.go`

### Removed

- `cockroachdb/errors` dependency
- Dead artifacts: `report/jscpd-report.json`, empty `pkg/` directory
