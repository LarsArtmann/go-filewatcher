# Branching-Flow Analysis: Comprehensive Improvement Plan

**Date**: 2026-04-10 06:33 CEST  
**Analysis Tool**: branching-flow all . --verbose  
**Project**: go-filewatcher  

---

## Executive Summary

The branching-flow multi-linter analysis identified **6 core issues** across 8 linters. Quality scores range from Good (90/100) to Excellent (100/100). This plan prioritizes fixes by **Customer Value × Impact / Effort** ratio.

---

## Issues Summary Table

| Rank | Issue | Severity | Impact | Effort | Customer Value | Location | Priority |
|------|-------|----------|--------|--------|----------------|----------|----------|
| 1 | **Critical Phantom Types** | Critical | High | Medium | **HIGH** | 5 locations | 🔴 P0 |
| 2 | **Event.Path Phantom Type** | High | High | Medium | **HIGH** | event.go:88 | 🔴 P0 |
| 3 | **Error Context Wrapping** | Medium | Medium | Low | **MEDIUM** | 10 locations | 🟡 P1 |
| 4 | **Watcher Large Struct** | Medium | High | High | **MEDIUM** | watcher.go:44 | 🟡 P1 |
| 5 | **DebounceEntry Mixin** | Low | Low | Low | **LOW** | debouncer.go:11 | 🟢 P2 |
| 6 | **Boolean Blindness** | Medium | Low | Medium | **LOW** | watcher.go:51 | 🟢 P2 |
| 7 | **Medium Phantom Types** | Medium | Medium | Low | **LOW** | 2 locations | 🟢 P2 |
| 8 | **Low Phantom Types (uint)** | Low | Low | Low | **LOW** | 11 locations | ⚪ P3 |

---

## Detailed Task Breakdown (≤12 min each)

### 🔴 P0: Critical Path (Customer-Facing Safety)

#### TASK-1: Fix 5 Critical Phantom Types (10 min)
**Files**: `debouncer.go`, `testing_helpers.go`, `watcher_walk.go`

| # | Violation | Current | New Type |
|---|-----------|---------|----------|
| 1 | debouncer.go:115 | `Debounce(key string, ...)` | `type DebounceKey string` |
| 2 | testing_helpers.go:73 | `assertLogContains(..., substr string)` | `type LogSubstring string` |
| 3 | testing_helpers.go:144 | `createTestFile(..., tmpDir string)` | `type TempDir string` |
| 4 | watcher_walk.go:22 | `addPath(root string)` | `type RootPath string` |
| 5 | watcher_walk.go:34 | `walkAndAddPaths(root string)` | `type RootPath string` |

**Rationale**: Prevents passing wrong string arguments at compile time.

---

#### TASK-2: Add Event.Path Phantom Type (12 min)
**File**: `event.go:88`

| Field | Current | New Type |
|-------|---------|----------|
| Event.Path | `string` | `type FilePath string` |

**Breaking Change**: YES - Public API change  
**Mitigation**: Add `String() string` method for easy conversion

**Rationale**: Core API type safety - prevents mixing path types with arbitrary strings.

---

### 🟡 P1: Important (Developer Experience)

#### TASK-3: Error Context Wrapping - watcher.go (10 min)
**Issues**: 9 error propagation issues

| Line | Current | Suggested Improvement |
|------|---------|----------------------|
| 91 | Context variable 'opts' lost | Add operation context |
| 98 | Context variable 'opts' lost | Add path resolution context |
| 102 | Context variable 'opts' lost | Add validation context |
| 105 | Context variable 'opts' lost | Add directory check context |
| 111 | Context variable 'opts' lost | Add fsnotify context |
| 188 | Context variable 'path' lost | Add "Add()" operation context |
| 197 | Context variable 'path' not included | Wrap addPath error |
| 210 | Context variable 'path' lost | Add "Remove()" operation context |
| 219 | Context variable 'path' lost | Add removal context |

**Rationale**: Better debugging experience with full error chains.

---

#### TASK-4: Error Context Wrapping - watcher_walk.go (5 min)
**Issue**: 1 error propagation issue at line 46

| Line | Current | Improvement |
|------|---------|-------------|
| 46 | Context variable 'd' lost | Add walk context with directory info |

---

#### TASK-5: Fix Watcher Large Struct (12 min)
**Problem**: 17 fields (threshold: 15)

**Proposed Split**:
```go
// WatcherConfig holds user-provided configuration
type WatcherConfig struct {
    Paths           []string
    Recursive       bool
    GlobalDebounce  time.Duration
    PerPathDebounce time.Duration
    SkipDotDirs     bool
    BufferSize      int
    IgnoreDirNames  []string
}

// WatcherState holds mutable runtime state
type WatcherState struct {
    Closed    bool
    Watching  bool
    WatchList []string
}

// Watcher reduces to:
type Watcher struct {
    fswatcher         *fsnotify.Watcher
    config            WatcherConfig
    filters           []Filter
    middleware        []Middleware
    errorHandler      func(error)
    onAdd             func(path string)
    mu                sync.RWMutex
    state             WatcherState
    debounceInterface DebouncerInterface
}
```

**Rationale**: Better separation of concerns, testability, maintainability.

---

### 🟢 P2: Nice-to-Have (Code Quality)

#### TASK-6: Implement DebounceEntry Mixin (8 min)
**Problem**: `debounceEntry` and `GlobalDebouncer` share 2 fields

**Shared Fields**:
- `fn func()`
- `timer *time.Timer`

**Implementation**:
```go
type debounceMixin struct {
    fn    func()
    timer *time.Timer
}

type debounceEntry struct {
    debounceMixin
}

type GlobalDebouncer struct {
    debounceMixin
    delay time.Duration
    mu    sync.Mutex
}
```

---

#### TASK-7: Medium/Low Phantom Types (8 min)
**Medium Priority**:
| Field | Current | New Type |
|-------|---------|----------|
| bufferSize | int | `type BufferSize int` |
| WatchCount | int | `type WatchCount int` |

**Low Priority (uint conversions)**:
| Parameter | Current | New Type |
|-----------|---------|----------|
| FilterMinSize minSize | int | uint |
| WithBufferSize size | int | uint |
| FilterMaxSize maxSize | int | uint |
| etc. | ... | uint |

---

#### TASK-8: Fix Boolean Blindness (10 min)
**Problem**: 4 bool fields = 4 bytes → could be 1 byte with bit flags

**Fields**: `recursive`, `skipDotDirs`, `closed`, `watching`

**Implementation**:
```go
type WatcherFlags byte

const (
    FlagRecursive WatcherFlags = 1 << iota
    FlagSkipDotDirs
    FlagClosed
    FlagWatching
)
```

**Rationale**: Memory optimization, more explicit state representation.

---

### ⚪ P3: Deferred (Low ROI)

#### TASK-9: Remaining uint conversions (5 min)
Convert remaining int parameters to uint where semantically correct (sizes, counts, durations already have proper types).

---

## Quality Score Targets

| Linter | Current | Target | Delta |
|--------|---------|--------|-------|
| Context | 90.0/100 | 95.0/100 | +5 |
| Phantom | 19 violations | 5 violations | -14 |
| BoolBlind | 1 violation | 0 violations | -1 |
| Anti-Patterns | 1 warning | 0 warnings | -1 |
| Mixins | 1 opportunity | 0 (implemented) | -1 |
| **Overall** | **Good** | **Excellent** | **+1 grade** |

---

## Risk Assessment

| Task | Risk Level | Breaking Change | Rollback Strategy |
|------|------------|-----------------|-------------------|
| TASK-1 (Critical Phantoms) | Low | No | Revert type aliases |
| TASK-2 (Event.Path) | Medium | Yes | Keep string alias with methods |
| TASK-3/4 (Error Wrap) | Low | No | String comparison still works |
| TASK-5 (Struct Split) | Medium | Yes | Add backward-compat methods |
| TASK-6 (Mixin) | Low | No | Internal only |
| TASK-7 (Med/Low Phantoms) | Low | No | Type aliases |
| TASK-8 (Bool Flags) | Medium | Yes | Document migration path |

---

## Definition of Done

- [ ] All P0 tasks complete
- [ ] All P1 tasks complete
- [ ] `just check` passes
- [ ] Tests pass with coverage maintained
- [ ] Status report written
- [ ] Detailed commit messages
- [ ] No regressions in public API (unless intentional breaking change)

---

## Time Estimate

| Priority | Tasks | Est. Time |
|----------|-------|-----------|
| P0 | 2 | 22 min |
| P1 | 3 | 27 min |
| P2 | 3 | 26 min |
| P3 | 1 | 5 min |
| **Total** | **9** | **~80 min** |

---

*Plan created by branching-flow analysis on 2026-04-10*
