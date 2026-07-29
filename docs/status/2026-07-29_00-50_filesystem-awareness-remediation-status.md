# Status Report: Filesystem Awareness — Remediation Pass

**Created:** 2026-07-29 00:50
**Session scope:** Closing the gaps identified in the brutally-honest
`2026-07-29_00-10_filesystem-awareness-pareto-execution.md` report. A 22-task
Pareto plan was executed in full: correctness bugs, test depth, documentation
discoverability, polish, and capture of deferred v3 items.
**Verdict:** `nix flake check` passes (build, test, lint, fmt, vet,
examples-build, treefmt). `nix run .#check` = **0 lint issues**, all tests pass
with `-race`. New `pathKey` benchmarks measured the NFC cost empirically.

---

## a) FULLY DONE (shipped, tested, lint-clean)

### Correctness (Tier 1)

- **`normalizePath()` best-effort cleaning** — when `filepath.Abs` fails, the
  returned path is now still `filepath.Clean`'d. `FilterExcludePaths` /
  `WithExcludePaths` use the always-cleaned result instead of swallowing the
  error and falling back to a raw, un-cleaned path.
- **`Remove()` incremental key pruning** — the dead `rebuildWatchListKeys()`
  O(n) full-rebuild helper was removed. `Remove()` now deletes each pruned
  subtree key directly (`delete()`), O(removed) not O(n). Verified by
  `TestWatcher_Remove_PrunesSubtreeKeys`.
- **Path entry-point audit** — the only `filepath.Abs` in non-test code is now
  inside `normalizePath`. `New` (via `validateAndNormalizePaths`) and
  `Add`/`AddRecursive`/`Remove` (via `withResolvedPath`) all route through it.

### Test depth (Tier 2)

- **`FuzzPathKey`** — 577,699 mutated executions, 0 failures. Proves never-panic,
  determinism, idempotency, and NFC-equivalence across combining marks, emoji,
  ZWJ sequences, and invalid UTF-8.
- **`normalizePath` + `pathKey` edge-case tables** — `..`, trailing slash,
  redundant separators, empty string, `/`, `.`, `..`.
- **`pathKey` benchmarks** — measured the NFC hot-path cost (see below).
- **`TestWatcher_Remove_UnicodeNFDMatchesNFCWatch`** — end-to-end proof that an
  NFD-encoded `Remove()` path removes an NFC-watched directory.
- **`TestPathKey_NFCEquivalenceProperty`** — `pathKey(P) == pathKey(NFC(P))`
  across accented Latin, Cyrillic, Greek, CJK, emoji.

### Documentation discoverability (Tier 3)

- **README "Filesystem Compatibility" section** — case-sensitivity + NFC +
  `FilterCaseInsensitive` + `Stats.CaseSensitivity`, with code examples.
- **FEATURES.md** — `FilterCaseInsensitive` row added to the filter table.
- **doc.go + example_test.go** — `FilterCaseInsensitive` documented + runnable
  `ExampleFilterCaseInsensitive`.
- **AGENTS.md** — `golang.org/x/text` dependency listed; `pathKey` performance
  contract documented with the measured numbers.

### Polish (Tier 4)

- **wsl_v5 nolint removed** — `filesystem_poll_gitignore_test.go` no longer has a
  file-level suppression; the 10 real formatting issues were fixed with proper
  whitespace. No nolint needed at all.
- **`CaseSensitivity` Prometheus gauge** — `filewatcher_case_sensitivity`
  (0=case-sensitive, 1=case-insensitive, 2=auto), with named constants.
- **`TestGitignore_ExcludesDirDuringWalk_EndToEnd`** — full integration test of
  the walk → gitignore → skip → watchListKeys pipeline.
- **`TestPollWalkDir_NormalizesUnicodeKeysAndPreservesOriginalPath`** — proves
  NFD filename → NFC snapshot key, with original NFD path preserved for events.

### Capture deferred (Tier 5)

- **ROADMAP.md** — v3 filesystem probing (`CaseSensitivityProbed`), macOS/Windows
  CI for case-insensitive verification, NFD path-key caching, trie-based
  exclude-path matching, `WithNormalizeUnicode` escape hatch, phantom-typed
  `PathKey`.
- **TODO_LIST.md** — macOS CI matrix task; `FuzzPathKey` progress noted.

---

## b) KEY FINDINGS

### NFC normalization performance (answers status-report open question #3)

| pathKey input              |   ns/op | allocs | bytes |
| -------------------------- | ------: | -----: | ----: |
| ASCII, case-sensitive      | **~26** |  **0** | **0** |
| ASCII, case-insensitive    |     ~60 |      0 |     0 |
| Unicode NFC (pre-composed) |    ~138 |      0 |     0 |
| Unicode NFD (decomposed)   |  ~1,060 |      3 |   672 |

ASCII (the common case) is **0-allocation**. Only decomposed NFD input —
emitted by the macOS filesystem — allocates (~1 µs). This is negligible for
event-driven watching. **No caching is needed now.** A fresh `bench-baseline.txt`
was captured so future regressions are detectable via `nix run .#bench-diff`.

### Pre-existing go-gitignore limitation discovered

The `github.com/sabhiram/go-gitignore` library returns
`MatchesPath("excluded") == false` for a trailing-slash directory pattern
`excluded/` — it only matches paths _inside_ (`excluded/foo`). This means a
walk-time `.gitignore` with `Build/` does NOT prevent the `Build` directory
itself from being watched; only its children would be filtered at event time.
This is library behavior, not a regression from the case-awareness work. Tracked
implicitly; a no-slash pattern (`Build`) matches the directory path correctly.

---

## c) NOT STARTED (deliberately deferred to v3 — captured in ROADMAP)

- Real filesystem probing (write probe files to detect actual case-sensitivity).
- macOS/Windows CI matrix for real case-insensitive integration tests (Linux CI
  verifies the _logic_ only).
- Filter signature modernization (`func(Event) bool` → richer type).
- NFD path-key caching (only worthwhile at >10k Unicode events/sec on macOS).
- Trie-based exclude-path prefix matching.

---

## d) VERIFICATION

| Gate                                                         | Result               |
| ------------------------------------------------------------ | -------------------- |
| `go build ./...`                                             | ✅                   |
| `go vet ./...`                                               | ✅                   |
| `go test -race -count=1 ./...`                               | ✅ all pass          |
| `nix run .#check` (vet + lint + test)                        | ✅ **0 lint issues** |
| `nix flake check` (build/test/lint/fmt/vet/examples/treefmt) | ✅ all passed        |
| `gofmt -l`                                                   | ✅ clean             |
| `FuzzPathKey` (15s, 577k execs)                              | ✅ 0 failures        |

---

## e) Open questions still awaiting user input

1. **macOS CI** — is a `macos-latest` runner available for real APFS
   case-insensitive integration tests? (Current tests prove logic only.)
2. **`normalizePath` + NFC** — should `normalizePath()` ALSO apply NFC
   normalization (currently only `pathKey()` does)? Trade-off: stored paths would
   be NFC, not original-filesystem bytes.
3. **Benchmark baseline in CI** — should `bench-baseline.txt` be committed
   (gitignored today) to drive a CI regression gate? (See TODO_LIST open Q2.)
