# Status Reports Index

Point-in-time session reports, newest first. Resolved reports are moved to
[`archived/`](./archived/) with their action items either done or harvested
into [`../../TODO_LIST.md`](../../TODO_LIST.md) — the archive is history, the
TODO list is the live backlog.

## Active (2026-10-07)

| Report | Topic | State |
| --- | --- | --- |
| [09-47 pareto-execution-ci-hardening](./2026-10-07_09-47_pareto-execution-ci-hardening.md) | CI hardening (floor guard, 26.04 readiness), event-channel race fix, docs debts; PRs #38–#44 | Merged through §f5; §f6+ in progress |
| [08-14 docs-health-full-archive-sweep](./2026-10-07_08-14_docs-health-full-archive-sweep.md) | All 2026-0* reports annotated & archived; full-archive sweep | Done; §b6 vendorHash re-verify open |
| [06-00 buildflow-green-pnpm-audit-zero-cache-war](./2026-10-07_06-00_buildflow-green-pnpm-audit-zero-cache-war.md) | BuildFlow green, pnpm audit zeroed via update+override, result-cache purge | Done; cache-key gap upstream (gated) |
| [05-17 buildflow-gomod-war-solved](./2026-10-07_05-17_buildflow-gomod-war-solved.md) | Root cause of the go.mod floor bumps (three tools, strace-proven) | Done; superseded by the structural floor guard (PR #38) |
| [04-34 buildflow-recovery-session](./2026-10-07_04-34_buildflow-recovery-session.md) | buildflow --fix recovery attempt | Superseded by 05-17 |
| [03-10 pr-review-ci-recovery-v2.4.1-release](./2026-10-07_03-10_pr-review-ci-recovery-v2.4.1-release.md) | PR review, master CI recovery, v2.4.1 release | Done; annotations resolved in PR #39 |
| [03-09 website-changelog-v2.4.1-incident](./2026-10-07_03-09_website-changelog-v2.4.1-incident.md) | Website fixes, changelog sync, v2.4.1 incident | Done |
| [02-24 v2.4.0-release-postmortem](./2026-10-07_02-24_v2.4.0-release-postmortem.md) | v2.4.0 release post-mortem & self-review | Done; sourced most of TODO_LIST's CI section |
| [00-42 go-daemon-synergy-readme-note](./2026-10-07_00-42_go-daemon-synergy-readme-note.md) | Auto-commit daemon + README note | Done |

## Archives

77 historical reports in [`archived/`](./archived/), oldest 2026-04-04. Row
uniformity is gate-checked (every action cell struck or `~~—~~`); the ~292
marker-free tables there are classified data-vs-open-task in TODO_LIST.

## Conventions

- File name: `YYYY-MM-DD_HH-MM_<slug>.md`, created at session end.
- Sections: a) fully done, b) partially done, c) not started, d) fucked up,
  e) process improvements, f) up to 50 next steps, g) questions.
- Move to `archived/` only when every §f item is done, harvested, or
  explicitly deferred into TODO_LIST.
