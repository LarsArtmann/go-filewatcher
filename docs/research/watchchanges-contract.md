# `WatchChanges(ctx, targetState)` — Contract Sketch

**Status:** Design / not yet implemented. Promote to `TODO_LIST.md` when scoped.
**Last updated:** 2026-07-26

## Problem

Today the watch set is **imperative**: `New(paths...)`, then `Add` / `AddRecursive`
/ `Remove` to mutate it. A sync/backup tool that wants the watch set to mirror a
**dynamic manifest** (config file, remote list of directories to sync) must diff
`WatchList()` against the desired set and call `Add`/`Remove` per path itself.
That diff-and-apply logic is mechanical, easy to get wrong, and belongs in the
library. `WatchChanges` is the **declarative reconciliation** API: declare the
desired watch set once; the watcher converges to it; calling again with the same
state is a no-op.

The closest analogy is "GitOps for file watching" — same idempotent-sync
semantics as `kubectl apply`.

## Goals

- **Idempotent:** `WatchChanges(s)` then `WatchChanges(s)` again changes nothing.
- **Full reconciliation:** after a successful call, `WatchList()` (collapsed to
  targets) equals `targetState`.
- **Reuse existing walk semantics:** gitignore, `excludePaths`, `DefaultIgnoreDirs`,
  recursion/depth, graceful ENOSPC handling, self-heal.
- **Concurrent-safe:** composes with the existing mutex model.
- **Cancellable:** respects `ctx` for long walks over big trees.
- **Observable:** reports what changed and what failed (partial failure is the
  norm on real filesystems).

## Non-goals

- Streaming/long-poll of state changes (possible later: `<-chan ChangeSet`).
- Mutation of filters/middleware (those are construction-time config; the watch
  *set* is what reconciles).
- Watching individual files (always directories, matching current model).

## Proposed types

```go
// WatchTarget describes one desired watch in a target state.
type WatchTarget struct {
	Path     string // resolved the same way as Add/AddRecursive (Abs-normalized)
	MaxDepth int    // -1 = full recursion (default), 0 = shallow dir only, N = depth cap
}

// WatchState is the declarative desired watch set passed to WatchChanges.
// An empty WatchState removes every watch (converges to "watch nothing").
type WatchState struct {
	Targets []WatchTarget
}

// WatchFailure records a path that could not be reconciled.
type WatchFailure struct {
	Path string
	Err  error
}

// ChangeSet reports the result of a reconciliation pass.
type ChangeSet struct {
	Added   []string      // paths newly watched
	Removed []string      // paths no longer watched
	Failed  []WatchFailure // paths that could not be added (ENOSPC, permission denied)
}

// WatchChanges reconciles the watcher's watch set to match target.
// Idempotent: two calls with the same target produce no net change.
// Returns the ChangeSet describing the diff and any per-path failures.
// A non-nil error means reconciliation could not complete (e.g. ctx cancelled,
// watcher closed); the ChangeSet still reports what was applied before the stop.
func (w *Watcher) WatchChanges(ctx context.Context, target WatchState) (ChangeSet, error)
```

## Semantics

1. **Diff, then apply.** Compute `toAdd = target - current`, `toRemove = current - target`.
   Apply removals first (frees inotify budget), then additions.
2. **Depth reuse.** `MaxDepth` mirrors `AddRecursive`: `0` = `Add`, `-1` = full
   walk, `N` = `addPathWithDepth`. No new depth model.
3. **Normalization.** Targets are Abs-normalized exactly like `withResolvedPath`.
   Duplicate targets after normalization collapse to one.
4. **Empty target.** `WatchState{}` removes everything (converge to "watch
   nothing"). This is *not* an error.
5. **Partial failure does not abort.** A path that fails to add (ENOSPC,
   permission denied) is recorded in `ChangeSet.Failed`, increments
   `Stats.WatchErrors`, and is added to `failedPaths` for self-heal — exactly as
   `walkAndAddPaths` behaves today. The call returns the partial `ChangeSet` and
   a **nil** error (failures are per-path, not global). This matches the
   documented "graceful ENOSPC" contract.
6. **Context.** If `ctx` is cancelled mid-walk, stop immediately, return the
   partial `ChangeSet` and `ctx.Err()`.
7. **Self-heal interaction.** Failed targets remain in `failedPaths`; the
   existing `selfHealLoop` retries them on schedule. A target that was removed
   is dropped from `failedPaths` too (no point healing a path we no longer want).
8. **Interaction with `New()` paths.** `WatchChanges` supersedes the initial
   `paths`. Callers who want additive behavior keep their own diff and call
   `Add`/`Remove`; callers who want declarative behavior use `WatchChanges`.

## Interaction with existing options

| Option            | Behavior under WatchChanges                                       |
| ----------------- | ----------------------------------------------------------------- |
| `WithGitignore`   | Still applied during the walk for each added target               |
| `WithExcludePaths`| Still honored (excluded subtrees are never added)                 |
| `WithMaxWatches`  | Budget still enforced; over-budget targets land in `Failed`       |
| `WithSelfHeal`    | Failed targets queued for retry as today                          |
| Filters/middleware| Unaffected — `WatchChanges` only reconciles the *watch set*       |

## Why `MaxDepth` per-target, not per-call

A manifest often mixes shallow and deep targets (watch `./config` shallowly but
`./src` recursively). Per-target depth matches `AddRecursive`'s vocabulary and
avoids a second knob. A package-level helper can lift a flat `[]string` into a
uniform-depth `WatchState`:

```go
func ShallowState(paths ...string) WatchState { /* MaxDepth=0 for each */ }
func RecursiveState(paths ...string) WatchState { /* MaxDepth=-1 for each */ }
```

## Open questions (resolve before implementation)

1. **Collapse reporting granularity.** `Added` is per-directory (post-walk) or
   per-target? Per-directory is honest but noisy for big trees; per-target loses
   visibility. Recommendation: per-directory (mirrors `WatchList`), with a
   `AddedRoots []string` summary of target roots added.
2. **Closed-watcher semantics.** Return `ErrWatcherClosed` immediately, matching
   `Add`/`Remove`. (Aligned with `checkClosedOp`.)
3. **Not-watching-yet.** Can `WatchChanges` be called before `Watch()`? Useful
   for setup. Recommendation: yes — it mutates `watchList` regardless of run
   state, same as `Add`.
4. **Streaming variant.** A future `<-chan ChangeSet` driven by an external
   state source (file watcher on the manifest itself) would let the watch set
   track a live config. Out of scope for v1; note in ROADMAP.
5. **Depth comparison.** When a path is already watched at depth 2 but the new
   target says depth 5, do we re-walk to add the deeper levels? Recommendation:
   yes — reconcile depth too, not just presence.

## Example usage

```go
w, _ := filewatcher.New(nil) // start empty
events, _ := w.Watch(ctx)

// Declarative: mirror a manifest read from disk.
manifest := loadSyncDirs() // []string
state := filewatcher.RecursiveState(manifest...)

cs, _ := w.WatchChanges(ctx, state)
log.Printf("added %d, removed %d, failed %d", len(cs.Added), len(cs.Removed), len(cs.Failed))

// Manifest changed later — converge again. Idempotent for unchanged paths.
state = filewatcher.RecursiveState(loadSyncDirs()...)
w.WatchChanges(ctx, state)
```

## Implementation outline (post-approval)

- Reuse `withResolvedPath` + `addPath` / `walkAndAddPaths` / `addPathWithDepth`.
- New unexported `reconcile(ctx, targets)` holding the diff/apply loop, under
  `w.mu`. Emit `Added`/`Removed`/`Failed` as it goes.
- Tests: idempotency (double-call no-op), add/remove/keep mix, empty-state clears
  all, partial-failure collection, ctx-cancellation partial result,
  closed-watcher error, depth reconciliation, normalization/dedup.
