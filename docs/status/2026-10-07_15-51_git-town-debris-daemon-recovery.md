# Status Report: Git Town Debris Sweep + Daemon Sabotage #5 Recovery

**Session:** 2026-10-07, ~15:22–15:51 CEST (incident segment after the 15:21
resume report; triggered by a `git town sync` conflict-pause on zombie
branches left by today's 18-PR rebase-merge wave)
**Machine:** shared box, load 20–80 on 32 cores (still above the <8 gate —
the quiet-machine batch stayed deferred). The auto-commit daemon struck
**once more during recovery** (sabotage #5 today, see §d) and aborted a
rebase mid-flight.

---

## a) FULLY DONE

### 1. Git Town paused-sync unwound cleanly

`git sync` (= `git town sync`) merges master into **every** local branch and
had conflict-paused on `chore/codespell-hygiene` (AGENTS.md + status-index
conflicts). `git town undo` returned to the exact pre-sync state (its WIP
stash popped back, merge aborted) — zero content lost, no manual conflict
resolution on a dead branch.

### 2. The Git Town WIP stash identified and preserved — it was NOT junk

Inspection showed the stash was the **`website/src/styles/global.out.css`
artifact reversion** (−1930 tracked lines) — i.e. real §f-22 cleanup work
from an earlier session, never committed. Re-stashed under a self-describing
name: `global.out.css artifact reversion (13:09 f22) — from
chore/codespell-hygiene`. Nothing orphaned.

### 3. All 18 zombie branches deleted, each with upstream-content proof

Local branches from the whole merge wave (`chore/codespell-hygiene`,
`ci/release-goexperiment`, `ci/website-build*` ×3, `ci/floor-guard-rollout`,
`docs/agents-split`, `docs/bookkeeping`, `docs/consistency-exemption-zero`,
`docs/status-1309-report`, `docs/status-index`, `docs/sweep-followups`,
`docs/t21-verification`, `docs/tick-website-ci`, `docs/todo-ticks`,
`docs/v3-decisions`, `test/reliability-semantics`). Method:
`git cherry master <branch>` per branch — `-` = patch-equivalent upstream
(rebase-merge-safe proof). The five `+` outliers each inspected individually:
three daemon auto-commit tips (junk snapshots, file-signature-matched to
already-merged work) and two real commits proven superseded by content diff
(master carries the newer pinned workflows; identical exemption content).
Plus `git worktree prune` — three stale `/tmp/gfw-*` registrations blocked
branch deletion (dirs already gone from disk). Repo is now **master + the
two open PR branches** — `git town sync` can never trip again.

### 4. Daemon sabotage #5 fully recovered

Mid-recovery the daemon committed a **68-file junk snapshot onto local
master** (`1763aec`, +2191/−4078 — zombie-tree resurrection incl. the css
artifact re-add and AGENTS.md churn). My PR branch forked from it before I
saw it; the daemon then **aborted my interactive rebase** mid-resolution.
Recovery used the atomic pattern: branch reset onto clean `d8c05cc` →
scripted AGENTS.md row insert → commit → `push --force-with-lease`, **all in
one bash invocation** (no gap for the daemon). Local master reset (--mixed,
AGENTS-allowed) to `d8c05cc`; the 68-file residue wiped after signature
verification (css reversion already safe in the named stash).
**Origin/master was never touched** — branch protection held again.

### 5. PR #58 — the incident's lesson recorded in AGENTS.md

New "Git Town syncs ALL local branches" gotcha row: debris sweeps must
enumerate `git branch -vv` (upstream `gone`) + `git worktree prune`, and
prove content with `git cherry` before `git branch -D`. **OPEN, CLEAN, all
checks green, exactly 1 commit** (after the sabotage rebuild).

### 6. PR #57 verified merge-ready

The 15:21 status-report PR: GitGuardian (the only non-required pending
check) resolved; **CLEAN, 1 commit**, all 16 required contexts green.

---

## b) PARTIALLY DONE

1. **Branch-debris program** — local branches and worktree registrations
   done; the **remote mirrors** remain on origin for the branches that
   merged without auto-delete (`ci/floor-guard-rollout`,
   `docs/bookkeeping`; most others show `gone`). Remote deletion is a push
   operation — gated on your OK (§g2).
2. **§f-22 `global.out.css` artifact** — the reversion is now *preserved and
   labeled*, but the actual fix (untrack + gitignore vs keep) is still an
   open decision; the stash should not outlive the decision.
3. **Incident forensics** — the failure *pattern* is now precisely
   documented (daemon junk commits on local master + resets; Git Town
   all-branch sync amplifies any stale branch), but the daemon's process
   identity/config remains unknown (§g1) — I still cannot see what runs it.
4. **Harvest** — this segment produced no TODO_LIST deltas (no §f item
   fully closed), so the snapshot count stays 39; the lessons live in PR
   #58's AGENTS.md row and this report.

---

## c) NOT STARTED

- **Quiet-machine batch** (gate load < 8; segment load 20–80): **T5** `nix
  flake check` + second vendorHash; **T16** BuildFlow triage on
  `acdb606-dirty`; **T22** bench re-capture; **T23** fuzz expansion; **T24**
  stress harness.
- **§f-3 from the 15:21 report: BuildFlow stale pnpm-audit cache purge** —
  flagged as load-independent **two reports in a row** now and still not
  executed. This is the segment's #1 self-criticism (§e-1).
- **User-gated**: daemon fate (§g1), remote-branch deletion + merge
  authorization (§g2), batch scheduling (§g3), **T17** upstream BuildFlow
  filings.
- **§f tail** (unchanged): html-validate wiring, Starlight editLink,
  release smoke-test script, per-OS expectation rewrites (34 skips),
  Website Build → required check, leg-promotion decision 2026-10-14, and
  the rest enumerated in §f.

---

## d) TOTALLY FUCKED UP

1. **The daemon's sabotage #5 hit during MY recovery operation** — junk
   commit onto local master, inherited by my fresh PR branch, then aborted
   my paused rebase. Recovery was clean (atomic rebuild, protection held,
   nothing lost), but the *window* was mine to prevent: I forked a branch
   without re-verifying master's position seconds beforehand, and I ran an
   **interactive multi-step rebase** — a paused-merge state is an open
   invitation to a daemon that resets. The atomic one-shot pattern should
   have been the first move, not the recovery move.
2. **My first zombie sweep was truncated** — `git branch -vv | head -15`
   listed 15 branches; I deleted 14 and only then discovered 4 more
   (including the checked-out one blocking its own deletion). A batch
   operation driven by a truncated list is a half-operation; the sweep had
   to run twice.
3. **PR #58 pushed before a parent check** — the standing rule (verify
   branch + commit count after `pr create`) *caught* the 2-commit fork, but
   the cheaper check (`git log --oneline origin/master..HEAD` before push)
   would have caught it pre-push and saved the force-push cycle. Rule
   applied late is a recovery loop.

Nothing was destroyed: every deletion was cherry- or diff-proofed, the css
reversion survived, and origin/master never moved.

---

## e) WHAT WE SHOULD IMPROVE

1. **Stop flagging, start executing the cheap items.** The BuildFlow cache
   purge is a documented one-liner, load-independent, and has now been
   "queued" in two consecutive reports while I handled bigger fires. Cheap
   load-independent items need a different queue than the quiet batch —
   they should execute *inside* any session that touches the repo.
2. **Under active daemon interference, every git mutation is one atomic
   bash invocation.** The paused rebase was the attack surface; the
   one-shot rebuild was immune. Make the atomic pattern the default for
   branch surgery in this repo, not the fallback.
3. **Batch operations enumerate exhaustively first** — no `head` on lists
   that drive destructive steps; `git branch -vv` in full, every time.
4. **Move the parent check pre-push**: `git log origin/master..HEAD` must
   show exactly the intended commit(s) BEFORE `git push`, not after
   `pr create`. The post-create check is the safety net, not the gate.
5. **Debris sweep checklist** (now in AGENTS.md via PR #58): branches
   (exhaustive) + `git worktree prune` + stashes (content-verified) +
   remote mirrors. A sweep that does four categories can't be "done" by a
   one-pattern grep.
6. **`git cherry master <branch>` is the standard upstream-proof tool** for
   rebase-merged branches (patch-id equivalence survives rebases that
   `git branch --merged` cannot see). It turned an 18-branch judgment call
   into a mechanical sweep with five targeted inspections.

---

## f) Up to 50 things we should get done next

Carries the 15:21 list forward; new items 1–6 from this segment, then the
standing backlog (unchanged items keep their 15:21 numbering in brackets
where useful).

| # | Item | Notes / gate |
| --- | --- | --- |
| 1 | **Merge PR #57** (15:21 report) — CLEAN, 1 commit | your OK (§g2) |
| 2 | **Merge PR #58** (AGENTS.md Git Town gotcha) — CLEAN, 1 commit; needs one update-branch cycle after #57 | your OK (§g2) |
| 3 | **BuildFlow stale pnpm-audit cache purge** — documented sqlite one-liner; flagged twice, executed zero times | load-independent; execute in ANY session |
| 4 | **Delete merged remote branches** on origin (`ci/floor-guard-rollout`, `docs/bookkeeping`, any others) | push op — your OK (§g2) |
| 5 | **§f-22 fix for real**: decide untrack+gitignore vs keep for `global.out.css`; the preserved stash executes it in one commit | decision + trivial PR |
| 6 | **Adopt the atomic-git-mutation rule** (one bash invocation for branch surgery; pre-push parent check) | process; AGENTS.md row already covers the sweep half |
| 7 | **Daemon decision** (§g1): disable / restrict / `wip:` prefix — 5 sabotages today | user answer; highest-leverage |
| 8 | **Quiet batch as one dedicated session** (§g3): T5 + T16 + T22 + T23 + T24 | load < 8 pre-flight |
| 9 | **Leg-promotion decision** after 2026-10-14 stability window + week-one review | user answer + data |
| 10 | **Per-OS test expectations** — replace 34 `skipOnWindows` skips with `filepath` assertions | prerequisite for 9 |
| 11 | **Website Build → required check** | green since #54 |
| 12 | **T17 (gated)**: two upstream BuildFlow filings (strace auto-bump evidence; result-cache key gap) | approval gate |
| 13 | **Release smoke-test script** before v2.4.3 | |
| 14 | **Re-check release-PR zero-job failure** on next release PR | |
| 15 | **Wire html-validate into Website Build** (devDep exists) | |
| 16 | **Starlight editLink/lastUpdated decision** | |
| 17 | **Website redeploy when there is something to ship** | v2.4.2 not on live changelog yet |
| 18 | **Overnight merge-ownership protocol** | |
| 19 | **Consumer sweep** after red-master windows | |
| 20 | **Audit the 292 marker-free tables** in archives | |
| 21 | **Benchmark job required-check role** post-race-fix | |
| 22 | **erraudit nolint sweep** + 4 `err113` LSP warnings on middleware_test.go | |
| 23 | **CODEOWNERS-style pointer** for docs/ | |
| 24 | **Floor-guard job on release PRs** confirmation | |
| 25 | **Event-channel lifecycle gotcha** → docs/gotchas.md | |
| 26 | **Weekly `-race -count=2`** cost/benefit | |
| 27 | **vite 8.3.3 verification** with pnpm update | |
| 28 | **24.04 pins decision 2026-11-19** | |
| 29 | **Dependabot vs floor guard** conflict check | |
| 30 | **minimumReleaseAge into pnpm-workspace.yaml** | |
| 31 | **Verify 09:47 report's `done at` hashes** | |
| 32 | **Archive 00:42–09:47 reports** (and eventually 13:09/15:21) when tails resolve | |
| 33 | **Branch-protection record → docs/ci-workflows.md** | |
| 34 | **close/reopen mergeStateStatus trick → crush-config lessons** | |
| 35 | **actionlint in pre-push verification** | |
| 36 | **examples/ README freshness pass** | |
| 37 | **ROADMAP → v3-decisions.md cross-link** | |
| 38 | **TODO_LIST Status Snapshot recount** at next sweep | |
| 39 | **`MiddlewareWriteFileLog` deprecation warning** (v3 prep) | |
| 40 | **Confirm examples compile on Windows/macOS** explicitly | |
| 41 | **`--frozen-lockfile` for local website builds** | |
| 42 | **gitleaks → required CI set?** decision | |
| 43 | **reliability_semantics_test.go placement convention** | |
| 44 | **Confirm live site got the v2.4.1 fix** at next deploy | |
| 45 | **buildflow doctor + list providers inventory** (refresh binary-version note) | load-independent |
| 46 | **Post-merge master-CI confirmation habit** for docs-only PRs | |
| 47 | **Status-dir conventions enforcement** (this report complies) | |
| 48 | **Harvest this §f into TODO_LIST.md** at next docs session | |
| 49 | **T25 close-out** once cache + evidence purges land | |
| 50 | **Next probe-leg data checkpoint** (weekly until promotion decision) | |

---

## g) Questions I cannot figure out myself

1. **Daemon fate — now 5 sabotages today** (4 in the afternoon wave, 1 more
   during this segment's recovery: a 68-file junk commit onto local master
   plus a mid-rebase abort). What runs it, what is it FOR, and can it be
   disabled or restricted to WIP branches / a `wip:` prefix? Every session
   pays the recovery tax; I cannot see its config from here.
2. **Merge + remote hygiene authorization:** merge the two CLEAN one-commit
   PRs (#57 status report, #58 AGENTS.md gotcha — sequential, strict
   up-to-date), and may I delete the merged remote branches on origin
   (`ci/floor-guard-rollout`, `docs/bookkeeping`, + any stragglers)?
   Remote deletion is a push operation, so it waits for your OK.
3. **Quiet-batch scheduling:** should T5/T16/T22/T23/T24 run as one
   dedicated low-load session (overnight cron / weekend) instead of
   opportunistic attempts? Load has not dropped below 20 all day. (The
   leg-promotion question rides the 2026-10-14 data regardless.)

---

*Prepared 2026-10-07 15:51 CEST. Master at `d8c05cc` (local == origin);
18 zombie branches + stale worktrees removed; PRs #57 and #58 OPEN and
CLEAN; the only unmerged local artifact is a labeled stash holding the
§f-22 css reversion.*
