# Agent Guide: go-filewatcher

**Go 1.26.7** | `github.com/larsartmann/go-filewatcher/v2` | **MIT License**

> **Companion docs:** [FEATURES.md](./FEATURES.md) (feature inventory) ·
> [ROADMAP.md](./ROADMAP.md) (long-term direction) · [TODO_LIST.md](./TODO_LIST.md)
> (actionable work) · [CHANGELOG.md](./CHANGELOG.md) (release history).
> This file is for enduring, hard-to-discover-from-code context only.

---

## Critical Commands

```bash
# Using Nix flake (recommended)
nix develop              # Enter development shell with Go and tools
direnv allow             # Auto-load environment on cd (requires direnv)

# Nix apps (run from anywhere, no need to be in dev shell)
nix run .#check          # Full quality: vet + lint + test
nix run .#ci             # Full CI: tidy + fmt + vet + lint + test
nix run .#lint-fix       # Auto-fix linter issues
nix run .#test           # Run tests with -race
nix run .#test-v         # Run tests with -race -v
nix run .#lint           # Run linter
nix run .#lint-tests     # Run linter (--tests explicit)
nix run .#bench          # Run benchmarks
nix run .#bench-baseline # Capture benchmark baseline (run from repo root)
nix run .#bench-diff     # Diff benchmarks vs baseline (hermetic benchstat)
nix run .#coverage       # Generate coverage report
nix run .#fmt            # Format Go code (writes — run from repo root)
nix run .#tidy           # Run go mod tidy (writes — run from repo root)
nix run .                # Default = check

# Nix quality gates
nix flake check          # Run all checks (build, test, lint, fmt, vet, examples-build)
nix build .              # Validate reproducible build
nix fmt                  # Format .nix files

# Inside dev shell (aliases are set automatically):
check       # nix run .#check
ci          # nix run .#ci
lint        # nix run .#lint
lint-fix    # nix run .#lint-fix
test        # nix run .#test
```

## Updating vendorHash

When `go.mod`/`go.sum` change, update `vendorHash` in `flake.nix`: run
`nix build .` and copy the `got:` hash from the error (placeholder-then-rebuild
procedure: [docs/nix-vendorhash.md](docs/nix-vendorhash.md).

## Non-Obvious Conventions

Uses `errors`/`fmt` only — sentinel errors, `%w` wrapping, `errors.Is`. Full
sample: [docs/error-handling.md](docs/error-handling.md).

### Single Package Layout

All code in **root package** (`filewatcher`). No `internal/` or `pkg/` subdirectories — all code lives in the package root.

### File Organization

All code lives in the root package `filewatcher`; the per-file responsibility
table (watcher.go, backend.go, filter.go, middleware.go, ...) moved to
[docs/file-organization.md](docs/file-organization.md).

### Examples (`examples/`)

Separate Go programs demonstrating usage. Each subdirectory is a standalone
`main` package that imports the library. Shared helpers (including
`demo.MustWatch`) live in `examples/demo/`. Build with `go run ./examples/<name>`
or `go build ./examples/...`. Part of the same Go module (no separate go.mod).

### Website (`website/`)

Separate Astro + Starlight site (`filewatcher.lars.software`, Firebase Hosting) with
its own flake/toolchain — not part of the Go module. Build `cd website && nix run
.#build`. The `/changelog` page is GENERATED from repo `CHANGELOG.md` by
`sync-changelog.mjs` on every build — never hand-edit the mdx; update CHANGELOG and
rebuild. Deploy procedure + quirks:
[docs/website-deploy-runbook.md](docs/website-deploy-runbook.md).

Website JS/TS is linted by BuildFlow's oxlint step under a NESTED
`website/.oxlintrc.json` (2026-10-12) carving website conventions out of the
root BuildFlow config: PascalCase components (filename-case off), intentional
content/config key order (sort-keys off), one-declaration-per-line (one-var /
sort-vars off), font-weight/timeout literals (no-magic-numbers off), `window`
globals (prefer-global-this off), console/process-exit/fs-sync allowed for
`scripts/**` + `*.config.*`. A nested oxlint config REPLACES the root `rules`
block (no merge) — carry the root's `oxc/*` exceptions (no-async-await etc.)
into it or they resurface. oxlint sort-imports orders by first local name,
case-sensitive (uppercase sorts before lowercase). `astro check` is BROKEN
here (TS 7.0 unsupported); the working typecheck is `tsc --noEmit` in
`website/` (what BuildFlow runs), plus `pnpm run build`.

---

## Critical Gotchas (index)

Full detail for every gotcha: [docs/gotchas.md](docs/gotchas.md). The one-liners:

1. Middleware order is REVERSED — first in `WithMiddleware` runs outermost.
2. `WithDebounce` = global (all events → one callback); `WithPerPathDebounce` = per-path.
3. `exhaustruct` linter: every struct field must be initialized.
4. Every test needs `t.Parallel()` (paralleltest linter).
5. Multi-op events resolve by priority Create > Write > Remove > Rename.
6. Chmod events are ignored (`convertEvent` returns nil).
7. `DefaultIgnoreDirs` is an exported global with `//nolint:gochecknoglobals` — keep the nolint.
8. `WithDebug` is active: wires slog debug logging through the pipeline.
9. `WithPolling` is active: starts a pollLoop alongside fsnotify (NFS/FUSE).
10. Circuit breaker: Closed → Open → HalfOpen; one event probes recovery.
11. `fswatcher.Add` errors (incl. ENOSPC) degrade instead of aborting the walk; `Stats.WatchErrors` counts them.
12. `maxWatches` auto-detected on Linux; override with `WithMaxWatches`; over-budget dirs skip silently.
13. `.gitignore` walking is default-on; trailing-slash patterns (`Build/`) do NOT match the dir itself.
14. Watch registration is batched (1000/batch, `runtime.Gosched()` between).
15. `WithExcludePaths`: absolute-path prefix matching, walk-time only.
16. `Remove()` removes the whole subtree, not just the top-level dir.
17. `Reset()` clears runtime state, keeps config; enables `Watch()` again after `Close()`.
18. `pathKey()` = NFC-normalized (+lowercased when case-insensitive); perf contract and bench-diff methodology in docs/gotchas.md #18.
19. `watchListKeys` map gives O(1) duplicate detection and self-heal checks.
20. Symlink cycle detection via `symlinkVisited` map (`WithFollowSymlinks`).
21. Middleware drops tracked via atomic flag → `Stats.EventsDroppedByMiddleware`.
22. `Add`/`AddRecursive`/`Watch` validate paths (`ErrPathNotFound`/`ErrPathNotDir`); `Remove` does not.
23. Error classification: permission/not-exist/ENOTDIR permanent (self-heal abandons); ENOSPC transient.
24. Poll mode applies the same skip/exclusion/gitignore logic as the initial walk.
25. `EventChannelDropOnFull`: non-blocking send, drops counted; default is blocking backpressure.
26. Watch-budget safety fraction applies ONLY to auto-detected limits; explicit `WithMaxWatches` wins, incl. through `Reset()`.
27. Windows/macOS CI legs: budget enforcement is inotify-only; `failedPaths` lookups MUST go through `pathKey()` (raw paths only work on Linux); POSIX-shape tests skip via `skipOnWindows(t)`.
28. Breaker state is only observable via `NewCircuitBreaker` + `WithCircuitBreaker` (the anonymous `MiddlewareCircuitBreaker` closure is invisible — health reports `CircuitNone`); `Healthy(ctx)` (go-health `CheckFunc`-compatible) treats HalfOpen and drop counters as healthy; `Reset()` does NOT clear `failedPaths`.

## Key Patterns

Pattern→location table (13 rows) moved to [docs/patterns.md](docs/patterns.md).

Default-guard convention + middleware resource-cleanup patterns (worked
examples): [docs/patterns.md](docs/patterns.md).

## Dependencies

```
github.com/fsnotify/fsnotify          # Core file watching (v1.10.1)
github.com/LarsArtmann/gogenfilter/v3 # Generated code detection (v3.6.1, published — NO local replace)
github.com/sabhiram/go-gitignore      # .gitignore pattern matching (zero transitive deps)
golang.org/x/text/unicode/norm        # NFC Unicode normalization in pathKey (v0.42.0)
golang.org/x/time/rate                # rate.Limiter for MiddlewareThrottle
```

gogenfilter v3 API notes (v3.6.1 migration semantics, sqlc detection):
[docs/gogenfilter-v3-notes.md](docs/gogenfilter-v3-notes.md).

## Named Types (phantom types)

`type X string` types for compile-time safety on path-like strings:
`EventPath`, `RootPath`, `DebounceKey`, `LogSubstring`, `TempDir`, `OpString`.
Constructors like `NewEventPath()`; full table + domain methods
(`.Base()/.Dir()/.Ext()/.Join()`): [docs/phantom-types.md](docs/phantom-types.md).

## Release / CI Gotchas (index)

Full detail, fix commands, and incident history:
[docs/release-ci-gotchas.md](docs/release-ci-gotchas.md).

| Gotcha                    | One-liner                                                                                                                                                                                    |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| release-please permission | Needs the repo's "can approve pull request" permission or the release PR fails. It cuts tag + GitHub Release on merge; `release.yml` only fires for manual `v*` tags.                        |
| release-please path-scope | `exclude-paths: ["website", "docs", ".github"]` — non-module commits use `chore/docs/build(website)` scopes, never `fix(website)`/`feat(website)`.                                           |
| golangci-lint version     | Pin one version across flake (`pkgs.golangci-lint`), `ci.yml`, and `release.yml`; `.golangci.yml` needs the `exhaustruct_v5` key (v2.13+).                                                   |
| Release-PR CI "workflow"  | Release-please PRs report failure with ZERO jobs (GitHub quirk). Judge releases by master CI on the merge commit.                                                                            |
| Publishing doc URLs       | GitHub strips dots from heading slugs (`v2.3-to-v2.4` → `v23-to-v24`); check `ls dist/guides/` before `gh release edit`.                                                                     |
| Auto-commit daemon        | It commits/resets local master while you work. `git fetch` + check `git log HEAD..origin/master` before pushing; never assume your commit is HEAD.                                           |
| Git Town `git sync`       | Merges master into EVERY local branch; one stale merged-PR branch conflict-pauses the whole run. Verify branch content with `git cherry master <branch>` before `git branch -D`.             |
| Branch protection         | REMOVED 2026-10-12 (direct pushes work). Restore with `docs/branch-protection-backup-2026-10-12.json`.                                                                                       |
| Windows/macOS CI legs     | POSIX-shape tests skip via `skipOnWindows(t)`; coverage steps need `shell: bash`. Probe legs only.                                                                                           |
| ubuntu-latest → 26.04     | Rollout completes 2026-11-19. Jobs are pinned to ubuntu-24.04 except the dual-OS test matrix, until 24.04 brownouts.                                                                         |
| go.mod language floor     | `go 1.26.7` is enforced by the `go-directive` nix check AND a blocking CI job. Three BuildFlow tools fight it — never fix a bump by editing the guards alone.                                |
| Nix sandbox / BuildFlow   | GOEXPERIMENT, nix-hash-fix, and pnpm-audit cache replay each need a specific workaround: see the doc before improvising.                                                                     |
| Accepted BuildFlow noise  | `file-size-check`, `cue-vet`, `go-auto-upgrade`, `nix-checker`, `dependabot`, `art-dupl`, `vulnix` findings are baselined or documented non-fixes. Read the doc's table before "fixing" one. |
