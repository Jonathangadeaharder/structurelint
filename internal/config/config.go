package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents a .structurelint.yml configuration file
type Config struct {
	Root        bool                   `yaml:"root"`
	Extends     interface{}            `yaml:"extends"` // string or []string
	Exclude     []string               `yaml:"exclude"` // Patterns to exclude from linting
	Rules       map[string]interface{} `yaml:"rules"`
	Overrides   []Override             `yaml:"overrides"`
	Layers      []Layer                `yaml:"layers"`      // Phase 1: Layer definitions
	Entrypoints []string               `yaml:"entrypoints"` // Phase 2: Entry points for orphan detection
}

// Override represents a configuration override for specific file patterns
type Override struct {
	Files []string               `yaml:"files"`
	Rules map[string]interface{} `yaml:"rules"`
}

// Layer represents an architectural layer definition (Phase 1)
type Layer struct {
	Name      string   `yaml:"name"`
	Path      string   `yaml:"path"`
	DependsOn []string `yaml:"dependsOn"`
}

// MaxDepthRule represents the max-depth rule configuration
type MaxDepthRule struct {
	Max int `yaml:"max"`
}

// MaxFilesRule represents the max-files-in-dir rule configuration
type MaxFilesRule struct {
	Max int `yaml:"max"`
}

// MaxSubdirsRule represents the max-subdirs rule configuration
type MaxSubdirsRule struct {
	Max int `yaml:"max"`
}

// NamingConventionRule represents naming convention patterns
type NamingConventionRule map[string]string

// FileExistenceRule represents file existence requirements
type FileExistenceRule map[string]string

// AllowedLocationsRule represents allowed file locations
type AllowedLocationsRule struct {
	Files       []string `yaml:"files"`
	Destination []string `yaml:"destination"`
	StartsWith  string   `yaml:"startsWith,omitempty"`
}

// DisallowedPatternsRule represents disallowed file patterns
type DisallowedPatternsRule []string

// Load loads a configuration file from the given path
func Load(path string) (*Config, error) {
	visited := make(map[string]bool)
	return loadWithVisited(path, visited)
}

// loadWithVisited loads a config and tracks visited paths to detect cycles
func loadWithVisited(path string, visited map[string]bool) (*Config, error) {
	// Normalize path for cycle detection
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path // fallback to original path if Abs fails
	}

	// Check for cycles
	if visited[absPath] {
		return nil, fmt.Errorf("circular dependency detected: config '%s' is already being loaded", path)
	}
	visited[absPath] = true

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Rules: make(map[string]interface{})}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Rules == nil {
		config.Rules = make(map[string]interface{})
	}

	// Resolve extends if present
	if config.Extends != nil {
		extendedConfigs, err := resolveExtendsWithVisited(config.Extends, filepath.Dir(path), visited)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve extends: %w", err)
		}

		// Merge extended configs with this config
		// Extended configs come first, so this config overrides them
		allConfigs := append(extendedConfigs, &config)
		merged := Merge(allConfigs...)

		// Clear the Extends field to avoid infinite recursion
		merged.Extends = nil

		return merged, nil
	}

	return &config, nil
}

// resolveExtends resolves the extends field to a list of configs (no cycle detection)
func resolveExtends(extends interface{}, baseDir string) ([]*Config, error) {
	visited := make(map[string]bool)
	return resolveExtendsWithVisited(extends, baseDir, visited)
}

// resolveExtendsWithVisited resolves extends with cycle detection
func resolveExtendsWithVisited(extends interface{}, baseDir string, visited map[string]bool) ([]*Config, error) {
	var extendPaths []string

	// Handle both string and []string
	switch v := extends.(type) {
	case string:
		extendPaths = []string{v}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				extendPaths = append(extendPaths, s)
			}
		}
	case []string:
		extendPaths = v
	default:
		return nil, fmt.Errorf("extends must be a string or array of strings, got %T", extends)
	}

	var configs []*Config
	for _, extendPath := range extendPaths {
		resolvedPath, err := resolveExtendPath(extendPath, baseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve extend path '%s': %w", extendPath, err)
		}

		config, err := loadWithVisited(resolvedPath, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to load extended config '%s': %w", resolvedPath, err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// resolveExtendPath resolves an extend path to an absolute file path
func resolveExtendPath(extendPath, baseDir string) (string, error) {
	// If it's an absolute path, use it as-is
	if filepath.IsAbs(extendPath) {
		if _, err := os.Stat(extendPath); err != nil {
			return "", fmt.Errorf("extended config not found: %w", err)
		}
		return extendPath, nil
	}

	// If it starts with ./ or ../, or is not an absolute path, treat as relative
	if strings.HasPrefix(extendPath, "./") ||
		strings.HasPrefix(extendPath, "../") ||
		(!filepath.IsAbs(extendPath) && filepath.VolumeName(extendPath) == "") {
		absPath := filepath.Join(baseDir, extendPath)
		if _, err := os.Stat(absPath); err != nil {
			return "", fmt.Errorf("extended config not found: %w", err)
		}
		return absPath, nil
	}

	// Future: Handle package names (e.g., @structurelint/preset-go)
	// For now, treat as a relative path
	absPath := filepath.Join(baseDir, extendPath)
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("extended config not found (package resolution not yet implemented): %w", err)
	}
	return absPath, nil
}

// FindConfigs finds all .structurelint.yml files from the given path up to the root
func FindConfigs(startPath string) ([]*Config, error) {
	var configs []*Config

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	currentPath := absPath

	for {
		configPath := filepath.Join(currentPath, ".structurelint.yml")
		if _, err := os.Stat(configPath); err == nil {
			config, err := Load(configPath)
			if err != nil {
				return nil, err
			}
			configs = append([]*Config{config}, configs...) // Prepend to maintain order
			if config.Root {
				break
			}
		}

		// Try .yaml extension
		configPath = filepath.Join(currentPath, ".structurelint.yaml")
		if _, err := os.Stat(configPath); err == nil && len(configs) == 0 {
			config, err := Load(configPath)
			if err != nil {
				return nil, err
			}
			configs = append([]*Config{config}, configs...)
			if config.Root {
				break
			}
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	// If no config found, return a default config
	if len(configs) == 0 {
		return []*Config{{Rules: make(map[string]interface{})}}, nil
	}

	return configs, nil
}

// Merge merges multiple configs into a single config
// Later configs override earlier ones
func Merge(configs ...*Config) *Config {
	result := &Config{
		Rules:     make(map[string]interface{}),
		Overrides: []Override{},
	}

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge rules
		for key, value := range config.Rules {
			result.Rules[key] = value
		}

		// Append overrides (they are processed in order)
		result.Overrides = append(result.Overrides, config.Overrides...)

		// Append exclude patterns
		if len(config.Exclude) > 0 {
			result.Exclude = append(result.Exclude, config.Exclude...)
		}

		// Append layers (Phase 1)
		if len(config.Layers) > 0 {
			result.Layers = append(result.Layers, config.Layers...)
		}

		// Append entrypoints (Phase 2)
		if len(config.Entrypoints) > 0 {
			result.Entrypoints = append(result.Entrypoints, config.Entrypoints...)
		}

		// Root flag is taken from the last config that sets it
		if config.Root {
			result.Root = config.Root
		}
	}

	return result
}

// GetRuleConfig extracts a typed rule configuration from the config
func GetRuleConfig[T any](config *Config, ruleName string) (T, bool) {
	var result T
	value, exists := config.Rules[ruleName]
	if !exists {
		return result, false
	}

	// Handle disabled rules (value is 0 or false)
	switch v := value.(type) {
	case int:
		if v == 0 {
			return result, false
		}
	case bool:
		if !v {
			return result, false
		}
	}

	// Try to convert the value to the target type
	// This is a simplified conversion; in production, you'd want more robust handling
	return result, true
}
