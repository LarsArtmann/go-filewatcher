# PRO/CONTRA: Adopting samber/ro in go-filewatcher

> Assessment of three adoption strategies: replacing the core API, optional adapter, and internal use.

---

## Option A: Make `ro.Observable[Event]` the core API (replace `<-chan Event`)

### PRO

| Argument                             | Why it matters                                                                                                               |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| **Multi-subscriber for free**        | Currently single channel — users fan-out themselves. Observables natively support N subscribers                              |
| **Rich operator vocabulary**         | 100+ operators (`Map`, `Scan`, `Buffer`, `Window`, `GroupBy`, `Retry`, `Catch`) replace custom middleware for many use cases |
| **Backpressure built-in**            | ro handles slow consumers; go-filewatcher currently drops events when buffer is full                                         |
| **Composability with other streams** | File events alongside HTTP responses, timers, signals — all one type system                                                  |
| **samber is a trusted author**       | `samber/lo` (18k+ stars), well-maintained, idiomatic Go style                                                                |

### CONTRA

| Argument                                | Why it matters                                                                                                                                                                                                                                    |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Breaks every existing user**          | `<-chan Event` → `Observable[Event]` is a total API rewrite. All consumers must change                                                                                                                                                            |
| **Reactive is not idiomatic Go**        | Go's philosophy: channels, goroutines, `select`. Observable chains are a paradigm mismatch. This will alienate Go developers who expect `for event := range ch`                                                                                   |
| **Huge dependency injection**           | go-filewatcher currently depends on 3 focused libs. `ro` is a framework (100+ files, its own module system, plugin architecture). It redefines the nature of the project                                                                          |
| **Overlap with existing features**      | Middleware, filters, debounce, deduplication, rate limiting — all already built domain-specific. Replacing them with generic `ro` operators _loses_ domain semantics (e.g., `FilterIgnoreDirs` is more meaningful than `ro.Filter(func(e) bool)`) |
| **Cold Observable semantics don't fit** | ro's default: each subscription creates new execution. File watching wants: single shared execution, multiple observers. You'd need `ro.Connectable`/`ro.Subject` everywhere, fighting the model                                                  |
| **Debugging complexity**                | Reactive stacks are notoriously hard to debug. A filter failing deep in a `ro.Pipe6(...)` chain gives worse errors than a clear middleware chain                                                                                                  |
| **Lost domain richness**                | `WatcherError` with codes/categories/stack traces doesn't map cleanly to `ro`'s error channel. `Stats()`, Prometheus collector, OTel middleware — all need rethinking                                                                             |
| **Execution model conflict**            | go-filewatcher runs `watchLoop` + `pollLoop` + `selfHealLoop` goroutines with `sync.WaitGroup` shutdown. ro has its own goroutine lifecycle. Merging these creates subtle race conditions                                                         |
| **API surface explosion**               | Users must learn both go-filewatcher _and_ ro's operator model. Current API: `New()`, `Watch()`, `Close()`. With ro: `Pipe`, `Subscribe`, `Unsubscribe`, `Connect`, `Subject`, backpressure strategies...                                         |

### Verdict

**DON'T.** The cost/benefit is deeply negative. You'd sacrifice Go idiomaticity, domain specificity, and existing users for a composability model that doesn't match file watching semantics.

---

## Option B: Optional `ro` adapter (sub-package or separate module)

Add a `filewatcher/adapter/ro` package that wraps a `Watcher` as `ro.Observable[Event]`.

### PRO

| Argument                       | Why it matters                                                                 |
| ------------------------------ | ------------------------------------------------------------------------------ |
| **Zero cost for non-ro users** | Optional dependency. Core API unchanged                                        |
| **Best of both worlds**        | Production robustness (self-heal, gitignore, polling) + reactive composability |
| **Clear separation**           | Domain logic stays in go-filewatcher; stream composition stays in ro           |
| **Low maintenance**            | ~30 lines of adapter code                                                      |
| **Attracts ro users**          | "Use go-filewatcher with your reactive pipeline" is a compelling story         |

### CONTRA

| Argument                       | Why it matters                                                                                                                                                              |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Another module to maintain** | Tests, compatibility tracking across both APIs                                                                                                                              |
| **Semantic impedance**         | `Watcher.Watch(ctx)` returns `<-chan Event` with lifecycle guarantees. Wrapping as Observable loses some control (when does the Observable "complete"? who owns `Close()`?) |
| **Testing burden**             | Must test against two framework versions (go-filewatcher + ro semver)                                                                                                       |
| **Small audience**             | Intersection of "ro users" AND "need production file watching" is likely small                                                                                              |

### Verdict

**LOW PRIORITY, but reasonable.** A `filewatcher/adapter/ro` sub-module is a clean additive feature. It doesn't compromise the core and costs ~50 lines total. But it shouldn't be a priority unless users ask for it.

---

## Option C: Use `ro` primitives internally (implementation detail)

Replace internal event processing pipelines with ro's operator chains.

### PRO

| Argument                                                          | Why it matters |
| ----------------------------------------------------------------- | -------------- |
| Could simplify internal debounce/filter/middleware implementation |

### CONTRA

| Argument                                 | Why it matters                                                                                                                                   |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Adds heavy dep for internal use only** | Users pay the dependency cost without seeing the benefit                                                                                         |
| **Fighting the domain**                  | go-filewatcher's middleware chain (`func(Handler) Handler`) is a cleaner fit for synchronous event processing than Observable's async push model |
| **Performance regression risk**          | ro has its own scheduler, goroutine pool, and allocation patterns. go-filewatcher's current direct channel writes are likely faster              |
| **Debugging internal ro chains**         | When a user reports "events stopped flowing," tracing through ro internals is harder than tracing `watchLoop` → `emitEvent`                      |
| **Loss of domain control**               | Self-heal retries, batch registration, inotify budget — these need tight control over goroutine lifecycle. ro abstracts that away                |

### Verdict

**DON'T.** Internal adoption adds dependency weight without user-facing benefit and fights the domain model.

---

## Summary

| Option                                  | Recommendation                                                  | Confidence |
| --------------------------------------- | --------------------------------------------------------------- | ---------- |
| **A: Replace core API with Observable** | **No.** Architectural overreach, breaks users, fights Go idioms | High       |
| **B: Optional adapter module**          | **Maybe later.** Clean additive, but wait for user demand       | Medium     |
| **C: Use ro internally**                | **No.** Dependency cost without user benefit                    | High       |

**The honest assessment:** go-filewatcher's current architecture — channels + functional options + middleware chain — is well-suited to its domain. It's idiomatic Go, zero-framework, and purpose-built. `ro` solves a real problem (reactive composition) but it's a _different_ problem than "production-grade file watching with all edge cases handled."

---

_Assessment date: 2026-06-08. Based on `samber/ro` fsnotify plugin source at `plugins/fsnotify/source.go` and `go-filewatcher` v2.2.0._
