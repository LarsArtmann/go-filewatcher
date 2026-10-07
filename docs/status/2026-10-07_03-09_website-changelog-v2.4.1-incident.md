# Status Report: Website Fixes, Changelog Sync & v2.4.1 Incident

**Date:** 2026-10-07 03:09 CEST
**Scope:** This continuation session — the `deploy` command, the card-overflow bug, the missing-2.3.0/2.4.0 changelog investigation, and the v2.4.1 mis-release they uncovered. Builds on the earlier `2026-10-07_02-24_v2.4.0-release-postmortem.md`.
**Headline:** Website deployed 3× with real fixes; website changelog now auto-synced from CHANGELOG.md (single source of truth); release-please path-scoped after a `fix(website)` commit leaked a Go-only-less patch release (v2.4.1). Two of my own commits caused the incidents I then had to fix — honest accounting below.

---

## a) FULLY DONE

| # | Item | Evidence |
|---|------|----------|
| 1 | **Website deployed** with the v2.4.0 docs (incl. migration guide) | Firebase target `filewatcher`, release complete; guide live |
| 2 | Fixed migration-guide link in v2.4.0 GitHub release notes (dots are stripped from slugs: real URL is `/guides/migration-v23-to-v24/`) | `gh release edit`, live fetch 200 |
| 3 | Committed pnpm's `minimumReleaseAgeExclude` additions so builds are reproducible | `843bd9d` |
| 4 | **Card-grid overflow into "On this page" TOC fixed** — root cause: CardGrid `1fr 1fr` tracks size to min-content; long identifiers (`WithMaxWatchesSafetyFraction`) pushed the grid into the TOC. Fix: `min-width: 0` + `overflow-wrap: anywhere` (unlayered, beats `starlight.components`), site-wide | `7c4ea79`, rule verified in live CSS bundle |
| 5 | **Website changelog was missing 2.2.1/2.3.0/2.4.0** — page was a hand-copied snapshot frozen at 2.2.0. Replaced with generated page: `website/scripts/sync-changelog.mjs` regenerates `changelog.mdx` from repo `CHANGELOG.md` on every `pnpm build`/`pnpm dev` | Script wired in `package.json`; live `/changelog` verified with all 11 versions, newest first (node https check: 2.4.1/2.4.0/2.3.0 all present) |
| 6 | **release-please path-scoped** — `release-please-config.json` + `.release-please-manifest.json` (seeded 2.4.1), workflow switched to config-file mode; `exclude-paths: ["website", "docs", ".github"]` | `744a0ee` + daemon's `5de2ff5`; run green in 17s, clean no-op, no spurious PR |
| 7 | **v2.4.1 release notes annotated** — "no Go code changes, safe to skip" note prepended | `gh release edit v2.4.1` |
| 8 | AGENTS.md updated: changelog-sync mechanism (never hand-edit the generated file), release-please path-scoping + commit-scope convention | `6afc6a7`, `744a0ee` |
| 9 | Website redeployed with 2.4.1 changelog entry picked up by the sync | live markers verified |

## b) PARTIALLY DONE

| Item | Done | Missing |
|------|------|---------|
| Verify v2.4.1 on the module proxy | tag/release exist, merged by user | never checked `go list -m @v2.4.1` / proxy serves it (it's a no-Go-changes release, so impact is nil, but unverified) |
| Website build reproducibility | pnpm workspace/lockfile committed | `node_modules` is still machine-local; no CI job builds the website, so drift/`astro: command not found`-class failures are only caught at manual deploy time (hit once this session) |
| AGENTS.md accuracy | gotchas added | the "Release / CI Gotchas" section is accreting entries from parallel sessions (go.mod recurrence, release-PR startup failures, mine) — correct but increasingly a wall of text; could use consolidation |

## c) NOT STARTED

1. pkg.go.dev verification for v2.4.0/v2.4.1.
2. FEATURES.md / TODO_LIST.md / ROADMAP.md / README.md refresh for v2.4.x.
3. Dependabot PRs #31, #32, #14.
4. CI guard for go-directive drift; Go 1.27 in the matrix (carried over from previous report).
5. `nix flake check` / `nix build .` vendorHash validation (carried over).
6. Website CI build job (gap made concrete this session: `astro: command not found` because node_modules was absent and nothing checks it).
7. GitHub Release for v2.4.0/2.4.1 render-check of full notes.

## d) TOTALLY FUCKED UP (self-inflicted, both fixed)

1. **`fix(website): stop card grids…` triggered v2.4.1.** I wrote that commit type/scope myself, hours after documenting that release-please parses every conventional commit repo-wide. The user merged the resulting PR; a Go patch release with zero Go changes shipped to the proxy. Immutable tag → cannot be undone, only annotated + prevented (exclude-paths + convention). **This was avoidable by me alone.**
2. **Release-notes link 404**: I published the v2.4.0 notes with `migration-v2.3-to-v2.4` without checking the built slug (dots stripped → `migration-v23-to-v24`). Caught one turn later during `deploy` verification; fixed via `gh release edit`. Verify URLs against the build output before publishing.
3. Near-miss: my push collided with the daemon twice more (rebase dance each time; nothing lost). Also two of my file sets got absorbed into daemon `chore: auto-commit` commits, discarding my conventional-commit messages — cosmetic, but it means the git history under-attributes this work.

## e) WHAT WE SHOULD IMPROVE

1. **Scope discipline**: non-module changes must use `chore/docs/build(website)` — now both configured (exclude-paths) and documented; the remaining failure mode is *me* forgetting again.
2. **Verify published URLs against the built slug** before `gh release edit`/deploy — a 5-second `ls dist/guides/` would have caught the 404.
3. **Website deploy is manual and unguarded**: no CI build, no link checker, no preview. Cheapest next step: a website build job + `html-validate` (already a devDep) in CI.
4. **Changelog page was duplicated content** — now generated. Generalize the instinct: any repo file mirrored into `website/` should be generated, never copied.
5. **Daemon absorbs in-flight work**: it committed my staged-but-uncommitted file sets twice (losing my commit messages), and raced my pushes three times. `git status` + fetch before every commit/push is now muscle memory, but the daemon excluding `.github/`, `go.mod`, `release-please*.json` would remove the whole class.

## f) NEXT — concrete, Pareto-ordered

1. Verify proxy serves v2.4.1 (`go list -m github.com/larsartmann/go-filewatcher/v2@v2.4.1`).
2. Add website CI job: `pnpm install && pnpm build` (+ `html-validate`), catching missing-node_modules/ESLint-class breakage before deploy.
3. Add a link checker to the website build (e.g. `html-validate` or an Astro dead-link plugin) — would have caught the migration-guide 404.
4. Deploy automation: `firebase deploy` from CI on master (or at least a workflow_dispatch) so deploys don't depend on this laptop.
5. Document the deploy runbook in AGENTS.md (install → build → deploy via flake apps, `trailingSlash: false` quirk, cleanUrls).
6. pkg.go.dev check for v2.4.0/2.4.1; request crawl if stale.
7. FEATURES.md: mark v2.4.x shipped features DONE.
8. TODO_LIST.md harvest + ROADMAP refresh.
9. README staleness pass (Go version, feature bullets, v2.4 links).
10. CI guard: fail if go.mod `go` directive > matrix Go version.
11. Add Go 1.27 to the test matrix.
12. Required status checks on master/release PRs.
13. `nix flake check` + `nix build .` (vendorHash) — still unrun since the go.mod fix.
14. `release.yml` `workflow_dispatch` trigger (still dead code for GITHUB_TOKEN tags).
15. Test asserting `ErrorContext.Event` population (added in the lint fix, untested).
16. Review/merge dependabot #31/#32; close stale #14.
17. Benchmark baseline re-capture post-v2.4.x.
18. Consider exposing gogenfilter sqlc output-dir config (recover deliberate `models.go` filtering).
19. README/website API docs: document gogenfilter v3.6 weak-filename semantics.
20. Consolidate the "Release / CI Gotchas" AGENTS.md section (parallel sessions stacked 5 subsections; tighten into a table).
21. Daemon: exclude `.github/`, `go.mod`, `release-please*.json`, `website/flake.nix` from auto-commits (cross-repo tooling — needs your buy-in).
22. Ubuntu 26 runner migration audit (**deadline 2026-10-12**, warnings already in every run log — was Oct 19 in the log text; treat the earlier date as the safe target).
23. Sweep `docs/status/` — mark items done by this session (website link, changelog sync, release-please scoping, v2.4.1 annotation).
24. Star/OG image check: does the new og image actually render on social cards (added by a parallel session)? Unverified.
25. Add `editLink`/`lastUpdated` to astro.config (website-launch skill retrofit list) — check if already on.
26. Demo video on the landing page (skill default; site currently has none) — bigger item, needs your call.
27. v2.4.1's CHANGELOG entry on the website is the release-please stub — acceptable, but the sync script could post-process; probably not worth it.
28. `changelog.mdx` is a committed build artifact — consider gitignoring it + generating in CI to avoid stale-checkout drift (low priority; build regenerates).
29. Check CodeRabbit "bot user not eligible for review" config.
30. Retro: add "verify built slugs before publishing URLs" to AGENTS.md release runbook.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Did you intend to merge the v2.4.1 release PR #36?** It contained no Go changes (my `fix(website)` triggered it). If it was an auto-merge/accident, I'll treat "user merge" as not-a-safety-net; if deliberate, fine — either way the annotation stands.
2. **Should website deploys be automated from CI** (on master push / workflow_dispatch with a Firebase service-account secret), or do you want to keep depops manual-and-local? Determines items 2/4 in the next list.
3. **Daemon governance (again, now with receipts):** it absorbed two of my in-flight file sets (losing my commit messages) and raced pushes three times this session. Is its config editable from here, or is it external tooling I should only document around?

---

**Bottom line:** every user-visible bug reported this session is fixed and verified live (deploy, TOC overflow, changelog 2.3.0/2.4.0), the changelog can never drift again, and release-please can no longer be triggered by website commits. The two incidents fixed along the way were both caused by my own earlier choices (`fix(website)` scope; unchecked URL slug) — the mechanisms now prevent both classes, but the better version of me checks slugs before publishing and scopes commits correctly the first time.
