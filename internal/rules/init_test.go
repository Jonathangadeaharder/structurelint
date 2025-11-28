package rules

import (
	"testing"
)

func TestInitRegistration(t *testing.T) {
	// Verify standard rules are registered
	expectedRules := []string{
		"max-depth",
		"max-files-in-dir",
		"max-subdirs",
		"file-existence",
		"regex-match",
		"disallowed-patterns",
		"enforce-layer-boundaries",
		"disallow-orphaned-files",
		"disallow-unused-exports",
	}

	for _, name := range expectedRules {
		if _, ok := GetFactory(name); !ok {
			t.Errorf("Rule '%s' is not registered", name)
		}
	}
}

func TestStandardRuleFactories(t *testing.T) {
	// Test max-depth factory
	factory, _ := GetFactory("max-depth")
	ctx := &RuleContext{
		Config: map[string]interface{}{"max": 5},
	}
	rule, err := factory(ctx)
	if err != nil {
		t.Errorf("max-depth factory failed: %v", err)
	}
	if rule == nil || rule.Name() != "max-depth" {
		t.Error("max-depth factory returned invalid rule")
	}

	// Test invalid config
	ctxInvalid := &RuleContext{
		Config: map[string]interface{}{},
	}
	_, err = factory(ctxInvalid)
	if err == nil {
		t.Error("max-depth factory should fail with missing max")
	}
}
