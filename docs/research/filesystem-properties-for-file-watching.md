# Filesystem Properties That Matter for File Watching

**Created:** 2026-07-28
**Status:** Research / Reference
**Scope:** Catalogue every filesystem property that affects the correctness, reliability,
and performance of event-driven file watching. Each property is assessed for its impact on
go-filewatcher and its current handling status.

---

## Table of Contents

1. [Case Sensitivity](#1-case-sensitivity)
2. [Unicode Normalization](#2-unicode-normalization)
3. [Event Backend Support](#3-event-backend-support)
4. [Timestamp Resolution](#4-timestamp-resolution)
5. [Atomicity of `rename()`](#5-atomicity-of-rename)
6. [Inode / File Identity](#6-inode--file-identity)
7. [Path Length & Filename Restrictions](#7-path-length--filename-restrictions)
8. [File Locking Semantics](#8-file-locking-semantics)
9. [Event Coalescing / Merging](#9-event-coalescing--merging)
10. [FSEvents Historical / Retroactive Nature](#10-fsevents-historical--retroactive-nature)
11. [Network Filesystem Disconnection](#11-network-filesystem-disconnection)
12. [Mount Boundaries & OverlayFS Layers](#12-mount-boundaries--overlayfs-layers)
13. [`IN_CLOSE_WRITE` vs `IN_MODIFY`](#13-in_close_write-vs-in_modify)
14. [Trailing Slashes & Path Canonicalization](#14-trailing-slashes--path-canonicalization)
15. [Summary: Priority Matrix](#15-summary-priority-matrix)

---

## 1. Case Sensitivity

### The Property

Some filesystems treat `File.txt` and `file.txt` as the same file (case-insensitive),
others as distinct files (case-sensitive). A third category preserves case for display
but compares case-insensitively.

### Filesystem Behavior Matrix

| Filesystem                  | Case Behavior   | Preserves Case      | Notes                     |
| --------------------------- | --------------- | ------------------- | ------------------------- |
| ext2 / ext3 / ext4          | Sensitive       | Yes                 | Default Linux             |
| XFS                         | Sensitive       | Yes                 | RHEL default              |
| Btrfs                       | Sensitive       | Yes                 |                           |
| ZFS                         | Sensitive       | Yes                 |                           |
| APFS (default)              | **Insensitive** | Yes                 | macOS default since 10.13 |
| APFS (case-sensitive)       | Sensitive       | Yes                 | Optional format           |
| HFS+ (default)              | **Insensitive** | Yes                 | macOS pre-APFS            |
| HFS+ / HFSX                 | Sensitive       | Yes                 | Optional format           |
| NTFS                        | **Insensitive** | Yes                 | Windows default           |
| FAT12 / FAT16 / FAT32       | **Insensitive** | **No** (uppercases) | Floppy, USB, SD           |
| exFAT                       | **Insensitive** | Yes                 | Large USB, SDXC           |
| CIFS / SMB (Windows server) | **Insensitive** | Yes                 |                           |
| CIFS / SMB (Samba)          | Configurable    | Yes                 | `case sensitive = yes/no` |
| NFS                         | Follows server  | Yes                 |                           |

### Impact on File Watching

Case sensitivity affects five internal operations:

1. **Watch deduplication** — adding `/dir/MyFile` and `/dir/myfile` on NTFS creates two
   watches for the same inode, wasting inotify budget.
2. **`Remove()` subtree matching** — calling `Remove("/dir/MyFile")` must match and remove
   watches for `/dir/myfile/subdir` on case-insensitive filesystems.
3. **Path exclusion** — `WithExcludePaths("/Build")` must catch events from `/build/output`.
4. **Debounce keying** — events for `/File.go` and `/file.go` must coalesce on
   case-insensitive filesystems to avoid duplicate processing.
5. **Self-heal** — retrying a failed path with different case must detect "already watched"
   and not create a duplicate.

### Status in go-filewatcher

**IMPLEMENTED (v2.4.0).** `WithCaseSensitivity(mode)` with `pathKey()` canonicalization.
See `filesystem.go`.

**Remaining gaps:**

- Poll loop (`watcher_poll.go`) does not use `pathKey()` for snapshot comparison.
- Gitignore matcher (`watcher_gitignore.go`) does not use `pathKey()` for prefix matching.
- User-facing filters (`filter.go`) do not use `pathKey()` — this is a design question
  (filters are `func(Event) bool` with no watcher reference).

---

## 2. Unicode Normalization

### The Property

Unicode allows the same visual character to be represented by different byte sequences.
`café` can be encoded as:

- **NFC** (composed): `c`, `a`, `f`, `é` (4 code points, 5 bytes in UTF-8)
- **NFD** (decomposed): `c`, `a`, `f`, `e`, `◌́` (5 code points, 6 bytes in UTF-8)

Both render identically, but their byte representations differ. Filesystems choose one
normalization form and enforce it on write.

### Filesystem Behavior Matrix

| Filesystem         | Normalization                       | Impact                                                        |
| ------------------ | ----------------------------------- | ------------------------------------------------------------- |
| **APFS**           | NFD enforced                        | `café.txt` (NFC input) stored as NFD                          |
| **HFS+**           | HFS+ variant of NFD (non-standard!) | Apple's own broken variant; subtly different from Unicode NFD |
| ext4 / XFS / Btrfs | None (raw bytes)                    | Stores whatever bytes the application writes                  |
| NTFS               | None (UTF-16 as-is)                 | Stores whatever the application writes                        |
| FAT / exFAT        | None                                | OS-dependent translation layer                                |

### Impact on File Watching

This is arguably **worse than case sensitivity** because it's invisible:

1. **Path mismatch**: User configures `WithExcludePaths("/home/user/café")` (typed as NFC).
   The filesystem stores it as NFD. Events arrive with NFD paths. The NFC exclude path
   never matches.
2. **Debounce failure**: Two events for the same file arrive — one with NFC path (from
   user action), one with NFD path (from filesystem normalization). They don't match as
   debounce keys, producing duplicate processing.
3. **Gitignore bypass**: `.gitignore` entry written as NFC, path compared as NFD — match
   fails, gitignore rules silently ignored.
4. **Cross-platform inconsistency**: A file created on Linux (raw bytes, NFC) synced to
   macOS (NFD enforced) has a different path byte sequence. Tools that compare paths
   byte-for-byte break.

### The macOS-Specific Trap

macOS is the worst offender. HFS+ uses a **non-standard** normalization table that differs
from official Unicode NFD in edge cases (certain Hangul compatibility characters, old
Greek combining marks). APFS switched to standard NFD but the damage is done: files
migrated from HFS+ may retain non-standard normalization. This means two macOS machines
can disagree on the byte representation of the same filename.

### Status in go-filewatcher

**NOT HANDLED.** `pathKey()` only lowercases. An NFC `é` (U+00E9) lowercased is still
NFC. An NFD `e` + `◌́` (U+0065 + U+0301) lowercased is still NFD. They don't match.

### Recommended Fix

Normalize all path keys to **NFC** using `golang.org/x/text/unicode/norm`:

```go
import "golang.org/x/text/unicode/norm"

func (w *Watcher) pathKey(path string) string {
    normalized := norm.NFC.String(path) // canonical composed form
    if w.effectiveCaseSensitivity == CaseInsensitive {
        return strings.ToLower(normalized)
    }
    return normalized
}
```

**Priority: P0.** This is the next correctness fix after case sensitivity.

---

## 3. Event Backend Support

### The Property

Event-driven file watching relies on OS kernel mechanisms. Not all filesystems are
supported by these mechanisms. When a filesystem doesn't fire events, the watcher produces
**zero output with zero errors** — the worst possible failure mode.

### Backend Support Matrix

| Backend                           | Filesystems with Event Support         | Filesystems WITHOUT Events                           |
| --------------------------------- | -------------------------------------- | ---------------------------------------------------- |
| **Linux inotify**                 | ext2/3/4, XFS, Btrfs, ZFS, F2FS, tmpfs | NFS, CIFS, most FUSE, `/proc`, `/sys`                |
| **macOS FSEvents**                | APFS, HFS+                             | Network mounts (unreliable), some FUSE               |
| **BSD kqueue** (`evfilt_vnode`)   | UFS, UFS2, ZFS, ext2                   | NFS, most FUSE                                       |
| **Windows ReadDirectoryChangesW** | NTFS, FAT32, exFAT, ReFS               | Network shares (unreliable), mapped drives (delayed) |

### Impact on File Watching

A user watching an NFS mount on Linux with default settings gets:

- `Watch()` returns successfully (no error — inotify doesn't know it's NFS)
- Zero events forever
- No diagnostic, no warning, no fallback

This is the #1 user-reported issue with file watching libraries across all languages.

### Status in go-filewatcher

**PARTIALLY HANDLED.** `WithPolling(true)` provides a fallback. But:

- The user must **manually know** to enable polling. There's no auto-detection.
- There's no diagnostic mode that detects "0 events after N seconds on a non-trivial
  directory tree" and suggests enabling polling.
- The poll loop supplements inotify; it doesn't replace it. On pure-NFS setups, both
  run, wasting resources.

### Recommended Improvements

1. **Dead-watch detection**: After 30 seconds with zero events on a directory known to
   have files, log a warning suggesting `WithPolling(true)`.
2. **Auto-polling for known network filesystems**: Detect NFS/CIFS mount types from
   `/proc/mounts` (Linux) or `mount` output and auto-enable polling.
3. **Document the matrix**: Add a troubleshooting table showing which option to enable
   per filesystem type.

---

## 4. Timestamp Resolution

### The Property

`os.FileInfo.ModTime()` returns the file's last modification time, but the resolution
(precision) of that timestamp varies wildly by filesystem. The poll loop uses `ModTime`
to detect changes — if the resolution is coarser than the edit interval, changes are
missed.

### Filesystem Behavior Matrix

| Filesystem              | mtime Resolution                | Practical Impact                    |
| ----------------------- | ------------------------------- | ----------------------------------- |
| FAT16 / FAT32           | **2 seconds**                   | Rapid edits within 2s are invisible |
| exFAT                   | 10ms                            | Usually fine                        |
| ext3 (without `iatime`) | 1 second                        | Edits within 1s may be missed       |
| ext4 / XFS / Btrfs      | 1 nanosecond                    | Excellent                           |
| APFS                    | 1 nanosecond                    | Excellent                           |
| NTFS                    | 100 nanoseconds                 | Excellent                           |
| HFS+                    | 1 second                        | Same 1s gap as ext3                 |
| NFS v3                  | 1 second (server-dependent)     | Unreliable for fast edits           |
| NFS v4                  | Sub-second (if server supports) | Better, but not guaranteed          |
| SMB / CIFS              | Server-dependent                | Often coarse (1s)                   |

### Impact on File Watching

The poll loop in `watcher_poll.go:108` compares:

```go
if curState.modTime != prevState.modTime || curState.size != prevState.size {
    w.pollEmitEvent(ctx, Write, path, curState, eventCh)
}
```

On FAT32, writing a file twice within 2 seconds:

- `modTime` doesn't change (2-second resolution)
- `size` may not change if the second write is the same length
- Result: **zero Write events** for the second edit

This is a **silent data loss** — the file changed, but the watcher doesn't know.

### Status in go-filewatcher

**NOT HANDLED.** The poll loop trusts `ModTime` at face value.

### Recommended Fix

For filesystems with known coarse timestamps, the poll loop should also compare:

1. **File content hash** (expensive but definitive — only when mtime+size are unchanged
   and the filesystem is known to have coarse timestamps)
2. **inode change time** (`ctime`) — tracks metadata changes, not just data writes
3. **A minimum inter-poll interval** that's longer than the filesystem's timestamp
   resolution (e.g., don't poll faster than every 2 seconds on FAT32)

### Detection

`syscall.Statfs` (Linux) or `unix.Statfs` can identify the filesystem type. Combined
with a known-resolution table, the poll loop could adapt dynamically.

---

## 5. Atomicity of `rename()`

### The Property

`rename()` is atomic on most local filesystems: at no point does the file exist under both
names or neither name. But on network and legacy filesystems, rename may decompose into
copy + delete, producing different event sequences.

### Filesystem Behavior Matrix

| Filesystem                        | `rename()` Atomic? | Event Sequence                                                |
| --------------------------------- | ------------------ | ------------------------------------------------------------- |
| ext4, XFS, Btrfs, ZFS, APFS, NTFS | ✅ Atomic          | Clean Rename event                                            |
| NFS                               | ⚠️ Not guaranteed  | May appear as CREATE + DELETE, especially after reconnection  |
| SMB / CIFS                        | ⚠️ Not guaranteed  | Often CREATE + DELETE, especially cross-directory             |
| FAT / exFAT                       | ❌ Not atomic      | Always copy + delete semantics                                |
| Cross-filesystem (any → any)      | ❌ Never atomic    | POSIX `rename` fails `EXDEV`; apps fall back to copy + delete |

### Impact on File Watching

On reliable local filesystems, `convertEvent()` maps `fsnotify.Rename` to the `Rename` Op.
The user's handler can assume a clean rename.

On network filesystems, the same logical operation produces:

- A `Create` for the new path
- A `Remove` for the old path
- **No `Rename` event at all**

Worse, on a flaky network, you might see only the Create (file appeared) and never the
Remove (old path not cleaned up), or vice versa. The watcher has no way to correlate
them.

### Status in go-filewatcher

**NOT HANDLED.** The watcher trusts the backend's event classification. No logic to
detect "this Remove + that Create with the same inode = Rename."

### Recommended Fix (Complex)

Inode-based rename correlation:

1. Maintain a path → inode map in the poll loop
2. When a Create and Remove happen close together with the same inode, emit a Rename
3. This requires inode tracking, which is a significant architecture change

For now, **documentation** is the pragmatic fix: warn users that rename events are
unreliable on network filesystems and suggest correlating Create + Remove pairs in their
handler.

---

## 6. Inode / File Identity

### The Property

A file has two identities: its **path** (what the user sees) and its **inode** (what the
filesystem tracks internally). Most files have a 1:1 path-to-inode mapping. But hard
links break this: multiple paths point to the same inode. Copy-on-write and reflinks
create subtler identity ambiguities.

### Filesystem Behavior Matrix

| Feature                 | Filesystems                                                      | Behavior                                   |
| ----------------------- | ---------------------------------------------------------------- | ------------------------------------------ |
| Hard links              | ext4, XFS, Btrfs, ZFS, APFS, NTFS (hardlinks within same volume) | Multiple paths, same inode, same data      |
| Copy-on-write snapshots | Btrfs, ZFS, APFS                                                 | Snapshot files share inodes until modified |
| Inline dedup            | ZFS, Btrfs, NetApp                                               | Different inodes, identical block content  |
| Reflinks                | XFS (`copy_file_range`), Btrfs, APFS                             | Instant copy shares data until written     |

### Impact on File Watching

A file watcher tracks **paths**, but the filesystem tracks **inodes**. This creates three
problems:

1. **Hard link ambiguity**: Writing through `/dir/A/file.txt` also modifies
   `/dir/B/file.txt` if they're hard-linked. inotify reports the event on the watched path
   only — the other path's watcher sees nothing.
2. **Snapshot events**: Creating a Btrfs snapshot doesn't fire events. But modifying a
   file _inside_ a snapshot fires on the snapshot path, not the original.
3. **Reflink confusion**: A reflink copy shares data blocks. A watcher might see two files
   with identical content and size but different paths. Deduplication logic that assumes
   1:1 path-to-content mapping breaks.

### Status in go-filewatcher

**NOT HANDLED.** Events are purely path-based. No inode tracking anywhere in the
codebase.

This is **correct for 99% of use cases** — most applications don't use hard links. But
it's worth documenting as a known limitation.

---

## 7. Path Length & Filename Restrictions

### The Property

Different filesystems impose different limits on total path length and individual filename
length. They also reserve certain names.

### Filesystem Behavior Matrix

| Filesystem | Max Path                     | Max Filename               | Illegal Characters                           | Reserved Names                                 |
| ---------- | ---------------------------- | -------------------------- | -------------------------------------------- | ---------------------------------------------- |
| ext4       | 4096 bytes                   | 255 bytes                  | `/`, `NUL`                                   | None                                           |
| XFS        | 4096 bytes                   | 255 bytes                  | `/`, `NUL`                                   | None                                           |
| NTFS       | 32,767 chars (with `\\?\`)   | 255 chars                  | `\`, `/`, `:`, `*`, `?`, `"`, `<`, `>`, `\|` | `CON`, `PRN`, `AUX`, `NUL`, `COM1-9`, `LPT1-9` |
| FAT32      | 260 chars                    | 255 chars                  | Same as NTFS                                 | Same as NTFS                                   |
| exFAT      | 32,760 chars                 | 255 chars                  | Same as NTFS                                 | Same as NTFS                                   |
| APFS       | ~1024 chars (NFD-normalized) | 255 chars (NFD-normalized) | `:` (historically), `NUL`                    | None                                           |
| HFS+       | ~1024 chars                  | 255 chars                  | `:` (converted to `/` in classic Mac OS)     | None                                           |
| CIFS / SMB | Server-dependent             | 255 chars                  | Server-dependent                             | Server-dependent                               |

### Impact on File Watching

1. **Windows reserved names**: A file named `CON.txt` is valid on ext4 but **impossible**
   on NTFS. If a watcher walks a directory tree synced from Linux to Windows, the walk
   silently fails on these names.
2. **Trailing dots/spaces on Windows**: Windows strips trailing dots and spaces from
   filenames (`file.` → `file`). A path compared before and after this normalization
   doesn't match.
3. **Path length overflow**: A path within the 4096-byte ext4 limit may exceed Windows'
   260-char `MAX_PATH` limit (unless long-path support is enabled). The watcher's `Add()`
   call fails with no clear diagnostic.
4. **Illegal characters**: Cross-platform sync tools may create files with `:` in the name
   on Linux. On macOS, `:` is historically converted to `/` (and vice versa in Terminal).
   This causes path confusion.

### Status in go-filewatcher

**NOT HANDLED.** No filename validation. This is acceptable for a library — it's the
application's concern — but path-length limits should be documented, especially for
recursive watching where path length grows with depth.

---

## 8. File Locking Semantics

### The Property

When a process opens a file, the OS may prevent other processes from reading or writing
it. The locking model varies by OS and filesystem.

### Filesystem Behavior Matrix

| Platform / Filesystem   | Locking Model                   | Impact                                                       |
| ----------------------- | ------------------------------- | ------------------------------------------------------------ |
| POSIX (ext4, XFS, etc.) | **Advisory** (`flock`, `fcntl`) | File is always readable; locks are cooperative               |
| Windows / NTFS          | **Mandatory** (by default)      | File may be **unreadable** while another process has it open |
| SMB / CIFS              | **Mandatory** (oplocks)         | Same as Windows, plus network oplock semantics               |
| NFS                     | NLM (Network Lock Manager)      | Advisory in practice; stale locks common on disconnect       |

### Impact on File Watching

`WithContentHashing()` reads the file to compute SHA-256. The behavior changes by
platform:

- **Linux**: `os.Open(file)` succeeds even if another process is writing. The hash is
  computed on a **partially-written file** — the hash is valid but meaningless.
- **Windows**: `os.Open(file)` may **fail with access denied** if another process has the
  file open exclusively (common for log files, databases, Office documents). The hash
  computation fails silently (returns empty string).
- **NFS**: Stale locks from disconnected clients may cause `open()` to hang or fail
  unpredictably.

### Status in go-filewatcher

**NOT HANDLED.** `hashFileContents` in `watcher_internal.go:368` swallows all errors and
returns empty string. The user has no way to distinguish:

- Hash is empty because the file was already removed (expected)
- Hash is empty because the file is locked (unexpected — Windows)
- Hash is empty because of a permission error (unexpected)
- Hash is empty because the path is a directory (expected)

### Recommended Fix

Return a typed error from `hashFileContents` that the caller can inspect:

```go
type HashError struct {
    Reason  HashErrorReason // HashErrorLocked, HashErrorPermission, HashErrorNotFound
    Op      string
    Path    string
    Err     error
}
```

---

## 9. Event Coalescing / Merging

### The Property

OS event backends don't guarantee 1:1 correspondence between filesystem operations and
events. Multiple operations may be merged, reordered, or split.

### Backend Behavior Matrix

| Backend                           | Coalescing Behavior                                                                                                             | Impact                                      |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| **Linux inotify**                 | Multiple writes within the kernel event buffer are merged into one `IN_MODIFY`. A `Create + Write` may arrive as just `Create`. | Fewer events than operations                |
| **macOS FSEvents**                | Events are **time-batched** (typically 1-second latency by design). Events are historical, not real-time.                       | Sub-second debounce is meaningless on macOS |
| **Windows ReadDirectoryChangesW** | Multiple operations may be batched in a single buffer read. Renames always produce `OLD_NAME + NEW_NAME` pairs.                 | Rename detection requires pairing logic     |
| **BSD kqueue**                    | Only reports changes since last `kevent` call. If the queue overflows, events are **dropped**.                                  | Silent event loss under load                |

### Impact on File Watching

The watcher's debounce and event-counting logic assumes it receives individual events.
But the backend may have already merged or dropped events before they reach the watcher:

1. **Merged writes**: An editor that does open → write → write → write → close may produce
   a single `Write` event (or none if the creates dominates).
2. **FSEvents latency floor**: macOS has an inherent ~1-second latency. A `WithDebounce(100ms)`
   setting has no effect — events arrive at most once per second anyway.
3. **kqueue overflow**: Under high load (thousands of rapid changes), the kqueue event
   queue can overflow. Events are silently dropped. The only signal is an `ENOBUFS` error
   on the next `kevent` call.

### Status in go-filewatcher

**PARTIALLY HANDLED.** Debounce coalesces events. But:

- No documentation of backend-specific latency floors
- No detection or handling of kqueue overflow
- No documentation that event counts are inherently non-deterministic

### Recommended Fix

**Documentation.** Add a "Backend Behavior" section to Troubleshooting.md that explains
these differences. Code-level mitigation is limited because the behavior is baked into
the OS kernel.

---

## 10. FSEvents Historical / Retroactive Nature

### The Property

macOS FSEvents is fundamentally different from Linux inotify. Rather than streaming
real-time events, it records a **journal of changes** and delivers them in batches.

### How FSEvents Works

1. The kernel records changes to a journal, keyed by event ID (a monotonic integer).
2. `FSEventStreamCreate` lets you ask "what changed since event ID X?"
3. Events are delivered at the **directory level** — FSEvents tells you
   "directory `/foo` changed," not "file `/foo/bar.txt` was modified."
4. The application must then `stat` or `readdir` to figure out what actually changed.
5. Events may arrive **seconds** after the actual change.
6. If your application wasn't running, you can query past events using "since event ID"
   — a capability unique to FSEvents.

### Impact on File Watching

fsnotify abstracts FSEvents, but the abstraction leaks:

1. **Coarse granularity**: FSEvents reports directory-level changes. fsnotify may deliver
   a single event for `/foo` when any file inside changed. The watcher's `handleNewDirectory`
   and per-path debounce logic may behave unexpectedly.
2. **High latency**: With default `latency=1.0s`, events arrive 1+ seconds after the
   change. Sub-second debounce settings have no effect.
3. **No per-file granularity**: If 100 files change in one directory within the latency
   window, FSEvents may deliver a single "directory changed" event. fsnotify expands this
   to individual file events by reading the directory, but this adds further latency.
4. **Historical events**: FSEvents can report events that happened while the watcher wasn't
   running. fsnotify does not expose this, but it's a unique capability worth considering
   for a "catch up on missed changes" feature.

### Status in go-filewatcher

**NOT HANDLED / NOT DOCUMENTED.** No macOS-specific documentation. No mitigation for
coarse granularity or high latency.

### Recommended Fix

1. **Document** the FSEvents behavior in Troubleshooting.md.
2. Consider a `WithPolling` auto-suggestion on macOS when sub-second debounce is
   configured (polling is more responsive than FSEvents for small directories).
3. Consider exposing the "since event ID" capability as a `CatchUp(since EventID)` method
   in a future v3 API.

---

## 11. Network Filesystem Disconnection

### The Property

Network filesystems (NFS, SMB/CIFS, SSHFS, 9P) can disconnect. When they do, filesystem
operations on mounted paths **hang** for varying durations before timing out or returning
errors.

### Filesystem Behavior Matrix

| Filesystem           | Disconnect Behavior                                       | Default Timeout |
| -------------------- | --------------------------------------------------------- | --------------- |
| **NFS (hard mount)** | All operations **hang indefinitely** until reconnect      | None (infinite) |
| **NFS (soft mount)** | Operations retry for `timeo × retrans`, then return error | ~60s            |
| **SMB / CIFS**       | Operations hang until reconnect or TCP timeout            | 15-60s          |
| **SSHFS**            | All operations hang until SSH timeout                     | 10-30s          |
| **9P (WSL2)**        | Usually reliable; drops on WSL restart                    | Immediate error |

### Impact on File Watching

When a network filesystem disconnects, every `os.Stat()` call inside the watcher blocks:

1. **`convertEvent`** (`watcher_internal.go:340`): Calls `os.Stat(fsEvent.Name)` to
   populate `IsDir`, `Size`, `ModTime`. On a disconnected NFS mount, this **hangs the
   entire event processing loop**.
2. **`handleNewDirectory`** (`watcher_internal.go:228`): Calls `os.Stat(path)` to check
   if a new path is a directory. Same hang.
3. **Poll loop** (`watcher_poll.go:60`): `filepath.WalkDir` calls `os.Stat` on every file.
   A disconnected mount hangs the poll loop indefinitely.
4. **`hashFileContents`** (`watcher_internal.go:368`): `os.Open` + `io.Copy` on a
   disconnected mount hangs.

The user sees a **frozen watcher** with no error message. Events stop flowing. The only
recovery is to kill and restart the process.

### Status in go-filewatcher

**NOT HANDLED.** All filesystem operations are synchronous with no timeout or context
cancellation. There is no mechanism to detect or recover from a hung network filesystem.

### Recommended Fix

1. **Context-aware stat**: Replace `os.Stat(path)` with a goroutine + `select` on
   `ctx.Done()`:

   ```go
   func statWithTimeout(ctx context.Context, path string, timeout time.Duration) (os.FileInfo, error) {
       type result struct {
           info os.FileInfo
           err  error
       }
       ch := make(chan result, 1)
       go func() {
           info, err := os.Stat(path)
           ch <- result{info, err}
       }()
       select {
       case r := <-ch:
           return r.info, r.err
       case <-time.After(timeout):
           return nil, fmt.Errorf("stat %q timed out after %v (possible network filesystem disconnection)", path, timeout)
       case <-ctx.Done():
           return nil, ctx.Err()
       }
   }
   ```

2. **Circuit breaker for stat calls**: If N consecutive stat calls time out, assume the
   filesystem is down and pause event processing until it recovers.
3. **Document** that watching network filesystems requires `WithPolling(true)` and a
   well-configured `WithPollInterval` that matches the network's expected latency.

---

## 12. Mount Boundaries & OverlayFS Layers

### The Property

Modern systems use layered filesystems (OverlayFS, union mounts, bind mounts). These
introduce complexities for file watching because the same logical path can be backed by
multiple physical layers.

### Scenario Matrix

| Scenario                            | Impact on File Watching                                                                                                                                                                                                          |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **OverlayFS (Docker containers)**   | Events fire on the **upper layer** (writable container layer) only. Reads from the lower layer (image) don't trigger events. Modifications to a file in the lower layer that gets copied-up produce a Create on the upper layer. |
| **Bind mounts**                     | inotify watches the underlying inode. Events on the bind mount path may or may not fire depending on kernel version. On older kernels, watching a bind mount path doesn't see events from the original path.                     |
| **autofs**                          | Directory may not exist (not yet mounted) when `Add()` is called. The watch fails. When the filesystem auto-mounts later, no watch is registered.                                                                                |
| **Submounts / cross-mount walking** | `filepath.WalkDir` does cross mount points by default. But inotify watches are per-inode; if a submount replaces the inode, the watch becomes stale.                                                                             |
| **FUSE passthrough**                | FUSE filesystems decide whether to implement inotify support. Most don't, making event-based watching impossible without polling.                                                                                                |

### Impact on File Watching — Docker / Kubernetes (Most Common Case)

Docker containers and Kubernetes pods heavily use OverlayFS. Watching `/app` in a
container means watching the overlay merged view:

1. **Image layer modifications** (lower layer): Don't fire events. If the base image is
   updated and the container is recreated, the watcher sees no events for the changed
   files — they simply appear on the first walk.
2. **Container writes** (upper layer): Fire events normally.
3. **Copy-up behavior**: When a file from the image is modified, OverlayFS copies it to
   the upper layer. The watcher sees this as a `Create` event, not a `Write` — because a
   new file appears in the watched (upper) layer.
4. **Volume mounts** (`-v host:container`): These are bind mounts. Events work normally
   on Linux (same inode). On macOS Docker Desktop, they go through osxfs/gRPC-FUSE, which
   has its own event limitations.

### Status in go-filewatcher

**NOT HANDLED.** No mount-boundary awareness. The watcher doesn't know whether a path is
an overlay, bind mount, or autofs.

**Partial mitigation:**

- `WithSelfHeal(interval)` partially mitigates autofs: failed watches are retried
  periodically. When the filesystem mounts, the retry succeeds.
- `WithFollowSymlinks(true)` helps with some symlink-based mount indirection.

### Recommended Fix

1. **Document** OverlayFS behavior in Troubleshooting.md — this is a common Docker/K8s
   confusion point.
2. Consider detecting mount type from `/proc/self/mountinfo` (Linux) and warning if the
   path is on an OverlayFS upper layer.
3. Consider auto-enabling self-heal when the path is on autofs.

---

## 13. `IN_CLOSE_WRITE` vs `IN_MODIFY`

### The Property

Linux inotify has two events related to file writing:

- **`IN_MODIFY`**: Fires on every `write()` system call. A file opened, written 3 times,
  and closed fires 3 `IN_MODIFY` events.
- **`IN_CLOSE_WRITE`**: Fires when a file that was opened for writing is **closed**.
  A file opened, written 3 times, and closed fires 1 `IN_CLOSE_WRITE` event.

### The Difference

| Event            | When                                         | File Safe to Read?                     | Use Case                        |
| ---------------- | -------------------------------------------- | -------------------------------------- | ------------------------------- |
| `IN_MODIFY`      | After each `write()` call                    | **No** — file may be partially written | Real-time monitoring, live tail |
| `IN_CLOSE_WRITE` | After `close()` on a file opened for writing | **Yes** — writer is done               | Build triggers, post-processing |

### Impact on File Watching

A text editor saving a file does: `open(WRONLY) → write(data) → close()`. The watcher
receives:

- `IN_MODIFY` (possibly multiple): File is mid-write — reading it now may return
  truncated or partial content.
- `IN_CLOSE_WRITE`: File is complete — safe to read, hash, or process.

If a watcher triggers a build on `Write` (mapped from `IN_MODIFY`), the build may read a
half-written file. This is a **common, hard-to-diagnose bug** in file-watcher-triggered
build systems.

### Status in go-filewatcher

**NOT HANDLED.** The `Write` Op maps to `fsnotify.Write`, which corresponds to
`IN_MODIFY`. There is no `CloseWrite` Op.

This is a known limitation of the fsnotify abstraction. fsnotify does not expose
`IN_CLOSE_WRITE` as a separate event type.

### Recommended Fix

This requires either:

1. **A new Op type** (`CloseWrite`) — but fsnotify doesn't expose it, so go-filewatcher
   would need to use `syscall.InotifyAddWatch` directly with the `IN_CLOSE_WRITE` mask,
   bypassing fsnotify. This is a significant change.
2. **Documentation** — warn users that `Write` means "file is being modified" not "file
   modification is complete." Recommend debounce with a sufficiently long window to allow
   the writer to finish.
3. **Heuristic** — after a `Write` event, wait for a quiet period (no new events for the
   same file for N ms), then treat the file as "settled." This is what debounce already
   does, but with a different framing.

---

## 14. Trailing Slashes & Path Canonicalization

### The Property

Paths can be represented in multiple equivalent forms: with or without trailing slashes,
with or without `.` / `..` components, with redundant separators, or with symlinks
unresolved.

### Platform Behavior

| Platform    | Behavior                                                                                                                                     |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **POSIX**   | `/foo/` and `/foo` refer to the same path. `filepath.Clean` strips trailing `/`. `/foo/../bar` normalizes to `/bar`.                         |
| **Windows** | `C:\foo\` and `C:\foo` are the same. But `C:` (drive current directory) ≠ `C:\` (drive root). Forward slashes are normalized to backslashes. |
| **macOS**   | Same as POSIX. Finder sometimes adds trailing slashes in AppleScript paths.                                                                  |

### Impact on File Watching

Path comparison bugs arise when the same path is represented differently:

1. **`Add()` / `Remove()` mismatch**: User calls `Add("/home/user")` then later
   `Remove("/home/user/")`. The trailing slash causes the prefix match to fail:
   `strings.HasPrefix("/home/user/file.txt", "/home/user/")` is true, but
   `strings.HasPrefix("/home/user/file.txt", "/home/user")` is also true — however,
   the exact match `"home/user" == "/home/user/"` is false, so the root path itself
   is not removed.
2. **Event path vs watch path**: The watch was registered on `/home/user` but the event
   arrives for `/home/user/file.txt`. This works for prefix matching but not for exact
   path matching in filters like `FilterExcludePaths`.
3. **Symlink resolution**: A user might watch `/var/www` which is a symlink to
   `/srv/www`. Events arrive with `/srv/www/file.txt` (resolved path) but the watcher
   compares against `/var/www/file.txt` (registered path). This mismatch breaks prefix
   matching and debounce keying.
4. **Relative vs absolute**: `Add("./src")` and events for `/home/user/project/src/file.go`
   don't match unless paths are normalized to absolute first.

### Status in go-filewatcher

**PARTIALLY HANDLED.**

- `New()` calls `filepath.Abs()` on all paths, converting relative to absolute.
- `withResolvedPath()` (used by `Add`, `Remove`) calls `filepath.Abs()`.
- `FilterExcludePaths` calls `filepath.Abs()` on configured paths.

**NOT HANDLED:**

- `filepath.Clean()` is **not** called anywhere. A path like `/home/user/../user/file.txt`
  is stored as-is and compared as-is.
- Symlink resolution (`filepath.EvalSymlinks`) is **not** called during path normalization
  (only during walk if `WithFollowSymlinks(true)` is set). This means the registered path
  and the event path may differ if one is a symlink and the other isn't.
- Trailing slashes are not stripped. `filepath.Abs` does not strip them.

### Recommended Fix

Add a `normalizePath(path string) string` function that:

```go
func normalizePath(path string) string {
    abs, err := filepath.Abs(path)
    if err != nil {
        abs = path
    }
    return filepath.Clean(abs)
}
```

Call this in `New()`, `Add()`, `Remove()`, `AddRecursive()`, and all filter constructors
that accept paths.

---

## 15. Summary: Priority Matrix

Ranked by **impact on correctness × likelihood of encountering the issue**:

| Priority | Property                                      | Current Status                             | Effort                  | Affected Users           |
| -------- | --------------------------------------------- | ------------------------------------------ | ----------------------- | ------------------------ |
| 🔴 P0    | Unicode normalization (NFC/NFD)               | Not handled                                | Medium                  | All macOS users          |
| 🔴 P0    | Network FS `stat()` hangs                     | Not handled                                | Low-Medium              | NFS/SMB/SSHFS users      |
| 🟡 P1    | Poll loop case-awareness                      | Not handled (gap in case-sensitivity impl) | Low                     | macOS + polling users    |
| 🟡 P1    | Gitignore case-awareness                      | Not handled (gap in case-sensitivity impl) | Low                     | macOS + gitignore users  |
| 🟡 P1    | Timestamp resolution (FAT 2s gap)             | Not handled                                | Low                     | USB/SD card users        |
| 🟡 P1    | Event backend auto-detection                  | Partial (manual polling exists)            | Medium                  | All NFS/FUSE users       |
| 🟡 P1    | Windows mandatory locking + hashing           | Not handled                                | Low                     | Windows users            |
| 🟡 P1    | Path canonicalization (Clean, trailing slash) | Partially handled                          | Low                     | All platforms            |
| 🟢 P2    | Rename atomicity (network FS)                 | Not handled                                | High                    | NFS/SMB users            |
| 🟢 P2    | FSEvents directory-level + latency            | Not documented                             | Low (docs only)         | macOS users              |
| 🟢 P2    | OverlayFS layer semantics                     | Not handled                                | High (kernel-dependent) | Docker/K8s users         |
| 🟢 P2    | `IN_CLOSE_WRITE` vs `IN_MODIFY`               | Not handled                                | Medium (new Op type)    | Build tool users         |
| ⚪ P3    | Hard link identity                            | Not handled                                | Very high               | Rare (advanced users)    |
| ⚪ P3    | Event coalescing documentation                | Not documented                             | Low (docs only)         | All users (expectations) |
| ⚪ P3    | Path length / reserved names                  | Not handled                                | Low (docs only)         | Windows users            |
| ⚪ P3    | kqueue overflow detection                     | Not handled                                | Medium                  | BSD users under load     |

---

## Appendix: Filesystem Quick Reference

A condensed reference of the most commonly encountered filesystems and their relevant
properties:

| Filesystem     | Case                | Unicode       | mtime Resolution | rename Atomic | Events               | Locking          |
| -------------- | ------------------- | ------------- | ---------------- | ------------- | -------------------- | ---------------- |
| ext4           | Sensitive           | Raw bytes     | 1ns              | Yes           | inotify ✅           | Advisory         |
| XFS            | Sensitive           | Raw bytes     | 1ns              | Yes           | inotify ✅           | Advisory         |
| Btrfs          | Sensitive           | Raw bytes     | 1ns              | Yes           | inotify ✅           | Advisory         |
| ZFS            | Sensitive           | Raw bytes     | 1ns              | Yes           | inotify ✅           | Advisory         |
| APFS (default) | **Insensitive**     | **NFD**       | 1ns              | Yes           | FSEvents ⚠️          | Advisory         |
| HFS+ (default) | **Insensitive**     | **HFS+ NFD**  | 1s               | Yes           | FSEvents ⚠️          | Advisory         |
| NTFS           | **Insensitive**     | None (UTF-16) | 100ns            | Yes           | ReadDirChangesW ✅   | **Mandatory**    |
| FAT32          | **Insensitive**     | None          | **2s**           | **No**        | ReadDirChangesW ✅   | None             |
| exFAT          | **Insensitive**     | None          | 10ms             | **No**        | ReadDirChangesW ✅   | None             |
| NFS v3         | Follows server      | Raw bytes     | **1s**           | ⚠️ Unreliable | **None** ⛔          | NLM (unreliable) |
| SMB/CIFS       | Usually insensitive | None          | Variable         | ⚠️ Unreliable | **None** ⛔          | Mandatory        |
| OverlayFS      | Sensitive           | Raw bytes     | 1ns              | Yes (upper)   | inotify (upper only) | Advisory         |
| tmpfs          | Sensitive           | Raw bytes     | 1ns              | Yes           | inotify ✅           | Advisory         |

Legend: ✅ Full support · ⚠️ Partial / caveats · ⛔ No support

---

_This document is a living reference. Update when new filesystem properties are discovered
or when go-filewatcher adds handling for existing ones._
