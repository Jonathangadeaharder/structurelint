package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// DisallowedPatternsRule prevents specific file or directory patterns
type DisallowedPatternsRule struct {
	Patterns []string
}

// Name returns the rule name
func (r *DisallowedPatternsRule) Name() string {
	return "disallowed-patterns"
}

// Check validates that disallowed patterns are not present
func (r *DisallowedPatternsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, pattern := range r.Patterns {
			if matchesGlobPattern(file.Path, pattern) {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("matches disallowed pattern '%s'", pattern),
				})
			}
		}
	}

	return violations
}

// matchesGlobPattern checks if a path matches a glob pattern including **
func matchesGlobPattern(path, pattern string) bool {
	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")

		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check prefix
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			// Check suffix
			if suffix != "" {
				// If suffix has a glob pattern, use filepath.Match
				if strings.ContainsAny(suffix, "*?[]") {
					matched, _ := filepath.Match(suffix, filepath.Base(path))
					return matched
				}
				// Otherwise check if path ends with suffix or contains it
				return strings.HasSuffix(path, suffix) || strings.Contains(path, suffix)
			}

			return true
		}
	}

	// Exact match
	if path == pattern {
		return true
	}

	// Try glob matching
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// Try matching full path
	matched, err = filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	return false
}

// NewDisallowedPatternsRule creates a new DisallowedPatternsRule
func NewDisallowedPatternsRule(patterns []string) *DisallowedPatternsRule {
	return &DisallowedPatternsRule{
		Patterns: patterns,
	}
}
