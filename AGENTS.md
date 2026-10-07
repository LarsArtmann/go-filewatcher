# Agent Guide: go-filewatcher

**Go 1.26.7** | `github.com/larsartmann/go-filewatcher/v2` | **MIT License**

> **Companion docs:** [FEATURES.md](./FEATURES.md) (feature inventory) ·
> [ROADMAP.md](./ROADMAP.md) (long-term direction) · [TODO_LIST.md](./TODO_LIST.md)
> (actionable work) · [CHANGELOG.md](./CHANGELOG.md) (release history).
> This file is for enduring, hard-to-discover-from-code context only.

---

## Critical Commands

```bash
# Using Nix flake (recommended)
nix develop              # Enter development shell with Go and tools
direnv allow             # Auto-load environment on cd (requires direnv)

# Nix apps (run from anywhere, no need to be in dev shell)
nix run .#check          # Full quality: vet + lint + test
nix run .#ci             # Full CI: tidy + fmt + vet + lint + test
nix run .#lint-fix       # Auto-fix linter issues
nix run .#test           # Run tests with -race
nix run .#test-v         # Run tests with -race -v
nix run .#lint           # Run linter
nix run .#lint-tests     # Run linter (--tests explicit)
nix run .#bench          # Run benchmarks
nix run .#bench-baseline # Capture benchmark baseline (run from repo root)
nix run .#bench-diff     # Diff benchmarks vs baseline (hermetic benchstat)
nix run .#coverage       # Generate coverage report
nix run .#fmt            # Format Go code (writes — run from repo root)
nix run .#tidy           # Run go mod tidy (writes — run from repo root)
nix run .                # Default = check

# Nix quality gates
nix flake check          # Run all checks (build, test, lint, fmt, vet, examples-build)
nix build .              # Validate reproducible build
nix fmt                  # Format .nix files

# Inside dev shell (aliases are set automatically):
check       # nix run .#check
ci          # nix run .#ci
lint        # nix run .#lint
lint-fix    # nix run .#lint-fix
test        # nix run .#test
```

## Updating vendorHash

When `go.mod` or `go.sum` changes, `vendorHash` in `flake.nix` must be updated:

```bash
# 1. Update dependencies
go get github.com/some/pkg@latest
# or: go mod tidy

# 2. Update vendorHash (Nix will compute the new hash)
nix flake update

# 3. Verify everything still works
nix run .#check
```

If `nix flake update` fails with a hash mismatch, set a temporary placeholder and rebuild:

```bash
# In flake.nix, set vendorHash to an empty string temporarily:
vendorHash = "";  # Will show correct hash in error message

# Then run:
nix build .  # Error will show correct hash

# Copy the hash from the error and set it properly:
vendorHash = "sha256-XXXX...";
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

| File                   | Responsibility                                                                                   |
| ---------------------- | ------------------------------------------------------------------------------------------------ |
| `watcher.go`           | Public API: New, Watch, Add, AddRecursive, Remove, Reset, WatchList, Stats                       |
| `backend.go`           | watchBackend interface + fsnotifyBackend adapter (test seam for fake backend injection)          |
| `watcher_internal.go`  | Event processing: watchLoop, middleware, emitEvent, debugLog, handleError                        |
| `watcher_walk.go`      | Directory walking: addPath, walkAndAddPaths, addBatch, symlink resolution, budget detection      |
| `watcher_gitignore.go` | .gitignore loading and matching: gitignoreCache, shouldSkipByGitignore                           |
| `watcher_selfheal.go`  | Self-healing: selfHealLoop, attemptSelfHeal, failed path tracking                                |
| `watcher_poll.go`      | Polling mode: pollLoop for NFS/FUSE environments                                                 |
| `filesystem.go`        | Filesystem case-sensitivity: FilesystemCaseSensitivity enum, pathKey(), resolveCaseSensitivity() |
| `filter.go`            | All Filter functions + FilterWithMeta and combinators                                            |
| `filter_gogen.go`      | Generated-code detection filter (gogenfilter v3 integration)                                     |
| `middleware.go`        | All Middleware functions (circuit breaker, error batch, correlation, exponential backoff)        |
| `metrics.go`           | PrometheusCollector, StatsFunc, CounterMetric, GaugeMetric                                       |
| `otel.go`              | OTelMiddleware, OTelSpan interface                                                               |
| `debouncer.go`         | Debouncer + GlobalDebouncer                                                                      |
| `event.go`             | Op type, Event type, JSON/Text marshaling                                                        |
| `errors.go`            | Sentinel errors, ErrorCode, ErrorCategory, WatcherError                                          |
| `options.go`           | Functional options (WithGitignore, WithExcludePaths, WithMaxWatches, etc.)                       |
| `phantom_types.go`     | Compile-time phantom types (EventPath, RootPath, DebounceKey, OpString, etc.)                    |

### Examples (`examples/`)

Separate Go programs demonstrating usage. Each subdirectory is a standalone
`main` package that imports the library. Shared helpers (including
`demo.MustWatch`) live in `examples/demo/`. Build with `go run ./examples/<name>`
or `go build ./examples/...`. Part of the same Go module (no separate go.mod).

### Website (`website/`)

Separate Astro + Starlight documentation site with its own `flake.nix`, deployed to
Firebase Hosting at `filewatcher.lars.software`. Not part of the Go module — has its
own `package.json` and Node toolchain. Build with `cd website && nix run .#build`.

The `/changelog` docs page is **generated** from the repo `CHANGELOG.md` by
`website/scripts/sync-changelog.mjs` on every `pnpm build`/`pnpm dev` — never edit
`website/src/content/docs/changelog.mdx` by hand; update `CHANGELOG.md` and rebuild.

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

### 8. WithDebug is Active (not a stub)

`WithDebug(logger)` wires real debug logging throughout the pipeline. The `debugLog` helper checks `w.debug` and calls `w.debugLogger.Debug()`. Log calls are in `watchLoop`, `processEvent`, `emitEvent`, `handleError`, `handleNewDirectory`, and `pollLoop`.

### 9. WithPolling is Active (not a stub)

`WithPolling(true)` starts a `pollLoop` goroutine in `Watch()` that maintains a filesystem snapshot and detects new/modified/removed files at `pollInterval`. Works alongside fsnotify for NFS/FUSE environments.

### 10. Circuit Breaker States

`MiddlewareCircuitBreaker` uses three states: `CircuitClosed` → `CircuitOpen` → `CircuitHalfOpen`. In half-open, only one event passes through to test recovery.

### 11. Graceful ENOSPC Handling

`fswatcher.Add()` errors (including ENOSPC) do not abort the entire walk. Instead:

- The error is logged via `handleError()`
- The `watchErrors` atomic counter is incremented
- Walking continues to add remaining directories
- `Stats.WatchErrors` tracks how many paths failed to add
- This allows the watcher to start in degraded mode instead of failing entirely

### 12. Inotify Budget Awareness

- `maxWatches` is auto-detected from `/proc/sys/fs/inotify/max_user_watches` on Linux
- Override with `WithMaxWatches(n)`
- When budget is exhausted, directories are skipped silently
- `Stats.WatchLimit` and `Stats.WatchBudgetUsed` track budget usage

### 13. .gitignore-Aware Walking

- Enabled by default (`WithGitignore(true)`)
- Loads `.gitignore` files during directory walking
- Directories matching gitignore patterns are skipped (not added to inotify)
- Uses `github.com/sabhiram/go-gitignore` (zero transitive deps)
- gitignore cache is stored per-directory for hierarchical matching
- **Trailing-slash directory patterns don't match the dir itself**: the
  go-gitignore library returns `MatchesPath("Build") == false` for pattern
  `Build/` — it only matches paths _inside_ (`Build/foo`). This means a
  `.gitignore` with `Build/` does NOT prevent the `Build` directory from being
  added to the watch list during walk. Use `Build` (no trailing slash) for
  patterns that need to skip the directory itself.

### 14. Batched Watch Registration

- Directories are collected during walk and added in batches of 1000
- `runtime.Gosched()` is called between batches to yield to event processing
- Reduces startup latency for large directory trees

### 15. Path-Level Exclusions

- `WithExcludePaths(paths...)` excludes absolute paths during walk
- Prefix matching: excluding `/home/user/forks` skips all subdirectories too
- Walk-time only: does not affect event filtering

### 16. Remove() Cleans Up Subdirectories

`Remove(path)` removes all subdirectory watches under the given path, not just
the top-level directory. This prevents watch leaks.

### 17. Reset() Method

`Reset()` clears runtime state while preserving configuration (filters,
middleware, debounce, options). Allows re-calling `Watch()` after `Close()`
without rebuilding from scratch.

### 18. Filesystem Case-Sensitivity Awareness

Different filesystems treat filename case differently: NTFS (Windows) and APFS
(macOS) are case-insensitive, while ext4/XFS/btrfs (Linux) are case-sensitive.
The watcher must be aware of this to avoid:

- **Duplicate watches**: `/dir/MyFile` and `/dir/myfile` on NTFS are the same file
- **Broken Remove()**: calling `Remove("/dir/MyFile")` must match `/dir/myfile`
- **Incorrect exclusion**: `WithExcludePaths("/Build")` must catch `/build/output`
- **Debounce misses**: events for `/File.go` and `/file.go` must coalesce

`WithCaseSensitivity(mode)` configures the mode:

- `CaseSensitivityAuto` (default): resolves per-platform (Windows/macOS → insensitive)
- `CaseSensitive`: paths differing only in case are distinct
- `CaseInsensitive`: paths differing only in case are identical

Internally, `pathKey()` returns a canonical key: NFC-normalized (always), then
lowercased on case-insensitive filesystems. NFC normalization fixes the invisible
mismatch where macOS stores filenames as NFD (decomposed) but user-configured paths
are NFC (composed) — without it, exclude matching, debounce, and gitignore silently
fail on any non-ASCII filename. This key is used for: `watchListKeys` dedup set,
`failedPaths`, `excludePaths`, debounce keys, `Remove()` subtree matching, poll
loop snapshot keys, and gitignore cache keys.

**pathKey performance contract** (measured via `BenchmarkPathKey_*`): NFC is
free for pure-ASCII paths (~26 ns/op, 0 allocs — the common case). Pre-composed
NFC Unicode is allocation-free (~140 ns/op). Only decomposed NFD input — emitted
by the macOS filesystem — allocates (~1 µs/op, 3 allocs, 672 B). This is
negligible for event-driven watching; caching is only worth considering for
extreme macOS throughput (>10k Unicode events/sec). ASCII and NFC paths never
allocate.

**Bench-diff regression verification** (pre-NFC vs post-NFC): zero allocation
regression on ALL hot-path benchmarks (ConvertEvent, PassesFilters, EmitEvent,
ShouldSkipDir, Stats, all Filters, all Middleware). The only allocation change
is +1 alloc in `New()` (one-time startup cost from NFC initialization in
pathKey).

**Bench-diff methodology**: benchmarks MUST run without parallel CPU-intensive
workloads (fuzz tests, builds, etc.). Previous bench-diff runs were invalidated
by running a 5-minute fuzz test (32 workers) simultaneously — the ±18-59% timing
variance was CPU contention, not NFC overhead. Allocation data is always valid
(deterministic regardless of CPU load). Rule: kill all background work before
`nix run .#bench-diff`.

### 19. O(1) Watched-Path Lookup

`watchListKeys` is a `map[string]struct{}` of pathKeys, maintained alongside
`watchList`. This enables:

- O(1) duplicate detection in `tryAddPath` (was O(n) `slices.Contains`)
- O(1) self-heal check in `isPathWatched` (was O(n) `slices.Contains`)

### 20. Symlink Cycle Detection

`WithFollowSymlinks(true)` now includes built-in cycle detection. The walker
tracks resolved symlink targets in `symlinkVisited map[string]struct{}` and
skips any target already visited. This prevents infinite recursion on cycles
(e.g., `/a/b -> /a`). The map is initialized/cleared by `walkAndAddPaths` via
a `topLevel` flag (so recursive calls from `handleFollowedSymlink` reuse the
same set).

### 21. Middleware Drop Tracking

Middleware that drops events (rate limit, dedup, circuit breaker, etc.) is
detected by wrapping the emit function with an `atomic.Bool` flag. If the
middleware chain returns `nil` without calling emit, the event is counted in
`Stats.EventsDroppedByMiddleware`. The `eventsProcessed` counter is now
incremented only when the event actually reaches the event channel (not when
the middleware chain returns nil).

### 22. Path Validation in Add/AddRecursive/Watch

`Add()`, `AddRecursive()`, and `Watch()` now `os.Stat` each path and return
`ErrPathNotFound` or `ErrPathNotDir` immediately for invalid paths. This
prevents non-existent paths from silently entering the self-heal retry loop.
`Remove()` does NOT validate (it's best-effort and should work on already-
deleted paths).

### 23. Syscall Error Classification

`categorizeError` recognizes `os.ErrPermission`, `os.ErrNotExist`, and
`syscall.ENOTDIR` as permanent errors (self-heal abandons them). `syscall.ENOSPC`
is classified as transient (resources may free up). This prevents self-heal
from retrying permission-denied paths forever.

### 24. Polling Mode Respects Exclusions

`pollWalkDir` now applies `shouldSkipDir`, `shouldExcludePath`,
`loadGitignoreForDir`, and `shouldSkipByGitignore` — the same skip logic as
the initial walk. Previously, the poll loop only checked `shouldSkipDir` and
ignored exclusions and gitignore rules.

### 25. Event Channel DropOnFull Mode

`WithEventChannelMode(EventChannelDropOnFull)` changes the channel send logic
in `emitEvent` to use a non-blocking send with a `default` case that increments
`eventsDroppedByBackpressure`. Default is `EventChannelBlocking` (blocking
send, preserves backpressure). The send logic is inlined into `emitEvent`'s
`trackedEmit` closure (not a separate `buildEmitFunc`) to avoid an extra heap
allocation per event.

### 26. Watch Budget Safety Fraction Only Affects Auto-Detected Limits

`WithMaxWatchesSafetyFraction(0.75)` reduces the inotify watch limit to leave
headroom on shared machines. It only applies to **auto-detected** limits
(`WithMaxWatches(0)` / default). Explicit limits set via `WithMaxWatches(n)` are
used as-is — the `maxWatchesExplicit` bool flag tracks this. `Reset()` also
respects explicit settings instead of always re-detecting from
`/proc/sys/fs/inotify/max_user_watches`.

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

### Default-guard convention

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

## CI Workflows

| Workflow               | Purpose                                                                     |
| ---------------------- | --------------------------------------------------------------------------- |
| `ci.yml`               | Test (race + coverage), lint, examples-build, benchmark                     |
| `commitlint.yml`       | Validates PR commit subjects follow conventional-commit format              |
| `docs-consistency.yml` | Checks README.md vs API_STABILITY.md deprecation claims don't drift         |
| `release-please.yml`   | Opens release PRs from conventional commits (automates CHANGELOG + version) |
| `release.yml`          | Triggered on `v*` tags — creates GitHub Release                             |

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

Run `nix run .#lint-fix` — it auto-fixes many issues.

---

## Dependencies

```
github.com/fsnotify/fsnotify          # Core file watching (v1.10.1)
github.com/LarsArtmann/gogenfilter/v3 # Generated code detection (v3.6.1, published — NO local replace)
github.com/sabhiram/go-gitignore      # .gitignore pattern matching (zero transitive deps)
golang.org/x/text/unicode/norm        # NFC Unicode normalization in pathKey (v0.42.0)
golang.org/x/time/rate                # rate.Limiter for MiddlewareThrottle
```

### gogenfilter v3 API

The library depends on `github.com/LarsArtmann/gogenfilter/v3` (v3.6.1, published module — no
local replace directive). The v3 API differs from v0/v2 in these ways (relevant when upgrading consumers):

- `NewFilter` returns `(*Filter, error)` — must handle error
- `WithFilterOptions` returns `(FilterConfig, error)` — must handle error
- `Enabled()` / `Disabled()` removed — auto-enables when configured
- `ShouldFilter` renamed to `Filter` — `f.Filter(path)` returns `(bool, error)`
- Generators: `FilterOapi`, `FilterDeepcopy`, `FilterWire`, `FilterMoq`

**v3.6 sqlc detection semantics** (differs from v3.2): weak filenames like `models.go` are NOT
treated as sqlc-generated on their own (hand-written models.go is too common). Strong `*.sql.go`
patterns still match by filename; the `// Code generated by sqlc. DO NOT EDIT.` content marker is
detected when content is passed (`FilterGeneratedCodeFull` + `ContentCheckEnabled`). Tests in
`filter_gogen_test.go` encode this contract — do not "fix" them back to filename-only matching.

## Named Types (phantom types)

Plain `type X string` named types for compile-time type safety on path-like strings:

| Type           | Purpose                             |
| -------------- | ----------------------------------- |
| `EventPath`    | Event file/directory paths          |
| `RootPath`     | Root directory paths during walking |
| `DebounceKey`  | Debouncer keys                      |
| `LogSubstring` | Log substring assertions (tests)    |
| `TempDir`      | Temp directory paths (tests)        |
| `OpString`     | Operation names                     |

**Usage:** Use constructor functions (e.g., `NewEventPath()`, `NewRootPath()`).

**EventPath has domain methods:** `.Base()`, `.Dir()`, `.Ext()`, `.Join()` for path operations.

---

## Release / CI Gotchas

### release-please needs a repo permission

`release-please.yml` fails with "GitHub Actions is not permitted to create or approve pull
requests" when the repo setting is off. Fix:

```bash
gh api -X PUT repos/LarsArtmann/go-filewatcher/actions/permissions/workflow \
  --input - <<'EOF'
{"default_workflow_permissions":"read","can_approve_pull_request_merge_requests":true}
EOF
```

On PR merge, release-please creates the tag AND the GitHub Release directly. `release.yml`
only fires for manually-pushed `v*` tags (GITHUB_TOKEN-created tags do not re-trigger workflows).

### release-please is path-scoped — website/docs/.github never trigger releases

`release-please.yml` runs in config-file mode (`release-please-config.json` +
`.release-please-manifest.json`). The root package sets
`exclude-paths: ["website", "docs", ".github"]`: commits touching ONLY those paths are
skipped by release-please. Before this existed, a `fix(website)` commit shipped v2.4.1 —
a Go patch release with zero Go changes (2026-10-07). Convention on top: use
`chore(website)` / `docs(website)` / `build(website)` scopes for non-module changes,
never `fix(website)` / `feat(website)`.

### golangci-lint version pin

CI (`ci.yml`) and `release.yml` pin golangci-lint **v2.14.0** — the same version the Nix flake
provides locally. The `.golangci.yml` uses the `exhaustruct_v5` settings key, which only exists
in v2.13+; v2.12 rejects the whole config at validation. When bumping the linter, bump all
three places together (flake, ci.yml, release.yml).

### Release PRs show CI "workflow file issue" startup failures

On release-please PRs, CI, Commitlint, and Docs Consistency fail with `conclusion: failure` and
ZERO jobs ("This run likely failed because of a workflow file issue") — hit both the 2.4.0 and
2.4.1 release PRs (2026-10-06/07). Not a YAML bug: the identical workflow files run green on
Dependabot PRs and on master pushes in the same hour. Treat it as GitHub-side; verify release
content via the master push CI on the merge commit instead. The release is unaffected —
release-please creates the tag and GitHub Release directly on merge.

### go.mod language version

The `go` directive must stay at the CI matrix floor (`go 1.26.x`). The local dev toolchain is
newer (Go 1.27); running go commands can silently bump the directive (2026-09-29 incident broke
CI for a week). If CI fails with "go.mod requires go >= 1.27", restore `go 1.26.7`.

Recurred 2026-10-07 (~1 day of red CI, all open PRs' Test/Lint/Examples failing): a parallel
session's go commands produced the bump and the auto-commit daemon committed and pushed it.
The bump can land on origin even when you ran no go commands yourself — check the `go` line
whenever the daemon commits go.mod, and expect dependent PR CI to recover only after master
is fixed plus a branch update/re-run.

Forensic root cause (2026-10-07, strace-proven): **three BuildFlow tools fight the directive**,
all buildflow-internal (writer TIDs show zero execve). All three are skipped via
`skip_steps` in `.buildflow.yml`:

- `go-version-auto-configure` — wants a major.minor-only directive.
- `go-mod-update` — minor mode bumps the directive to the latest Go release.
- `go-structure-linter` (new in buildflow 202b114) — reports "Go 1.27 is available" as an
  error finding and **auto-repairs by bumping the directive mid-run**, leaving backups in
  `/tmp/go-structure-linter-backups/go.mod.<timestamp>.bak`.

`go-mod-normalize` is exonerated: it writes a candidate downgrade, the dependency-floor
gate (`go mod tidy -diff`) rejects it, and it atomically restores (net-zero; "kept:
downgrade is not dependency-floor-safe" WARN in every run log). Restore command:
`sed -i 's/^go 1\.27$/go 1.26.7/' go.mod`.

**Worktrees and nix**: `nix flake` evaluation IGNORES dirty state in linked worktrees
(drv is not suffixed `-dirty`; it builds the COMMITTED go.mod). In a worktree the floor
must be committed for nix to see it; the main checkout uses the dirty tree as usual.

### Auto-commit daemon races

The daemon commits (and occasionally resets) local master while you work. Before pushing,
`git fetch` and check `git log HEAD..origin/master`; rebase your commits onto origin/master if
they diverge. Never assume your last local commit is still HEAD.

## Known Issues

### Website build (pnpm 11) build-script approvals

Build-script approvals live in `website/pnpm-workspace.yaml` under `allowBuilds:` (`esbuild: true`) — pnpm v11 ignores `pnpm.*` in `package.json` and silently skips unapproved postinstall scripts, so `astro build` then fails on a missing esbuild binary. A placeholder value (e.g. `esbuild: set this to true or false`) silently disables the whole key (cmdguard incident, fixed 2026-09-19).

### Nix sandbox needs GOEXPERIMENT=jsonv2

`event.go` imports `encoding/json/v2` (GOEXPERIMENT-gated). CI sets `GOEXPERIMENT: jsonv2` workflow-wide (ci.yml) and direnv exports it locally, but nix sandbox derivations receive neither, so every Go target fails with "build constraints exclude all Go files in .../encoding/json/v2". The flake sets it wherever Go compiles: `packages.default` and the four `runCommand` checks via `env.GOEXPERIMENT`, and every `mkApp`/`mkBenchApp` script exports it (apps run outside direnv). Any new Go-building derivation or app must set it too (fixed 2026-10-07).

### buildflow nix-hash-fix cannot repair this flake

`nix-hash-fix` has failed 15+/15 runs here: after a hash mismatch it reports "the stale hash was not found verbatim in any .nix file" even when `vendorHash = "sha256-…"` sits verbatim in flake.nix (BuildFlow repo bug; the 2026-10-07 fix was hand-applied after the fixer gave up). Until fixed upstream, update the `vendorHash` in flake.nix to the `got:` hash from `nix build` output by hand, then verify with `nix build`. Consequence: do NOT extract vendorHash into a separate file (the nix-checker suggestion) — it would break the one manual repair path that works.

### buildflow pnpm-audit result cache replays after lockfile-only fixes

The detector result cache keys pnpm-audit on `package.json` but NOT `pnpm-lock.yaml`, so
vulnerabilities fixed via `pnpm audit --fix=update` (lockfile-only change) keep gating the
full pipeline with the stale 9-error finding set even though `pnpm audit` in `website/` is
clean. Surgical purge (the reference's DB path is wrong; the real DB is `cache.db`):

```bash
nix shell nixpkgs#sqlite -c sqlite3 ~/.cache/buildflow/cache.db \
  "DELETE FROM result_cache WHERE value LIKE '%pnpm-audit%';"
```

A no-cache run (`BUILDFLOW_NO_RESULT_CACHE=1`) proves the fix but does NOT overwrite the
stale entry (BuildFlow gap), and it re-executes every step — under high machine load the
re-run hits step timeouts (go killed at spawn) and produces false step failures. Purge the
entry, then re-run the normal cached pipeline. Note `pnpm audit --fix` in pnpm 11 needs an
explicit strategy (`--fix=update` re-resolves the lockfile; `--fix=override` adds
overrides to `website/pnpm-workspace.yaml`). 2026-10-07: 12 vulns (9 high) → 0 via
update + one override (`postcss-selector-parser@<7.1.6: ^7.1.6`, forced major bump,
website build verified green after).
