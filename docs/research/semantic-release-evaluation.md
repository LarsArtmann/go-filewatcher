# Semantic-release / Conventional Commits — Evaluation

**Status:** Evaluation (no tooling change proposed yet). Decides whether
commit-message-driven versioning is worth adopting to reduce manual CHANGELOG
drift.
**Last updated:** 2026-07-26

## The question

CHANGELOG.md is maintained **by hand** in Keep a Changelog format under an
`[Unreleased]` section, cut into a versioned section when a tag is pushed. This
drifts: entries are forgotten, phrasing is inconsistent, and the version bump
level is a manual judgment call. Could **commit-message-driven versioning**
derive both the changelog and the version bump from the git history automatically?

## Evidence from this repo

The commit history is **already mostly conventional**. Of the last 200 commit
subjects:

| Type                 | Count  |
| -------------------- | ------ |
| `chore:`             | 51     |
| `docs:`              | 44     |
| `fix:`               | 28     |
| `feat:`              | 24     |
| `refactor:`          | 20     |
| `test:`              | 14     |
| `style:`             | 2      |
| `ci:`                | 2      |
| **non-conventional** | **15** |

~92% follow `type(scope): subject`. The 15 outliers are older summary commits
(`functionality`, `project guidelines for AI assistants`) and a few capitalized
ones (`Fix CI: ...`). **Adoption cost is low** because the convention is already
the norm — the gap is enforcement, not culture.

The auto-git commit daemon (per global AGENTS.md) already emits conventional
subjects, so automation is not a blocker.

## Current release path

- `.github/workflows/release.yml` triggers on `v*.*.*` tags, runs tests + lint,
  then calls `softprops/action-gh-release` with `generate_release_notes: true`.
- `generate_release_notes` produces a GitHub release body from PR titles +
  commit list, but it is **not** a structured changelog and does **not** touch
  `CHANGELOG.md`. Version and tag are chosen manually.
- `.goreleaser.yml` exists but is **not invoked** (ROADMAP notes this).

So today: **manual version + manual CHANGELOG + auto release notes.** The drift
is real.

## Options considered

### A. `semantic-release` (JS ecosystem)

- Derives semver bump + changelog from conventional commits; pushes tag, release,
  and changelog automatically.
- **Cons for a Go library:** Node toolchain in CI; fights the Go "tag is the
  source of truth" model (`go get` resolves from tags); does not understand Go
  module `/v2` major-version paths. Net: heavy, ecosystem-mismatched.

### B. `release-please` (Google) — **recommended**

- A GitHub Action that reads conventional commits, opens a **release PR**
  updating `CHANGELOG.md` and a version file, and creates the tag when the PR is
  merged. Reviewable, no surprise releases.
- Native Keep a Changelog output (matches the existing `CHANGELOG.md` style).
- Go-aware: handles `go.mod` major-version suffixes and tag formats.
- Drops in alongside the existing `release.yml` (replaces the manual tag step).
- **This is the closest fit** to a Go library that already writes conventional
  commits and already uses Keep a Changelog.

### C. `goreleaser` (already configured, unused)

- Cross-platform binary release tool. It does **not** derive versions from
  commit messages — it consumes a tag. Useful for shipping binaries (ROADMAP
  "cross-platform release artifacts"), orthogonal to changelog drift. **Not a
  substitute** for B, but **complementary**: release-please cuts the tag,
  goreleaser ships binaries from it.

### D. Status quo + `commitlint` gate only

- Add a CI check rejecting non-conventional subjects, keep CHANGELOG manual.
- Catches the 15 outliers and prevents regression, but does **nothing** for
  drift — the manual editing burden remains. Cheapest, weakest.

## Recommendation

**Adopt `release-please` (B), paired with a `commitlint`/actionlint gate (D).**

- release-please automates the version bump **and** the `CHANGELOG.md` entry from
  the conventional commits this repo already writes — directly removing the drift
  the task names.
- The commitlint gate enforces the 8% that currently slip, protecting the
  automation's input quality.
- goreleaser (C) stays as a separate, later task for binary artifacts; it is not
  blocked by or conflicting with release-please.

### Why not semantic-release

It is the more famous tool but the wrong shape for a Go module: it wants to own
the version lifecycle in a Node-friendly way and would add a Node dependency to a
pure-Go library's CI for no capability that release-please doesn't already
provide natively.

## Tradeoffs / risks

- **Release-please changes the release cadence** from "tag whenever" to "merge
  the release PR." This is more deliberate but less spontaneous. Acceptable for a
  library; document it.
- **Conventional-commit enforcement is now load-bearing.** A bad `feat:` when you
  meant `fix:` bumps a minor instead of a patch. Mitigated by the release PR
  being reviewable before merge.
- **The auto-git daemon must keep emitting conventional subjects.** It already
  does; add a note to its config so a future change doesn't silently break
  release-please's parsing.
- **Initial one-time wiring** in `.github/workflows/` (~1–2 hours) and migrating
  the `[Unreleased]` block once.

## Decision needed

This is a reversible, CI-only change (no library API impact). If approved, the
scoped task moves to `TODO_LIST.md`:

1. Add `release-please-action` workflow (config: Go, Keep a Changelog, `v` tag
   prefix).
2. Add a commit-subject lint gate (commitlint or a simple regex action).
3. One-time: cut a release to let release-please bootstrap from the last tag.
4. Update `ROADMAP.md` / `CONTRIBUTING.md` to describe the new flow.

Until then, this evaluation is the deliverable.
