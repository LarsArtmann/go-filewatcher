# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Core watcher: `New()`, `Watch(ctx)→<-chan Event`, `Add()`, `Close()`
- 9 functional options: debounce, per-path debounce, filter, extensions, ignore dirs, ignore hidden, recursive, middleware, error handler
- 11 composable filters: Extensions, IgnoreExtensions, IgnoreDirs, IgnoreHidden, Operations, NotOperations, Glob, And, Or, Not
- 7 middleware: Logging, Recovery, RateLimit, Filter, OnError, Metrics, WriteFileLog
- Per-key `Debouncer` and `GlobalDebouncer`
- 4 sentinel errors with `cockroachdb/errors`
- Channel-based event streaming with context cancellation
- Automatic recursive directory watching with dynamic new-dir detection
- 50 tests, 86.1% coverage, race detector clean
