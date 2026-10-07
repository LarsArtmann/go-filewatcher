# v3 Decision Ledger

Decisions and constraints that any v3 cut must respect or explicitly overturn.
Point-in-time like the rest of [research/](./INDEX.md): entries carry their
date and the evidence that produced them. Newest first.

---

## Constraints (non-negotiable inputs)

### C1. `encoding/json/v2` requires a go-directive floor bump (2026-10-07)

`event.go` imports `encoding/json/v2`, which is GOEXPERIMENT-gated today and —
decisively — **requires the `go` directive to be ≥ 1.27 when compiled by the
Go 1.27 toolchain** ("json.Marshal requires go1.27 or later (file is
go1.26)", CI run 37581977239). Consequences:

- The repo floor is pinned at `go 1.26.7` (CI `go-directive` job + flake check
  both assert it), so a Go 1.27 CI leg is **structurally impossible** until the
  floor moves. A 1.27 matrix entry was tried and removed the same day.
- **v3 must decide as one atomic change:** floor bump to ≥ 1.27 **and**
  keep/replace the json/v2 usage **and** re-add the 1.27 CI leg **and** update
  both floor guards (`go-directive` CI job, `checks.go-directive` in
  flake.nix). Half-states do not build.
- Alternative recorded for v3: drop `encoding/json/v2` (back to `encoding/json`
  or a hand-rolled encoder) and keep the 1.26.7 floor. The v2 usage is limited
  to event serialization; the perf win was not benchmarked against v1.

## Decisions taken (v2, binding context for v3)

### D1. Keep release-please, path-scoped (2026-10-07)

Config-file mode with `exclude-paths: ["website", "docs", ".github"]`.
Commits touching only excluded paths never cut releases; a Go-less release
(v2.4.1, shipped 2026-10-07) proved the guard works. v3 keeps this wiring.

### D2. Required status checks + linear history (2026-10-07)

8 required contexts, `enforce_admins`, `required_linear_history`. The
auto-commit daemon can no longer push. v3 inherits this protection model;
release PRs are the only release path (workflow_dispatch verified as the
manual fallback, 2026-10-07).

### D3. Windows/macOS legs are probe-grade (2026-10-07)

The matrix runs macos-15 + windows-2025 (informational, not required).
Budget enforcement is inotify-only; `failedPaths` lookups must use `pathKey()`;
POSIX-shape assertions skip via `skipOnWindows(t)`. v3 either promotes the
legs to required (after per-OS test expectations exist) or documents why not.

## Open questions for the v3 planning pass

- Floor bump timing: ride the Go 1.27 `encoding/json/v2` stabilization (is it
  out of GOEXPERIMENT by then?) vs drop json/v2 first and decouple.
- Deprecation inventory: `WithWatchedIgnoreDirs`, `WithOnError`,
  `MiddlewareRateLimit`, the two-arg `ErrorHandler` signature, a runtime
  warning for `MiddlewareWriteFileLog` (see ROADMAP "API Evolution").
- `WithNormalizeUnicode(false)` escape hatch (ROADMAP: always-on NFC today).
- `CaseSensitivityProbed`: probe the real mount instead of `runtime.GOOS`
  (ROADMAP: probing is the remaining gap).
- Poll-mode event quality: `WithPollDeduplicate(true)` /
  `WithPollDetectRenames(true)` are designed-but-unbuilt.
