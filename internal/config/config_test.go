package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
rules:
  max-depth:
    max: 5
  naming-convention:
    "*.ts": "camelCase"
`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := Load(configFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !config.Root {
		t.Error("Expected Root to be true")
	}

	if len(config.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(config.Rules))
	}

	if _, ok := config.Rules["max-depth"]; !ok {
		t.Error("Expected max-depth rule to exist")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	config, err := Load("/nonexistent/file.yml")

	if err != nil {
		t.Fatalf("Expected no error for missing file, got %v", err)
	}

	if config.Rules == nil {
		t.Error("Expected Rules map to be initialized")
	}
}

func TestMerge(t *testing.T) {
	config1 := &Config{
		Root: false,
		Rules: map[string]interface{}{
			"max-depth": map[string]interface{}{"max": 5},
		},
	}

	config2 := &Config{
		Root: true,
		Rules: map[string]interface{}{
			"max-depth":  map[string]interface{}{"max": 7}, // Override
			"max-files": map[string]interface{}{"max": 10}, // New rule
		},
	}

	merged := Merge(config1, config2)

	if !merged.Root {
		t.Error("Expected Root to be true from config2")
	}

	maxDepth := merged.Rules["max-depth"].(map[string]interface{})["max"]
	if maxDepth != 7 {
		t.Errorf("Expected max-depth to be 7, got %v", maxDepth)
	}

	if _, ok := merged.Rules["max-files"]; !ok {
		t.Error("Expected max-files rule from config2")
	}
}

func TestMergeWithLayers(t *testing.T) {
	config1 := &Config{
		Layers: []Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	config2 := &Config{
		Layers: []Layer{
			{Name: "app", Path: "src/app/**", DependsOn: []string{"domain"}},
		},
	}

	merged := Merge(config1, config2)

	if len(merged.Layers) != 2 {
		t.Errorf("Expected 2 layers, got %d", len(merged.Layers))
	}

	if merged.Layers[0].Name != "domain" {
		t.Errorf("Expected first layer to be 'domain', got %s", merged.Layers[0].Name)
	}

	if merged.Layers[1].Name != "app" {
		t.Errorf("Expected second layer to be 'app', got %s", merged.Layers[1].Name)
	}
}

func TestMergeWithEntrypoints(t *testing.T) {
	config1 := &Config{
		Entrypoints: []string{"src/index.ts"},
	}

	config2 := &Config{
		Entrypoints: []string{"src/main.ts"},
	}

	merged := Merge(config1, config2)

	if len(merged.Entrypoints) != 2 {
		t.Errorf("Expected 2 entrypoints, got %d", len(merged.Entrypoints))
	}
}

func TestMergeWithOverrides(t *testing.T) {
	config1 := &Config{
		Overrides: []Override{
			{
				Files: []string{"src/**"},
				Rules: map[string]interface{}{"max-depth": 5},
			},
		},
	}

	config2 := &Config{
		Overrides: []Override{
			{
				Files: []string{"tests/**"},
				Rules: map[string]interface{}{"max-depth": 0},
			},
		},
	}

	merged := Merge(config1, config2)

	if len(merged.Overrides) != 2 {
		t.Errorf("Expected 2 overrides, got %d", len(merged.Overrides))
	}
}

func TestFindConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "src", "components")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Root config
	rootConfig := filepath.Join(tmpDir, ".structurelint.yml")
	if err := os.WriteFile(rootConfig, []byte("root: true\nrules:\n  max-depth:\n    max: 5"), 0644); err != nil {
		t.Fatal(err)
	}

	// Sub config
	subConfig := filepath.Join(subDir, ".structurelint.yml")
	if err := os.WriteFile(subConfig, []byte("rules:\n  max-depth:\n    max: 10"), 0644); err != nil {
		t.Fatal(err)
	}

	configs, err := FindConfigs(subDir)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find both configs (root and sub)
	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// First config should be root
	if !configs[0].Root {
		t.Error("Expected first config to be root")
	}
}
