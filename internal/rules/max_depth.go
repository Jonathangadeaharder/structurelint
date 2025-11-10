// Package rules provides rule implementations for structurelint.
//
// @structurelint:no-test Simple rule implementation tested via rules_test.go integration test
package rules

import (
	"fmt"

	"github.com/structurelint/structurelint/internal/walker"
)

// MaxDepthRule enforces a maximum directory nesting depth
type MaxDepthRule struct {
	MaxDepth int
}

// Name returns the rule name
func (r *MaxDepthRule) Name() string {
	return "max-depth"
}

// Check validates the maximum depth constraint
func (r *MaxDepthRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for _, file := range files {
		if file.Depth > r.MaxDepth {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("exceeds maximum depth of %d (depth: %d)", r.MaxDepth, file.Depth),
			})
		}
	}

	return violations
}

// NewMaxDepthRule creates a new MaxDepthRule
func NewMaxDepthRule(maxDepth int) *MaxDepthRule {
	return &MaxDepthRule{
		MaxDepth: maxDepth,
	}
}
