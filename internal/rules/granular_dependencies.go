// @structurelint:ignore test-adjacency Granular dependency validation is tested through integration tests
package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// GranularDependencyRule validates fine-grained module-to-module dependencies
type GranularDependencyRule struct {
	Rules []config.DependencyRule
	Graph *graph.ImportGraph
}

// Name returns the rule name
func (r *GranularDependencyRule) Name() string {
	return "granular-dependencies"
}

// Check validates granular dependency rules
func (r *GranularDependencyRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if r.Graph == nil || len(r.Rules) == 0 {
		return []Violation{}
	}

	var violations []Violation

	// Check each dependency rule
	for _, rule := range r.Rules {
		// Handle pre-built rules
		if rule.Rule != "" {
			violations = append(violations, r.checkPrebuiltRule(rule)...)
			continue
		}

		// Handle from/to rules
		violations = append(violations, r.checkFromToRule(rule)...)
	}

	return violations
}

// checkPrebuiltRule validates pre-built rules like "no-circular"
func (r *GranularDependencyRule) checkPrebuiltRule(rule config.DependencyRule) []Violation {
	var violations []Violation

	switch rule.Rule {
	case "no-circular":
		violations = append(violations, r.checkCircularDependencies(rule)...)
	default:
		// Unknown pre-built rule, skip
	}

	return violations
}

// checkCircularDependencies detects circular dependencies
func (r *GranularDependencyRule) checkCircularDependencies(rule config.DependencyRule) []Violation {
	var violations []Violation

	// Build a map to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var checkCycle func(node string, path []string) bool
	checkCycle = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true

		// Add current node to path
		currentPath := append(path, node)

		// Check all dependencies
		if deps, ok := r.Graph.Dependencies[node]; ok {
			for _, dep := range deps {
				if !visited[dep] {
					if checkCycle(dep, currentPath) {
						return true
					}
				} else if recStack[dep] {
					// Found a cycle
					cyclePath := append(currentPath, dep)
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Path:    node,
						Message: fmt.Sprintf("circular dependency detected: %s", formatCyclePath(cyclePath)),
					})
					return true
				}
			}
		}

		recStack[node] = false
		return false
	}

	// Check all files
	for file := range r.Graph.Dependencies {
		if !visited[file] {
			checkCycle(file, []string{})
		}
	}

	return violations
}

// checkFromToRule validates from/to path restrictions
func (r *GranularDependencyRule) checkFromToRule(rule config.DependencyRule) []Violation {
	var violations []Violation

	// Check all dependencies
	for sourceFile, targets := range r.Graph.Dependencies {
		// Check if source matches "from" selector
		if !r.matchesSelector(sourceFile, rule.From) {
			continue
		}

		// Check each target
		for _, targetPath := range targets {
			// Resolve target to actual file
			var targetFile string
			for _, file := range r.Graph.AllFiles {
				if strings.Contains(file, targetPath) {
					targetFile = file
					break
				}
			}

			if targetFile == "" {
				continue
			}

			// Check if target matches "to" selector
			if r.matchesSelector(targetFile, rule.To) {
				severity := "error"
				if rule.Severity != "" {
					severity = rule.Severity
				}

				message := fmt.Sprintf("[%s] %s: dependency from '%s' to '%s' is not allowed",
					severity, rule.Name, sourceFile, targetFile)

				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    sourceFile,
					Message: message,
				})
			}
		}
	}

	return violations
}

// matchesSelector checks if a file path matches a dependency selector
func (r *GranularDependencyRule) matchesSelector(filePath string, selector config.DependencySelector) bool {
	// Check positive match
	if selector.Path != "" {
		matched, err := filepath.Match(selector.Path, filePath)
		if err != nil {
			return false
		}
		if !matched {
			// Try glob pattern
			if !matchesDependencyGlob(filePath, selector.Path) {
				return false
			}
		}
	}

	// Check negative match
	if selector.PathNot != "" {
		matched, err := filepath.Match(selector.PathNot, filePath)
		if err == nil && matched {
			return false
		}
		// Try glob pattern
		if matchesDependencyGlob(filePath, selector.PathNot) {
			return false
		}
	}

	return true
}

// matchesDependencyGlob checks if a path matches a glob pattern (supports **)
func matchesDependencyGlob(path, pattern string) bool {
	// Simple ** support
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}

			return true
		}
	}

	matched, _ := filepath.Match(pattern, path)
	return matched
}

// formatCyclePath formats a cycle path for display
func formatCyclePath(path []string) string {
	if len(path) == 0 {
		return ""
	}

	var formatted []string
	for _, p := range path {
		formatted = append(formatted, filepath.Base(p))
	}

	return strings.Join(formatted, " → ")
}

// NewGranularDependencyRule creates a new GranularDependencyRule
func NewGranularDependencyRule(rules []config.DependencyRule, importGraph *graph.ImportGraph) *GranularDependencyRule {
	return &GranularDependencyRule{
		Rules: rules,
		Graph: importGraph,
	}
}
