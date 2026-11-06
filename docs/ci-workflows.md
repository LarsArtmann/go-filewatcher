# CI Workflows

| Workflow               | Purpose                                                                     |
| ---------------------- | --------------------------------------------------------------------------- |
| `ci.yml`               | Test (race + coverage), lint, examples-build, benchmark                     |
| `commitlint.yml`       | Validates PR commit subjects follow conventional-commit format              |
| `docs-consistency.yml` | Checks README.md vs API_STABILITY.md deprecation claims don't drift         |
| `release-please.yml`   | Opens release PRs from conventional commits (automates CHANGELOG + version) |
| `release.yml`          | Triggered on `v*` tags — creates GitHub Release                             |

## Branch protection (removed 2026-10-12)

Master protection was LIVE 2026-10-07 → 2026-10-12 and then removed on request
(direct pushes and the auto-commit daemon work again). The retired configuration
— 8 required contexts, strict up-to-date, enforce-admins, linear history, no
review requirement, repo auto-merge disabled — is preserved verbatim in
[branch-protection-backup-2026-10-12.json](./branch-protection-backup-2026-10-12.json)
and restores in one call:

```bash
gh api -X PUT repos/LarsArtmann/go-filewatcher/branches/master/protection \
  --input docs/branch-protection-backup-2026-10-12.json
```

If it is ever re-enabled: strict up-to-date forces sequential PR merges
(`gh pr update-branch N` → wait green → `gh pr merge N --rebase`) and the
auto-commit daemon cannot push. (`src: 2026-10-07_15-51 §f33`)

---

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
