package filewatcher

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Filter determines whether a file event should be processed.
// Return true to keep the event, false to discard it.
type Filter func(event Event) bool

// FilterExtensions creates a filter that only passes events for files
// matching one of the given extensions. Extensions should include the
// dot prefix (e.g., ".go", ".md").
func FilterExtensions(exts ...string) Filter {
	return makeExtFilter(exts, true)
}

// FilterIgnoreExtensions creates a filter that discards events for files
// matching one of the given extensions.
func FilterIgnoreExtensions(exts ...string) Filter {
	return makeExtFilter(exts, false)
}

func makeExtFilter(exts []string, include bool) Filter {
	extSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		extSet[strings.ToLower(ext)] = struct{}{}
	}

	return func(event Event) bool {
		ext := strings.ToLower(filepath.Ext(event.Path))
		_, found := extSet[ext]

		return found == include
	}
}

// FilterIgnoreDirs creates a filter that discards events for files
// within directories matching any of the given directory names.
// Directory names are matched against path components (e.g., "vendor"
// matches both "vendor" and "pkg/vendor").
func FilterIgnoreDirs(dirs ...string) Filter {
	dirSet := make(map[string]struct{}, len(dirs))
	for _, dir := range dirs {
		dirSet[dir] = struct{}{}
	}

	return func(event Event) bool {
		for part := range dirSet {
			sep := string(filepath.Separator)
			if strings.Contains(event.Path, sep+part+sep) ||
				strings.HasSuffix(event.Path, sep+part) ||
				filepath.Base(event.Path) == part {
				return false
			}
		}

		return true
	}
}

// FilterExcludePaths creates a filter that discards events for files
// matching any of the given exact paths. Paths are matched after
// normalization (absolute path conversion).
//
// This differs from FilterIgnoreDirs which matches directory names anywhere
// in the path. FilterExcludePaths requires exact path matches.
//
// Example:
//
//	// Exclude specific files
//	watcher, _ := filewatcher.New("./src",
//	    filewatcher.WithFilter(filewatcher.FilterExcludePaths(
//	        "/home/user/project/src/generated.go",
//	        "/home/user/project/src/vendor",
//	    )),
//	)
func FilterExcludePaths(paths ...string) Filter {
	pathSet := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		// Normalize to absolute path for consistent matching
		abs, err := filepath.Abs(path)
		if err == nil {
			pathSet[abs] = struct{}{}
		} else {
			// Fall back to original path if Abs fails
			pathSet[path] = struct{}{}
		}
	}

	return func(event Event) bool {
		_, excluded := pathSet[event.Path]

		return !excluded
	}
}

// FilterIgnoreHidden creates a filter that discards events for hidden
// files and directories (those starting with a dot).
func FilterIgnoreHidden() Filter {
	return func(event Event) bool {
		base := filepath.Base(event.Path)
		if strings.HasPrefix(base, ".") {
			return false
		}

		for part := range strings.SplitSeq(event.Path, string(filepath.Separator)) {
			if strings.HasPrefix(part, ".") && part != "." && part != ".." {
				return false
			}
		}

		return true
	}
}

// FilterOperations creates a filter that only passes events matching
// one of the given operations.
func FilterOperations(ops ...Op) Filter {
	return makeOpFilter(ops, true)
}

// FilterNotOperations creates a filter that discards events matching
// any of the given operations.
func FilterNotOperations(ops ...Op) Filter {
	return makeOpFilter(ops, false)
}

func makeOpFilter(ops []Op, include bool) Filter {
	opSet := make(map[Op]struct{}, len(ops))
	for _, op := range ops {
		opSet[op] = struct{}{}
	}

	return func(event Event) bool {
		_, found := opSet[event.Op]

		return found == include
	}
}

// FilterGlob creates a filter that only passes events for files
// matching the given glob pattern.
func FilterGlob(pattern string) Filter {
	return func(event Event) bool {
		matched, err := filepath.Match(pattern, filepath.Base(event.Path))
		if err != nil {
			return false
		}

		return matched
	}
}

// FilterRegex creates a filter that only passes events for paths
// matching the given regular expression pattern. The pattern is
// pre-compiled at creation time for efficiency.
// Panics if the pattern is invalid (use regexp.Compile for runtime validation).
func FilterRegex(pattern string) Filter {
	re := regexp.MustCompile(pattern)

	return func(event Event) bool {
		return re.MatchString(event.Path)
	}
}

// filterFileStat extracts common file stat logic used by size/time filters.
// Returns (info, true, true) if stat succeeded and event is a file.
// Returns (nil, true, false) if event is a directory.
// Returns (nil, false, false) if stat fails.
func filterFileStat(event Event) (os.FileInfo, bool, bool) {
	if event.IsDir {
		return nil, true, false // isFile=false, shouldFilter=false
	}

	info, err := os.Stat(event.Path)
	if err != nil {
		return nil, false, false // stat failed, shouldFilter=false
	}

	return info, true, true // stat succeeded, isFile=true, shouldFilter=true
}

// makeSizeFilter creates a filter that applies a size comparison.
// Use >= for min size, <= for max size.
func makeSizeFilter(threshold int64, isMin bool) Filter {
	return func(event Event) bool {
		info, isFile, shouldFilter := filterFileStat(event)
		if !shouldFilter {
			return isFile // directories pass through (true), stat fails filter out (false)
		}

		if isMin {
			return info.Size() >= threshold
		}

		return info.Size() <= threshold
	}
}

// FilterMinSize creates a filter that only passes events for files
// with size greater than or equal to the given minimum size in bytes.
// Directory events are not filtered by size.
func FilterMinSize(minSize int64) Filter {
	return makeSizeFilter(minSize, true)
}

// FilterMaxSize creates a filter that only passes events for files
// with size less than or equal to the given maximum size in bytes.
// Directory events are not filtered by size.
func FilterMaxSize(maxSize int64) Filter {
	return makeSizeFilter(maxSize, false)
}

// FilterModifiedSince creates a filter that only passes events for files
// modified after the given time. Directory events are not filtered by time.
// Useful for ignoring old files during initial scan.
func FilterModifiedSince(minTime time.Time) Filter {
	return func(event Event) bool {
		info, isFile, shouldFilter := filterFileStat(event)
		if !shouldFilter {
			return isFile // directories pass through (true), stat fails filter out (false)
		}

		return info.ModTime().After(minTime)
	}
}

// FilterMinAge creates a filter that only passes events for files
// that are at least the given age old. Directory events are not filtered.
// Useful for ignoring recently modified files (e.g., during save operations).
func FilterMinAge(age time.Duration) Filter {
	return func(event Event) bool {
		info, isFile, shouldFilter := filterFileStat(event)
		if !shouldFilter {
			return isFile // directories pass through (true), stat fails filter out (false)
		}

		return time.Since(info.ModTime()) >= age
	}
}

// FilterAnd combines multiple filters with AND logic.
// All filters must return true for the event to pass.
func FilterAnd(filters ...Filter) Filter {
	return func(event Event) bool {
		for _, f := range filters {
			if !f(event) {
				return false
			}
		}

		return true
	}
}

// FilterOr combines multiple filters with OR logic.
// At least one filter must return true for the event to pass.
func FilterOr(filters ...Filter) Filter {
	return func(event Event) bool {
		for _, f := range filters {
			if f(event) {
				return true
			}
		}

		return false
	}
}

// FilterNot inverts a filter.
func FilterNot(f Filter) Filter {
	return func(event Event) bool {
		return !f(event)
	}
}
