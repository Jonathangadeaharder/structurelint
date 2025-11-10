// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// FileExistenceRule validates file existence requirements
type FileExistenceRule struct {
	Requirements map[string]string // pattern -> requirement (e.g., "index.ts" -> "exists:1")
}

// Name returns the rule name
func (r *FileExistenceRule) Name() string {
	return "file-existence"
}

// Check validates file existence requirements
func (r *FileExistenceRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Group files by directory
	filesByDir := make(map[string][]walker.FileInfo)
	for _, file := range files {
		if !file.IsDir {
			dir := file.ParentPath
			filesByDir[dir] = append(filesByDir[dir], file)
		}
	}

	// Check each directory against requirements
	for dir := range dirs {
		for pattern, requirement := range r.Requirements {
			if err := r.checkRequirement(dir, pattern, requirement, filesByDir[dir]); err != nil {
				displayPath := dir
				if displayPath == "" {
					displayPath = "."
				}
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    displayPath,
					Message: err.Error(),
				})
			}
		}
	}

	return violations
}

// checkRequirement checks a single file existence requirement for a directory
func (r *FileExistenceRule) checkRequirement(dir, pattern, requirement string, dirFiles []walker.FileInfo) error {
	// Parse the requirement (e.g., "exists:1", "exists:0", "exists:1-10")
	parts := strings.Split(requirement, ":")
	if len(parts) != 2 || parts[0] != "exists" {
		return fmt.Errorf("invalid requirement format: %s", requirement)
	}

	countSpec := parts[1]
	minCount, maxCount := r.parseCountSpec(countSpec)

	// Handle special .dir and .file patterns
	var matchCount int
	if pattern == ".dir" {
		// Count subdirectories
		for _, file := range dirFiles {
			if file.IsDir && file.ParentPath == dir {
				matchCount++
			}
		}
	} else {
		// Count matching files
		// Handle OR patterns like "index.ts|index.js"
		patterns := strings.Split(pattern, "|")
		matched := make(map[string]bool)

		for _, file := range dirFiles {
			if file.ParentPath != dir {
				continue
			}
			for _, p := range patterns {
				if r.fileMatchesPattern(file, p) && !matched[file.Path] {
					matchCount++
					matched[file.Path] = true
					break
				}
			}
		}
	}

	// Check if count is within range
	if matchCount < minCount {
		return fmt.Errorf("requires at least %d file(s) matching '%s', found %d", minCount, pattern, matchCount)
	}
	if maxCount >= 0 && matchCount > maxCount {
		return fmt.Errorf("requires at most %d file(s) matching '%s', found %d", maxCount, pattern, matchCount)
	}

	return nil
}

// parseCountSpec parses count specifications like "1", "0", "1-10"
func (r *FileExistenceRule) parseCountSpec(spec string) (min, max int) {
	if strings.Contains(spec, "-") {
		parts := strings.Split(spec, "-")
		min, _ = strconv.Atoi(parts[0])
		max, _ = strconv.Atoi(parts[1])
		return min, max
	}

	count, _ := strconv.Atoi(spec)
	return count, count
}

// fileMatchesPattern checks if a file matches a pattern
func (r *FileExistenceRule) fileMatchesPattern(file walker.FileInfo, pattern string) bool {
	filename := filepath.Base(file.Path)

	// Exact match
	if filename == pattern {
		return true
	}

	// Glob match
	matched, err := filepath.Match(pattern, filename)
	if err == nil && matched {
		return true
	}

	return false
}

// NewFileExistenceRule creates a new FileExistenceRule
func NewFileExistenceRule(requirements map[string]string) *FileExistenceRule {
	return &FileExistenceRule{
		Requirements: requirements,
	}
}
