# TODO List

**Last Updated:** 2026-06-18

Short- and mid-term actionable work. Each item is scoped and owned by no one in
particular — pick one, do it, tick the box.

Long-term direction and raw ideas live in [ROADMAP.md](./ROADMAP.md).
Completed history lives in [CHANGELOG.md](./CHANGELOG.md) and
[FEATURES.md](./FEATURES.md).

---

## 🔴 HIGH Priority

- [ ] **Harden or quarantine flaky tests** — `TestWatcher_Stats_Metrics` and
      `TestWatcher_Watch_WithMiddleware` intermittently fail due to fsnotify write
      coalescing. Either make assertions event-count-agnostic or `t.Skip` with a
      tracked issue and a re-enable plan.
- [ ] **Wire goreleaser publish pipeline** — `.goreleaser.yml` exists but no
      GitHub Actions workflow runs it on tag. Add a `release.yml` job that builds
      cross-platform artifacts on `v*` tags.
- [ ] **Deprecation audit for v3** — `WithWatchedIgnoreDirs` is the only
      declared deprecation. Inventory other candidates (e.g., two-arg
      `ErrorHandler` signature) and list them in API_STABILITY.md so v3 scope is
      clear before cutting.

---

## 🟡 MEDIUM Priority

### Testing

- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Document any platform-specific skips.
- [ ] **Error simulation testing** — build a fake `fsnotify.Watcher` that can
      inject ENOSPC, permission denied, and closed-watcher errors. Use it to
      exercise `MiddlewareCircuitBreaker`, `MiddlewareErrorRecovery`, and self-heal
      deterministically.
- [ ] **Expand fuzz tests** — currently covers `ParseFamily`, `Classify`, error
      formatting. Add fuzzers for `FilterAnd/Or/Not` composition, `Event` JSON
      round-trip, gitignore matcher.
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under load.
- [ ] **Extract `drainEvents` to a testutil package** — currently inlined in
      multiple tests. Centralize for reuse.

### Documentation

- [ ] **OpenTelemetry end-to-end example** — `OTelMiddleware` exists but the
      README has no tracing setup walkthrough. Add a runnable example showing
      spans propagating to a real exporter.
- [ ] **Prometheus collector quickstart** — add `MustRegister(coll, opts...)`
      helper or a documented snippet showing standard namespace/subsystem wiring.
- [ ] **Docs freshness CI gate** — add a check that FEATURES.md/README.md
      mention every exported symbol (could be generated from `go doc -all`).

### Ecosystem Integration

- [ ] **Integrate into file-and-image-renamer** — dogfood in a real consumer.
- [ ] **Integrate into dynamic-markdown-site** — dogfood.
- [ ] **Integrate into auto-deduplicate** — dogfood.
- [ ] **Integrate into Cyberdom** — dogfood.

---

## 🟢 LOW Priority

- [ ] **Localizable error messages** — sentinel errors are English-only today.
      Evaluate whether to externalize message templates (likely v3 scope).
- [ ] **Semantic-release / conventional commits** — evaluate whether
      commit-message-driven versioning reduces manual CHANGELOG drift.
- [ ] **Explore fsnotify v2 API changes** — monitor upstream for breaking
      changes; no action until a release candidate lands.
- [ ] **`WatchChanges(ctx, targetState)`** — idempotent sync API for
      sync/backup workflows. Sketch the contract before implementing.
- [ ] **Zero-allocation event path** — `ConvertEvent/Create` is 3 allocs.
      Investigate pooling or stack-allocated `Event` for hot paths.
- [ ] **Lazy `FilterAnd` short-circuit** — currently evaluates all sub-filters.
      Return on first `false` for measurably cheaper composition.

---

## Status Snapshot

| Metric          | Value | Status |
| --------------- | ----- | ------ |
| Linter issues   | 0     | ✅     |
| Build           | Clean | ✅     |
| Tests           | 100%  | ✅     |
| Flaky tests     | 2     | 🟡     |
| HIGH priority   | 3     | 🔴     |
| MEDIUM priority | 11    | 🟡     |
| LOW priority    | 6     | 🟢     |
