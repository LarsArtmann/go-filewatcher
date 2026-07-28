package filewatcher

import (
	"runtime"
	"strings"
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
		return "unknown"
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
func (w *Watcher) pathKey(path string) string {
	if w.effectiveCaseSensitivity == CaseInsensitive {
		return strings.ToLower(path)
	}

	return path
}
