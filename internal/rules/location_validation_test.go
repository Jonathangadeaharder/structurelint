package rules

import (
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestTestLocationRule_Check(t *testing.T) {
	tests := []struct {
		name               string
		integrationTestDir string
		allowAdjacent      bool
		filePatterns       []string
		exemptions         []string
		files              []walker.FileInfo
		wantViolCount      int
	}{
		{
			name:               "test adjacent to source - allowed",
			integrationTestDir: "tests",
			allowAdjacent:      true,
			filePatterns:       []string{"**/*_test.go"},
			files: []walker.FileInfo{
				{Path: "main.go", IsDir: false},
				{Path: "main_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "test in integration dir - allowed",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			filePatterns:       []string{"**/*_test.go"},
			files: []walker.FileInfo{
				{Path: "tests/integration_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "orphaned test - adjacent not allowed",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			filePatterns:       []string{"**/*_test.go"},
			files: []walker.FileInfo{
				{Path: "main_test.go", IsDir: false},
			},
			wantViolCount: 1,
		},
		{
			name:               "test with source nearby - allowed when adjacent enabled",
			integrationTestDir: "tests",
			allowAdjacent:      true,
			filePatterns:       []string{"**/*_test.go"},
			files: []walker.FileInfo{
				{Path: filepath.Join("src", "utils.go"), IsDir: false},
				{Path: filepath.Join("src", "utils_test.go"), IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "exempted test - no violation",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			filePatterns:       []string{"**/*_test.go"},
			exemptions:         []string{"testdata/**"},
			files: []walker.FileInfo{
				{Path: "testdata/fixture_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "python test ignored by file pattern",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			filePatterns:       []string{"**/*_test.go"}, // Only Go tests
			files: []walker.FileInfo{
				{Path: "tests/test_module.py", IsDir: false}, // Python test
			},
			wantViolCount: 0, // Should be ignored (doesn't match pattern)
		},
		{
			name:               "multiple language patterns",
			integrationTestDir: "tests",
			allowAdjacent:      true,
			filePatterns:       []string{"**/*_test.go", "**/*.test.ts"},
			files: []walker.FileInfo{
				{Path: "main.go", IsDir: false},
				{Path: "main_test.go", IsDir: false},
				{Path: "utils.ts", IsDir: false},
				{Path: "utils.test.ts", IsDir: false},
				{Path: "test_module.py", IsDir: false}, // Python - ignored
			},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewTestLocationRule(tt.integrationTestDir, tt.allowAdjacent, tt.filePatterns, tt.exemptions)

			// Act
			violations := rule.Check(tt.files, nil)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestTestLocationRule_isTestFile(t *testing.T) {
	rule := &TestLocationRule{}

	tests := []struct {
		path string
		want bool
	}{
		{"main_test.go", true},
		{"utils.test.ts", true},
		{"component.spec.js", true},
		{"test_module.py", true},
		{"FileTest.java", true},
		{"file_spec.rb", true},
		{"main.go", false},
		{"utils.ts", false},
		{"module.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := rule.isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestTestLocationRule_getSourceFileName(t *testing.T) {
	rule := &TestLocationRule{}

	tests := []struct {
		testFileName   string
		ext            string
		wantSourceName string
	}{
		{"main_test.go", ".go", "main.go"},
		{"utils.test.ts", ".ts", "utils.ts"},
		{"component.spec.js", ".js", "component.js"},
		{"test_module.py", ".py", "module.py"},
	}

	for _, tt := range tests {
		t.Run(tt.testFileName, func(t *testing.T) {
			got := rule.getSourceFileName(tt.testFileName, tt.ext)
			if got != tt.wantSourceName {
				t.Errorf("getSourceFileName(%q, %q) = %q, want %q",
					tt.testFileName, tt.ext, got, tt.wantSourceName)
			}
		})
	}
}

func TestTestLocationRule_Name(t *testing.T) {
	// Arrange
	rule := NewTestLocationRule("tests", true, []string{"**/*_test.go"}, nil)

	// Act
	name := rule.Name()

	// Assert
	if name != "test-location" {
		t.Errorf("Name() = %v, want test-location", name)
	}
}

func TestTestLocationRule_matchesFilePattern(t *testing.T) {
	tests := []struct {
		name         string
		filePatterns []string
		path         string
		want         bool
	}{
		{
			name:         "go test matches pattern",
			filePatterns: []string{"**/*_test.go"},
			path:         "internal/foo_test.go",
			want:         true,
		},
		{
			name:         "python test does not match go pattern",
			filePatterns: []string{"**/*_test.go"},
			path:         "tests/test_bar.py",
			want:         false,
		},
		{
			name:         "multiple patterns - go matches",
			filePatterns: []string{"**/*_test.go", "**/*.test.ts"},
			path:         "src/utils_test.go",
			want:         true,
		},
		{
			name:         "multiple patterns - typescript matches",
			filePatterns: []string{"**/*_test.go", "**/*.test.ts"},
			path:         "src/utils.test.ts",
			want:         true,
		},
		{
			name:         "multiple patterns - python does not match",
			filePatterns: []string{"**/*_test.go", "**/*.test.ts"},
			path:         "tests/test_module.py",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := &TestLocationRule{FilePatterns: tt.filePatterns}

			// Act
			got := rule.matchesFilePattern(tt.path)

			// Assert
			if got != tt.want {
				t.Errorf("matchesFilePattern(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
