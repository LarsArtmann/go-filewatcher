package filewatcher

import (
	"runtime"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// FilesystemCaseSensitivity controls whether path comparisons treat character case.
//
// Different filesystems behave differently: NTFS (Windows) and APFS (macOS) are
// case-insensitive by default, while ext4/XFS/btrfs (Linux) are case-sensitive.
// Choosing the wrong mode can cause duplicate watches, missed Remove() calls,
// or inconsistent debouncing on paths that differ only in case.
type FilesystemCaseSensitivity int

const (
	// CaseSensitivityAuto selects the platform default: case-insensitive on
	// Windows and macOS, case-sensitive everywhere else.
	CaseSensitivityAuto FilesystemCaseSensitivity = iota
	// CaseSensitive treats paths that differ only in case as distinct.
	CaseSensitive
	// CaseInsensitive treats paths that differ only in case as identical.
	CaseInsensitive
)

// String returns a human-readable representation of the case-sensitivity mode.
func (c FilesystemCaseSensitivity) String() string {
	switch c {
	case CaseSensitive:
		return "case-sensitive"
	case CaseInsensitive:
		return "case-insensitive"
	case CaseSensitivityAuto:
		return "auto"
	default:
		return categoryStringUnknown
	}
}

// resolveCaseSensitivity returns the effective case-sensitivity mode.
// Auto resolves to case-insensitive on Windows and macOS, case-sensitive elsewhere.
func resolveCaseSensitivity(mode FilesystemCaseSensitivity) FilesystemCaseSensitivity {
	if mode != CaseSensitivityAuto {
		return mode
	}

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return CaseInsensitive
	}

	return CaseSensitive
}

// pathKey returns a canonical string key for path lookups.
// On case-insensitive filesystems the key is lowercased so that paths differing
// only in case collide in maps and equality checks.
//
// NFC normalization is always applied (even on case-sensitive filesystems) so
// that NFC and NFD representations of the same Unicode path produce identical
// keys. This is critical on macOS, where the filesystem stores filenames as NFD
// (decomposed) but user-configured paths are typically NFC (composed). Without
// normalization, exclude-path matching, debounce deduplication, and gitignore
// prefix checks silently fail on any non-ASCII filename.
// norm.NFC.String is idempotent on ASCII, so pure-ASCII paths are unaffected.
func (w *Watcher) pathKey(path string) string {
	normalized := norm.NFC.String(path)

	if w.effectiveCaseSensitivity == CaseInsensitive {
		return strings.ToLower(normalized)
	}

	return normalized
}
