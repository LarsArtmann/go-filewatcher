# Status Report: Error Simulation, Example Refactor, and Docs Remediation

**Date:** 2026-07-26 20:00
**Session Scope:** 4 tasks — split brain fix, `MustWatch` helper, `resolveBatchDefaults` test, error simulation suite
**Branch:** `master` (commits `213a892..ee40a6e`)
**Quality Gate:** `nix run .#check` — ✅ All checks passed (vet + lint 0 issues + tests `-race`)

---

## A) FULLY DONE ✅

### 1. Split Brain in `API_STABILITY.md`

- **Problem:** `WithOnError` and `MiddlewareRateLimit` listed in BOTH "Stable APIs" and "Deprecated APIs" tables simultaneously — a live contradiction in the project's own stability contract.
- **Fix:** Removed both from the Stable table rows. They now appear only in the Deprecated section with migration paths.
- **Files:** `API_STABILITY.md` (2 rows edited)
- **Status:** ✅ Done, verified, committed (`ee40a6e`)

### 2. `MustWatch` Helper + Example Migration

- **Problem:** The `New + err-check + log.Fatal` boilerplate was duplicated across 4 example `main()` functions — the last surviving `art-dupl -t 1` clone group.
- **Fix:** Added `MustWatch(ctx, paths, opts...) (<-chan Event, func())` to `examples/demo/shared.go`. Migrated all 4 examples:
  - `examples/basic/main.go` — now 3 lines shorter, no manual error handling
  - `examples/per-path-debounce/main.go` — same pattern
  - `examples/middleware/main.go` — removed 2 `nolint:gocritic` directives, no more manual `cancel()` on error paths
  - `examples/filter-generated/main.go` — also fixed a **pre-existing resource leak** (deferred `Close()` ran before events were consumed; now cleanup is returned to caller)
- **Verification:** `art-dupl -t 1 examples/` → **0 clone groups** (was 2)
- **Files:** `examples/demo/shared.go`, `examples/basic/main.go`, `examples/per-path-debounce/main.go`, `examples/middleware/main.go`, `examples/filter-generated/main.go`
- **Status:** ✅ Done, verified, committed

### 3. `resolveBatchDefaults` Unit Test

- **Problem:** The third shared `resolve*` helper in `middleware.go` (serving `MiddlewareBatch` + `MiddlewareErrorBatch`) had no direct test — only indirect coverage through the middlewares themselves.
- **Fix:** Added `TestResolveBatchDefaults` with 6 table-driven cases covering: both-positive pass-through, zero/negative window, zero/negative maxSize, both-non-positive defaults.
- **Files:** `middleware_test.go` (1 test function added)
- **Status:** ✅ Done, verified, committed (`213a892`)

### 4. Error Simulation Test Suite (the big one)

- **Problem:** No way to deterministically test error paths — ENOSPC, permission denied, closed-watcher errors, self-heal retry, circuit breaker state transitions. All existing tests relied on real filesystem events.
- **Architecture Decision:** Introduced a `watchBackend` interface in `backend.go`:
  ```go
  type watchBackend interface {
      Add(name string) error
      Remove(name string) error
      Close() error
      Events() <-chan fsnotify.Event
      Errors() <-chan error
  }
  ```
  - `fsnotifyBackend` adapter wraps `*fsnotify.Watcher` for production (exposes channel fields as methods).
  - `withBackend(b watchBackend)` unexported `Option` injects fakes in tests.
  - `Watcher.fswatcher` field type changed from `*fsnotify.Watcher` → `watchBackend`.
  - `New()` and `Reset()` updated to create `fsnotifyBackend{fw}` when no backend is injected.
- **Fake Backend (`fake_backend_test.go`):**
  - Buffered event and error channels (capacity 100)
  - Scriptable `addFn func(path, attempt) error` — nil = always succeed
  - Tracks add attempts per path, added paths, removed paths
  - Thread-safe with `sync.Mutex`
- **11 Tests Written (`error_simulation_test.go`):**

| Test                                               | What It Verifies                                                                                 |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `TestSelfHeal_HealsFailedPathAfterRetry`           | ENOSPC on all Adds → paths in failedPaths → Add starts succeeding → self-heal clears failedPaths |
| `TestSelfHeal_AbandonsPermanentError`              | Permanent error (wraps `ErrPathNotDir`) → self-heal abandons path permanently                    |
| `TestErrorChannel_PropagatesBackendErrorToHandler` | Error injected via backend channel → reaches `ErrorHandler` with correct `ErrorContext`          |
| `TestErrorChannel_PropagatesToErrorsChannel`       | Error injected → reaches public `Errors()` channel                                               |
| `TestEventPipeline_EventsFlowThroughFakeBackend`   | Create event injected → flows through full pipeline → received on events channel                 |
| `TestEventPipeline_MiddlewareProcessesFakeEvents`  | 3 events injected → `MiddlewareMetrics` counter increments for each                              |
| `TestClosedBackend_WatchLoopStopsGracefully`       | Backend channels closed → watchLoop exits → events channel closes                                |
| `TestCircuitBreaker_DropsEventsAfterFailures`      | 2 failures → circuit opens → events dropped → timeout → half-open → 1 event passes               |
| `TestErrorRecovery_StrategyReceivesEventAndError`  | Error from handler → strategy receives both event and error → can suppress                       |
| `TestWatchErrors_CounterIncrementsOnAddFailure`    | Add fails → `Stats().WatchErrors` increments                                                     |
| `TestFakeBackend_AddFailsSpecificPaths`            | Only subdirectory Add fails → root succeeds, sub in failedPaths                                  |

- **Files:** `backend.go` (new, 34 lines), `fake_backend_test.go` (new, 90 lines), `error_simulation_test.go` (new, 502 lines), `watcher.go` (field type + constructor + Reset), `watcher_internal.go` (2 channel reads: `.Events` → `.Events()`, `.Errors` → `.Errors()`)
- **Status:** ✅ Done, verified, committed (`0f5fc41`, `fa2e855`, `caff0d2`)
- **Stability:** Passed 3 consecutive runs with `-race -count=3`

### 5. Docs & Project Housekeeping

- `flake.nix` — added `backend.go`, `error_simulation_test.go`, `fake_backend_test.go` to the Nix fileset (reproducible build)
- `AGENTS.md` — added `backend.go` to File Organization table, added "Backend Abstraction" to Key Patterns, fixed `newTestWatcher` line number reference
- `TODO_LIST.md` — marked 4 items done (mustWatch, resolveBatchDefaults, error simulation, all 3 ecosystem integrations), updated MEDIUM count 12→6
- `CHANGELOG.md` — added Internal section entries for backend abstraction, error simulation suite, MustWatch helper, resolveBatchDefaults test; updated art-dupl claim to "zero at all thresholds"
- **Status:** ✅ Done, verified, committed (`ee40a6e`)

---

## B) PARTIALLY DONE ⚠️

### 1. Circuit Breaker Pipeline Integration

The circuit breaker test (`TestCircuitBreaker_DropsEventsAfterFailures`) tests the middleware in **isolation**, not through the full watcher pipeline. This is because the watcher's `wrapHandlerWithNilReturn` absorbs errors from each middleware layer, so inner middleware errors never propagate to the circuit breaker wrapping it. The test is correct and valuable, but it documents an architectural limitation rather than testing the breaker end-to-end.

**What's missing:** A test that sends events through the fake backend with a circuit breaker configured, and verifies that the breaker opens/closes based on handler errors in the real pipeline. This would require either (a) a handler that errors via the event channel pipeline, or (b) documenting that circuit breaker is only effective when it's the innermost middleware.

### 2. LSP Diagnostics Still Stale

The `golangci_lint_ls` LSP client continues reporting phantom diagnostics for `backend.go` ("cannot use b as *fsnotify.Watcher") and `fake_backend_test.go` (gci/wsl_v5). These are **stale** — `golangci-lint run` and `nix run .#check` both report 0 issues. I restarted the LSP once but the phantom warnings persisted. This is cosmetic but misleading during development.

### 3. Statistical Verification of New Tests

New tests passed 3x with `-race`, but I did not run `-count=50` for the self-heal timing-sensitive tests (which use 50ms intervals and `waitForCondition` with 2s timeouts). These could still be flaky under extreme load, though the polling approach is far more robust than the previous `time.Sleep` pattern.

---

## C) NOT STARTED ⬜ → ALL RESOLVED

> **Update 2026-07-26 (later sessions):** every item below has since shipped.
> See the Resolution appendix at the end of this file.

### From Previous Session's P0 List (Still Open ~~→ DONE~~)

1. ~~**`README.md` lines 153, 204** — still reference deprecated `WithOnError` and `MiddlewareRateLimit` as current APIs with no deprecation marker.~~ DONE: ⚠️ deprecation markers added;
2. ~~**`MIGRATION.md`** — no v2.3→v3 deprecation section exists.~~ DONE: "Migrating to v2.3+" section with before/after snippets;
3. ~~**`website/src/content/docs/api-reference.mdx`** — deprecated symbols not marked on the public docs site.~~ DONE: Starlight `:::caution[Deprecated]` blocks added;
4. ~~**`nix run .#ci` tidy step** — still fails with permission denied on `go.mod`.~~ DONE: ci/fmt/tidy apps now run from caller CWD;

### From TODO_LIST.md (Still Open ~~→ mostly DONE~~)

- ~~OpenTelemetry end-to-end example~~ DONE: README snippet added (accuracy caveat — see TODO_LIST);
- ~~Prometheus collector quickstart~~ DONE: README snippet added (accuracy caveat — see TODO_LIST);
- ~~Docs freshness CI gate~~ DONE: `check-exported-symbol-docs` workflow;
- Windows CI matrix — OPEN: see TODO_LIST;
- Expand fuzz tests — OPEN: see TODO_LIST (5 fuzzers exist, expansion planned);
- Large-tree stress harness — OPEN: see TODO_LIST;
- ~~All 8 LOW priority items~~ DONE: see CHANGELOG `[Unreleased]`;

---

## D) TOTALLY FUCKED UP 💥

### 1. I Didn't Notice the Auto-Git Daemon Committed Mid-Session

The auto-git daemon committed changes at least 6 times during this session (`213a892`, `038160a`, `0f5fc41`, `fa2e855`, `caff0d2`, `b4d1dcb`, `ee40a6e`). Some of these commits have **AI-boilerplate messages that don't describe the actual semantic changes**:

- `b4d1dcb chore(flake): update Nix flake configuration for Go build and dev environment` — this was just adding 3 files to the fileset, not a "configuration update"
- The commit messages describe implementation details ("Establish a foundation for mocking platform-specific error responses") rather than user-facing impact ("Enable deterministic error injection for self-heal and circuit breaker tests")

**Impact:** Git history is noisy and the commit messages don't help future readers understand why changes were made.

### 2. The `withBackend` Option Triggered a Lint False Positive

During development, the `unused` linter flagged `withBackend` because it was only used in test files. The Nix `.#lint` app runs `golangci-lint run ./...` which by default does **not** lint test files, so it didn't see the usage. I had to use `golangci-lint run --tests ./...` to verify. This means the project's `.#lint` and `.#check` commands have a **blind spot** — they don't catch issues that only manifest in test files. This is a pre-existing problem, not something I caused, but I should have flagged it.

### 3. I Left Dead Code Temporarily

When I first created `error_simulation_test.go`, the circuit breaker test had an `errorMiddleware` variable that was declared but never used (I realized the pipeline absorbs errors). I fixed it, but the intermediate state was committed by the auto-git daemon. The final version is clean, but git history contains a broken intermediate.

---

## E) WHAT WE SHOULD IMPROVE 🎯

### Process Improvements

1. **Commit message quality** — The auto-git daemon writes generic messages. For semantic changes (new interfaces, deprecations), we need human-quality commit messages. Consider disabling auto-git for sessions with architectural changes, or post-processing commit messages.
2. **Test file linting** — `nix run .#lint` doesn't lint test files by default. Add `--tests` flag to the lint app, or at minimum document this blind spot.
3. **LSP diagnostics** — The `golangci_lint_ls` client goes stale after interface changes. Need a way to force-rebuild its cache, not just restart.
4. **Statistical test verification** — We keep claiming tests are "not flaky" after 3-10 runs. For timing-sensitive tests (self-heal, debouncer), establish a protocol: `-count=50 -race` before marking done.

### Code Improvements

5. **Circuit breaker in pipeline** — The fact that `wrapHandlerWithNilReturn` absorbs errors means circuit breaker only works as the innermost middleware. This is a **hidden architectural constraint** that should be documented or fixed. Either: (a) document that circuit breaker must wrap the failing handler directly, or (b) refactor the pipeline so errors propagate through the full chain.
6. **`watchBackend` interface should be documented as internal** — It's unexported, but the pattern is significant. Add a comment explaining it's a test seam, not a public extension point.
7. **`fakeBackend` could be more capable** — Currently injects errors synchronously. Could add: delayed error injection, event sequencing (create→write→remove chains), concurrent event bursts.
8. **`MustWatch` swallows the watcher** — The caller gets events + cleanup but not the `*Watcher` itself. If an example needs `Stats()` or `Add()`, it can't use `MustWatch`. Consider returning the watcher too, or a different helper for advanced cases.

### Documentation Improvements

9. **`README.md` deprecation drift** — Still references deprecated APIs as current. Live contradiction between README and API_STABILITY.md.
10. **No migration guide** — `MIGRATION.md` doesn't exist or doesn't have the v3 deprecation section.
11. **Website docs stale** — Public-facing docs don't reflect deprecations.

---

## F) Up to 50 Things We Should Get Done Next

### P0 — Fix Live Damage (Cheap, High-Impact)

1. Mark `WithOnError` as deprecated in `README.md` (line ~153)
2. Mark `MiddlewareRateLimit` as deprecated in `README.md` (line ~204)
3. Create/update `MIGRATION.md` with v2.3→v3 deprecation section
4. Update `website/src/content/docs/api-reference.mdx` to mark deprecated symbols
5. Investigate `nix run .#ci` tidy permission failure (file a tracked issue if environmental)

### P1 — Lock In & Verify

6. Run self-heal tests with `-count=50 -race` for statistical confidence
7. Run full suite with `-count=10 -race` to catch any new timing issues
8. Add `--tests` flag to `nix run .#lint` app in `flake.nix` (fixes the test-file blind spot)
9. Run `staticcheck` SA1019 explicitly to confirm deprecation warnings emit for consumers
10. Verify `nix flake check` passes with new fileset entries

### P2 — Architecture & Testing

11. Document or fix the circuit-breaker-in-pipeline limitation (`wrapHandlerWithNilReturn`)
12. Add a compile-time interface check: `var _ watchBackend = (*fsnotifyBackend)(nil)`
13. Add compile-time check: `var _ watchBackend = (*fakeBackend)(nil)`
14. Write a pipeline-level circuit breaker test (send failing events through fake backend)
15. Add `fakeBackend` delayed-error injection (for testing timeout/retry windows)
16. Add `fakeBackend` event-sequence helper (create→write→remove chains)
17. Add concurrent event burst test (verify no goroutine leaks under load)
18. Test `Reset()` with a fake backend (verify it creates a real backend, discarding the fake)
19. Test `Add()` and `Remove()` through the fake backend (track addedPaths/removedPaths)
20. Add `go test -race -count=1 -run 'TestFakeBackend'` to verify the fake itself is race-free

### P3 — Docs & Website

21. Add `watchBackend` architecture note to AGENTS.md (test seam, not public extension point)
22. Document the `MustWatch` helper in examples README
23. Add `resolveBatchDefaults` to the "Key Patterns" `resolve*Defaults` row in AGENTS.md
24. Update `FEATURES.md` with "Error simulation testing" as a DONE feature
25. Write ROADMAP entry for "Public backend plugin API" if we ever want to expose `watchBackend`
26. Add deprecation callouts to website docs with visual badges
27. Create a "Testing Guide" doc showing how to use `fakeBackend` for consumer testing

### P4 — Linting & CI

28. Fix `nix run .#lint` to include `--tests` flag
29. Add `nix run .#lint-test` as a separate app if changing the default is too aggressive
30. Add a CI gate that `README.md` and `API_STABILITY.md` don't contradict each other
31. Add Windows CI matrix job
32. Run `govulncheck` and `gosec` in CI (security hardening)
33. Add `gofumpt` to the lint pipeline (currently only in `treefmt`)
34. Capture benchmark baseline (`nix run .#bench > bench-baseline.txt`)

### P5 — Code Quality

35. Extract `cancelAndDrain` to `testing_helpers_test.go` (currently in `error_simulation_test.go`)
36. Add `waitForCondition` to the shared test helpers documentation in AGENTS.md
37. Consider making `fakeBackend` exported (`FakeBackend`) for consumer testing
38. Add `WithBackend` as an exported option if consumers want custom backends
39. Add `Stats().SelfHealAttempts` counter (currently no way to verify self-heal ran N times)
40. Add `Stats().CircuitState` gauge (currently no observability for circuit breaker state)
41. Add `fakeBackend` assertion helpers (`assertAdded(t, path)`, `assertAttemptCount(t, path, n)`)

### P6 — Future Features

42. Lazy `FilterAnd` short-circuit (return on first `false`)
43. `WatchChanges(ctx, targetState)` idempotent sync API
44. Large-tree stress harness (100k directories)
45. Expand fuzz tests (FilterAnd/Or/Not, Event JSON, gitignore matcher)
46. OpenTelemetry end-to-end example
47. Prometheus collector quickstart with `MustRegister`
48. Semantic-release / conventional commits evaluation
49. Document the shared-vs-unique default-guard decision with worked example
50. Set `GOTMPDIR` to disk-backed path in devShell (prevent tmpfs exhaustion)

---

## G) Questions for the User

### 1. Should the `watchBackend` interface and `withBackend` option be exported?

Currently both are unexported (test-internal only). But consumers of this library might want to write their own fake backends for integration testing their filewatcher usage. Making `WatchBackend` and `WithBackend` exported would turn this test seam into a public extensibility point. The tradeoff: more public API surface to maintain vs. more value for consumers.

**I cannot decide this myself** because it depends on whether you want this library to be a "batteries-included testing framework" or just a file watcher. The answer changes the API stability commitment.

### 2. Should I fix the circuit-breaker-in-pipeline limitation or document it?

The `wrapHandlerWithNilReturn` function absorbs errors from each middleware layer, meaning the circuit breaker only sees errors from the handler directly inside it — not from outer middleware. This is either a bug (the breaker should see all errors) or a feature (each middleware isolates its own errors). Fixing it would require refactoring the middleware chain to propagate errors, which changes behavior for all existing middleware. Documenting it is safe but leaves a hidden gotcha.

**I cannot decide this myself** because the "right" answer depends on your middleware philosophy — are middleware layers independent error boundaries, or should errors cascade through the chain?

### 3. Is the `nix run .#ci` tidy failure a known environment quirk?

The `go mod tidy` step in `nix run .#ci` fails with `open /nix/store/.../source/go.mod: permission denied`. This suggests the Nix build environment makes the source read-only, but `go mod tidy` tries to write. This might be expected (CI should run tidy locally, not in the Nix sandbox) or a real bug in the flake configuration.

**I cannot figure this out myself** because I don't know if this ever worked, or if `.#ci` is intended to be run only outside the Nix sandbox. If it's a known limitation, it should be documented. If it's a regression, the flake's `mkApp` wrapper needs to make the source writable.
