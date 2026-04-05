# Agent Guide: go-filewatcher

**Go 1.26.1** | `github.com/larsartmann/go-filewatcher`

---

## Critical Commands

```bash
just check    # Full quality: tidy, fmt, vet, lint, test
just ci       # Same as check
just lint-fix # Auto-fix linter issues
```

---

## Non-Obvious Conventions

### Error Handling: Standard Library

Uses `errors` and `fmt` from the standard library:

```go
import (
    "errors"
    "fmt"
)

// Creating sentinel errors
var ErrPathNotFound = errors.New("path not found")

// Wrapping with context
return fmt.Errorf("path %q: %w", path, err)

// Checking
if errors.Is(err, ErrPathNotFound) { ... }
```

### Single Package Layout

All code in **root package** (`filewatcher`). No `internal/` or `pkg/` subdirectories — all code lives in the package root.

### File Organization

| File                  | Responsibility                                        |
| --------------------- | ----------------------------------------------------- |
| `watcher.go`          | Public API: New, Watch, Add, Remove, WatchList, Stats |
| `watcher_internal.go` | Event processing: watchLoop, middleware, emitEvent    |
| `watcher_walk.go`     | Directory walking: addPath, walkAndAddPaths           |
| `filter.go`           | All Filter functions                                  |
| `middleware.go`        | All Middleware functions                               |
| `debouncer.go`        | Debouncer + GlobalDebouncer                           |
| `event.go`            | Op type, Event type, JSON/Text marshaling             |
| `errors.go`           | Sentinel errors                                       |
| `options.go`          | Functional options                                    |

---

## Critical Gotchas

### 1. Middleware Order Is Reversed

```go
WithMiddleware(
    MiddlewareRecovery(),   // Runs LAST (innermost)
    MiddlewareLogging(nil), // Runs FIRST (outermost)
)
```

### 2. Two Debounce Modes (Different Semantics)

```go
WithDebounce(d)           // Global: ALL events → ONE callback
WithPerPathDebounce(d)    // Per-path: EACH file → separate callback
```

### 3. Strict Linter: `exhaustruct`

**All struct fields must be initialized** — no zero values allowed:

```go
// WRONG — fails lint
w := &Watcher{fswatcher: fs}

// RIGHT — all fields
w := &Watcher{
    fswatcher: fs,
    paths: paths,
    recursive: true,
    // ... every field
}
```

### 4. Required: `t.Parallel()` in All Tests

```go
func TestXxx(t *testing.T) {
    t.Parallel()  // REQUIRED (enforced by paralleltest linter)
    // ...
}
```

### 5. Event Priority (Multiple Ops)

Create > Write > Remove > Rename — highest wins.

### 6. Chmod Events Ignored

Not mapped to any Op, `convertEvent()` returns `nil`.

### 7. Exported Global with Nolint

```go
//nolint:gochecknoglobals // Intentionally exported for users
var DefaultIgnoreDirs = []string{".git", "vendor", ...}
```

Don't remove the nolint — this is intentional.

---

## Key Patterns

| Pattern            | Where                                          |
| ------------------ | ---------------------------------------------- |
| Functional Options | `options.go` — `type Option func(*Watcher)`    |
| Middleware Chain   | `middleware.go` — applied in **reverse** order |
| Filter Composition | `filter.go` — `FilterAnd()`, `FilterOr()`      |

---

## Linter Cheat Sheet

50+ linters enabled. Key ones that bite:

| Linter             | Rule                                  |
| ------------------ | ------------------------------------- |
| `exhaustruct`      | All struct fields must be initialized |
| `wrapcheck`        | All errors must be wrapped            |
| `paralleltest`     | All tests must use `t.Parallel()`     |
| `gochecknoglobals` | No globals unless `//nolint`          |
| `gci`              | Import order matters                  |

Run `just lint-fix` — it auto-fixes many issues.

---

## Dependencies

```
github.com/fsnotify/fsnotify    # Core file watching
```

Only dependency. Keep it minimal.
