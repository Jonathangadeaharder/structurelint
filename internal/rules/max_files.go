package rules

import (
	"fmt"

	"github.com/structurelint/structurelint/internal/walker"
)

// MaxFilesRule enforces a maximum number of files per directory
type MaxFilesRule struct {
	MaxFiles int
}

// Name returns the rule name
func (r *MaxFilesRule) Name() string {
	return "max-files-in-dir"
}

// Check validates the maximum files constraint
func (r *MaxFilesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for path, dirInfo := range dirs {
		if dirInfo.FileCount > r.MaxFiles {
			displayPath := path
			if displayPath == "" {
				displayPath = "."
			}
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    displayPath,
				Message: fmt.Sprintf("contains %d files, exceeds maximum of %d", dirInfo.FileCount, r.MaxFiles),
			})
		}
	}

	return violations
}

// NewMaxFilesRule creates a new MaxFilesRule
func NewMaxFilesRule(maxFiles int) *MaxFilesRule {
	return &MaxFilesRule{
		MaxFiles: maxFiles,
	}
}
