# Docs-Health Ratchet Drain: 1786 → 0 Findings, 15 Reports Archived

**Date:** 2026-11-06 08:20 CET
**Session scope:** docs-health AUDIT over the whole documentation corpus (living
docs + 92 archived reports + the active status batch), plus the ratchet drain the
2026-11-06 buildflow run had queued: `docs-gates` annotate 1786 → 0 and audit-all
16 → 0, then archive everything provably resolved.
**Outcome:** `docs-gates run .` is 5/5 PASS with **zero** tolerated findings
(`audit_max: 0`, `annotate_max: 0`) and no ratchet slack, `nix flake check` and
`nix run .#check` are green, and 15 status reports moved into `archived/`.

| Signal              | Before                          | After                              |
| ------------------- | ------------------------------- | ---------------------------------- |
| annotate findings   | 1786 (pinned at 1786)           | **0** (pin 0)                      |
| audit-all           | 16 unharvested (pinned at 17)   | **0** (pin 0)                      |
| cadence             | 17 reports unarchived, 30 days  | **4** reports, 0 days (pin 8 / 21) |
| archived reports    | 77                              | **92**                             |
| live TODO items     | 39 claimed / ~58 actual (dupes) | **64** (deduplicated, each once)   |
| dead relative links | 83+                             | **0**                              |

---

## a) FULLY DONE

1. **Recovered 1765 rows of silently destroyed content.** The archived tables had
   been rewritten as `| ~~ | ~~N~~ | ~~task~~ |` — a whole-row strikethrough
   converted per cell by a 2026-10-07 sweep that **dropped every trailing cell and
   every verdict** (Resolution columns, `File(s)` columns, `30min` efforts, and
   the `done — …` evidence all gone). Rows were rebuilt from the pre-sweep tree
   (`3b56ef0^`) and re-struck per cell, preserving every original cell plus the
   verdict. 1765 rows / 154 tables / 49 files. The `| ~~ |` "bare tilde" defect
   those rows carried — and the 51 blockquotes that documented it as accepted
   residue — are gone.
2. **Drained the annotate ratchet from 1786 to 0 findings** and pinned
   `annotate_max: 0`. Work included: 32 `OBSOLETE —` verdicts translated into the
   gate's marker vocabulary (`no longer applicable —`), 33 open items un-struck
   and re-routed (a struck row with an "OPEN →" tail is a lie the gate catches),
   2 rows rebuilt around an escaped `\|` that had split them into pieces, 7 prose
   bullets given real verdicts from their report's own Resolution section, 4 prose
   verdicts given marker words (`answered` → `decided`, `addressed` → `done`,
   `routed:` → `superseded — routed:`), and 5 backticked plurals de-glued
   (`` `Event`s `` → `` `Event` values ``) so rule F stops firing.
3. **Drained audit-all from 16 to 0 by verifying, not by asserting.** Each item
   got evidence I actually read: `website/pnpm-lock.yaml` resolves
   `fast-uri@3.1.9`, `source-map-js@1.2.2`, `http-cache-semantics@4.3.0`; all four
   hashes cited in the 09:47 report resolve (`git cat-file`); `docs/ci-workflows.md`
   carries the branch-protection record plus the restore JSON; ROADMAP links
   `docs/research/v3-decisions.md`; `nix run .#lint` reports **0 issues** and
   `nix flake check` passes **all 9 checks**; `/tmp/index.astro.bak` and
   `/tmp/oxlint-fresh.txt` are gone. Two genuinely open items (Lighthouse audit,
   health-API soak policy) were **harvested into TODO_LIST** rather than closed.
4. **Archived 15 resolved reports** — 13 from 2026-10-07 (release/website/buildflow
   war batch) and 2 from 2026-10-12 (event-mapping audit, oxlint hardening) — each
   with an `ARCHIVED 2026-11-06` banner, ledger-marked per-row verdict tables, a
   per-file manifest in `docs/status/archived/README.md`, and a rewritten
   `docs/status/README.md` (4 active reports, archive pointer, conventions).
   Cadence pins tightened 19/30 → 8/21.
5. **Rebuilt `TODO_LIST.md`.** The 2026-11-06 "session backlog" section repeated
   15+ items that already lived in the thematic sections (CSS direction,
   actionlint, weekly `-race`, `reliability_semantics_test.go`, public-presence
   polish, file-size campaign, each twice). All `[x]` done rows were removed (the
   skill's rule: a completed TODO belongs in CHANGELOG, never in the list), the
   two resolved ratchet campaigns (docs-annotate, archive ritual) dropped, the
   stale "Urgent: Ubuntu 26 audit — deadline 2026-10-12" rewritten as the
   2026-11-19 re-evaluation, and the snapshot recounted from repo state (64 open
   items, coverage gate 85%, tests passing).
6. **Fixed 83 broken archive links.** Every archived report's
   `[TODO_LIST.md](../../TODO_LIST.md)` was one directory short (archived/ is three
   levels deep); the corpus-wide relative-link scan is now 0 dead links.
7. **Corrected living-doc drift** — FEATURES header `v2.4.1` → `v2.5.0`, the false
   "Windows CI: currently CI runs Linux only" row replaced with the probe-leg
   reality, `30 examples` → 29, `90%` coverage → the CI's actual **85%**, README
   `18 middleware` → 19 and the filter claim made exact (`20 + combinators`),
   ROADMAP's two stale "Linux-only CI" claims and its v2.4.1-era version paragraph
   rewritten, CHANGELOG's `[Unreleased] / Nothing yet.` stub and the double blank
   line after `## [2.5.0]` removed.
8. **Gave AGENTS.md the docs-gates knowledge it was missing** — command block plus
   a gate/fix table (freshness, audit-all, annotate, routing, cadence) and the two
   rules that cost real time today: strike **per cell**, and **format, then re-run
   annotate** (dprint rewrites markdown and can break markers it does not
   understand). 207 lines, budget 220.
9. **Verified the quality gates end to end**: `docs-gates run .` 5/5 PASS,
   `nix run .#check` (vet + lint 0 issues + `-race` tests) green, `nix flake check`
   all checks passed, `dprint` clean on every file I touched.

## b) PARTIALLY DONE

1. **457 routing verdicts are machine-verified, not human-verified.** Each struck
   item says "no longer open here — routed to …" because
   `docs-gates freshness --audit-all` proved its tokens exist in the living docs.
   That is the repo's own standard, but a wrong routing _target_ would read exactly
   like a correct one. Spot-checking the batch is queued (TODO_LIST).
2. **`docs/planning/2026-10-07_08-17_master-pareto-execution-plan.md`** — 30 days
   old, 150 task rows, never given per-task evidence. It now carries an honest
   "superseded, kept for provenance" banner instead of a fake archive; the real
   decision (reconstruct verdicts vs archive with the banner as manifest) is
   queued.
3. **`dprint check` is not trustworthy as a local gate** — it flags
   `.config/metadata.yaml` (4-space indent) and `website/pnpm-lock.yaml`, both
   pre-existing and both passing the project's `treefmt`. I left them alone; the
   choice (format vs explicit `excludes`) is queued.
4. **AGENTS.md now describes the docs-gates workflow that is NOT enforced.** The
   binary is a local install (`~/.local/bin/docs-gates`); there is no flake check,
   no CI job, no BuildFlow step. I softened the wording to say so, but the gate is
   still only as reliable as the human who remembers to run it.

## c) NOT STARTED

1. **Wiring `docs-gates` into `flake.nix` checks and CI** (the single highest-value
   follow-up: it converts today's discipline into a machine gate).
2. **Per-item human review of the 457 routed verdicts.**
3. **Archival decision for the 150-row Pareto plan.**
4. **A check-rows/bare-rows sweep over the older archive corpus** (77 pre-2026-11-06
   files) — `check-rows` was run on today's moved files only.
5. **CHANGELOG entry for this documentation work** — no release was cut, so the
   work is only in git history and this report.

## d) TOTALLY FUCKED UP!

Not "fucked up" as in broken, but four things I got wrong and had to repair inside
this session — worth recording because they are the failure modes of this ritual:

1. **I trusted my own keyword matcher and it wrote a false verdict.** The
   `source-map-js` done-pattern also matched an item that merely _listed_ the
   package ("enumerate the 9 pnpm-audit errors (js-yaml, fast-uri, source-map-js,
   …)"), so it claimed a version bump as the evidence for an enumeration task.
   Worse, `html-validate` was marked **done** by the "website CI build job" pattern
   even though html-validate is still unwired. Both were caught by re-reading, both
   corrected — but keyword-matching historical text is exactly how a docs audit
   launders a lie.
2. **I reverted my own committed-in-flight work with `git restore`.** I assumed my
   `fix_audit_items.py` edits were uncommitted; the auto-commit daemon had already
   landed them, so `git restore --source=HEAD` was a no-op for those and I had to
   redo the reasoning. Lesson: with a commit daemon running, `git status` is not a
   reliable statement about what "my uncommitted work" is.
3. **My first repair shape was wrong and I shipped it before testing it.**
   Attempt #1 struck table rows and _hoisted_ each table's shared verdict into one
   `> **Table verdict …:**` line — cleaner to read, and **1649 A/evidence findings
   appeared**: the gate wants the verdict marker outside the strike, per row. I ran
   the gate, saw 1649, reverted all 49 files and rebuilt with a per-row verdict.
   The lesson is the skill's own: run the gate on a one-file sample before a
   49-file mutation.
4. **I overwrote `TODO_LIST.md` from memory of a partial read** and the editor
   refused ("modified since last read") — correct behaviour, but it shows I was
   about to replace a 537-line living doc while holding only chunks of it in
   context. A full read preceded the successful write, but that was luck of the
   guard, not discipline.

Also worth flagging honestly: **I claimed a gate that does not run.** The first
version of the AGENTS.md section I wrote said `docs-gates run .` "is a real quality
gate, not advisory" — in a repo where nothing runs it automatically. Softened, and
the wiring is now a TODO.

## e) WHAT WE SHOULD IMPROVE!

1. **Make the documentation gate mechanical, then stop writing prose about it.**
   Every docs-health session rediscovers the same ratchets by hand; a
   `checks.docs-gates` flake entry + one CI job removes the whole class.
2. **Never let a sweep destructively re-shape tables.** The 2026-10-07 pass lost
   the content of 1765 rows because a script converted `~~| row |~~` per cell
   without a round-trip assertion. Any future table rewrite must assert
   `cells(before) == cells(after)` (order and count) before writing.
3. **Verify twice, and with the gate rather than by eye.** The output of this
   session was correct only because each repair was followed by `docs-gates annotate`
   and `freshness --audit-all`. The two false verdicts (d1) survived _my_ reading
   and died only to a second pass.
4. **Keywords are not evidence.** A done-pattern should require the item's own
   subject (version bumped, file created) to appear in the verdict, not a package
   name that the item merely mentions.
5. **Keep the archive honest by policy, not by pressure.** Cadence thresholds pushed
   the batch out at exactly 30 days; the right response was the annotate work
   (which is what happened), not a raised pin. Make that sentence a comment in
   `.docs-gates.json` so the next session does not "solve" it by editing the number.
6. **Report the ratio.** 1765 restored rows vs 1 broken linter link is the shape of
   this work — silent content loss is far more dangerous than any finding the gate
   prints, and no gate would ever have caught it.

## f) NEXT — up to 50, honest count: 34

P1 (highest value first):

1. Wire `docs-gates` into the flake checks and CI (`checks.docs-gates` + one job) so
   a docs regression fails a build instead of waiting for the next audit.
2. Hand-verify the routed verdicts in the 2026-10 archive batch (457 items) and
   correct any wrong routing target.
3. Give `docs/planning/` the same archival pass — the 150-row Pareto plan either
   gets per-task verdicts or moves to `docs/planning/archived/` with its banner as
   the manifest row.
4. Settle the two dprint-unclean files (`.config/metadata.yaml`,
   `website/pnpm-lock.yaml`): format them or exclude them explicitly.
5. Run `check-rows` / `bare-rows` over the 77 older archived reports (today's sweep
   gated only the 15 moved files) and classify the remaining marker-free data tables.
6. Verify the 2026-11-06 session's changes on CI, including the Windows and macOS
   probe legs.
7. Ubuntu 26 runner re-evaluation after the 2026-11-19 rollout; then delete the two
   `cue-vet` baseline entries.
8. BuildFlow binary upgrade + skip-list re-verify; triage the 9-tools-unavailable
   warning, `go-auto-upgrade` ×12, `nix-checker` ×8, `vulnix`.
9. Add a CI check that `.release-please-manifest.json` matches the newest `v*` tag.
10. Confirm the release-commit CI run is green (v2.5.0 merge `b77b280` + `2bd47ff`).
11. Rebuild and deploy the website so `/changelog` shows 2.4.3 + 2.5.0.
12. Document the health API on the website (it is invisible on
    filewatcher.lars.software today).
13. Rewrite the published v2.4.3 release notes (they list PRs that shipped in 2.4.2).
14. Re-check pkg.go.dev shows `v2.5.0` and record the confirmation.
15. Release a documentation-only CHANGELOG note so this sweep is discoverable
    without reading git history.
16. Website link checker (`html-validate` is still unwired, as corrected above).
17. Record a Lighthouse audit of the built site.
18. Split the 20 oversized Go files (file-size campaign) — start with
    `watcher_test.go` (1860 lines).
19. Decide whether test files are exempt from the 350-line budget and ask BuildFlow
    upstream for a per-glob limit instead of inflating the number.
20. Add `actionlint` to pre-push verification.
21. Decide on a weekly `-race -count=2` CI run.
22. Re-drive seed `1793946570936674137` under concurrent load and harden the flake.
23. Large-tree stress harness (100k-directory fixture).
24. Investigate the July 6-hour Benchmark hang.
25. Per-OS test expectations replacing the `skipOnWindows` skips, then promote the
    probe legs to required checks.
26. Expand fuzz tests (combinators, Event JSON round-trip, gitignore matcher).
27. Add a compiled `examples/otel` so the README pattern cannot rot.
28. Add a `Healthy`-in-action example under `examples/`.
29. Health API graduate-to-Stable soak policy note in `API_STABILITY.md`.
30. Document the `MiddlewareCircuitBreaker` → `NewCircuitBreaker` migration on the
    website.
31. Cross-repo (crush-config): `lessons.md` entry + buildflow skill DB path fix.
32. **errata:** the plan says "Pareto first" but the _first_ action that paid off
    today was the mechanical repair — decide whether `docs-health` should run the
    gate repair before the prose audit in every future sweep.
33. Keep sweeping: move each newly created report into `archived/` within 21 days
    with per-item verdicts, so the cadence never has to be raised.
34. Add a short "how to close a report" recipe next to `docs/status/README.md`
    conventions — annotate → `check-rows` → `dprint` → `check-rows` again →
    `docs-gates run` → `git mv`.

## g) QUESTIONS — only the owner can answer (3)

1. **Should `docs-gates` become a hard build gate in this repo?** Wiring it into
   `flake.nix` checks + CI means a documentation defect blocks a release PR
   (including the automated release-please one). The alternative is keeping it a
   pre-commit discipline. Your call on strictness vs noise during release weeks.

2. **How should the 457 "routed" verdicts be treated — as sufficient, or do you
   want the specific TODO_LIST row cited in each verdict?** Citing a row per item is
   verifiable but means ~450 edits and volatile line references; the current form
   names the living doc and relies on the audit gate's token proof.

3. **Do you want the 2026-10-07 Pareto plan (`docs/planning/`, 150 rows) and the
   broader archive treated as "closed history with a banner" or as "per-task evidence
   required"?** That answer decides whether the remaining corpus work is a one-hour
   move or a multi-session reconstruction, and it sets the precedent for
   `docs/planning/` generally.
