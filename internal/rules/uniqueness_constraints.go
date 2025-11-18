package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// UniquenessConstraintsRule ensures only one file in a directory matches a pattern
// This prevents "dual implementation anti-patterns" like:
//   - vocabulary_service.py and vocabulary_service_clean.py
//   - UserRepository.java and UserRepository2.java
type UniquenessConstraintsRule struct {
	// Map of pattern to constraint type
	// e.g., "*_service.py" -> "singleton"
	Constraints map[string]string
}

// Name returns the rule name
func (r *UniquenessConstraintsRule) Name() string {
	return "uniqueness-constraints"
}

// Check validates that only one file per directory matches each pattern
func (r *UniquenessConstraintsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Group files by directory
	filesByDir := make(map[string][]walker.FileInfo)
	for _, file := range files {
		if file.IsDir {
			continue
		}
		dir := filepath.Dir(file.Path)
		filesByDir[dir] = append(filesByDir[dir], file)
	}

	// Check each directory for constraint violations
	for dir, dirFiles := range filesByDir {
		for pattern, constraintType := range r.Constraints {
			switch constraintType {
			case "singleton", "unique":
				v := r.checkSingletonConstraint(dir, dirFiles, pattern)
				violations = append(violations, v...)
			}
		}
	}

	return violations
}

// checkSingletonConstraint checks that only one file matches the pattern in the directory
func (r *UniquenessConstraintsRule) checkSingletonConstraint(dir string, files []walker.FileInfo, pattern string) []Violation {
	var violations []Violation
	var matchingFiles []string

	for _, file := range files {
		filename := filepath.Base(file.Path)
		if matchesUniquenessPattern(filename, pattern) {
			matchingFiles = append(matchingFiles, file.Path)
		}
	}

	// If more than one file matches, report violation
	if len(matchingFiles) > 1 {
		fileList := strings.Join(matchingFiles, ", ")
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    dir,
			Message: fmt.Sprintf("multiple files match pattern '%s' (singleton constraint violated): %s", pattern, fileList),
		})
	}

	return violations
}

// matchesUniquenessPattern checks if a filename matches a pattern
// Supports glob patterns like *_service.py
func matchesUniquenessPattern(filename, pattern string) bool {
	// Simple glob matching
	matched, err := filepath.Match(pattern, filename)
	if err == nil && matched {
		return true
	}

	// Also support suffix matching for convenience
	// e.g., "*_service.py" matches "user_service.py"
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(filename, suffix)
	}

	return false
}

// NewUniquenessConstraintsRule creates a new UniquenessConstraintsRule
func NewUniquenessConstraintsRule(constraints map[string]string) *UniquenessConstraintsRule {
	return &UniquenessConstraintsRule{
		Constraints: constraints,
	}
}
