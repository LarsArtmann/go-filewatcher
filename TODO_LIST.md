# TODO List

**Last Updated:** 2026-07-26

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## ✅ Completed — v2.3.0 release prep (2026-07-26)

All HIGH and MEDIUM correctness/tooling items from the 2026-07-26 harvest are
done. See CHANGELOG for details.

- [x] Mark `WithOnError` + `MiddlewareRateLimit` deprecated in README.md
- [x] Add v2.3→v3 deprecation section to MIGRATION.md
- [x] Mark deprecated symbols in api-reference.mdx with deprecation badges
- [x] Fix `nix run .#ci` tidy permission failure (run from caller CWD)
- [x] Update FEATURES.md "Error simulation testing" PLANNED → DONE
- [x] Fix `MiddlewareWriteFileLog` fd leak (`NewFileLogMiddleware` + `WithCleanup`)
- [x] Wire up `addAttemptCount` in self-heal and fake-backend tests
- [x] Fix `MiddlewareDeduplicate` cleanup-on-%100 quirk (event counter)
- [x] Add `nix run .#lint-tests` app (explicit `--tests` flag)
- [x] Add `examples/` to nix `fileset` + `examples-build` check
- [x] Vendor `benchstat` hermetically via `buildGoModule`
- [x] Re-capture clean `bench-baseline.txt` (`-run=^$`)
- [x] Add commitlint CI gate (`.github/workflows/commitlint.yml`)
- [x] Wire `release-please` (`.github/workflows/release-please.yml`)
- [x] Add README vs API_STABILITY docs-consistency CI gate

---

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
| HIGH priority   | 0     | ✅     |
| MEDIUM priority | 7     | 🟡     |

---

## Open questions (blockers, not tasks — need user decision)

These are **not** actionable until answered. Routed here from the 2026-07-26
self-review so they are not lost; they do not belong in the checklist above.

The `WatchChanges(ctx, targetState)` open design questions (reporting
granularity, closed-watcher semantics, depth reconciliation) live in
`docs/research/watchchanges-contract.md`; the implementation becomes a TODO
once those are answered.
