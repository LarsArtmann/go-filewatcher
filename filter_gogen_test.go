package filewatcher

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/gogenfilter"
)

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_SingleFilters(t *testing.T) {
	tests := []struct {
		name     string
		options  []gogenfilter.FilterOption
		event    Event
		expected bool
	}{
		{
			name:     "sqlc models.go - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterSQLC},
			event:    Event{Path: "/project/db/models.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "sqlc query.sql.go - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterSQLC},
			event:    Event{Path: "/project/db/query.sql.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "sqlc custom.sql.go - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterSQLC},
			event:    Event{Path: "/project/db/users.sql.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "regular .go file - not filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterSQLC},
			event:    Event{Path: "/project/main.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: true,
		},
		{
			name:     "templ generated file - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterTempl},
			event:    Event{Path: "/project/components/page_templ.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "go-enum generated file - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterGoEnum},
			event:    Event{Path: "/project/types/status_enum.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "protobuf generated file - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterProtobuf},
			event:    Event{Path: "/project/api/user.pb.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "protobuf grpc file - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterProtobuf},
			event:    Event{Path: "/project/api/user_grpc.pb.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "mockgen file with suffix - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterMockgen},
			event:    Event{Path: "/project/mocks/service_mock.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "mockgen file with prefix - filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterMockgen},
			event:    Event{Path: "/project/mocks/mock_service.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name:     "directories are never filtered",
			options:  []gogenfilter.FilterOption{gogenfilter.FilterSQLC},
			event:    Event{Path: "/project/db/models.go", Op: Op(0), Timestamp: time.Time{}, IsDir: true},
			expected: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filter := FilterGeneratedCode(testCase.options...)

			result := filter(testCase.event)
			if result != testCase.expected {
				t.Errorf(
					"FilterGeneratedCode() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.event.Path,
				)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_MultipleOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []gogenfilter.FilterOption
		event    Event
		expected bool
	}{
		{
			name: "sqlc with multiple options - filtered",
			options: []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterProtobuf,
			},
			event:    Event{Path: "/project/db/models.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name: "templ with multiple options - filtered",
			options: []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterProtobuf,
			},
			event:    Event{Path: "/project/page_templ.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name: "protobuf with multiple options - filtered",
			options: []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterProtobuf,
			},
			event:    Event{Path: "/project/api/user.pb.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: false,
		},
		{
			name: "regular file with multiple options - not filtered",
			options: []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterProtobuf,
			},
			event:    Event{Path: "/project/main.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: true,
		},
		{
			name: "go-enum with multiple options - not filtered (not in list)",
			options: []gogenfilter.FilterOption{
				gogenfilter.FilterSQLC,
				gogenfilter.FilterTempl,
				gogenfilter.FilterProtobuf,
			},
			event:    Event{Path: "/project/status_enum.go", Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			expected: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filter := FilterGeneratedCode(testCase.options...)

			result := filter(testCase.event)
			if result != testCase.expected {
				t.Errorf(
					"FilterGeneratedCode() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.event.Path,
				)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_DefaultAll(t *testing.T) {
	// When no options are provided, all generators should be checked
	filter := FilterGeneratedCode()

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/db/models.go", false},          // sqlc
		{"/project/page_templ.go", false},         // templ
		{"/project/status_enum.go", false},        // go-enum
		{"/project/api/user.pb.go", false},        // protobuf
		{"/project/mocks/service_mock.go", false}, // mockgen
		{"/project/main.go", true},                // regular
		{"/project/utils.go", true},               // regular
	}

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			event := Event{Path: testCase.path, Op: Op(0), Timestamp: time.Time{}, IsDir: false}

			result := filter(event)
			if result != testCase.expected {
				t.Errorf(
					"FilterGeneratedCode() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.path,
				)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCodeFull_WithContent(t *testing.T) {
	// Create a temporary file with content for content-based detection
	tmpDir := t.TempDir()

	// Create a file with sqlc filename pattern but WITHOUT sqlc content
	// When content check is enabled, this file will be filtered by filename
	// (since models.go is a sqlc pattern), regardless of content
	sqlcFilenameRegularContent := tmpDir + "/models.go"

	err := writeFile(sqlcFilenameRegularContent, []byte("package main\n\nfunc main() {}"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a file that IS sqlc generated with proper content marker
	sqlcFile := tmpDir + "/query.sql.go"

	sqlcContent := "// Code generated by sqlc. DO NOT EDIT.\n\npackage db"

	err = writeFile(sqlcFile, []byte(sqlcContent))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with content checking enabled
	filter := FilterGeneratedCodeFull(true, gogenfilter.FilterSQLC)

	// File with sqlc filename pattern is filtered even without sqlc content
	// because filename-based detection happens first
	// Filter returns false to discard, so we expect false here
	sqlcFilenameEvent := Event{Path: sqlcFilenameRegularContent, Op: Op(0), Timestamp: time.Time{}, IsDir: false}
	if filter(sqlcFilenameEvent) {
		t.Errorf("Expected sqlc filename pattern file to be filtered (return false)")
	}

	// File with sqlc content marker should also be filtered
	// Filter returns false to discard
	sqlcEvent := Event{Path: sqlcFile, Op: Op(0), Timestamp: time.Time{}, IsDir: false}
	if filter(sqlcEvent) {
		t.Errorf("Expected sqlc file to be filtered with content check (return false)")
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCodeWithFilter(t *testing.T) {
	// Create a gogenfilter.Filter instance
	genFilter := gogenfilter.NewFilter(true, []gogenfilter.FilterOption{gogenfilter.FilterAll})

	filter := FilterGeneratedCodeWithFilter(genFilter)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/db/models.go", false},  // filtered
		{"/project/page_templ.go", false}, // filtered
		{"/project/main.go", true},        // not filtered
		{"/project/vendor/lib.go", true},  // not filtered (no vendor pattern)
	}

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			event := Event{Path: testCase.path, Op: Op(0), Timestamp: time.Time{}, IsDir: false}

			result := filter(event)
			if result != testCase.expected {
				t.Errorf(
					"FilterGeneratedCodeWithFilter() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.path,
				)
			}
		})
	}

	// Check metrics were recorded
	stats := genFilter.GetStats()
	if stats.TotalFilesChecked == 0 {
		t.Error("Expected metrics to be recorded")
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestGeneratedCodeDetector(t *testing.T) {
	detector := NewGeneratedCodeDetector(gogenfilter.FilterSQLC, gogenfilter.FilterProtobuf)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/project/db/models.go", true},   // sqlc
		{"/project/api/user.pb.go", true}, // protobuf
		{"/project/main.go", false},       // regular
		{"/project/page_templ.go", false}, // templ (not in detector options)
	}

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			result := detector.IsGenerated(testCase.path)
			if result != testCase.expected {
				t.Errorf("IsGenerated() = %v, want %v for path %s", result, testCase.expected, testCase.path)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestGeneratedCodeDetector_GetReason(t *testing.T) {
	detector := NewGeneratedCodeDetector(gogenfilter.FilterSQLC, gogenfilter.FilterTempl)

	tests := []struct {
		path         string
		expected     gogenfilter.FilterReason
		shouldFilter bool
	}{
		{"/project/db/models.go", gogenfilter.ReasonSQLC, true},
		{"/project/page_templ.go", gogenfilter.ReasonTempl, true},
		{"/project/main.go", gogenfilter.ReasonNotFiltered, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			reason := detector.GetReason(testCase.path)
			if reason != testCase.expected {
				t.Errorf("GetReason() = %v, want %v for path %s", reason, testCase.expected, testCase.path)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestGeneratedCodeDetector_IsGeneratedWithContent(t *testing.T) {
	detector := NewGeneratedCodeDetector(gogenfilter.FilterGeneric)

	// Content with "// Code generated by" marker
	generatedContent := "// Code generated by mockgen. DO NOT EDIT.\n\npackage mocks"
	regularContent := "package main\n\nfunc main() {}"

	if !detector.IsGeneratedWithContent("/any/path.go", generatedContent) {
		t.Error("Expected generated content to be detected")
	}

	if detector.IsGeneratedWithContent("/any/path.go", regularContent) {
		t.Error("Expected regular content to not be detected as generated")
	}
}

// writeFile is a helper to write test files.
func writeFile(path string, content []byte) error {
	err := os.WriteFile(path, content, 0o600)
	if err != nil {
		return fmt.Errorf("writeFile %q: %w", path, err)
	}

	return nil
}
