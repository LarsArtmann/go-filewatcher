# Archived Status Reports

Closed history — every forward-looking item in these reports is resolved inline
(`~~strikethrough~~` verdicts citing evidence). Do **not** harvest from here;
live work lives in [../../TODO_LIST.md](../../TODO_LIST.md).

**Archive manifest (2026-10-07 docs-health sweep — 76 files):**

| Batch | Files | Classification → deciding reason |
| --- | --- | --- |
| 2026-04-04 ×9, 2026-04-05 ×4 | initial implementation → retrospective | ARCHIVE — all forward items done (≤v2.1.0), obsolete, or routed to TODO_LIST |
| 2026-04-10 → 2026-04-15 ×15 + planning ×2 | branching-flow analyses, comprehensive statuses, task plans | ARCHIVE — race fixes, error system, phantom types, v0.1.0/v2.0.0 all shipped; unimplemented ideas consciously rejected |
| 2026-04-20 → 2026-05-04 ×9 + planning ×1 | post-v0.1.0 improvements, branded-id, gogenfilter v3, MIT/LICENSE | ARCHIVE — branded-id reverted (a4d6f4c); v3 integration shipped; residue routed (goreleaser Q1, Windows CI, fuzz) |
| 2026-05-22 → 2026-06-03 ×8 + planning ×2 | nix migration, module fix, v0.4.0-inotify (→v2.2.0), v2.2.0 release, pareto plan | ARCHIVE — releases shipped; inotify plan fully executed (never tagged v0.4.0) |
| 2026-05-25 ×3, 2026-07-25/26 ×6 | quality sprint, dedup sessions, todo clearance | ARCHIVE — prior-pass annotations completed; residue routed to TODO_LIST |
| 2026-07-13 → 2026-07-29 ×12 + planning ×4 | docs-health passes, v2.3.0 release, filesystem-awareness series | ARCHIVE — shipped in v2.3.0/v2.4.0; v3 residues already in TODO_LIST/ROADMAP |
| 2026-08-12 ×4 | reliability edge-case series (3 sessions + feedback completion) | ARCHIVE — v2.4.0 shipped the options/stats; poll dedup/rename detection → ROADMAP |

**Completeness:** every file carries ≥1 inline strikethrough; `grep -rLn '~~'` over this
directory returns nothing. A handful of micro-items retain older resolution markers from
the 2026-07-26/27 passes instead of this sweep's format — those are equally closed.

**Not archived (active):** `docs/status/2026-10-07_*` (seven reports from the v2.4.x
release, website, and buildflow sessions of 2026-10-07).
