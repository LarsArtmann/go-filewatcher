# PRO/CONTRA: Integrating `samber/do/v2`

**Date:** 2026-04-04  
**Status:** Evaluated — Recommendation: **Do not integrate**

---

## Context

**go-filewatcher** is a focused, ~600 LOC library with 2 direct deps (`fsnotify`, `cockroachdb/errors`), using functional options and middleware chains.

**samber/do/v2** is a type-safe DI container with generics: `Provide`/`Invoke`, scoped hierarchies, lifecycle hooks (HealthChecker, Shutdowner), lazy/eager/transient loading.

---

## PRO

| # | Argument | Detail |
|---|----------|--------|
| 1 | **Testability via injection** | Could inject a mock `fsnotify.Watcher` or `DebouncerInterface` instead of relying on interface checks at runtime (`watcher.go:466`). Currently, testing the watch loop requires real filesystem events. DI would make the `fswatcher` field injectable. |
| 2 | **Lifecycle management** | `Watcher` already has `Close()` and could implement `Shutdowner` + `Healthchecker` from `samber/do`. Consumers using `do` could get automatic shutdown ordering if they register the watcher as a service. |
| 3 | **Scoped debounce instances** | `do.Scope` could model per-tenant or per-request watcher instances with parent-child isolation — potentially interesting for multi-tenant file watching. |
| 4 | **Named service registration** | Multiple watchers with different configs could be registered as `do.ProvideNamed(injector, "config-watcher", ...)` and `do.ProvideNamed(injector, "source-watcher", ...)`, making composition in larger apps cleaner. |
| 5 | **Framework alignment** | If `go-cqrs-lite` ecosystem adopts `samber/do`, alignment would reduce friction for shared consumers. |

---

## CONTRA

| # | Argument | Detail |
|---|----------|--------|
| 1 | **Library vs Application mismatch** | `samber/do` solves wiring problems in **applications**. `go-filewatcher` is a **library**. DI containers belong at the composition root — the consumer's `main()`. Adding DI _inside_ a library imposes a framework choice on every consumer. |
| 2 | **Violates "minimal deps" principle** | `AGENTS.md` explicitly states: _"Keep it minimal — no other deps."_ Adding `samber/do` would be a philosophical reversal. It would also add transitive dependencies and increase binary size for every consumer. |
| 3 | **Functional options already solve this** | The `Option func(*Watcher)` pattern already provides clean, composable configuration. DI provides no benefit over what `WithDebounce()`, `WithFilter()`, `WithMiddleware()` already deliver — it just adds indirection. |
| 4 | **No complex dependency graph** | The watcher has exactly 1 real dependency (`fsnotify.Watcher`) and optional pluggable interfaces (`DebouncerInterface`, `Filter`, `Middleware`). There is no dependency graph to manage. DI shines with 10-50+ services; here it's 1. |
| 5 | **Breaks the public API** | Currently: `filewatcher.New(paths, opts...)`. With DI: consumers must create an `Injector`, call `Provide`, then `Invoke`. This is a worse DX for the 95% use case. |
| 6 | **Exhaustruct linter conflict** | The project enforces `exhaustruct` (all struct fields initialized). `samber/do`'s lazy instantiation pattern means structs are created inside provider closures — making field initialization harder to audit and more error-prone. |
| 7 | **Consumer already can use DI** | Nothing prevents a consumer from doing `do.Provide(injector, func(i do.Injector) (*filewatcher.Watcher, error) { return filewatcher.New(...) })`. The library doesn't need to _depend on_ `samber/do` to _work with_ it. |
| 8 | **Added complexity for no behavioral gain** | No current bug, feature gap, or architectural pain point is solved by DI. The middleware chain, filter composition, and debounce strategies already work correctly and are well-tested. |

---

## Verdict

**Recommendation: Do not integrate.**

`samber/do/v2` is a well-designed DI toolkit, but it solves a problem `go-filewatcher` doesn't have. The project is a focused library with a flat dependency tree, not an application with a complex service graph.

### Better alternatives per desired benefit

| Desired benefit | Better approach |
|---|---|
| Testability | Extract `fsnotify.Watcher` behind an interface in `watcher.go` — zero new deps |
| Lifecycle hooks | Add `HealthCheck() error` and let consumers wrap it in `do.Shutdowner` at their level |
| Multi-watcher management | Document a pattern: `do.ProvideNamed(injector, "my-watcher", ...)` at consumer side |
| Scoped instances | Consumers create multiple `filewatcher.New()` calls in their own scope hierarchy |

The single highest-impact improvement would be **extracting `fsnotify.Watcher` behind an internal interface** — that gives testability without any new dependency.
