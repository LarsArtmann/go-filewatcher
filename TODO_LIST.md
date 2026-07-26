# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

Items cite evidence as `(code: path` / `src: docs/status/<report>.md)` so each is
auditable. Harvested from recent status reports on 2026-07-26.

---

## 🟠 HIGH Priority — v2.3.0 release prep

The next release publishes the `WithOnError` / `MiddlewareRateLimit` deprecations
to consumers. Those deprecations are **invisible in user-facing docs today** —
ship these before tagging.

- [ ] **Mark `WithOnError` + `MiddlewareRateLimit` deprecated in `README.md`** —
      add "(deprecated, use `WithErrorHandler` / `MiddlewareThrottle`)" callouts
      at the Options and Middleware sections.
      (`README.md` currently has zero `deprecated` matches; `src: 2026-07-26_18-53`, `2026-07-26_20-00`)
- [ ] **Add v2.3→v3 deprecation section to `MIGRATION.md`** with before/after
      snippets for both newly-deprecated symbols.
      (`MIGRATION.md` has no v2.3/deprecation content; `src: 2026-07-26_18-53`)
- [ ] **Mark deprecated symbols in `website/src/content/docs/api-reference.mdx`**
      with visual deprecation badges.
      (`src: 2026-07-26_18-53`, `2026-07-26_20-00`)
- [ ] **Fix `nix run .#ci` tidy permission failure** — the `ci` app does
      `cd "${self}"` into the read-only nix store, then `go mod tidy` fails:
      `open /nix/store/.../go.mod: permission denied`. Either run tidy against a
      writable copy or drop tidy from the hermetic `ci` app.
      (`src: 2026-07-26_18-53`, `2026-07-26_20-00`; reproduced 2026-07-26)
- [ ] **Update `FEATURES.md` "Error simulation testing" 🔵 PLANNED → 🟢 DONE** —
      the framework shipped (`error_simulation_test.go`, `fake_backend_test.go`).
      (`FEATURES.md:153` is stale; `src: 2026-07-26_20-00`)

---

## 🟡 MEDIUM Priority

### Correctness & bugs

- [ ] **`MiddlewareWriteFileLog` fd leak** — opens the log file lazily but never
      closes it (`middleware.go:345` has no `Close`/`defer`). Long-lived watchers
      with `Reset()` cycles may leak descriptors. Add a `Close()` hook or register
      cleanup with the watcher lifecycle.
      (`middleware.go:342-391` no `Close`; `src: 2026-07-26_21-00`)
- [ ] **Delete or wire up dead `addAttemptCount`** — `fakeBackend.addAttemptCount`
      is unused (`gopls unusedfunc`). Either remove it or add a test that uses it.
      (`fake_backend_test.go:85`; `src: 2026-07-26_21-00`)
- [ ] **`MiddlewareDeduplicate` cleanup-on-%100 quirk** — cleanup fires when
      `len(seen)%100 == 0`; a workload hovering at a multiple of 100 could clean
      every event. Consider a time-based or counter-rollover trigger.
      (`middleware.go:213`; `src: 2026-07-26_21-00`)

### Tooling & CI

- [ ] **Add `--tests` flag to `nix run .#lint`** — currently lint omits test
      files (golangci `tests:` is on, but the app gives no `--tests` opt-in for
      the strict path). Or add a separate `.#lint-test` app.
      (`flake.nix` lint app; `src: 2026-07-26_20-00`)
- [ ] **Add `examples/` to a nix build check** — `examples/` (incl.
      `filter-generated`) is absent from the `flake.nix` `fileset.unions`, so
      `nix build` never compiles the examples. Add a check that at least builds them.
      (`flake.nix:31-75` no examples; `src: 2026-07-26_21-00`)
- [ ] **Vendor `benchstat` hermetically** + fix bench app CWD convention —
      `nix run .#bench-diff` uses `go run golang.org/x/perf/cmd/benchstat@latest`
      (network, non-reproducible); `bench-baseline`/`bench-diff` break the
      `cd "${self}"` convention by running in caller CWD. Resolve via flake input
      or `buildGoModule`, and document the run-from-repo-root requirement.
      (`flake.nix` bench apps; `src: 2026-07-26_21-00`)
- [ ] **Re-capture clean `bench-baseline.txt`** — current baseline is polluted
      with `slog` stderr lines (captured via `2>&1 | tee`). Re-run with stderr
      discarded so the reference is clean for `benchstat`.
      (`bench-baseline.txt`; `src: 2026-07-26_21-00`)
- [ ] **Add commitlint / conventional-commit CI gate** — enforces the
      commit-message convention that the release-please recommendation depends on
      (15/200 recent subjects are non-conventional).
      (`src: 2026-07-26_21-00` semantic-release-evaluation.md)
- [ ] **Wire `release-please`** per `docs/research/semantic-release-evaluation.md`
      — release-PR workflow deriving version + CHANGELOG from conventional commits.
      Moves here on approval; pairs with the commitlint gate above.
      (`src: 2026-07-26_21-00`)
- [ ] **CI gate: README vs API_STABILITY contradiction** — add a check that
      deprecation/stability claims don't drift between the two files.
      (`src: 2026-07-26_20-00`)

### Testing

- [ ] **Windows CI matrix** — add a `windows-latest` job to `ci.yml` that runs
      `go test ./...`. Windows has different event semantics (no inotify);
      document any platform-specific skips.
- [ ] **Expand fuzz tests** — currently covers `ParseFamily`, `Classify`, error
      formatting. Add fuzzers for `FilterAnd/Or/Not` composition, `Event` JSON
      round-trip, and the gitignore matcher.
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under load.
- [ ] **Compile-time interface checks for the backend seam** — add
      `var _ watchBackend = (*fsnotifyBackend)(nil)` and
      `var _ watchBackend = (*fakeBackend)(nil)` so interface drift fails at build.
      (`backend.go`, `fake_backend_test.go`; `src: 2026-07-26_20-00`)
- [ ] **Fake-backend coverage gaps** — test `Reset()`, `Add()`, `Remove()`, and
      pipeline-level circuit-breaker behavior through `fakeBackend`; add a
      concurrent-event-burst test for goroutine leaks under load.
      (`src: 2026-07-26_20-00`)
- [ ] **Benchmark `FilterAnd` with many sub-filters** — current benchmark uses 2;
      prove the short-circuit payoff at scale (e.g. 10 filters, first rejects).
      (`benchmark_test.go:644`; `src: 2026-07-26_21-00`)

### Documentation

- [ ] **OpenTelemetry end-to-end example** — `OTelMiddleware` exists but the
      README has no tracing setup walkthrough. Add a runnable example showing
      spans propagating to a real exporter.
- [ ] **Prometheus collector quickstart** — add a `MustRegister(coll, opts...)`
      helper or a documented snippet showing standard namespace/subsystem wiring.
- [ ] **Docs freshness CI gate** — add a check that FEATURES.md/README.md
      mention every exported symbol (could be generated from `go doc -all`).
- [ ] **Link research docs from `ROADMAP.md`** — `watchchanges-contract.md` and
      `semantic-release-evaluation.md` are orphaned (referenced only from this
      file). Cross-link under "API evolution" and "Operational".
      (`src: 2026-07-26_21-00`)
- [ ] **Document `MustWatch` helper in `examples/README.md`** and add an
      `examples/` build note to AGENTS.md "File Organization".
      (`src: 2026-07-26_20-00`)
- [ ] **CONTRIBUTING.md: document the bench baseline workflow** + the
      `bench-baseline.txt` format (count=6, no `-race`) so captures stay comparable.
      (`src: 2026-07-26_21-00`)
- [ ] **Audit `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`** — likely stale post-migration;
      delete or mark historical.
      (`src: 2026-07-26_21-00`)
- [ ] **Add `docs/research/INDEX.md`** so research docs are discoverable.
      (`src: 2026-07-26_21-00`)

---

## Status Snapshot

| Metric          | Value | Status |
| --------------- | ----- | ------ |
| Linter issues   | 0     | ✅     |
| Build           | Clean | ✅     |
| Tests           | 100%  | ✅     |
| Flaky tests     | 0     | ✅     |
| Broken benches  | 0     | ✅     |
| HIGH priority   | 5     | 🟠     |
| MEDIUM priority | 22    | 🟡     |

---

## Open questions (blockers, not tasks — need user decision)

These are **not** actionable until answered. Routed here from the 2026-07-26
self-review so they are not lost; they do not belong in the checklist above.

- **Benchmark baseline: CI-enforceable (committed) or local-only (gitignored)?**
  The TODO said "gitignored"; the ROADMAP says "benchmark freshness CI." These
  contradict — a gitignored file can never drive CI.
- **Is hermetic `benchstat` a hard requirement for the bench-diff task to count
  as done, or is `@latest` acceptable for now?**
- **Keep the CWD-inconsistent bench apps, or restructure (explicit `--out` /
  writable project path)?**

A fourth open question — the `WatchChanges(ctx, targetState)` open design
questions (reporting granularity, closed-watcher semantics, depth
reconciliation) — lives in `docs/research/watchchanges-contract.md`; the
implementation becomes a TODO once those are answered.
