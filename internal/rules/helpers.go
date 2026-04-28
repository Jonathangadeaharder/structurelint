// Package rules provides shared helper functions for rule implementations.
package rules

import (
	"path/filepath"
	"strings"
)

// MatchesGlobPattern checks if a path matches a glob pattern including **
func MatchesGlobPattern(path, pattern string) bool {
	// Normalize paths to use forward slashes for consistent matching across platforms
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

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

// WorkflowFile represents a parsed GitHub Actions workflow file
type WorkflowFile struct {
	Name string                 `yaml:"name"`
	On   interface{}            `yaml:"on"`
	Jobs map[string]WorkflowJob `yaml:"jobs"`
}

// WorkflowJob represents a job in a GitHub Actions workflow
type WorkflowJob struct {
	Name  string           `yaml:"name"`
	RunsOn interface{}     `yaml:"runs-on"`
	Steps []WorkflowStep   `yaml:"steps"`
}

// WorkflowStep represents a step in a workflow job
type WorkflowStep struct {
	Name string                 `yaml:"name"`
	Uses string                 `yaml:"uses"`
	Run  string                 `yaml:"run"`
	If   string                 `yaml:"if"`
	With map[string]interface{} `yaml:"with"`
}
