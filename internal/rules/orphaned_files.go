package rules

import (
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// OrphanedFilesRule detects files that are not imported by any other file
type OrphanedFilesRule struct {
	Graph              *graph.ImportGraph
	Entrypoints        []string // Top-level entrypoints (backward compatibility)
	EntryPointPatterns []string // Additional entry point patterns from rule config
}

// Name returns the rule name
func (r *OrphanedFilesRule) Name() string {
	return "disallow-orphaned-files"
}

// Check validates that no files are orphaned (unreferenced)
func (r *OrphanedFilesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if r.Graph == nil {
		return []Violation{}
	}

	var violations []Violation

	for _, filePath := range r.Graph.AllFiles {
		// Skip entrypoints
		if r.isEntrypoint(filePath) {
			continue
		}

		// Skip configuration and documentation files
		if r.isConfigOrDocFile(filePath) {
			continue
		}

		// Check if file has any incoming references
		refCount, exists := r.Graph.IncomingRefs[filePath]
		if !exists || refCount == 0 {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    filePath,
				Message: "file is orphaned (not imported by any other file)",
			})
		}
	}

	return violations
}

// isEntrypoint checks if a file is an entrypoint
func (r *OrphanedFilesRule) isEntrypoint(filePath string) bool {
	// Check against configured entrypoints (top-level config, backward compatibility)
	for _, pattern := range r.Entrypoints {
		if matchesEntrypointPattern(filePath, pattern) {
			return true
		}
	}

	// Check against entry point patterns from rule config
	for _, pattern := range r.EntryPointPatterns {
		if matchesEntrypointPattern(filePath, pattern) {
			return true
		}
	}

	// Common entrypoint patterns
	base := filepath.Base(filePath)
	commonEntrypoints := []string{
		"main.go",
		"main.ts",
		"main.js",
		"main.py",
		"index.ts",
		"index.js",
		"app.ts",
		"app.js",
		"app.py",
		"__init__.py",
		"manage.py", // Django management script
	}

	for _, entry := range commonEntrypoints {
		if base == entry {
			return true
		}
	}

	// Test files are entrypoints
	if strings.Contains(filePath, "_test.") || strings.Contains(filePath, ".test.") ||
		strings.Contains(filePath, ".spec.") || strings.HasSuffix(filePath, "_test.go") {
		return true
	}

	return false
}

// matchesEntrypointPattern checks if a file matches an entrypoint pattern
func matchesEntrypointPattern(filePath, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(filePath, prefix) {
				return false
			}

			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(filePath))
				return matched
			}

			return true
		}
	}

	// Exact match
	if filePath == pattern {
		return true
	}

	// Glob match
	matched, err := filepath.Match(pattern, filepath.Base(filePath))
	if err == nil && matched {
		return true
	}

	return false
}

// NewOrphanedFilesRule creates a new OrphanedFilesRule
func NewOrphanedFilesRule(importGraph *graph.ImportGraph, entrypoints []string) *OrphanedFilesRule {
	return &OrphanedFilesRule{
		Graph:              importGraph,
		Entrypoints:        entrypoints,
		EntryPointPatterns: []string{},
	}
}

// WithEntryPointPatterns adds entry point patterns to the rule
func (r *OrphanedFilesRule) WithEntryPointPatterns(patterns []string) *OrphanedFilesRule {
	r.EntryPointPatterns = patterns
	return r
}

// isConfigOrDocFile checks if a file is a configuration or documentation file
func (r *OrphanedFilesRule) isConfigOrDocFile(filePath string) bool {
	base := filepath.Base(filePath)

	// Configuration files
	configFiles := []string{
		".structurelint.yml",
		".structurelint.yaml",
		"package.json",
		"tsconfig.json",
		"go.mod",
		"go.sum",
		"setup.py",
		"pyproject.toml",
		"Makefile",
		".gitignore",
		".eslintrc",
		".prettierrc",
	}

	for _, config := range configFiles {
		if base == config || strings.HasPrefix(base, config) {
			return true
		}
	}

	// Documentation files
	if strings.HasSuffix(base, ".md") || strings.HasSuffix(base, ".txt") {
		return true
	}

	return false
}
