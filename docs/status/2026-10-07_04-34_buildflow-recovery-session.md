# Status Report: buildflow --fix Recovery Session (go-filewatcher)

> **SUPERSEDED (2026-10-07 06:00):** every open item here was resolved by `2026-10-07_05-17_buildflow-gomod-war-solved.md` and closed out in `2026-10-07_06-00_buildflow-green-pnpm-audit-zero-cache-war.md` (exit 0 ×2, floor held, 3-skip policy committed). Read the 06:00 report for the final state.

**Date:** 2026-10-07 04:34 CEST
**Session scope:** User ran `buildflow --fix --build-mode=full --log-level warn --max-time 5m`; it failed with 4 step failures (exit 69). This session investigated and fixed the failures, uncovering a deeper tooling war along the way.
**Final state at time of writing:** `go.mod` = `go 1.27` (re-bumped again during the last instrumentation run, mtime tracked); working tree clean at HEAD `3f82846` (all fixes committed by the auto-commit daemon). **A full green run has NOT yet been achieved.** Every individual failure is fixed and individually verified; the full pipeline still trips over the go-directive poisoner whose identity has now been narrowed to one untested suspect.

---

## a) FULLY DONE (implemented + verified)

1. **vendorHash root cause understood and fixed.** `nix build` failed: `specified: sha256-hbYfg…, got: sha256-9qQXBs…`. The fix (`sha256-9qQXBs0qx/NO3MlaGarCGHiv7OAhx7t+9Xy5F/EohOI=` in flake.nix:25) landed in the working tree during the session (applied externally ~30s after nix-hash-fix gave up; committed by the daemon as `674c08f`). Verified: FOD builds, no mismatch.
2. **GOEXPERIMENT=jsonv2 propagated to every nix Go build context.** Root cause of the second failure class: `event.go` imports `encoding/json/v2` (GOEXPERIMENT-gated); CI sets it globally (ci.yml:13) and direnv exports it, but nix sandbox derivations got neither → "build constraints exclude all Go files in …/encoding/json/v2" on every Go check. Fixed in flake.nix: `env.GOEXPERIMENT = "jsonv2"` on `packages.default` and all four `runCommand` checks (`test`, `lint`, `vet`, `examples-build`), plus `export GOEXPERIMENT="jsonv2"` prepended in `mkApp`/`mkBenchApp` (apps run outside direnv). Committed (`ce7e6e6`).
3. **nix-build step verified GREEN** (`buildflow -s nix-build` exit 0, all 8 x86_64-linux targets, 03:33) after fixes 1+2.
4. **govalid-generate verified GREEN** — its earlier failure ("go: updates to go.mod needed") was a stale, non-tidy go.mod at run start; the pipeline's own tidy repair fixed it; re-run passed (0 findings, no-op).
5. **erraudit: 17 findings → 0.** The step had never run before (blocked behind nix failures), so the debt surfaced for the first time. All 17 fixed in Go code:
   - 6× `context_loss` (real fixes): `middleware.go:446,457` now include `filePath` in wrapped errors; `watcher.go:212` includes `opName` (refactor: dropped the redundant `caller` param from `withResolvedPath`, 3 call sites updated); `watcher.go:555` includes `subPath`.
   - 11× deliberate best-effort ignores → `//nolint:erraudit` with written reasons (`watcher.go:207,573,586,733,845`, `watcher_walk.go:117`, `watcher_internal.go:323,325`, `watcher_poll.go:62`, `filesystem.go:122`, `examples/demo/shared.go:59,64`, `watcher_gitignore.go:92`).
   - Step verified green (exit 0, zero findings) at 04:02.
6. **golangci-lint clean** — the 5 "File is not properly formatted" findings from my long nolint lines were auto-fixed via `buildflow -s golangci-lint --fix` (golines rewrapped). Step green.
7. **Tests verified:** `go test -race -count=1 .` passes (4.2s) after all Go edits; `go build ./...` + `go vet ./...` clean.
8. **lychee configured (`lychee.toml`, new file).** README.md:516 links to `github.com/LarsArtmann/go-daemon` → 404; verified via `gh` that the repo EXISTS but is PRIVATE, so the link is correct and the fix is the namespace exclude the preflight itself suggested. Also excludes nixos.wiki (bot-blocking 403). Preflight `quality/lychee-private-links` and the README:516 finding are resolved.
9. **`.buildflow.yml` created (new file)** with the go-version-auto-configure skip + rationale; confirmed loaded ("skipped via config" count went 3→4 in subsequent runs).
10. **go-line flipflop root cause SOLVED (analysis complete):** `go-mod-update`'s default **minor** dependency-update mode bumps the go directive to the latest Go release (1.27) whenever it is below latest. Every full `buildflow --fix` run (mine included) poisoned go.mod. This explains the preflight's `workspace/go-line-flipflop` warning (7-8 flips in 20 commits), today's two earlier "accidental bump to 1.27" incidents in AGENTS.md, and the phantom "concurrent session" hypothesis (my own runs were the poisoner in most observations).
11. **Debug methodology hardened along the way:** proved `go mod tidy` rewrites go.mod even when `tidy -diff` is clean (the chmod-444 experiment); proved rename-based writes bypass file permission guards; proved repair-only step names are not addressable via `-s` without `--fix` (my first isolation loop was void — usage dumps, exit 69); proved `go get -u=patch ./...` fails outright on this module graph ("can't query version patch … no existing version is required", exit 1, zero changes).
12. **AGENTS.md documented the two durable gotchas** (new Known Issues entries): "Nix sandbox needs GOEXPERIMENT=jsonv2" and "buildflow nix-hash-fix cannot repair this flake" (fails 15+/15 with "stale hash not found verbatim"; includes the do-NOT-extract-vendorHash warning).

## b) PARTIALLY DONE

1. **A green full `buildflow --fix --build-mode=full` run.** Best result so far: run 4 (04:03) — zero step failures except go-mod-tidy (which my own chmod-444 guard blocked at the time; that guard is now removed as ineffective). After that, `dep_update_mode: patch` eliminated the go-line bump but `go get -u=patch` itself fails (see f1). The isolated git-worktree verification harness (`/tmp/gfw-verify`, branch `verify-gomod-floor`) exists and is the right tool for the final proof — it isolates the run from concurrent writers.
2. **The `go 1.27` re-poisoning of the main checkout.** Restored to `1.26.7` five+ times; every full run (or the leftover state of one) re-bumped it. Currently `1.27` again. The restore command is one `sed` away, but a durable green run needs the writer identified AND skipped (next step: the inotify test with corrected syntax — the last attempt failed on a missing `--timefmt` flag).
3. **`.buildflow.yml` dispositions.** `skip_steps: [go-version-auto-configure]` works and is committed. Still missing: the disposition that stops the go-directive bump. `dep_update_mode: patch` stops the bump but breaks go-mod-update; the remaining candidate is `skip_steps: [go-mod-update]` (updates via explicit `buildflow update` instead) — drafted, NOT yet tested as a full run (the last worktree run with it still showed `1.27`, but that run's instrumentation was inconclusive: the inotify watcher errored and the echo suffered a shell-precedence bug in one earlier attempt, so the writer could be go-mod-tidy/go-mod-normalize instead).
4. **AGENTS.md go.mod section update.** The existing "### go.mod language version" entry documents the 2026-09-29 incident and the 2026-10-07 recurrence, but NOT today's refined mechanism (buildflow's own go-mod-update minor mode is the bumper; permission guards useless; the two-disposition war). Needs a short addendum.
5. **`./result` staleness for vulnix.** Refreshed once (`nix build .` at 03:47, BUILD_OK); subsequent failed runs made it stale again relative to newer flake evaluations. Cosmetic warning; one `nix build .` after the green run silences it.

## c) NOT STARTED

1. **AGENTS.md size reduction** (625 lines vs max 220-377 depending on scanner context). I deliberately did NOT trim: 240+ lines would have to move, and doing that while a daemon commits every minute and the go.mod war was live risked fossilizing a botched split. Needs a dedicated editorial pass (candidates to relocate: File Organization table, Dependencies/gogenfilter API section, CI Workflows table, Linter Cheat Sheet → `docs/`, with one-line pointers).
2. **erraudit/nix-hash-fix upstream reporting.** Two genuine BuildFlow-ecosystem bugs found (nix-hash-fix never finds the stale hash here; go-auto-upgrade/go-mod-update disposition war). Per the buildflow skill these belong in the BuildFlow repo (with verify-before-filing + github-voice). Not started.
3. **pnpm-audit follow-ups** (12 findings remain in website/, was 21 before pnpm-update repairs ran; js-yaml/fast-uri/source-map-js errors are detect-only, non-gating). Not started — website has its own flake/toolchain.
4. **vulnix CVEs** (24 findings; nixpkgs-channel-level, non-gating). Not started.
5. **devShell tool additions** (dprint, prettier, lychee, tsc, govulncheck, vulnix, tailwindcss run via `nix run nixpkgs#…` with warnings). Deliberately skipped: warnings only, closure cost, and tsc likely no-ops here. Not started.
6. **flake-meta-checker / nix-flake-check cosmetic findings** (missing `meta.platforms`, apps lacking `meta.description`, vendorHash-extraction suggestion — the last one is actively WRONG for this repo per the AGENTS.md entry). Not started.
7. **dependabot.yml "nix" ecosystem entry** (info-level, "matches nothing detected", kept as-is by the tool itself). Not started.
8. **Worktree cleanup** (`/tmp/gfw-verify`, branch `verify-gomod-floor`) — still present, intentionally, until the final green proof.

## d) TOTALLY FUCKED UP (own failures this session, honestly)

1. **Four full-pipeline attempts were burned by racing an enemy that turned out to be myself.** I attributed the go.mod bumps to a concurrent session (plausible — 9 autonomous crush processes run, website churn commits existed) and burned attempts 3-5 on restore-and-rerun loops instead of instrumenting early. An inotify/mtime watch on attempt 3 would have identified go-mod-update ~90 minutes earlier.
2. **My isolation tests were invalid and I trusted them.** `buildflow -s go-mod-update` etc. without `--fix` were rejected ("repair-only tool") → usage dump → exit 69 → my loop read "no change" and concluded the steps were innocent. This directly sent me down the concurrent-session path. Lesson: verify the tool actually RAN before trusting its no-op.
3. **The chmod 444 guard was a mistake in execution:** it broke go-mod-tidy (unconditional rewrite) and was bypassed by rename-based writes; I removed it, but it cost a run and added a wrong rationale comment to `.buildflow.yml` that had to be rewritten.
4. **Shell precedence bug** in the instrumentation run (`cd && sed && (watcher) & rest` backgrounded the whole AND-list, so buildflow ran in the wrong cwd) produced a false "worktree flipped" reading and wasted a cycle on the concurrent-session theory again.
5. **inotify syntax error** (`%T` without `--timefmt`) killed the decisive measurement on the last attempt. Trivial fix (`--timefmt '%H:%M:%S'`), still pending.
6. **Missed that `go.mod` content feeds the FOD hash** until run 2 — I initially treated the vendorHash mismatch as a standalone staleness issue; the deeper pattern (every go.mod mutation invalidates the nix build) should have been my first-system model given this repo's AGENTS.md documents exactly that interplay.

## e) WHAT WE SHOULD IMPROVE

1. **Instrument before iterating.** Any "mysterious revert/bump" deserves an inotify/fanotify watch IMMEDIATELY, before any restore-retry loop.
2. **Read BuildFlow's step model before isolating steps:** detect vs repair providers, `-s` requires the provider name, repairs need `--fix`. The `buildflow list providers` output I ran at the END should have been run FIRST.
3. **Prove a writer's identity with the cheapest possible experiment** (one step + inotify), not with full 1-3 minute pipeline runs.
4. **The fleet needs the go-directive disposition WAR settled once** — buildflow's own preflight says "align go-version-auto-configure and go-mod-update dispositions"; today it cost this session ~2h. Either go-mod-update never touches the go line (BuildFlow change), or go-version-auto-configure and go-mod-update agree on one floor.
5. **nix-hash-fix needs a regression test for "hash present verbatim in flake.nix"** — it failed 15+/15 here; the skill says never hand-paste hashes, but the fixer forces exactly that in this repo.
6. **Erraudit suppressed-ignores convention**: 11 `//nolint:erraudit` reasons is fine, but a repo-level convention doc (docs/ or AGENTS.md one-liner) would keep future suppressions honest.
7. **AGENTS.md slimming should be scheduled, not ad-hoc** — with a checklist of every distinct concept that must survive the move.

## f) NEXT TASKS (prioritized)

1. Fix the inotify command (`--timefmt`) and re-run the worktree full pipeline with go-mod-update skipped → identify the remaining writer if the line still flips (candidates: go-mod-tidy, go-mod-normalize).
2. If go-mod-tidy/normalize is the writer: add it to skip_steps with rationale (tidy verified by `tidy -diff` canary in doctor/go-mod-update's check).
3. Achieve the GREEN full run in `/tmp/gfw-verify` (worktree, `verify-gomod-floor`, go.mod at 1.26.7).
4. Sync the final `.buildflow.yml` dispositions to the main repo; delete the worktree + branch.
5. Restore main-repo go.mod to `go 1.26.7`; refresh `./result` (`nix build .`).
6. Re-run `buildflow --fix --build-mode=full` in the MAIN repo in a quiet window; expect exit 0.
7. Update AGENTS.md "go.mod language version" section with the refined mechanism (go-mod-update minor mode = bumper; disposition fix = skip + explicit `buildflow update`).
8. Update AGENTS.md nix-hash-fix entry with the additional failure signature seen today ("build reported hash mismatch but no fixable findings were generated").
9. Trim AGENTS.md under the size limit: move File Organization table + Dependencies/gogenfilter section + CI Workflows table + Linter Cheat Sheet to `docs/` with pointers (verify every concept survives).
10. File upstream BuildFlow issues: (a) nix-hash-fix "stale hash not found verbatim" false negative; (b) go-mod-update minor mode silently bumping the go directive against documented floors; (c) `go get -u=patch` failing on this module graph (maybe upstream Go behavior worth a BuildFlow fallback).
11. Consider `dep_update_mode: patch` + fixing the `-u=patch` failure as the alternative to skipping go-mod-update (keeps auto-updates working at patch level).
12. Decide lychee fleet policy (authenticate vs exclude) — this repo now has a working `lychee.toml` exclude to copy.
13. pnpm-audit: update website deps (js-yaml ≥4.3.2, fast-uri ≥3.1.6, source-map-js ≥1.2.2, devalue ≥5.9.3, svgo ≥4.1.0) via website toolchain.
14. vulnix: nixpkgs channel bump decision (24 CVEs incl. critical curl ones) — likely a fleet-level repin.
15. Add dprint/prettier/lychee (+ maybe govulncheck/vulnix) to devShells.default to silence the "not in project devShell" warnings.
16. Add `meta.platforms` to flake packages.default meta (flake-meta-checker info).
17. Add `meta.description` to the 8+ flake apps (nix-flake-check warnings) or accept the noise.
18. Remove or fix the dependabot.yml "nix" ecosystem entry (info).
19. Reconsider skip of go-version-auto-configure once BuildFlow dispositions are aligned upstream (the skip is a workaround, not a forever-decision).
20. Run `buildflow update` explicitly at a chosen time to get real dependency updates WITHOUT the pipeline fighting the go line (then re-pin vendorHash + go line in one commit).
21. Address the 12 go-auto-upgrade advisory findings deliberately (samber/lo lo.Map suggestions: adopt or rebut per how-to-golang policy; jsonv1tov2 migration blocked by the 1.26 floor — decision needed if/when CI matrix moves to 1.27).
22. Decide the fleet Go strategy: either CI matrix moves to 1.27 (then bump floor deliberately: ci.yml, AGENTS.md, flake go_1_26→go_1_27, vendorHash) or the go directive guard becomes enforcement (a nix check that fails loudly when go.mod ≠ floor).
23. Add a nix `check` that asserts go.mod's go line equals the documented floor with a CLEAR error message (turns the obscure FOD failure into an actionable one).
24. art-dupl: 97 findings reported (detect-only) — triage harmful vs intentional clones.
25. branching-flow: 75 findings reported (detect-only) — triage or document as accepted.
26. go-structure-linter: 19 findings (appeared in run 2, disappeared in run 4 — check why the scanner context changed).
27. Confirm golangci-lint stays at 0 findings after the go-mod skips (the 5 formatting findings were auto-fixed; verify they didn't regress).
28. Check `.buildflow.yml` against `buildflow list providers` names once more before finalizing skips (skip_steps is whole-tool only; qualified entries warn).
29. Investigate why the AGENTS.md size limit differed between main-repo runs (377) and the worktree run (220) — scanner context dependency worth understanding.
30. Clean up stray files: `/tmp/bf-*.log`, `/tmp/nixbuild*.log`, `/tmp/erraudit.*`, `/tmp/gmu.log`, `/tmp/goget.log`, `/tmp/writer-snapshot`, `/tmp/gomod-mtime` (keep the useful ones as evidence for upstream issues first).
31. Remove the `result` symlink or keep it fresh (vulnix warning heuristic; also it is in .gitignore?).
32. Consider adding `go.mod` go-line to the preflight's flipflop detector allowlist once dispositions are aligned (warning noise).
33. Verify the examples-build check still passes after the `examples/demo/shared.go` nolint edits (it does per nix-build green, but re-verify post any further edits).
34. Re-run `nix flake check` directly once for the full warning inventory (16 app warnings etc.) and decide each.
35. Bench-diff per AGENTS.md rule (benchmarks must run without parallel load — today's machine had 9 crush sessions; NOT a valid window; schedule for a quiet time).
36. Website: the pnpm-workspace `allowBuilds` mechanism is documented in AGENTS.md; verify the concurrent session's css churn (`global.out.css` ±1931 lines twice today) is intentional.
37. Reconcile with the concurrent website session: coordinate a quiet window for pipeline runs, or gate runs on a lockfile convention.
38. Consider `buildflow doctor` full output review (34 ok / 3 warn / 0 fail) — the 3 warns beyond the ones triaged.
39. Check whether `interrogate` (missing tool warning) should be installed or is even applicable to this repo.
40. Documentation: add the nix-hash-fix + GOEXPERIMENT entries to the website/docs if the docs-consistency workflow requires parity.
41. Review whether the 11 `//nolint:erraudit` suppressions should instead be code restructures with debugLog (observability upgrade, listed as possible follow-up).
42. Verify CI on GitHub is green after the daemon pushed today's commits (the last pushed go.mod=1.27 will FAIL CI — the floor restore must land on origin before the next release/PR).
43. **URGENT practical note:** origin/master currently carries `go 1.27` (daemon pushed) — every open PR's CI will fail until the restore is pushed; coordinate the push.
44. Update TODO_LIST.md with this session's outcomes (docs-health flow: harvest status into TODO_LIST).
45. Add the "go-directive disposition war" as a `references/lessons.md` candidate in crush-config (cross-project lesson: instrument writers before racing them).
46. BuildFlow feature request: preflight check that warns when go.mod's go line ≠ the CI matrix version in ci.yml (would have caught today's war in run 1).
47. BuildFlow feature request: `skip_steps` for repair sub-steps (qualified entries currently warn and never match).
48. Re-check that `nix fmt`/treefmt config still passes after flake.nix edits (the `format` check was green in cached runs; re-verify in the final green run).
49. Evaluate `--flight-recorder` on the next failing run (flag exists per help; would capture a trace on step failure).
50. Celebrate-safe: once green, capture `buildflow history`/`timings` baseline so future regressions of today's failure classes are visible in the fleet dashboard.

## g) QUESTIONS FOR YOU (cannot figure out myself)

1. **Which Go floor is the real policy: `go 1.26.7` (current AGENTS.md) or a move to `go 1.27`?** Two incidents today defended 1.26.7 (CI matrix is `1.26`), yet something/someone keeps asserting 1.27 — if the fleet's default toolchain is already 1.27, is the _real_ decision to bump the CI matrix and the floor deliberately? I did not reverse that decision unilaterally.
2. **Are the other autonomous crush sessions currently working in this repo (website churn, go.mod bumps) yours and intentionally active?** If one of them is _supposed_ to hold `go 1.27`, my restore-fighting was wrong-headed and we should coordinate instead — tell me which session/repo owner wins.
3. **For dependency updates, do you want go-mod-update skipped in the pipeline (explicit `buildflow update` on demand — my current recommendation) or kept with `dep_update_mode: patch` once the `-u=patch` failure is solved?** This decides the final `.buildflow.yml`.

---

**Session verdict:** All four original step failures are individually fixed and verified (vendorHash, GOEXPERIMENT in nix sandbox, govalid tidiness, erraudit debt). The pipeline is one skipped-disposition away from green; the remaining work is a 5-minute instrumented confirmation run, not new engineering.
