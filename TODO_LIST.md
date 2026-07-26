# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## Testing & Platform

- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Windows has different event semantics (no inotify);
      document any platform-specific skips.
- [ ] **Expand fuzz tests** — current corpus covers `FilterRegex`,
      `FilterExtensions`, `FilterIgnoreGlobs`, `OpUnmarshalText`, `FilterMinSize`.
      Add fuzzers for `FilterAnd`/`FilterOr`/`FilterNot` composition, `Event`
      JSON round-trip, and the gitignore matcher.
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under
      load.

## Documentation Accuracy

- [ ] **Fix README Prometheus snippet** — the polling-loop example uses
      `prometheus.ExemplarAdder` (not a real type) and would panic at runtime.
      Replace with a proper `prometheus.Collector` (Describe + Collect) wrapper,
      or reduce to API-surface pseudo-code. The library is zero-dep by design, so
      any snippet referencing `prometheus/client_golang` cannot be
      compile-verified without a test-only dependency — decide the approach.
      (`README.md:309`; `src: 2026-07-26_22-23 §d1`)
- [ ] **Fix README OTel snippet** — references `stdouttracer.New()` (correct
      package is `stdouttrace`) and `trace.Attribute` (likely renamed to
      `attribute.KeyValue` in current OTel SDK). Verify against the actual SDK
      API or mark as version-specific pseudo-code.
      (`README.md:354,380`; `src: 2026-07-26_22-23 §d2`)
- [ ] **`docs/DOMAIN_LANGUAGE.md` freshness pass** — the glossary is missing
      key terms that shipped in v2.2+: `ContentHash`, `MatchResult`,
      `FilterWithMeta`, `ErrorCategory`, `CircuitBreaker` states
      (`CircuitClosed`/`CircuitOpen`/`CircuitHalfOpen`). Verify each term
      against code and add entries.
      (`src: 2026-07-26_18-39 §c1`)
- [ ] **Shrink docs-consistency exemption list** — the
      `check-exported-symbol-docs` CI gate exempts 36 of 124 exported symbols
      (29%). Notable gaps: `BatchError`, `CircuitState`, `ErrorCategory`,
      `ErrorHandler`, `IsPermanentError`, `IsTransientError`. Document a few
      each pass until the list is empty.
      (`src: 2026-07-26_22-23 §b2`)

## Architecture

- [ ] **Document `wrapHandlerWithNilReturn` limitation** — the pipeline's error
      absorption means error-aware middleware (circuit breaker, error recovery)
      only observe failures when they are the innermost layer. This is by design
      (prevents middleware error cascades) but is an undocumented constraint.
      Write an ADR or a doc comment explaining the tradeoff.
      (`src: 2026-07-26_20-00 §B1, 2026-07-26_22-23 §e5`)

---

## Status Snapshot

| Metric          | Value | Status |
| --------------- | ----- | ------ |
| Linter issues   | 0     | ✅     |
| Build           | Clean | ✅     |
| Tests           | 100%  | ✅     |
| Flaky tests     | 0     | ✅     |
| Broken benches  | 0     | ✅     |
| Open items      | 8     | 🟡     |

---

## Open questions (blockers, not tasks — need user decision)

These are **not** actionable until answered. Routed here from the 2026-07-26
self-reviews so they are not lost; they do not belong in the checklist above.

1. **Is `.goreleaser.yml` dead config to delete, or an unfinished wiring task?**
   `release.yml` uses `softprops/action-gh-release` with auto-generated notes and
   has no goreleaser step. Now that release-please is wired in, goreleaser may be
   fully obsolete. Deleting it simplifies FEATURES (Cross-platform releases →
   not-planned); keeping it means it should be wired in eventually.
   (`src: 2026-07-26_18-39 §g Q2`)

2. **Should the benchmark baseline be committed (CI-enforceable) or stay
   gitignored (local-only)?** The baseline is currently gitignored per the
   original TODO, but the ROADMAP says "benchmark freshness CI." A gitignored
   file can never drive a CI regression gate. Machine-specific noise makes
   committed baselines imperfect, but they are the only way CI can catch
   regressions.
   (`src: 2026-07-26_21-00 §g Q1`)

3. **The `WatchChanges(ctx, targetState)` open design questions** (reporting
   granularity, closed-watcher semantics, depth reconciliation) live in
   `docs/research/watchchanges-contract.md`; the implementation becomes a TODO
   once those are answered.

4. **Should the `watchBackend` interface be exported?** Currently unexported
   (test-internal only). Consumers might want their own fake backends for
   integration testing. Tradeoff: more public API surface vs. more consumer
   value. This is a v3 decision — exporting it is a breaking commitment.
   (`src: 2026-07-26_20-00 §G Q1`)
