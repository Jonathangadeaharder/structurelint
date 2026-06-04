package core

import (
	"testing"
)

func TestToViolation(t *testing.T) {
	result := CheckResult{
		Path:    "test.yml",
		Message: "missing step",
		Fix:     "add step",
	}

	viol := result.ToViolation()

	if viol.Rule != "github-workflows" {
		t.Errorf("expected rule 'github-workflows', got '%s'", viol.Rule)
	}
	if viol.Path != "test.yml" {
		t.Errorf("expected path 'test.yml', got '%s'", viol.Path)
	}
	if viol.Message != "missing step" {
		t.Errorf("expected message 'missing step', got '%s'", viol.Message)
	}
	if len(viol.Suggestions) != 1 || viol.Suggestions[0] != "add step" {
		t.Errorf("expected suggestions ['add step'], got %v", viol.Suggestions)
	}

	// Test without fix suggestion
	resultNoFix := CheckResult{
		Path:    "test.yml",
		Message: "missing step",
	}
	violNoFix := resultNoFix.ToViolation()
	if len(violNoFix.Suggestions) != 0 {
		t.Errorf("expected no suggestions, got %v", violNoFix.Suggestions)
	}
}
