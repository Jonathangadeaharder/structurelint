// Package structure provides rule implementations for structurelint.
package structure

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// MaxSubdirsRule enforces a maximum number of subdirectories per directory
type MaxSubdirsRule struct {
	MaxSubdirs int
}

// Name returns the rule name
func (r *MaxSubdirsRule) Name() string {
	return "max-subdirs"
}

// Check validates the maximum subdirectories constraint
func (r *MaxSubdirsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	for path, dirInfo := range dirs {
		if dirInfo.SubdirCount > r.MaxSubdirs {
			displayPath := path
			if displayPath == "" {
				displayPath = "."
			}
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    displayPath,
				Message: fmt.Sprintf("contains %d subdirectories, exceeds maximum of %d", dirInfo.SubdirCount, r.MaxSubdirs),
			})
		}
	}

	return violations
}

// NewMaxSubdirsRule creates a new MaxSubdirsRule
func NewMaxSubdirsRule(maxSubdirs int) *MaxSubdirsRule {
	return &MaxSubdirsRule{
		MaxSubdirs: maxSubdirs,
	}
}
