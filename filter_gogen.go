package filewatcher

import (
	"os"

	"github.com/LarsArtmann/gogenfilter"
)

// FilterGeneratedCode creates a filewatcher filter that excludes auto-generated
// Go code files detected by gogenfilter. It supports two-phase detection:
// filename-based (zero I/O) and content-based (reads file).
//
// Use FilterGeneratedCodeOptions for zero-I/O detection (filename patterns only),
// or FilterGeneratedCodeFull for detection that also reads file content.
//
// Example:
//
//	watcher, _ := filewatcher.New("./src",
//	    filewatcher.WithFilter(filewatcher.FilterGeneratedCode(
//	        gogenfilter.FilterSQLC,
//	        gogenfilter.FilterProtobuf,
//	    )),
//	)
//
// This filter returns false for generated files (excluding them from events),
// and true for non-generated files (allowing them through).
func FilterGeneratedCode(options ...gogenfilter.FilterOption) Filter {
	return FilterGeneratedCodeFull(false, options...)
}

// FilterGeneratedCodeFull creates a filter with configurable content checking.
//
// If checkContent is false: only filename-based detection (zero I/O).
// If checkContent is true: filename + content detection (may read files).
//
// Content checking is more accurate but requires file I/O. For file watching
// scenarios, filename-only detection is usually sufficient since generated
// files typically have distinctive naming patterns.
//
// The filter returns true to keep (non-generated) files, false to discard
// (generated) files.
func FilterGeneratedCodeFull(checkContent bool, options ...gogenfilter.FilterOption) Filter {
	// Default to all generators if none specified
	opts := options
	if len(opts) == 0 {
		opts = []gogenfilter.FilterOption{gogenfilter.FilterAll}
	}

	// Build the options map for detection
	optMap := make(map[gogenfilter.FilterOption]bool)

	for _, opt := range opts {
		if opt == gogenfilter.FilterAll {
			// Enable all specific options
			for _, specific := range []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterGoEnum,
				gogenfilter.FilterProtobuf,
				gogenfilter.FilterMockgen,
				gogenfilter.FilterStringer,
			} {
				optMap[specific] = true
			}

			optMap[gogenfilter.FilterGeneric] = true
		} else {
			optMap[opt] = true
		}
	}

	return func(event Event) bool {
		// Directories are not filtered by generated code detection
		if event.IsDir {
			return true
		}

		// Phase 1: Filename-based detection (zero I/O)
		reason := gogenfilter.DetectReason(event.Path, "", optMap)
		if reason != gogenfilter.ReasonNotFiltered {
			return false // Filter out generated file
		}

		// Phase 2: Content-based detection (if requested)
		if checkContent {
			content, err := os.ReadFile(event.Path)
			if err == nil {
				reason = gogenfilter.DetectReason(event.Path, string(content), optMap)
				if reason != gogenfilter.ReasonNotFiltered {
					return false // Filter out generated file
				}
			}
		}

		return true // Keep non-generated files
	}
}

// FilterGeneratedCodeWithFilter creates a filter using an existing
// gogenfilter.Filter instance. This allows for more advanced configuration
// including custom filesystems and include/exclude patterns.
//
// The gogenfilter.Filter should be created with metrics enabled for tracking.
//
// Example:
//
//	genFilter := gogenfilter.NewFilter(true, []gogenfilter.FilterOption{gogenfilter.FilterAll})
//	genFilter.WithExcludePatterns([]string{"*_custom.go"})
//
//	watcher, _ := filewatcher.New("./src",
//	    filewatcher.WithFilter(filewatcher.FilterGeneratedCodeWithFilter(genFilter)),
//	)
func FilterGeneratedCodeWithFilter(genFilter *gogenfilter.Filter) Filter {
	return func(event Event) bool {
		// Directories are not filtered by generated code detection
		if event.IsDir {
			return true
		}

		// Use gogenfilter's ShouldFilter method
		// Returns true if file should be filtered (excluded)
		return !genFilter.ShouldFilter(event.Path)
	}
}

// GeneratedCodeDetector provides a reusable detector for generated code.
// Useful when you need to check files outside of the event filter context.
type GeneratedCodeDetector struct {
	options map[gogenfilter.FilterOption]bool
}

// NewGeneratedCodeDetector creates a new detector with the specified options.
func NewGeneratedCodeDetector(options ...gogenfilter.FilterOption) *GeneratedCodeDetector {
	opts := options
	if len(opts) == 0 {
		opts = []gogenfilter.FilterOption{gogenfilter.FilterAll}
	}

	optMap := make(map[gogenfilter.FilterOption]bool)

	for _, opt := range opts {
		if opt == gogenfilter.FilterAll {
			for _, specific := range []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterGoEnum,
				gogenfilter.FilterProtobuf,
				gogenfilter.FilterMockgen,
				gogenfilter.FilterStringer,
			} {
				optMap[specific] = true
			}

			optMap[gogenfilter.FilterGeneric] = true
		} else {
			optMap[opt] = true
		}
	}

	return &GeneratedCodeDetector{options: optMap}
}

// IsGenerated checks if a file path represents generated code using
// filename-based detection only (zero I/O).
func (d *GeneratedCodeDetector) IsGenerated(filePath string) bool {
	reason := gogenfilter.DetectReason(filePath, "", d.options)

	return reason != gogenfilter.ReasonNotFiltered
}

// IsGeneratedWithContent checks if a file is generated using both
// filename and content detection.
func (d *GeneratedCodeDetector) IsGeneratedWithContent(filePath string, content string) bool {
	reason := gogenfilter.DetectReason(filePath, content, d.options)

	return reason != gogenfilter.ReasonNotFiltered
}

// GetReason returns the specific reason why a file was detected as generated,
// or gogenfilter.ReasonNotFiltered if it's not generated.
func (d *GeneratedCodeDetector) GetReason(filePath string) gogenfilter.FilterReason {
	return gogenfilter.DetectReason(filePath, "", d.options)
}
