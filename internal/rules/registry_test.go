package rules

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	// Test Register and GetFactory
	Register("test-rule", func(ctx *RuleContext) (Rule, error) {
		return nil, nil
	})

	factory, ok := GetFactory("test-rule")
	if !ok {
		t.Error("Expected factory to be registered")
	}
	if factory == nil {
		t.Error("Expected factory to be non-nil")
	}

	_, ok = GetFactory("non-existent")
	if ok {
		t.Error("Expected non-existent rule to return false")
	}
}

func TestRuleContext_Helpers(t *testing.T) {
	config := map[string]interface{}{
		"intVal":    42,
		"floatVal":  42.0,
		"stringVal": "test",
		"mapVal": map[string]interface{}{
			"key": "value",
		},
		"sliceVal": []interface{}{"a", "b"},
	}

	ctx := &RuleContext{Config: config}

	// Test GetInt
	if val, ok := ctx.GetInt("intVal"); !ok || val != 42 {
		t.Errorf("GetInt(intVal) = %v, %v; want 42, true", val, ok)
	}
	if val, ok := ctx.GetInt("floatVal"); !ok || val != 42 {
		t.Errorf("GetInt(floatVal) = %v, %v; want 42, true", val, ok)
	}
	if _, ok := ctx.GetInt("stringVal"); ok {
		t.Error("GetInt(stringVal) should be false")
	}

	// Test GetStringMap
	if val, ok := ctx.GetStringMap("mapVal"); !ok || val["key"] != "value" {
		t.Errorf("GetStringMap(mapVal) = %v, %v; want map[key:value], true", val, ok)
	}

	// Test GetStringSlice
	if val, ok := ctx.GetStringSlice("sliceVal"); !ok || len(val) != 2 {
		t.Errorf("GetStringSlice(sliceVal) = %v, %v; want [a b], true", val, ok)
	}
}
