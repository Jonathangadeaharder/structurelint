// Package quality provides code quality metric rules for structurelint.
//
// Rules in this package:
// - max-cognitive-complexity: Checks function cognitive complexity against threshold
// - max-halstead-effort: Checks function Halstead effort against threshold
// - disallow-unused-exports: Detects files with exports but no importers
package quality

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

// Package-level storage for the registry to wire up the unused-exports rule.

func init() {
	rules.Register("max-cognitive-complexity", func(ctx *rules.RuleContext) (rules.Rule, error) {
		max, ok := ctx.GetInt("max")
		if !ok {
			return nil, fmt.Errorf("missing 'max' parameter for max-cognitive-complexity")
		}
		filePatterns, _ := ctx.GetStringSlice("file-patterns")

		rule := NewMaxCognitiveComplexityRule(max, filePatterns)

		if testMax, ok := ctx.GetInt("test-max"); ok && testMax > 0 {
			rule = rule.WithTestMax(testMax)
		}

		return rule, nil
	})

	rules.Register("max-halstead-effort", func(ctx *rules.RuleContext) (rules.Rule, error) {
		raw, ok := ctx.Config["max"]
		if !ok {
			return nil, fmt.Errorf("missing 'max' parameter for max-halstead-effort")
		}

		var max float64
		switch v := raw.(type) {
		case float64:
			max = v
		case int:
			max = float64(v)
		default:
			return nil, fmt.Errorf("invalid 'max' parameter for max-halstead-effort: must be a number")
		}

		if max <= 0 {
			return nil, fmt.Errorf("'max' must be positive for max-halstead-effort")
		}

		filePatterns, _ := ctx.GetStringSlice("file-patterns")
		return NewMaxHalsteadEffortRule(max, filePatterns), nil
	})
}
