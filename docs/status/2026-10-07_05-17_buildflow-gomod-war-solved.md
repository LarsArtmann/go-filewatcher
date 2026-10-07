# Status: buildflow go.mod War SOLVED (root cause + proof) — 2026-10-07 05:17

> **SUPERSEDED (2026-10-07 06:00):** closed out by `2026-10-07_06-00_buildflow-green-pnpm-audit-zero-cache-war.md` — findings gate cleared honestly (pnpm-audit 12→0), pipeline exit 0 twice, master pushed. §b/c residues below are resolved there except where marked.

**Session window:** 04:37–05:17 CEST (resumed from `2026-10-07_04-34_buildflow-recovery-session.md`)
**Mission:** make `buildflow --fix --build-mode=full --log-level warn --max-time 5m` pass in go-filewatcher.

---

## Headline

**The go.mod `1.26.7 → 1.27` flip war is SOLVED with hard forensic evidence, and a full pipeline run has now passed with ZERO failed steps and go.mod holding `go 1.26.7` end-to-end** (proof run, worktree commit `6e10a39`, 05:12–05:13). Remaining exit-69 is purely the findings gate on **website-level** pnpm-audit (9) + type-check (4) errors — a different work class, untouched by this session.

Two bumpers were identified and neutralized:

1. `go-mod-update` (known from previous session) — now skipped.
2. **`go-structure-linter` — NEW, discovered this session**: buildflow-internal code in the **freshly rebuilt binary** (202b114; a concurrent session rebuilt buildflow mid-session). It reports "go.mod specifies Go 1.26, but Go 1.27 is available" as an **error finding and auto-repairs it by bumping the go directive mid-run**, leaving a backup in `/tmp/go-structure-linter-backups/`.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                        | Evidence                                                                                                                                                                                                                                                                                                                                              |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **Attribution of the 1.27 bumper by strace-PID forensics**                                                                                                  | `strace -f -e trace=execve,renameat,openat` over two full runs: 4 distinct writer TIDs (2058918, 2060731, 2061560, 2062368) had **zero execve lines** → all are buildflow-internal threads, not subprocesses. TID 2062368 wrote `/tmp/go-structure-linter-backups/go.mod.20261007_050836_000.bak` then rewrote go.mod (05:08:36; reproduced 05:09:44) |
| 2  | **`go-fix` proven innocent**                                                                                                                                | Isolated `buildflow -s go-fix --fix` on go.mod=1.26.7: no bump (88 ms, no-op)                                                                                                                                                                                                                                                                         |
| 3  | **`go-mod-normalize` proven net-zero**                                                                                                                      | Source read (`BuildFlow/tools/gomod/module.go:300-400`): it writes a candidate downgrade (1.26.7→1.26, MOVED_TO #1), runs the dep-floor gate (`go mod tidy -diff`, GOWORK=off), gate rejects, restores original (MOVED_TO #2). Matches the observed double-rename pattern + "kept: downgrade is not dependency-floor-safe" WARN                       |
| 4  | **`.buildflow.yml` final policy (3 skips) with verified rationale**                                                                                         | skip: `go-version-auto-configure`, `go-mod-update`, `go-structure-linter`; comments cite strace verification. On main (working tree) + worktree                                                                                                                                                                                                       |
| 5  | **PROOF RUN: pipeline mechanically green, floor holds**                                                                                                     | Worktree `6e10a39` (floor COMMITTED — required, see #6): `BuildFlow passed with warnings 58/70, 0 failed steps`, nix-build GREEN (go-modules FOD built), **go.mod = `go 1.26.7` after the run**; only the normalize candidate+restore dance appeared in inotify                                                                                       |
| 6  | **New gotcha discovered: nix ignores dirty state in linked worktrees**                                                                                      | Worktree run at 05:00 evaluated the **committed** (poisoned) go.mod — drv name had no `-dirty`, unpacked the same source store path (`nvszy1dh…`) as main-repo runs. In worktrees the floor must be **committed** for nix to see it                                                                                                                   |
| 7  | **Attribution instrument kit built & proven**                                                                                                               | inotifywait with `--timefmt` on file+dir (CLOSE_WRITE = in-place write; MOVED_TO = atomic rename), 3 s process pollers, strace execve mapping; recovered the 04:55 flip that pure log analysis couldn't                                                                                                                                               |
| 8  | **Stale worktree git lock removed** (left by a killed strace run) and worktree recreated fresh at HEAD (`445c275`), then floor+config committed (`6e10a39`) |                                                                                                                                                                                                                                                                                                                                                       |
| 9  | **origin/master verified clean** (`go 1.26.7`): all poison is local-only; nothing pushed                                                                    |                                                                                                                                                                                                                                                                                                                                                       |
| 10 | Main-repo `.buildflow.yml` updated to the 3-skip policy (working tree, daemon will commit)                                                                  |                                                                                                                                                                                                                                                                                                                                                       |

## b) PARTIALLY DONE

| # | Item                                   | State                                                                                                                                                                                  |
| - | -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Main repo go.mod restore**           | Main go.mod = `go 1.27` at HEAD `445c275` (daemon-committed flip from 04:55). Needs one `sed` + let the daemon commit                                                                  |
| 2 | **Exit 0**                             | 0 failed steps achieved; exit 69 now comes ONLY from the findings gate: 13 findings ≥ error = pnpm-audit (9, website JS) + type-check (4, website TS). Not yet triaged/fixed/baselined |
| 3 | **Cold-cache full validation on main** | Proof run was cache-warm (30.7 s, 93% hits). A cold ~4–5 min run on main after sync is still owed                                                                                      |
| 4 | **AGENTS.md updates**                  | Needed: three-bumper mechanism in "go.mod language version"; worktree-nix-dirty-tracking gotcha; go-structure-linter skip rationale. Not yet written                                   |
| 5 | **Worktree cleanup**                   | `/tmp/gfw-verify` + branch `verify-gomod-floor` still exist (kept until main-repo sync + cold run are done)                                                                            |

## c) NOT STARTED

1. pnpm-audit remediation in `website/` (9 vulnerable deps).
2. Website type-check 4 errors.
3. AGENTS.md size trim (625 → ≤220; pre-existing, deliberately deferred).
4. Upstream BuildFlow issue: go-structure-linter auto-bumping go directives is a fleet-wide hazard.
5. Sweep other fleet repos for the same new-tool bump (mr-sync, bank-sync, cqrs-lite…).
6. Push master (13+ commits ahead) — awaiting instruction.
7. Background monitor shells: killed this session (071/07B/07C); no leftovers.

## d) TOTALLY FUCKED UP (honest list)

1. **Ignored the binary-freshness warning for ~40 minutes.** Run 1 (04:42) literally warned "binary was built at 202b114 but HEAD is 754fcc8 … results may not reflect current code". A concurrent session rebuilt buildflow and introduced go-structure-linter DURING my session — the skill's triage table says "Tool behaves like an old version → buildflow doctor" FIRST. Checking `buildflow doctor` / provider list at 04:43 would have named the new step immediately and saved three full pipeline runs (04:42, 04:55, 05:00 ≈ 12 min machine time + ~40 min wall time).
2. **The 05:00 worktree run was invalid and I didn't know it**: nix evaluated the committed poisoned go.mod, not my sed'ed working tree (linked-worktree dirty-tracking blind spot). I burned a 4-minute run "proving" nothing. Should have inspected the drv name / source store path BEFORE launching.
3. **The 04:55 main-repo rerun was launched into a known storm**: at 04:55:19 there were THREE concurrent `buildflow --fix --build-mode=full` processes (mine + two foreign `--budget 5m` runs) plus active nix builds/deploys from other sessions. Attribution contamination was predictable; I ran anyway.
4. **Instrumentation wasn't validated before dependence**: first process poller lost its redirect (output only in job buffer — nearly lost the decisive 04:55 evidence); I initially declared it "dead".
5. **Kill hygiene**: killing the strace run left `index.lock` in the worktree git dir, blocking the next commit (found and removed, but sloppy).
6. **Hypothesis-driven rerunning over evidence**: I tested suspects one by one (go-fix isolated, source-diving normalize/tidy/fix-gomod-stale/require-floor) before switching to direct PID attribution at ~05:07. strace at minute 10 ends the war; every extra rerun re-poisoned go.mod and re-muddied state.

## e) WHAT WE SHOULD IMPROVE (principles extracted)

1. **Binary drift = re-baseline everything.** On a shared machine with ~9 concurrent agent sessions, a fleet binary can gain NEW tools mid-session. Any behavior change → `buildflow doctor` + `buildflow list providers` diff BEFORE any rerun.
2. **Attribute, then configure.** PID-level forensics (strace) beats skip-and-rerun cycles; each blind rerun mutates the system you're debugging.
3. **New names in output are suspects, not noise.** go-structure-linter appeared in run 1's findings list and I treated it as background while hunting "the real" bumper.
4. **Nix + worktrees: commit or it didn't happen.** Dirty state is invisible to flake evaluation in linked worktrees.
5. **Validate instruments before trusting them** (dry-run the watcher/poller once).
6. **Detect-only vs repair-capable distinction matters**: a tool that "reports" a finding may also _fix_ it; check the repair list ("Repair ran (not measured)") and the writes, not just the findings.
7. **Concurrency inventory before rerunning**: `ps` for sibling buildflows/nix builds should be a pre-run checklist item on this machine.

## f) UP TO 50 THINGS TO DO NEXT (prioritized, this-session-derived)

**Immediate sync (main repo)**

1. Restore main go.mod → `go 1.26.7` (daemon commits).
2. Commit main `.buildflow.yml` 3-skip policy.
3. Cold full run on main → expect "0 failed steps", exit 69 only via findings gate.
4. `nix build .` to refresh `./result` and confirm vendorHash `9qQXBs…` still valid at the floor.
5. Remove worktree `git worktree remove --force /tmp/gfw-verify`; delete branch `verify-gomod-floor`.
6. Verify no background monitors/buildflow stragglers remain (`ps aux | grep -E 'buildflow|inotifywait'`).

**AGENTS.md / docs**
7. Rewrite "go.mod language version" gotcha: three bumpers (go-version-auto-configure, go-mod-update, go-structure-linter), skip policy, strace verification date.
8. Add gotcha: "nix flake eval ignores dirty state in linked worktrees — commit experiment state".
9. Add gotcha: go-structure-linter (202b114+) auto-bumps go directive; backup dir `/tmp/go-structure-linter-backups/`.
10. Record the attribution recipe (inotify --timefmt + strace execve mapping) in crush-config `references/lessons.md` (cross-project lesson, via commit).
11. Update/annotate `docs/status/2026-10-07_04-34_buildflow-recovery-session.md` with a pointer to this report (its item #1 "instrument inotify" is now DONE and answered).
12. Reconcile TODO_LIST.md with the new work items below.

**Findings gate (the last mile to exit 0)**
13. `buildflow -s pnpm-audit --format finding` — enumerate the 9 errors (js-yaml, fast-uri, source-map-js, http-cache-semantics + 5 more).
14. Upgrade `js-yaml` ≥ 4.3.2 (website/).
15. Upgrade `fast-uri` ≥ 3.1.6.
16. Upgrade `source-map-js` ≥ 1.2.2.
17. Upgrade `http-cache-semantics` ≥ 4.2.1.
18. Fix remaining 5 pnpm-audit errors.
19. `buildflow -s type-check --format finding` — fix the 4 website TS errors.
20. Decide gate policy: fix-all vs `--fail-on` config vs findings baseline for website/.
21. Re-run full pipeline → target **exit 0**.
22. Review dependabot-auto-configure info finding (nix ecosystem entry "matches nothing") — keep or drop the dependabot entry.
23. Triage go-auto-upgrade's 12 findings (jsonv1tov2 skips are deliberate; lo.Map suggestions need a decision: adopt samber/lo or suppress).
24. nix-checker "extract vendorHash to file" suggestion contradicts AGENTS.md (nix-hash-fix bug note) → document as deliberate non-fix (known-tool-bug pattern).
25. art-dupl 97 + branching-flow 14 findings → schedule a deduplicate-code session (detect-only today).
26. nix-flake-check: add `meta.description` to flake apps (16 warnings).
27. flake-meta-checker: add `platforms` attribute or suppress.
28. vulnix 20 findings (binutils/bison/coreutils CVEs) → nixpkgs bump or document accepted debt.
29. Fix "binary X not in project devShell" warnings (dprint, prettier, tsc, tailwindcss, lychee, vulnix, govulncheck) → add to devShells.default or accept.
30. Identify the "9 tools unavailable (health check failed)" via `buildflow -v` and fix or document.

**BuildFlow upstream (this machine's repo is actively developed)**
31. File issue: go-structure-linter must not auto-repair the go directive (or must respect GOTOOLCHAIN=local floors / repo policy); fleet-wide breakage vector.
32. Suggest rule-level skip granularity (`.buildflow.yml` per-rule suppression) instead of whole-step skips.
33. Consider making nix-hash-fix aware of this flake (still fails 15+/15 with "stale hash not found verbatim" — documented AGENTS.md).
34. Rebuild + reinstall fleet buildflow binary (nix build . && nix run .#reinstall in ~/projects/BuildFlow) — currently stale vs HEAD.
35. After rebuild, verify the three skip_steps names still resolve (provider renames would silently unskip).

**Repo hygiene / CI**
36. Add CI guard asserting `grep -q '^go 1.26.7$' go.mod` (fail fast on directive poison).
37. Sweep sibling repos' go.mod histories for the same 1.27 flips (mr-sync, bank-sync, cqrs-lite…).
38. Push master once user approves; verify open PR CIs go green.
39. Confirm daemon didn't commit partial `.buildflow.yml` states (`git log -p -- .buildflow.yml`).
40. Investigate the two non-linter writer TIDs (2058918/2060731) if ever relevant — proof run shows they're net-zero today.
41. Coordinate with the concurrent website session (observed `nix run .#deploy`, `.#mr-sync`, `.#pre-deploy-check` at 04:55) before pushing website-affecting changes.
42. Verify `website/src/styles/global.out.css` (+1931/−1931 churn from concurrent sessions) didn't break the website build.
43. Performance budget: full runs exceed the 1 m default budget → set an explicit `--budget` in the standard command or document the ignore.
44. Evaluate exporting `GOTOOLCHAIN=local` in `.envrc` (buildflow's fix-hint referenced shell pinning mismatches).
45. After AGENTS.md trim: consider re-enabling go-structure-linter WITHOUT its go-version repair (pending upstream fix) to keep the structure checks.
46. bench-diff discipline: schedule `nix run .#bench-diff` only with an idle machine (AGENTS.md rule; today's machine was saturated).
47. Check whether any open PRs touch the files this session changed (conflict risk before push).
48. Add the flip-flop preflight check's advice to docs: after any go.mod incident, run `git log -p -G'^go ' -- go.mod` (it's the fastest history X-ray).
49. Keep `/tmp/proof-run.log`, `/tmp/strace-bf.log`, `/tmp/strace-bf2.log`, `/tmp/gomod-*.log` until the next session confirms main-repo green (they are the evidence chain).
50. Write the close-out: after main-repo exit 0, append a "RESOLVED" note linking this report from README-adjacent changelog or the 04:34 report.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **BuildFlow upstream intent:** go-structure-linter's go-directive auto-bump — should I file an issue/PR against BuildFlow (you own it; a dev session is literally active right now), or is the bump intentional behavior I should accommodate differently (e.g., rule-level suppression API instead of skipping the whole step)?
2. **Findings-gate policy for exit 0:** fix the 9 pnpm-audit + 4 type-check website findings for real, or relax/baseline the gate for `website/` (it's a separate Astro subproject inside the same pipeline)?
3. **Push timing:** master is 13+ commits ahead including the restored `go 1.26.7` floor; open-PR CIs validate against origin. Push now (with the floor) or hold until the findings gate reaches exit 0 locally?

---

**Evidence chain (files, still on disk):** `/tmp/proof-run.log` (green run) · `/tmp/strace-bf.log` + `/tmp/strace-bf2.log` (writer attribution) · `/tmp/gomod-watch.log` + `/tmp/gomod-worktree-watch.log` (flip timelines) · `docs/status/2026-10-07_04-34_buildflow-recovery-session.md` (previous session).
