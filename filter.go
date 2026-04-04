package filewatcher

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Filter determines whether a file event should be processed.
// Return true to keep the event, false to discard it.
type Filter func(event Event) bool

// FilterExtensions creates a filter that only passes events for files
// matching one of the given extensions. Extensions should include the
// dot prefix (e.g., ".go", ".md").
func FilterExtensions(exts ...string) Filter {
	extSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		extSet[strings.ToLower(ext)] = struct{}{}
	}
	return func(event Event) bool {
		ext := strings.ToLower(filepath.Ext(event.Path))
		_, ok := extSet[ext]
		return ok
	}
}

// FilterIgnoreExtensions creates a filter that discards events for files
// matching one of the given extensions.
func FilterIgnoreExtensions(exts ...string) Filter {
	extSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		extSet[strings.ToLower(ext)] = struct{}{}
	}
	return func(event Event) bool {
		ext := strings.ToLower(filepath.Ext(event.Path))
		_, ignore := extSet[ext]
		return !ignore
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
	opSet := make(map[Op]struct{}, len(ops))
	for _, op := range ops {
		opSet[op] = struct{}{}
	}
	return func(event Event) bool {
		_, ok := opSet[event.Op]
		return ok
	}
}

// FilterNotOperations creates a filter that discards events matching
// any of the given operations.
func FilterNotOperations(ops ...Op) Filter {
	opSet := make(map[Op]struct{}, len(ops))
	for _, op := range ops {
		opSet[op] = struct{}{}
	}
	return func(event Event) bool {
		_, exclude := opSet[event.Op]
		return !exclude
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
func FilterRegex(pattern string) Filter {
	re := regexp.MustCompile(pattern)
	return func(event Event) bool {
		return re.MatchString(event.Path)
	}
}

// FilterMinSize creates a filter that only passes events for files
// with size greater than or equal to the given minimum size in bytes.
// Directory events are not filtered by size.
func FilterMinSize(minSize int64) Filter {
	return func(event Event) bool {
		if event.IsDir {
			return true // Don't filter directories by size
		}
		info, err := os.Stat(event.Path)
		if err != nil {
			return false // If we can't stat, filter out
		}
		return info.Size() >= minSize
	}
}

// FilterCustom is an alias for creating custom filters.
// It provides a named escape hatch for complex filtering logic.
func FilterCustom(fn func(Event) bool) Filter {
	return fn
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
