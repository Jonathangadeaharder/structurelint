package rules

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// RegexMatchRule validates files against regex patterns
type RegexMatchRule struct {
	Patterns map[string]string // file pattern -> regex requirement
}

// Name returns the rule name
func (r *RegexMatchRule) Name() string {
	return "regex-match"
}

// Check validates regex matching requirements
func (r *RegexMatchRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for _, file := range files {
		if file.IsDir {
			continue
		}

		for pattern, regexSpec := range r.Patterns {
			if matchesGlobPattern(file.Path, pattern) {
				if !r.matchesRegex(file, pattern, regexSpec) {
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Path:    file.Path,
						Message: fmt.Sprintf("does not match required pattern '%s'", regexSpec),
					})
				}
			}
		}
	}

	return violations
}

// matchesRegex checks if a file matches a regex requirement
func (r *RegexMatchRule) matchesRegex(file walker.FileInfo, pattern, regexSpec string) bool {
	// Handle special syntax: "regex:..." or "regex:!..." (negation)
	negate := false
	regexPattern := regexSpec

	if strings.HasPrefix(regexSpec, "regex:") {
		regexPattern = strings.TrimPrefix(regexSpec, "regex:")
	}

	if strings.HasPrefix(regexPattern, "!") {
		negate = true
		regexPattern = strings.TrimPrefix(regexPattern, "!")
	}

	// Handle directory substitution ${0}, ${1}, etc.
	regexPattern = r.substituteDirectories(file.Path, pattern, regexPattern)

	// Compile and match
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		// If regex is invalid, skip this check
		return true
	}

	filename := filepath.Base(file.Path)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	matched := re.MatchString(nameWithoutExt)

	// Apply negation if needed
	if negate {
		return !matched
	}

	return matched
}

// substituteDirectories replaces ${0}, ${1}, etc. with directory names from the path
func (r *RegexMatchRule) substituteDirectories(filePath, pattern, regexPattern string) string {
	// Extract directory components from the pattern and file path
	// For example: "components/*/Button.tsx" and "components/atoms/Button.tsx"
	// ${0} should be replaced with "atoms"

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(filePath, "/")

	// Find wildcard positions in pattern
	wildcardIndices := []int{}
	for i, part := range patternParts {
		if part == "*" || part == "**" {
			wildcardIndices = append(wildcardIndices, i)
		}
	}

	// Replace ${n} with corresponding directory from path
	result := regexPattern
	for i, idx := range wildcardIndices {
		if idx < len(pathParts) {
			placeholder := fmt.Sprintf("${%d}", i)
			replacement := pathParts[idx]
			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}

	return result
}

// NewRegexMatchRule creates a new RegexMatchRule
func NewRegexMatchRule(patterns map[string]string) *RegexMatchRule {
	return &RegexMatchRule{
		Patterns: patterns,
	}
}
