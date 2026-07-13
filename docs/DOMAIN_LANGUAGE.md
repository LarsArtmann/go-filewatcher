# Domain Language

A **Unified Language** for `go-filewatcher` — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a customer, define it here.

## Glossary

| Term                 | Definition                                                                     | Context                           |
| -------------------- | ------------------------------------------------------------------------------ | --------------------------------- |
| Watcher              | The central type that monitors filesystem paths and emits events               | Core abstraction of the library   |
| Event                | A single filesystem change notification with path, op, timestamp, and metadata | The primary output of the watcher |
| Op                   | The operation type: Create, Write, Remove, or Rename                           | Enumerated on every Event         |
| Filter               | A predicate function `(Event) bool` that decides which events to emit          | Filtering pipeline stage          |
| Middleware           | A function wrapping event handlers for cross-cutting concerns                  | Middleware pipeline stage         |
| Debouncer            | Coalesces rapid successive events into a single emission                       | Debouncing stage                  |
| Watch Budget         | The inotify watch limit; auto-detected or overridden via `WithMaxWatches`      | Resilience context                |
| Self-healing         | Auto-retry of failed watch registrations at a configurable interval            | Resilience context                |
| Polling Mode         | Fallback event detection via periodic filesystem snapshots (NFS/FUSE)          | Resilience context                |
| Gitignore-aware Walk | Directory walking that skips paths matching `.gitignore` patterns              | Walk-time optimization            |

## Entities

Objects with identity and lifecycle.

| Term    | Definition                                                 | Context                                   |
| ------- | ---------------------------------------------------------- | ----------------------------------------- |
| Watcher | Monitors filesystem paths, manages goroutines and channels | Created via `New()`, closed via `Close()` |

## Value Objects

Immutable objects defined by attributes.

| Term         | Definition                                                              | Context                      |
| ------------ | ----------------------------------------------------------------------- | ---------------------------- |
| Event        | Filesystem change notification (path, op, timestamp, size, modtime)     | Emitted on the event channel |
| Op           | Operation enum: `Create`, `Write`, `Remove`, `Rename`                   | Part of every Event          |
| Filter       | Predicate `(Event) bool` — composable via AND/OR/NOT                    | Filtering pipeline           |
| Stats        | Runtime counters: events processed, filtered, errors, watch budget      | Observability                |
| WatcherError | Structured error with category (transient/permanent), code, stack       | Error handling               |
| EventPath    | Phantom-typed path string with `.Base()`, `.Dir()`, `.Ext()`, `.Join()` | Type safety on event paths   |
| ErrorCode    | Typed string constant for programmatic error matching                   | Error handling               |

## Events

Things that happen in the domain.

| Term   | Definition                      | Context                                     |
| ------ | ------------------------------- | ------------------------------------------- |
| Create | A file or directory was created | Highest priority when multiple ops coalesce |
| Write  | A file was written to           | Second priority                             |
| Remove | A file or directory was removed | Third priority                              |
| Rename | A file or directory was renamed | Lowest priority; Chmod events are ignored   |

## Commands

Actions the system can perform.

| Term         | Definition                                         | Context                             |
| ------------ | -------------------------------------------------- | ----------------------------------- |
| New          | Create a watcher from paths with options           | Entry point                         |
| Watch        | Start the event loop, return `<-chan Event`        | Starts goroutine, channel on cancel |
| WatchOnce    | Return the first event then close                  | One-shot mode                       |
| Add          | Dynamically add a path to an existing watcher      | Runtime path management             |
| AddRecursive | Add a path with a recursion depth limit            | 0=flat, -1=full, N=depth-limited    |
| Remove       | Remove a path and all subdirectory watches         | Subtree-aware cleanup               |
| Reset        | Clear runtime state while preserving configuration | Allows re-Watch after Close         |
| Close        | Stop watching, close channels, clean up goroutines | Idempotent                          |

## Bounded Contexts

Subsystems with distinct vocabulary.

| Context        | Description                                                             |
| -------------- | ----------------------------------------------------------------------- |
| Core Watching  | Path management, event loop, channel streaming, context cancellation    |
| Filtering      | Predicates that decide event emission; composable via AND/OR/NOT        |
| Middleware     | Wrapping event handlers for logging, metrics, rate limiting, resilience |
| Debouncing     | Coalescing rapid events: global (all events) or per-path (each file)    |
| Resilience     | ENOSPC handling, inotify budget, self-healing, polling fallback         |
| Observability  | Stats, debug logging, Prometheus collector, OpenTelemetry tracing       |
| Error Handling | Sentinel errors, typed codes, structured WatcherError, error channel    |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
