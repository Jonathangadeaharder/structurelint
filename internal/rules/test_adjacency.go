// Package rules provides rule implementations for structurelint.
package rules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
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
		if file.IsDir {
			continue
		}

		// Check if this file matches any of our patterns
		if !r.matchesFilePattern(file.Path) {
			continue
		}

		// Skip if exempted
		if r.isExempted(file.Path) {
			continue
		}

		// Skip if this IS a test file
		if r.isTestFile(file.Path) {
			continue
		}

		hasNoTestDirective, reason := r.hasNoTestDirective(file.AbsPath)

		// Look for corresponding test file
		testFileName := r.getTestFileName(file.Path)
		hasTest := false

		for _, f := range filesByDir[file.ParentPath] {
			if filepath.Base(f.Path) == testFileName {
				hasTest = true
				break
			}
		}

		// Validate consistency
		if hasNoTestDirective && hasTest {
			// File declares no test needed but has a test file - warn about inconsistency
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("declares @structurelint:no-test (%s) but test file '%s' exists - remove directive or test file", reason, testFileName),
			})
		} else if !hasNoTestDirective && !hasTest {
			// File should have a test but doesn't
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("missing adjacent test file '%s' (or add @structurelint:no-test directive with reason)", testFileName),
			})
		}
	}

	return violations
}

// checkSeparatePattern validates that source files have tests in separate test directory
func (r *TestAdjacencyRule) checkSeparatePattern(files []walker.FileInfo) []Violation {
	var violations []Violation

	// Build map of source files and test files
	sourceFiles := make(map[string]walker.FileInfo)
	testFiles := make(map[string]bool)

	for _, file := range files {
		if file.IsDir {
			continue
		}

		if !r.matchesFilePattern(file.Path) {
			continue
		}

		// Check if file is in test directory
		if strings.HasPrefix(file.Path, r.TestDir+"/") {
			// This is a test file
			// Extract the source path it should correspond to
			testFiles[file.Path] = true
		} else if !r.isTestFile(file.Path) && !r.isExempted(file.Path) {
			sourceFiles[file.Path] = file
		}
	}

	// Check each source file for corresponding test in test directory
	for sourcePath := range sourceFiles {
		expectedTestPath := r.getExpectedTestPath(sourcePath)

		if !testFiles[expectedTestPath] {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    sourcePath,
				Message: fmt.Sprintf("missing test file '%s' in '%s/' directory", filepath.Base(expectedTestPath), r.TestDir),
			})
		}
	}

	return violations
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

// hasNoTestDirective checks if a file contains the @structurelint:no-test directive
// Returns (hasDirective, reason)
func (r *TestAdjacencyRule) hasNoTestDirective(absPath string) (bool, string) {
	file, err := os.Open(absPath)
	if err != nil {
		return false, ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	// Only scan first 50 lines (directives should be at the top)
	for scanner.Scan() && lineCount < 50 {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Look for @structurelint:no-test directive
		if strings.Contains(line, "@structurelint:no-test") {
			// Extract reason after the directive
			parts := strings.SplitN(line, "@structurelint:no-test", 2)
			if len(parts) == 2 {
				reason := strings.TrimSpace(parts[1])
				if reason == "" {
					reason = "no reason provided"
				}
				return true, reason
			}
			return true, "no reason provided"
		}
	}

	return false, ""
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
