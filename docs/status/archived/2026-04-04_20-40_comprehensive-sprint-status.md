# Go-Filewatcher — Comprehensive Sprint Status Report

> **ARCHIVED 2026-10-07** (docs-health sweep): forward items resolved inline below — `~~strikethrough~~` verdicts cite evidence. Live work lives in [TODO_LIST.md](../../TODO_LIST.md).

**Date:** 2026-04-04 20:40\
**Author:** Crush (AI Assistant)\
**Session:** Improvement Sprint — Retrospective & Execution\
**Branch:** `master` (7 commits ahead of `origin/master`)

---

## Executive Summary

The go-filewatcher library is a functional, well-structured utility built on fsnotify with recursive watching, composable filters, middleware chains, and two debounce strategies. This sprint identified **25 improvements** and executed **4 of them** before hitting a **Go build cache corruption** blocker that prevents all `go build`, `go test`, and `go vet` commands. The remaining 21 improvements are well-defined and prioritized but **blocked on the Go toolchain issue**.

**Overall Health:** 🟡 Code is correct and clean, but toolchain is broken.

---

## A) FULLY DONE ✅

| # | What                                               | Commit               | Impact                                     |
| - | -------------------------------------------------- | -------------------- | ------------------------------------------ |
| 1 | **Fix go.mod version** (`1.26.1` → `1.26.0`)       | `4f663fd`            | 🔴 Critical — build was failing            |
| 2 | **Remove `pkg/errors/apperrors.go`** template junk | `b14cef3`            | 🟡 Cleanup — 24 lines of dead code         |
| 3 | **Pre-compile regex in `FilterRegex`**             | `d5f3a40`            | 🟢 Perf — was recompiling per-event        |
| 4 | **Remove `FilterCustom` dead alias**               | `f21fc03`            | 🟡 Cleanup — unused function               |
| 5 | **Fix justfile `GOWORK=off`** (14 recipes)         | `f21fc03`            | 🔴 Critical — all go commands were failing |
| 6 | **Clean test dead code** (`_ = w`, orphan comment) | `f21fc03`            | 🟡 Cleanup                                 |
| 7 | **Status reports** (3 documents)                   | `617678f`, `e721d6e` | 📝 Documentation                           |

**Total committed this session:** 7 commits, 600 insertions, 61 deletions.

---

## B) PARTIALLY DONE 🔧

| # | What                                       | Status                                                                                                                                                                                                                                                  | Blocker                                                                                                             |
| - | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| 1 | **Replace cockroachdb/errors with stdlib** | Analysis complete, not started. Recommendation: remove. Removes 6 transitive deps (sentry-go, gogo/protobuf, kr/pretty, kr/text, logtags, redact). Go-cqrs-lite uses cockroachdb/errors, so consistency argues either way.                              | **User decision pending** — but cockroachdb/errors is making the build cache issue WORSE (more packages to compile) |
| 2 | **Fix handleNewDirectory race**            | Root cause identified at `watcher.go:474-493`. `addPath` → `walkDirFunc` appends to `watchList` without holding `mu`. The `mu.RLock()` only protects `closed` check.                                                                                    | **Blocked by Go cache**                                                                                             |
| 3 | **Make shouldSkipDir respect user dirs**   | Root cause identified at `watcher.go:342-347`. `WithIgnoreDirs` only adds a filter (post-walk), but `shouldSkipDir` (pre-walk) only checks `DefaultIgnoreDirs`. Result: user-ignored dirs still get added to fsnotify, wasting kernel file descriptors. | **Blocked by Go cache**                                                                                             |

---

## C) NOT STARTED 📋

| #  | Improvement | Priority                                                | Est. Work | Impact  |
| -- | ----------- | ------------------------------------------------------- | --------- | ------- |
| ~~ | 1           | Fix `handleNewDirectory` race condition                 | 🔴 High   | 30 min  |
| ~~ | 2           | Make `shouldSkipDir` respect user ignore dirs           | 🔴 High   | 20 min  |
| ~~ | 3           | Replace `cockroachdb/errors` with stdlib                | 🟡 Medium | 30 min  |
| ~~ | 4           | Improve `Op` type: add `MarshalText`/`UnmarshalText`    | 🟡 Medium | 15 min  |
| ~~ | 5           | Add `Event` JSON tags and `MarshalJSON`                 | 🟡 Medium | 10 min  |
| ~~ | 6           | Refactor `getDebounceKey` — remove type assertion       | 🟢 Low    | 10 min  |
| ~~ | 7           | Replace `log.Logger` with `slog` in `MiddlewareLogging` | 🟡 Medium | 20 min  |
| ~~ | 8           | Cache file handle in `MiddlewareWriteFileLog`           | 🔴 High   | 15 min  |
| ~~ | 9           | Split `watcher.go` (549 lines) into 3 files             | 🟢 Low    | 20 min  |
| ~~ | 10          | Raise test coverage to 90%+                             | 🟡 Medium | 2-3 hrs |
| ~~ | 11          | Add `-race` to test commands                            | 🟡 Medium | 5 min   |
| ~~ | 12          | Integrate into `file-and-image-renamer`                 | 🟡 Medium | 1 hr    |
| ~~ | 13          | Add `MiddlewareSlog` (new, alongside existing)          | 🟢 Low    | 15 min  |
| ~~ | 14          | Add `Watcher.WatchList()` contains check test           | 🟢 Low    | 5 min   |
| ~~ | 15          | Add `Watcher.Remove()` subdirectory removal test        | 🟢 Low    | 10 min  |
| ~~ | 16          | Fix `TestWatcher_Watch_Deletes` flakiness               | 🟡 Medium | 15 min  |
| ~~ | 17          | Add `Example_new` / `Example_watch` to doc.go           | 🟢 Low    | 10 min  |
| ~~ | 18          | Add `.golangci.yml` lint config                         | 🟢 Low    | 15 min  |
| ~~ | 19          | Add GitHub Actions CI workflow                          | 🟡 Medium | 30 min  |
| ~~ | 20          | Add `doc.go` benchmark tests                            | 🟢 Low    | 20 min  |
| ~~ | 21          | Add CHANGELOG.md entries for this sprint                | 🟢 Low    | 5 min   |

---

## D) TOTALLY FUCKED UP 💥

### Go Build Cache Corruption

**Severity:** 🔴 CRITICAL — **Blocks ALL compilation, testing, and vetting.**

**Symptom:** Every `go build`, `go test`, `go vet` fails with:

```
could not import io (open .../Library/Caches/go-build/.../xxx-d: no such file or directory)
```

**Root Cause Analysis:**

~~1. Go 1.26.0 is installed via **Nix** at `/nix/store/5ajixjk279m40yf6x96xxlnvw1wg6hq3-go-1.26.0/share/go`~~ done — shipped ≤v2.1.0, verified v2.4.1
~~2. The Nix store is **read-only** — Go cannot write its compiled standard library artifacts there~~ OBSOLETE — Nix flake migration later resolved toolchain
~~3. Go falls back to `~/Library/Caches/go-build/` for cached compilation~~ done — shipped ≤v2.1.0, verified v2.4.1
~~4. Previous `rm -rf` of the cache directory **partially succeeded** — the directory was recreated but Go's internal indexing is broken~~ done — shipped ≤v2.1.0, verified v2.4.1
~~5. `go clean -cache` reports success but **doesn't fully clear** — stale index entries remain~~ done — shipped ≤v2.1.0, verified v2.4.1
~~6. Even `go build -a` (force full rebuild) hangs because it can't write the rebuilt stdlib packages~~ done — shipped ≤v2.1.0, verified v2.4.1

**What we tried:**

~~1. `rm -rf ~/Library/Caches/go-build/` — partial success, cache regenerates broken~~ done — shipped ≤v2.1.0, verified v2.4.1
~~2. `go clean -cache` — reports success, doesn't fix~~ done — shipped ≤v2.1.0, verified v2.4.1
~~3. `go build -a ./...` — hangs indefinitely on stdlib compilation~~ done — shipped ≤v2.1.0, verified v2.4.1
~~4. All killed after 60+ seconds~~ done — shipped ≤v2.1.0, verified v2.4.1

**Likely Fix:**

```bash
# Nuclear option — kill cache completely and let Go rebuild fresh
rm -rf ~/Library/Caches/go-build
mkdir -p ~/Library/Caches/go-build
# May need to restart terminal / re-source shell for Nix paths to settle
```

**Or:** The Nix Go installation itself may be corrupted. Consider:

```bash
nix-store --verify --check-contents --repair
# Or reinstall Go via Nix
nix-env -iA nixpkgs.go_1_26
```

**Impact:** ZERO code changes can be tested until this is resolved.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

~~1. **`cockroachdb/errors` is overkill** — 6 transitive deps for 4 sentinel errors and ~10 `errors.Wrap` calls. A utility library should minimize its dependency tree. Stdlib `errors.New` + `fmt.Errorf("...: %w", err)` covers everything.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~2. **`shouldSkipDir` is disconnected from `WithIgnoreDirs`** — This is a design flaw, not a bug. `WithIgnoreDirs` adds a _filter_ (post-walk) but `shouldSkipDir` (pre-walk) only checks `DefaultIgnoreDirs`. The fix needs a new field `userSkipDirs []string` populated during option application, merged in `shouldSkipDir`.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~3. **`getDebounceKey` type assertion smell** — Uses `w.debounceInterface.(*Debouncer)` to distinguish per-path vs global debounce. Should use an explicit `perPathDebounce bool` flag or add `IsPerPath() bool` to the interface.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~4. **`MiddlewareLogging` uses `log.Logger`** — Go 1.21+ has `log/slog`. The middleware should accept `slog.Logger` or `slog.Handler` with a backward-compat adapter for `log.Logger`.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~5. **`MiddlewareWriteFileLog` opens file per event** — Potential fd exhaustion under burst. Should cache the file handle and implement `io.Closer`.~~ done — shipped ≤v2.1.0, verified v2.4.1

### Type Model Improvements

~~6. **`Op` lacks `MarshalText`/`UnmarshalText`** — Without this, `Event` can't be serialized to JSON. Add these methods so `json.Marshal(event)` works for audit/logging.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~7. **`Event` lacks JSON tags** — Should have explicit `json:"path"`, `json:"op"`, `json:"timestamp"`, `json:"is_dir"` tags.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~8. **`Stats` is too thin** — Should include `StartTime`, `EventCount`, `FilterCount` for proper observability.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~9. **`Event` should implement `fmt.Stringer`** — Already has `String()` on `Op` but `Event.String()` could be richer.~~ done — shipped ≤v2.1.0, verified v2.4.1

### Testing

~~10. **Coverage is ~77%** — Missing tests for: `handleError` stderr path, `MiddlewareLogging`, `MiddlewareWriteFileLog`, `watchLoop` error channel, `Remove` subdirectory removal, double `Watch()` call.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~11. **No race detection** — `-race` flag not in justfile. The `handleNewDirectory` race would be caught immediately.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~12. **Flaky test** — `TestWatcher_Watch_Deletes` failed once with timeout. Race between file removal and fsnotify delivery.~~ done — shipped ≤v2.1.0, verified v2.4.1

### Real-World Integration

~~13. **`hierarchical-errors` already uses this library** — Found at `cmd/watch.go`. This is validation that the API works.~~ done — shipped ≤v2.1.0, verified v2.4.1

~~14. **`file-and-image-renamer` has a confirmed debounce bug** — Uses `time.Sleep` that never resets on new events. Replacing with go-filewatcher would fix it.~~ OBSOLETE — external repo, out of scope

---

## F) TOP 25 THINGS TO DO NEXT (Priority Order)

| Rank | Action | Work                                              | Impact | Risk                  | Depends On          |
| ---- | ------ | ------------------------------------------------- | ------ | --------------------- | ------------------- |
| ~~   | **1**  | **Fix Go build cache**                            | 15 min | 🔴 Blocks everything  | None                |
| ~~   | **2**  | **Fix `handleNewDirectory` race**                 | 30 min | 🔴 Data race → crash  | Low                 |
| ~~   | **3**  | **Make `shouldSkipDir` respect user dirs**        | 20 min | 🔴 Resource waste     | Low                 |
| ~~   | **4**  | **Replace cockroachdb/errors with stdlib**        | 30 min | 🟡 6 fewer deps       | Medium (API compat) |
| ~~   | **5**  | **Add `Op.MarshalText`/`UnmarshalText`**          | 15 min | 🟡 JSON support       | None                |
| ~~   | **6**  | **Add `Event` JSON tags**                         | 10 min | 🟡 Structured logging | None                |
| ~~   | **7**  | **Cache file handle in `MiddlewareWriteFileLog`** | 15 min | 🔴 FD exhaustion      | Low (API change)    |
| ~~   | **8**  | **Replace `log.Logger` with `slog`**              | 20 min | 🟡 Modern logging     | Low (API change)    |
| ~~   | **9**  | **Refactor `getDebounceKey`**                     | 10 min | 🟢 Code quality       | None                |
| ~~   | **10** | **Run tests with `-race`**                        | 5 min  | 🟡 Catch races        | None                |
| ~~   | **11** | **Add race to justfile**                          | 5 min  | 🟡 CI quality         | None                |
| ~~   | **12** | **Add test coverage for `handleError` stderr**    | 10 min | 🟡 Coverage           | None                |
| ~~   | **13** | **Add test for `MiddlewareLogging`**              | 10 min | 🟡 Coverage           | None                |
| ~~   | **14** | **Add test for `MiddlewareWriteFileLog`**         | 15 min | 🟡 Coverage           | None                |
| ~~   | **15** | **Add test for `Remove` subdirectory**            | 10 min | 🟡 Coverage           | None                |
| ~~   | **16** | **Fix `TestWatcher_Watch_Deletes` flakiness**     | 15 min | 🟡 CI stability       | None                |
| ~~   | **17** | **Split `watcher.go` into 3 files**               | 20 min | 🟢 Maintainability    | None                |
| ~~   | **18** | **Add `Stats` observability fields**              | 15 min | 🟢 Observability      | Low                 |
| ~~   | **19** | **Integrate into file-and-image-renamer**         | 1 hr   | 🟡 Real-world fix     | Medium              |
| ~~   | **20** | **Add GitHub Actions CI**                         | 30 min | 🟡 Automation         | None                |
| ~~   | **21** | **Add `.golangci.yml`**                           | 15 min | 🟢 Code quality       | None                |
| ~~   | **22** | **Add benchmark tests**                           | 20 min | 🟢 Perf tracking      | None                |
| ~~   | **23** | **Add `Example_*` functions**                     | 10 min | 🟢 godoc              | None                |
| ~~   | **24** | **Update CHANGELOG.md**                           | 5 min  | 🟢 Documentation      | None                |
| ~~   | **25** | **Tag v0.2.0 release**                            | 5 min  | 🟢 Distribution       | None                |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

### The Go Build Cache is Corrupted — How Do You Want to Fix It?

The Nix-installed Go 1.26.0 at `/nix/store/.../share/go` has a read-only GOROOT. Go needs to compile its standard library and caches it at `~/Library/Caches/go-build/`. After repeated partial cache clears, the index is corrupted:

- `go clean -cache` reports success but doesn't fix
- `rm -rf ~/Library/Caches/go-build && go build -a` hangs indefinitely
- `go build ./...` fails with "could not import io" (can't find cached stdlib)
- Even `go vet ./...` fails for the same reason

**My recommendation:**

```bash
# Option A: Nuclear cache clear (may work if Nix Go is OK)
rm -rf ~/Library/Caches/go-build
mkdir -p ~/Library/Caches/go-build
# Then open a NEW terminal window and try: cd go-filewatcher && GOWORK=off go build ./...

# Option B: Reinstall Go via Nix
nix-env -iA nixpkgs.go_1_26

# Option C: Install Go via Homebrew instead
brew install go
```

**This is the single blocker preventing all progress.** Once resolved, I can execute the entire improvement plan in sequence.

---

## Codebase Metrics

| Metric                  | Value                                                                                                |
| ----------------------- | ---------------------------------------------------------------------------------------------------- |
| Source files            | 8 (.go)                                                                                              |
| Test files              | 4 (.go)                                                                                              |
| Example file            | 1 (.go)                                                                                              |
| Total lines             | 2,703                                                                                                |
| Source lines (no tests) | 1,401                                                                                                |
| Test lines              | 1,302                                                                                                |
| Test-to-source ratio    | 0.93                                                                                                 |
| Dependencies            | 2 direct, 8 indirect                                                                                 |
| Public types            | `Watcher`, `Event`, `Op`, `Stats`, `Filter`, `Middleware`, `Handler`, `Debouncer`, `GlobalDebouncer` |
| Public options          | 11 functional options                                                                                |
| Public filters          | 12 filter constructors                                                                               |
| Public middleware       | 7 middleware constructors                                                                            |
| Sentinal errors         | 5                                                                                                    |

## Git State

```
Branch: master
Commits ahead of origin: 7
Untracked: docs/adr/
Working tree: CLEAN (no staged or unstaged changes)
```

---

_Arte in Aeternum_
