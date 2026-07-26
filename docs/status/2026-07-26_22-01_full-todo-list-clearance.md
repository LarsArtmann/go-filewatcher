# TODO List Clearance — All 15 Items Complete

**Date:** 2026-07-26_22-01
**Scope:** All HIGH (5) + MEDIUM (10) priority items from the 2026-07-26 TODO harvest.

## Summary

Every item from the pasted TODO list is implemented, tested, lint-clean, and documented.

### HIGH Priority (5/5 ✅)

| #   | Item                                       | Resolution                                                                                   |
| --- | ------------------------------------------ | -------------------------------------------------------------------------------------------- |
| 1   | Mark deprecations in README.md             | Added ⚠️ deprecated callouts to `WithOnError` and `MiddlewareRateLimit` table rows           |
| 2   | v2.3→v3 migration section in MIGRATION.md  | Added before/after snippets for both deprecated symbols                                      |
| 3   | Deprecation badges in api-reference.mdx    | Starlight `:::caution[Deprecated]` blocks, moved deprecated symbols to dedicated subsections |
| 4   | Fix `nix run .#ci` tidy permission failure | `ci`/`fmt`/`tidy` apps now run from caller CWD (no `cd "${self}"` into read-only store)      |
| 5   | FEATURES.md error simulation PLANNED→DONE  | Moved to Developer Experience section as ✅                                                  |

### MEDIUM Priority (10/10 ✅)

| #   | Item                               | Resolution                                                                                                |
| --- | ---------------------------------- | --------------------------------------------------------------------------------------------------------- |
| 6   | MiddlewareWriteFileLog fd leak     | Added `NewFileLogMiddleware` (returns closer) + `WithCleanup` option; `Close()` calls registered cleanups |
| 7   | Dead `addAttemptCount`             | Wired up in `TestSelfHeal_HealsFailedPathAfterRetry` and `TestFakeBackend_AddFailsSpecificPaths`          |
| 8   | MiddlewareDeduplicate %100 quirk   | Replaced `len(seen)%100` with `eventCount%dedupeCleanupInterval` (event counter)                          |
| 9   | `--tests` lint flag                | Added `nix run .#lint-tests` app (explicit `--tests` opt-in)                                              |
| 10  | examples/ in nix build             | Added `./examples` to `fileset.unions` + `examples-build` check in `checks`                               |
| 11  | Hermetic benchstat                 | `buildGoModule` derivation in flake; `bench-diff` uses it instead of `go run @latest`                     |
| 12  | Clean bench-baseline.txt           | Regenerated with `-run=^$` (skips test functions that produce slog output)                                |
| 13  | Commitlint CI gate                 | `.github/workflows/commitlint.yml` — validates conventional-commit format on PRs                          |
| 14  | release-please                     | `.github/workflows/release-please.yml` — release-type: go, Keep a Changelog                               |
| 15  | README vs API_STABILITY drift gate | `.github/workflows/docs-consistency.yml` — checks deprecation claims cross-file                           |

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
