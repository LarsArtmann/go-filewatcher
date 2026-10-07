# Status Report: Pareto Execution Session — CI hardening, race fix, docs debts

**Date:** 2026-10-07 09:47 CEST
**Scope:** This session only ("GET SHIT DONE — the whole TODO list", executing the 08:17
Master Pareto Execution Plan). Plan tasks T1–T9 attempted; T10–T27 untouched.
**End state:** 4 open PRs, all green (#38 CLEAN/MERGEABLE, #39/#40/#41 green) · master
itself RED on the 1.27 leg (superseded by #38, which removes that leg) · 1 real data
race found by CI and fixed (2 follow-up iterations) · branch protection now ON
(direct pushes declined — every change, including the daemon's, must go through PRs).

---

## a) FULLY DONE

1. **T1 (push + CI): master pushed** — `3b56ef0` (78-file docs sweep) and `ddd15e7`
   (floor guard) landed on origin. CI verdict on `ddd15e7`: 1.26 legs + lint + docs
   GREEN, 1.27 leg RED (see c.1 — that leg is removed by #38).
2. **T2 (floor guard): the war class is now structural.** `Go Directive Floor` CI job
   asserts `go 1.26.7` and gates the test matrix; `nix flake check` runs the twin
   `go-directive` check; AGENTS.md documents both. Failure path proven locally
   (bumped-copy test) and the job is green in real CI.
3. **T3 (Ubuntu 26 audit, deadline 10-12 — done 5 days early).** All 10 jobs across 5
   workflows inventoried (pinned-SHA actions, explicit Go — no Python/apt-key/cgroup
   exposure). Adaptation: test job now matrices **ubuntu-24.04 + ubuntu-26.04** —
   26.04 proven green in CI runs 37585089337/37588681273 BEFORE the 2026-10-19
   label migration; all other jobs pinned to 24.04. Re-evaluate pins at 24.04
   brownout announcement (expected 2027).
4. **T4 (release gating): master is protected** — strict required status checks (8
   contexts), enforce_admins, linear history. The v2.4.0 "merged while red" class is
   structurally dead. Side effect discovered live: **direct pushes are declined too**
   ("protected branch hook declined") — the daemon can no longer push master; every
   change needs a PR. `release.yml` gained `workflow_dispatch` with a tag input
   (GITHUB_TOKEN tags never re-trigger workflows); dispatch dry-run NOT yet done (c.4).
5. **Data race: found by CI, fixed, CI-verified.** The new Benchmark-job run on PR
   #37 caught `DATA RACE: closechan vs chansend` — the fsnotify loop closed the
   shared event channel while the poll loop was mid-send. Root fix: per-Watch
   `chUsers` WaitGroup + closer goroutine (`341b1ac`), plus a regression test
   (cancel-mid-emit, drain to closed channel, under -race).
6. **T6 (missed 03-10 report): annotated + harvested** — 30 inline evidence-citing
   verdicts; residue routed: July 6h Benchmark hang, post-release smoke-test script,
   minimumReleaseAge window, merge-ownership protocol, consumer sweep (→ TODO_LIST),
   master-health automation (→ ROADMAP), Q10 (v2.4.1 keep/retract) + Q11 (daemon
   manifest policy) (→ open questions). Open items 36 → 39.
7. **T7 (archive row gate): 86/86 uniform.** `check-rows.py` initially failed with
   **1,768 PARTIAL rows** — the 08:14 sweep's buggy wrapper had left bare `~~` in
   first cells. A cell-aware normalizer struck every cell uniformly across 8 files
   (+18 residual rows in 2 files: code-span-hidden `~~`, empty TOTAL cells), with a
   marker note per repaired file. Both gates pass: presence + row uniformity. The
   292 marker-free (data?) tables are recorded as a bounded follow-up audit.
8. **T8 (exemption zero): PR #40.** The docs-consistency gate's real gaps were exactly
   `FilterGeneratedCodeFull` + `FilterGeneratedCodeWithFilter` (the earlier "already
   in FEATURES.md" observation was a misread of filter_gogen.go). Both documented
   (content-check semantics + custom-instance wrapper); exemptions removed; gate
   verified green locally with the shrunk list (137 symbols, 0 missing).
9. **T9 (AGENTS.md split): PR #41 — 630 → 221 lines.** Nine reference blocks moved to
   docs/ (file-organization, patterns, ci-workflows, linter-cheatsheet,
   gogenfilter-v3-notes, error-handling, nix-vendorhash, phantom-types, gotchas).
   Concept survival is SCRIPT-VERIFIED, not claimed: zero original headings missing
   from the AGENTS+docs corpus, all 26 gotchas verbatim in docs/gotchas.md, all
   pointers resolve. Session-critical content stays inline (commands, conventions,
   gotcha index, Release/CI gotchas + floor, Known Issues). Duplicated
   `## Release / CI Gotchas` heading fixed.
10. **Bonus hardening (from the 03-10 report's own list):** every CI job now has
    `timeout-minutes` (5–15; the Benchmark job's only ceiling was the 6h default —
    the July PR #14 hang burned exactly that), and the dev shell pins
    `GOTOOLCHAIN=local` so a directive bump fails loudly instead of silently
    downloading a newer toolchain.
11. **Load-contamination proven, not guessed.** Local full-suite failures under
    machine load 27 were shown to be environmental via a HEAD-baseline worktree
    (the pre-fix tree failed the same random pure-unit tests). CI (idle runners) is
    the authoritative gate; the race-fix PR went fully green there.

## b) PARTIALLY DONE

1. **PR #38 is CLEAN and MERGEABLE but NOT merged** — my `gh pr merge 38 --rebase
   --delete-item=false` used a nonexistent flag; gh printed usage (truncated by
   `tail`) and merged nothing. One correct command from done (see f.1).
2. **PRs #39/#40/#41 are green but will need a rebase** after #38 lands (strict
   up-to-date requirement); none conflict with it (disjoint files).
3. **T5 (full `nix flake check`)** ran once: failed ONLY at treefmt (my new flake
   block was unformatted) — fixed via `nix fmt`; the full re-run is deferred until
   the machine is quieter (test derivations under load 27 flake out). Second
   vendorHash leg (benchstat) not yet proven this session.
4. **T1's "CI green on origin" is only true transitively**: master HEAD `ddd15e7` is
   RED on the 1.27 leg; #38 removes that leg and green-lights master on merge.
5. **TODO_LIST post-merge bookkeeping** (ticking CI items, pulling #38's hashes into
   the 03-10 annotations' "done at" for the follow-ups) — prepared but pending merge.

## c) NOT STARTED

1. **Go 1.27 matrix leg** — removed deliberately (PR #38): under Go 1.27,
   `encoding/json/v2` requires the go directive ≥ 1.27 — structurally incompatible
   with the pinned 1.26.7 floor (CI run 37581977239). Re-adding REQUIRES the
   intentional floor bump + json/v2 usage review. Documented in AGENTS.md; a v3
   decision input (T26).
2. **Plan tasks T10–T27 untouched** (status index, Windows/macOS CI, test trio,
   website cluster, BuildFlow, dependabot PRs, pkg.go.dev, runbook, docs
   verification, bench re-capture, fuzz, stress harness, hygiene, v3 doc,
   cross-repo bundle).
3. **T17 upstream BuildFlow issues** — still user-gated (Q6), not even drafted.
4. **release.yml dispatch dry-run** — trigger exists, never dispatched.
5. **Daemon under protection** — the daemon's pushes now decline; its local commits
   pile up on diverged local master (e.g. `e759746` duplicates #38 content). A
   local-reset + daemon-policy decision is needed (see g.1).

## d) TOTALLY FUCKED UP

1. **Silent merge failure.** `gh pr merge` with a bogus flag failed and I read the
   truncated output as success-adjacent. The exact failure mode the 03-10 report's
   e.1 warned about ("gh's quiet output misled me"), repeated same-day. Verify
   `gh pr view --json state` after every merge — non-negotiable now.
2. **Daemon race chaos, round two.** My first T2 commit was split by the daemon
   (ci.yml absorbed separately), my squash+amend created non-FF chaos, a push got
   rejected, and I burned several rounds reconciling — including one amend of a
   DAEMON commit (content preserved, message lost to history). No work was lost,
   but the "commit explicitly and fast" lesson now needs an upgrade: under branch
   protection, commit to a FEATURE branch fast, never fight for master.
3. **Commitlint self-block.** PR #37 (the first PR ever through the new gate) failed
   commitlint because MY commit subject was 74 chars (limit 72). The branch needed a
   clean-history rebuild (#37 closed → #38). Check subject length before pushing.
4. **Race fix needed three CI round-trips.** Fix 1 (closer goroutine) raced
   `Reset()`'s Once reassignment (not enrolled in `w.wg`); fix 2 tripped
   `waitgroupgo` (modernize → `WaitGroup.Go`). I reasoned about senders but not
   about every lifecycle actor touching the Once — Reset() was knowable upfront
   (watcher.go:744). Lifecycle changes demand enumerating ALL actors first.
5. **Two self-inflicted syntax/copy slips**, both caught immediately: an edit
   emitted `})()` (invalid), and T8 `cp`'d the whole FEATURES.md instead of
   patching (worked only because the checkout was otherwise clean).

## e) WHAT WE SHOULD IMPROVE

1. **Verify-then-act, enforced mechanically:** alias `gh pr merge` behind a wrapper
   that polls `gh pr view --json state` until MERGED, or never again merge by hand.
2. **Local commitlint gate:** run the workflow's subject checks (≤72 chars,
   `type: subject`) before every push — two of today's CI round-trips were
   self-inflicted format violations.
3. **Lifecycle checklist:** when touching channel/goroutine lifecycle, enumerate
   every actor on the shared state (loops, closer, Close, Reset, WatchOnce) BEFORE
   writing the fix; the race detector only surfaces what the test schedule hits.
4. **check-rows immediately after any bulk annotate sweep** — presence gates hide
   misplaced markers; row uniformity caught 1,768 of them today.
5. **Machine-load discipline:** under load >~15, local full-suite runs are noise
   (proven via baseline worktree); use CI as the gate and save local suites for
   quiet windows (AGENTS.md bench discipline generalizes to tests).
6. **Protection rollout needs a daemon plan on day zero:** master protection without
   a daemon workflow = daemon commit pile-up on a diverged local master. Decide the
   daemon's new role (g.1) before the divergence grows.

## f) Up to 50 things we should get done next

Impact-ordered. 1–8 are immediate mechanics; 9+ follow the 08:17 plan order.

1. Merge PR #38 (CLEAN/MERGEABLE) with `gh pr merge 38 --rebase`, then VERIFY
   `state == MERGED`; watch the master push run go green (1% item complete).
2. Rebase + merge #39 (docs sweep follow-ups), #40 (exemption zero), #41 (AGENTS
   split) in order, verifying each; watch master stay green after each.
3. Post-merge bookkeeping in TODO_LIST: tick CI-guard/required-checks/dispatch/
   26.04-matrix/check-rows/03-10 items; update the 03-10 annotations' `done at`
   hashes to the final squashed/rebased ones.
4. Reset local master to origin and clear the daemon's diverged local commits
   (`git fetch && git switch master && git reset --hard origin/master` is banned by
   policy — use `git switch` + `git reset origin/master` (mixed) after confirming
   nothing unique is local; e759746's content is already in #38).
5. T4.3: dry-run `workflow_dispatch` on release.yml (dispatch on v2.4.1, cancel
   before the release step — proves the test+lint path without touching the release).
6. T10: `docs/status/README.md` index (active reports + archived pointer) +
   CHANGELOG v2.4.0 cosmetic blank-line fix.
7. T5 re-run: full `nix flake check` on a quiet machine; verify the second
   vendorHash leg (flake.nix benchstat); record results in the 08-14 report §b6.
8. T11+T12: Windows/macOS CI matrix jobs (extend the os matrix; platform skips for
   inotify-only tests; triage+fix; document Windows semantics in Troubleshooting).
9. T13: reliability test trio (ContentHashMaxSize(0) semantics pin,
   BudgetCap==WatchLimit @ fraction 1.0, ErrorContext.Event population).
10. T14: website CI job (`pnpm install && pnpm build` + html-validate) — PR-gated
    like everything else now; T15: website dep updates (fast-uri/astro) + dedupe +
    redeploy + sharp spot-check.
11. T18: dependabot PRs #31/#32 (review + merge — they now need rebase onto the new
    master and green required checks), close stale #14.
12. T16: BuildFlow binary upgrade + findings triage (9-tools warning, go-auto-
    upgrade ×12, nix-checker ×8, nix-build-verify fate, vulnix CVEs).
13. T17 (user-gated): draft the two upstream BuildFlow issues (go-structure-linter
    auto-bump with strace evidence; result-cache ignoring pnpm-lock.yaml) — submit
    only on approval.
14. T19: pkg.go.dev crawl state for v2.4.0/v2.4.1; OG image check; editLink/
    lastUpdated; changelog.mdx gitignore decision.
15. T20: website deploy runbook (flake apps, trailingSlash/cleanUrls quirk,
    slug-verify step) — given the AGENTS split, put it in docs/ with a pointer.
16. T21: content-verify DOMAIN_LANGUAGE / API_STABILITY / Troubleshooting /
    MIGRATION / ARCHITECTURE against code; fix findings.
17. T22: benchmark re-capture (quiet machine; then README table refresh; Q2 policy
    question still gates "committed baseline").
18. T23: fuzz expansion (combinators, Event JSON round-trip, gitignore matcher,
    case-insensitive filter+middleware) + corpus run.
19. T24: large-tree stress harness (100k-dir fixture; batch/budget/self-heal
    assertions; CI gating decision).
20. T25: hygiene (gitleaks + codespell once; /tmp evidence-log cleanup now that CI
    is green).
21. T26: v3 decision doc (docs/research/v3-decisions.md) — now also records the
    go-directive-floor-to-1.27 constraint (json/v2 gate) as a first-class input.
22. T27: cross-repo bundle (crush-config lessons.md entry; buildflow skill DB path
    fix + fan-out verify).
23. Add the go-directive floor to the release-please release-PR validation path —
    release PRs touch version files only, but confirm the floor job reports on them.
24. Decide Benchmark job's role in required checks (currently NOT required — it
    caught today's race, but races are schedule-dependent; weigh flake-blocking vs
    detection value).
25. Consider `continue-on-error` for the Benchmark job on PRs if it proves flaky
    post-race-fix (its failures today were real, not noise — revisit after a week).
26. Wire the dev-shell `GOTOOLCHAIN=local` story into AGENTS.md gotcha (it landed in
    flake.nix + the CI PR, but the gotcha section's "Restore" instructions predate
    it).
27. After the 24.04/26.04 matrix bakes for a week: drop the 24.04 pins for the
    non-test jobs OR keep until brownouts (documented decision point 2026-11-19).
28. Dependabot: verify gomod PRs will pass the new required checks (they rebase
    automatically; the floor guard blocks any directive bump — watch for conflicts).
29. Add a repo-level `CODEOWNERS`-style pointer so docs/ PRs (new files from #41)
    get reviewed by the right person — even solo, it documents intent.
30. Sweep the remaining `//nolint:erraudit` unknown-linter warnings in CI logs
    (golangci-lint v2.14 doesn't know `erraudit` — the nolint directives are dead
    weight or version-mismatched; audit them).
31. Re-check the zero-job release-PR startup failures on the NEXT release PR
    (03-10 f.12; transient-race hypothesis untested).
32. Post-release smoke-test script (TODO_LIST new item) — write it before the next
    release, not after.
33. Move `minimumReleaseAge` window into pnpm-workspace.yaml (TODO_LIST new item).
34. Overnight merge-ownership protocol (TODO_LIST new item) — now MORE relevant:
    release PRs are the only remaining release path and they need green checks.
35. Consumer sweep (TODO_LIST new item) — dependents' CI vs red-master windows.
36. Audit the 292 marker-free tables in archives (TODO_LIST new item) — classify
    data vs open-task tables.
37. Investigate the July 6h Benchmark hang (TODO_LIST new item) — timeout caps bound
    the blast radius now; root cause still unknown.
38. Re-verify vite 8.3.3 lands with the T15 pnpm update (03-10 b.2 — still open).
39. Confirm the deployed website got the v2.4.1 fix (03-10 f.24) during T15's
    redeploy.
40. Record the protection configuration (contexts, strict, enforce_admins) in
    AGENTS.md's Release/CI gotchas — it's live repo state now, not just intent.
41. Consider requiring the `Go Directive Floor` context in release-please PR branch
    protection explicitly if release PRs ever target a different branch.
42. Keep a local `git fetch` cadence: with protection, origin is authoritative;
    local master is disposable (confirm no unique work before any reset).
43. Re-run `buildflow doctor` after the AGENTS split — the 220-line finding should
    clear; record the new line count in TODO_LIST when it does.
44. Update FEATURES.md row for "Generated-code detection" to cross-reference the
    two new rows from #40 (consistency of the filter family presentation).
45. Verify the 26.04 runner leg picks up the benchmark job eventually (benchmark
    still pinned to 24.04 — fine, but note the asymmetry).
46. Add the event-channel lifecycle (closer goroutine, chUsers) to AGENTS gotcha #25
    or docs/gotchas.md — the design is subtle enough to deserve a paragraph.
47. Consider a `-race -count=2` weekly CI run (race was schedule-dependent; a
    second count increases catch probability) — cost/benefit decision.
48. Sweep stale `website/src/styles/global.out.css` (+1931 daemon-committed lines)
    — likely a build artifact that shouldn't be tracked; decide gitignore vs
    keep (touches the website cluster).
49. T28 candidate: `docs/README.md` index of all docs/ reference files created by
    #41 (nine files, no index yet).
50. After everything merges: cut the CHANGELOG entry for the CI hardening + race
    fix (release-please will pick up `fix:`/`ci:` — verify it doesn't cut a
    Go release for docs-only commits; path-scoping already handles it).

## g) Questions I cannot figure out myself

1. **What is the daemon's role under branch protection?** Its master pushes are now
   declined by design; its commits pile up on a diverged local master. Options:
   (a) add the daemon's credential to the branch-protection bypass list (keeps
   direct-push convenience, weakens the gate), (b) leave it blocked and treat local
   master as scratch (my recommendation — I work on feature branches anyway), or
   (c) reconfigure the daemon to commit onto the current feature branch only. The
   daemon's config/auth is yours — I can implement any of the three once you pick.
2. **Merge the three green doc PRs (#39/#40/#41) without your review?** They're
   mechanics of the sweep you already approved (annotations, archive repair,
   AGENTS split, exemption zero), all concept-survival-verified, but #41 restructures
   a file every future session reads. Say the word and they merge; or eyeball #41
   first and I merge #39/#40 immediately.
3. **Continue T10–T27 now?** 21 plan tasks remain. Say "keep going" and I proceed
   in plan order (PR-per-task, merge as green, your gates respected); say "pause"
   and everything above is parked, verified, and resumable.

---

_Point-in-time snapshot. PRs referenced: #38 (CI hardening + race fix, CLEAN),
#39 (docs sweep follow-ups + this report), #40 (exemption zero), #41 (AGENTS split) —
all green at time of writing; master red only on the leg #38 removes._
