package rules

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// CompositeRule combines multiple rules with logical operators
type CompositeRule struct {
	name        string
	description string
	operator    CompositeOperator
	rules       []Rule
}

// CompositeOperator defines how rules are combined
type CompositeOperator int

const (
	// OperatorAND requires all rules to pass (no violations from any rule)
	OperatorAND CompositeOperator = iota

	// OperatorOR requires at least one rule to pass (at least one rule has no violations)
	OperatorOR

	// OperatorNOT inverts the first rule's result
	OperatorNOT

	// OperatorXOR requires exactly one rule to pass
	OperatorXOR
)

// NewCompositeRule creates a new composite rule
func NewCompositeRule(name, description string, operator CompositeOperator, rules ...Rule) *CompositeRule {
	return &CompositeRule{
		name:        name,
		description: description,
		operator:    operator,
		rules:       rules,
	}
}

// Name returns the rule name
func (r *CompositeRule) Name() string {
	return r.name
}

// Check executes the composite rule
func (r *CompositeRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	switch r.operator {
	case OperatorAND:
		return r.checkAND(files, dirs)
	case OperatorOR:
		return r.checkOR(files, dirs)
	case OperatorNOT:
		return r.checkNOT(files, dirs)
	case OperatorXOR:
		return r.checkXOR(files, dirs)
	default:
		return []Violation{{
			Rule:    r.name,
			Path:    ".",
			Message: fmt.Sprintf("Unknown composite operator: %d", r.operator),
		}}
	}
}

// checkAND requires all rules to pass (reports violations from all rules)
func (r *CompositeRule) checkAND(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var allViolations []Violation

	for _, rule := range r.rules {
		violations := rule.Check(files, dirs)
		// Prefix violations with composite rule name
		for _, v := range violations {
			v.Rule = fmt.Sprintf("%s[%s]", r.name, v.Rule)
			allViolations = append(allViolations, v)
		}
	}

	return allViolations
}

// checkOR requires at least one rule to pass (only reports violations if all rules fail)
func (r *CompositeRule) checkOR(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var allViolations [][]Violation

	// Collect violations from all rules
	for _, rule := range r.rules {
		violations := rule.Check(files, dirs)
		allViolations = append(allViolations, violations)
	}

	// Check if at least one rule passed (has no violations)
	for _, violations := range allViolations {
		if len(violations) == 0 {
			return nil // At least one rule passed - composite passes
		}
	}

	// All rules failed - create a combined violation
	var ruleNames []string
	for _, rule := range r.rules {
		ruleNames = append(ruleNames, rule.Name())
	}

	return []Violation{{
		Rule:    r.name,
		Path:    ".",
		Message: fmt.Sprintf("All constituent rules failed: %s", strings.Join(ruleNames, ", ")),
	}}
}

// checkNOT inverts the first rule's result
func (r *CompositeRule) checkNOT(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if len(r.rules) == 0 {
		return []Violation{{
			Rule:    r.name,
			Path:    ".",
			Message: "NOT operator requires at least one rule",
		}}
	}

	violations := r.rules[0].Check(files, dirs)

	if len(violations) > 0 {
		// Rule failed (has violations) - NOT inverts to pass
		return nil
	}

	// Rule passed (no violations) - NOT inverts to fail
	return []Violation{{
		Rule:    r.name,
		Path:    ".",
		Message: fmt.Sprintf("NOT(%s) failed: rule passed when it should have failed", r.rules[0].Name()),
	}}
}

// checkXOR requires exactly one rule to pass
func (r *CompositeRule) checkXOR(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	passCount := 0
	var failedRules []string

	for _, rule := range r.rules {
		violations := rule.Check(files, dirs)
		if len(violations) == 0 {
			passCount++
		} else {
			failedRules = append(failedRules, rule.Name())
		}
	}

	if passCount == 1 {
		return nil // Exactly one rule passed - XOR succeeds
	}

	if passCount == 0 {
		return []Violation{{
			Rule:    r.name,
			Path:    ".",
			Message: fmt.Sprintf("XOR failed: no rules passed (all failed: %s)", strings.Join(failedRules, ", ")),
		}}
	}

	return []Violation{{
		Rule:    r.name,
		Path:    ".",
		Message: fmt.Sprintf("XOR failed: %d rules passed (expected exactly 1)", passCount),
	}}
}

// --- Composite rule builders ---

// AllOf creates a composite rule that requires all sub-rules to pass
func AllOf(name, description string, rules ...Rule) Rule {
	return NewCompositeRule(name, description, OperatorAND, rules...)
}

// AnyOf creates a composite rule that requires at least one sub-rule to pass
func AnyOf(name, description string, rules ...Rule) Rule {
	return NewCompositeRule(name, description, OperatorOR, rules...)
}

// NotRule creates a composite rule that inverts a sub-rule's result
func NotRule(name, description string, rule Rule) Rule {
	return NewCompositeRule(name, description, OperatorNOT, rule)
}

// ExactlyOneOf creates a composite rule that requires exactly one sub-rule to pass
func ExactlyOneOf(name, description string, rules ...Rule) Rule {
	return NewCompositeRule(name, description, OperatorXOR, rules...)
}

// --- Example composite rules ---

// ExampleTestingStrategyRule ensures either unit tests or integration tests exist
func ExampleTestingStrategyRule() Rule {
	unitTestsExist := NewFileExistenceRule(map[string]string{
		"*_test.go": "Unit tests should exist",
	})

	integrationTestsExist := NewFileExistenceRule(map[string]string{
		"tests/integration/*_test.go": "Integration tests should exist",
	})

	return AnyOf(
		"testing-strategy",
		"Project must have either unit tests or integration tests",
		unitTestsExist,
		integrationTestsExist,
	)
}

// ExampleDocumentationCompletenessRule ensures all documentation types exist
func ExampleDocumentationCompletenessRule() Rule {
	readmeExists := NewFileExistenceRule(map[string]string{
		"README.md": "README must exist",
	})

	contributingExists := NewFileExistenceRule(map[string]string{
		"CONTRIBUTING.md": "CONTRIBUTING guide must exist",
	})

	licenseExists := NewFileExistenceRule(map[string]string{
		"LICENSE*": "LICENSE file must exist",
	})

	return AllOf(
		"documentation-completeness",
		"Project must have complete documentation",
		readmeExists,
		contributingExists,
		licenseExists,
	)
}

// ExampleArchitectureConsistencyRule ensures consistent architecture style
func ExampleArchitectureConsistencyRule() Rule {
	// This is a simplified example showing the concept
	cleanArchRule := NewFileExistenceRule(map[string]string{
		"internal/domain/*":         "Clean architecture: domain layer",
		"internal/application/*":    "Clean architecture: application layer",
		"internal/infrastructure/*": "Clean architecture: infrastructure layer",
	})

	hexagonalRule := NewFileExistenceRule(map[string]string{
		"internal/core/*":    "Hexagonal: core",
		"internal/adapters/*": "Hexagonal: adapters",
		"internal/ports/*":   "Hexagonal: ports",
	})

	return ExactlyOneOf(
		"architecture-consistency",
		"Project must follow exactly one architecture style",
		cleanArchRule,
		hexagonalRule,
	)
}

// --- Conditional rules ---

// ConditionalRule executes a rule only if a condition is met
type ConditionalRule struct {
	name      string
	condition func(files []walker.FileInfo, dirs map[string]*walker.DirInfo) bool
	rule      Rule
}

// NewConditionalRule creates a rule that only executes if a condition is met
func NewConditionalRule(name string, condition func([]walker.FileInfo, map[string]*walker.DirInfo) bool, rule Rule) *ConditionalRule {
	return &ConditionalRule{
		name:      name,
		condition: condition,
		rule:      rule,
	}
}

// Name returns the rule name
func (r *ConditionalRule) Name() string {
	return r.name
}

// Check executes the rule only if condition is met
func (r *ConditionalRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if !r.condition(files, dirs) {
		return nil // Condition not met - rule doesn't apply
	}

	return r.rule.Check(files, dirs)
}

// IfProjectHas creates a conditional rule that checks for file existence
func IfProjectHas(pattern string, rule Rule) Rule {
	return NewConditionalRule(
		fmt.Sprintf("if-has-%s", pattern),
		func(files []walker.FileInfo, dirs map[string]*walker.DirInfo) bool {
			for _, file := range files {
				// Use simple pattern matching
				if strings.Contains(file.Path, pattern) {
					return true
				}
			}
			return false
		},
		rule,
	)
}

// IfProjectLanguage creates a conditional rule based on project language
func IfProjectLanguage(extension string, rule Rule) Rule {
	return NewConditionalRule(
		fmt.Sprintf("if-language-%s", extension),
		func(files []walker.FileInfo, dirs map[string]*walker.DirInfo) bool {
			for _, file := range files {
				if strings.HasSuffix(file.Path, extension) {
					return true
				}
			}
			return false
		},
		rule,
	)
}
