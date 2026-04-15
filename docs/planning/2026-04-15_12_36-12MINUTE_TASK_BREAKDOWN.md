# 12-Minute Task Breakdown (60 Tasks)

**Date:** 2026-04-15 12:36  
**Max Task Time:** 12 minutes each  
**Total Tasks:** 60

---

## Table 1: Race Condition Fix (Tasks 1-12) - BLOCKING

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 1 | Remove emitWg field from Watcher struct | 5min | CRITICAL | - |
| 2 | Remove emitWg initialization in New() | 5min | CRITICAL | Task 1 |
| 3 | Remove emitWg.Add/Done from emitEvent | 5min | CRITICAL | Task 1 |
| 4 | Revert watchLoop defer to close eventCh | 5min | CRITICAL | - |
| 5 | Add defer/recover in buildEmitFunc | 5min | CRITICAL | - |
| 6 | Remove eventCh field from Watcher | 5min | CRITICAL | - |
| 7 | Remove eventCh assignment in Watch() | 5min | CRITICAL | Task 6 |
| 8 | Simplify Close() remove eventCh close | 5min | CRITICAL | - |
| 9 | Build to check compilation | 3min | CRITICAL | Tasks 1-8 |
| 10 | Run tests without -race | 5min | CRITICAL | Task 9 |
| 11 | Run tests with -race | 10min | CRITICAL | Task 10 |
| 12 | Commit race fix | 5min | CRITICAL | Task 11 |

**Table 1 Total:** ~53 minutes

---

## Table 2: Context Cancellation Fix (Tasks 13-18) - BLOCKING

| # | Task | Time | Priority | Depends On |
| 3 | Fix context cancellation test hang | 10min | CRITICAL | Task 12 |
| 14 | Add test timeout handling | 5min | HIGH | Task 13 |
| 15 | Verify test passes with -race | 5min | CRITICAL | Task 14 |
| 16 | Commit test fix | 5min | HIGH | Task 15 |

**Table 2 Total:** ~25 minutes

---

## Table 3: Cleanup & Tagging (Tasks 19-25) - HIGH PRIORITY

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 19 | Update TODO_LIST.md with completions | 5min | HIGH | Task 16 |
| 20 | Run linter to verify 0 issues | 5min | HIGH | Task 16 |
| 21 | Review git status | 3min | MEDIUM | Task 20 |
| 22 | Write v0.1.0 release notes | 10min | HIGH | Task 20 |
| 23 | Create v0.1.0 tag | 5min | HIGH | Task 22 |
| 24 | Push tags to origin | 2min | HIGH | Task 23 |
| 25 | Verify release on GitHub | 5min | MEDIUM | Task 24 |

**Table 3 Total:** ~35 minutes

---

## Table 4: Event Enhancements (Tasks 26-35) - MEDIUM PRIORITY

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 26 | Add Size field to Event struct | 5min | MEDIUM | Task 25 |
| 27 | Update convertEvent to populate Size | 10min | MEDIUM | Task 26 |
| 28 | Test Event.Size with unit tests | 5min | MEDIUM | Task 27 |
| 29 | Commit Event.Size | 5min | MEDIUM | Task 28 |
| 30 | Add ModTime field to Event struct | 5min | MEDIUM | Task 29 |
| 31 | Update convertEvent to populate ModTime | 5min | MEDIUM | Task 30 |
| 32 | Test Event.ModTime | 5min | MEDIUM | Task 31 |
| 33 | Commit Event.ModTime | 5min | MEDIUM | Task 32 |
| 34 | Update README with new fields | 10min | MEDIUM | Task 33 |
| 35 | Commit README update | 5min | MEDIUM | Task 34 |

**Table 4 Total:** ~55 minutes

---

## Table 5: Filter Enhancements (Tasks 36-48) - MEDIUM PRIORITY

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 36 | Add FilterExcludePaths function | 10min | MEDIUM | Task 35 |
| 37 | Test FilterExcludePaths | 5min | MEDIUM | Task 36 |
| 38 | Commit FilterExcludePaths | 5min | MEDIUM | Task 37 |
| 39 | Add FilterMinAge function | 10min | MEDIUM | Task 38 |
| 40 | Test FilterMinAge | 5min | MEDIUM | Task 39 |
| 41 | Commit FilterMinAge | 5min | MEDIUM | Task 40 |
| 42 | Add FilterMaxSize function | 10min | MEDIUM | Task 41 |
| 43 | Test FilterMaxSize | 5min | MEDIUM | Task 42 |
| 44 | Commit FilterMaxSize | 5min | MEDIUM | Task 43 |
| 45 | Update filter documentation | 10min | MEDIUM | Task 44 |
| 46 | Commit filter docs | 5min | MEDIUM | Task 45 |
| 47 | Run all tests | 5min | MEDIUM | Task 46 |
| 48 | Commit any test fixes | 5min | MEDIUM | Task 47 |

**Table 5 Total:** ~75 minutes

---

## Table 6: Advanced Features (Tasks 49-55) - MEDIUM PRIORITY

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 49 | Add WithPolling option skeleton | 5min | MEDIUM | Task 48 |
| 50 | Research fsnotify polling support | 10min | MEDIUM | Task 49 |
| 51 | Implement polling fallback | 10min | MEDIUM | Task 50 |
| 52 | Test WithPolling | 5min | MEDIUM | Task 51 |
| 53 | Commit WithPolling | 5min | MEDIUM | Task 52 |
| 54 | Add symlink following support | 10min | MEDIUM | Task 53 |
| 55 | Test symlink following | 5min | MEDIUM | Task 54 |

**Table 6 Total:** ~50 minutes

---

## Table 7: Middleware & WatchOnce (Tasks 56-60) - MEDIUM PRIORITY

| # | Task | Time | Priority | Depends On |
|---|------|------|----------|------------|
| 56 | Add MiddlewareDeduplicate | 10min | MEDIUM | Task 55 |
| 57 | Test MiddlewareDeduplicate | 5min | MEDIUM | Task 56 |
| 58 | Commit MiddlewareDeduplicate | 5min | MEDIUM | Task 57 |
| 59 | Add Watcher.WatchOnce() | 10min | MEDIUM | Task 58 |
| 60 | Test WatchOnce | 5min | MEDIUM | Task 59 |

**Table 7 Total:** ~35 minutes

---

## Summary

| Table | Tasks | Est. Time | Priority |
|-------|-------|-----------|----------|
| 1: Race Fix | 12 | 53min | CRITICAL |
| 2: Context Fix | 4 | 25min | CRITICAL |
| 3: Cleanup & Tag | 7 | 35min | HIGH |
| 4: Event Enhance | 10 | 55min | MEDIUM |
| 5: Filter Enhance | 13 | 75min | MEDIUM |
| 6: Advanced Features | 7 | 50min | MEDIUM |
| 7: Middleware | 5 | 35min | MEDIUM |

**Total:** 58 tasks, ~328 minutes (~5.5 hours)

**Critical Path:** Tables 1-3 (113 minutes) → Release v0.1.0

**Full Implementation:** All tables (328 minutes) → Release v2.0.0

---

## Execution Strategy

### Option A: Minimal Viable Release (2 hours)

Execute Tables 1-3 only:
- Fix race condition
- Tag v0.1.0
- Ship stable release

### Option B: Feature-Complete Release (6 hours)

Execute Tables 1-7:
- Fix race condition
- Add all medium priority features
- Tag v2.0.0
- Comprehensive release

### Option C: Staged Release (Recommended)

1. **Today** (2 hours): Tables 1-3 → v0.1.0
2. **This week** (4 hours): Tables 4-7 → v2.0.0

---

**Prepared for execution.**
