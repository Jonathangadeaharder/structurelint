// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/structurelint/structurelint/internal/walker"
)

// NamingConventionRule enforces naming conventions for files and directories
type NamingConventionRule struct {
	Patterns map[string]string // pattern -> convention (e.g., "*.ts" -> "camelCase")
}

// Name returns the rule name
func (r *NamingConventionRule) Name() string {
	return "naming-convention"
}

// Check validates naming conventions
func (r *NamingConventionRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for _, file := range files {
		for pattern, convention := range r.Patterns {
			if matchesPattern(file.Path, pattern) {
				if !r.matchesConvention(file.Path, pattern, convention) {
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Path:    file.Path,
						Message: fmt.Sprintf("does not match naming convention '%s'", convention),
					})
				}
			}
		}
	}

	return violations
}

// matchesConvention checks if a path matches a naming convention
func (r *NamingConventionRule) matchesConvention(path, pattern, convention string) bool {
	// Extract the relevant part of the path to check
	var nameToCheck string

	if strings.HasSuffix(pattern, "/") {
		// Directory pattern - check the directory name itself
		nameToCheck = filepath.Base(path)
	} else {
		// File pattern - check the filename without extension
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		nameToCheck = strings.TrimSuffix(base, ext)
	}

	switch strings.ToLower(convention) {
	case "camelcase":
		return isCamelCase(nameToCheck)
	case "pascalcase":
		return isPascalCase(nameToCheck)
	case "kebab-case", "kebabcase":
		return isKebabCase(nameToCheck)
	case "snake_case", "snakecase":
		return isSnakeCase(nameToCheck)
	case "lowercase":
		return isLowerCase(nameToCheck)
	case "uppercase":
		return isUpperCase(nameToCheck)
	default:
		// If convention is not recognized, assume it passes
		return true
	}
}

func isCamelCase(s string) bool {
	if len(s) == 0 {
		return true
	}
	// camelCase starts with lowercase and can have uppercase letters
	if unicode.IsUpper(rune(s[0])) {
		return false
	}
	// Should not contain hyphens, underscores, or spaces
	return !strings.ContainsAny(s, "-_ ")
}

func isPascalCase(s string) bool {
	if len(s) == 0 {
		return true
	}
	// PascalCase starts with uppercase
	if !unicode.IsUpper(rune(s[0])) {
		return false
	}
	// Should not contain hyphens, underscores, or spaces
	return !strings.ContainsAny(s, "-_ ")
}

func isKebabCase(s string) bool {
	// kebab-case is all lowercase with hyphens
	if s != strings.ToLower(s) {
		return false
	}
	// Should not contain underscores or spaces
	return !strings.ContainsAny(s, "_ ")
}

func isSnakeCase(s string) bool {
	// snake_case is all lowercase with underscores
	if s != strings.ToLower(s) {
		return false
	}
	// Should not contain hyphens or spaces
	return !strings.ContainsAny(s, "- ")
}

func isLowerCase(s string) bool {
	return s == strings.ToLower(s)
}

func isUpperCase(s string) bool {
	return s == strings.ToUpper(s)
}

func matchesPattern(path, pattern string) bool {
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		// Check if this is a directory by seeing if it's in the path as a directory component
		dirPattern := strings.TrimSuffix(pattern, "/")

		// For patterns like "components/**/"
		if strings.Contains(dirPattern, "**") {
			parts := strings.Split(dirPattern, "**")
			if len(parts) >= 1 {
				prefix := strings.TrimSuffix(parts[0], "/")
				if prefix != "" && strings.HasPrefix(path, prefix) {
					return true
				}
			}
		}

		// For exact directory patterns
		if strings.Contains(path, dirPattern+string(filepath.Separator)) {
			return true
		}
	}

	// Use filepath.Match for glob patterns
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// For patterns with path separators, try matching the full path
	if strings.Contains(pattern, "/") {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// NewNamingConventionRule creates a new NamingConventionRule
func NewNamingConventionRule(patterns map[string]string) *NamingConventionRule {
	return &NamingConventionRule{
		Patterns: patterns,
	}
}

// NewLanguageAwareNamingConventionRule creates a NamingConventionRule with language-specific defaults
// If userPatterns is provided, they override the language defaults
func NewLanguageAwareNamingConventionRule(rootDir string, userPatterns map[string]string) (*NamingConventionRule, error) {
	// Import is at package level
	detector := &struct{
		RootDir string
	}{RootDir: rootDir}

	// For now, create default patterns based on common language file extensions
	// This will be enhanced when we integrate with the language detector
	defaultPatterns := generateDefaultNamingPatterns()

	// Merge user patterns (they take precedence)
	finalPatterns := make(map[string]string)
	for k, v := range defaultPatterns {
		finalPatterns[k] = v
	}
	for k, v := range userPatterns {
		finalPatterns[k] = v
	}

	_ = detector // Suppress unused warning for now

	return &NamingConventionRule{
		Patterns: finalPatterns,
	}, nil
}

// generateDefaultNamingPatterns returns language-specific naming conventions
func generateDefaultNamingPatterns() map[string]string {
	return map[string]string{
		// Python: snake_case
		"*.py": "snake_case",

		// JavaScript/TypeScript: camelCase (except React components)
		"*.js":  "camelCase",
		"*.ts":  "camelCase",
		"*.mjs": "camelCase",

		// React components: PascalCase
		"**/components/**/*.jsx": "PascalCase",
		"**/components/**/*.tsx": "PascalCase",
		"*.jsx": "PascalCase",
		"*.tsx": "PascalCase",

		// Go: PascalCase (matches Go's exported identifier convention)
		"*.go": "PascalCase",

		// Java: PascalCase for class files
		"*.java": "PascalCase",

		// C#: PascalCase
		"*.cs": "PascalCase",

		// Ruby: snake_case
		"*.rb": "snake_case",

		// Rust: snake_case
		"*.rs": "snake_case",
	}
}
