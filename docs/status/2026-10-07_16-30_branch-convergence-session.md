# Branch Convergence — "MERGE all branches into master" Session Report

| | |
| --- | --- |
| **Date** | 2026-10-07 16:30 CEST |
| **Session scope** | Integrate all 5 unmerged branches into master under live branch protection; recover from two conflict-paused `git sync` runs and one daemon reset |
| **Start state** | master `d8c05cc`; local branches `docs/git-town-branch-hygiene`, `docs/status-1521-report`, `docs/status-1551-report` (each +1 commit); diverged remote branches `origin/docs/bookkeeping`, `origin/ci/floor-guard-rollout` (no PRs); local master later carried 14 commits (5 of them merge commits) that could never be pushed |
| **End state** | master `f88be1a` == origin/master, clean; only `master` exists locally and remotely; zero open PRs; all content upstream; required CI green on every merged PR (incl. windows/macos probe legs); tree byte-identical to the locally `nix flake check`-validated tree (9/9 checks) |

---

## a) FULLY DONE

1. **PR #58 merged** (`docs/git-town-branch-hygiene`, rebase merge) — AGENTS.md git-town
   branch-debris gotcha row landed. Evidence: `260664e` on origin/master.
2. **PR #57 merged** (`docs/status-1521-report`, update-branch + rebase merge) — 15:21
   resume-session report, TODO_LIST ticks, status index rows. Evidence: `64b25f6`;
   all 8 required checks green on the updated head.
3. **PR #59 merged** (`docs/status-1551-report`, squash merge) — 15:51 git-town recovery
   report. The docs/status/README.md conflict (both sides prepending index rows) was
   resolved on the branch (rows ordered newest-first: 15-51 → 15-21 → 13-09), branch
   pushed (fast-forward `137660c..0151baa`), then squashed. Evidence: `f88be1a`.
4. **`origin/docs/bookkeeping` deleted after airtight zero-contribution proof** —
   `git merge-tree --write-tree f88be1a origin/docs/bookkeeping` produced a tree
   with an EMPTY diff vs master; every file it touched was already upstream
   (verified identical at `d8c05cc` before the doc merges, and net-zero after).
5. **`origin/ci/floor-guard-rollout` deleted after the same proof** — its ci.yml /
   release.yml / AGENTS.md conflicts all resolved to master's strictly newer
   versions (master had the probe legs, `workflow_dispatch` release path with the
   v2.4.2 dry-run note, ubuntu-24.04 pins, `shell: bash` + awk coverage gate);
   net merge contribution: zero files changed.
6. **Final tree validated locally** — `nix flake check` on the merged local state
   (`369e0ac`): all 9 checks passed (go-directive floor `1.26.7`, build, test,
   lint, vet, fmt, treefmt, examples-build, go-modules). Proved remote master's
   tree byte-identical afterwards: `git diff f88be1a 369e0ac` empty.
7. **Local branch cleanup with upstream proofs** — `git cherry master <branch>`
   showed `-` (patch-equivalent) for git-town-branch-hygiene and
   status-1521-report; status-1551-report's tree-diff vs master tip is empty
   (the single `+` is the squash-format patch-id mismatch). All three deleted.
8. **Two conflict-paused `git sync` runs recovered** — first via `git town undo`,
   second via `git rebase --abort` (already unwound) + `git switch -C master
   origin/master` realignment, after the rebase-replay of local merge commits
   was identified as the wrong operation under PR-based integration.
9. **One daemon mid-merge reset survived without data loss** — the `reset: moving
   to HEAD` during the first bookkeeping merge was diagnosed from the reflog;
   the merge was redone in a single chained command.

## b) PARTIALLY DONE

1. **New git-town hazard not yet recorded in AGENTS.md** — the lesson "under
   branch protection, realign local master to origin/master before `git sync`;
   rebasing unpushed merge commits conflict-pauses" is identified and applied
   but not written down. The existing "Git Town syncs ALL local branches" row
   covers a sibling case, not this variant. Blocker: any AGENTS.md change now
   needs its own PR. Effort: S.
2. **This report not yet committed/PR'd** — written per the harness no-commit
   rule; the auto-commit daemon will pick it up locally, but the index row +
   report reaching origin needs the usual PR path (same for every future
   report). Effort: S.
3. **TODO_LIST.md not yet harvested from this report** — section (f) below is
   the input for a docs-health HARVEST pass; TODO_LIST still shows the
   pre-session state. Effort: M.

## c) NOT STARTED (observed during this session; no new research)

1. **stash@{0} resolution** — `global.out.css artifact reversion` from deleted
   `chore/codespell-hygiene`; dangling since before this session. Needs a user
   decision (see section g).
2. **Auto-commit daemon policy under branch protection** — it committed junk
   snapshots on branches and reset local master mid-merge this session; no
   decision made on re-scoping/disabling it (see section g).
3. **Merge-convention documentation** — #57/#58 used rebase merges, #59 needed
   squash (rebase would replay the conflicting original patch); the convention
   is undocumented.
4. **Open items visible in merged reports/TODO_LIST** (not touched this
   session): buildflow pnpm-audit result-cache purge (queued per 15:21 report),
   remote mirrors + website css decision (open per 15:51 report), §g questions,
   §b6 vendorHash re-verify, f12/T16 quiet-machine batch, website dependabot
   alerts (fast-uri, astro), html-validate in website CI, website link checker,
   probe-leg promotion decision, per-OS test expectations, go 1.27 floor bump
   (v3 decision, deliberately deferred).

## d) TOTALLY FUCKED UP

1. **Strategy inversion: local merges first, PR flow second — under live branch
   protection.** AGENTS.md says "Direct pushes to master are declined —
   everything goes through PRs", and I read that exact row mid-session. The
   entire local-merge phase (merge commits `b2de4f0`, `12f10a5`, `b600e11`,
   `369e0ac` + fix commit `683245b`) could never land; the durable path was
   PRs #57–#59 plus deletion proofs for the two PR-less branches. Severity:
   wasted ~30 min and manufactured the exact local-master divergence that then
   conflict-paused the user's `git sync` twice. Mitigation that rescued value:
   the local merges doubled as conflict-resolution rehearsal and enabled the
   `nix flake check` validation + zero-contribution deletion proofs. Root
   cause: I optimized for "get branches integrated" locally instead of
   checking the merge POLICY first.
2. **Committed a known-bad auto-merge result.** The bookkeeping merge took the
   branch's stale "Open items 40" and duplicated the count note; I noticed only
   after committing and needed fix commit `683245b`. Should have diffed the
   merge result tree against pre-merge master before `git commit`. Severity: M
   (self-inflicted, caught, fixed).
3. **Stale merge-tree preview trusted.** Previewed `origin/docs/bookkeeping`
   against pre-merge master ("clean"), then merged the three docs branches
   which changed AGENTS.md and docs/status/README.md — the inputs changed and
   the real merge conflicted exactly where the stale preview said clean.
4. **Daemon reset blindside.** The auto-commit daemon's `reset: moving to HEAD`
   wiped the first bookkeeping merge state mid-resolution. This behavior is
   documented in AGENTS.md ("the daemon commits and occasionally resets local
   master while you work") and I still left gaps between merge → inspect →
   edit → commit until the second attempt.
5. **Handoff gap: did not warn about `git sync`.** After the local merges I
   reported "ahead 14, not pushed" without flagging that the user's
   muscle-memory `git sync` would conflict-pause on the unpushed merge
   commits. The user hit the pause twice and had to paste terminal output back
   to me. The second pause was avoidable entirely had master been realigned at
   the end of the local phase.
6. **Minor: stale conversation-start state.** The session snapshot claimed
   branch `ci/windows-macos-probe`; reality was `docs/status-1551-report` on a
   different tip (18-branch git-town debris sweep had happened between).
   Caught within two commands; noted as a reminder to always re-verify.

Honesty checks: no lies told — but the earlier "waiting for your go-ahead to
push" framing was incomplete (see #5). Split brain: local master (merge-based)
vs origin/master (rebase-based) was a genuine split brain for ~30 minutes,
resolved by realignment. Ghost systems: none introduced; the two PR-less
remote branches were the ghost risk and were proven-empty then deleted.

## e) WHAT WE SHOULD IMPROVE

1. **Check merge POLICY before merge mechanics.** Under branch protection, the
   first question is "how does content land on master here?" — PRs — and only
   then "what conflicts will each PR hit?" Local merges become a rehearsal
   tool, not the deliverable.
2. **Re-run `git merge-tree` immediately before every merge.** Inputs move;
   a preview older than the last mutation is a lie.
3. **Chain state-mutating git commands when a reset-happy daemon shares the
   repo.** merge + restore + commit in ONE invocation was the fix; make it the
   default posture on this machine.
4. **Diff the merge result against pre-merge master before committing.** Catches
   silently-bad auto-merges (stale counters, duplicated blocks) at zero cost.
5. **Every handoff message must cross-check the user's next likely command.**
   "master is ahead, unpushed" must come with "do NOT `git sync` until X" when
   the state makes their habitual command dangerous.
6. **docs/status/README.md prepend-row convention guarantees conflicts** —
   three conflicts on the same 5-line table today. Bottom-append convention,
   or a tiny `new-status-report.sh` that inserts the row, removes the friction.
7. **Rehearsal-merge pattern worth keeping** — resolving conflicts locally,
   validating the tree, then translating the resolutions onto the PR flow was
   fast and safe. Keep, but never commit rehearsal merges to master.

## f) Up to 50 things to get done next (ranked by impact; session-observed only)

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Decide auto-commit daemon policy under branch protection (re-scope to branches / disable on master / remove) — it reset master mid-merge twice today and commits junk snapshots | High | S | Cleanup |
| 2 | PR: AGENTS.md gotcha — "realign local master to origin/master before `git sync` when local master carries unpushed commits" | High | S | Documentation |
| 3 | Run docs-health HARVEST: pull this report's section (f) into TODO_LIST.md / ROADMAP.md | High | M | Documentation |
| 4 | Resolve `stash@{0}` (codespell-hygiene css reversion): drop or land — user decision (see g1) | Medium | S | Cleanup |
| 5 | Verify `nix flake check` on `f88be1a` itself (tree proven identical to validated `369e0ac`; cheap certainty) | Medium | S | Quality |
| 6 | Re-run `git town sync` once to confirm the converged state balances clean | Medium | S | Quality |
| 7 | Investigate whether the daemon still creates local master commits post-protection (silent divergence would recur) | High | S | Bug |
| 8 | Document merge convention: rebase-merge default, squash for conflict-carrying PRs (what actually happened with #57/#58 vs #59) | Medium | S | Documentation |
| 9 | PR: add `new-status-report.sh` (or bottom-append convention) to kill docs/status/README.md index-row conflicts | Medium | S | Feature |
| 10 | PR: this report + docs/status/README.md index row | Medium | S | Documentation |
| 11 | Website: fix dependabot alerts `fast-uri` (high) + `astro` (medium) in website/package.json | High | M | Bug |
| 12 | Wire html-validate into website CI (devDep already present) | Medium | M | Feature |
| 13 | buildflow pnpm-audit result-cache purge (queued per 15:21 report; known surgical sqlite delete) | Medium | S | Cleanup |
| 14 | Website link checker (Astro dead-link plugin) — would have caught the migration-guide 404 | Medium | M | Feature |
| 15 | Probe-leg promotion decision: make macos-15/windows-2025 test legs required or keep informational | High | S | Decision |
| 16 | Per-OS test expectations doc (which POSIX-shape tests skip on Windows, via skipOnWindows) | Medium | M | Documentation |
| 17 | Work through follow-up items from the probe-leg wave (TODO count note) | Medium | M | Quality |
| 18 | Clean `/tmp` evidence logs post-green-CI (`bf-*.log`, `gomod-*.log`, `strace-*.log`, `gfw-main-run*.log`) | Low | S | Cleanup |
| 19 | Remote mirror setup for the repo (open per 15:51 report) | Medium | M | Feature |
| 20 | Website CSS decision (open per 15:51 report) | Low | S | Decision |
| 21 | §g questions from the 15:21 report — re-ask and close | Medium | S | Decision |
| 22 | §b6 vendorHash re-verify (open per 08-14 archive row) | Low | S | Quality |
| 23 | f12/T16 quiet-machine batch (queued per 09-47 row) | Medium | M | Quality |
| 24 | Go 1.27 floor bump decision: revisit encoding/json/v2 usage or accept the bump (v3 decision, deliberately deferred) | Medium | L | Decision |
| 25 | Archive 15-21 + 15-51 status reports once their open items resolve (docs-health ANNOTATE flow) | Low | S | Documentation |
| 26 | Fix docs/status/README.md hardcoded "Active (2026-10-07)" heading — needs a rolling-date scheme | Low | S | Documentation |
| 27 | Audit other clones/machines for stale local masters carrying merge commits (same sync-conflict trap as this session) | Medium | S | Cleanup |
| 28 | Consider a pre-sync guard: warn when local master has commits not on origin (prevents today's pause pattern mechanically) | Medium | M | Quality |
| 29 | Run `buildflow doctor` + `list providers` (machine-drift check; fleet binary changed mid-session yesterday per AGENTS.md) | Medium | S | Quality |
| 30 | Confirm release-please is unaffected by the branch cleanup (watch the next commit wave for a correct release PR) | Medium | S | Quality |
| 31 | GitGuardian third-party check sits IN_PROGRESS on PRs (non-required, slow) — suppress the noise or accept | Low | S | Quality |
| 32 | Update TODO_LIST.md status snapshot + open-item count after harvesting this report | Medium | S | Documentation |
| 33 | website-build.yml job is non-required on PRs (noted in the 15:21 DONE text) — require it or record the accepted risk | Medium | S | Decision |
| 34 | Confirm GitHub's auto-delete-head-branch setting is intentional (it silently removed all four merged PR branches today — desired, but undocumented) | Low | S | Documentation |

(34 items — stopped where specificity would have become filler. Items 11–24
were observed in merged reports/TODO_LIST during this session, not re-researched.)

## g) Questions I cannot answer myself

1. **`stash@{0}`** (`global.out.css artifact reversion`, from deleted
   `chore/codespell-hygiene`): I can diff its content, but not know whether the
   css reversion is still wanted upstream. Drop it, or resurrect onto a branch?
2. **What is the auto-commit daemon FOR now?** With branch protection live it
   cannot push, its master commits are junk snapshots, and it reset master
   mid-merge twice today. Keep it scoped to feature branches, disable it on
   master, or retire it? This is your infrastructure call — it changes how I
   guard every future git operation on this machine.
3. **Merge convention going forward:** AGENTS.md documents the rebase-merge
   loop (`gh pr merge N --rebase`), but #59 only merged via squash because the
   original commit's patch conflicts after resolution. Standardize on squash
   for everything, or keep rebase-as-default with squash as the documented
   exception?

---

*Report convention: point-in-time snapshot. Section (f) is the HARVEST input
for TODO_LIST.md/ROADMAP.md. Waiting for instructions.*
