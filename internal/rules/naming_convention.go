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
					// Extract filename for better error messages
					base := filepath.Base(file.Path)
					ext := filepath.Ext(base)
					nameToCheck := strings.TrimSuffix(base, ext)

					// Detect actual convention
					actualConvention := r.detectConvention(nameToCheck)

					// Generate fix suggestions
					suggestions := r.generateSuggestions(nameToCheck, convention, ext)

					violations = append(violations, Violation{
						Rule:        r.Name(),
						Path:        file.Path,
						Message:     fmt.Sprintf("does not match naming convention '%s'", convention),
						Expected:    convention,
						Actual:      actualConvention,
						Suggestions: suggestions,
						Context:     fmt.Sprintf("Pattern: %s", pattern),
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

// detectConvention attempts to detect what naming convention a string uses
func (r *NamingConventionRule) detectConvention(name string) string {
	if isCamelCase(name) {
		return "camelCase"
	}
	if isPascalCase(name) {
		return "PascalCase"
	}
	if isKebabCase(name) {
		return "kebab-case"
	}
	if isSnakeCase(name) {
		return "snake_case"
	}
	if isLowerCase(name) {
		return "lowercase"
	}
	if isUpperCase(name) {
		return "UPPERCASE"
	}
	return "unknown/mixed"
}

// generateSuggestions generates fix suggestions for naming convention violations
func (r *NamingConventionRule) generateSuggestions(name, targetConvention, ext string) []string {
	var suggestions []string

	// Generate the correct name based on target convention
	correctName := r.convertToConvention(name, targetConvention)

	// Add suggestion to rename
	suggestions = append(suggestions, fmt.Sprintf("Rename to '%s%s'", correctName, ext))

	// Add suggestion to exclude
	suggestions = append(suggestions, "Add to exclude patterns if intentional")

	// Add suggestion to override
	suggestions = append(suggestions, "Use override rule for this specific file/directory")

	return suggestions
}

// convertToConvention converts a name to the target convention
func (r *NamingConventionRule) convertToConvention(name, convention string) string {
	// Split name into words (handles camelCase, PascalCase, snake_case, kebab-case)
	words := splitIntoWords(name)

	switch strings.ToLower(convention) {
	case "camelcase":
		return toCamelCase(words)
	case "pascalcase":
		return toPascalCase(words)
	case "kebab-case", "kebabcase":
		return toKebabCase(words)
	case "snake_case", "snakecase":
		return toSnakeCase(words)
	case "lowercase":
		return strings.ToLower(strings.Join(words, ""))
	case "uppercase":
		return strings.ToUpper(strings.Join(words, ""))
	default:
		return name
	}
}

// splitIntoWords splits a name into words handling various conventions
func splitIntoWords(name string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range name {
		if r == '_' || r == '-' {
			// Snake_case or kebab-case separator
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else if unicode.IsUpper(r) && i > 0 {
			// PascalCase or camelCase word boundary
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			currentWord.WriteRune(r)
		} else {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// toCamelCase converts words to camelCase
func toCamelCase(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += strings.Title(strings.ToLower(words[i]))
	}
	return result
}

// toPascalCase converts words to PascalCase
func toPascalCase(words []string) string {
	var result string
	for _, word := range words {
		result += strings.Title(strings.ToLower(word))
	}
	return result
}

// toKebabCase converts words to kebab-case
func toKebabCase(words []string) string {
	var lower []string
	for _, word := range words {
		lower = append(lower, strings.ToLower(word))
	}
	return strings.Join(lower, "-")
}

// toSnakeCase converts words to snake_case
func toSnakeCase(words []string) string {
	var lower []string
	for _, word := range words {
		lower = append(lower, strings.ToLower(word))
	}
	return strings.Join(lower, "_")
}
