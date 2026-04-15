//nolint:testpackage // Tests internal functions, must be in same package
package filewatcher

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/gogenfilter"
)

//nolint:gochecknoglobals // Test data must be package-level for funlen compliance
var (
	generatedCodeFilterOptions = []gogenfilter.FilterOption{
		gogenfilter.FilterSQLC,
		gogenfilter.FilterTempl,
		gogenfilter.FilterProtobuf,
	}

	sqlcEventCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{"models.go", "/project/db/models.go", false},
		{"query.sql.go", "/project/db/query.sql.go", false},
		{"users.sql.go", "/project/db/users.sql.go", false},
	}

	templEventCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{"page_templ.go", "/project/components/page_templ.go", false},
	}

	goEnumEventCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{"status_enum.go", "/project/types/status_enum.go", false},
	}

	protobufEventCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{"user.pb.go", "/project/api/user.pb.go", false},
		{"user_grpc.pb.go", "/project/api/user_grpc.pb.go", false},
	}

	mockgenEventCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{"service_mock.go", "/project/mocks/service_mock.go", false},
		{"mock_service.go", "/project/mocks/mock_service.go", false},
	}

	multipleOptionsTestCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "sqlc with multiple options", path: "/project/db/models.go", expected: false},
		{name: "templ with multiple options", path: "/project/page_templ.go", expected: false},
		{name: "protobuf with multiple options", path: "/project/api/user.pb.go", expected: false},
		{name: "regular file with multiple options", path: "/project/main.go", expected: true},
		{name: "go-enum with multiple options (not in list)", path: "/project/status_enum.go", expected: true},
	}
)

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_SingleFilters(t *testing.T) {
	t.Run("SQLC", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterSQLC)
		for _, tc := range sqlcEventCases {
			t.Run(tc.name, testFilter(filter, tc.path, tc.expected))
		}
	})

	t.Run("Templ", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterTempl)
		for _, tc := range templEventCases {
			t.Run(tc.name, testFilter(filter, tc.path, tc.expected))
		}
	})

	t.Run("GoEnum", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterGoEnum)
		for _, tc := range goEnumEventCases {
			t.Run(tc.name, testFilter(filter, tc.path, tc.expected))
		}
	})

	t.Run("Protobuf", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterProtobuf)
		for _, tc := range protobufEventCases {
			t.Run(tc.name, testFilter(filter, tc.path, tc.expected))
		}
	})

	t.Run("Mockgen", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterMockgen)
		for _, tc := range mockgenEventCases {
			t.Run(tc.name, testFilter(filter, tc.path, tc.expected))
		}
	})

	t.Run("RegularFile", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterSQLC)
		testFilter(filter, "/project/main.go", true)(t)
	})

	t.Run("DirectoriesNotFiltered", func(t *testing.T) {
		filter := FilterGeneratedCode(gogenfilter.FilterSQLC)

		event := Event{
			Path:      "/project/db/models.go",
			Op:        Op(0),
			Timestamp: time.Time{},
			IsDir:     true,
		}
		if filter(event) != true {
			t.Error("directories should never be filtered")
		}
	})
}

func testFilter(filter Filter, path string, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		event := Event{Path: path, Op: Op(0), Timestamp: time.Time{}, IsDir: false}
		if filter(event) != expected {
			t.Errorf("FilterGeneratedCode() = %v, want %v for path %s", !expected, expected, path)
		}
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_MultipleOptions(t *testing.T) {
	multiFilter := FilterGeneratedCode(generatedCodeFilterOptions...)

	for _, testCase := range multipleOptionsTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			if multiFilter(
				Event{Path: testCase.path, Op: Op(0), Timestamp: time.Time{}, IsDir: false},
			) != testCase.expected {
				t.Errorf(
					"FilterGeneratedCode() = %v, want %v for path %s",
					!testCase.expected,
					testCase.expected,
					testCase.path,
				)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_DefaultAll(t *testing.T) {
	// When no options are provided, all generators should be checked
	filter := FilterGeneratedCode()

	runFilterSubtests(t, []filterTestCase{
		{"/project/db/models.go", false},          // sqlc
		{"/project/page_templ.go", false},         // templ
		{"/project/status_enum.go", false},        // go-enum
		{"/project/api/user.pb.go", false},        // protobuf
		{"/project/mocks/service_mock.go", false}, // mockgen
		{"/project/main.go", true},                // regular
		{"/project/utils.go", true},               // regular
	}, filter)
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
	filter := FilterGeneratedCodeFull(ContentCheckEnabled, gogenfilter.FilterSQLC)

	// File with sqlc filename pattern is filtered even without sqlc content
	// because filename-based detection happens first
	// Filter returns false to discard, so we expect false here
	sqlcFilenameEvent := Event{
		Path:      sqlcFilenameRegularContent,
		Op:        Op(0),
		Timestamp: time.Time{},
		IsDir:     false,
	}
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

	runFilterSubtests(t, []filterTestCase{
		{"/project/db/models.go", false},  // filtered
		{"/project/page_templ.go", false}, // filtered
		{"/project/main.go", true},        // not filtered
		{"/project/vendor/lib.go", true},  // not filtered (no vendor pattern)
	}, filter)

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
				t.Errorf(
					"IsGenerated() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.path,
				)
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
				t.Errorf(
					"GetReason() = %v, want %v for path %s",
					reason,
					testCase.expected,
					testCase.path,
				)
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

// filterTestCase represents a test case for filter testing.
type filterTestCase struct {
	path     string
	expected bool
}

// runFilterSubtests runs subtests for a filter function with the given test cases.
func runFilterSubtests(t *testing.T, tests []filterTestCase, filter func(Event) bool) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			event := Event{Path: testCase.path, Op: Op(0), Timestamp: time.Time{}, IsDir: false}

			result := filter(event)

			if result != testCase.expected {
				t.Errorf(
					"filter() = %v, want %v for path %s",
					result,
					testCase.expected,
					testCase.path,
				)
			}
		})
	}
}
