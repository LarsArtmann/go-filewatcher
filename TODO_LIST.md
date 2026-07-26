# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## 🟡 MEDIUM Priority

### Testing

- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Windows has different event semantics (no inotify);
      document any platform-specific skips.
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
