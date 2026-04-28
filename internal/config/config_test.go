package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
rules:
  max-depth:
    max: 5
  naming-convention:
    "*.ts": "camelCase"
`

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	// Act
	config, err := Load(configFile)

	// Assert
	require.NoError(t, err)
	assert.True(t, config.Root)
	assert.Equal(t, 2, len(config.Rules))
	assert.Contains(t, config.Rules, "max-depth")
}

func TestLoadNonExistentFile(t *testing.T) {
	config, err := Load("/nonexistent/file.yml")

	require.NoError(t, err)
	assert.NotNil(t, config.Rules)
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
			"max-depth": map[string]interface{}{"max": 7},  // Override
			"max-files": map[string]interface{}{"max": 10}, // New rule
		},
	}

	merged := Merge(config1, config2)

	assert.True(t, merged.Root)

	maxDepth := merged.Rules["max-depth"].(map[string]interface{})["max"]
	assert.Equal(t, 7, maxDepth)

	assert.Contains(t, merged.Rules, "max-files")
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

	assert.Equal(t, 2, len(merged.Layers))
	assert.Equal(t, "domain", merged.Layers[0].Name)
	assert.Equal(t, "app", merged.Layers[1].Name)
}

func TestMergeWithEntrypoints(t *testing.T) {
	config1 := &Config{
		Entrypoints: []string{"src/index.ts"},
	}

	config2 := &Config{
		Entrypoints: []string{"src/main.ts"},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 2, len(merged.Entrypoints))
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

	assert.Equal(t, 2, len(merged.Overrides))
}

func TestFindConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "src", "components")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Root config
	rootConfig := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(rootConfig, []byte("root: true\nrules:\n  max-depth:\n    max: 5"), 0644))

	// Sub config
	subConfig := filepath.Join(subDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(subConfig, []byte("rules:\n  max-depth:\n    max: 10"), 0644))

	configs, err := FindConfigs(subDir)

	require.NoError(t, err)

	// Should find both configs (root and sub)
	assert.Equal(t, 2, len(configs))

	// First config should be root
	assert.True(t, configs[0].Root)
}

func TestMergeWithExclude(t *testing.T) {
	config1 := &Config{
		Exclude: []string{"node_modules/**", "dist/**"},
	}

	config2 := &Config{
		Exclude: []string{"vendor/**"},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 3, len(merged.Exclude))

	// Verify all patterns are present
	assert.Contains(t, merged.Exclude, "node_modules/**")
	assert.Contains(t, merged.Exclude, "dist/**")
	assert.Contains(t, merged.Exclude, "vendor/**")
}

func TestLoadWithExclude(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
exclude:
  - testdata/**
  - build/**
rules:
  max-depth:
    max: 5
`

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	config, err := Load(configFile)

	require.NoError(t, err)
	assert.Equal(t, 2, len(config.Exclude))
	assert.Equal(t, "testdata/**", config.Exclude[0])
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	// Use invalid YAML with mismatched indentation and tabs/spaces
	content := "\troot: true\n\t  rules:\n\t\tinvalid:\t  [\n"

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	_, err := Load(configFile)

	assert.Error(t, err)
}

func TestFindConfigsWithYamlExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Use .yaml extension instead of .yml
	configFile := filepath.Join(tmpDir, ".structurelint.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("root: true\nrules:\n  max-depth:\n    max: 5"), 0644))

	configs, err := FindConfigs(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	assert.True(t, configs[0].Root)
}

func TestMergeWithNilConfig(t *testing.T) {
	config1 := &Config{
		Rules: map[string]interface{}{
			"max-depth": 5,
		},
	}

	merged := Merge(config1, nil)

	assert.Equal(t, 1, len(merged.Rules))
}

func TestFindConfigsNoConfigFound(t *testing.T) {
	tmpDir := t.TempDir()

	configs, err := FindConfigs(tmpDir)

	require.NoError(t, err)

	// Should return default config
	assert.Equal(t, 1, len(configs))
	assert.NotNil(t, configs[0].Rules)
}
