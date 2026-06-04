package ci

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func TestInitRegistration(t *testing.T) {
	// Retrieve factory
	factory, ok := rules.GetFactory("github-workflows")
	if !ok {
		t.Fatal("github-workflows rule not registered in init")
	}

	ctx := &rules.RuleContext{
		Config: map[string]interface{}{
			"sveltekit": map[string]interface{}{
				"required-coverage": 85.0,
			},
			"golang": map[string]interface{}{
				"required-coverage": 90.0,
			},
			"python": "invalid-config-format",
		},
	}

	rule, err := factory(ctx)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	if rule.Name() != "github-workflows" {
		t.Errorf("expected rule name 'github-workflows', got '%s'", rule.Name())
	}
}

func TestExtractConfig(t *testing.T) {
	fullCfg := map[string]interface{}{
		"sub": map[string]interface{}{
			"key": "value",
		},
		"non-map": "string",
	}

	// Case 1: Valid sub config map
	sub := extractConfig("sub", fullCfg)
	if sub == nil || sub["key"] != "value" {
		t.Errorf("expected extractConfig to return map with key=value, got %v", sub)
	}

	// Case 2: Config is nil
	if res := extractConfig("sub", nil); res != nil {
		t.Errorf("expected nil for nil input, got %v", res)
	}

	// Case 3: Target is not a map
	if res := extractConfig("non-map", fullCfg); res != nil {
		t.Errorf("expected nil for non-map value, got %v", res)
	}

	// Case 4: Target does not exist
	if res := extractConfig("missing", fullCfg); res != nil {
		t.Errorf("expected nil for missing key, got %v", res)
	}
}
