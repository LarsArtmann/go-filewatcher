# TODO List Clearance — All 15 Items Complete

> **ARCHIVED 2026-10-07** (docs-health sweep): forward items resolved inline below — `~~strikethrough~~` verdicts cite evidence. Live work lives in [TODO_LIST.md](../../TODO_LIST.md).

**Date:** 2026-07-26_22-01
**Scope:** All HIGH (5) + MEDIUM (10) priority items from the 2026-07-26 TODO harvest.

## Summary

Every item from the pasted TODO list is implemented, tested, lint-clean, and documented.

### HIGH Priority (5/5 ✅)

| #  | Item | Resolution                                 |
| -- | ---- | ------------------------------------------ |
| ~~ | ~~1~~ | ~~Mark deprecations in README.md~~ |
| ~~ | ~~2~~ | ~~v2.3→v3 migration section in MIGRATION.md~~ |
| ~~ | ~~3~~ | ~~Deprecation badges in api-reference.mdx~~ |
| ~~ | ~~4~~ | ~~Fix `nix run .#ci` tidy permission failure~~ |
| ~~ | ~~5~~ | ~~FEATURES.md error simulation PLANNED→DONE~~ |
> Row-level marker note (2026-10-07 second pass): the earlier sweep's buggy wrapper left a bare `~~` in the first cell of these rows. Cells are now uniformly struck; this marks the table resolved wholesale. Per-row outcomes: read the era's git history — this file is archived, closed history.

### MEDIUM Priority (10/10 ✅)

| #  | Item | Resolution                         |
| -- | ---- | ---------------------------------- |
| ~~ | ~~6~~ | ~~MiddlewareWriteFileLog fd leak~~ |
| ~~ | ~~7~~ | ~~Dead `addAttemptCount`~~ |
| ~~ | ~~8~~ | ~~MiddlewareDeduplicate %100 quirk~~ |
| ~~ | ~~9~~ | ~~`--tests` lint flag~~ |
| ~~ | ~~10~~ | ~~examples/ in nix build~~ |
| ~~ | ~~11~~ | ~~Hermetic benchstat~~ |
| ~~ | ~~12~~ | ~~Clean bench-baseline.txt~~ |
| ~~ | ~~13~~ | ~~Commitlint CI gate~~ |
| ~~ | ~~14~~ | ~~release-please~~ |
| ~~ | ~~15~~ | ~~README vs API_STABILITY drift gate~~ |

## Verification

- **Build:** Clean
- **Vet:** Clean
- **Lint:** 0 issues (50+ linters)
- **Tests:** All pass with `-race`
- **Format:** All files formatted (gofumpt + nix fmt)
- **flake.nix syntax:** Valid

## New Public API

| Symbol                                                  | Type     | Status   |
| ------------------------------------------------------- | -------- | -------- |
| `WithCleanup(fn func() error)`                          | Option   | Evolving |
| `NewFileLogMiddleware(path) (Middleware, func() error)` | Function | Evolving |

## New Nix Apps/Checks

| App/Check                         | Purpose                             |
| --------------------------------- | ----------------------------------- |
| `nix run .#lint-tests`            | Explicit test-file linting          |
| `checks.examples-build`           | Compiles `./examples/...`           |
| `benchstat` (internal derivation) | Hermetic benchstat for `bench-diff` |

## New CI Workflows

| Workflow               | Purpose                                        |
| ---------------------- | ---------------------------------------------- |
| `commitlint.yml`       | Conventional-commit format validation          |
| `release-please.yml`   | Automated release PRs from commits             |
| `docs-consistency.yml` | README ↔ API_STABILITY deprecation drift check |

## Files Changed

**Go source:** `middleware.go`, `options.go`, `watcher.go`, `middleware_test.go`, `error_simulation_test.go`, `fake_backend_test.go`
**Nix:** `flake.nix`
**Docs:** `README.md`, `MIGRATION.md`, `FEATURES.md`, `API_STABILITY.md`, `CHANGELOG.md`, `TODO_LIST.md`, `AGENTS.md`, `website/src/content/docs/api-reference.mdx`
**CI:** `.github/workflows/{commitlint,release-please,docs-consistency}.yml`
**Data:** `bench-baseline.txt` (regenerated clean)
