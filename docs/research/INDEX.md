# Research Index

Point-in-time research documents that informed design and tooling decisions.
These are **not living docs** — they capture analysis as of their date and may
reference APIs, versions, or tooling that has since changed. Cross-referenced
from [ROADMAP.md](../../ROADMAP.md) where relevant to long-term planning.

---

## Documents

| Document                                                                           | Date       | Topic                                                                | Decision                                               |
| ---------------------------------------------------------------------------------- | ---------- | -------------------------------------------------------------------- | ------------------------------------------------------ |
| [watchchanges-contract.md](./watchchanges-contract.md)                             | 2026-07-26 | Event-contract analysis for the `Watch` → `Events` channel semantics | Informs v3 API evolution                               |
| [semantic-release-evaluation.md](./semantic-release-evaluation.md)                 | 2026-07-26 | Tradeoff analysis: semantic-release vs release-please vs manual      | Pending decision; release-please currently wired in CI |
| [go-filewatcher-vs-ro-fsnotify.md](./go-filewatcher-vs-ro-fsnotify.md)             | 2026-06-08 | Competitive comparison with `fsnotify` wrapper libraries             | Positioning reference                                  |
| [adopting-samber-ro-pro-contra.md](./adopting-samber-ro-pro-contra.md)             | 2026-06-08 | Evaluation of `samber/ro` for functional utilities                   | Not adopted (adds dependency for marginal benefit)     |
| [go-error-family-adoption-analysis.html](./go-error-family-adoption-analysis.html) | 2026-06-03 | Evaluation of `LarsArtmann/go-error-family` for typed errors         | Not adopted; stdlib `errors`/`fmt` used instead        |

---

## When to Update

- **New research** → add a row to the table above and cross-link from
  `ROADMAP.md` or `TODO_LIST.md`.
- **Research becomes decision** → update the "Decision" column and link the
  resulting ADR in `docs/adr/`.
- **Research goes stale** → add a note at the top of the document pointing to
  the current source of truth. Do not delete (historical context has value).
