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

// Fix generates fixes for naming convention violations
func (r *NamingConventionRule) Fix(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Fix {
	var fixes []Fix

	for _, file := range files {
		for pattern, convention := range r.Patterns {
			if matchesPattern(file.Path, pattern) {
				if !r.matchesConvention(file.Path, pattern, convention) {
					// Generate a fix to rename the file
					newName := r.convertToConvention(file.Path, pattern, convention)
					if newName != "" && newName != file.Path {
						fixes = append(fixes, Fix{
							FilePath: file.Path,
							Action:   "rename",
							OldValue: file.AbsPath,
							NewValue: filepath.Join(filepath.Dir(file.AbsPath), filepath.Base(newName)),
						})
					}
				}
			}
		}
	}

	return fixes
}

// convertToConvention converts a filename to the specified convention
func (r *NamingConventionRule) convertToConvention(path, pattern, convention string) string {
	// Extract the relevant part of the path to convert
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Convert the name
	var converted string
	switch strings.ToLower(convention) {
	case "camelcase":
		converted = toCamelCase(nameWithoutExt)
	case "pascalcase":
		converted = toPascalCase(nameWithoutExt)
	case "kebab-case", "kebabcase":
		converted = toKebabCase(nameWithoutExt)
	case "snake_case", "snakecase":
		converted = toSnakeCase(nameWithoutExt)
	case "lowercase":
		converted = strings.ToLower(nameWithoutExt)
	case "uppercase":
		converted = strings.ToUpper(nameWithoutExt)
	default:
		return ""
	}

	// Reconstruct the full path
	newBase := converted + ext
	if dir == "." {
		return newBase
	}
	return filepath.Join(dir, newBase)
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	// Split on common delimiters
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}
	return result
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	words := splitWords(s)
	var result string
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return result
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "_")
}

// splitWords splits a string into words based on common delimiters and case changes
func splitWords(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if r == '-' || r == '_' || r == ' ' {
			// Delimiter found
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else if i > 0 && unicode.IsUpper(r) && unicode.IsLower(rune(s[i-1])) {
			// CamelCase boundary (lowercase to uppercase)
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

// NewNamingConventionRule creates a new NamingConventionRule
func NewNamingConventionRule(patterns map[string]string) *NamingConventionRule {
	return &NamingConventionRule{
		Patterns: patterns,
	}
}
