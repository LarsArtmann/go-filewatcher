# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## 🟡 MEDIUM Priority

### Testing

- [x] **Add a `mustWatch` helper to `examples/demo`** and migrate the example
      `main()` functions to it. Eliminates the last `art-dupl -t 1` clone group
      (the `New + err-check + log.Fatal` envelope) AND the recurring
      `gocritic`/`mnd` linter fights in `examples/` in one stroke.
- [x] **Unit test for `resolveBatchDefaults`** — the third shared `resolve*`
      helper in `middleware.go` is still only indirectly tested (it serves
      `MiddlewareBatch` + `MiddlewareErrorBatch`). `resolveRateLimitDefaults` and
      `resolveMaxFailures` both have direct table-driven coverage now; close the
      gap.
- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Windows has different event semantics (no inotify);
      document any platform-specific skips.
- [x] **Error simulation testing** — build a fake `fsnotify.Watcher` that can
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

### Ecosystem Integration (dogfooding) — ✅ All Done

- [x] Integrate into `dynamic-markdown-site`.
- [x] Integrate into `auto-deduplicate`.
- [x] Integrate into Cyberdom.

---

## 🟢 LOW Priority

- [x] **Capture a benchmark baseline** — `bench-baseline.txt` (gitignored) is
      now captured via `nix run .#bench-baseline`; `nix run .#bench-diff` runs
      fresh benches through `benchstat` against it. Pairs with the ROADMAP
      "benchmark freshness CI" idea.
- [x] **Set `GOTMPDIR` to a disk-backed path** in the `flake.nix` devShell
      (`shellHook` exports `${XDG_CACHE_HOME:-$HOME/.cache}/go-filewatcher/gotmp`)
      to prevent tmpfs exhaustion during long sessions.
- [x] **gocritic `exitAfterDefer` exception for `examples/`** — added a
      path+text exclusion in `.golangci.yml` (`log.Fatal` in example `main()` is
      intentional); per-line directives are the wrong fight.
- [x] **Test asserting every middleware default const is used** —
      `TestMiddlewareDefaultConsts_AllUsed` (AST-based) guards against dead
      constants if a middleware's defaulting is refactored away.
- [x] **Document the shared-vs-unique default-guard decision** — added a
      worked-example "Default-guard convention" subsection in `AGENTS.md`.
- [x] **Lazy `FilterAnd` short-circuit** — verified `FilterAnd` already
      short-circuits; added `TestFilterAndShortCircuitsOnFirstFalse` as a
      regression guard.
- [x] **`WatchChanges(ctx, targetState)` contract** — sketched in
      `docs/research/watchchanges-contract.md` (idempotent sync/reconcile API for
      sync/backup workflows). Implementation deferred to a scoped task.
- [x] **Semantic-release / conventional commits** — evaluated in
      `docs/research/semantic-release-evaluation.md`; recommends `release-please` + a commit-subject lint gate over `semantic-release` for a Go module.

---

## Status Snapshot

| Metric          | Value | Status |
| --------------- | ----- | ------ |
| Linter issues   | 0     | ✅     |
| Build           | Clean | ✅     |
| Tests           | 100%  | ✅     |
| Flaky tests     | 0     | ✅     |
| Broken benches  | 0     | ✅     |
| MEDIUM priority | 6     | 🟡     |
| LOW priority    | 0     | ✅     |
