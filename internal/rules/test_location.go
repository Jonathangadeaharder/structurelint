package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// TestLocationRule validates that test files are in appropriate locations
// Test files should either be:
// 1. Adjacent to their source code, OR
// 2. In a designated integration test directory
type TestLocationRule struct {
	IntegrationTestDir string   // Directory for integration tests (e.g., "tests", "test")
	AllowAdjacent      bool     // Allow tests adjacent to source
	Exemptions         []string // Patterns to exempt from checking
}

// Name returns the rule name
func (r *TestLocationRule) Name() string {
	return "test-location"
}

// Check validates test file locations
func (r *TestLocationRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Build a map of source files (non-test files)
	sourceFiles := make(map[string]bool)
	for _, file := range files {
		if file.IsDir {
			continue
		}

		if !r.isTestFile(file.Path) {
			// Add source file with and without extension for matching
			base := filepath.Base(file.Path)
			ext := filepath.Ext(base)
			nameWithoutExt := strings.TrimSuffix(base, ext)
			dir := file.ParentPath

			// Store both full path and just the directory+name
			sourceFiles[file.Path] = true
			sourceFiles[filepath.Join(dir, nameWithoutExt)] = true
		}
	}

	// Check each test file
	for _, file := range files {
		if file.IsDir {
			continue
		}

		if !r.isTestFile(file.Path) {
			continue
		}

		if r.isExempted(file.Path) {
			continue
		}

		// Check if test file is in integration test directory
		inIntegrationDir := r.IntegrationTestDir != "" && strings.HasPrefix(file.Path, r.IntegrationTestDir+"/")

		if inIntegrationDir {
			// Test is in integration test directory - this is allowed
			continue
		}

		// Check if test is adjacent to source code
		if r.AllowAdjacent && r.hasAdjacentSource(file.Path, sourceFiles) {
			// Test is adjacent to its source - this is allowed
			continue
		}

		// Test file is misplaced
		message := "test file not adjacent to source code"
		if r.IntegrationTestDir != "" {
			message += fmt.Sprintf(" and not in integration test directory '%s/'", r.IntegrationTestDir)
		}

		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    file.Path,
			Message: message,
		})
	}

	return violations
}

// isTestFile checks if a file is a test file
func (r *TestLocationRule) isTestFile(path string) bool {
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
		"test_",    // Python: test_file.py
	}

	for _, pattern := range patterns {
		if strings.Contains(nameWithoutExt, pattern) || strings.Contains(base, pattern) {
			return true
		}
	}

	return false
}

// hasAdjacentSource checks if a test file has a corresponding source file in the same directory
func (r *TestLocationRule) hasAdjacentSource(testPath string, sourceFiles map[string]bool) bool {
	dir := filepath.Dir(testPath)
	base := filepath.Base(testPath)
	ext := filepath.Ext(base)

	// Extract the source file name from the test file name
	sourceFileName := r.getSourceFileName(base, ext)

	if sourceFileName == "" {
		return false
	}

	sourcePath := filepath.Join(dir, sourceFileName)

	// Check both exact match and without extension
	if sourceFiles[sourcePath] {
		return true
	}

	// Also check without extension
	nameWithoutExt := strings.TrimSuffix(sourceFileName, ext)
	return sourceFiles[filepath.Join(dir, nameWithoutExt)]
}

// getSourceFileName extracts the source file name from a test file name
func (r *TestLocationRule) getSourceFileName(testFileName, ext string) string {
	nameWithoutExt := strings.TrimSuffix(testFileName, ext)

	// Remove common test patterns
	patterns := map[string]string{
		"_test":  "",
		".test":  "",
		".spec":  "",
		"_spec":  "",
		"Test":   "",
		"test_":  "",
	}

	for pattern, replacement := range patterns {
		if strings.Contains(nameWithoutExt, pattern) {
			nameWithoutExt = strings.ReplaceAll(nameWithoutExt, pattern, replacement)
			return nameWithoutExt + ext
		}
	}

	return ""
}

// isExempted checks if a file is exempted from validation
func (r *TestLocationRule) isExempted(path string) bool {
	for _, exemption := range r.Exemptions {
		if matchesGlobPattern(path, exemption) {
			return true
		}
	}
	return false
}

// NewTestLocationRule creates a new TestLocationRule
func NewTestLocationRule(integrationTestDir string, allowAdjacent bool, exemptions []string) *TestLocationRule {
	return &TestLocationRule{
		IntegrationTestDir: integrationTestDir,
		AllowAdjacent:      allowAdjacent,
		Exemptions:         exemptions,
	}
}
