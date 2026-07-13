# Status Report: go-filewatcher Public Presence — DNS + Firebase + Website

**Date:** 2026-07-13 21:58
**Session Scope:** Full public presence overhaul across two repos: go-filewatcher (README + website + GitHub metadata) and domains (DNS CNAME + ACME challenge for Firebase Hosting)
**Status:** Website live on `filewatcher.web.app`. Custom domain DNS configured in Terraform but NOT applied (blocked on Namecheap API credentials). Firebase custom domain claim created but waiting on DNS propagation.

---

## a) FULLY DONE

### Website (`go-filewatcher/website/`) — Live & Deployed

- **59 source files**, Astro 7 + Starlight + Tailwind v4, built from scratch following go-atomic-write/gogenfilter patterns
- **13 pages** generated, builds in ~1.6s, zero errors
- **Deployed** to Firebase Hosting site `filewatcher` in project `lars-software`
- **Live URL:** `https://filewatcher.web.app` — verified HTTP 200, all security headers active (HSTS, COOP, CORP, X-Frame-Options, etc.)
- Landing page: hero with GitHub stars fetch, code preview with copy button, 6 feature cards, 4-step pipeline, 9-row comparison matrix, 3 use case cards, CTA
- 11 Starlight docs pages: installation, quick-start, filtering, middleware, debouncing, resilience, observability, api-reference, changelog, contributing, related-tools
- Violet accent (#8b5cf6), eye/watch SVG logo, dark/light theme with FOUC-free init
- Firebase config: multi-site target, security headers, immutable asset caching, clean URLs, Nix flake with dev/build/preview/deploy apps

### README.md — Rewritten

- Professional centered-header format matching reference repos
- Why? section, comparison table, how-it-works pipeline, design decisions, dependencies table, error handling, API stability
- All API tables preserved (24 options, 17+ filters, 18 middleware)
- Removed emoji-heavy style, added Go Report Card badge

### GitHub Repo Metadata — Updated

- **Description:** "A high-performance, composable file system watcher for Go. Built on fsnotify with automatic recursion, 17+ filters, 18 middleware, and production-grade resilience."
- **Homepage URL:** `https://filewatcher.lars.software`
- **Topics:** 14 topics (added: event-driven, file-system-watcher, inotify, observability, polling, resilience, hierarchical-watcher)

### Firebase Hosting — Site Created & Deployed

- `firebase hosting:sites:create filewatcher --project lars-software` — executed successfully
- `firebase deploy --only hosting:filewatcher` — 65 files uploaded, version released
- Site URL confirmed: `https://filewatcher.web.app`

### Firebase Custom Domain — Claimed

- REST API call to `firebasehosting.googleapis.com/v1beta1/sites/filewatcher/domains` succeeded
- Domain `filewatcher.lars.software` registered with `DOMAIN_ACTIVE` status
- Returns `CERT_PENDING` / `DNS_MISSING` — waiting for DNS propagation

### Terraform DNS Configuration — Written & Validated

- Two new records added to `domains/lars.software.tf`:
  1. CNAME `filewatcher` → `filewatcher.web.app.` (TTL 1799)
  2. TXT `_acme-challenge.filewatcher` → `LYzZFQvIIMYOUALIO2R99_ofNoLWrUdUWgGAnX29K1M` (TTL 1799)
- `terraform validate` — passes
- `terraform fmt -check` — passes
- Format matches existing patterns in the file (gogenfilter, go-atomic-write, go-output, go-workflow-auditlog)

### Status Report — Previous Session

- Written to `docs/status/2026-07-13_21-22_public-presence-overhaul-status.md`

---

## b) PARTIALLY DONE

### Custom Domain `filewatcher.lars.software`

- Firebase side: domain claimed, ACME challenge token retrieved
- DNS side: Terraform config written, formatted, validated
- **NOT APPLIED** — `terraform apply` blocked because `terraform.tfvars` contains placeholder API key (`REPLACE_WITH_YOUR_API_KEY`)
- Current IP (89.65.239.240) may also not be whitelisted in Namecheap
- Firebase provisioning shows `DNS_MISSING` — the CNAME hasn't propagated because it hasn't been created yet

### Website Content Depth

- Docs pages are accurate and functional but thin — essentially the README tables reformatted into Starlight
- Lacks deep walkthroughs, multi-file examples, troubleshooting scenarios, visual diagrams

---

## c) NOT STARTED

### Terraform Apply

- Cannot apply DNS changes without real Namecheap API key
- The `terraform.tfvars` file exists but contains `namecheap_api_key = "REPLACE_WITH_YOUR_API_KEY"`
- User must provide credentials or apply manually via Namecheap dashboard

### DNS Propagation Verification

- Never verified whether the DNS records propagated after applying (because we never applied)
- No `dig` or `nslookup` checks done against `filewatcher.lars.software`

### SSL Certificate Provisioning

- Firebase reports `CERT_PENDING` — will remain pending until DNS propagates
- Once CNAME + TXT records are live and propagated, Firebase auto-provisions the SSL cert

### Website CI/CD Pipeline

- No GitHub Actions workflow for auto-deploying website on push
- No Firebase deploy token configured as a GitHub secret

### Custom OG Image

- No Open Graph image created for social sharing
- `og:image` meta tag absent from landing page

### Website typecheck / html-validate

- `npm run typecheck` never run
- `html-validate` config exists but never executed against build output

---

## d) TOTALLY FUCKED UP

Nothing is broken or corrupted. However:

1. **Placeholder API key** — The `terraform.tfvars` in the domains repo has `namecheap_api_key = "REPLACE_WITH_YOUR_API_KEY"`. This is not my file to edit (it contains credentials and is gitignored), but it blocked the apply. I should have checked for this earlier in the session instead of at apply time.

2. **Left temp files** — `/tmp/add-firebase-domain.js` and `/tmp/check-firebase-domain.js` were created for the REST API calls. Harmless but untidy. Could have been inline.

---

## e) WHAT WE SHOULD IMPROVE

### Critical

1. **Unblock DNS apply** — Either get the real Namecheap API key into `terraform.tfvars`, or add the two records manually via Namecheap dashboard. Without this, the custom domain never works.
2. **Verify SSL provisioning** — After DNS propagates (minutes to hours), check that Firebase transitions from `CERT_PENDING` to `CERT_OK`.
3. **Verify custom domain serves content** — `curl -sI https://filewatcher.lars.software` should return 200 after DNS + SSL.

### Quality

4. **Website content expansion** — Each guide page should have real walkthroughs, not just reformatted README tables.
5. **Add OG image** — Create a 1200x630 image for social sharing.
6. **Add favicon PNG variants** — Only SVG exists; need PNG fallbacks + apple-touch-icon.
7. **Run typecheck** — `npm run typecheck` in the website to catch TypeScript issues.
8. **Website CI/CD** — Auto-deploy on push to master.

---

## f) UP TO 50 THINGS TO GET DONE NEXT

### DNS & Domain (do first — currently blocked)

1. Get real Namecheap API key into `domains/terraform.tfvars`
2. Whitelist current public IP (89.65.239.240) in Namecheap API dashboard OR confirm it's already whitelisted
3. Run `nix run nixpkgs#opentofu -- apply -target=namecheap_domain_records.lars_software` from domains repo
4. Verify DNS propagation: `dig filewatcher.lars.software CNAME`
5. Verify ACME challenge TXT propagated: `dig _acme-challenge.filewatcher.lars.software TXT`
6. Check Firebase domain status transitions to `CERT_OK`
7. Verify `https://filewatcher.lars.software` returns HTTP 200
8. Verify HTTPS certificate is valid (not self-signed, no warnings)

### Website CI/CD

9. Create `.github/workflows/website-deploy.yml` in go-filewatcher
10. Add `FIREBASE_TOKEN` as GitHub secret
11. Configure workflow to build + deploy on push to master
12. Add website build status badge to README

### Website Content

13. Expand Filtering guide with real multi-file project examples
14. Expand Middleware guide with full e2e walkthrough
15. Expand Resilience guide with NFS/FUSE setup walkthrough
16. Expand Observability guide with actual Prometheus YAML config
17. Expand Observability guide with actual OpenTelemetry collector config
18. Add Troubleshooting docs page (ENOSPC, NFS, large monorepos)
19. Add Migration from raw fsnotify docs page
20. Add Migration from v1 to v2 docs page
21. Add Performance Tuning guide (batch sizes, buffer sizes, filter ordering)
22. Add architecture diagrams (pipeline flow, middleware order) via Mermaid or images

### Design & Assets

23. Create OG image (1200x630) for social sharing
24. Create PNG favicon variants (32x32, 192x192, 512x512) from SVG
25. Add apple-touch-icon (180x180) to head
26. Create favicon.ico for legacy browser support
27. Redesign logo SVG to be more distinctive
28. Add hero image or background graphic to landing page
29. Add staggered row reveal animation to comparison table
30. Improve feature card hover states

### Code Quality

31. Run `npm run typecheck` and fix any TypeScript strict errors
32. Run html-validate against build output
33. Add ESLint config for website (Astro + TypeScript)
34. Test website in Firefox and Safari
35. Run Lighthouse audit on built site
36. Fix any accessibility issues found
37. Test prefers-color-scheme media query (dark/light toggle)
38. Verify CSP headers work with inline scripts

### Cross-Repo Consistency

39. Align website flake.nix nixpkgs revision with root project
40. Update go-filewatcher AGENTS.md to mention website/ directory
41. Update go-filewatcher FEATURES.md to mention documentation website
42. Verify root .gitignore covers website/node_modules and website/dist
43. Add website to root project's flake.nix or devShell

### Repo Hygiene

44. Remove temp files: `/tmp/add-firebase-domain.js`, `/tmp/check-firebase-domain.js`
45. Decide on package-lock.json: commit or gitignore
46. Clean up website/flake.lock if nixpkgs diverges from root

### README

47. Add website screenshot/GIF to README
48. Add "Why not X?" section comparing popular alternatives
49. Add Codecov badge if coverage reporting exists
50. Add "Used by" / "Adopters" section if projects use the library

---

## g) TOP 2 QUESTIONS I CANNOT ANSWER MYSELF

### 1. What is the real Namecheap API key?

The `terraform.tfvars` contains `namecheap_api_key = "REPLACE_WITH_YOUR_API_KEY"`. I cannot apply the DNS changes without it. Either provide the key, or add the two DNS records manually via the Namecheap dashboard:

- CNAME: `filewatcher` → `filewatcher.web.app.` (TTL 1799)
- TXT: `_acme-challenge.filewatcher` → `LYzZFQvIIMYOUALIO2R99_ofNoLWrUdUWgGAnX29K1M` (TTL 1799)

### 2. Is the current public IP (89.65.239.240) whitelisted in Namecheap?

Even with a valid API key, the Namecheap API rejects calls from non-whitelisted IPs. If this IP is not whitelisted (and dynamic IPs may change), terraform apply will fail. Need either a whitelisted static IP, the current IP added to the whitelist, or manual DNS record creation via the dashboard.
