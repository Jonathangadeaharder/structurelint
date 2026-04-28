// Package structure provides rule implementations for structurelint.
package structure

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
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
func (r *DisallowedPatternsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Separate patterns into disallowed and allowed (negations)
	var disallowedPatterns []string
	var allowedPatterns []string

	for _, pattern := range r.Patterns {
		if strings.HasPrefix(pattern, "!") {
			allowedPatterns = append(allowedPatterns, strings.TrimPrefix(pattern, "!"))
		} else {
			disallowedPatterns = append(disallowedPatterns, pattern)
		}
	}

	for _, file := range files {
		for _, pattern := range disallowedPatterns {
			if rules.MatchesGlobPattern(file.Path, pattern) {
				// Check if this file matches any allowed pattern (exceptions)
				isAllowed := false
				for _, allowPattern := range allowedPatterns {
					if rules.MatchesGlobPattern(file.Path, allowPattern) {
						isAllowed = true
						break
					}
				}

				if !isAllowed {
					violations = append(violations, rules.Violation{
						Rule:    r.Name(),
						Path:    file.Path,
						Message: fmt.Sprintf("matches disallowed pattern '%s'", pattern),
					})
				}
			}
		}
	}

	return violations
}

// NewDisallowedPatternsRule creates a new DisallowedPatternsRule
func NewDisallowedPatternsRule(patterns []string) *DisallowedPatternsRule {
	return &DisallowedPatternsRule{
		Patterns: patterns,
	}
}
