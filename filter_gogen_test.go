package filewatcher

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LarsArtmann/gogenfilter/v3"
)

var (
	generatedCodeFilterOptions = []gogenfilter.FilterOption{
		gogenfilter.FilterSQLC,
		gogenfilter.FilterTempl,
		gogenfilter.FilterProtobuf,
	}

	// Test paths for generated code detection.
	testPathModelsGo     = "/project/db/models.go"
	testPathQuerySQLGo   = "/project/db/query.sql.go"
	testPathUsersSQLGo   = "/project/db/users.sql.go"
	testPathPageTemplGo  = "/project/page_templ.go"
	testPathUserPbGo     = "/project/api/user.pb.go"
	testPathMainGo       = "/project/main.go"
	testPathStatusEnumGo = "/project/status_enum.go"

	// gogenfilter v3.6+ no longer treats the weak "models.go" filename as
	// sqlc-generated on its own: a hand-written models.go is too common. Strong
	// "*.sql.go" patterns still match by filename; content markers catch the rest.
	sqlcEventCases = []testCaseName{
		{testCaseName: "models.go", path: testPathModelsGo, expected: true},
		{testCaseName: "query.sql.go", path: testPathQuerySQLGo, expected: false},
		{testCaseName: "users.sql.go", path: testPathUsersSQLGo, expected: false},
	}

	templEventCases = []testCaseName{
		{testCaseName: "page_templ.go", path: testPathPageTemplGo, expected: false},
	}

	goEnumEventCases = []testCaseName{
		{testCaseName: "status_enum.go", path: testPathStatusEnumGo, expected: false},
	}

	multipleOptionsTestCases = []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "sqlc weak filename with multiple options", path: testPathModelsGo, expected: true},
		{name: "sqlc strong filename with multiple options", path: testPathQuerySQLGo, expected: false},
		{name: "templ with multiple options", path: testPathPageTemplGo, expected: false},
		{name: "protobuf with multiple options", path: testPathUserPbGo, expected: false},
		{name: "regular file with multiple options", path: testPathMainGo, expected: true},
		{
			name:     "go-enum with multiple options (not in list)",
			path:     testPathStatusEnumGo,
			expected: true,
		},
	}
)

var (
	protobufEventCases = twoTestCases(
		"user.pb.go",
		"/project/api/user.pb.go",
		"user_grpc.pb.go",
		"/project/api/user_grpc.pb.go",
	)
	mockgenEventCases = twoTestCases(
		"service_mock.go",
		"/project/mocks/service_mock.go",
		"user_mock.go",
		"/project/mocks/user_mock.go",
	)
	mockeryEventCases = twoTestCases(
		"mock_service.go",
		"/project/mocks/mock_service.go",
		"mock_user.go",
		"/project/mocks/mock_user.go",
	)
)

// twoTestCases creates a []testCaseName with two entries for files that should NOT be filtered.
func twoTestCases(name1, path1, name2, path2 string) []testCaseName {
	return []testCaseName{
		{testCaseName: name1, path: path1, expected: false},
		{testCaseName: name2, path: path2, expected: false},
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_SingleFilters(t *testing.T) {
	runSingleFilterSubtests(t, "SQLC", gogenfilter.FilterSQLC, sqlcEventCases)
	runSingleFilterSubtests(t, "Templ", gogenfilter.FilterTempl, templEventCases)
	runSingleFilterSubtests(t, "GoEnum", gogenfilter.FilterGoEnum, goEnumEventCases)
	runSingleFilterSubtests(t, "Protobuf", gogenfilter.FilterProtobuf, protobufEventCases)
	runSingleFilterSubtests(t, "Mockgen", gogenfilter.FilterMockgen, mockgenEventCases)
	runSingleFilterSubtests(t, "Mockery", gogenfilter.FilterMockery, mockeryEventCases)

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
			Size:      0,
			ModTime:   time.Time{},
		}
		if filter(event) != true {
			t.Error("directories should never be filtered")
		}
	})
}

// runSingleFilterSubtests runs subtests for a single filter type.
func runSingleFilterSubtests(
	t *testing.T,
	name string,
	filterOption gogenfilter.FilterOption,
	cases []testCaseName,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		filter := FilterGeneratedCode(filterOption)
		for _, tc := range cases {
			t.Run(tc.testCaseName, testFilter(filter, tc.path, tc.expected))
		}
	})
}

func testFilter(filter Filter, path string, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		if filter(newTestEvent(path)) != expected {
			t.Errorf("FilterGeneratedCode() = %v, want %v for path %s", !expected, expected, path)
		}
	}
}

// newTestEvent creates a test event for the given path.
func newTestEvent(path string) Event {
	return Event{
		Path:      path,
		Op:        Op(0),
		Timestamp: time.Time{},
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCode_MultipleOptions(t *testing.T) {
	multiFilter := FilterGeneratedCode(generatedCodeFilterOptions...)

	for _, testCase := range multipleOptionsTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			if multiFilter(newTestEvent(testCase.path)) != testCase.expected {
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

	runFilterSubtests(t, []testCaseName{
		{testCaseName: testPathModelsGo, path: testPathModelsGo, expected: true},
		{testCaseName: testPathQuerySQLGo, path: testPathQuerySQLGo, expected: false},
		{testCaseName: testPathPageTemplGo, path: testPathPageTemplGo, expected: false},
		{testCaseName: testPathStatusEnumGo, path: testPathStatusEnumGo, expected: false},
		{testCaseName: testPathUserPbGo, path: testPathUserPbGo, expected: false},
		{
			testCaseName: "/project/mocks/service_mock.go",
			path:         "/project/mocks/service_mock.go",
			expected:     false,
		},
		{testCaseName: testPathMainGo, path: testPathMainGo, expected: true},
		{testCaseName: "/project/utils.go", path: "/project/utils.go", expected: true},
	}, filter)
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCodeFull_WithContent(t *testing.T) {
	// Create a temporary file with content for content-based detection
	tmpDir := t.TempDir()

	// Create a file with the weak sqlc filename pattern but WITHOUT the sqlc
	// content marker: gogenfilter v3.6+ keeps it, because the weak filename
	// alone is not proof of sqlc generation.
	sqlcFilenameRegularContent := tmpDir + "/models.go"

	err := writeFile(sqlcFilenameRegularContent, []byte("package main\n\nfunc main() {}"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a file with a strong sqlc filename pattern and proper content marker
	sqlcFile := tmpDir + "/query.sql.go"

	sqlcContent := "// Code generated by sqlc. DO NOT EDIT.\n\npackage db"

	err = writeFile(sqlcFile, []byte(sqlcContent))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a file with a WEAK sqlc filename but WITH the content marker:
	// content-based detection must catch what the filename alone cannot.
	if err := os.MkdirAll(tmpDir+"/generated", 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	sqlcWeakNameWithMarker := tmpDir + "/generated/models.go"

	err = writeFile(sqlcWeakNameWithMarker, []byte(sqlcContent))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with content checking enabled
	filter := FilterGeneratedCodeFull(ContentCheckEnabled, gogenfilter.FilterSQLC)

	// Weak sqlc filename without marker is kept: filename alone is not proof
	// (gogenfilter v3.6+), and the content check finds no marker either.
	// Filter returns false to discard, so we expect true (keep) here
	sqlcFilenameEvent := Event{
		Path:      sqlcFilenameRegularContent,
		Op:        Op(0),
		Timestamp: time.Time{},
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}
	if !filter(sqlcFilenameEvent) {
		t.Errorf("Expected weak sqlc filename without marker to be kept (return true)")
	}

	// Weak sqlc filename WITH the content marker must be filtered by the
	// content check — this is exactly what content detection is for
	sqlcWeakNameEvent := Event{
		Path:      sqlcWeakNameWithMarker,
		Op:        Op(0),
		Timestamp: time.Time{},
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}
	if filter(sqlcWeakNameEvent) {
		t.Errorf("Expected weak sqlc filename with marker to be filtered (return false)")
	}

	// File with sqlc content marker should also be filtered
	// Filter returns false to discard
	sqlcEvent := Event{
		Path:      sqlcFile,
		Op:        Op(0),
		Timestamp: time.Time{},
		IsDir:     false,
		Size:      0,
		ModTime:   time.Time{},
	}
	if filter(sqlcEvent) {
		t.Errorf("Expected sqlc file to be filtered with content check (return false)")
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestFilterGeneratedCodeWithFilter(t *testing.T) {
	// Create a gogenfilter.Filter instance
	config, configErr := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
	if configErr != nil {
		t.Fatalf("Failed to create filter config: %v", configErr)
	}

	genFilter, filterErr := gogenfilter.NewFilter(config)
	if filterErr != nil {
		t.Fatalf("Failed to create filter: %v", filterErr)
	}

	filter := FilterGeneratedCodeWithFilter(genFilter)

	runFilterSubtests(t, []testCaseName{
		{testCaseName: testPathModelsGo, path: testPathModelsGo, expected: true},
		{testCaseName: testPathQuerySQLGo, path: testPathQuerySQLGo, expected: false},
		{testCaseName: testPathPageTemplGo, path: testPathPageTemplGo, expected: false},
		{testCaseName: testPathMainGo, path: testPathMainGo, expected: true},
		{testCaseName: "/project/vendor/lib.go", path: "/project/vendor/lib.go", expected: true},
	}, filter)
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestGeneratedCodeDetector(t *testing.T) {
	detector := NewGeneratedCodeDetector(gogenfilter.FilterSQLC, gogenfilter.FilterProtobuf)

	tests := []struct {
		path     string
		expected bool
	}{
		{testPathQuerySQLGo, true},   // sqlc (strong filename)
		{testPathModelsGo, false},    // weak sqlc filename alone is not proof (v3.6+)
		{testPathUserPbGo, true},     // protobuf
		{testPathMainGo, false},      // regular
		{testPathPageTemplGo, false}, // templ (not in detector options)
	}

	for _, tc := range tests { //nolint:varnamelen // idiomatic table-driven test variable
		t.Run(tc.path, func(t *testing.T) {
			result := detector.IsGenerated(tc.path)
			if result != tc.expected {
				t.Errorf(
					"IsGenerated() = %v, want %v for path %s",
					result,
					tc.expected,
					tc.path,
				)
			}
		})
	}
}

//nolint:paralleltest // Test files cannot be parallel due to file system operations
func TestGeneratedCodeDetector_GetReason(t *testing.T) {
	detector := NewGeneratedCodeDetector(gogenfilter.FilterSQLC, gogenfilter.FilterTempl)

	tests := []struct {
		path     string
		expected gogenfilter.FilterReason
	}{
		{testPathQuerySQLGo, gogenfilter.ReasonSQLC},
		{testPathModelsGo, gogenfilter.ReasonNotFiltered}, // weak filename alone (v3.6+)
		{testPathPageTemplGo, gogenfilter.ReasonTempl},
		{testPathMainGo, gogenfilter.ReasonNotFiltered},
	}

	for _, tc := range tests { //nolint:varnamelen // idiomatic table-driven test variable
		t.Run(tc.path, func(t *testing.T) {
			reason := detector.GetReason(tc.path)
			if reason != tc.expected {
				t.Errorf(
					"GetReason() = %v, want %v for path %s",
					reason,
					tc.expected,
					tc.path,
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

// testCaseName is a test case with a name field for subtest naming.
type testCaseName struct {
	testCaseName string
	path         string
	expected     bool
}

// runFilterSubtests runs subtests for a filter function with the given test cases.
func runFilterSubtests(t *testing.T, tests []testCaseName, filter func(Event) bool) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.path, func(t *testing.T) {
			result := filter(newTestEvent(testCase.path))

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
