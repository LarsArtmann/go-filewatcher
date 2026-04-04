# ADR: Integrating `samber/do/v2`

**Date:** 2026-04-04  
**Status:** Evaluated — Recommendation: **Do not integrate**  
**Revised:** 2026-04-04 (added gaps from reflection)

---

## Context

**go-filewatcher** is a focused library (~600 LOC, 2 direct deps: `fsnotify`, `cockroachdb/errors`) providing composable file system watching. It uses functional options (`Option func(*Watcher)`), middleware chains (`Middleware func(Handler) Handler`), composable filters (`Filter func(Event) bool`), and channel-based event streaming.

**samber/do/v2** is a type-safe dependency injection container for Go using generics. Key features: `Provide`/`Invoke` service registration, scoped hierarchies, lifecycle hooks (HealthChecker, Shutdowner), lazy/eager/transient loading, dependency-aware parallel shutdown.

---

## Existing Composition Patterns (what DI would replace)

The codebase already uses Go's first-class functions for composition — the same patterns DI frameworks automate:

| Pattern | Current implementation | DI equivalent |
|---------|----------------------|---------------|
| Configuration | `Option func(*Watcher)` | `do.Provide` with config struct |
| Filtering | `Filter func(Event) bool` | Service-level filter injection |
| Middleware | `Middleware func(Handler) Handler` | Service decorator chain |
| Debounce strategy | `DebouncerInterface` (2 implementations) | `do.Provide` with interface |
| Error handling | `WithErrorHandler(func(error))` | `do.ProvideValue` |

These are already composable, testable, and idiomatic Go. DI would add indirection without adding capability.

---

## Code Comparison: Current API vs DI API

### Current (clean, 3 lines):

```go
watcher, err := filewatcher.New(
    []string{"./src"},
    filewatcher.WithExtensions(".go"),
    filewatcher.WithDebounce(500*time.Millisecond),
)
```

### With samber/do (7 lines, requires library knowledge):

```go
injector := do.New()
do.ProvideValue(injector, &WatchConfig{Paths: []string{"./src"}})
do.Provide(injector, func(i do.Injector) (*filewatcher.Watcher, error) {
    cfg := do.MustInvoke[*WatchConfig](i)
    return filewatcher.New(cfg.Paths,
        filewatcher.WithExtensions(".go"),
        filewatcher.WithDebounce(500*time.Millisecond),
    )
})
watcher := do.MustInvoke[*filewatcher.Watcher](injector)
```

The DI version adds 4 lines of boilerplate and requires understanding `samber/do`'s API — for zero behavioral gain.

---

## PRO

| # | Argument | Detail |
|---|----------|--------|
| 1 | **Testability via injection** | Could inject a mock `fsnotify.Watcher` instead of relying on real filesystem events in tests. Currently, the watch loop is tested with real I/O (`watcher_test.go:107-251`). |
| 2 | **Lifecycle management** | `Watcher` already implements `io.Closer` (`watcher.go:71`). Adding `Shutdowner`/`Healthchecker` from `samber/do` would give automatic shutdown ordering in DI-aware applications. |
| 3 | **Scoped debounce instances** | `do.Scope` could model per-tenant watcher instances with parent-child isolation for multi-tenant scenarios. |
| 4 | **Named service registration** | Multiple watchers with different configs: `do.ProvideNamed(injector, "config-watcher", ...)`. |
| 5 | **Framework alignment** | If `go-cqrs-lite` ecosystem adopts `samber/do`, alignment reduces friction. |

---

## CONTRA

| # | Argument | Detail |
|---|----------|--------|
| 1 | **Library vs Application mismatch** | DI solves wiring in **applications**. `go-filewatcher` is a **library**. DI containers belong at the composition root — the consumer's `main()`. Adding DI _inside_ a library imposes a framework choice on every consumer. |
| 2 | **Violates "minimal deps" principle** | `AGENTS.md`: _"Keep it minimal — no other deps."_ `samber/do` adds transitive dependencies and increases binary size. |
| 3 | **Functional options already solve this** | `WithDebounce()`, `WithFilter()`, `WithMiddleware()` provide clean composable configuration. DI adds indirection without capability. |
| 4 | **No complex dependency graph** | 1 real dependency (`fsnotify.Watcher`) and 3 pluggable interfaces (`DebouncerInterface`, `Filter`, `Middleware`). DI shines with 10-50+ services. |
| 5 | **Breaks the public API** | `filewatcher.New(paths, opts...)` → `Injector` + `Provide` + `Invoke`. Worse DX for the 95% use case. |
| 6 | **Exhaustruct linter conflict** | Project enforces `exhaustruct` (`.golangci.yml:28`). Lazy instantiation in provider closures makes field auditing harder. |
| 7 | **Consumer already can use DI** | `do.Provide(injector, func(i do.Injector) (*filewatcher.Watcher, error) { return filewatcher.New(...) })`. Library doesn't need to _depend on_ `samber/do` to _work with_ it. |
| 8 | **No behavioral gain** | No bug, feature gap, or architectural pain is solved. Middleware chains, filter composition, and debounce strategies already work correctly. |
| 9 | **Go ecosystem convention** | Go favors explicit dependency passing over DI containers. Standard library (`http.Handler`, `io.Reader`) uses interfaces + functions, not containers. |
| 10 | **Testing already works** | `watcher_test.go` achieves ~90% coverage with real filesystem. The marginal testability gain from DI doesn't justify the architectural cost. |

---

## Related Work: Go DI Landscape

| Library | Approach | Best for | Fit for this project |
|---------|----------|----------|---------------------|
| `samber/do/v2` | Runtime DI, generics, scopes | Large applications with 20+ services | ❌ Overkill |
| `google/wire` | Compile-time DI, code gen | Applications wanting type safety | ❌ Overkill, adds build step |
| `uber-go/fx` | Runtime DI, reflection | Uber-style microservices | ❌ Heavy, opinionated |
| `sarulabs/di` | Runtime DI, scopes | Web applications | ❌ Overkill |
| **None (current)** | Functional options + interfaces | **Focused libraries** | ✅ Correct choice |

The Go community consensus for libraries: accept interfaces, return structs, use functional options. No DI container needed.

---

## Quantitative Analysis

| Metric | Current | With samber/do |
|--------|---------|----------------|
| Direct dependencies | 2 | 3+ |
| Public API surface | `New()`, options, `Watch()` | + Injector, Provide, Invoke |
| Lines of boilerplate to create watcher | 3-5 | 7-10 |
| Concepts to learn | options pattern | options + DI container + scopes |
| Test dependencies | none | samber/do (in test scope) |
| Binary size impact | minimal | +samper/do transitive deps |

---

## Verdict

**Recommendation: Do not integrate.**

`samber/do/v2` is well-designed, but `go-filewatcher` has no problem it solves. The project is a focused library, not an application with a complex service graph.

---

## Actionable Improvements (independent of this decision)

These improvements address the legitimate needs that motivated the DI evaluation, without adding a DI dependency.

### High impact, low work

| # | Improvement | Work | Impact |
|---|-------------|------|--------|
| 1 | **Extract `fsnotify.Watcher` behind internal interface** | Small | High — enables mock-based watch loop testing without real I/O |
| 2 | **Add `HealthCheck() error` to `Watcher`** | Small | Medium — consumers using DI can wrap it at their level |
| 3 | **Document DI integration pattern** in README | Tiny | Medium — shows consumers how to use with `samber/do`, `wire`, `fx` |

### Medium impact, medium work

| # | Improvement | Work | Impact |
|---|-------------|------|--------|
| 4 | **Use `log/slog` in middleware** (stdlib since Go 1.21) | Medium | Medium — structured logging instead of `log.Logger` |
| 5 | **Add `Event` batch accumulation** | Medium | Medium — useful for consumers who want to process events in batches |

### Already done (this session)

| # | Improvement | Commit |
|---|-------------|--------|
| ✅ | Remove unused nolint directive | `3eaf3e4` |
| ✅ | Replace custom `contains()` with `strings.Contains` | `de57c1e` |
| ✅ | Fix stale `pkg/errors/` reference in AGENTS.md | `83d08ad` |
| ✅ | Add `Pending()` to `GlobalDebouncer` for API consistency | `813328a` |
| ✅ | Add `TextMarshaler`/`TextUnmarshaler` to `Op` + json tags to `Event` | `6d934dc` |
