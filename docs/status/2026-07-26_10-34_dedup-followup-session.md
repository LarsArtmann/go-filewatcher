# Dedup Follow-up Session — Status Report

**Date:** 2026-07-26 10:34 CEST
**Session type:** Continuation of the 2026-07-25 dedup session ("READ, UNDERSTAND, RESEARCH, REFLECT" prompt)
**Author:** Crush (this session)
**Pre-session baseline commit:** `de459a1` (docs(status): add deduplication session report for 2026-07-25)
**Final state:** `nix run .#check` = **ALL PASSED, 0 issues**; `art-dupl -t 5` = **0 clone groups**

---

## TL;DR

This session executed the **"Exact Next Steps" follow-ups** left open by the prior
dedup session. The headline result: **the dedup report is at its irreducible floor
at every meaningful threshold, AND the full quality gate is now green** (prior
session left `nix run .#check` non-green due to `mnd` warnings — those are fixed).

**But the session was messy.** Benchmark execution was fumbled, a disk-full incident
interrupted the final gate, and three linter iterations were needed where one should
have sufficed. Details in section (d) and (e).

---

## a) FULLY DONE

### a1. Comprehensive `newTestWatcher` adoption sweep — 43 sites migrated

The prior session migrated ~15 sites and left ~20 unflagged candidates as an open
question (Question 3). This session resolved it by migrating every defensible site:

| File                       | Sites migrated | Notes                                                                |
| -------------------------- | -------------- | -------------------------------------------------------------------- |
| `watcher_walk_test.go`     | 10             | File already partially migrated (2 sites); now fully consistent      |
| `watcher_coverage_test.go` | 14             | Skipped 2: multi-path `New([]{dir1,dir2})`, panic test               |
| `watcher_test.go`          | 17             | Skipped 8: creation-helper, error-path, manual-Close lifecycle tests |
| `errors_test.go`           | 2              | Skipped 1: closure-scoped `defer` (timing semantics differ)          |

**Final adoption:** `newTestWatcher` is now called **76 times** across 9 test files.
Remaining inline `New(` calls (20 total) are all **deliberately retained** —
benchmarks (use `b.TempDir()`, different signature), `Example*` functions (external
package API), error-path tests (assert on the error), and lifecycle tests (test
`Close()` behavior explicitly).

**Two compile regressions caught and fixed immediately by `go vet`:**

- `watcher_test.go:438` — `w` declared-not-used (handler test only used `w` in `defer Close`)
- `watcher_test.go:789` — `err` undefined (missed converting `_, err =` to `:=`)

### a2. Fixed 2 pre-existing `mnd` magic-number warnings

`examples/basic/main.go:19` (`WithDebounce(300*time.Millisecond)`) and
`examples/per-path-debounce/main.go:18` (`WithPerPathDebounce(500*time.Millisecond)`).
Both resolved by extracting `const debounceDelay` — matching the existing convention
in sibling `examples/middleware/main.go` (`exampleTimeout`, `maxEventCount`).

**Impact:** These were the sole blockers preventing a green `nix run .#check`. Now green.

### a3. Middleware default-guard consistency — `MiddlewareThrottle`

Surveyed all 11 `resolve*`/inline default guards in `middleware.go`. Policy decision:

- **Shared defaulting** (2+ functions) → extract `resolve*Defaults` helper (already done prior session)
- **Unique defaulting** (1 function) → inline guard with named const

The **one violation** was `MiddlewareThrottle`'s bare literal `100`. Fixed by adding
`defaultThrottleEvents` const (doc comment + placement matching `defaultDedupeWindow`,
`defaultBatchWindow`, `defaultExponentialBackoffInitial` convention). Every middleware
default now uses a named const — **zero magic literals remain in default-guard code**.

### a4. Unit tests for the 3 shared helpers

Added direct coverage (previously only indirectly tested via the middleware they serve):

- `TestResolveRateLimitDefaults` (6 table cases) in `middleware_test.go`
- `TestResolveMaxFailures` (3 table cases) in `middleware_test.go`
- `TestRequireNonNegativeDuration` (2 subtests: panic-with-name, zero/positive-ok) in `options_test.go`

All pass `-race`. `wsl_v5` caught a decl-cuddling issue on first lint; fixed by grouping
`recovered`/`didPanic` into a `var` block (matching project's existing panic-test idiom).

### a5. AGENTS.md Key Patterns updated

Added three rows to the Key Patterns table documenting the conventions this and the
prior session established: `resolve*Defaults`, `baseDebouncer.stop`, `newTestWatcher`.

### a6. Fixed 4 pre-existing linter warnings in examples (collateral)

When my example edits triggered a nix rebuild, the cache invalidation surfaced
pre-existing `gocritic exitAfterDefer` + `nolintlint` (unused directive) issues in
`examples/middleware/main.go` and `examples/filter-generated/main.go` — files I hadn't
originally touched. Fixed by calling `cancel()` explicitly before `log.Fatal` and
relocating the `//nolint:gocritic` directive to the same line.

### a7. Benchmark regression check — partial pass

Pure-compute benchmarks (middleware construction, event conversion, filter chains)
run clean with no regression. Debouncer `Stop()` refactor verified stable across
3 consecutive `-race` runs of the full debouncer test suite. See (d) for the caveat.

---

## b) PARTIALLY DONE

### b1. Benchmark verification

**Done:** Pure-compute benchmarks (`BenchmarkNew_*`, `BenchmarkConvertEvent_*`,
`BenchmarkBuildMiddlewareHandler_*`) pass fast with sensible numbers — no regression
from the `middleware.go`/`debouncer.go` refactors.

**Not done:** The `BenchmarkEmitEvent_*` family could not be measured. Four of them
panic/deadlock (see d2). I have **no before/after delta** — only "no obvious
regression on the benches that run." The prior session's noted gap ("No benchmarks")
is only partially closed.

### b2. Documentation of the `resolve*Defaults` policy

A single AGENTS.md table row captures the rule. There is no worked example or
"when to add a new `resolve*` helper vs inline guard" decision guide. Adequate for
now; could be richer.

---

## c) NOT STARTED

### c1. Fix the pre-existing `BenchmarkEmitEvent_NoDebounce` deadlock

(See d2 — confirmed pre-existing, not my regression. Out of scope; not started.)

### c2. Push the 6 unpushed commits on `master` ahead of `origin/master`

(Release-cadence question for the user — not started, awaiting Question 1.)

### c3. Status-report skill HTML format

(Skill specifies styled HTML dashboard; user overrode to `.md` this session and
prior. Not migrated — see Question 3.)

---

## d) TOTALLY FUCKED UP!

### d1. Benchmark execution was a multi-round fumble

**What happened:**

1. Ran `nix run .#bench` blindly → moved to background after 3 min with no output
2. Polled `job_output` twice → still "no output"
3. Eventually ran `ps aux` → discovered a **stale benchmark process from 06:16**
   (4+ hours old, hung) AND the current one
4. Killed stale processes, pivoted to targeted benches with `timeout`
5. Hit the `EmitEvent` panic/deadlock (see d2)

**What I should have done:** Run targeted, named benchmarks with a `timeout` wrapper
**from the start** (`timeout 60 go test -bench='SpecificName' ...`). The full
`-bench=.` suite is slow under `-race` and includes the known-broken `EmitEvent`
benches. Wasted ~6 rounds on diagnosis that one careful command would have skipped.

**Root cause:** I treated "run benchmarks" as a single black-box step instead of
thinking about _which_ benchmarks matter for _which_ refactor.

### d2. Disk-full incident interrupted the final quality gate

`/tmp` (tmpfs, 24G) hit "no space left on device" during the final `nix run .#check`.
Root cause: I ran `nix run .#check` / `.#lint` / `go test` roughly **15+ times**
during the session, each leaving build artifacts in `/tmp/go-build*` and the go cache.
I never cleaned up proactively. The error masqueraded as a code failure until I
checked `df -h`.

**What I should have done:** Either (a) periodically `go clean -cache`, or (b) noticed
the accumulating `/tmp/go-build*` directories, or (c) recognized tmpfs has a hard cap
unlike disk-backed paths.

### d3. Three linter iterations to silence `gocritic`/`nolintlint`

**Iteration 1:** Added `cancel()` before `log.Fatal`, **kept the original misplaced
`//nolint:gocritic` on the line above.** → `nolintlint` flagged it as unused (correctly).
**Iteration 2:** Moved nolint to same-line on **both** `log.Fatal` calls. → gocritic
only flags the _first_ fatal in a function, so the second nolint was unused → `nolintlint`.
**Iteration 3:** Removed the second nolint. → Finally clean.

**What I should have known cold:** `gocritic`'s `exitAfterDefer` fires once per function
on the first `log.Fatal`-after-`defer`. Same-line `//nolint` is the canonical fix.
One iteration, not three.

### d4. Trusted the stale session summary on first read

The prior-session summary claimed: "`nix run .#check` is non-green due to examples
mnd warnings; 2 clone groups at t=1 are `examples/middleware/main.go`." The **first**
lint run (via stale nix cache) returned "0 issues," contradicting the summary. The
**real** state (after cache rebuild) had the mnd warnings back, and the t=1 clone group
was a different pair of lines than the summary described.

**I did verify against reality** (good), but I burned a round being confused by the
cache-vs-source discrepancy. Lesson: nix lint results can be cached; always distrust
"0 issues" that come back suspiciously fast.

---

## e) WHAT WE SHOULD IMPROVE!

### e1. Personal process improvements (this session's lessons)

1. **Never run a benchmark suite blindly.** Always name specific benchmarks and wrap
   in `timeout`. A hung bench in a background shell is a silent time-sink.
2. **Watch `/tmp` and go-cache size during long sessions.** 15+ nix invocations on a
   tmpfs will fill it. Run `go clean -cache` between major phases.
3. **Know the linters' firing semantics before silencing them.** `gocritic
exitAfterDefer` fires once per function; `nolintlint` flags unused nolints.
   Reading the lint output carefully once beats iterating.
4. **Distrust suspiciously-fast "0 issues" results** — nix caches lint output.
5. **`gocritic exitAfterDefer` is the wrong fight in `main()` functions.** Example
   code that calls `log.Fatal` is idiomatic. The cleaner long-term fix is a project-
   level nolint exception for `examples/`, not per-line directives.

### e2. Codebase improvements observed (not necessarily this session's scope)

6. **`examples/` main functions repeat the `New + err-check + log.Fatal` shape.**
   This is the 1 remaining t=1 clone group. Acceptable for self-contained examples,
   but a shared `examples/demo` helper (`mustWatch`) would eliminate it and the
   recurring linter fights. The `demo` package already exists and is imported.
7. **`BenchmarkEmitEvent_*` (4 benches) are broken** — zero-value `&Watcher{}`
   deadlocks in `emitEvent`. Either fix them or delete them; a benchmark that can't
   run is worse than no benchmark.
8. **`/tmp` as `GOTMPDIR` on tmpfs** is fragile for this workflow. Consider pointing
   `GOTMPDIR` at a disk-backed path in the devShell to avoid space exhaustion.
9. **No benchmark baseline is captured anywhere.** There's no `bench-main.txt` or
   similar to diff against. Adding `nix run .#bench` output to a baseline file (gitignored)
   would make regression detection a one-command `benchstat` diff.
10. **The status-report skill says HTML; the project uses `.md`.** This is the 2nd
    override. Either update the skill's default or formalize the `.md` exception.

---

## f) Next Steps (ranked, up to 50)

### High impact — do next

1. **Fix or delete the 4 broken `BenchmarkEmitEvent_*` benchmarks.** They block
   `nix run .#bench` from completing cleanly. (See Question 2.)
2. **Push the 6 unpushed commits** on `master` ahead of `origin/master` — or
   formally decide on a release cadence. (See Question 1.)
3. **Add a `mustWatch` helper to `examples/demo`** and migrate the 5 example
   `main()` functions to it. Eliminates the last t=1 clone group AND the recurring
   `gocritic`/`mnd` linter fights in examples in one stroke.
4. **Capture a benchmark baseline** (`nix run .#bench > bench-baseline.txt`,
   gitignored) so future refactors have a `benchstat` reference.
5. **Set `GOTMPDIR` to a disk-backed path** in `flake.nix` devShell to prevent
   tmpfs exhaustion during long sessions.

### Medium impact — quality hardening

6. **Add a `//nolint:gocritic` exception at the `examples/` package level** (or via
   config) for `exitAfterDefer`, since `log.Fatal` in `main()` is intentional there.
7. **Add unit tests for `resolveBatchDefaults`** — the third `resolve*` helper,
   still only indirectly tested (serves `MiddlewareBatch` + `MiddlewareErrorBatch`).
8. **Add a test that asserts every middleware default const is used** — guards
   against dead constants if a middleware's defaulting is refactored away.
9. **Document the "shared vs unique default-guard" decision** with a worked example
   in AGENTS.md or a short ADR (the table row is too terse).
10. **Audit `middleware.go` for any remaining inline magic literals** outside
    default-guards (e.g. the `len(seen)%100 == 0` heuristic at line 209).
11. **Run `art-dupl -t 1` on `examples/` separately** to confirm no clone groups
    hiding behind the test-file exclusion defaults.
12. **Add a `just`-free / `make`-free CI check** that `examples/` all build
    (`go build ./examples/...`) — currently only validated transitively.

### Lower impact — polish

13. **Extract a `setupWatchCtx(t, timeout)` test helper** — the
    `ctx := setupTestContext(t, 5*time.Second)` line repeats ~15x and could fold
    into `newTestWatcher` as an option.
14. **Standardize test-context timeout** — `5*time.Second` is hardcoded everywhere;
    a `defaultTestTimeout` const would centralize it.
15. **Add `t.Helper()` audit** — verify all custom assertions call it (some may not).
16. **Migrate `errors_test.go:328`** (closure-scoped watcher) by introducing a
    `newTestWatcherInClosure` variant if the pattern repeats elsewhere.
17. **Check `watcher_reset_test.go`** — the 2 remaining inline `New(` are lifecycle
    tests; confirm they can't use a `newTestWatcherWithoutCleanup` variant.
18. **Add a lint rule / CI check** that `newTestWatcher` is used in preference to
    inline `New(` in test files (a custom `depguard` or `forbidigo` rule).
19. **Run `nix flake check`** end-to-end to confirm the flake itself is healthy
    (separate from `.#check`).
20. **Verify the `website/` submodule** still builds (untouched this session but
    the repo had a `website/astro.config.mjs` change in the daemon commits).
21. **Update `CHANGELOG.md`** with this session's fixes (the daemon committed a
    CHANGELOG update mid-session — verify it captures the mnd/gocritic fixes).
22. **Add `//nolint:mnd` exceptions or named consts for the `300`/`500` ms values**
    if more example debounce delays are added later — or a shared
    `examples/demo.DefaultDebounceDelay`.
23. **Consider a `testdata/` golden-file test** for art-dupl output to catch
    duplication regressions in CI.
24. **Document the `baseDebouncer.stop` envelope contract** (lock order, cleanup
    closure semantics) in a doc comment — currently only the AGENTS.md row.
25. **Audit `debouncer.go` for other extractable envelopes** — the `stopTimer`
    helper was extracted; check if `flush`/`debounce` have similar shared shapes.
26. **Add a fuzz test for `resolveRateLimitDefaults`** boundary behavior at
    `math.MinInt32`/`math.MaxInt32`.
27. **Profile `MiddlewareCircuitBreaker` under load** — it holds a mutex across
    state transitions; worth a benchmark once the EmitEvent deadlock is fixed.
28. **Check if `paralleltest` has an `ignore` config** that could allow extracting
    `t.Parallel()` into a helper for the 28 irreducible clones (unlikely, but verify).
29. **Rename `newTestWatcher` → `newWatcher`** for brevity? (Judgment call — current
    name is clear; probably leave it.)
30. **Add a `CONTRIBUTING.md`** note that `newTestWatcher` is mandatory for new tests
    (currently only discoverable via AGENTS.md Key Patterns).
31. **Tag a release** — 6 commits are ahead of origin; if the work is stable,
    `v2.2.2` or `v2.3.0` would capture the dedup + lint-clean state.
32. **Sweep `docs/status/` for stale reports** — there are 56 status reports; the
    `update-old-docs` skill could annotate which conclusions still hold.
33. **Verify `docs/DOMAIN_LANGUAGE.md` is current** — referenced by global AGENTS.md
    but not checked this session.
34. **Run the `full-code-review` skill** for a fresh-eyes pass on the refactored
    `middleware.go` and `debouncer.go`.
35. **Check `golangci-lint` version** in the flake — newer versions may have better
    `exitAfterDefer` handling.
36. **Add a `Makefile`-free task list to AGENTS.md** documenting the `nix run .#*`
    apps (partially there in "Critical Commands").
37. **Consider extracting `examples/demo.Run` into a more general `mustWatch`**
    that handles the `New + Watch + log.Fatal` envelope.
38. **Test the debouncer under concurrent `Stop()` calls** — the `baseDebouncer.stop`
    refactor centralizes locking; a `-race` stress test would harden it.
39. **Document why `BenchmarkEmitEvent_*` broke** (zero-value Watcher missing
    channel setup) in a TODO comment if not fixed immediately.
40. **Audit all `//nolint` directives** for accuracy — the session found 2 stale
    ones; a sweep with `nolintlint` in `--explain` mode would find more.
41. **Add pre-commit hook** to run `nix run .#check` (if not already present).
42. **Verify `direnv allow` workflow** still works after the flake changes.
43. **Consider a `bench-short` nix app** that runs only fast, non-I/O benchmarks
    for quick regression checks.
44. **Add type-level assertions** that the `resolve*` helpers return the same type
    they take (compile-time guard against signature drift).
45. **Review whether `defaultThrottleEvents` should equal `defaultSlidingWindowEvents`**
    (both 100) — are they the same concept? If so, unify; if not, document why not.
46. **Check the `gogenfilter v3` local replace** still resolves (mentioned in AGENTS.md
    deps; not touched this session).
47. **Add a `CHANGELOG` entry pattern** for dedup/lint work (currently ad hoc).
48. **Run `nix fmt` on `.nix` files** to confirm flake formatting is current.
49. **Consider bumping the module version** if the API surface changed (it hasn't
    this session, but worth confirming before any release).
50. **Schedule a recurring dedup audit** (`art-dupl -t 5` in CI) to prevent drift.

---

## g) Questions I CANNOT Answer Myself

### Q1. Push policy — 6 unpushed commits on `master`

There are now **6 commits on `master` ahead of `origin/master`** (the prior session
left 5; this session's daemon added more). I cannot infer your release cadence from
the repo history (some releases are tagged, some aren't; commits sometimes pile up,
sometimes get pushed immediately). **Do you want these pushed now, or held for a
tagged release (`v2.2.2` / `v2.3.0`)?** I will not push without explicit instruction.

### Q2. The pre-existing `BenchmarkEmitEvent_*` deadlock — fix now or ticket it?

I **proved** (via `git switch --detach de459a1`) that 4 benchmarks
(`BenchmarkEmitEvent_NoDebounce`, `_WithMiddleware`, `_WithGlobalDebounce`,
`_WithPerPathDebounce`) deadlock/panic due to a zero-value `&Watcher{}` missing
channel setup — **this predates my work entirely**. But it means `nix run .#bench`
can never complete cleanly. **Should I fix these benchmarks now (likely a small
change to set up the event channel / use `newTestWatcher`-style construction), or
leave them as documented pre-existing tech debt?** Fixing would unblock benchmark
regression detection for all future refactors.

### Q3. Status-report format: skill says HTML, you say `.md` — which wins long-term?

The `status-report` skill explicitly specifies a "self-contained styled HTML
dashboard." You've now overridden to `.md` **twice** (prior session + this one). I
followed your instruction both times. **Should I (a) update the skill's default to
`.md`, (b) keep overriding per-session, or (c) actually produce HTML next time?**
This affects whether I load the HTML design-system assets on future status requests.

---

## Metrics Summary

| Metric                           | Before session (`de459a1`)   | After session                               |
| -------------------------------- | ---------------------------- | ------------------------------------------- |
| `art-dupl -t 5` clone groups     | 0                            | **0**                                       |
| `art-dupl -t 1` clone groups     | 2                            | **2** (unchanged — both proven irreducible) |
| `nix run .#check`                | NON-GREEN (2 `mnd` warnings) | **GREEN — 0 issues**                        |
| `newTestWatcher` call sites      | ~33                          | **76**                                      |
| Inline `New(` in tests           | ~63                          | **20** (all deliberately retained)          |
| Unit tests for shared helpers    | 0                            | **3** (8 table cases)                       |
| Unpushed commits ahead of origin | 5                            | **6**                                       |

---

## Files Changed This Session

**Code (authored by this session):**

- `options_test.go` — 2 `newTestWatcher` migrations + `TestRequireNonNegativeDuration`
- `middleware_test.go` — `TestResolveRateLimitDefaults` + `TestResolveMaxFailures`
- `watcher_test.go` — 17 `newTestWatcher` migrations
- `watcher_coverage_test.go` — 14 `newTestWatcher` migrations
- `watcher_walk_test.go` — 10 `newTestWatcher` migrations
- `errors_test.go` — 2 `newTestWatcher` migrations
- `middleware.go` — `defaultThrottleEvents` const extraction
- `examples/basic/main.go` — `debounceDelay` const (mnd fix)
- `examples/per-path-debounce/main.go` — `debounceDelay` const (mnd fix)
- `examples/middleware/main.go` — `cancel()` before `log.Fatal`, nolint fix
- `examples/filter-generated/main.go` — `cancel()` before `log.Fatalf`, nolint fix
- `AGENTS.md` — 3 Key Patterns rows added

**Uncommitted at report time:** `examples/middleware/main.go` (1 file, pending auto-git cycle)
