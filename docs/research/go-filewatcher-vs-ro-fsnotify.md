# go-filewatcher vs samber/ro fsnotify Plugin

> **TL;DR** — These are different abstractions solving different problems. `ro/fsnotify` is a 40-line adapter that puts raw fsnotify events into a reactive stream. `go-filewatcher` is a full production file-watching SDK. They are complementary, not competing.

---

## What Each Library Is

|                          | **go-filewatcher**                                                 | **samber/ro fsnotify**                                                 |
| ------------------------ | ------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| **What**                 | Purpose-built, production-grade file watcher library               | One plugin (~40 LOC) in a general-purpose reactive streams framework   |
| **Source size**          | ~3,000+ lines across 20+ files                                     | **40 lines** (`source.go`)                                             |
| **Philosophy**           | Batteries-included: everything a production watcher needs          | Thin fsnotify → Observable adapter; power comes from `ro` operators    |
| **When to reach for it** | "I need a production file watcher with all the edge cases handled" | "I already use `ro` and want to treat file events as reactive streams" |

---

## Feature Matrix

### Core Watching

| Capability                  | go-filewatcher                                              | ro/fsnotify                                                       |
| --------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------------- |
| Recursive directory walking | Built-in, with depth control (`AddRecursive(path, depth)`)  | No — only watches paths you pass                                  |
| Gitignore-aware walking     | Built-in via `go-gitignore`                                 | None                                                              |
| Dynamic path management     | `Add()`, `AddRecursive()`, `Remove()`, `Reset()` at runtime | Fixed at construction — paths passed to `NewFSListener(paths...)` |
| Symlink resolution          | Configurable via `WithFollowSymlinks()`                     | None                                                              |
| Path exclusions             | `WithExcludePaths()` for prefix-based exclusion during walk | None                                                              |
| Batched watch registration  | Batches of 1000 dirs with `runtime.Gosched()` yields        | None                                                              |

### Event Processing

| Capability                   | go-filewatcher                                                                                                | ro/fsnotify                                                                      |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| **Debouncing**               | Two modes: global + per-path, built-in                                                                        | Via `ro.ThrottleTime` / `ro.DebounceTime` operators                              |
| **Filtering**                | 15+ built-in filters (ext, glob, regex, size, time, ops, hidden, dirs) + composable (`FilterAnd`, `FilterOr`) | Via `ro.Filter` operator — you write your own predicates                         |
| **Middleware pipeline**      | Full chain: logging, recovery, rate limiting, deduplication, circuit breaker, metrics, OTel tracing           | N/A — `ro` operators _are_ the pipeline (`ro.Pipe`, `ro.Map`, `ro.Filter`, etc.) |
| **Generated code filtering** | `gogenfilter` integration (oapi, deepcopy, wire, moq)                                                         | None                                                                             |
| **Content hashing**          | SHA-256 on events via `WithContentHashing()`                                                                  | None                                                                             |
| **Event enrichment**         | `IsDir`, `Size`, `ModTime`, `Hash`, `Timestamp` on every event                                                | Raw `fsnotify.Event` (`Name`, `Op` only)                                         |

### Reliability

| Capability                | go-filewatcher                                                                       | ro/fsnotify                                              |
| ------------------------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------- |
| Self-healing              | Retries failed watch registrations periodically                                      | None                                                     |
| Polling fallback          | Built-in for NFS/FUSE/Docker volumes                                                 | None                                                     |
| inotify budget management | Auto-detects `/proc/sys/fs/inotify/max_user_watches`, graceful degradation on ENOSPC | None — errors terminate the Observable                   |
| Graceful shutdown         | `Close()` with `sync.WaitGroup`, channel close ordering, `sync.Once`                 | Teardown function closes fsnotify watcher on unsubscribe |
| Thread safety             | Documented guarantees per method, RWMutex, atomic counters                           | Context cancellation + ro's subscription model           |

### Error Handling

| Capability      | go-filewatcher                                                                                                               | ro/fsnotify                                          |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| Error taxonomy  | Structured: sentinel errors, `WatcherError` with error codes + categories (transient/permanent), stack traces, error channel | Errors emitted on Observable error stream            |
| Retry semantics | Categorized as transient (retryable) or permanent                                                                            | Via `ro.Retry` / `ro.Catch` operators                |
| Error channel   | Dedicated `<-chan error` via `Errors()`                                                                                      | Part of the Observable contract (`OnError` callback) |

### Observability

| Capability     | go-filewatcher                                                | ro/fsnotify             |
| -------------- | ------------------------------------------------------------- | ----------------------- |
| Prometheus     | Built-in collector (counters + gauges), dependency-free       | None — compose yourself |
| OpenTelemetry  | Built-in span middleware (`OTelMiddleware`)                   | None — compose yourself |
| Debug logging  | `WithDebug(logger)` for verbose pipeline logging              | None                    |
| Stats snapshot | `Stats()` returns counters, uptime, watch limit, budget usage | None                    |

### Serialization

| Capability       | go-filewatcher                                           | ro/fsnotify                                      |
| ---------------- | -------------------------------------------------------- | ------------------------------------------------ |
| JSON marshaling  | Full `encoding.TextMarshaler`/`JSON` on `Op` and `Event` | Standard `fsnotify.Event` (no custom marshaling) |
| slog integration | `Event.LogValue()` for structured logging                | None                                             |

---

## Architecture

| Dimension               | go-filewatcher                                             | ro/fsnotify                                                            |
| ----------------------- | ---------------------------------------------------------- | ---------------------------------------------------------------------- |
| **Pattern**             | Imperative, channel-based (`<-chan Event`)                 | Declarative, reactive (Observable/Observer)                            |
| **Composability model** | Functional options + middleware chain + filter combinators | `ro.Pipe` with 100+ operators (Map, Filter, Merge, Scan, Buffer, etc.) |
| **Multi-subscriber**    | Single channel per watcher                                 | Observable supports multiple subscribers natively                      |
| **Backpressure**        | Buffered channel with configurable size                    | Built-in backpressure strategies                                       |
| **Dependency count**    | 3 (fsnotify, go-gitignore, gogenfilter) + x/time/rate      | 2 (fsnotify, samber/ro) — but `ro` itself is a large framework         |
| **Coupling**            | Self-contained library                                     | Coupled to `ro` reactive framework                                     |
| **Event type**          | Custom `Event` struct with enriched metadata               | Raw `fsnotify.Event`                                                   |
| **Type safety**         | Phantom types (`EventPath`, `RootPath`, `DebounceKey`)     | None — raw `fsnotify.Event`                                            |

---

## Code Comparison

### Basic Usage

**go-filewatcher:**

```go
watcher, _ := filewatcher.New([]string{"./src"},
    filewatcher.WithDebounce(300*time.Millisecond),
    filewatcher.WithExtensions(".go"),
    filewatcher.WithIgnoreDirs("vendor", "node_modules"),
)

events, _ := watcher.Watch(ctx)
for event := range events {
    fmt.Println(event.Op, event.Path)
}
```

**ro/fsnotify:**

```go
observable := ro.Pipe2(
    rofsnotify.NewFSListener("./src"),
    ro.Filter(func(e fsnotify.Event) bool {
        return filepath.Ext(e.Name) == ".go"
    }),
    ro.ThrottleTime[fsnotify.Event](300*time.Millisecond),
)

subscription := observable.Subscribe(ro.NewObserver(
    func(e fsnotify.Event) { fmt.Println(e.Op, e.Name) },
    func(err error) { log.Println(err) },
    func() {},
))
defer subscription.Unsubscribe()
```

### Adding Middleware / Operators

**go-filewatcher — built-in middleware:**

```go
watcher, _ := filewatcher.New([]string{"./src"},
    filewatcher.WithMiddleware(
        MiddlewareRecovery(),
        MiddlewareLogging(slog.Default()),
        MiddlewareRateLimit(100),
        MiddlewareDeduplicate(200*time.Millisecond),
    ),
)
```

**ro/fsnotify — compose with operators:**

```go
observable := ro.Pipe4(
    rofsnotify.NewFSListener("./src"),
    ro.Filter(func(e fsnotify.Event) bool { return true }),
    ro.ThrottleTime[fsnotify.Event](100*time.Millisecond),
    ro.DistinctUntilChanged(/* comparator */),
    ro.Map(func(e fsnotify.Event) string { return e.Name }),
)
```

### Multi-Subscriber

**go-filewatcher** — single channel, fan-out yourself:

```go
events, _ := watcher.Watch(ctx)
// Fan-out manually
ch1 := make(chan filewatcher.Event)
ch2 := make(chan filewatcher.Event)
go func() {
    for e := range events {
        ch1 <- e
        ch2 <- e
    }
}()
```

**ro/fsnotify** — native multi-subscriber:

```go
obs := rofsnotify.NewFSListener("./src")
sub1 := obs.Subscribe(handler1)
sub2 := obs.Subscribe(handler2)
```

---

## Where Each Wins

### go-filewatcher

1. **Production robustness** — Self-healing, graceful ENOSPC handling, inotify budget tracking, polling fallback for NFS/FUSE
2. **Recursive watching** — Walks directory trees, handles new directories, depth-limited recursion
3. **File-watcher-specific features** — Gitignore awareness, content hashing, generated code detection, path exclusions, symlink resolution
4. **Observability out of the box** — Prometheus, OTel, debug logging, structured stats
5. **Error taxonomy** — Structured errors with codes, categories, stack traces
6. **Zero framework coupling** — No dependency on a reactive streams framework
7. **Dynamic path management** — Add/remove paths at runtime

### ro/fsnotify

1. **Composability with other streams** — File events become first-class reactive citizens alongside timers, HTTP responses, signals
2. **Multi-subscriber** — Multiple observers without extra plumbing
3. **Backpressure** — Built-in strategies in `ro`
4. **Rich transformations** — `ro.Map`, `ro.Scan`, `ro.Buffer`, `ro.Window`, `ro.GroupBy` and 100+ operators
5. **Simplicity** — 40 lines of source. Zero-friction for basic "file changed" notifications

---

## Decision Guide

| Use go-filewatcher when...                             | Use ro/fsnotify when...                                                    |
| ------------------------------------------------------ | -------------------------------------------------------------------------- |
| You need recursive directory watching                  | You already use `ro` for reactive streams                                  |
| Deploying to NFS/FUSE/Docker volumes                   | You need multi-subscriber file events                                      |
| You need Prometheus/OTel observability                 | You want to compose file events with other streams (HTTP, timers, signals) |
| You care about inotify limits and graceful degradation | You just need basic "file changed" notifications                           |
| You want gitignore-aware walking                       | Your use case is simple and you value minimalism                           |
| You need structured errors with retry semantics        | You're already invested in the reactive paradigm                           |
| You want dynamic add/remove of paths at runtime        |                                                                            |

---

## Could They Work Together?

Yes. You could wrap `go-filewatcher` as a `ro.Observable` source — getting both the reactive composability of `ro` and the production robustness of `go-filewatcher`:

```go
func NewRobustFSListener(paths []string, opts ...filewatcher.Option) ro.Observable[filewatcher.Event] {
    return ro.NewUnsafeObservableWithContext(func(ctx context.Context, dest ro.Observer[filewatcher.Event]) ro.Teardown {
        w, err := filewatcher.New(paths, opts...)
        if err != nil {
            dest.ErrorWithContext(ctx, err)
            return nil
        }

        events, err := w.Watch(ctx)
        if err != nil {
            dest.ErrorWithContext(ctx, err)
            return nil
        }

        go func() {
            for event := range events {
                dest.NextWithContext(ctx, event)
            }
            dest.CompleteWithContext(ctx)
        }()

        return func() { _ = w.Close() }
    })
}
```

This gives you recursive walking, gitignore, self-healing, polling, Prometheus, OTel — all composable with `ro.Pipe` operators.

---

_Research date: 2026-06-08. Based on `samber/ro` fsnotify plugin source at `plugins/fsnotify/source.go` and `go-filewatcher` v2.2.0._
