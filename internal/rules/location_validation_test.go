package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestTestLocationRule_Check(t *testing.T) {
	// Arrange
	// Act
	// Assert
	tests := []struct {
		name               string
		integrationTestDir string
		allowAdjacent      bool
		exemptions         []string
		files              []walker.FileInfo
		wantViolCount      int
	}{
		{
			name:               "test adjacent to source - allowed",
			integrationTestDir: "tests",
			allowAdjacent:      true,
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
			files: []walker.FileInfo{
				{Path: "tests/integration_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "orphaned test - adjacent not allowed",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			files: []walker.FileInfo{
				{Path: "main_test.go", IsDir: false},
			},
			wantViolCount: 1,
		},
		{
			name:               "test with source nearby - allowed when adjacent enabled",
			integrationTestDir: "tests",
			allowAdjacent:      true,
			files: []walker.FileInfo{
				{Path: "src/utils.go", IsDir: false},
				{Path: "src/utils_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:               "exempted test - no violation",
			integrationTestDir: "tests",
			allowAdjacent:      false,
			exemptions:         []string{"testdata/**"},
			files: []walker.FileInfo{
				{Path: "testdata/fixture_test.go", IsDir: false},
			},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewTestLocationRule(tt.integrationTestDir, tt.allowAdjacent, tt.exemptions)
			violations := rule.Check(tt.files, nil)

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
	rule := NewTestLocationRule("tests", true, nil)
	if got := rule.Name(); got != "test-location" {
		t.Errorf("Name() = %v, want test-location", got)
	}
}
