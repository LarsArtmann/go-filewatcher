# Roadmap

**Last Updated:** 2026-07-26

Long-term direction and raw ideas for go-filewatcher. Items here are
**exploratory and not yet committed** — they graduate to
[TODO_LIST.md](./TODO_LIST.md) when scoped into bounded, estimable work. For the
current feature snapshot, see [FEATURES.md](./FEATURES.md).

---

## Themes

1. **Hardening** — broader platform coverage, error-path testing, performance
   regressions.
2. **Ecosystem integration** — first-class support for observability stacks and
   dogfooding in real consumers.
3. **API evolution** — toward v3 with feedback-driven breaking changes.
4. **Operational excellence** — automated releases, dependency hygiene, docs
   freshness, reproducible benchmarks.

Concrete tasks under these themes live in
[TODO_LIST.md](./TODO_LIST.md). The ideas below are the longer-term, less
certain directions worth exploring.

---

## Ideas Worth Exploring

### Platform Coverage

- **macOS FSEvents edge cases** — rename semantics differ from Linux; document
  or test the divergences (current CI is Linux-only).
- **BSD/kqueue** — fsnotify supports it; verify our assumptions (budget
  detection, batched registration) hold.

### API Evolution

- **v3 planning** — accumulated deprecations (`WithWatchedIgnoreDirs`) and
  awkward signatures (the two-arg `ErrorHandler`) suggest a v3 cleanup pass.
  The concrete deprecation inventory is a TODO_LIST task; this is the strategic
  framing: gather breaking changes over the next 6–12 months before cutting.
  See [docs/research/watchchanges-contract.md](./docs/research/watchchanges-contract.md)
  for the event-contract analysis that informs v3 API decisions.
- **Streaming filter protocol** — current `Filter` is a sync bool. Consider
  returning `(keep bool, err error)` or a channel-based variant for filters
  that need async I/O (e.g. remote manifest lookup).

### Observability

- **pprof endpoints for watcher introspection** — expose watch-list size,
  debouncer queue depth, filter rejection counts via `net/http/pprof`.

### Performance

- **Zero-allocation event path** — current `ConvertEvent`/`Create` is 3 allocs;
  investigate pooling or stack-allocated `Event` for hot paths.
- **Benchmark freshness CI** — local baseline capture (`nix run .#bench-baseline`)
  and diff (`nix run .#bench-diff`) tooling exist, but the baseline is gitignored.
  The open question is whether to commit a sanitized baseline and add a CI
  regression gate with tolerance (see TODO_LIST open questions).

### Reliability

- **Race-detector-on-CI flake quarantine** — two formerly-flaky tests
  (`TestWatcher_Stats_Metrics`, `TestWatcher_Watch_WithMiddleware`) are now
  hardened via the `waitForCondition` polling helper. The open question is
  whether to also add a `-count=50 -race` statistical gate in CI for
  timing-sensitive tests (self-heal, debouncer intervals).

### Operational

- **Cross-platform release artifacts via goreleaser** — `.goreleaser.yml` is
  configured but **not invoked** by `release.yml` (which uses
  `softprops/action-gh-release` with auto-generated notes). Now that
  release-please handles versioning + CHANGELOG, goreleaser may be dead config
  to delete (see TODO_LIST open questions). Wiring it in would ship compiled
  cross-platform binaries on tag. See FEATURES.md "Cross-platform releases"
  (currently PARTIALLY DONE).
- **Dependency freshness SLO** — current policy is "update within 24h"; codify
  with Dependabot status checks.
- **Docs freshness gate** — the `check-exported-symbol-docs` CI workflow exists
  and ratchets against a 36-symbol exemption list (29% of the API). The
  exploratory angle is auto-generating doc tables from `go doc -all` source
  parsing to eliminate manual sync entirely.
  See [docs/research/INDEX.md](./docs/research/INDEX.md) for related research.
- **Automated release tooling** — release-please is wired in
  (`.github/workflows/release-please.yml`), auto-generating release PRs from
  conventional commits. The commitlint gate
  (`.github/workflows/commitlint.yml`) enforces the commit format that feeds it.
  See [docs/research/semantic-release-evaluation.md](./docs/research/semantic-release-evaluation.md)
  for the tradeoff analysis that informed this choice.

---

## Non-Goals

These are explicitly **out of scope** to keep the library focused:

- **CLI tooling** — go-filewatcher is a library, not a CLI. Consumers build
  their own.
- **Database-backed event journaling** — out of scope; users compose
  `MiddlewareBatch` with their own persistence layer.
- **Cross-language bindings** — Go-only. Other languages should bind to fsnotify
  directly.
- **GUI / TUI** — not the library's job.

---

## Versioning Strategy

| Track         | Cadence                       | Trigger                                     |
| ------------- | ----------------------------- | ------------------------------------------- |
| Patch (x.y.Z) | As needed                     | Bug fixes, no API changes                   |
| Minor (x.Y.0) | Monthly–quarterly             | New options, filters, middleware (additive) |
| Major (X.0.0) | When breaking changes pile up | Removed deprecations, signature changes     |

See [API_STABILITY.md](./API_STABILITY.md) for the full stability policy. There
are 68 commits ahead of the latest tag (`v2.2.1`). The `[Unreleased]` CHANGELOG
section covers deprecations (`WithOnError`, `MiddlewareRateLimit`), the backend
test-seam abstraction, error simulation, CI automation (commitlint,
release-please, docs-consistency), and extensive lint/dedup hardening — a
`v2.3.0` would capture all of it. Release-please is wired in and will
auto-generate the release PR from conventional commits.
