# Roadmap

**Last Updated:** 2026-07-13

Long-term direction and raw ideas for go-filewatcher. Items here are **not yet
committed** — they graduate to [TODO_LIST.md](./TODO_LIST.md) when scoped and
actionable.

For the current feature snapshot, see [FEATURES.md](./FEATURES.md).

---

## Themes

1. **Hardening** — broader platform coverage, error-path testing, performance regressions
2. **Ecosystem integration** — first-class support for observability stacks and DI patterns
3. **API evolution** — toward v3 with feedback-driven breaking changes
4. **Operational excellence** — automated releases, dependency hygiene, docs freshness

---

## Ideas Worth Exploring

### Platform Coverage

- **Windows CI matrix** — current CI is Linux-only; Windows has different event
  semantics (no inotify, different rename behavior). Worth at least smoke tests.
- **macOS FSEvents edge cases** — rename semantics differ from Linux; document
  or test the divergences.
- **BSD/kqueue** — fsnotify supports it; verify our assumptions hold.

### Testing & Reliability

- **Error simulation framework** — inject fsnotify failures (ENOSPC, permission
  denied, watcher closed) to exercise error middleware paths deterministically.
- **Property-based / fuzz testing expansion** — current fuzz tests cover
  `ParseFamily`, `Classify`, error formatting. Expand to filter combinators,
  event marshaling, gitignore matching.
- **Large-tree stress harness** — synthetic 100k-directory tree to validate the
  inotify budget, batched registration, and self-heal under load.
- **Race-detector-on-CI flake quarantine** — two flaky tests are documented in
  AGENTS.md. Either harden them or mark them as `t.Skip` with a tracked issue.

### API Evolution

- **v3 planning** — accumulated deprecations (`WithWatchedIgnoreDirs`) and
  awkward signatures (`ErrorHandler` two-arg form) suggest a v3 cleanup pass.
  Gather breaking changes over the next 6-12 months before cutting.
- **Idempotent sync API** — `WatchChanges(ctx, targetState)` that emits events
  until the filesystem matches a declared target. Useful for sync/backup tools.
- **Streaming filter protocol** — current `Filter` is sync bool. Consider
  returning `(keep bool, err error)` or a channel-based variant for filters
  that need async I/O (e.g., remote manifest lookup).

### Observability

- **Prometheus default labels** — `PrometheusCollector` is pluggable but
  requires user wiring. Ship a `MustRegister(coll, opts...)` helper that
  attaches standard namespace/subsystem.
- **OpenTelemetry default exporter wiring example** — `OTelMiddleware` is
  zero-dependency but the README does not show end-to-end tracing setup.
- **pprof endpoints for watcher introspection** — expose watch-list size,
  debouncer queue depth, filter rejection counts via `net/http/pprof`.

### Performance

- **Zero-allocation event path** — current `ConvertEvent/Create` is 3 allocs;
  investigate pooling or stack-allocated `Event` for hot paths.
- **Lazy filter evaluation short-circuit** — `FilterAnd` evaluates all filters
  today; consider returning on first `false` for measurably cheaper composition.
- **Benchmark freshness CI** — current benchmarks are saved as artifacts;
  add automated regression comparison against `main` with a tolerance.

### Operational

- **Automated release pipeline** — tag-triggered release via `release.yml` is
  wired (tests + lint + GitHub Release). Consider adding cross-compilation
  artifacts via goreleaser to the workflow.
- **Semantic-release / conventional commits** — evaluate whether commit-message-
  driven versioning reduces manual CHANGELOG drift.
- **Dependency freshness SLO** — current policy is "update within 24h"; codify
  with Dependabot status checks.
- **Docs freshness gate** — add a CI check that FEATURES.md/README.md hashes
  match the source API surface (e.g., generated from `go doc`).

---

## Non-Goals

These are explicitly **out of scope** to keep the library focused:

- **CLI tooling** — go-filewatcher is a library, not a CLI. Consumers build their own.
- **Database-backed event journaling** — out of scope; users compose `MiddlewareBatch`
  with their own persistence layer.
- **Cross-language bindings** — Go-only. Other languages should bind to fsnotify directly.
- **GUI / TUI** — not the library's job.

---

## Versioning Strategy

| Track         | Cadence                       | Trigger                                     |
| ------------- | ----------------------------- | ------------------------------------------- |
| Patch (x.y.Z) | As needed                     | Bug fixes, no API changes                   |
| Minor (x.Y.0) | Monthly–quarterly             | New options, filters, middleware (additive) |
| Major (X.0.0) | When breaking changes pile up | Removed deprecations, signature changes     |

See [API_STABILITY.md](./API_STABILITY.md) for the full stability policy.
