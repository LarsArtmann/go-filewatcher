# Critical Gotchas (Full Detail)

One-line index lives in `AGENTS.md`; this file carries the full detail,
verbatim from `AGENTS.md` (moved 2026-10-07, carrying-capacity split).

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

### 27. Windows/macOS Test Legs: Inotify-Only Features and POSIX-Shape Assertions

The CI matrix probes windows-2025 and macos-15. Three platform truths keep
those legs green:

- **Watch budgets are inotify-only.** `detectMaxWatches` reads
  `/proc/sys/fs/inotify/max_user_watches`; on macOS/Windows there is no kernel
  cap, so budget enforcement (and `WithMaxWatchesSafetyFraction`) is a no-op
  there. Tests asserting budget behavior need no skips (they test the Linux
  path), but docs must not claim cross-platform budget enforcement.
- **`failedPaths` is keyed by `pathKey()`, never raw paths.** On
  case-insensitive systems the key is lowercased, so a test that looks up
  `failedPaths[subDir]` with a mixed-case raw path works on Linux by accident
  and fails on macOS (found 2026-10-07 via the probe leg). Always look up
  through `watcher.pathKey(path)`.
- **POSIX-shape assertions skip on Windows** via `skipOnWindows(t)` in
  `testing_helpers_test.go`: tests like `TestCleanPath` feed `/a/b/` literals
  and assert forward-slash output — on Windows `filepath.Clean` legitimately
  returns `a\b`. The library behavior is correct per-OS; the assertions are
  POSIX contracts pending per-OS expectations. `ExampleEventPath` is split
  into `example_path_test.go` (!windows) and `example_path_windows_test.go`
  because Example funcs cannot call `t.Skip`.

