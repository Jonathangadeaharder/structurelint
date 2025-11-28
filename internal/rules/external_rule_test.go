package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

type mockPlugin struct {
	name       string
	violations []Violation
	err        error
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Check(ctx context.Context, files []walker.FileInfo, config map[string]interface{}) ([]Violation, error) {
	return m.violations, m.err
}

func TestExternalRuleAdapter(t *testing.T) {
	// Test successful execution
	plugin := &mockPlugin{
		name: "test-plugin",
		violations: []Violation{
			{Rule: "test-plugin", Path: "file.go", Message: "violation"},
		},
	}

	adapter := NewExternalRuleAdapter(plugin, nil)
	if adapter.Name() != "test-plugin" {
		t.Errorf("Name() = %s, want test-plugin", adapter.Name())
	}

	violations := adapter.Check(nil, nil)
	if len(violations) != 1 {
		t.Errorf("Check() returned %d violations, want 1", len(violations))
	}

	// Test failure
	pluginErr := &mockPlugin{
		name: "fail-plugin",
		err:  errors.New("plugin failed"),
	}

	adapterErr := NewExternalRuleAdapter(pluginErr, nil)
	violationsErr := adapterErr.Check(nil, nil)
	if len(violationsErr) != 1 {
		t.Errorf("Check() returned %d violations on error, want 1", len(violationsErr))
	}
	if violationsErr[0].Message != "Plugin execution failed: plugin failed" {
		t.Errorf("Unexpected error message: %s", violationsErr[0].Message)
	}
}
