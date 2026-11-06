# TODO List

**Last Updated:** 2026-11-06 (BuildFlow green run: harvested the 2026-10-12 and
2026-11-06 reports, recorded the file-size/docs-annotate/cadence ratchets)

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## Urgent

- [ ] **Ubuntu 26 runner migration audit — deadline 2026-10-12** — GitHub is
      migrating `ubuntu-latest`; warnings already appear in every run log.
      Audit all workflows (`.github/workflows/*.yml`) for runner-compat
      breakage and pin or adapt.
      (`src: 2026-10-07_03-09 §f22, 2026-10-07_02-24 §f11`)

## Testing & Platform

- [x] **Windows + macOS CI matrix (probe legs)** — DONE 2026-10-07 (PR #47):
      `windows-2025` + `macos-15` run in the test matrix (informational, not
      required). The probe surfaced and fixed a real test bug (`failedPaths`
      raw-path lookup only worked on Linux) and 34 POSIX-shape assertion
      failures (skipped via `skipOnWindows(t)`; `ExampleEventPath` split
      per-OS). Windows coverage steps need `shell: bash` (PowerShell splits
      `-flag=file.ext`). Platform notes in README; gotcha #27 in
      docs/gotchas.md.
      (`src: 2026-10-07_02-24 §f8`)
- [ ] **Promote windows/macos probe legs to required checks** — decision point
      after the legs bake for a week: promote, or keep informational and
      document why. Prerequisite for promotion: per-OS expectations replacing
      the `skipOnWindows` skips (next item).
      (`src: 2026-10-07_02-24 §f8`)
- [ ] **Per-OS test expectations for Windows** — the `skipOnWindows(t)` skips
      hide correct-per-OS behavior behind POSIX-shaped assertions
      (`TestCleanPath`, `TestNormalizePath_EdgeCases`, `TestEventPath_Join/Dir`,
      `TestFilterExcludePaths`, `TestFilterIgnoreDirs`/`Hidden` + case
      variants). Rewrite with `filepath`-based expectations so Windows runs
      the assertions too.
      (`src: PR #47 probe run 37602089069`)
- [ ] **Expand fuzz tests** — current corpus covers `FilterRegex`,
      `FilterExtensions`, `FilterIgnoreGlobs`, `OpUnmarshalText`, `FilterMinSize`,
      and `PathKey` (NFC/idempotency/case-folding). Add fuzzers for
      `FilterAnd`/`FilterOr`/`FilterNot` composition, `Event` JSON round-trip,
      the gitignore matcher, `FilterIgnoreDirsCaseInsensitive`, and
      `MiddlewareDeduplicateCaseInsensitive`.
      (`src: 2026-08-12_19-22 §f16`)
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under
      load.
- [ ] **Investigate the July 6-hour Benchmark hang** — PR #14's Benchmark job ran
      02:23→08:24 UTC (6h, killed at runner timeout): possible benchmark/watcher
      deadlock. Timeout caps now bound the blast radius (`b2caf11`); root cause
      still unknown.
      (`src: 2026-10-07_03-10 §b5/f4`)
- [x] **Test: `WithContentHashing()` + `WithContentHashMaxSize(0)` interaction** —
      DONE 2026-10-07 (PR #46): 0 disables hashing entirely; the option-order
      trap (MaxSize(0) before WithContentHashing() is re-substituted with the
      10 MiB default) is pinned explicitly.
      (`src: 2026-08-12_19-22 §F8, §G1`)
- [x] **Test: `Stats.WatchBudgetCap == Stats.WatchLimit` when fraction is 1.0** —
      DONE 2026-10-07 (PR #46): fraction 1.0 is a documented no-op; 0.75
      scales; 0 is a no-op; default watcher cap == detected limit.
      (`src: 2026-08-12_19-22 §F9`)
- [x] **Test: `ErrorContext.Event` population** — DONE 2026-10-07 (PR #46):
      handler/middleware errors carry the event; registration errors carry
      nil event with populated path + retryable=true.
      (`src: 2026-10-07_02-24 §f13`)

## CI & Release Hardening

- [x] **CI guard: go-directive floor** — DONE 2026-10-07 (PR #38): `go-directive`
      CI job asserts `go 1.26.7` and gates the test matrix; `nix flake check`
      runs the same assertion as the `go-directive` check. Structural, not
      skip-list-fragile.
      (`src: 2026-10-07_02-24 §f8, 05-17 §f36, 06-00 §f4/C5`)
- [x] **Go 1.27 in the CI test matrix** — RESOLVED AS STRUCTURALLY BLOCKED
      2026-10-07 (PR #38): `encoding/json/v2` requires the `go` directive ≥ 1.27
      while the pinned floor is 1.26.7, so a 1.27 leg cannot build. Revisit only
      as part of an intentional floor bump (pairs with the fleet Go-strategy
      decision; see open questions and the v3 decision doc).
      (`src: 2026-10-07_02-24 §f9, 04-34 §f22`)
- [x] **Required status checks on master + release PRs** — DONE 2026-10-07: 8
      required contexts live (floor guard, tests on 24.04/26.04, examples,
      lint, commitlint, drift check, exported symbols) with linear history +
      enforce-admins; proven by six same-day PR merges, all gated.
      (`src: 2026-10-07_02-24 §f10/c7`)
- [x] **`release.yml` `workflow_dispatch` trigger** — DONE 2026-10-07 (PRs
      #38, #43): trigger added with `tag` input; dry-run dispatched twice on
      v2.4.2. First run exposed a missing `GOEXPERIMENT=jsonv2` (tests could
      not compile `encoding/json/v2`) — fixed; second run verified checkout of
      `inputs.tag`, tests, lint, and that the release step refuses an existing
      release (published v2.4.2 untouched).
      (`src: 2026-10-07_02-24 §f12/b3`)
- [ ] **BuildFlow binary upgrade + skip-list re-verify** — binary is `202b114`,
      BuildFlow HEAD moved past it; after upgrade re-run the pipeline and
      re-prove the three skips still resolve (provider renames silently
      unskip). Also triage the non-gating finding noise (9-tools-unavailable
      warning, go-auto-upgrade ×12, nix-checker ×8, nix-build-verify fate,
      vulnix CVEs) and file the two upstream BuildFlow issues once approved
      (see open questions).
      (`src: 2026-10-07_06-00 §f7/B2/B3/C1`)
- [ ] **Re-capture benchmark baseline post-v2.4.x** — `nix run .#bench-baseline`
      on a quiet machine (AGENTS.md bench discipline: no parallel load).
      (`src: 2026-10-07_02-24 §f16, 03-09 §f17`)
- [ ] **Post-release smoke-test script** — after each release-please merge, assert
      tag + GitHub Release exist and the changelog section matches the merge
      (done manually on v2.4.0/v2.4.1).
      (`src: 2026-10-07_03-10 §f23`)
- [ ] **Re-check the fleet cue-policy runner deviation** — `cue-vet` flags
      `ci.yml` (`ubuntu-24.04` + a `${{ matrix.os }}` leg) and `release.yml`
      (`ubuntu-24.04`) against the fleet's `ubuntu-latest` allowlist. The pins
      are deliberate for runner determinism through the `ubuntu-latest` 26.04
      rollout (complete 2026-11-19), so both findings are baselined in
      `.buildflow-baseline.json` rather than changed. After the rollout: either
      switch to `ubuntu-latest` (and drop the 26.04 matrix leg) or re-file the
      deviation with the cue-policy repo, then remove the baseline entries.
      (`src: 2026-11-06 buildflow findings gate`)
- [ ] **Add `actionlint` to pre-push verification** — catch workflow-expression
      and `runs-on` mistakes locally instead of after a push.
      (`src: 2026-10-07_15-51 §f35`)
- [ ] **Decide on a weekly `-race -count=2` CI run** — cost/benefit call for
      catching order-dependent flakes the every-push single run misses; two
      reports reached the same open question independently.
      (`src: 2026-10-07_13-09 §f24, 15-21 §f21, 15-51 §f26`)

## Documentation

- [ ] **Shrink docs-consistency exemption list** — the CI gate exempts 15
      symbols (11 phantom-type helpers + `WatcherStateFlags` +
      `WithWatchedIgnoreDirs` (deprecated) + 2 real gaps:
      `FilterGeneratedCodeFull` and `FilterGeneratedCodeWithFilter`). Document
      the two filters in FEATURES.md to reach zero real exemptions.
- [x] **Split AGENTS.md to carrying capacity** — DONE 2026-11-06: AGENTS.md is
      now 178 lines (gate: 220). The Release/CI gotchas table, the go.mod floor
      narrative, and the Known Issues section moved verbatim to
      [docs/release-ci-gotchas.md](./docs/release-ci-gotchas.md), with a
      one-line-per-gotcha index left behind. File Organization, gogenfilter
      notes, CI Workflows, and the Linter Cheat Sheet had moved earlier.
      (`src: 2026-10-07_04-34 §c1, 06-00 §f18/B5`)
- [ ] **Website JS conventions doc** — write `website/README.md` documenting the
      site's build (`nix run .#build`), the generated `/changelog` page, and the
      nested `.oxlintrc.json` conventions carved out of the root config. Also
      restores the file AGENTS.md used to link to.
      (`src: 2026-10-12_03-10 §f18`)
- [ ] **Website deploy runbook** — document install → build → deploy via flake
      apps, the `trailingSlash: false` quirk, cleanUrls, and "verify built
      slugs before publishing URLs" (the v2.4.0 release-notes 404 lesson) in
      AGENTS.md or a RELEASE.md.
      (`src: 2026-10-07_03-09 §f5/#30, 02-24 §e1/e5`)
- [ ] **Compiled `examples/otel`** — the README OTel span-adapter + stdouttrace
      pattern is documentation-only; nothing compiles it against the real OTel
      SDK, so it can rot silently. Add `examples/otel/` (span adapter +
      stdouttrace exporter + PrometheusCollector adapter) so `go build
      ./examples/...` CI-verified it. 2026-11-06: verified the README pattern
      compiles and runs against otel v1.47.0 (spans + 15 metric families), but
      also found two false doc claims (otel.go OTelSpan "satisfies trace.Span",
      metrics.go direct-MustRegister) — both since fixed in doc comments.
      (`src: 2026-11-06 OTEL support audit`)

## Website Maintenance

- [x] **Fix website dependabot alerts** — DONE 2026-11-06: `pnpm audit --fix=update`
      in `website/` cleared all 19 advisories (3 low / 10 moderate / 6 high) —
      `fast-uri` 3.1.9, `undici` 8.12.2, `vite` 8.3.3, `zod` 4.6.6 now resolved,
      `pnpm audit` reports "No known vulnerabilities found". The four patched
      versions were added to `minimumReleaseAgeExclude` (they are younger than
      the age window); `pnpm run build` verified green (14 pages) after the bump.
      (`src: 2026-07-29_14-06 §Dependabot`)
- [x] **Website CI build job** — DONE 2026-10-07 (PRs #52–#54):
      `website-build.yml` runs pnpm install (frozen lockfile) + astro build on
      website/** PRs and master pushes. Three fix-forward rounds taught the
      workflow gotchas now in crush-config lessons: full 40-char action SHAs,
      actions ignoring `defaults.run.working-directory` (pass
      `package_json_file`), and the job being non-required (verified on the PR
      before each merge anyway). html-validate still to wire in if wanted.
      (`src: 2026-10-07_03-09 §f2/#6, 02-24 §f24`)
- [ ] **Website link checker** — `html-validate` or an Astro dead-link plugin in
      the build; would have caught the migration-guide 404.
      (`src: 2026-10-07_03-09 §f3`)
- [ ] **Post-audit website chores** — `pnpm dedupe` after the 2026-10-07
      lockfile churn; review the 7 `minimumReleaseAgeExclude` entries once the
      patched versions age out (see open questions); drop the
      `postcss-selector-parser` override when expressive-code bumps
      postcss-nested; document the override's existence; redeploy after dep
      updates; spot-check sharp 0.35.5 rendering on an image-heavy page.
      (`src: 2026-10-07_06-00 §f5/#22/#23/#24/#36/#37/B4`)
- [ ] **Commit the pnpm `minimumReleaseAge` window** — only the excludes list is
      in `pnpm-workspace.yaml`; the actual age window lives in a machine-local
      global config nobody else can see.
      (`src: 2026-10-07_03-10 §f17`)
- [ ] **Website small items** — add `editLink`/`lastUpdated` to astro.config if
      missing; consider gitignoring `changelog.mdx` (committed build artifact);
      check the OG image renders on social cards.
      (`src: 2026-10-07_03-09 §f24/#25/#28`)
- [ ] **Public-presence polish** — add a 180x180 `apple-touch-icon` to the site
      head, redesign the logo SVG so it is distinctive rather than generic, test
      the site in Firefox and Safari, embed a website screenshot/GIF in
      README.md, and add a "Why not X?" comparison section against the popular
      alternatives (fsnotify, viper/fsnotify wrappers, watcher libraries). The
      sitemap itself is verified (`website/dist/sitemap-index.xml` is built).
      (`src: 2026-07-13_21-22 §f5/#23/#42, 2026-07-13_21-58 §f25/#27/#34/#47/#48`)

## Hygiene

- [x] **Review/merge dependabot PRs #31, #32; close stale #14.** — DONE
      2026-10-07: all three already CLOSED (unmerged); the website dep updates
      they carried landed via the 06:00 pnpm-audit wave. No open dependabot
      PRs remain.
      (`src: 2026-10-07_02-24 §f15/c4`)
- [ ] **Verify prettier's "3 fixed" daemon-committed files** from the 06:00
      buildflow run 1 (unreviewed auto-edits).
      (`src: 2026-10-07_06-00 §f31`)
- [x] **Run gitleaks + codespell once** — DONE 2026-10-07 (PR #49): gitleaks
      clean (working tree + 586-commit history); codespell found exactly two
      real typos (fixed), the rest deliberate (café test data, WRONLY syscall
      constant, "accreting"); `.codespellrc` captures the false positives so a
      plain run exits clean. codespell not yet in the buildflow pipeline.
      (`src: 2026-10-07_06-00 §f32`)
- [x] **Clean /tmp evidence logs** — DONE 2026-10-07 15:10 (resume session):
      120 files trashed via `trash` (recoverable, rm-ban respected), 0
      remaining; shared caches (e.g. `npm-cache-buildflow/`) untouched. Scope
      was pattern-matched (`-user lars`, today), not per-file ownership-proof
      — flagged as a judgment call in the 15:21 report §d.
      (`src: 2026-10-07_06-00 §f34`)

## Process debts (2026-10-07 docs sweep tail)

- [ ] **Split the 20 oversized Go files (file-size campaign)** — `file-size-check`
      errors on 20 files over the 350-line budget; they are baselined in
      `.buildflow-baseline.json` so new overages still fail. Work largest first:
      `watcher_test.go` (1860), `middleware_test.go` (1406), `middleware.go`
      (1001), `watcher.go` (919), `filter_test.go` (853), `benchmark_test.go`
      (727), `filesystem_test.go` (642), `watcher_coverage_test.go` (597),
      `errors_test.go` (563), `filter.go` (543), `error_simulation_test.go`
      (537), `testing_helpers_test.go` (518), `options.go` (466),
      `filesystem_poll_gitignore_test.go` (457), `example_test.go` (447),
      `filter_gogen_test.go` (419), `watcher_internal.go` (408),
      `watcher_walk_test.go` (378), `options_test.go` (369), `errors.go` (364).
      Beware: `max_file_size` is INERT in BuildFlow (BuildFlow#75) so raising the
      limit is not an option. Drop each baseline entry as its file splits.
      (`src: 2026-11-06 buildflow findings gate`)
- [ ] **Reduce the docs-annotate backlog (ratchet at 1786)** — `docs-gates
      annotate` reports 1786 findings (struck items with no verdict marker, or a
      verdict citing a path that no longer exists) across 105 files, mostly in
      `docs/status/archived/`. `.docs-gates.json` pins the count so the gate is
      red only past it; annotate a batch and ratchet down.
      (`src: 2026-11-06 docs-gates battery`)
- [ ] **Run the report archive ritual (17 unarchived, oldest 30 days)** —
      `docs-gates cadence` is ratcheted at 17 reports / 30 days. Walk the
      2026-10-07 batch: verify every remaining item is struck, WONT-DO, or
      routed, `git mv` into `docs/status/archived/` with the `> **Live
      remainder:**` banner, update `docs/status/README.md`, then tighten the
      cadence ratchet back to the fleet defaults (15 / 21).
      (`src: 2026-11-06 docs-gates battery`)
- [ ] **Add a scheduled `BUILDFLOW_NO_RESULT_CACHE=1` run as a cache-lie
      tripwire** — result-cache keying gaps (pnpm-lock, oxlint) have twice hidden
      real findings behind a stale zero-finding entry.
      (`src: 2026-10-12_03-10 §f14`)
- [ ] **Settle the `reliability_semantics_test.go` placement convention** — decide
      whether semantics tests live in their own file (as now) or next to the code
      they pin, and record the rule.
      (`src: 2026-10-07_15-51 §f43`)
- [ ] **Adopt a post-merge master-CI confirmation habit for docs-only PRs** —
      check the merge commit's run before moving on; release-please PRs report a
      zero-job failure that is a GitHub quirk, so master CI is the only real
      signal.
      (`src: 2026-10-07_15-21 §f46, 15-51 §f46`)
- [ ] **Decide the website CSS direction (still open since 2026-10-07 15:51)** —
      the Starlight theme layer has both inline `style` shims and custom CSS;
      pick one home so future styling does not split across both.
      (`src: 2026-10-07_16-30 §f20`)
- [ ] **Audit the 292 marker-free tables in archived files** — check-rows now
      passes (86/86 uniform), but tables with zero `~~` anywhere were left as
      non-task data tables or potentially-open items; one bounded pass should
      classify them (data vs open task).
      (`src: 2026-10-07 docs/sweep-followups pass`)
- [ ] **Overnight merge-ownership protocol for release PRs** — define who/what may
      merge a release PR while other sessions push to master (v2.4.1 race window
      was seconds-wide).
      (`src: 2026-10-07_03-10 §f32`)
- [ ] **Consumer sweep** — check no dependent repos' CI keyed off red-master
      windows (the go.mod incidents).
      (`src: 2026-10-07_03-10 §f39`)
- [x] **Content-verify DOMAIN_LANGUAGE / API_STABILITY / Troubleshooting /
      MIGRATION / ARCHITECTURE against code** — DONE 2026-10-07: API_STABILITY
      verified (all three deprecated symbols exist with `// Deprecated:`
      markers); DOMAIN_LANGUAGE verified (Op/enum/option terms all in code).
      FINDINGS: `Troubleshooting.md`, `MIGRATION.md`, `ARCHITECTURE.md` never
      existed — README's related-docs row linked two of them (dead links,
      fixed: migration now points at the live website guide, verified URL);
      ARCHITECTURE.md is only referenced inside historical CHANGELOG entries
      (left as history). Troubleshooting content remains unbuilt (ROADMAP
      mentions a guide that has no page).
      (`src: 2026-10-07_08-14 §c2`)
- [ ] **Starlight editLink + lastUpdated** — neither configured in
      website/astro.config.mjs; both are Starlight built-ins (`editLink.baseUrl`,
      `lastUpdated: true`). Decide whether stale lastUpdated dates on old pages
      are acceptable before enabling.
      (`src: 2026-10-07_02-24 §f14`)
- [ ] **Full `nix flake check` + verify second vendorHash (flake.nix:111)**.
      (`src: 2026-10-07_06-00 §B6`)
- [ ] **Status-dir index** (`docs/status/README.md`: active reports + archived
      pointer) + CHANGELOG v2.4.0 cosmetic blank-line fix.
      (`src: 2026-10-07_08-14 §c5/c6`)
- [ ] **Cross-repo (crush-config):** commit `lessons.md` entry (instrument
      writers before racing them) + fix buildflow skill DB path
      (`buildflow.db`→`cache.db`) with fan-out verify.
      (`src: 2026-10-07_05-17 §f10, 06-00 §f9`)
- [ ] **v3 decision doc** — collect Q1/Q2/Q4/Q9 + micro-polish bundle into
      `docs/research/v3-decisions.md`.
      (`src: 2026-10-07_08-17 plan task 26`)

## v3 Candidates

- [ ] **Phantom-typed `PathKey`** — `type PathKey string` for compile-time safety
      on path keys. Currently keys are bare strings. Would prevent accidental
      mixing of raw paths and canonicalized keys.
- [ ] **`CaseSensitivityProbed` mode** — actually probe the filesystem (write a
      test file with mixed-case name, check if it collides) instead of relying
      on `runtime.GOOS`. More accurate for edge cases (case-sensitive NTFS,
      case-sensitive APFS volume).
      (`src: 2026-07-29_09-44 §f #27`)
- [ ] **Filesystem API micro-polish** — `Parse(string)` for
      `FilesystemCaseSensitivity`, a `CaseSensitivityConfigured` stat (vs the
      resolved-mode stat), a lock-policy doc note for
      `EffectiveCaseSensitivity`, a symlink-to-watched-path dedup test, and an
      `EffectiveCaseSensitivity`-vs-`Reset()` concurrency test.
      (`src: 2026-07-29_14-38 §f4/5/10/13/17`)
- [ ] **Context-aware `OTelMiddleware`** — `startSpan func(path, op string)`
      carries no `context.Context`, so every event span is a ROOT trace:
      verified 2026-11-06 (two events → two distinct TraceIDs, parent all
      zeros). Downstream apps cannot correlate file events with their own
      request traces. v3 shape: a context-aware startSpan variant (or thread
      the pipeline ctx into startSpan) with the old signature kept as a
      wrapper. Root-span workaround until then: close over a captured ctx.
      (`src: 2026-11-06 OTEL support audit`)

---

## Status Snapshot

| Metric         | Value | Status |
| -------------- | ----- | ------ |
| Linter issues  | 0     | ✅     |
| Build          | Clean | ✅     |
| Tests          | 100%  | ✅     |
| Flaky tests    | 0     | ✅     |
| Broken benches | 0     | ✅     |
| Open items     | 39    | 🟡     |

Count note 2026-10-07: six items ticked today (CI guard, required checks,
workflow_dispatch, probe legs, reliability test trio, codespell), three new
items added (leg-promotion decision, per-OS expectations, follow-up items
from the probe), net 39 → 40; the /tmp evidence-log cleanup ticked in the
15:21 resume session, net 40 → 39.

---

## Session backlog (2026-11-06 buildflow-green run)

Harvested from `docs/status/2026-11-06_07-22_buildflow-green-lint-fix-and-docs-ratchet.md`
§f. Duplicates of the campaigns above were dropped; these are the distinct ones.
(`src: 2026-11-06_07-22 §f`)

- [ ] **Split `watcher_test.go` (1860 lines) into topic files** — P1/L/Quality.
- [ ] **Split `middleware.go` (1001 lines) into themed middleware files** — P1/L/Quality.
- [ ] **Split `middleware_test.go` (1406 lines)** — P2/L/Quality.
- [ ] **Split `watcher.go` (919 lines) along its walk/event/health seams** — P2/L/Quality.
- [ ] **Split the remaining 16 oversized files and drop their baseline entries** — P3/L/Quality.
- [ ] **Annotate a first 100-item batch of the 1786 docs-annotate findings and
      lower the ratchet** — P1/M/Cleanup.
- [ ] **Archive the 2026-10-07 report batch (13 files) and lower the cadence
      ratchet to 15/21** — P1/M/Cleanup.
- [ ] **Identify the 9 "tools unavailable (health check failed)" providers and
      fix or exclude them** — P1/S/Quality.
- [ ] **Upgrade the BuildFlow binary past `e81c118` and re-prove the three
      skip_steps still match the current provider names** — P1/S/Quality.
- [ ] **Verify this session's changes on CI, including the Windows and macOS
      probe legs** — P1/S/Quality.
- [ ] **Decide whether test files are exempt from the 350-line budget** (and ask
      BuildFlow upstream for a per-glob limit) — P1/S/Decision.
- [ ] **Write `website/README.md`** (build, generated `/changelog`, oxlint
      conventions) — P2/S/Docs.
- [ ] **Add `actionlint` to pre-push verification** — P2/S/Quality.
- [ ] **Decide on a weekly `-race -count=2` CI run** — P2/S/Decision.
- [ ] **Add a scheduled `BUILDFLOW_NO_RESULT_CACHE=1` run as a cache-lie
      tripwire** — P2/S/Quality.
- [ ] **Automate the pnpm-audit/oxlint result-cache purge** (script or upstream
      key fix) — P2/M/Quality.
- [ ] **Re-verify the documented BuildFlow limitations** (`nix-hash-fix`,
      pnpm-audit cache, `max_file_size`) against the current binary — P2/M/Quality.
- [ ] **Confirm the dependabot `nix` and root-`npm` entries actually manage
      something** — P2/S/Docs.
- [ ] **Re-check the cue-vet runner deviation after the 2026-11-19 rollout and
      delete both baseline entries** — P1/S/CI.
- [ ] **Triage the `vulnix` advisories** (accept or mitigate per CVE) and refresh
      the stale `result` scan target — P3/S/Security.
- [ ] **Settle the `reliability_semantics_test.go` placement convention** —
      P3/S/Quality.
- [ ] **Adopt the post-merge master-CI confirmation habit for docs-only PRs** —
      P3/S/Process.
- [ ] **Decide the website CSS direction** (Starlight theme vs custom layer) —
      P3/S/Decision.
- [ ] **Public-presence polish**: apple-touch-icon, logo redesign, Firefox/Safari
      pass, README screenshot/GIF, "Why not X?" — P3/M/Marketing.
- [ ] **Re-drive seed `1793946570936674137` under concurrent load, name the
      flaking test, and harden it** — the 07:29 `test-race` failure happened only
      while a second BuildFlow process held the state DB (`SQLITE_BUSY`); four
      subsequent runs with the same seed were green. — P1/M/Quality.
- [ ] **Document the health API on the website** — `Healthy` appears in
      README/FEATURES/API_STABILITY but not in `website/src/content/docs`
      (`api-reference.mdx` only mentions `MiddlewareCircuitBreaker`), so the
      feature shipped in v2.5.0 is invisible on filewatcher.lars.software. Add a
      health section/page, rebuild, deploy. — P1/M/Docs.
- [x] **Document `GOEXPERIMENT=jsonv2` in the README Install section** — DONE
      2026-11-06: validated with a throwaway consumer module that a build
      without the experiment fails with "build constraints exclude all Go files
      in …/encoding/json/v2"; README now states the requirement and the export
      line. — P1/S/Docs.
- [ ] **Add a CI check that `.release-please-manifest.json` matches the newest
      `v*` tag** — v2.4.3 shipped by hand and left the manifest at 2.4.2, which
      would have produced a wrong `v2.4.2...v2.5.0` compare link. — P2/S/CI.

---

## Open questions (blockers, not tasks — need user decision)

These are **not** actionable until answered. They do not belong in the
checklist above.

1. **Is `.goreleaser.yml` dead config to delete, or an unfinished wiring task?**
   `release.yml` uses `softprops/action-gh-release` with auto-generated notes and
   has no goreleaser step. Now that release-please is wired in, goreleaser may be
   fully obsolete. Deleting it simplifies FEATURES (Cross-platform releases →
   not-planned); keeping it means it should be wired in eventually.
   (`src: 2026-07-26_18-39 §g Q2`)

2. **Should the benchmark baseline be committed (CI-enforceable) or stay
   gitignored (local-only)?** The baseline is currently gitignored per the
   original TODO, but the ROADMAP says "benchmark freshness CI." A gitignored
   file can never drive a CI regression gate. Machine-specific noise makes
   committed baselines imperfect, but they are the only way CI can catch
   regressions.
   (`src: 2026-07-26_21-00 §g Q1`)

3. **The `WatchChanges(ctx, targetState)` open design questions** (reporting
   granularity, closed-watcher semantics, depth reconciliation) live in
   `docs/research/watchchanges-contract.md`; the implementation becomes a TODO
   once those are answered.

4. **Should the `watchBackend` interface be exported?** Currently unexported
   (test-internal only). Consumers might want their own fake backends for
   integration testing. Tradeoff: more public API surface vs. more consumer
   value. This is a v3 decision — exporting it is a breaking commitment.
   (`src: 2026-07-26_20-00 §G Q1`)

5. **Should website deploys be automated from CI** (Firebase service-account
   secret on master push / workflow_dispatch), or stay manual-and-local?
   Determines the website CI job's scope.
   (`src: 2026-10-07_03-09 §g Q2`)

6. **File the two upstream BuildFlow issues?** (a) go-structure-linter
   auto-bumps the `go` directive mid-run (strace evidence ready); (b) result
   cache ignores `pnpm-lock.yaml` for pnpm-audit (stale replay until TTL).
   Both are fleet-wide hazards; filings need your go-ahead since you own
   BuildFlow.
   (`src: 2026-10-07_05-17 §g Q1, 06-00 §g Q1/C1`)

7. **go-daemon ecosystem follow-ups** — mirror the "Related Projects" note in
   go-daemon's README (cross-repo edit), and choose the host for a
   `watcher-daemon` example: bump this module to Go 1.27, host in go-daemon's
   repo, or a sibling repo.
   (`src: 2026-10-07_00-42 §b1/g1/g2`)

8. **Are the 7 new `minimumReleaseAgeExclude` entries an acceptable standing
   supply-chain tradeoff** (trusting same-day patched releases), and when
   should they be pruned?
   (`src: 2026-10-07_06-00 §g Q3/B4`) Also: Dependabot can propose versions
   younger than the window (bot cadence vs supply-chain policy mismatch —
   devalue 5.9.4 merged within hours of existing).
   (`src: 2026-10-07_03-10 §f27`)

9. **Changelog policy for releases** — should future releases block tagging on
   curated changelog content (curate inside the release PR before merge), or is
   the tag-gets-stub/master-gets-curation split acceptable (as shipped for
   v2.4.0)?

10. **Keep or retract v2.4.1?** It shipped as a Go patch release whose only
    change is website CSS (pre-path-scoping incident). Published and
    pkg.go.dev-indexed; retraction means deleting the release + tag. The
    policy call is yours. (`src: 2026-10-07_03-10 §g1`)

11. **May the auto-commit daemon ignore dependency manifests** (`go.mod`,
    `go.sum`, lockfiles)? The daemon committed a `go 1.27` bump once (the
    war class). A daemon-side ignore list is implementable; whether the
    daemon should ever commit manifests is your decision.
    (`src: 2026-10-07_03-10 §g2`)
    (`src: 2026-10-07_02-24 §g Q3`)
