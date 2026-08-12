# Reliability Improvements for Edge Cases — All Consumers

**Date:** 2026-08-12
**From:** Crush review → go-filewatcher
**Version reviewed:** v2.3.0
**Scope:** changes that benefit every consumer of the library, not just the PMA use case.

---

## Executive Summary

go-filewatcher is already a sophisticated, well-tested library. The code paths that are exercised daily (filters, debounce, batched walk, ENOSPC self-heal, case-insensitive/NFC keys) are solid. The gaps that remain are mostly in **less-common but high-impact combinations**: polling mode with exclusions, symlink cycles, slow-consumer backpressure, middleware drop observability, and a few small bugs that can panic or silently misbehave. This document lists concrete, actionable fixes.

Each item includes:
- the consumer-visible symptom,
- the root cause,
- the suggested change with file references,
- the proposed default behavior (preserving backward compatibility where possible).

---

## 1. Polling Mode Ignores Walk-Time Exclusions

**Symptom:** A consumer configures `WithExcludePaths("/home/user/forks")` and `WithPolling(true)`. Native fsnotify events correctly ignore `/forks`, but the poll loop still walks it and emits `Create`/`Write`/`Remove` events for it.

**Root cause:** `pollWalkDir` only calls `shouldSkipDir(d.Name())` (`watcher_poll.go:67`). It never checks `shouldExcludePath` or `shouldSkipByGitignore`. It also walks `.git`, `node_modules`, and any gitignored subtrees.

**Suggested fix:** Make `pollWalkDir` use the same skip logic as the initial walk:

```go
// internal/watcher_poll.go
func (w *Watcher) pollWalkDir(rootPath string, snapshot map[string]fileState) {
    _ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return nil
        }

        if d.IsDir() {
            if w.shouldSkipDir(d.Name()) ||
               w.shouldExcludePath(path) ||
               w.shouldSkipByGitignore(path) {
                return filepath.SkipDir
            }
        }
        // ...
    })
}
```

Load `.gitignore` for each directory visited by calling `w.loadGitignoreForDir(path)` before the skip check, mirroring `walkDirFunc` (`watcher_walk.go:156-160`).

**Backward compatibility:** This is a bug fix — consumers who explicitly excluded paths would expect the exclusions to apply in all modes. It should be safe to change without an opt-in.

---

## 2. Polling Mode Emits Duplicate Events for Native fsnotify Changes

**Symptom:** With `WithPolling(true)`, a single file change can produce two events: one from fsnotify and one from the poll loop. This breaks consumers that rely on exactly-once semantics, even if they add deduplication middleware.

**Root cause:** The poll loop and fsnotify run independently. There is no mechanism for the poll loop to suppress changes that fsnotify already reported. The deduplication middleware exists but is not used by default and cannot dedupe across different event paths (e.g., `Create` vs `Write` for the same file).

**Suggested fix (minimum):** Document the limitation clearly in `Troubleshooting.md` and the `WithPolling` option doc: "Polling is a fallback, not a duplicate-suppression layer. When both backends are active, consumers should use `MiddlewareDeduplicate` or rely on idempotent handling."

**Suggested fix (better):** Give the poll loop a "last native event timestamp" per path. When `watchLoop` processes an fsnotify event, record `pathKey` + `time.Now()` in a small LRU/sync.Map. When the poll loop detects a change, skip emitting if the same path had a native event within one poll interval. This is a heuristic, but it removes the common double-fire case.

**Backward compatibility:** The deduplication heuristic should be opt-in via a new option such as `WithPollDeduplicate(true)` to avoid surprising consumers who expect the poll loop to report everything.

---

## 3. Poll Loop Does Not Detect Renames

**Symptom:** A consumer on NFS/FUSE relies on polling. A file is renamed from `a.txt` to `b.txt`. The poll loop emits `Remove a.txt` + `Create b.txt`, but never a `Rename` event. Consumers that batch renames specially (e.g., git tooling) see two unrelated events.

**Root cause:** `pollDetectChanges` only compares `modTime`/`size` for modifications and `existence` for creates/removes (`watcher_poll.go:100-126`). It has no inode-tracking mechanism.

**Suggested fix:** Add `ino` to `fileState` (from `os.Stat` on Linux/BSD, or `FileInfo` via `Sys()`). When a file disappears and a new file appears with the same inode in the same parent directory, emit a `Rename` event from the old path to the new path instead of separate `Remove` + `Create`. This is platform-specific and may be unavailable on some filesystems; fall back to the current behavior when inode is zero or unavailable.

**Backward compatibility:** Emitting `Rename` instead of `Remove`+`Create` is a behavioral change. Gate it behind a new option such as `WithPollDetectRenames(true)` so existing consumers are not affected.

---

## 4. `WithDebug(nil)` Can Panic

**Symptom:** A consumer writes `filewatcher.WithDebug(nil)` and then enables debug mode. The watcher panics on the first debug log.

**Root cause:** The option doc says "If logger is nil, log/slog.Default() is used" (`options.go:206-209`), but the implementation simply stores `nil`:

```go
func WithDebug(logger *slog.Logger) Option {
    return func(w *Watcher) {
        w.debug = true
        w.debugLogger = logger  // nil if caller passed nil
    }
}
```

`debugLog` then calls `w.debugLogger.Debug(...)` without a nil check (`watcher_internal.go:21-25`).

**Suggested fix:** Honor the documented contract:

```go
func WithDebug(logger *slog.Logger) Option {
    return func(w *Watcher) {
        w.debug = true
        if logger == nil {
            logger = slog.Default()
        }
        w.debugLogger = logger
    }
}
```

Also add a nil guard in `debugLog` as defense in depth:

```go
func (w *Watcher) debugLog(msg string, args ...any) {
    if w.debug && w.debugLogger != nil {
        w.debugLogger.Debug(msg, args...)
    }
}
```

**Backward compatibility:** Pure bug fix. Existing code that passes a non-nil logger is unaffected.

---

## 5. Middleware Drops Are Not Visible in `Stats`

**Symptom:** A consumer uses `MiddlewareRateLimit`, `MiddlewareThrottle`, `MiddlewareDeduplicate`, or `MiddlewareCircuitBreaker` and later calls `Stats()`. `EventsFilteredOut` only counts filter drops, not middleware drops. The reported number of processed events is higher than what actually reached the consumer, making it hard to debug lost events.

**Root cause:** `eventsFilteredOut` is only incremented in `passesFilters` (`watcher_internal.go:90`). Middleware that returns `nil` to drop events does not increment any counter.

**Suggested fix:** Add a `MiddlewareEventsDropped` atomic counter and increment it whenever a middleware explicitly drops an event. For example:

- In `rateLimiterMiddleware`: when `!limiter.Allow()` (`middleware.go:112-113`).
- In `MiddlewareDeduplicate`: when a duplicate is detected (`middleware.go:236-240`).
- In `MiddlewareCircuitBreaker`: when the circuit is open or half-open probe already sent (`middleware.go:581-596`).
- In `MiddlewareExponentialBackoff`: when `now.Before(state.dropUntil)` (`middleware.go:865`).
- In `MiddlewareBatch`: when batch size is exceeded and the flushed batch is consumed, but the current event is still emitted? Actually batching doesn't drop; note that it currently emits the event even after flushing (`middleware.go:324`). That is a separate issue (see item 6).

Expose the new counter in `Stats` as `EventsDroppedByMiddleware uint64`.

**Backward compatibility:** Additive field in `Stats`. No breaking change.

---

## 6. `MiddlewareBatch` Emits the Triggering Event After the Flush

**Symptom:** A consumer uses `MiddlewareBatch(window, maxSize, flush)` with `maxSize=10`. When the 10th event arrives, the first 9 are flushed, but the 10th event is also immediately emitted to the next handler via `next(ctx, event)` (`middleware.go:324`). The consumer receives the batch via `flush` and the individual event via the normal channel.

**Root cause:** The batch-full branch flushes the batch but then unconditionally calls `next(ctx, event)`.

**Suggested fix:** The batch-full branch should not emit the individual event; the event is already in the flushed batch. Remove the `return next(ctx, event)` call and return `nil` after the flush. The timer-expiry branch already does this correctly (it does not call `next`).

**Backward compatibility:** This is a bug fix. The current behavior is a duplicate emission that no consumer would rely on intentionally. However, because it is a behavior change, call it out in the changelog and consider a minor version bump.

---

## 7. `MiddlewareBatch` Timer Flush Errors Are Not Reported to the Error Handler

**Symptom:** The timer-based flush in `MiddlewareBatch` fails (e.g., network destination unavailable), but the error is only logged with `slog.Error` (`middleware.go:339-343`). It does not propagate to the watcher's configured error handler or `Errors()` channel.

**Root cause:** The timer goroutine cannot return an error through the middleware `Handler` interface, so it falls back to logging.

**Suggested fix:** Two options:

1. **Capture the error and re-emit it on the next event:** store the last flush error in `batchState`. When the next event arrives, if a flush error is pending, return it to the caller. This routes the error through `wrapWithMiddleware` → `handleError`.
2. **Expose a batch-specific error channel or callback:** add an optional `onFlushError` parameter. This is more explicit but adds API surface.

Option 1 is simpler and fits the existing error model. The timer goroutine sets `state.flushErr = err`; the next event handler returns it; `wrapWithMiddleware` dispatches it to `handleError`.

**Backward compatibility:** No API change. Errors that were previously swallowed now reach the error handler.

---

## 8. `Errors()` Channel Can Drop Errors Silently

**Symptom:** A consumer reads errors from `Errors()`. Under high error load (e.g., many ENOSPC errors during a big tree walk), the channel fills and subsequent errors are dropped. The consumer has no way to know how many errors were lost.

**Root cause:** `handleError` sends non-blockingly (`watcher_internal.go:278-285`):

```go
select {
case w.errorsCh <- err:
default:
    // Channel is full or closed, drop the error
}
```

**Suggested fix:** Add an `ErrorsDropped` counter to `Stats` and increment it in the `default` branch. This gives consumers a way to detect that they need to increase `WithBuffer` or read errors more frequently.

Optionally, expose `WithErrorBufferSize(int)` to decouple the error channel size from the event channel size. Today both use `bufferSize`.

**Backward compatibility:** Additive counter and optional option. No breaking change.

---

## 9. Symlink Following Can Loop and Duplicate Events

**Symptom:** A consumer enables `WithFollowSymlinks(true)`. A directory contains a symlink to an ancestor or to another already-watched directory. The watcher can enter infinite recursion or emit duplicate events for the same physical directory under different paths.

**Root cause:** `walkDirFunc` resolves a symlink and calls `walkAndAddPaths(NewRootPath(resolved))` recursively (`watcher_walk.go:130-146`). It does not track resolved targets. `filepath.WalkDir` does not detect cycles through symlinks because it is given a real directory path each time.

**Suggested fix:** Add a `resolvedSymlinkTargets map[string]struct{}` to the walker state (or to `Watcher` for the duration of the walk). Before recursing into a resolved symlink target, check if the canonical path is already being walked. If yes, skip it and log a debug message. Use `pathKey(resolved)` as the key so it respects case-insensitivity and NFC normalization.

Also deduplicate watch-list entries for symlink targets that are already watched under their real path. If `/a` and symlink `/b -> /a` are both configured, only watch `/a` and document that events are reported once.

**Backward compatibility:** Add cycle detection as the default behavior for `WithFollowSymlinks(true)`. This is a safety improvement; the only consumers affected are those accidentally relying on loops, which is unlikely.

---

## 10. Newly Created Directories Are Watched Even When the Event Was Filtered Out

**Symptom:** A consumer filters out `Create` events in a specific subdirectory. A new directory is created there. Despite the filter, the watcher adds the directory to fsnotify and starts watching it.

**Root cause:** `handleFilteredEvent` calls `handleNewDirectory` for any `Create` event that did not pass filters (`watcher_internal.go:102-105`). The intent is to keep the watch list up to date even for filtered directories, but it can surprise consumers who expect filters to be authoritative.

**Suggested fix:** Provide a new option `WithWatchFilteredDirectories(bool)` defaulting to `true` (current behavior) for backward compatibility. When `false`, `handleFilteredEvent` does not add new directories. This lets consumers that truly want to ignore certain subtrees avoid inotify budget being spent on them.

**Backward compatibility:** Default preserves existing behavior. New option is opt-in.

---

## 11. `FilterExcludePaths` and `FilterGitignore` Are Sensitive to Path Form

**Symptom:** A consumer passes a relative path to `FilterGitignore(repoRoot)` or `FilterExcludePaths(paths...)`. Events for the same files are not filtered because the event path is absolute while the configured path is relative.

**Root cause:** `FilterExcludePaths` stores `cleanPath(path)` but compares to `event.Path` which is always absolute (`filter.go:107-117`). `FilterGitignore` computes `filepath.Rel(repoRoot, event.Path)` without first making `repoRoot` absolute (`filter.go:457-463`).

**Suggested fix:** In `FilterExcludePaths`, normalize configured paths to absolute with `filepath.Abs` (or `cleanPath` after making absolute). In `FilterGitignore`, make `repoRoot` absolute before computing the relative path. If `filepath.Abs` fails, fall back to the current behavior and skip the normalization.

**Backward compatibility:** Bug fix. Consumers who already pass absolute paths are unaffected. Consumers who pass relative paths will see more consistent behavior.

---

## 12. `MiddlewareDeduplicate` Does Not Normalize Path Keys

**Symptom:** On macOS, a file named `café.txt` (NFC) emits a `Write` event. The filesystem reports it as `cafe\u0301.txt` (NFD). A second event for the same file does not get deduplicated because the dedupe key is the raw `event.Path` string.

**Root cause:** `MiddlewareDeduplicate` uses `event.Path` directly as the map key (`middleware.go:217`). It does not apply the same NFC normalization or case-folding that `pathKey()` uses for watch-list deduplication.

**Suggested fix:** Use `watcher.pathKey(event.Path)` (or a helper that does not require a watcher instance) to build the dedupe key. This ensures deduplication matches the watcher's canonical path representation. If the middleware is not tied to a watcher, replicate the `norm.NFC.String(...)` logic plus optional case-folding based on a mode parameter.

**Backward compatibility:** This is a bug fix on macOS and case-insensitive filesystems. It may cause fewer duplicate events, which is the desired behavior. No breaking API change.

---

## 13. Slow Consumers Can Block the Entire Watch Loop

**Symptom:** A consumer reads events slowly. The event channel fills. The `watchLoop` goroutine blocks on `eventCh <- e` (`watcher_internal.go:138`), which stops processing new fsnotify events. Fsnotify's internal queue overflows and events are dropped by the kernel, with no indication to the consumer.

**Root cause:** `buildEmitFunc` uses a blocking select with `w.done` and `ctx.Done()` as the only escape paths. There is no timeout or drop-on-full behavior.

**Suggested fix:** Add a configurable `WithEventChannelMode(mode)` option with at least two modes:

- `Blocking` (current default): preserve backpressure semantics.
- `DropOnFull`: if the channel is full, drop the event and increment a `EventsDroppedByBackpressure` counter in `Stats`. Optionally log a single warning per burst.

For `DropOnFull`, use a non-blocking send:

```go
select {
case eventCh <- e:
default:
    w.eventsDroppedByBackpressure.Add(1)
    w.debugLog("event dropped: channel full", slog.String("path", e.Path))
}
```

This is valuable for consumers that prefer losing events over blocking the watcher.

**Backward compatibility:** Default remains blocking. New mode is opt-in.

---

## 14. Error Classification Does Not Recognize Syscall Errors

**Symptom:** `selfHealLoop` retries paths that will never succeed (e.g., permission denied on a directory, or a path that is no longer a directory). It retries them forever because the error is categorized as `CategoryUnknown` instead of `CategoryPermanent`.

**Root cause:** `categorizeError` only matches sentinel errors (`errors.go:217-252`). It does not inspect `errors.Is(err, os.ErrPermission)`, `errors.Is(err, os.ErrNotExist)`, or `errors.Is(err, syscall.ENOTDIR)`. `selfHeal` uses `IsPermanent()` to abandon paths, but `CategoryUnknown` is not permanent.

**Suggested fix:** Extend `categorizeError` to recognize common syscall errors:

```go
if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
    return CategoryPermanent
}
```

Also consider `errors.Is(err, syscall.ENOSPC)` as transient (it may resolve when resources free up), which is already handled by continuing retries.

**Backward compatibility:** Better classification means `selfHeal` abandons truly permanent failures faster. This is a reliability improvement, not a breaking change.

---

## 15. `WithMaxWatches` Uses System-Wide Total, Not Remaining Budget

**Symptom:** A consumer on a shared machine sets `WithMaxWatches(0)` (auto-detect). The library reads `/proc/sys/fs/inotify/max_user_watches` and uses the full value as its limit. Other processes already own many watches, so the library exhausts the real remaining budget and gets `ENOSPC` errors for many paths.

**Root cause:** `detectMaxWatches` returns the system limit, not the remaining budget. There is no portable way to query per-process inotify usage, but the library can at least be more conservative.

**Suggested fix:** Provide a `WithMaxWatchesSafetyFraction(float64)` option that defaults to `0.75`. The effective limit becomes `int(float64(maxWatches) * fraction)`. This leaves headroom for other processes and the kernel. Add a `Stats` field `WatchBudgetCap` showing the computed cap.

**Backward compatibility:** Changing the default fraction would reduce the number of watches the library attempts. To avoid surprising consumers, default to `1.0` (current behavior) and document that `0.75` is recommended for shared or multi-tenant environments. Alternatively, default to `0.75` and call it out as a behavioral change in a minor release.

---

## 16. `AddRecursive` and `Add` Do Not Validate That the Path Exists

**Symptom:** A consumer calls `watcher.Add("/nonexistent")` or `watcher.AddRecursive("/nonexistent", -1)`. The path is silently added to `failedPaths` and retried by self-heal forever, even though the path will never exist.

**Root cause:** `withResolvedPath` only calls `normalizePath`, not `os.Stat` (`watcher.go:174-189`). The underlying `fsnotify.Add` fails, but the failure is treated as transient.

**Suggested fix:** In `withResolvedPath`, after resolving the path, stat it. If it does not exist or is not a directory, return `ErrPathNotFound` or `ErrPathNotDir` immediately. These are already permanent errors, so self-heal will not retry them.

```go
abs, resolveErr := normalizePath(path)
if resolveErr != nil { ... }

info, statErr := os.Stat(abs)
if statErr != nil {
    return fmt.Errorf("%w: path %q", ErrPathNotFound, abs)
}
if !info.IsDir() {
    return fmt.Errorf("%w: path %q", ErrPathNotDir, abs)
}

return fn(abs)
```

**Backward compatibility:** Bug fix. Invalid paths now fail fast instead of entering the retry loop. This is more reliable and less surprising.

---

## 17. `WithIgnoreDirs` Walk-Time Skipping Is Case-Sensitive

**Symptom:** On a case-insensitive filesystem (APFS, NTFS), a directory is named `BUILD` or `Node_Modules`. The walk-time `shouldSkipDir` does not match it because it compares against the lowercase `DefaultIgnoreDirs` and user-provided names exactly (`watcher_walk.go:180-191`). The directory is added to inotify and consumes budget.

**Root cause:** `shouldSkipDir` does not apply the effective case-sensitivity mode to directory-name matching. The event-time `FilterIgnoreDirs` wrapper does not help here because the directory is already watched.

**Suggested fix:** In `shouldSkipDir`, when `effectiveCaseSensitivity == CaseInsensitive`, compare lowercased `name` against lowercased entries in `DefaultIgnoreDirs` and `ignoreDirNames`. Keep the case-sensitive fast path for case-sensitive filesystems.

**Backward compatibility:** Bug fix on case-insensitive filesystems. On case-sensitive filesystems, behavior is unchanged.

---

## 18. `FilterIgnoreDirs` Event-Time Matching Does Not Handle Case-Only Variants

**Symptom:** A consumer on macOS uses `WithIgnoreDirs("node_modules")`. A file event arrives for `.../Node_Modules/foo.js`. The event is not filtered because the event-time `FilterIgnoreDirs` does exact string matching (`filter.go:71-88`).

**Root cause:** Same as item 17, but at the filter layer. The event path is compared raw.

**Suggested fix:** In `FilterIgnoreDirs`, normalize both the configured directory names and the path components using the same case/NFC logic as `pathKey()`. Because filters are not tied to a watcher, consider adding a `FilterIgnoreDirsCaseInsensitive(dirs...)` variant or a helper that normalizes based on `runtime.GOOS`. The simpler fix is to make the existing `FilterIgnoreDirs` case-insensitive for directory-name matching, since directory names like `node_modules` and `vendor` are conventionally lowercase on all platforms.

**Backward compatibility:** This changes matching for case-sensitive filesystems too. To be safe, add a new `FilterIgnoreDirsCaseInsensitive` filter and update `WithIgnoreDirs` to use the case-sensitive version by default (current behavior). Document the new option for consumers on case-insensitive filesystems.

---

## 19. Content Hashing and Generated-Code Detection Read Files Synchronously in the Event Loop

**Symptom:** A consumer enables `WithContentHashing()` or `FilterGeneratedCodeFull(ContentCheckEnabled)`. A burst of large file events causes the watcher to block on disk I/O, delaying all subsequent events.

**Root cause:** `hashFile` (`filter.go:419-442`) and `FilterGeneratedCodeFull` content mode (`filter_gogen.go:75-82`) call `os.Open`/`os.ReadFile` synchronously from the `watchLoop` goroutine. There is no size pre-check in `FilterGeneratedCodeFull` (unlike `hashFile`'s 10 MiB cap) and no async I/O.

**Suggested fix:**
1. Add a `WithContentHashMaxSize(bytes int64)` option and use it in `convertEvent`/`hashFileContents`. Default to 10 MiB (current behavior) to preserve compatibility.
2. For `FilterGeneratedCodeFull` with `ContentCheckEnabled`, add a similar max-size cap and skip content checks for files larger than the cap.
3. Consider adding a warning in the docs: content hashing and content-based detection are not suitable for high-throughput or large-file scenarios. For large files, use filename-only detection.

**Backward compatibility:** Default caps match current behavior. New options are opt-in.

---

## 20. `MiddlewareWriteFileLog` Can Leak File Descriptors on Long-Running Watchers

**Symptom:** A consumer uses `MiddlewareWriteFileLog("audit.log")` with a long-lived watcher or a watcher that calls `Reset()` periodically. The file handle is opened lazily and never closed by the library.

**Root cause:** `MiddlewareWriteFileLog` returns only the middleware, not the closer (`middleware.go:428-431`). The docs already warn about this and recommend `NewFileLogMiddleware + WithCleanup`.

**Suggested fix:** Deprecate `MiddlewareWriteFileLog` more strongly. Add a runtime warning (log to stderr once) when it is used, directing consumers to `NewFileLogMiddleware`. In v3, remove it entirely.

**Backward compatibility:** Warning only; no behavior change.

---

## 21. `Reset()` Does Not Preserve `failedPaths` or Self-Heal State

**Symptom:** A consumer watches a large tree, hits ENOSPC, enables self-heal, then calls `Reset()` and `Watch()` again. The failed paths are forgotten, so the watcher starts from scratch instead of prioritizing previously failed paths.

**Root cause:** `Reset()` clears `watchList`, `watchListKeys`, `failedPaths` is not reset explicitly but the struct is not touched after creation. Actually `failedPaths` is not reset in `Reset()` (it is initialized in `New`). Wait, looking at `Reset()` (`watcher.go:639-697`): it resets `watchList`, `watchListKeys`, `done`, `eventCh`, `errorsCh`, `cleanups`, metrics, debouncer, gitignore cache, and re-detects max watches. It does **not** clear `failedPaths`. That means after Reset, `failedPaths` still contains entries from the previous run. Is that good or bad? The old failed paths may no longer be valid. But the paths were never removed, so they are likely still valid. However, `Watch()` calls `addPath` for each configured root, which will re-walk and add paths. The old `failedPaths` entries are not used until `selfHealLoop` runs. If a path that was previously failed is now added successfully, `tryAddPath` deletes it from `failedPaths`. So stale entries are cleaned up. This seems acceptable, not a bug. Maybe skip this item.

Actually, `Reset()` creates a new `fsnotify.Watcher`, but the old fswatcher is closed. The `failedPaths` map retains old path strings. If the configured root paths changed via options? No, options preserve `w.paths`. So stale entries are for paths that were originally under the roots. They may or may not be visited again. If they are not visited, self-heal will retry them. That could be useful (recovering from ENOSPC after Reset) or wasteful (if the path was removed). Not a major issue.

I'll remove this item or keep it as a minor observation. Let's keep it as a suggestion to clear failedPaths on Reset or document that it is retained.

Actually, I'll remove it to avoid overloading. There are plenty of other items.

---

## 22. `WatcherError.Stack` Captures the Stack at Error Creation, Not Error Origin

**Symptom:** A consumer receives a `WatcherError` with a stack trace that points to `NewWatcherError` rather than to the actual line where the underlying error occurred.

**Root cause:** `NewWatcherError` calls `debug.Stack()` at the time it wraps the error (`errors.go:207-214`). The stack is from the wrapper, not from the original failure site.

**Suggested fix:** Accept an optional stack trace parameter or use `runtime.Caller` to skip the wrapper frames. For example, add `NewWatcherErrorWithStack(op, path, err, stack []byte)` and have the caller capture the stack at the original failure site. Alternatively, document the current behavior: the stack shows where the `WatcherError` was created, which is useful for tracing the error-handling path, not the underlying syscall failure.

**Backward compatibility:** Keep existing constructor; add a new constructor. No breaking change.

---

## 23. `Watch()` Does Not Re-check Path Existence

**Symptom:** A watcher is created with paths that exist at `New()` time. Between `New()` and `Watch()`, a path is deleted. `Watch()` attempts to add it and fails with a generic fsnotify error instead of a clear `ErrPathNotFound`.

**Root cause:** `Watch()` calls `addPath` for each configured root without re-statting them (`watcher.go:338-343`).

**Suggested fix:** In `Watch()`, before adding each root, verify it still exists with `os.Stat`. If not, return a wrapped `ErrPathNotFound` with the path name. This gives consumers a clear error for race conditions between configuration and start.

**Backward compatibility:** Bug fix. Paths that were deleted before `Watch()` now fail with a clear error instead of a backend-specific one.

---

## 24. Documentation Gaps for Consumers

**Symptom:** Consumers discover edge-case behavior through production failures rather than docs.

**Suggested additions to `Troubleshooting.md`:**

1. **Polling mode and exclusions:** explain that `WithExcludePaths` and `.gitignore` filtering apply to fsnotify only unless the polling fixes from item 1 are implemented. If not implemented, document the current limitation.
2. **Symlink cycles:** warn that `WithFollowSymlinks(true)` can follow cycles and that consumers should avoid symlinks to ancestors.
3. **Slow consumers and channel backpressure:** explain that the event channel is blocking by default and that a slow consumer can cause event loss at the kernel level.
4. **Middleware drops and metrics:** explain that `EventsFilteredOut` does not include middleware drops, and point to `EventsDroppedByMiddleware` once added.
5. **Content hashing I/O:** warn that `WithContentHashing()` reads every matched file synchronously and is not suitable for large files or high-throughput scenarios.

**Backward compatibility:** Docs only.

---

## 25. Testing Gaps to Close

The following scenarios are either not tested or only tested on Linux. Closing them would catch regressions in the items above:

| Scenario | Where to add | Priority |
| --- | --- | --- |
| Polling loop with `WithExcludePaths` | `watcher_poll_test.go` (new) | High |
| Polling loop double-event with native fsnotify | `watcher_poll_test.go` | Medium |
| Symlink cycle in `WithFollowSymlinks` | `watcher_walk_test.go` | High |
| Symlink to already-watched real path | `watcher_walk_test.go` | Medium |
| `WithDebug(nil)` does not panic | `options_test.go` | High |
| `MiddlewareBatch` does not emit individual event on full batch | `middleware_test.go` | High |
| `MiddlewareDeduplicate` with NFD/NFC paths | `middleware_test.go` | Medium |
| `Errors()` channel drops counted in `Stats` | `watcher_internal_test.go` | Medium |
| `FilterGitignore` with relative `repoRoot` | `filter_test.go` | Medium |
| `Add`/`AddRecursive` with non-existent path | `watcher_test.go` | High |
| Case-insensitive `WithIgnoreDirs` walk-time skip | `watcher_walk_test.go` | Medium |
| `selfHeal` abandons permission-denied paths | `watcher_selfheal_test.go` | High |
| macOS CI matrix for case/NFC behavior | `.github/workflows/ci.yml` | High (per `TODO_LIST.md`) |
| Windows CI matrix | `.github/workflows/ci.yml` | Medium (per `TODO_LIST.md`) |

---

## Suggested Sequencing

**Phase 1 — correctness fixes (patch release):**
- Items 1, 4, 6, 11, 12, 16, 23.
- Add tests for each.

**Phase 2 — observability improvements (minor release):**
- Items 5, 8, 15 (with default 1.0 for safety).
- Add `Stats` fields and update `PrometheusCollector`/`metrics.go`.

**Phase 3 — advanced edge cases (minor release):**
- Items 2, 3, 9, 10, 13, 17, 18, 19.
- These require new options or more invasive changes; implement with opt-in defaults.

**Phase 4 — docs and CI:**
- Items 24, 25.

---

## Bottom Line

The library is production-ready for the common path. The most important fixes for all consumers are:

1. **Apply exclusions and gitignore to polling** (item 1) — otherwise polling mode silently violates the consumer's filtering contract.
2. **Fix `WithDebug(nil)` panic** (item 4) — a trivial bug that violates the documented API.
3. **Fix `MiddlewareBatch` duplicate emission** (item 6) — middleware consumers are currently receiving double events.
4. **Add cycle detection for symlinks** (item 9) — currently a real foot-gun with infinite-walk potential.
5. **Validate paths in `Add`/`AddRecursive` and `Watch`** (items 16, 23) — prevent permanent retry loops and give clear errors.
6. **Classify syscall errors as permanent** (item 14) — stop self-heal from retrying permission-denied paths forever.

These six changes are high-impact, low-surprise, and should be the top priority.
