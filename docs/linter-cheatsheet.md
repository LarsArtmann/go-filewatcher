# Linter Cheat Sheet


50+ linters enabled. Key ones that bite:

| Linter             | Rule                                  |
| ------------------ | ------------------------------------- |
| `exhaustruct`      | All struct fields must be initialized |
| `wrapcheck`        | All errors must be wrapped            |
| `paralleltest`     | All tests must use `t.Parallel()`     |
| `gochecknoglobals` | No globals unless `//nolint`          |
| `gci`              | Import order matters                  |

Run `nix run .#lint-fix` — it auto-fixes many issues.

---

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
