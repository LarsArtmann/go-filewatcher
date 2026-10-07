# CI Workflows


| Workflow               | Purpose                                                                     |
| ---------------------- | --------------------------------------------------------------------------- |
| `ci.yml`               | Test (race + coverage), lint, examples-build, benchmark                     |
| `commitlint.yml`       | Validates PR commit subjects follow conventional-commit format              |
| `docs-consistency.yml` | Checks README.md vs API_STABILITY.md deprecation claims don't drift         |
| `release-please.yml`   | Opens release PRs from conventional commits (automates CHANGELOG + version) |
| `release.yml`          | Triggered on `v*` tags — creates GitHub Release                             |

---

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
