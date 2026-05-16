// Package structure provides rule implementations for structurelint.
package structure

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// MaxDepthOverride raises (or lowers) the depth limit for files matching a glob.
type MaxDepthOverride struct {
	Pattern string
	Max     int
}

// MaxDepthRule enforces a maximum directory nesting depth, with optional
// per-glob overrides for paths like SvelteKit `src/routes/**` that are
// legitimately deeper than the project default.
type MaxDepthRule struct {
	MaxDepth  int
	Overrides []MaxDepthOverride
}

func (r *MaxDepthRule) Name() string {
	return "max-depth"
}

func (r *MaxDepthRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation
	for _, file := range files {
		limit := r.limitFor(file.Path)
		if file.Depth > limit {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("exceeds maximum depth of %d (depth: %d)", limit, file.Depth),
			})
		}
	}
	return violations
}

func (r *MaxDepthRule) limitFor(path string) int {
	for _, o := range r.Overrides {
		if rules.MatchesGlobPattern(path, o.Pattern) {
			return o.Max
		}
	}
	return r.MaxDepth
}

func NewMaxDepthRule(maxDepth int) *MaxDepthRule {
	return &MaxDepthRule{MaxDepth: maxDepth}
}

func NewMaxDepthRuleWithOverrides(maxDepth int, overrides []MaxDepthOverride) *MaxDepthRule {
	return &MaxDepthRule{MaxDepth: maxDepth, Overrides: overrides}
}
