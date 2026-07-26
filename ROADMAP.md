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
- **Benchmark freshness CI** — current benchmarks are saved as artifacts; add
  automated regression comparison against `main` with a tolerance. (Capturing a
  local baseline is a TODO_LIST quick win.)

### Reliability

- **Race-detector-on-CI flake quarantine** — two flaky tests are documented in
  AGENTS.md. The hardening work is tracked in TODO_LIST; the open question here
  is whether `t.Skip` with a tracked issue is preferable to making the
  assertions event-count-agnostic.

### Operational

- **Cross-platform release artifacts via goreleaser** — `.goreleaser.yml` is
  configured but **not invoked** by `release.yml` (which uses
  `softprops/action-gh-release` with auto-generated notes). Wiring goreleaser in
  would ship compiled cross-platform binaries on tag. See FEATURES.md
  "Cross-platform releases" (currently PARTIALLY DONE).
- **Dependency freshness SLO** — current policy is "update within 24h"; codify
  with Dependabot status checks.
- **Docs freshness gate** — a CI check that FEATURES.md/README.md hashes match
  the source API surface (e.g. generated from `go doc`). The concrete task is in
  TODO_LIST; the exploratory angle is auto-generating doc tables from source.
  See [docs/research/INDEX.md](./docs/research/INDEX.md) for related research.
- **Automated release tooling** — evaluate semantic-release or release-please
  for fully automated changelog and version bumping. See
  [docs/research/semantic-release-evaluation.md](./docs/research/semantic-release-evaluation.md)
  for the tradeoff analysis.

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
are 21 commits ahead of the latest tag (`v2.2.1`); a `v2.2.2` or `v2.3.0` would
capture the dedup, lint-clean, and example-hygiene work landed since.
