// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// TestAdjacencyRule enforces test file existence requirements
type TestAdjacencyRule struct {
	Pattern      string   // "adjacent" or "separate"
	TestDir      string   // Directory for separate tests (e.g., "tests", "test")
	FilePatterns []string // Patterns to check (e.g., "**/*.go", "**/*.ts")
	Exemptions   []string // Patterns to exempt from checking
}

// Name returns the rule name
func (r *TestAdjacencyRule) Name() string {
	return "test-adjacency"
}

// Check validates test file adjacency requirements
func (r *TestAdjacencyRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	switch r.Pattern {
	case "adjacent":
		violations = r.checkAdjacentPattern(files)
	case "separate":
		violations = r.checkSeparatePattern(files)
	}

	return violations
}

// checkAdjacentPattern validates that source files have adjacent test files
func (r *TestAdjacencyRule) checkAdjacentPattern(files []walker.FileInfo) []Violation {
	var violations []Violation

	// Group files by directory
	filesByDir := make(map[string][]walker.FileInfo)
	for _, file := range files {
		if !file.IsDir {
			dir := file.ParentPath
			filesByDir[dir] = append(filesByDir[dir], file)
		}
	}

	// Check each source file for corresponding test file
	for _, file := range files {
		if v := r.checkAdjacentFile(file, filesByDir); v != nil {
			violations = append(violations, *v)
		}
	}

	return violations
}

func (r *TestAdjacencyRule) checkAdjacentFile(file walker.FileInfo, filesByDir map[string][]walker.FileInfo) *Violation {
	if file.IsDir {
		return nil
	}

	// Check if this file matches any of our patterns
	if !r.matchesFilePattern(file.Path) {
		return nil
	}

	// Skip if exempted
	if r.isExempted(file.Path) {
		return nil
	}

	// Skip if this IS a test file
	if r.isTestFile(file.Path) {
		return nil
	}

	// Check if file has directive to ignore this rule
	hasIgnoreDirective, reason := ShouldIgnoreFile(file, r.Name())

	// Look for corresponding test file
	testFileName := r.getTestFileName(file.Path)
	hasTest := false

	for _, f := range filesByDir[file.ParentPath] {
		if f.Path == testFileName {
			hasTest = true
			break
		}
	}

	if !hasTest && !hasIgnoreDirective {
		return &Violation{
			Rule:    r.Name(),
			Path:    file.Path,
			Message: "missing adjacent test file",
		}
	}

	if !hasTest && hasIgnoreDirective && reason != "" {
		return &Violation{
			Rule:    r.Name(),
			Path:    file.Path,
			Message: fmt.Sprintf("missing adjacent test file (exemption reason: %s)", reason),
		}
	}

	return nil
}

// checkSeparatePattern validates that source files have tests in separate test directory
func (r *TestAdjacencyRule) checkSeparatePattern(files []walker.FileInfo) []Violation {
	var violations []Violation

	// Build map of source files and test files
	sourceFiles := make(map[string]walker.FileInfo)
	testFiles := make(map[string]bool)

	r.categorizeFiles(files, sourceFiles, testFiles)

	// Check each source file for corresponding test in test directory
	for sourcePath, sourceFile := range sourceFiles {
		if v := r.checkSeparateFile(sourcePath, sourceFile, testFiles); v != nil {
			violations = append(violations, *v)
		}
	}

	return violations
}

func (r *TestAdjacencyRule) categorizeFiles(files []walker.FileInfo, sourceFiles map[string]walker.FileInfo, testFiles map[string]bool) {
	for _, file := range files {
		if file.IsDir {
			continue
		}

		if !r.matchesFilePattern(file.Path) {
			continue
		}

		// Check if file is in test directory
		if strings.HasPrefix(file.Path, r.TestDir+"/") {
			testFiles[file.Path] = true
		} else if !r.isTestFile(file.Path) && !r.isExempted(file.Path) {
			sourceFiles[file.Path] = file
		}
	}
}

func (r *TestAdjacencyRule) checkSeparateFile(sourcePath string, sourceFile walker.FileInfo, testFiles map[string]bool) *Violation {
	// Check if file has directive to ignore this rule
	if hasIgnore, _ := ShouldIgnoreFile(sourceFile, r.Name()); hasIgnore {
		return nil
	}

	expectedTestPath := r.getExpectedTestPath(sourcePath)

	if !testFiles[expectedTestPath] {
		return &Violation{
			Rule:    r.Name(),
			Path:    sourcePath,
			Message: fmt.Sprintf("missing test file '%s' in '%s/' directory (or add @structurelint:no-test/@structurelint:ignore directive)", filepath.Base(expectedTestPath), r.TestDir),
		}
	}

	return nil
}

// matchesFilePattern checks if a file matches any of the configured patterns
func (r *TestAdjacencyRule) matchesFilePattern(path string) bool {
	if len(r.FilePatterns) == 0 {
		return false
	}

	for _, pattern := range r.FilePatterns {
		if matchesGlobPattern(path, pattern) {
			return true
		}
	}
	return false
}

// isExempted checks if a file is exempted from test requirements
func (r *TestAdjacencyRule) isExempted(path string) bool {
	for _, exemption := range r.Exemptions {
		if matchesGlobPattern(path, exemption) {
			return true
		}
	}
	return false
}

// isTestFile checks if a file is a test file
func (r *TestAdjacencyRule) isTestFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Common test file patterns
	patterns := []string{
		"_test",    // Go: file_test.go
		".test",    // TypeScript: file.test.ts
		".spec",    // JavaScript/TypeScript: file.spec.js
		"_spec",    // Ruby: file_spec.rb
		"Test",     // Java: FileTest.java
	}

	for _, pattern := range patterns {
		if strings.Contains(nameWithoutExt, pattern) {
			return true
		}
	}

	return false
}

// getTestFileName returns the expected test file name for a source file
func (r *TestAdjacencyRule) getTestFileName(sourcePath string) string {
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Determine test file pattern based on extension
	switch ext {
	case ".go":
		return nameWithoutExt + "_test" + ext
	case ".ts", ".tsx":
		return nameWithoutExt + ".test" + ext
	case ".js", ".jsx":
		return nameWithoutExt + ".spec" + ext
	case ".py":
		return "test_" + base
	default:
		return nameWithoutExt + "_test" + ext
	}
}

// getExpectedTestPath returns the expected test file path in the separate test directory
func (r *TestAdjacencyRule) getExpectedTestPath(sourcePath string) string {
	dir := filepath.Dir(sourcePath)

	// Get test file name
	testFileName := r.getTestFileName(sourcePath)

	// Construct path in test directory, mirroring source structure
	if dir == "." || dir == "" {
		return filepath.Join(r.TestDir, testFileName)
	}

	return filepath.Join(r.TestDir, dir, testFileName)
}

// NewTestAdjacencyRule creates a new TestAdjacencyRule
func NewTestAdjacencyRule(pattern, testDir string, filePatterns, exemptions []string) *TestAdjacencyRule {
	return &TestAdjacencyRule{
		Pattern:      pattern,
		TestDir:      testDir,
		FilePatterns: filePatterns,
		Exemptions:   exemptions,
	}
}
