# Archived Status Reports

Closed history — every forward-looking item in these reports is resolved inline
(`~~strikethrough~~` verdicts citing evidence or routing). Do **not** harvest
from here; live work lives in [../../../TODO_LIST.md](../../../TODO_LIST.md).

**Archive manifests:**

| Sweep                        | Files | Classification → deciding reason                                                                                                                                                                               |
| ---------------------------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-10-07 docs-health sweep | 76    | eight batches below — see that sweep's own rows for the per-batch reason                                                                                                                                       |
| 2026-11-06 docs-health sweep | 15    | 13 × 2026-10-07 + 2 × 2026-10-12 reports — ARCHIVE: every §f item struck with its closing evidence or its routing target; `docs-gates freshness --audit-all` reports 0 unharvested items over the whole corpus |

**2026-10-07 sweep — per batch:**

| Batch                                     | Files                                                                            | Classification → deciding reason                                                                                       |
| ----------------------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| 2026-04-04 ×9, 2026-04-05 ×4              | initial implementation → retrospective                                           | ARCHIVE — all forward items done (≤v2.1.0), obsolete, or routed to TODO_LIST                                           |
| 2026-04-10 → 2026-04-15 ×15 + planning ×2 | branching-flow analyses, comprehensive statuses, task plans                      | ARCHIVE — race fixes, error system, phantom types, v0.1.0/v2.0.0 all shipped; unimplemented ideas consciously rejected |
| 2026-04-20 → 2026-05-04 ×9 + planning ×1  | post-v0.1.0 improvements, branded-id, gogenfilter v3, MIT/LICENSE                | ARCHIVE — branded-id reverted (a4d6f4c); v3 integration shipped; residue routed (goreleaser Q1, Windows CI, fuzz)      |
| 2026-05-22 → 2026-06-03 ×8 + planning ×2  | nix migration, module fix, v0.4.0-inotify (→v2.2.0), v2.2.0 release, pareto plan | ARCHIVE — releases shipped; inotify plan fully executed (never tagged v0.4.0)                                          |
| 2026-05-25 ×3, 2026-07-25/26 ×6           | quality sprint, dedup sessions, todo clearance                                   | ARCHIVE — prior-pass annotations completed; residue routed to TODO_LIST                                                |
| 2026-07-13 → 2026-07-29 ×12 + planning ×4 | docs-health passes, v2.3.0 release, filesystem-awareness series                  | ARCHIVE — shipped in v2.3.0/v2.4.0; v3 residues already in TODO_LIST/ROADMAP                                           |
| 2026-08-12 ×4                             | reliability edge-case series (3 sessions + feedback completion)                  | ARCHIVE — v2.4.0 shipped the options/stats; poll dedup/rename detection → ROADMAP                                      |

**2026-11-06 sweep — per file:**

| File                                               | Classification → deciding reason                                                                      |
| -------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| 2026-10-07_00-42 go-daemon synergy note            | ARCHIVE — §f was self-declared brainstorm fuel; the two live threads sit in TODO_LIST open question 7 |
| 2026-10-07_02-24 v2.4.0 release postmortem         | ARCHIVE — release shipped; follow-ups done or routed (required checks, dependabot, bench, fuzz)       |
| 2026-10-07_03-09 website changelog v2.4.1 incident | ARCHIVE — incident fixed (path-scoped release-please, generated changelog); residues in TODO_LIST     |
| 2026-10-07_03-10 PR review + v2.4.1 release        | ARCHIVE — review and release done; the two standing lessons recorded in AGENTS.md                     |
| 2026-10-07_04-34 buildflow recovery attempt        | ARCHIVE — superseded by 05-17 (same afternoon)                                                        |
| 2026-10-07_05-17 go.mod war solved                 | ARCHIVE — root cause attributed and structurally guarded (go-directive job); dep upgrades verified    |
| 2026-10-07_06-00 buildflow green + pnpm audit war  | ARCHIVE — pipeline green, 0 advisories; cache-key gap routed to TODO_LIST                             |
| 2026-10-07_08-14 docs-health full archive sweep    | ARCHIVE — that sweep's own tail closed by the 2026-11-06 sweep                                        |
| 2026-10-07_09-47 pareto execution + CI hardening   | ARCHIVE — floor guard, race fix, docs debts landed; quiet batch in TODO_LIST                          |
| 2026-10-07_13-09 CI hardening continuation         | ARCHIVE — v2.4.2 shipped; html-validate routed (still unwired)                                        |
| 2026-10-07_15-21 resume + debris cleanup           | ARCHIVE — stashes dropped, /tmp swept, hashes verified                                                |
| 2026-10-07_15-51 git-town debris + daemon recovery | ARCHIVE — branches deleted with proofs; branch-protection record in docs/ci-workflows.md              |
| 2026-10-07_16-30 branch convergence session        | ARCHIVE — all branches integrated; daemon/CSS decisions live in TODO_LIST                             |
| 2026-10-12_02-25 event-mapping docs audit          | ARCHIVE — audit done; its harvest + VERIFY items closed by the 2026-11-06 pass                        |
| 2026-10-12_03-10 oxlint website lint hardening     | ARCHIVE — nested oxlint config shipped; `/tmp` cleanup verified; residues routed                      |

**Completeness:** every file carries ≥1 inline strikethrough; `grep -rLn '~~'`
over this directory returns nothing. A handful of micro-items retain older
resolution markers from the 2026-07-26/27 passes instead of the sweep format —
those are equally closed. Marker-free **data** tables (session metadata, state
tables) stay in place by design: they record what happened, not open work.

**Not archived (active):** the four `docs/status/2026-11-06_*` reports from the
health-API, release, and docs-ratchet sessions.
