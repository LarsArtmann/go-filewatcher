# Status Reports Index

Point-in-time session reports, newest first. Resolved reports are moved to
[`archived/`](./archived/) with their action items either done or harvested
into [`../../TODO_LIST.md`](../../TODO_LIST.md) — the archive is history, the
TODO list is the live backlog.

## Active (2026-11-06)

| Report                                                                                                           | Topic                                                                                                                                                                          | State                                                                                                     |
| ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------- |
| [08-20 docs-health ratchet drain](./2026-11-06_08-20_docs-health-ratchet-drain-and-archive-sweep.md)             | Full corpus audit: 1765 destroyed table rows rebuilt from `3b56ef0^`, annotate 1786 → 0, audit-all 16 → 0, 15 reports archived, TODO_LIST deduplicated, 83 archive links fixed | Done; 4 follow-ups queued (gate wiring, routing review, planning pass, dprint stragglers)                 |
| [07-54 v2.5.0 shipped + consumer caveat](./2026-11-06_07-54_v2.5.0-release-shipped-and-consumer-caveat.md)       | v2.5.0 tagged and published, proxy listing proof, consumer build against the released health API, CHANGELOG/manifest drift repaired                                            | Done; no §f section (release record only); follow-ups in TODO_LIST                                        |
| [07-22 buildflow-green lint fix + docs ratchet](./2026-11-06_07-22_buildflow-green-lint-fix-and-docs-ratchet.md) | BuildFlow exit 69 → 0: varnamelen rename in `health_test.go`, AGENTS.md 266 → 179, flake app meta, pnpm audit 19 → 0, docs-gates 4 red → 5 PASS, 22 findings baselined         | Steps done; f1–f26 harvested into TODO_LIST; 3 owner questions open (skips, 350-line policy, runner pins) |
| [06-46 health API session](./2026-11-06_06-46_go-filewatcher-health-api-session.md)                              | `Healthy(ctx)` + introspectable circuit breaker shipped                                                                                                                        | Done; §f27 (graduate-to-Stable soak policy) harvested into TODO_LIST                                      |
| [06-37 otel audit + go floor restore](./2026-11-06_06-37_otel-audit-and-go-floor-restore.md)                     | OTel support audit, `go 1.26.7` floor restored                                                                                                                                 | Done; `nix run .#lint` + `nix flake check` verified 2026-11-06                                            |

## Archives

92 reports from 2026-04-04 → 2026-10-12 in [`archived/`](./archived/). Two
batches moved on 2026-11-06 after the docs-health sweep struck every
forward-looking item with its evidence or its routing target:

- **13 reports from 2026-10-07** — the v2.4.1/v2.4.2 release, website, and
  buildflow-war sessions;
- **2 reports from 2026-10-12** — the event-mapping docs audit and the oxlint
  website-lint hardening.

Those §f tables carry `<!-- check-rows:ledger -->` because their per-row verdicts
are the point of the table.

Marker-free **data** tables (session metadata, state tables) remain in archived
reports by design — they record what happened, not open work. The
data-vs-open-task classification of the older corpus is tracked in TODO_LIST.

## Conventions

- File name: `YYYY-MM-DD_HH-MM_<slug>.md`, created at session end.
- Sections: a) fully done, b) partially done, c) not started, d) fucked up,
  e) process improvements, f) up to 50 next steps, g) questions.
- Move to `archived/` only when every §f item is done, harvested, or explicitly
  deferred into TODO_LIST — and after `docs-gates run` is green.
