package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestTestAdjacencyRule_Check_Adjacent(t *testing.T) {
	tests := []struct {
		name          string
		filePatterns  []string
		exemptions    []string
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:         "Go file with test - no violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main.go", ParentPath: "", IsDir: false},
				{Path: "main_test.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "Go file without test - violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 1,
		},
		{
			name:         "TypeScript file with test - no violation",
			filePatterns: []string{"**/*.ts"},
			files: []walker.FileInfo{
				{Path: "utils.ts", ParentPath: "src", IsDir: false},
				{Path: "utils.test.ts", ParentPath: "src", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "exempted file - no violation",
			filePatterns: []string{"**/*.go"},
			exemptions:   []string{"cmd/**/*.go"},
			files: []walker.FileInfo{
				{Path: "cmd/main.go", ParentPath: "cmd", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "test file itself - no violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main_test.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "multiple files in same dir - some missing tests",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "file1.go", ParentPath: "", IsDir: false},
				{Path: "file1_test.go", ParentPath: "", IsDir: false},
				{Path: "file2.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 1, // file2.go missing test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewTestAdjacencyRule("adjacent", "", tt.filePatterns, tt.exemptions)

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

func TestTestAdjacencyRule_getTestFileName(t *testing.T) {
	rule := &TestAdjacencyRule{}

	tests := []struct {
		sourcePath   string
		wantTestName string
	}{
		{"main.go", "main_test.go"},
		{"utils.ts", "utils.test.ts"},
		{"component.tsx", "component.test.tsx"},
		{"helper.js", "helper.spec.js"},
		{"module.py", "test_module.py"},
		{"src/file.go", "file_test.go"},
	}

	for _, tt := range tests {
		t.Run(tt.sourcePath, func(t *testing.T) {
			got := rule.getTestFileName(tt.sourcePath)
			if got != tt.wantTestName {
				t.Errorf("getTestFileName(%q) = %q, want %q", tt.sourcePath, got, tt.wantTestName)
			}
		})
	}
}

func TestTestAdjacencyRule_isTestFile(t *testing.T) {
	rule := &TestAdjacencyRule{}

	tests := []struct {
		path string
		want bool
	}{
		{"main_test.go", true},
		{"utils.test.ts", true},
		{"component.spec.js", true},
		{"test_module.py", false}, // Python test prefix not detected by basename check
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

func TestTestAdjacencyRule_Name(t *testing.T) {
	rule := NewTestAdjacencyRule("adjacent", "", []string{"**/*.go"}, nil)
	if got := rule.Name(); got != "test-adjacency" {
		t.Errorf("Name() = %v, want test-adjacency", got)
	}
}
