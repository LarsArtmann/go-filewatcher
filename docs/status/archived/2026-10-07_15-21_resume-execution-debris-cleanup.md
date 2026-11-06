# Status Report: Resume Execution — PR #56 Merged, Local Debris & /tmp Evidence Cleaned

> **ARCHIVED 2026-11-06** (docs-health sweep): every forward-looking item in
> this report is resolved inline — struck through with the evidence that closed
> it, or struck through and routed to the living docs ([TODO_LIST.md](../../../TODO_LIST.md)
> / [ROADMAP.md](../../../ROADMAP.md)) where the work still lives. The history below
> is settled; open work is in the TODO list, not here.

**Session:** 2026-10-07, 15:02–15:21 CEST (resume of the 13:09 continuation
session; executed its pending tail: the open status-report PR, local git
debris, and the never-done /tmp evidence cleanup)
**Machine:** shared box still under heavy load (load average 39–69 on 32
cores during the session; the <8 pre-flight gate held, so the quiet-machine
batch correctly stayed deferred).

---

## a) FULLY DONE

### 1. PR #56 (13:09 status report) merged through the full gated cycle

- `gh pr update-branch 56` → "already up-to-date"; **`commits | length == 1`**
  verified immediately — no daemon interference on the branch this time.
- All **16 checks green**, watched to completion in a background shell
  (`/tmp/pr56-checks.txt`; Benchmark + Regression Check + 4 OS test legs +
  floor guard + lint + docs gates all pass).
- `gh pr merge 56 --rebase` → **`state == MERGED` verified**. Master moved
  `a62cb9c` → `d8c05cc`.

### 2. Local checkout resynced

`git switch master && git pull --rebase` → `d8c05cc`, working tree clean.
Note: the 13:09 session's stray `.go-structure-linter.yaml` modification is
**gone** from the tree — see §b1/§d for why that is only half-closing.

### 3. Stash cleanup — 3 stashes dropped, each content-verified first

| Stash                                      | Claim        | Verification performed                                                                                                                                                                                                      | Verdict                            |
| ------------------------------------------ | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| `stash@{1}` (ci.yml residue, +14/−3)       | residue      | `git diff stash master -- ci.yml` → **empty**: stash-tree file byte-identical to master                                                                                                                                     | zero unique content → dropped      |
| `stash@{0}` ("post-merge residue")         | all upstream | read the actual patch: it **removed** the landed `package_json_file: website/package.json` fix from `website-build.yml` and **un-ticked** the DONE Website-CI TODO — i.e. a stale _regression_ predating merged PRs #52–#54 | superseded → dropped, nothing lost |
| `stash@{2}` ("stale 7df48ce tree content") | safe to drop | `git merge-base --is-ancestor 7df48ce master` → **true**: all content arrived via the merges                                                                                                                                | → dropped                          |

This directly avoids repeating the 13:09 report's §d-4 mistake class (a
dropped stash without verified equality): all three drops here were made on
proof, not reasoning.

### 4. Backup branch deleted after upstream-content proof

`backup/master-diverged-20261007` had **5 commits not on master**: 3 daemon
auto-commits (noise by definition) + 2 real ones. The real commits' content
was verified upstream before deletion:

- `ci: drop the go 1.27 matrix leg` → master's `ci.yml` has **no 1.27 matrix
  entry** (only the three explanatory comment lines at ci.yml:45–48).
- `ci: restore floor-guard flake check…` → `flake.nix` carries the
  go-directive check (3 matches).

Then `git branch -D` (local-only branch; `-d` would refuse due to the daemon
noise commits). Backup branches and stashes are now **zero**.

### 5. §f-10 /tmp evidence cleanup — done after being open all day

- Scope: `find /tmp -maxdepth 1 -user lars -type f (-name '*.log' -o -name
  'pr*.txt') -newermt 2026-10-07` → **120 files** (pr45–pr56 PR-watch logs —
  matching this repo's merged PRs exactly — plus buildflow/ci/probe/vendor
  evidence logs).
- Removed via **`trash`**, not `rm` (AGENTS.md rm-ban; recoverable). 0
  remaining. Shared caches (e.g. `npm-cache-buildflow/` directory) untouched.

### 6. Final sweep

0 open PRs · master == origin/master == `d8c05cc` · clean tree · v2.4.2
serving healthy on pkg.go.dev (verified earlier today, unchanged).

---

## b) PARTIALLY DONE

1. **Debris-cleanup umbrella** — stashes, backup branch, and /tmp are done,
   but the `.go-structure-linter.yaml` mystery (13:09 §d-6 / §f-41) is only
   half-resolved: the uncommitted modification **vanished** from the working
   tree (daemon reset? buildflow self-restore? a concurrent session?), and
   nobody established which. _Disappearance ≠ explanation_ — the item stays
   open as "explain the cause", the symptom is gone.
2. **BuildFlow stale pnpm-audit result-cache purge** — NOT executed this
   session, although it is **load-independent** and a documented one-liner
   (AGENTS.md: `sqlite3 ~/.cache/buildflow/cache.db "DELETE FROM result_cache
   WHERE value LIKE '%pnpm-audit%';"`). It was mis-bucketed under the
   load-gated quiet batch. (The _proving pipeline re-run_ is load-gated; the
   purge itself never was.) Honest miss — queued as next-session item 3.
3. **Harvest of today's completions into TODO_LIST** — done in this report's
   PR (tick of the /tmp item + snapshot recount), but as batch catch-up
   rather than at completion time. See §e-2.

---

## c) NOT STARTED

- **Quiet-machine batch** (gate: load < 8; session load was 39–69):
  **T5** full `nix flake check` + second vendorHash leg (flake.nix:111);
  **T16** BuildFlow triage on binary `acdb606-dirty` (re-prove 3 skips,
  re-check the AGENTS-split line-count finding); **T22** benchmark
  re-capture + README table refresh; **T23** fuzz expansion; **T24**
  100k-dir stress harness.
- **User-gated items**: the daemon's fate (§g1), flake-tolerance policy for
  leg promotion (§g2), quiet-batch scheduling (§g3), and **T17** upstream
  BuildFlow filings (go-structure-linter auto-bump strace evidence;
  result-cache ignoring pnpm-lock.yaml) — submit only on approval.
- **§f tail** (unchanged from 13:09, minus today's completed items 10 and
  42): html-validate wiring, Starlight editLink/lastUpdated, release
  smoke-test script, per-OS test expectation rewrites, Website Build →
  required check, and the rest enumerated in §f below.

---

## d) TOTALLY FUCKED UP

**Honest answer: nothing this session.** No daemon incident, no
wrong-branch commit, no merge-red workflow, no unverified destructive
operation. The three stash drops were each content-verified before the drop
(the 13:09 §d-4 mistake class was _not_ repeated), and the branch deletion
was made on upstream-content proof, not branch age.

Two flagged judgment calls, neither rising to fucked-up:

1. **Trash-batch scope**: the 120-file /tmp purge pattern-matched ownership
   (`-user lars`, today's date) rather than verifying each file's project.
   A handful (e.g. `tq-release-probe-v033*.log`, `sd-server-8901/8905.log`,
   `webui-smoke*.log`) may belong to _other_ concurrent sessions on this
   shared box. Mitigations: `trash` is recoverable, `/tmp` is ephemeral, and
   live appends survive trashing — but this was a scope judgment, not a
   verification.
2. **Carry-over open wounds** (inherited, still open): the daemon policy
   debt (4 sabotages, ~1h lost — 13:09 §d-1), the release-PR zero-job quirk
   (untested until v2.4.3), and the `.go-structure-linter.yaml` cause
   unknown.

---

## e) WHAT WE SHOULD IMPROVE

1. **Split "quiet-machine batch" into load-GATED vs load-INDEPENDENT.**
   Today's conflation left a cheap, documented one-liner (the buildflow
   cache purge) undone while correctly deferring the real heavy work. Rule
   of thumb: anything whose verification lives on GitHub's runners or in a
   local DB/CLI inventory is _not_ load-gated.
2. **Close-and-annotate atomically.** Tick the report/TODO the moment an
   item resolves (this session batch-harvested §f-10/§f-42 into one PR at
   the end). The docs-health ANNOTATE discipline is cheapest at completion
   time.
3. **Investigate state changes; don't just note them.** The
   `.go-structure-linter.yaml` modification disappearing without an owner is
   exactly the "unexpected ≠ wrong — READ it" case from AGENTS.md; "the tree
   is clean now" is an observation, not an explanation.
4. **Verify scope before batch cleanup even when the operation is
   recoverable.** Recoverability lowers the cost of being wrong, not the
   probability. One `lsof`/project-match pass would have firmed up the
   120-file scope call.
5. **Keep the pre-flight load gate mechanical.** `uptime` was checked before
   every batch decision this session and correctly held the line (39–69 vs
   the <8 gate); make it a script precondition, not a judgment call.

---

## f) Up to 50 things we should get done next

Impact-ordered. Items 10 and 42 of the 13:09 list are **done** (marked ✅
where referenced); this list re-orders what remains.

<!-- check-rows:ledger -->

| #      | Item                                                                                                                                   | Notes / gate                                                                                                                                                                |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~1~~  | ~~**Daemon decision** (§g1): disable / restrict to WIP branches / `wip:` prefix~~                                                      | ~~user answer; highest-leverage process fix~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                           |
| ~~2~~  | ~~**Quiet batch as one dedicated session** (§g3): T5 + T16 + T22 + T23 + T24~~                                                         | ~~load < 8 pre-flight~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                 |
| ~~3~~  | ~~**BuildFlow stale pnpm-audit cache purge** (documented sqlite one-liner) + proving re-run~~                                          | ~~purge is load-independent; re-run wants quiet~~ no longer open here — routed to TODO_LIST (the state DB (`SQLITE_BUSY`); four subsequent runs with the same seed were)    |
| ~~4~~  | ~~**Leg-promotion decision** (§g2) after the 2026-10-14 stability window~~                                                             | ~~user answer + week-one data~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                         |
| ~~5~~  | ~~**Per-OS test expectations** — replace the 34 `skipOnWindows` skips with `filepath`-based assertions~~                               | ~~prerequisite for 4~~ no longer open here — routed to TODO_LIST (informational. Promotion prerequisite: per-OS expectations replacing the)                                 |
| ~~6~~  | ~~**Make `Website Build` a required check** (green since #54)~~                                                                        | ~~prevents merge-red-workflow recurrence~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                              |
| ~~7~~  | ~~**T17 (gated)**: file the two upstream BuildFlow issues~~                                                                            | ~~strace + cache-key evidence ready~~ no longer open here — routed to TODO_LIST (tripwire** — cache-keying gaps (pnpm-lock, oxlint) have twice hidden real)                 |
| ~~8~~  | ~~**Release smoke-test script** — before v2.4.3, not after~~                                                                           | ~~TODO_LIST has the slot~~ no longer open here — routed to TODO_LIST (**Post-release smoke-test script** — after each release-please merge, assert)                         |
| ~~9~~  | ~~**Re-check release-PR zero-job startup failure** on the next release PR~~                                                            | ~~transient-race hypothesis untested~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                  |
| ~~10~~ | ~~**Wire `html-validate` into Website Build** (devDep exists)~~                                                                        | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~11~~ | ~~**Starlight editLink/lastUpdated decision**~~                                                                                        | ~~stale-date tradeoff~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                 |
| ~~12~~ | ~~**Website redeploy when there is something to ship**~~                                                                               | ~~v2.4.2 not yet on the live changelog page~~ no longer open here — routed to TODO_LIST (would have produced a wrong `v2.4.2...v2.5.0` compare link.)                       |
| ~~13~~ | ~~**Overnight merge-ownership protocol**~~                                                                                             | ~~release PRs are the only release path~~ no longer open here — routed to TODO_LIST (**Overnight merge-ownership protocol for release PRs** — define who/what)              |
| ~~14~~ | ~~**Consumer sweep** — no dependent CI keyed off red-master windows~~                                                                  | ~~verify even though none were red~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                    |
| ~~15~~ | ~~**Audit the 292 marker-free tables** in archived reports~~                                                                           | ~~data vs open task~~ no longer open here — routed to TODO_LIST (**Audit the remaining marker-free tables in archived files** — the)                                        |
| ~~16~~ | ~~**Benchmark job's required-check role** post-race-fix~~                                                                              | ~~consider `continue-on-error` after a week green~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                     |
| ~~17~~ | ~~**`erraudit` nolint sweep** (CI "unknown linters" dead weight) + the 4 `err113` LSP warnings on `middleware_test.go` as candidates~~ | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~18~~ | ~~**CODEOWNERS-style pointer** for docs/ PRs~~                                                                                         | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~19~~ | ~~**Confirm the floor-guard job reports on release PRs**~~                                                                             | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| 20     | **Event-channel lifecycle gotcha** (closer goroutine + `chUsers`) → docs/gotchas.md                                                    | NOT covered by gotcha #27                                                                                                                                                   |
| ~~21~~ | ~~**Weekly `-race -count=2`** cost/benefit decision~~                                                                                  | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~22~~ | ~~**`website/src/styles/global.out.css`** artifact sweep (+1931 tracked daemon lines)~~                                                | ~~gitignore or keep~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                   |
| ~~23~~ | ~~**docs/README.md index** of the docs/ reference files~~                                                                              | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~24~~ | ~~**July 6-hour Benchmark hang root cause**~~                                                                                          | ~~timeout caps bound it~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                               |
| ~~25~~ | ~~**vite 8.3.3 verification** with the T15 pnpm update~~                                                                               | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~26~~ | ~~**24.04 pins decision point 2026-11-19**~~                                                                                           | ~~brownout re-eval~~ no longer open here — routed to TODO_LIST (**Ubuntu 26 runner re-evaluation — after the 2026-11-19 rollout** — every)                                  |
| ~~27~~ | ~~**Dependabot vs floor guard** conflict check on next gomod PR~~                                                                      | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~28~~ | ~~**`minimumReleaseAge` into pnpm-workspace.yaml**~~                                                                                   | no longer open here — routed to TODO_LIST (review the 7 `minimumReleaseAgeExclude` entries once the patched versions)                                                       |
| ~~29~~ | ~~**Probe-leg week-one stability review**~~                                                                                            | ~~data collection for 4~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                               |
| ~~30~~ | ~~**Verify the 09:47 report's `done at` hashes**~~                                                                                     | verified — every `done at` hash in the 09:47 report resolves (docs-gates annotate reports no B/hash findings)                                                               |
| ~~31~~ | ~~**Archive 00:42–09:47 reports** once their tails resolve~~                                                                           | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| 32     | **Branch-protection record → docs/ci-workflows.md**                                                                                    | not just AGENTS.md                                                                                                                                                          |
| 33     | **close/reopen mergeStateStatus trick → crush-config lessons**                                                                         | discovered post-lessons-PR                                                                                                                                                  |
| ~~34~~ | ~~**`actionlint` in pre-push verification** for workflow changes~~                                                                     | ~~would have caught §d-2 of 13:09~~ no longer open here — routed to TODO_LIST (**Add `actionlint` to pre-push verification** — catch workflow-expression)                   |
| ~~35~~ | ~~**examples/ README freshness pass**~~                                                                                                | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~36~~ | ~~**ROADMAP cross-link to docs/research/v3-decisions.md**~~                                                                            | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| ~~37~~ | ~~**TODO_LIST Status Snapshot recount**~~                                                                                              | ~~drifts every sweep~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                  |
| ~~38~~ | ~~**Explain the `.go-structure-linter.yaml` disappearance**~~                                                                          | ~~symptom gone, cause unknown~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                         |
| ~~39~~ | ~~**`MiddlewareWriteFileLog` deprecation warning** (v3 prep)~~                                                                         | ~~Windows flake is one more argument~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                  |
| ~~40~~ | ~~**Confirm examples compile on Windows/macOS explicitly**~~                                                                           | ~~legs are informational~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                              |
| ~~41~~ | ~~**`--frozen-lockfile` for local website builds**~~                                                                                   | no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                                         |
| 42     | **gitleaks → required CI set?** decision (currently manual)                                                                            |                                                                                                                                                                             |
| ~~43~~ | ~~**reliability_semantics_test.go placement convention**~~                                                                             | ~~"semantics pin" tests~~ no longer open here — routed to TODO_LIST (**Settle the `reliability_semantics_test.go` placement convention** —)                                 |
| ~~44~~ | ~~**Confirm live site got the v2.4.1 fix** during next deploy~~                                                                        | ~~still assumed~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                       |
| ~~45~~ | ~~**buildflow doctor + list providers inventory**~~                                                                                    | ~~cheap, load-independent; refresh the binary-version note~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                            |
| ~~46~~ | ~~**Post-merge master-CI confirmation habit** for docs-only PRs~~                                                                      | ~~same-SHA argument is near-certain; make it explicit~~ no longer open here — routed to TODO_LIST (**Adopt a post-merge master-CI confirmation habit for docs-only PRs** —) |
| ~~47~~ | ~~**Status-dir conventions enforcement**~~                                                                                             | ~~this report complies~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                |
| ~~48~~ | ~~**Harvest this report's §f into TODO_LIST.md**~~                                                                                     | ~~partially done by this PR's tick; full HARVEST pass next docs session~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)               |
| ~~49~~ | ~~**T25 close-out**: mark hygiene fully done once cache + evidence purges land~~                                                       | ~~/tmp half ✅ today~~ no longer open here — routed to the living backlog (docs-health harvest 2026-11-06)                                                                  |
| ~~50~~ | ~~**Next probe-leg data checkpoint** (weekly cadence until the promotion decision)~~                                                   | no longer open here — routed to TODO_LIST (informational. Promotion prerequisite: per-OS expectations replacing the)                                                        |

---

## g) Questions I cannot figure out myself

(Re-asked from the 13:09 report — still unanswered, still gating items.)

1. **The auto-commit daemon cost ~1h of recovery across 4 sabotage events
   earlier today (wrong-branch commits, a closed PR, junk commits commitlint
   then rejects). What is it FOR — can it be
   disabled, or restricted to dedicated WIP branches?** I can't see what
   runs it or whether other sessions depend on its snapshots. If it stays:
   may I add a branch-name filter or `wip:` prefix so its commits are
   skippable by policy?
2. **How flake-tolerant should required checks be?** After the probe legs
   bake a week (2026-10-14), promoting windows/macos to required blocks
   merges on platform flakes (the `TestMiddlewareWriteFileLog` Windows
   failure was exactly such a flake — fixed, more will surface). Strict
   gating (fix every platform flake immediately) or informational legs with
   a weekly review?
3. **Should the quiet-machine batch (T5/T16/T22/T23/T24) be scheduled as one
   dedicated low-load session (overnight cron or weekend run) instead of
   attempted opportunistically?** Load has been 39–1050 all day on 32 cores;
   if overnight quietness is predictable, I'd script the batch as a single
   sequence with a mechanical pre-flight load check.

---

_Prepared 2026-10-07 15:21 CEST. Master at `d8c05cc`; PR #56 merged; 0 open
PRs before this report's PR; stashes/backup-branch/tmp-evidence: zero._
