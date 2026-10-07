# Key Patterns: Default-Guard Convention & Middleware Cleanup

Full detail moved from `AGENTS.md` (2026-10-07). Summary kept in AGENTS.md:
every tunable middleware default substitutes a **named const** via a
non-positive guard; shared defaulting lives in `resolve*Defaults` helpers
(`resolveRateLimitDefaults`, `resolveBatchDefaults`, `resolveMaxFailures`),
unique defaulting stays inline. `TestMiddlewareDefaultConsts_AllUsed` guards
the inventory. Middleware has no lifecycle hook — pair factories with
`WithCleanup(fn)` for resources.


_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._


Every middleware that accepts a tunable (duration, count, threshold) must
substitute a **named const** when the caller passes a non-positive value — never
a magic literal. The decision of _where_ the defaulting lives has one rule:

> **Shared defaulting → `resolve*Defaults` helper. Unique defaulting → inline guard.**

- **Two or more functions share the same defaulting** → extract a `resolve*Defaults`
  helper. This is the DRY path: one named const, one guard, one test target.
- **Exactly one function uses the default** → keep an inline `if x <= 0` guard with
  a named const. A helper for a single caller is indirection without benefit.

**Worked example** (`middleware.go`):

```go
// SHARED — two middlewares (SlidingWindowRateLimit + ErrorRateLimit) reuse the
// same window default and the same non-positive guard. Extract.
const defaultRateLimitWindow = time.Second

func resolveRateLimitDefaults(maxValue, defaultMax int, window time.Duration) (int, time.Duration) {
    if maxValue <= 0 {
        maxValue = defaultMax
    }
    if window <= 0 {
        window = defaultRateLimitWindow
    }
    return maxValue, window
}

func MiddlewareSlidingWindowRateLimit(maxEvents int, window time.Duration) Middleware {
    maxEvents, window = resolveRateLimitDefaults(maxEvents, defaultSlidingWindowEvents, window)
    // ...
}

// UNIQUE — only MiddlewareThrottle uses this default. Inline guard, named const.
const defaultThrottleEvents = 100

func MiddlewareThrottle(maxEvents, burst int) Middleware {
    if maxEvents <= 0 {
        maxEvents = defaultThrottleEvents
    }
    // ...
}
```

The three shared helpers are `resolveRateLimitDefaults`, `resolveBatchDefaults`,
and `resolveMaxFailures`. Each has direct table-driven coverage, and
`TestMiddlewareDefaultConsts_AllUsed` guards the whole inventory: if a default
const loses its call site in a refactor, the test fails before linters notice.

### Middleware Resource Cleanup

`Middleware` is `func(Handler) Handler` — a function type with no lifecycle hook.
Middleware that hold resources (e.g., file handles) cannot close them automatically.
The watcher provides `WithCleanup(fn func() error)` to register cleanup functions
called on `Close()` (after all goroutines/channels are torn down) and cleared on `Reset()`.

Pattern: factory returns `(Middleware, func() error)`; caller pairs them:

```go
mw, closeLog := NewFileLogMiddleware("audit.log")
watcher, _ := New(paths, WithMiddleware(mw), WithCleanup(closeLog))
defer watcher.Close() // closeLog is called automatically
```

`MiddlewareWriteFileLog` wraps `NewFileLogMiddleware` for backward compat but
does NOT close the file — long-lived watchers should use `NewFileLogMiddleware`.

---

## Key Patterns

| Pattern               | Where                                                                             |
| --------------------- | --------------------------------------------------------------------------------- |
| Functional Options    | `options.go` — `type Option func(*Watcher)`                                       |
| Middleware Chain      | `middleware.go` — applied in **reverse** order                                    |
| Filter Composition    | `filter.go` — `FilterAnd()`, `FilterOr()`                                         |
| `resolve*Defaults`    | `middleware.go` — see [Default-guard convention](#default-guard-convention) below |
| `baseDebouncer.stop`  | `debouncer.go` — lock/markStopped/cleanup/unlock/wait in one place                |
| Backend Abstraction   | `backend.go` — `watchBackend` interface; `withBackend()` injects fakes            |
| `newTestWatcher`      | `testing_helpers_test.go:432` — standard `New + cleanup` for all tests            |
| Case-sensitivity      | `filesystem.go` — `pathKey()`, `resolveCaseSensitivity()`, `WithCaseSensitivity`  |
| O(1) path lookup      | `watcher.go` — `watchListKeys` map alongside `watchList` for dedup + O(1) checks  |
| Symlink cycle detect  | `watcher_walk.go` — `handleFollowedSymlink()` + `symlinkVisited` map              |
| Middleware drop track | `watcher_internal.go` — `emitEvent` wraps emit with `atomic.Bool` flag            |
| Path validation       | `watcher.go` — `validateDirExists()` in Add/AddRecursive/Watch                    |
| DropOnFull mode       | `watcher_internal.go` — non-blocking send in `emitEvent` when `eventDropOnFull`   |
