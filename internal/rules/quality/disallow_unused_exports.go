// Disallow Unused Exports Rule
//
// Detects exported symbols that are never imported anywhere in the project.
// Uses the import graph for cross-file symbol resolution.
// Supports Go, Python, JavaScript, and TypeScript.
//
// @structurelint:ignore test-adjacency Covered by disallow_unused_exports_test.go
package quality

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// UnusedExportsRule detects exported symbols that are never imported
type UnusedExportsRule struct {
	ImportGraph    *graph.ImportGraph
	ExcludePatterns []string // Glob patterns for files to exclude from unused-export checking
	EntryPointPatterns []string // Files matching these patterns are considered "entry points" — their exports are always considered used
}

// Name returns the rule name
func (r *UnusedExportsRule) Name() string {
	return "disallow-unused-exports"
}

// Check validates that all exports are used somewhere
func (r *UnusedExportsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	if r.ImportGraph == nil {
		return nil
	}

	// Build a set of files that are imported (have dependents)
	importedFiles := r.buildImportedFiles()

	var violations []rules.Violation

	for filePath, exports := range r.ImportGraph.Exports {
		// Normalize path
		normalPath := filepath.ToSlash(filePath)

		// Skip excluded patterns
		if r.isExcluded(normalPath) {
			continue
		}

		// Skip entry points — their exports are considered public API
		if r.isEntryPoint(normalPath) {
			continue
		}

		// Collect all export names
		var exportNames []string
		for _, exp := range exports {
			exportNames = append(exportNames, exp.Names...)
		}
		if len(exportNames) == 0 {
			continue
		}

		// Check if this file is imported at all
		if !importedFiles[normalPath] {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    filePath,
				Message: fmt.Sprintf("exports %s but is never imported", formatExportNames(exportNames)),
			})
		}
	}

	return violations
}

// buildImportedFiles builds a set of files that are imported by at least one other file
func (r *UnusedExportsRule) buildImportedFiles() map[string]bool {
	imported := make(map[string]bool)

	for _, deps := range r.ImportGraph.Dependencies {
		for _, importPath := range deps {
			resolved := r.resolveImport(importPath)
			if resolved != "" {
				imported[resolved] = true
			}
		}
	}

	return imported
}

// resolveImport resolves an import path to a file in the project
func (r *UnusedExportsRule) resolveImport(importPath string) string {
	normalPath := filepath.ToSlash(importPath)

	// Direct match
	if r.fileExists(normalPath) {
		return normalPath
	}

	// Try with common extensions
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py"} {
		if r.fileExists(normalPath + ext) {
			return normalPath + ext
		}
	}

	// Try as directory/index pattern (directory/index.ts, etc.)
	dirIndexes := []string{
		"/index.ts", "/index.tsx", "/index.js", "/index.jsx",
	}
	for _, idx := range dirIndexes {
		candidate := normalPath + idx
		if r.fileExists(candidate) {
			return candidate
		}
	}

	return ""
}

func (r *UnusedExportsRule) fileExists(path string) bool {
	for _, file := range r.ImportGraph.AllFiles {
		if filepath.ToSlash(file) == path {
			return true
		}
	}
	return false
}

func (r *UnusedExportsRule) isExcluded(path string) bool {
	for _, pattern := range r.ExcludePatterns {
		if rules.MatchesGlobPattern(path, pattern) {
			return true
		}
	}
	return false
}

func (r *UnusedExportsRule) isEntryPoint(path string) bool {
	for _, pattern := range r.EntryPointPatterns {
		if rules.MatchesGlobPattern(path, pattern) {
			return true
		}
	}
	return false
}

// formatExportNames formats a list of export names for display
func formatExportNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return fmt.Sprintf("'%s'", names[0])
	}
	if len(names) <= 3 {
		quoted := make([]string, len(names))
		for i, name := range names {
			quoted[i] = fmt.Sprintf("'%s'", name)
		}
		return strings.Join(quoted, ", ")
	}
	return fmt.Sprintf("'%s', '%s', '%s' and %d more", names[0], names[1], names[2], len(names)-3)
}

// NewUnusedExportsRule creates a new UnusedExportsRule
func NewUnusedExportsRule(importGraph *graph.ImportGraph, excludePatterns, entryPointPatterns []string) *UnusedExportsRule {
	return &UnusedExportsRule{
		ImportGraph:      importGraph,
		ExcludePatterns:   excludePatterns,
		EntryPointPatterns: entryPointPatterns,
	}
}
