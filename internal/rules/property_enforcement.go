package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// PropertyEnforcementRule enforces various dependency management properties
// including cyclic dependency detection, dependency limits, and depth constraints
type PropertyEnforcementRule struct {
	Graph                  *graph.ImportGraph
	MaxDependenciesPerFile int
	MaxDependencyDepth     int
	DetectCycles           bool
	ForbiddenPatterns      []string // Patterns like "internal/** -> external/**"
}

// PropertyEnforcementConfig holds the configuration for the rule
type PropertyEnforcementConfig struct {
	MaxDependenciesPerFile int      `yaml:"max_dependencies_per_file"`
	MaxDependencyDepth     int      `yaml:"max_dependency_depth"`
	DetectCycles           bool     `yaml:"detect_cycles"`
	ForbiddenPatterns      []string `yaml:"forbidden_patterns"`
}

func (r *PropertyEnforcementRule) Name() string {
	return "property-enforcement"
}

func (r *PropertyEnforcementRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if r.Graph == nil {
		return []Violation{}
	}

	var violations []Violation

	// Check 1: Detect cyclic dependencies
	if r.DetectCycles {
		violations = append(violations, r.detectCycles()...)
	}

	// Check 2: Enforce max dependencies per file
	if r.MaxDependenciesPerFile > 0 {
		violations = append(violations, r.checkMaxDependencies()...)
	}

	// Check 3: Enforce max dependency depth
	if r.MaxDependencyDepth > 0 {
		violations = append(violations, r.checkDependencyDepth()...)
	}

	// Check 4: Check forbidden dependency patterns
	if len(r.ForbiddenPatterns) > 0 {
		violations = append(violations, r.checkForbiddenPatterns()...)
	}

	return violations
}

// detectCycles finds circular dependencies in the import graph
func (r *PropertyEnforcementRule) detectCycles() []Violation {
	var violations []Violation

	// Track visited nodes and current path for cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var currentPath []string

	var dfs func(string) bool
	dfs = func(file string) bool {
		visited[file] = true
		recStack[file] = true
		currentPath = append(currentPath, file)

		// Check all dependencies of this file
		if deps, exists := r.Graph.Dependencies[file]; exists {
			for _, dep := range deps {
				// Resolve dependency to actual file path
				depFile := r.resolveImportToFile(dep, file)
				if depFile == "" {
					continue
				}

				if !visited[depFile] {
					if dfs(depFile) {
						return true
					}
				} else if recStack[depFile] {
					// Cycle detected! Build the cycle path
					cycleStart := -1
					for i, p := range currentPath {
						if p == depFile {
							cycleStart = i
							break
						}
					}
					if cycleStart >= 0 {
						cyclePath := append(currentPath[cycleStart:], depFile)
						violations = append(violations, Violation{
							Rule:    r.Name(),
							Path:    file,
							Message: "cyclic dependency detected",
							Context: fmt.Sprintf("cycle: %s", strings.Join(cyclePath, " -> ")),
							Suggestions: []string{
								"Break the cycle by introducing an interface or abstraction",
								"Restructure the code to remove circular imports",
								"Consider dependency inversion principle",
							},
						})
					}
					return true
				}
			}
		}

		currentPath = currentPath[:len(currentPath)-1]
		recStack[file] = false
		return false
	}

	// Run DFS from each unvisited file
	for file := range r.Graph.Dependencies {
		if !visited[file] {
			currentPath = []string{}
			dfs(file)
		}
	}

	return violations
}

// checkMaxDependencies ensures no file has too many direct dependencies
func (r *PropertyEnforcementRule) checkMaxDependencies() []Violation {
	var violations []Violation

	for file, deps := range r.Graph.Dependencies {
		depCount := len(deps)
		if depCount > r.MaxDependenciesPerFile {
			violations = append(violations, Violation{
				Rule:     r.Name(),
				Path:     file,
				Message:  fmt.Sprintf("file has too many dependencies (%d > %d)", depCount, r.MaxDependenciesPerFile),
				Expected: fmt.Sprintf("at most %d dependencies", r.MaxDependenciesPerFile),
				Actual:   fmt.Sprintf("%d dependencies", depCount),
				Suggestions: []string{
					"Consider breaking this file into smaller, focused modules",
					"Use dependency injection to reduce coupling",
					"Look for common dependencies that could be grouped",
				},
			})
		}
	}

	return violations
}

// checkDependencyDepth ensures dependency chains don't exceed maximum depth
func (r *PropertyEnforcementRule) checkDependencyDepth() []Violation {
	var violations []Violation

	// Calculate the maximum dependency depth for each file
	depthCache := make(map[string]int)

	var calculateDepth func(string, map[string]bool) int
	calculateDepth = func(file string, visiting map[string]bool) int {
		// Check cache first
		if depth, cached := depthCache[file]; cached {
			return depth
		}

		// Check for cycles (if we're visiting this node already)
		if visiting[file] {
			return 0 // Cycle detected, return 0 to avoid infinite recursion
		}

		// Mark as visiting
		visiting[file] = true
		defer delete(visiting, file)

		// Get dependencies
		deps, exists := r.Graph.Dependencies[file]
		if !exists || len(deps) == 0 {
			depthCache[file] = 0
			return 0
		}

		// Find maximum depth among all dependencies
		maxDepth := 0
		for _, dep := range deps {
			depFile := r.resolveImportToFile(dep, file)
			if depFile == "" {
				continue
			}

			depth := calculateDepth(depFile, visiting)
			if depth > maxDepth {
				maxDepth = depth
			}
		}

		result := maxDepth + 1
		depthCache[file] = result
		return result
	}

	// Check depth for each file
	for file := range r.Graph.Dependencies {
		depth := calculateDepth(file, make(map[string]bool))
		if depth > r.MaxDependencyDepth {
			violations = append(violations, Violation{
				Rule:     r.Name(),
				Path:     file,
				Message:  fmt.Sprintf("dependency chain too deep (%d > %d)", depth, r.MaxDependencyDepth),
				Expected: fmt.Sprintf("dependency depth at most %d", r.MaxDependencyDepth),
				Actual:   fmt.Sprintf("dependency depth %d", depth),
				Suggestions: []string{
					"Flatten the dependency hierarchy",
					"Consider using facades or abstraction layers",
					"Review the architectural design for unnecessary layering",
				},
			})
		}
	}

	return violations
}

// checkForbiddenPatterns validates that forbidden dependency patterns are not violated
func (r *PropertyEnforcementRule) checkForbiddenPatterns() []Violation {
	var violations []Violation

	for _, pattern := range r.ForbiddenPatterns {
		// Parse pattern: "source_pattern -> target_pattern"
		parts := strings.Split(pattern, "->")
		if len(parts) != 2 {
			// Report malformed pattern as a violation
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    "configuration",
				Message: fmt.Sprintf("invalid forbidden pattern syntax: '%s'", pattern),
				Context: "Forbidden pattern must be in the form 'source_pattern -> target_pattern'",
				Suggestions: []string{
					"Ensure the pattern contains exactly one '->' separating source and target patterns",
					"Example: 'internal/** -> external/**'",
					"Check your .structurelint.yml configuration",
				},
			})
			continue
		}

		sourcePattern := strings.TrimSpace(parts[0])
		targetPattern := strings.TrimSpace(parts[1])

		// Validate that patterns are not empty
		if sourcePattern == "" || targetPattern == "" {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    "configuration",
				Message: fmt.Sprintf("forbidden pattern has empty source or target: '%s'", pattern),
				Context: "Both source and target patterns must be non-empty",
				Suggestions: []string{
					"Ensure both patterns before and after '->' are non-empty",
					"Example: 'internal/** -> external/**'",
				},
			})
			continue
		}

		// Check all dependencies
		for file, deps := range r.Graph.Dependencies {
			// Check if source file matches source pattern
			if !matchPattern(file, sourcePattern) {
				continue
			}

			// Check each dependency
			for _, dep := range deps {
				depFile := r.resolveImportToFile(dep, file)
				if depFile == "" {
					continue
				}

				// Check if dependency matches forbidden target pattern
				if matchPattern(depFile, targetPattern) {
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Path:    file,
						Message: fmt.Sprintf("forbidden dependency from '%s' to '%s'", sourcePattern, targetPattern),
						Context: fmt.Sprintf("file '%s' imports '%s'", file, depFile),
						Suggestions: []string{
							"Restructure code to avoid this dependency",
							"Use dependency inversion or abstraction",
							fmt.Sprintf("Files matching '%s' should not depend on '%s'", sourcePattern, targetPattern),
						},
					})
				}
			}
		}
	}

	return violations
}

// resolveImportToFile attempts to resolve an import path to an actual file path
func (r *PropertyEnforcementRule) resolveImportToFile(importPath string, sourceFile string) string {
	// Get list of files to search - prefer AllFiles if available, otherwise use Dependencies keys
	var filesToSearch []string
	if len(r.Graph.AllFiles) > 0 {
		filesToSearch = r.Graph.AllFiles
	} else {
		// Fallback: collect all files from Dependencies map (keys + values)
		fileSet := make(map[string]bool)
		for file := range r.Graph.Dependencies {
			fileSet[file] = true
		}
		for _, deps := range r.Graph.Dependencies {
			for _, dep := range deps {
				fileSet[dep] = true
			}
		}
		for file := range fileSet {
			filesToSearch = append(filesToSearch, file)
		}
	}

	// First, check if the import path is already a valid file path
	// This handles cases where dependencies are already resolved file paths
	for _, file := range filesToSearch {
		if file == importPath {
			return importPath
		}
	}

	// For relative imports, resolve relative to source file
	if strings.HasPrefix(importPath, ".") {
		sourceDir := filepath.Dir(sourceFile)
		resolved := filepath.Join(sourceDir, importPath)

		// Check if this file exists in our graph
		for _, file := range filesToSearch {
			if file == resolved {
				return resolved
			}
		}

		// Try with common extensions
		for _, ext := range []string{".go", ".py", ".ts", ".js", ".java", ".cs", ".cpp", ".hpp"} {
			candidate := resolved + ext
			for _, file := range filesToSearch {
				if file == candidate {
					return candidate
				}
			}
		}
	}

	// For absolute imports, search in all files using suffix matching
	// Only use HasSuffix to avoid false matches (e.g., api/client matching api/client_test.go)
	for _, file := range filesToSearch {
		if strings.HasSuffix(file, importPath) {
			return file
		}
	}

	return ""
}

// matchPattern checks if a path matches a glob-like pattern
// Note: This is a simplified implementation that supports:
// - Patterns ending with '/**' (e.g., 'internal/**')
// - Patterns starting with '**/' (e.g., '**/test.go')
// - Simple glob patterns with '*' and '?' wildcards
// More complex patterns like 'foo/**/bar' are not currently supported.
func matchPattern(path, pattern string) bool {
	// Normalize paths
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Handle ** glob patterns (simplified implementation)
	if strings.Contains(pattern, "**") {
		// Pattern ending with /** matches any file under the prefix
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			return strings.HasPrefix(path, prefix)
		}
		// Pattern starting with **/ matches any file with the suffix
		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")
			return strings.HasSuffix(path, suffix) || strings.Contains(path, "/"+suffix)
		}
	}

	// Use filepath.Match for simple patterns
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}

	// Try matching the full path
	matched, _ = filepath.Match(pattern, path)
	return matched
}

// NewPropertyEnforcementRule creates a new PropertyEnforcementRule
func NewPropertyEnforcementRule(importGraph *graph.ImportGraph, config PropertyEnforcementConfig) *PropertyEnforcementRule {
	return &PropertyEnforcementRule{
		Graph:                  importGraph,
		MaxDependenciesPerFile: config.MaxDependenciesPerFile,
		MaxDependencyDepth:     config.MaxDependencyDepth,
		DetectCycles:           config.DetectCycles,
		ForbiddenPatterns:      config.ForbiddenPatterns,
	}
}
