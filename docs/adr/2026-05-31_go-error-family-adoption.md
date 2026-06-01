# ADR: Should go-filewatcher adopt go-error-family?

**Date:** 2026-05-31
**Status:** Rejected

## Context

[go-error-family](https://github.com/larsartmann/go-error-family) is a structured error protocol for Go, providing 5 behavioral Families (Rejection, Conflict, Transient, Corruption, Infrastructure), machine-readable codes, CLI boundary handling, message templates, diagnostics, and agent analysis. Both libraries are by the same author. Should go-filewatcher adopt it as a dependency?

## Decision

**No.** go-filewatcher will keep its existing domain-specific error system.

## Conceptual Overlap

| go-error-family | go-filewatcher | Notes |
|---|---|---|
| `Family` (5 families) | `ErrorCategory` (Transient/Permanent/Unknown) | filewatcher's domain is simpler — 3 categories suffice |
| `Coded` interface → `ErrorCode() string` | `ErrorCode` type + `Code()` method | Same concept, different implementation |
| `Classified` interface → `ErrorFamily() Family` | `WatcherError.IsTransient()` / `.IsPermanent()` | Same intent, different granularity |
| `Contextual` interface → `ErrorContext() map[string]string` | `ErrorContext` struct (Operation/Path/Event/Retryable) | Struct is more domain-specific |
| `errorfamily.New()` constructors | `NewWatcherError()` constructor | filewatcher's is simpler and **unused in production code** |
| `Classify()` + `RegisterClassification()` | `categorizeError()` (hardcoded switch) | filewatcher's sentinels are fixed; no extensibility needed |
| `HandleError()` → stderr + exit code | `handleError()` → channel/callback/stderr | filewatcher has no CLI boundary |
| Diagnostics (`diagnose/`) | None | Filesystem/Network diagnostics don't apply here |
| Agent analysis (`agent/`) | None | Debug agent is overkill for a watcher library |

## Reasoning

### 1. Mismatch of domain complexity

go-error-family solves the **CLI/HTTP boundary problem** — where errors leave your program and meet humans or downstream systems. go-filewatcher is a **library** with no `main()`, no exit codes, no HTTP handlers. Its errors are consumed programmatically by callers via `errors.Is()`. The 5-family model (Rejection/Conflict/Transient/Corruption/Infrastructure) with Audience/Tone metadata is designed for a different problem.

### 2. The current system is fit for purpose

The existing `errors.go` provides:
- 11 sentinel errors with clear domain meaning
- `ErrorCategory` (Transient/Permanent) — sufficient for "should I retry?"
- `WatcherError` with Op/Path/Err/Category/Stack — exactly the fields that matter for a file watcher
- `ErrorCode` for programmatic matching
- `ErrorContext` + `ErrorHandler` for runtime error dispatch

This is domain-specific, lean, and has **zero external dependencies**.

### 3. NewWatcherError is never called in production code

The production path is `fmt.Errorf("%w: ...", sentinel)` — callers use `errors.Is()` for matching. `WatcherError` is exported but unused in the hot path. The structured error system is barely exercised. Adopting go-error-family would add a dependency for something that's essentially dead weight already.

### 4. Dependency cost exceeds value

Adding `github.com/larsartmann/go-error-family` as a dependency:
- Pulls a new import into every consumer's dependency tree
- Introduces a shared type system consumers must now understand
- Creates coupling between two evolving libraries by the same author
- The zero-dep claim of error-family is nice, but the **conceptual** dependency is heavy

### 5. go-error-family is v0.2.0

Not yet v1.0. Adopting it means buying into an unstable API for a domain (error classification) that's already solved in filewatcher.

## Potential Internal Cleanup (Separate Concern)

The current `errors.go` could be simplified independently:
- `NewWatcherError()` is unused in production code → candidate for removal or deprecation
- `ErrorCode` constants duplicate the sentinel identity → could be derived at runtime
- `ErrorContext` struct overlaps with `WatcherError` fields

This is internal cleanup, not a reason to adopt an external library.

## Conclusion

The libraries solve different problems at different abstraction levels. go-error-family is for CLI/HTTP error presentation at system boundaries. go-filewatcher needs simple programmatic error classification for library consumers. The current system is adequate, dependency-free, and domain-specific. Adopting go-error-family would be over-engineering.
