# Release & CI Gotchas

> Moved out of `AGENTS.md` on 2026-11-06 (AGENTS.md size gate: max 220 lines).
> `AGENTS.md` keeps a one-line index of each gotcha and links here.

## Release / CI gotchas

| Gotcha                                    | Rule                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| release-please needs a repo permission    | Fails with "GitHub Actions is not permitted to create or approve pull requests" when off. Fix: `gh api -X PUT repos/LarsArtmann/go-filewatcher/actions/permissions/workflow --input - <<< '{"default_workflow_permissions":"read","can_approve_pull_request_merge_requests":true}'`. On PR merge release-please creates the tag AND the GitHub Release; `release.yml` only fires for manually-pushed `v*` tags.                                                                                                                                                                                                                                                                                                                                                                            |
| release-please is path-scoped             | Config-file mode with `exclude-paths: ["website", "docs", ".github"]` — commits touching ONLY those paths never cut releases. Convention: `chore/docs/build(website)` for non-module changes, never `fix(website)`/`feat(website)` (that shipped a Go-less v2.4.1 on 2026-10-07).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| golangci-lint version pin                 | `ci.yml` and `release.yml` pin `version: v2.15.0`; the flake uses nixpkgs `pkgs.golangci-lint` (2.15.0 today) — keep them aligned, and remember `.golangci.yml` needs the `exhaustruct_v5` key (v2.13+; v2.12 rejects the config).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Release-PR CI "workflow file issue"       | Release-please PRs show `conclusion: failure` with ZERO jobs (hit v2.4.0 + v2.4.1). GitHub-side quirk — identical workflows run green on Dependabot PRs and master pushes. Verify releases via master CI on the merge commit; the release itself is unaffected.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| Verify built slugs before publishing URLs | GitHub slugs strip dots: `migration-v2.3-to-v2.4` → `/guides/migration-v23-to-v24/`. Check `ls dist/guides/` (or the live URL) before `gh release edit`/deploy — the unchecked slug 404'd in the published v2.4.0 notes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Auto-commit daemon races                  | The daemon commits (and occasionally resets) local master while you work. `git fetch` + check `git log HEAD..origin/master` before pushing; rebase if diverged. Never assume your last local commit is still HEAD. Since branch protection was removed (2026-10-12) it CAN push again; during the 2026-10-07 → 2026-10-12 protection window its stale commits piled up locally and needed `git rebase --skip` when their content was already upstream.                                                                                                                                                                                                                                                                                                                                     |
| Git Town syncs ALL local branches         | `git sync` (= `git town sync`) merges master into every local branch — one stale merged-PR branch conflict-pauses the whole run (2026-10-07: 18 zombie branches from the merge wave tripped it on `chore/codespell-hygiene`; `git town undo` unwinds cleanly). Debris sweeps must enumerate `git branch -vv` (upstream `gone`) and run `git worktree prune` (stale /tmp worktree registrations block branch deletion). Before `git branch -D`, prove content is upstream: `git cherry master <branch>` (`-` = patch-equivalent; inspect every `+` — daemon auto-commit tips are junk snapshots, real commits need a content diff vs master).                                                                                                                                               |
| Branch protection REMOVED (2026-10-12)    | REMOVED 2026-10-12 (Lars: no master protection needed — direct pushes work again). Was LIVE 2026-10-07 → 2026-10-12: 8 required contexts (Go Directive Floor, Test 1.26 × ubuntu-24.04, Test 1.26 × ubuntu-26.04, Examples Build, Lint, Validate Commit Messages, README vs API_STABILITY Drift Check, Exported Symbols in Docs), strict up-to-date, `enforce_admins`, `required_linear_history`, NO review requirement, repo-level auto-merge DISABLED. Restore one-liner: `gh api -X PUT repos/LarsArtmann/go-filewatcher/branches/master/protection --input docs/branch-protection-backup-2026-10-12.json`. If ever re-enabled: strict up-to-date forces sequential PR merges (`gh pr update-branch N` → wait green → `gh pr merge N --rebase`) and the auto-commit daemon cannot push. |
| Windows/macOS CI legs (probe)             | The test matrix also runs macos-15 + windows-2025 (informational). Coverage steps need `shell: bash` — PowerShell splits `-coverprofile=coverage.out` into `-coverprofile=coverage` + `.out`. POSIX-shape tests skip via `skipOnWindows(t)`; `failedPaths` lookups MUST use `pathKey()`. Full detail: docs/gotchas.md #27.                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| ubuntu-latest migrates to 26.04           | Rollout starts 2026-10-19, complete 2026-11-19 (actions/runner-images#14748). Audited 2026-10-07: all workflows are pinned-SHA actions + explicit Go versions, no Python/apt-key/cgroup exposure. The CI test matrix runs both ubuntu-24.04 and ubuntu-26.04 (early compat proof); all other jobs are pinned to ubuntu-24.04 for determinism through the rollout window. Re-evaluate pins when GitHub announces 24.04 brownouts (expected 2027).                                                                                                                                                                                                                                                                                                                                           |

## go.mod language version

The `go` directive must stay at the CI matrix floor (`go 1.26.7`); the local
toolchain is newer and three BuildFlow tools silently bump it (strace-proven
2026-10-07, all skipped via `skip_steps` in `.buildflow.yml`):
`go-version-auto-configure` (wants major.minor-only), `go-mod-update` (minor
mode bumps to latest Go), and `go-structure-linter` (buildflow 202b114+
auto-repairs its "Go 1.27 is available" finding mid-run, backups in
`/tmp/go-structure-linter-backups/`). `go-mod-normalize` is exonerated
(candidate downgrade is rejected by the dependency-floor gate, then atomically
restored). Restore: `sed -i 's/^go 1\.27$/go 1.26.7/' go.mod`. If a daemon
bump also dragged in deps that themselves require go 1.27 (happened 2026-11-06:
gogenfilter v3.7.0 declares `go 1.27`), `go get <dep>@<older>` deadlocks under
`GOTOOLCHAIN=local` because it must load the current graph first — downgrade
via `go mod edit -require=<dep>@<older>` (no graph validation), then `go mod
tidy`. The dev shell
also pins `GOTOOLCHAIN=local` (flake.nix devShell, PR #38) so the local
toolchain never silently switches to a newer one that would re-apply the bump.
The bump can
land on origin even when you ran no go commands (daemon commits) — check the
`go` line whenever the daemon commits go.mod; dependent PR CI recovers only
after master is fixed plus a branch update/re-run.

**Structural guard (added 2026-10-07)**: the floor is now enforced, not just
remembered — a `Go Directive Floor` CI job asserts `go 1.26.7` and blocks the
test matrix (`needs: go-directive`), and `nix flake check` runs the same
assertion as the `go-directive` check. Intentional bumps must change go.mod
AND both guards in one commit.

**No go 1.27 CI leg (deliberate, 2026-10-07)**: a 1.27 matrix entry was
tried and removed the same day — under the Go 1.27 toolchain,
`encoding/json/v2` (imported in `event.go`) requires the go directive to be
≥ 1.27, so the leg cannot build while the floor is pinned at 1.26.7 (CI run
37581977239: "json.Marshal requires go1.27 or later (file is go1.26)").
Re-adding the 1.27 leg REQUIRES the intentional floor bump to 1.27 — which
also means revisiting the json/v2 usage or accepting the bump. This is
recorded as a v3 decision (docs/research/v3-decisions.md).

**Worktrees and nix**: flake evaluation IGNORES dirty state in linked
worktrees (no `-dirty` drv suffix; it builds the COMMITTED go.mod). Commit the
floor before any worktree proof run.

**BuildFlow on this shared machine**: a fleet binary can gain NEW tools
mid-session — on any behavior change run `buildflow doctor` + `buildflow list
providers` BEFORE rerunning (a stale-binary run cost ~2h on 2026-10-07). Under
high load (many concurrent agent sessions), steps get killed at spawn and
report false failures; serialize heavy runs or wait for a quiet window.
Observed 2026-10-07 (second time): the binary changed from `202b114` to
`acdb606-dirty` between sessions — treat the version in TODO_LIST's BuildFlow
item as a snapshot, not a constant.

## Known issues

### Website build (pnpm 11) build-script approvals

Build-script approvals live in `website/pnpm-workspace.yaml` under `allowBuilds:` (`esbuild: true`) — pnpm v11 ignores `pnpm.*` in `package.json` and silently skips unapproved postinstall scripts, so `astro build` then fails on a missing esbuild binary. A placeholder value (e.g. `esbuild: set this to true or false`) silently disables the whole key (cmdguard incident, fixed 2026-09-19).

### Nix sandbox needs GOEXPERIMENT=jsonv2

`event.go` imports `encoding/json/v2` (GOEXPERIMENT-gated). CI sets `GOEXPERIMENT: jsonv2` workflow-wide (ci.yml) and direnv exports it locally, but nix sandbox derivations receive neither, so every Go target fails with "build constraints exclude all Go files in .../encoding/json/v2". The flake sets it wherever Go compiles: `packages.default` and the four `runCommand` checks via `env.GOEXPERIMENT`, and every `mkApp`/`mkBenchApp` script exports it (apps run outside direnv). Any new Go-building derivation or app must set it too (fixed 2026-10-07).

### buildflow nix-hash-fix cannot repair this flake

`nix-hash-fix` has failed 15+/15 runs here: after a hash mismatch it reports "the stale hash was not found verbatim in any .nix file" even when `vendorHash = "sha256-…"` sits verbatim in flake.nix (BuildFlow repo bug; the 2026-10-07 fix was hand-applied after the fixer gave up). Until fixed upstream, update the `vendorHash` in flake.nix to the `got:` hash from `nix build` output by hand, then verify with `nix build`. Consequence: do NOT extract vendorHash into a separate file (the nix-checker suggestion) — it would break the one manual repair path that works.

### buildflow pnpm-audit result cache replays after lockfile-only fixes

The detector result cache keys pnpm-audit on `package.json` but NOT `pnpm-lock.yaml`, so
vulnerabilities fixed via `pnpm audit --fix=update` (lockfile-only change) keep gating the
full pipeline with the stale 9-error finding set even though `pnpm audit` in `website/` is
clean. Surgical purge (the reference's DB path is wrong; the real DB is `cache.db`):

```bash
nix shell nixpkgs#sqlite -c sqlite3 ~/.cache/buildflow/cache.db \
  "DELETE FROM result_cache WHERE value LIKE '%pnpm-audit%';"
```

A no-cache run (`BUILDFLOW_NO_RESULT_CACHE=1`) proves the fix but does NOT overwrite the
stale entry (BuildFlow gap), and it re-executes every step — under high machine load the
re-run hits step timeouts (go killed at spawn) and produces false step failures. Purge the
entry, then re-run the normal cached pipeline. Note `pnpm audit --fix` in pnpm 11 needs an
explicit strategy (`--fix=update` re-resolves the lockfile; `--fix=override` adds
overrides to `website/pnpm-workspace.yaml`). 2026-10-07: 12 vulns (9 high) → 0 via
update + one override (`postcss-selector-parser@<7.1.6: ^7.1.6`, forced major bump,
website build verified green after).

2026-10-12: the same replay class hit the `oxlint` entry — a stale zero-finding
JSON masked two real `no-unused-vars` errors (dead imports in sync-changelog.mjs).
Purge is fleet-DB shared, so scope by repo hash: `SELECT DISTINCT repo_hash FROM
result_cache WHERE value LIKE '%go-filewatcher%'`, then
`DELETE FROM result_cache WHERE repo_hash='<hash>' AND value LIKE '%"Tool":"oxlint"%'`.

## Accepted BuildFlow noise (baselined or documented non-fixes)

A full `buildflow --fix --build-mode=full` run ends with a set of warning/info
findings that are deliberate. Each one is either baselined in
`.buildflow-baseline.json` (the gate stays armed for NEW findings) or recorded
here as a non-fix, so nobody re-litigates them every session.

| Finding                                                                 | Disposition                                                                                                                                                                                                |
| ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `file-size-check`: 20 Go files over 350 lines (error)                   | Baselined. Split campaign tracked in `TODO_LIST.md`; drop entries as files split. Raising `max_file_size` is NOT an option — the key is inert in BuildFlow (BuildFlow#75, re-verified 2026-11-06).         |
| `cue-vet`: `ci.yml` / `release.yml` use `ubuntu-24.04` (+ an OS matrix) | Baselined. Deliberate runner determinism through the `ubuntu-latest` 26.04 rollout (complete 2026-11-19); re-check then and delete the entries (revisit task in `TODO_LIST.md`).                           |
| `go-auto-upgrade`: "jsonv1tov2 migration skipped" (×3)                  | Non-fix while the floor is `go 1.26.7`: `encoding/json/v2` needs a ≥ 1.27 directive, so migrating would stop the module compiling. Re-triage only with an intentional floor bump.                          |
| `go-auto-upgrade`: `lo.Map` suggestions                                 | Non-fix: `github.com/samber/lo` is not a dependency and adding one for a for-range loop is not an auto-rewrite BuildFlow may make. Adopting `lo` is a separate dependency decision.                        |
| `go-structure-linter`: "consider `internal/`" / "consider `assets/`"    | Non-fix: this library is deliberately one flat `filewatcher` package (see AGENTS.md "Single Package Layout"); `assets/` is about `website/dist` build output, which is generated.                          |
| `go-structure-linter`: "Go 1.27 is available"                           | Non-fix: the `go 1.26.7` floor is policy, enforced by the `go-directive` check and CI job.                                                                                                                 |
| `nix-checker`: extract `vendorHash` / `hash` into their own files       | Non-fix, deliberately reversed advice: `nix-hash-fix` cannot repair this flake, so the inline `vendorHash` is the only working manual repair path (see "buildflow nix-hash-fix cannot repair this flake"). |
| `dependabot-auto-configure`: nix entry "matches nothing"                | Tool false positive: `flake.nix` and `flake.lock` both exist at `/`. The `nix` ecosystem entry stays.                                                                                                      |
| `art-dupl` / `branching-flow`                                           | Advisory structural noise: duplicated benchmark scaffolding and the intentionally wide `Watcher` struct. Not gated.                                                                                        |
| `vulnix` (21 advisories)                                                | Comes from the nixpkgs toolchain closure (`binutils`, `bison`, `brotli`, `gcc`), not this repo's Go dependencies. Fixed by nixpkgs bumps, tracked upstream.                                                |
