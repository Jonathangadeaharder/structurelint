package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/walker"
)

func TestNew(t *testing.T) {
	// Act
	linter := New()

	// Assert
	if linter == nil {
		t.Fatal("New() returned nil")
	}

	if linter.config != nil {
		t.Error("Expected config to be nil initially")
	}
}

func TestGetRuleConfig_Exists(t *testing.T) {
	// Arrange
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	// Act
	cfg, ok := linter.getRuleConfig("max-depth")

	// Assert
	if !ok {
		t.Error("Expected rule to exist")
	}

	if cfg == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestGetRuleConfig_NotExists(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	_, ok := linter.getRuleConfig("nonexistent-rule")
	if ok {
		t.Error("Expected rule not to exist")
	}
}

func TestGetRuleConfig_NilConfig(t *testing.T) {
	linter := &Linter{
		config: nil,
	}

	_, ok := linter.getRuleConfig("max-depth")
	if ok {
		t.Error("Expected rule not to exist when config is nil")
	}
}

func TestGetRuleConfig_NilRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: nil,
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	if ok {
		t.Error("Expected rule not to exist when Rules is nil")
	}
}

func TestGetRuleConfig_DisabledByZero(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": 0,
			},
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	if ok {
		t.Error("Expected rule to be disabled when value is 0")
	}
}

func TestGetRuleConfig_DisabledByFalse(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": false,
			},
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	if ok {
		t.Error("Expected rule to be disabled when value is false")
	}
}

func TestGetRuleConfig_EnabledByTrue(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"enforce-layer-boundaries": true,
			},
		},
	}

	_, ok := linter.getRuleConfig("enforce-layer-boundaries")
	if !ok {
		t.Error("Expected rule to be enabled when value is true")
	}
}

func TestIsRuleEnabled_Enabled(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	if !linter.isRuleEnabled("max-depth") {
		t.Error("Expected rule to be enabled")
	}
}

func TestIsRuleEnabled_Disabled(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": 0,
			},
		},
	}

	if linter.isRuleEnabled("max-depth") {
		t.Error("Expected rule to be disabled")
	}
}

func TestIsRuleEnabled_NotExists(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	if linter.isRuleEnabled("nonexistent-rule") {
		t.Error("Expected rule to be disabled when it doesn't exist")
	}
}

func TestLint_BasicRules(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple project structure
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a config file with max-depth rule
	configContent := `
root: true
rules:
  max-depth:
    max: 2
`
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a deeply nested file (should violate max-depth: 2)
	deepDir := filepath.Join(srcDir, "level1", "level2", "level3")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	deepFile := filepath.Join(deepDir, "file.ts")
	if err := os.WriteFile(deepFile, []byte("// test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run linter
	linter := New()
	violations, err := linter.Lint(tmpDir)

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Should have violations for exceeding max depth
	if len(violations) == 0 {
		t.Error("Expected violations for max-depth rule")
	}

	// Verify at least one violation is about depth
	hasDepthViolation := false
	for _, v := range violations {
		if v.Rule == "max-depth" {
			hasDepthViolation = true
			break
		}
	}

	if !hasDepthViolation {
		t.Error("Expected at least one max-depth violation")
	}
}

func TestLint_NoConfig(t *testing.T) {
	// Create a temporary directory without config
	tmpDir, err := os.MkdirTemp("", "linter-test-noconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple file
	testFile := filepath.Join(tmpDir, "test.ts")
	if err := os.WriteFile(testFile, []byte("// test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run linter (should not fail even without config)
	linter := New()
	violations, err := linter.Lint(tmpDir)

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Should have no violations (no rules configured)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations without config, got %d", len(violations))
	}
}

func TestLint_WithLayers(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "linter-layers-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	domainDir := filepath.Join(tmpDir, "src", "domain")
	appDir := filepath.Join(tmpDir, "src", "application")

	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create domain file that violates layer boundaries by importing from application
	domainFile := filepath.Join(domainDir, "user.ts")
	domainContent := `
import { UserService } from '../application/userService';
export class User {}
`
	if err := os.WriteFile(domainFile, []byte(domainContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create application file
	appFile := filepath.Join(appDir, "userService.ts")
	appContent := `
import { User } from '../domain/user';
export class UserService {}
`
	if err := os.WriteFile(appFile, []byte(appContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config with layers
	configContent := `
root: true
rules:
  enforce-layer-boundaries:
    enabled: true
layers:
  - name: domain
    path: src/domain/**
    dependsOn: []
  - name: application
    path: src/application/**
    dependsOn: [domain]
`
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run linter
	linter := New()
	violations, err := linter.Lint(tmpDir)

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Should have violations for layer boundaries
	hasLayerViolation := false
	for _, v := range violations {
		if v.Rule == "enforce-layer-boundaries" {
			hasLayerViolation = true
			break
		}
	}

	if !hasLayerViolation {
		t.Error("Expected layer boundary violation")
	}
}

func TestCreateRules_MaxDepth(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	rules := linter.createRules([]walker.FileInfo{}, nil)

	// Should have created max-depth rule
	hasMaxDepth := false
	for _, rule := range rules {
		if rule.Name() == "max-depth" {
			hasMaxDepth = true
			break
		}
	}

	if !hasMaxDepth {
		t.Error("Expected max-depth rule to be created")
	}
}

func TestCreateRules_NoRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	rules := linter.createRules([]walker.FileInfo{}, nil)

	if len(rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(rules))
	}
}

func TestCreateRules_MultipleRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
				"max-files-in-dir": map[string]interface{}{
					"max": 10,
				},
				"max-subdirs": map[string]interface{}{
					"max": 8,
				},
			},
		},
	}

	rules := linter.createRules([]walker.FileInfo{}, nil)

	// Should have created 3 rules
	if len(rules) < 3 {
		t.Errorf("Expected at least 3 rules, got %d", len(rules))
	}

	// Check that all rules are present
	ruleNames := make(map[string]bool)
	for _, rule := range rules {
		ruleNames[rule.Name()] = true
	}

	expectedRules := []string{"max-depth", "max-files-in-dir", "max-subdirs"}
	for _, expectedRule := range expectedRules {
		if !ruleNames[expectedRule] {
			t.Errorf("Expected rule '%s' to be created", expectedRule)
		}
	}
}
