package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxFilesRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		maxFiles      int
		files         []walker.FileInfo
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:     "GivenFileCountWithinLimit_ThenReturnsNoViolations",
			maxFiles: 10,
			files: []walker.FileInfo{
				{Path: "src/file1.go", ParentPath: "src", IsDir: false},
				{Path: "src/file2.go", ParentPath: "src", IsDir: false},
				{Path: "src/file3.go", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"src": {},
			},
			wantViolCount: 0,
		},
		{
			name:     "GivenTestFiles_ThenExcludedFromCount",
			maxFiles: 5,
			files: []walker.FileInfo{
				{Path: "src/file1.go", ParentPath: "src", IsDir: false},
				{Path: "src/file2.go", ParentPath: "src", IsDir: false},
				{Path: "src/file3.go", ParentPath: "src", IsDir: false},
				{Path: "src/file1_test.go", ParentPath: "src", IsDir: false},
				{Path: "src/file2_test.go", ParentPath: "src", IsDir: false},
				{Path: "src/file3_test.go", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"src": {},
			},
			wantViolCount: 0, // 3 non-test files, tests don't count
		},
		{
			name:     "GivenExcessiveNonTestFiles_ThenReturnsViolation",
			maxFiles: 2,
			files: []walker.FileInfo{
				{Path: "src/file1.go", ParentPath: "src", IsDir: false},
				{Path: "src/file2.go", ParentPath: "src", IsDir: false},
				{Path: "src/file3.go", ParentPath: "src", IsDir: false},
				{Path: "src/file1_test.go", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"src": {},
			},
			wantViolCount: 1, // 3 non-test files exceeds limit of 2
		},
		{
			name:     "GivenTypeScriptTestFiles_ThenExcludedFromCount",
			maxFiles: 3,
			files: []walker.FileInfo{
				{Path: "src/component.tsx", ParentPath: "src", IsDir: false},
				{Path: "src/utils.ts", ParentPath: "src", IsDir: false},
				{Path: "src/component.test.tsx", ParentPath: "src", IsDir: false},
				{Path: "src/utils.spec.ts", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"src": {},
			},
			wantViolCount: 0, // 2 non-test files
		},
		{
			name:     "GivenPythonTestFiles_ThenExcludedFromCount",
			maxFiles: 2,
			files: []walker.FileInfo{
				{Path: "src/module.py", ParentPath: "src", IsDir: false},
				{Path: "src/test_module.py", ParentPath: "src", IsDir: false},
				{Path: "src/module_test.py", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"src": {},
			},
			wantViolCount: 0, // 1 non-test file
		},
		{
			name:          "GivenEmptyDirectory_ThenReturnsNoViolations",
			maxFiles:      5,
			files:         []walker.FileInfo{},
			dirs:          map[string]*walker.DirInfo{"src": {}},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewMaxFilesRule(tt.maxFiles)

			// Act
			violations := rule.Check(tt.files, tt.dirs)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}

			for _, v := range violations {
				if v.Rule != "max-files-in-dir" {
					t.Errorf("violation rule = %v, want max-files-in-dir", v.Rule)
				}
			}
		})
	}
}

func TestMaxFilesRule_WhenCheckingIfTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Go
		{"file_test.go", true},
		{"file.go", false},

		// TypeScript/JavaScript
		{"component.test.ts", true},
		{"component.test.tsx", true},
		{"utils.spec.js", true},
		{"utils.spec.jsx", true},
		{"component.ts", false},

		// Python
		{"test_module.py", true},
		{"module_test.py", true},
		{"module.py", false},

		// Ruby
		{"file_spec.rb", true},
		{"file.rb", false},

		// Java
		{"FileTest.java", true},
		{"FileIT.java", true},
		{"File.java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Arrange
			rule := &MaxFilesRule{}

			// Act
			got := rule.isTestFile(tt.path)

			// Assert
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMaxFilesRule_WhenGettingName(t *testing.T) {
	// Arrange
	rule := NewMaxFilesRule(10)

	// Act
	got := rule.Name()

	// Assert
	if got != "max-files-in-dir" {
		t.Errorf("Name() = %v, want max-files-in-dir", got)
	}
}
