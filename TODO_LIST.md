# TODO List

**Last Updated:** 2026-11-06 (docs-health audit: docs-gates now passes with zero
tolerated findings — annotate 1786 → 0, audit-all 16 → 0, 15 resolved reports
archived, this list deduplicated so every item appears exactly once)

Short- and mid-term actionable work. Each item is scoped — pick one, do it, tick
the box. An item lives here only when it is bounded and estimable; vague or
long-term ideas live in [ROADMAP.md](./ROADMAP.md). Completed work is recorded in
[CHANGELOG.md](./CHANGELOG.md), never here.

---

## Next up

- [ ] **Ubuntu 26 runner re-evaluation — after the 2026-11-19 rollout** — every
      workflow still pins `ubuntu-24.04` (plus a 24.04 + 26.04 test matrix) for
      runner determinism; the original 2026-10-12 audit deadline passed with the
      pins intact. When the `ubuntu-latest` rollout completes: switch the jobs to
      `ubuntu-latest` (dropping the 26.04 matrix leg) or re-file the deviation
      with the cue-policy repo, then delete the two `cue-vet` entries from
      `.buildflow-baseline.json`.
      (`src: 2026-10-07_02-24 §f11, 2026-11-06 buildflow findings gate`)
- [ ] **Verify the 2026-11-06 session's changes on CI** — the health API, docs
      ratchet, and lint fixes were proven locally only; confirm the push runs
      green, including the Windows and macOS probe legs.
      (`src: 2026-11-06_07-22 §f10`)
- [ ] **BuildFlow binary upgrade + skip-list re-verify** — the binary is older
      than BuildFlow HEAD; after upgrading, re-prove the `skip_steps` still
      resolve (provider renames silently unskip) and triage the non-gating
      finding noise (9-tools-unavailable warning, `go-auto-upgrade` ×12,
      `nix-checker` ×8, `nix-build-verify` fate, `vulnix` CVEs). The two
      upstream BuildFlow filings stay gated on open question 6.
      (`src: 2026-10-07_06-00 §f7/B2/B3/C1, 2026-11-06_07-22 §f9`)

## Testing & Platform

- [ ] **Promote the windows/macos probe legs to required checks** — decision
      point after the legs bake: promote and document why, or keep them
      informational. Promotion prerequisite: per-OS expectations replacing the
      `skipOnWindows` skips.
      (`src: 2026-10-07_02-24 §f8, 15-51 §f9`)
- [ ] **Per-OS test expectations for Windows** — the `skipOnWindows(t)` skips
      hide correct-per-OS behavior behind POSIX-shaped assertions
      (`TestCleanPath`, `TestNormalizePath_EdgeCases`, `TestEventPath_Join/Dir`,
      `TestFilterExcludePaths`, `TestFilterIgnoreDirs`/`Hidden` + case
      variants). Rewrite with `filepath`-based expectations so Windows runs the
      assertions too.
      (`src: PR #47 probe run 37602089069, 15-51 §f10`)
- [ ] **Expand fuzz tests** — the corpus covers `FilterRegex`,
      `FilterExtensions`, `FilterIgnoreGlobs`, `OpUnmarshalText`, `FilterMinSize`,
      and `PathKey` (NFC/idempotency/case-folding). Add fuzzers for
      `FilterAnd`/`FilterOr`/`FilterNot` composition, `Event` JSON round-trip,
      the gitignore matcher, `FilterIgnoreDirsCaseInsensitive`, and
      `MiddlewareDeduplicateCaseInsensitive`.
      (`src: 2026-08-12_19-22 §f16`)
- [ ] **Large-tree stress harness** — synthetic 100k-directory fixture that
      validates batched registration, budget enforcement, and self-heal under
      load. Routed here from four reports; still the biggest untested claim.
- [ ] **Investigate the July 6-hour Benchmark hang** — PR #14's Benchmark job ran
      02:23→08:24 UTC (6h, killed at the runner timeout): possible
      benchmark/watcher deadlock. Timeout caps bound the blast radius
      (`b2caf11`); root cause still unknown.
      (`src: 2026-10-07_03-10 §b5/f4`)
- [ ] **Re-drive seed `1793946570936674137` under concurrent load** — the 07:29
      `test-race` failure happened only while a second BuildFlow process held
      the state DB (`SQLITE_BUSY`); four subsequent runs with the same seed were
      green. Reproduce under load, name the flaking test, and harden it.
      (`src: 2026-11-06_07-22 §f25`)

## CI & Release Hardening

- [ ] **Re-capture the benchmark baseline post-v2.4.x** — `nix run .#bench-baseline`
      on a quiet machine (AGENTS.md bench discipline: no parallel load).
      (`src: 2026-10-07_02-24 §f16, 03-09 §f17`)
- [ ] **Post-release smoke-test script** — after each release-please merge, assert
      the tag + GitHub Release exist and the changelog section matches the merge
      (done by hand for v2.4.0/v2.4.1/v2.5.0).
      (`src: 2026-10-07_03-10 §f23, 15-51 §f13`)
- [ ] **Add `actionlint` to pre-push verification** — catch workflow-expression
      and `runs-on` mistakes locally instead of after a push.
      (`src: 2026-10-07_15-51 §f35`)
- [ ] **Decide on a weekly `-race -count=2` CI run** — cost/benefit call for
      catching order-dependent flakes that the every-push single run misses.
      (`src: 2026-10-07_13-09 §f24, 15-21 §f21, 15-51 §f26`)
- [ ] **Add a CI check that `.release-please-manifest.json` matches the newest
      `v*` tag** — v2.4.3 shipped by hand and left the manifest at 2.4.2, which
      would have produced a wrong `v2.4.2...v2.5.0` compare link.
      (`src: 2026-11-06_07-22 §f28`)
- [ ] **Generate the GitHub Release body from the CHANGELOG section at tag time**
      so release notes have one source (today: release-please output plus manual
      edits).
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Add `docs/release-checklist.md`** — prerequisites (GOEXPERIMENT, Go
      floor), the CI gate before tagging, then proxy + pkg.go.dev + site
      verification.
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Automate the pnpm-audit/oxlint result-cache purge** (script or upstream
      key fix) — stale cache entries have twice hidden real findings.
      (`src: 2026-11-06_07-22 §f16, 2026-10-07_06-00 §f33`)
- [ ] **Re-verify the documented BuildFlow limitations** — `nix-hash-fix`,
      pnpm-audit cache replay, and `max_file_size` (inert in BuildFlow#75)
      against the current binary.
      (`src: 2026-11-06_07-22 §f17`)
- [ ] **Add a scheduled `BUILDFLOW_NO_RESULT_CACHE=1` run as a cache-lie
      tripwire** — cache-keying gaps (pnpm-lock, oxlint) have twice hidden real
      findings behind a stale zero-finding entry.
      (`src: 2026-10-12_03-10 §f14`)
- [ ] **Identify the 9 "tools unavailable (health check failed)" providers** and
      fix or exclude them.
      (`src: 2026-11-06_07-22 §f8`)
- [ ] **Confirm the dependabot `nix` and root-`npm` entries manage something** —
      only `npm_and_yarn` and `gomod` PRs have been observed.
      (`src: 2026-11-06_07-22 §f18`)
- [ ] **Triage the `vulnix` advisories** (accept or mitigate per CVE) and refresh
      the stale `result` scan target.
      (`src: 2026-11-06_07-22 §f20`)
- [ ] **Confirm the release-commit CI run is green** — the v2.5.0 merge
      (`b77b280`) and the follow-up docs commit (`2bd47ff`); fix anything it
      flagged.
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Verify the Windows/macOS probe legs ran on the release commit.**
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Confirm the release did not silently move a consumer-visible dependency
      floor** (`go.yaml.in/yaml/v3` went 3.0.5 → 3.0.4 indirect across
      v2.4.3 → v2.5.0).
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Re-check pkg.go.dev shows `v2.5.0`** and record the confirmation (it
      still listed 2.4.3 at release time; the proxy already serves 2.5.0).
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Rewrite the published v2.4.3 release notes** — GitHub's auto-generated
      body lists PRs #43–#60, most of which shipped in 2.4.2.
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Bump a real consumer to `v2.5.0` and run its tests** (e.g. a `go-health`
      example) — release proof currently rests on a throwaway `/tmp` module.
      (`src: 2026-11-06_07-22 §f` release cluster)
- [ ] **Decide whether test files are exempt from the 350-line budget** — if
      they are, ask BuildFlow upstream for a per-glob limit instead of inflating
      the number (see the file-size campaign below).
      (`src: 2026-11-06_07-22 §f11`)

## Documentation

- [ ] **Shrink the docs-consistency exemption list** — the CI gate exempts 15
      symbols (11 phantom-type helpers + `WatcherStateFlags` +
      `WithWatchedIgnoreDirs` (deprecated) + 2 real gaps:
      `FilterGeneratedCodeFull` and `FilterGeneratedCodeWithFilter`). Document
      the two filters in FEATURES.md to reach zero real exemptions.
- [ ] **Website JS conventions doc** — write `website/README.md` documenting the
      site's build (`nix run .#build`), the generated `/changelog` page, and the
      nested `.oxlintrc.json` conventions carved out of the root config. Also
      restores the file AGENTS.md used to link to.
      (`src: 2026-10-12_03-10 §f18, 2026-11-06_07-22 §f12`)
- [ ] **Website deploy runbook** — document install → build → deploy via flake
      apps, the `trailingSlash: false` quirk, cleanUrls, and "verify built slugs
      before publishing URLs" (the v2.4.0 release-notes 404 lesson).
      (`src: 2026-10-07_03-09 §f5/#30, 02-24 §e1/e5`)
- [ ] **Compiled `examples/otel`** — the README OTel span-adapter + stdouttrace
      pattern is documentation-only; nothing compiles it against the real OTel
      SDK, so it can rot silently. Add `examples/otel/` (span adapter +
      stdouttrace exporter + PrometheusCollector adapter) so
      `go build ./examples/...` in CI verifies it. 2026-11-06: the README
      pattern was proven against otel v1.47.0 (spans + 15 metric families) and
      two false doc claims were fixed.
      (`src: 2026-11-06 OTEL support audit`)
- [ ] **Document the health API on the website** — `Healthy` appears in
      README/FEATURES/API_STABILITY but not in `website/src/content/docs`
      (`api-reference.mdx` only mentions `MiddlewareCircuitBreaker`), so the
      v2.5.0 feature is invisible on filewatcher.lars.software. Add a health
      section/page, rebuild, deploy.
      (`src: 2026-11-06_07-22 §f26`)
- [ ] **Health API graduate-to-Stable note** — record in `API_STABILITY.md` the
      soak policy that governs `Healthy()` / `NewCircuitBreaker()` leaving
      Experimental (how long, which consumers must adopt it, what signals count),
      so the graduation is a rule rather than a later judgement call.
      (`src: 2026-11-06_06-46 §f27`)
- [ ] **Document the `MiddlewareCircuitBreaker` → `NewCircuitBreaker` migration**
      in the website middleware guide.
      (`src: 2026-11-06_07-22 §f` docs cluster)
- [ ] **Add a `Healthy`-in-action example under `examples/`** so
      `go build ./examples/...` compiles the health wiring.
      (`src: 2026-11-06_07-22 §f` docs cluster)
- [ ] **Verify or footnote the "Go 1.27 defaults to json/v2" claim** in README
      and the v2.5.0 release notes; it is currently an unsourced expectation.
      (`src: 2026-11-06_07-22 §f` docs cluster)
- [ ] **Cross-repo (crush-config):** commit the `lessons.md` entry (instrument
      writers before racing them) and fix the buildflow skill DB path
      (`buildflow.db` → `cache.db`) with fan-out verify.
      (`src: 2026-10-07_05-17 §f10, 06-00 §f9`)
- [ ] **v3 decision doc** — collect open questions 1/2/4/9 plus the micro-polish
      bundle into `docs/research/v3-decisions.md`.
      (`src: 2026-10-07_08-17 plan task 26`)
- [ ] **Add a "peer agent active" check to the session-start ritual** on shared
      machines (`ps` for peer sessions; treat unexpected worktree diffs as
      someone else's).
      (`src: 2026-11-06_07-22 §f` process cluster)
- [ ] **Audit the remaining marker-free tables in archived files** — the
      cell-boundary sweep made every struck row uniform, but tables with zero
      `~~` anywhere were never classified as data-vs-open-task. One bounded pass
      closes it.
      (`src: 2026-10-07 docs/sweep-followups pass`)

## Website

- [ ] **Website link checker** — `html-validate` (devDep already present) or an
      Astro dead-link plugin in the build; would have caught the
      migration-guide 404.
      (`src: 2026-10-07_03-09 §f3, 16-30 §f14, 15-51 §f15`)
- [ ] **Record a Lighthouse audit of the built site** — run Lighthouse against
      `https://filewatcher.lars.software` (or the built `website/dist`), record
      the scores, and fix critical SEO/a11y findings.
      (`src: 2026-07-13_21-22 §f34, 2026-07-13_21-58 §f35`)
- [ ] **Post-audit website chores** — `pnpm dedupe` after the lockfile churn;
      review the 7 `minimumReleaseAgeExclude` entries once the patched versions
      age out; drop the `postcss-selector-parser` override when expressive-code
      bumps postcss-nested; document the override; redeploy after dep updates;
      spot-check sharp 0.35.5 rendering on an image-heavy page.
      (`src: 2026-10-07_06-00 §f5/#22/#23/#24/#36/#37/B4`)
- [ ] **Commit the pnpm `minimumReleaseAge` window** — only the excludes list is
      in `pnpm-workspace.yaml`; the actual age window lives in a machine-local
      global config nobody else can see.
      (`src: 2026-10-07_03-10 §f17, 15-51 §f30`)
- [ ] **Starlight `editLink` + `lastUpdated`** — neither is configured in
      `website/astro.config.mjs`, though both are built-ins
      (`editLink.baseUrl`, `lastUpdated: true`). Decide whether stale
      `lastUpdated` dates on old pages are acceptable before enabling. Same
      pass: consider gitignoring the generated `changelog.mdx` and check the OG
      image renders on social cards.
      (`src: 2026-10-07_02-24 §f14, 03-09 §f24/#25/#28, 15-51 §f16`)
- [ ] **Decide the website CSS direction (open since 2026-10-07 15:51)** — the
      Starlight theme layer has both inline `style` shims and custom CSS; pick
      one home so future styling does not split across both.
      (`src: 2026-10-07_16-30 §f20, 15-51 §f5, 2026-11-06_07-22 §f23`)
- [ ] **Public-presence polish** — a 180x180 `apple-touch-icon` in the site head,
      a distinctive logo SVG, a Firefox/Safari pass, a website screenshot/GIF in
      README.md, and a "Why not X?" section comparing popular alternatives
      (fsnotify itself and the watcher wrappers) so the split brain with the
      README comparison table dies. The sitemap itself is verified
      (`website/dist/sitemap-index.xml` is built).
      (`src: 2026-07-13_21-22 §f5/#23/#42, 2026-07-13_21-58 §f25/#27/#34/#47/#48`)
- [ ] **Rebuild and deploy the website so `/changelog` shows 2.4.3 + 2.5.0** —
      the site is deployed by hand, so releases stay invisible there until
      someone runs the runbook.
      (`src: 2026-11-06_07-22 §f` website cluster)

## Hygiene & Process debts

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
      Beware: `max_file_size` is INERT in BuildFlow (BuildFlow#75). Drop each
      baseline entry as its file splits.
      (`src: 2026-11-06 buildflow findings gate, 2026-11-06_07-22 §f1–f5`)
- [ ] **Settle the `reliability_semantics_test.go` placement convention** —
      decide whether semantics tests live in their own file (as now) or next to
      the code they pin, and record the rule.
      (`src: 2026-10-07_15-51 §f43, 2026-11-06_07-22 §f21`)
- [ ] **Adopt a post-merge master-CI confirmation habit for docs-only PRs** —
      check the merge commit's run before moving on; release-please PRs report a
      zero-job failure that is a GitHub quirk, so master CI is the only real
      signal.
      (`src: 2026-10-07_15-21 §f46, 15-51 §f46, 2026-11-06_07-22 §f22`)
- [ ] **Overnight merge-ownership protocol for release PRs** — define who/what
      may merge a release PR while other sessions push to master (the v2.4.1
      race window was seconds-wide).
      (`src: 2026-10-07_03-10 §f32`)
- [ ] **Consumer sweep** — check that no dependent repo's CI keyed off the
      red-master windows from the `go.mod` incidents.
      (`src: 2026-10-07_03-10 §f39`)
- [ ] **Verify prettier's "3 fixed" daemon-committed files** from the 2026-10-07
      06:00 buildflow run (unreviewed auto-edits) — confirm the formatter gates
      now cover them or drop the concern with evidence.
      (`src: 2026-10-07_06-00 §f31`)

## v3 Candidates

- [ ] **Phantom-typed `PathKey`** — `type PathKey string` for compile-time safety
      on path keys. Keys are bare strings today, so raw paths and canonicalized
      keys can be mixed silently.
- [ ] **`CaseSensitivityProbed` mode** — actually probe the filesystem (write a
      test file with a mixed-case name, check whether it collides) instead of
      relying on `runtime.GOOS`. More accurate for edge cases (case-sensitive
      NTFS, case-sensitive APFS volume).
      (`src: 2026-07-29_09-44 §f #27`)
- [ ] **Filesystem API micro-polish** — `Parse(string)` for
      `FilesystemCaseSensitivity`, a `CaseSensitivityConfigured` stat (vs the
      resolved-mode stat), a lock-policy doc note for
      `EffectiveCaseSensitivity`, a symlink-to-watched-path dedup test, and an
      `EffectiveCaseSensitivity`-vs-`Reset()` concurrency test.
      (`src: 2026-07-29_14-38 §f4/5/10/13/17`)
- [ ] **Context-aware `OTelMiddleware`** — `startSpan func(path, op string)`
      carries no `context.Context`, so every event span is a ROOT trace
      (verified 2026-11-06: two events → two distinct TraceIDs, parent all
      zeros). Downstream apps cannot correlate file events with their own
      request traces. v3 shape: a context-aware `startSpan` variant (or thread
      the pipeline ctx into `startSpan`) with the old signature kept as a
      wrapper. Root-span workaround until then: close over a captured ctx.
      (`src: 2026-11-06 OTEL support audit`)

---

## Status Snapshot

| Metric             | Value          | Status |
| ------------------ | -------------- | ------ |
| Linter issues      | 0              | ✅     |
| Build              | Clean          | ✅     |
| Test suite         | 100% passing   | ✅     |
| Coverage (CI gate) | ≥ 85% enforced | ✅     |
| Flaky tests known  | 0              | ✅     |
| Broken benches     | 0              | ✅     |
| Open items         | 60             | 🟡     |
| docs-gates         | 5/5 PASS       | ✅     |

`docs-gates run` is the standing documentation gate: freshness, audit-all,
annotate, routing, and cadence all pass with **zero** tolerated findings and no
ratchet slack (`.docs-gates.json`). A red gate means a report item was not
harvested, a struck item carries no verdict, or a report outlived its cadence.

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

8. **Are the 7 `minimumReleaseAgeExclude` entries an acceptable standing
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
    (`src: 2026-10-07_03-10 §g2, 02-24 §g Q3`)
