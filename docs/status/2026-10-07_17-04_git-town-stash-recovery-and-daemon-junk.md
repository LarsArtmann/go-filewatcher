# Git-Town Stash Recovery, PR #60, and Daemon Junk Commit — Session Report

| | |
| --- | --- |
| **Date** | 2026-10-07 17:04 CEST |
| **Session scope** | Continuation of the 16:30 branch-convergence session: recover the status report swallowed by `git sync`'s WIP stash, triage + park the daemon's 70-file junk commit, land the report properly via PR #60 |
| **Start state** | master `115d8da` (daemon junk commit, 70 files, unpushable) + 2 uncommitted report files; report then swallowed into a "Git Town WIP" stash by the user's `git sync`; two rejected direct pushes |
| **End state** | master `1518cfe` == origin/master, clean; report landed via PR #60 (rebase-merged, all required checks green); junk commit preserved on `park/daemon-115d8da-70files`; only the pre-existing codespell stash remains |

---

## a) FULLY DONE

1. **PR #60 merged** — 16:30 status report + index row landed on origin/master as
   `1518cfe` via rebase merge; all 8 required checks green (incl. windows/macos
   legs). Evidence: merge output, master log, auto-deleted head branch.
2. **"Git Town WIP" stash forensics** — `git stash show --stat` proved
   stash@{0} contained exactly the two report files (212-line report + 1 README
   row) before recovery; files restored onto the PR branch from the stash, then
   the stash dropped (`c70ac5f`) — no hostage data left behind.
3. **Daemon junk commit `115d8da` triaged and parked** — verified it did NOT
   revert today's merged content (ci.yml probe legs present ✓, both status
   reports present ✓, TODO_LIST "Open items 39" ✓) before touching anything;
   preserved on branch `park/daemon-115d8da-70files` instead of deleted, since
   its 70 files may be another session's uncommitted work.
4. **227/14 diff anomaly root-caused before merge** — the commit's README diff
   showed 15 insertions / 14 deletions instead of the expected 1/1; reconciled
   exactly: 212 report lines + 1 new index row + 14 table rows re-padded by a
   formatter pass (daemon or concurrent session) that was sitting UNCOMMITTED
   in the working tree when I edited the file. No hidden content shipped.
5. **master realigned and verified** — `git switch -C master origin/master`
   after the merge; working tree clean; local branches = `master` + park
   branch only; zero open PRs; `git sync` is safe to run again.

## b) PARTIALLY DONE

1. **`115d8da` origin unknown** — triaged as content-safe and parked, but WHOSE
   70 files these are (which session/machine/process, and why
   `website/src/styles/global.out.css` shrank by ~1931 lines) is undetermined.
   Blocker: requires the daemon owner (user). Effort to close: S once answered.
2. **Daemon policy** — now THREE incidents today (mid-merge reset at ~16:05,
   70-file master commit at 16:30:48, uncommitted formatter debris that
   poisoned my PR diff). Still no decision; question asked twice, unanswered.
3. **16:30 report's open items** — its section (f) (34 tasks) and (g) (3
   questions) remain harvested-nowhere / unanswered; this report adds to the
   pile instead of closing any of it.

## c) NOT STARTED

1. **AGENTS.md gotcha: git-town sync does `git add -A` + `git stash`** — any
   uncommitted file becomes a "Git Town WIP" stash hostage on the next sync.
   Directly caused this session's recovery work; not yet written down.
2. **AGENTS.md gotcha: realign local master before `git sync`** — carried from
   the 16:30 report (b1); still unwritten; still the root cause of sync
   conflict-pause #2 and the rejected-push noise in pause #3.
3. **global.out.css lifecycle decision** — a ~1900-line generated artifact
   that origin/master CONTAINS while two separate changes today tried to
   revert it (the codespell stash from the 13:09 session, and the daemon's
   115d8da). Whether it should be committed, gitignored, or regenerated has
   never been decided; it keeps generating debris.
4. **End-of-turn "clean tree" invariant** — the rule "no session ends with
   uncommitted files on this machine" is identified as necessary but not
   implemented (no hook, no checklist item, no skill note).

## d) TOTALLY FUCKED UP

1. **Repeat of the exact failure mode I documented an hour earlier.** The
   16:30 report's section (e)5 says: "Every handoff message must cross-check
   the user's next likely command." I then ended the 16:30 turn with TWO
   uncommitted files on top of an unpushable master while the user's
   muscle-memory `git sync` was the obvious next command. It ran `git add -A`
   + `git stash` and swallowed the report. My own report file needed rescuing
   from the trap my own report described. Severity: the deliverable was one
   `git stash drop` away from being entombed.
2. **Committed unvetted working-tree changes into PR #60.** I caught the
   anomaly only because the insertion count looked wrong (227/14 vs expected
   213/1) and investigated AFTER push. The 14 reformatted rows were benign
   this time; they could have been anything — the daemon had just proven it
   writes to the working tree uninvited. My 16:30 (e)4 said "diff results
   before committing"; I did not diff before committing.
3. **Late detection of `115d8da`.** The daemon committed 70 files at 16:30:48
   — WHILE I was writing the report about its previous sabotage. I noticed it
   only via the trailing `git status` at the end of the report task, and the
   first thing the user's next command hit was that commit (rejected push).
   On this machine, `git log origin/master..master` belongs in every
   checkpoint, not in the epilogue.
4. **Missed failure surfaced only by user paste (again).** The failed edit
   ("file not found" — the report file had been stashed out from under me)
   and the swallowed stash were discovered because the user pasted their
   terminal output, not because I detected the state change. Both incidents
   today share this shape; the fix is the checkpoint habit from #3, not
   sharper forensics.

Honesty checks: no lies told; the 16:30 report's claims all held (verified
again during 115d8da triage). Split brain: local master (junk commit) vs
origin/master — third instance today, closed by realignment. Ghost systems:
the park branch is a deliberate, labeled parking lot — not a ghost; it has an
owner decision attached. No code was touched this session; no tests were at
risk (docs-only, tree validated at `369e0ac` == `f88be1a` content, report-only
commits since).

## e) WHAT WE SHOULD IMPROVE

1. **End-of-turn invariant: master == origin/master, tree clean — every
   session, every time.** On this machine any uncommitted file is one
   `git sync` away from a WIP stash, and any local master commit is one sync
   away from a conflict-pause. "Waiting for instructions" is not a safe
   parking state here; "merged and aligned" is.
2. **Stage-diff discipline: never commit without counting.** `git diff
   --cached --stat` reconciled against the intended change (files + lines)
   would have caught the 14 formatter lines pre-push.
3. **Daemon checkpoint: `git log origin/master..master` at every task
   boundary.** The daemon commits mid-task; noticing at the end costs a
   recovery cycle, noticing immediately costs one branch command.
4. **Park-branch pattern promoted to standard** — unknown-author commits get
   parked with a descriptive branch name, never deleted, never left on
   master. Worked cleanly this session.
5. **Formatter debris: identify the writer.** Something re-padded
   docs/status/README.md in the working tree without committing. If it is a
   concurrent session's formatter, fine — but then agents must treat the
   worktree as shared and RE-READ files immediately before editing (I did
   view before edit, which is the only reason the diff was reconcilable).
6. **Report-to-PR pipeline should be one motion** — write file → branch →
   commit → PR immediately. The gap between "write report" and "land report"
   is where both of today's incidents lived.

## f) Things to get done next (session-observed; ranked)

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Decide daemon policy NOW (3rd incident today): disable on master / scope to branches / retire — it committed 70 unvetted files to master and reset a merge earlier today | High | S | Decision |
| 2 | Identify the author of `115d8da`'s 70 files (another agent session? daemon sweep?) — then keep, cherry-pick, or drop `park/daemon-115d8da-70files` | High | S | Cleanup |
| 3 | Decide `global.out.css` lifecycle: generated ~1900-line artifact present on origin while two changes today reverted it — commit, gitignore, or build-step it away | High | M | Bug |
| 4 | PR: AGENTS.md gotcha — "git-town sync runs `git add -A` + stash: uncommitted files become 'Git Town WIP' stash hostages" | High | S | Documentation |
| 5 | PR: AGENTS.md gotcha — "realign local master to origin/master before `git sync` when local master carries unpushed commits" (carried from 16:30, still unwritten) | High | S | Documentation |
| 6 | Implement the end-of-turn invariant (clean tree + aligned master) as a hook or checklist step for agents on this machine | High | M | Quality |
| 7 | Run docs-health HARVEST: fold 16:30 (f) 34 items + this report's items into TODO_LIST.md / ROADMAP.md — two reports of tasks are currently entombed | High | M | Documentation |
| 8 | Ask-and-close the 3 carried questions from 16:30 (daemon policy, stash@{0}, merge convention) — daemon policy is now question #1 here too | Medium | S | Decision |
| 9 | Resolve `stash@{0}` (codespell-hygiene global.out.css reversion) — likely same answer as item 3; decide together | Medium | S | Cleanup |
| 10 | Document merge convention: rebase default, squash for conflict-carrying PRs (carried; both patterns used today) | Medium | S | Documentation |
| 11 | Run `nix flake check` on `1518cfe` for cheap certainty (content is report-only on top of the validated tree) | Low | S | Quality |
| 12 | Consider a daemon guard: block auto-commits touching >N files or master entirely (would have prevented 115d8da and the mid-merge reset) | Medium | M | Quality |
| 13 | Re-run `git town sync` once to confirm the steady state balances clean (post-#60) | Low | S | Quality |
| 14 | Watch the next release-please run: today landed many docs commits; confirm no spurious release PR (path-scoped excludes should prevent it) | Medium | S | Quality |
| 15 | Add `git log origin/master..master` to the session checkpoint ritual (skill/AGENTS note) — catches daemon commits at boundaries instead of epilogues | Medium | S | Quality |
| 16 | All 34 items in the 16:30 report's (f) section remain open — that list stays canonical for website/buildflow/CI-probe work; harvest first (item 7), then execute | High | L | (meta) |

(16 items; the 16:30 report's 34 remain canonical and are referenced, not
duplicated. Nothing here required researching beyond this session's events.)

## g) Questions I cannot answer myself

1. **Whose work is `115d8da`?** Did you (or another agent session on this
   machine) have ~70 files of uncommitted changes in this repo around
   16:30 — including a website docs sweep and the `global.out.css` shrink?
   I triaged it content-safe and parked it on `park/daemon-115d8da-70files`:
   keep, cherry-pick parts, or drop?
2. **Should `website/src/styles/global.out.css` exist in the repo at all?**
   Origin/master contains the full ~1900-line artifact; the 13:09 session
   stashed a reversion of it; the daemon's junk commit reverts it again.
   Generated-artifact-in-repo vs gitignore-and-build is a workflow decision
   the runbook may or may not settle — I have not researched it per your
   instruction to stay in-scope.
3. **Final call on the auto-commit daemon?** Third incident today. Options:
   (a) keep but hard-block master + large sweeps, (b) scope to dedicated
   branches, (c) retire it — agents commit deliberately. This is your
   infrastructure; every git guard I build on this machine depends on the
   answer.

---

*Point-in-time snapshot; section (f) is HARVEST input. Landing via PR #61
per the report-to-PR pipeline established this session. Waiting for
instructions.*
