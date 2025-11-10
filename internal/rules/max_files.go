// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"

	"github.com/structurelint/structurelint/internal/walker"
)

// MaxFilesRule enforces a maximum number of files per directory
type MaxFilesRule struct {
	MaxFiles int
}

// Name returns the rule name
func (r *MaxFilesRule) Name() string {
	return "max-files-in-dir"
}

// Check validates the maximum files constraint
// Test files are excluded from the count
func (r *MaxFilesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Count non-test files per directory
	fileCountPerDir := make(map[string]int)
	for _, file := range files {
		if file.IsDir {
			continue
		}
		if r.isTestFile(file.Path) {
			continue
		}
		fileCountPerDir[file.ParentPath]++
	}

	// Check each directory
	for path := range dirs {
		count := fileCountPerDir[path]
		if count > r.MaxFiles {
			displayPath := path
			if displayPath == "" {
				displayPath = "."
			}
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    displayPath,
				Message: fmt.Sprintf("contains %d non-test files, exceeds maximum of %d", count, r.MaxFiles),
			})
		}
	}

	return violations
}

// isTestFile checks if a file is a test file based on naming conventions
func (r *MaxFilesRule) isTestFile(path string) bool {
	// Common test file patterns across languages
	return hasTestPattern(path, "_test.go") || // Go
		hasTestPattern(path, ".test.ts") || // TypeScript
		hasTestPattern(path, ".test.tsx") || // TypeScript JSX
		hasTestPattern(path, ".test.js") || // JavaScript
		hasTestPattern(path, ".test.jsx") || // JavaScript JSX
		hasTestPattern(path, ".spec.ts") || // TypeScript spec
		hasTestPattern(path, ".spec.tsx") || // TypeScript JSX spec
		hasTestPattern(path, ".spec.js") || // JavaScript spec
		hasTestPattern(path, ".spec.jsx") || // JavaScript JSX spec
		hasTestPattern(path, "_spec.rb") || // Ruby
		hasTestPattern(path, "test_") || // Python test_*.py
		hasTestPattern(path, "_test.py") || // Python *_test.py
		hasTestPattern(path, "Test.java") || // Java *Test.java
		hasTestPattern(path, "IT.java") // Java *IT.java (integration tests)
}

// hasTestPattern checks if the file path contains a test pattern
func hasTestPattern(path, pattern string) bool {
	return len(path) >= len(pattern) && path[len(path)-len(pattern):] == pattern ||
		   len(path) > len(pattern) && path[:len(pattern)] == pattern
}

// NewMaxFilesRule creates a new MaxFilesRule
func NewMaxFilesRule(maxFiles int) *MaxFilesRule {
	return &MaxFilesRule{
		MaxFiles: maxFiles,
	}
}
