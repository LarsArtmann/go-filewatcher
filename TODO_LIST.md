# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## 🔴 HIGH Priority

- [ ] **Fix or delete the 4 broken `BenchmarkEmitEvent_*` benchmarks** —
      `BenchmarkEmitEvent_NoDebounce`, `_WithMiddleware`, `_WithGlobalDebounce`,
      `_WithPerPathDebounce` (in `benchmark_test.go`) deadlock/panic because they
      construct a zero-value `&Watcher{}` with no event channel. Proven
      pre-existing (reproduced at commit `de459a1`). They block `nix run .#bench`
      from completing cleanly and prevent regression detection for every future
      refactor. Either wire up the channel / use `newTestWatcher`-style
      construction, or delete them.
- [ ] **Harden or quarantine flaky tests** — `TestWatcher_Stats_Metrics` and
      `TestWatcher_Watch_WithMiddleware` intermittently fail due to fsnotify
      write coalescing (see AGENTS.md "Known Issues"). Either make assertions
      event-count-agnostic or `t.Skip` with a tracked issue and a re-enable plan.
- [ ] **Deprecation audit for v3** — `WithWatchedIgnoreDirs` is the only declared
      deprecation. Inventory other candidates (e.g. the two-arg `ErrorHandler`
      signature) and list them in `API_STABILITY.md` so v3 scope is clear before
      cutting.

---

## 🟡 MEDIUM Priority

### Testing

- [ ] **Add a `mustWatch` helper to `examples/demo`** and migrate the 5 example
      `main()` functions to it. Eliminates the last `art-dupl -t 1` clone group
      (the `New + err-check + log.Fatal` envelope) AND the recurring
      `gocritic`/`mnd` linter fights in `examples/` in one stroke.
- [ ] **Unit test for `resolveBatchDefaults`** — the third shared `resolve*`
      helper in `middleware.go` is still only indirectly tested (it serves
      `MiddlewareBatch` + `MiddlewareErrorBatch`). `resolveRateLimitDefaults` and
      `resolveMaxFailures` both have direct table-driven coverage now; close the
      gap.
- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Windows has different event semantics (no inotify);
      document any platform-specific skips.
- [ ] **Error simulation testing** — build a fake `fsnotify.Watcher` that can
      inject ENOSPC, permission denied, and closed-watcher errors. Use it to
      exercise `MiddlewareCircuitBreaker`, `MiddlewareErrorRecovery`, and
      self-heal deterministically.
- [ ] **Expand fuzz tests** — currently covers `ParseFamily`, `Classify`, error
      formatting. Add fuzzers for `FilterAnd/Or/Not` composition, `Event` JSON
      round-trip, and the gitignore matcher.
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under
      load.

### Documentation

- [ ] **OpenTelemetry end-to-end example** — `OTelMiddleware` exists but the
      README has no tracing setup walkthrough. Add a runnable example showing
      spans propagating to a real exporter.
- [ ] **Prometheus collector quickstart** — add a `MustRegister(coll, opts...)`
      helper or a documented snippet showing standard namespace/subsystem wiring.
- [ ] **Docs freshness CI gate** — add a check that FEATURES.md/README.md
      mention every exported symbol (could be generated from `go doc -all`).

### Ecosystem Integration (dogfooding)

- [ ] Integrate into `file-and-image-renamer`.
- [ ] Integrate into `dynamic-markdown-site`.
- [ ] Integrate into `auto-deduplicate`.
- [ ] Integrate into Cyberdom.

---

## 🟢 LOW Priority

- [ ] **Capture a benchmark baseline** — `nix run .#bench > bench-baseline.txt`
      (gitignored) so future refactors have a one-command `benchstat` diff
      reference. Pairs with the ROADMAP "benchmark freshness CI" idea.
- [ ] **Set `GOTMPDIR` to a disk-backed path** in the `flake.nix` devShell to
      prevent tmpfs exhaustion during long sessions (15+ nix invocations filled a
      24G `/tmp` during the 2026-07-26 session).
- [ ] **Add a package-level `//nolint:gocritic` exception for `examples/`**
      (`exitAfterDefer`) — `log.Fatal` in example `main()` functions is
      intentional; per-line directives are the wrong fight.
- [ ] **Test asserting every middleware default const is used** — guards against
      dead constants if a middleware's defaulting is refactored away.
- [ ] **Document the shared-vs-unique default-guard decision** with a worked
      example in `AGENTS.md` (the current Key Patterns table row is too terse;
      see `resolve*Defaults` convention).
- [ ] **Lazy `FilterAnd` short-circuit** — currently evaluates all sub-filters;
      return on first `false` for measurably cheaper composition.
- [ ] **`WatchChanges(ctx, targetState)`** — idempotent sync API for sync/backup
      workflows. Sketch the contract before implementing.
- [ ] **Semantic-release / conventional commits** — evaluate whether
      commit-message-driven versioning reduces manual CHANGELOG drift.

---

## Status Snapshot

| Metric          | Value | Status |
| --------------- | ----- | ------ |
| Linter issues   | 0     | ✅     |
| Build           | Clean | ✅     |
| Tests           | 100%  | ✅     |
| Flaky tests     | 2     | 🟡     |
| Broken benches  | 4     | 🔴     |
| HIGH priority   | 3     | 🔴     |
| MEDIUM priority | 13    | 🟡     |
| LOW priority    | 8     | 🟢     |
