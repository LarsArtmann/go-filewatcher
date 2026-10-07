# Website Deploy Runbook

How `filewatcher.lars.software` gets built and deployed (Firebase Hosting).
The website lives in [`website/`](../website/) — a separate Astro + Starlight
toolchain with its own flake, **not part of the Go module**.

## Deploy

```bash
cd website
nix run .#deploy   # pnpm run build + firebase deploy --only hosting
```

Requires Firebase credentials (google-github-auth or `firebase login`
beforehand). Other apps: `nix run .#dev` (local dev server), `nix run
.#preview` (build + preview), `nix run .#build` (build only).

## Gotchas that have bitten

1. **GitHub slugs strip dots.** A guide named `migration-v2.3-to-v2.4.mdx`
   publishes as `/guides/migration-v23-to-v24/`. Before publishing links
   (`gh release edit`, release notes, README), verify with
   `ls dist/guides/` or fetch the live URL — the unchecked slug 404'd in the
   published v2.4.0 notes (2026-10-06).
2. **`trailingSlash: false` + `cleanUrls: true`** (`firebase.json`): URLs have
   NO trailing slash and NO `.html`. Write links accordingly.
3. **pnpm 11 build-script approvals**: `esbuild` must stay approved under
   `allowBuilds:` in `website/pnpm-workspace.yaml`, or `astro build` fails on
   a missing esbuild binary. A placeholder value silently disables the whole
   key.
4. **The `/changelog` page is generated.** `sync-changelog.mjs` reads the repo
   root `CHANGELOG.md` on every build — never hand-edit the mdx; update
   CHANGELOG.md and rebuild.
5. **Dependency security**: `pnpm audit --fix` in pnpm 11 needs an explicit
   strategy (`--fix=update` re-resolves the lockfile; `--fix=override` adds
   overrides to `pnpm-workspace.yaml`). Verify the build after any override —
   forced major bumps can break the build.
