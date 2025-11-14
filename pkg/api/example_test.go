// @structurelint:ignore test-adjacency Example tests demonstrating public API usage
// @structurelint:ignore file-content Example tests use testable examples format, not AAA/GWT pattern
package api_test

import (
	"fmt"
	"testing"

	"github.com/structurelint/structurelint/pkg/api"
)

// Example: Using the basic programmatic API
func ExampleLinter_basic() {
	// Create a new linter
	linter := api.NewLinter()

	// Run linting on current directory
	violations, err := linter.Lint(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print violations
	for _, v := range violations {
		fmt.Printf("[%s] %s: %s\n", v.Rule, v.Path, v.Message)
	}

	if len(violations) == 0 {
		fmt.Println("✓ All checks passed")
	}
}

// Example: Using the linter with custom configuration
func ExampleLinter_withConfig() {
	// Create custom configuration
	cfg := api.NewConfig().
		EnableRule("no-empty-files", true).
		EnableRule("naming-convention", map[string]string{
			"**/*.go": "snake_case",
		}).
		AddExclude("vendor/**").
		AddExclude("node_modules/**").
		AddLayer("domain", "internal/domain/**").
		AddLayer("infrastructure", "internal/infrastructure/**")

	// Create linter with config
	linter := api.NewLinter().WithConfig(cfg)

	// Run linting
	violations, err := linter.Lint("./myproject")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d violation(s)\n", len(violations))
}

// Example: Using the linter in production mode
func ExampleLinter_productionMode() {
	// Create linter with production mode (excludes test files)
	linter := api.NewLinter().WithProductionMode(true)

	violations, err := linter.Lint(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// This will only report unused exports in production code
	for _, v := range violations {
		if v.Rule == "disallow-unused-exports" {
			fmt.Printf("Dead code: %s\n", v.Path)
		}
	}
}

// Example: Using the fluent API for architectural rules
func TestArchRule_layerDependencies(t *testing.T) {
	// Define architectural rule: Domain layer should not depend on Infrastructure
	rule := api.NewArchRule().
		That(api.Layers().Matching("domain")).
		ShouldNot().
		DependOn(api.Layers().Matching("infrastructure"))

	// In a real test, you would load the project graph
	// violations := rule.Check(files)
	// if len(violations) > 0 {
	//     t.Errorf("Architectural rule violated: %v", violations)
	// }

	// This is a placeholder for the example
	_ = rule
}

// Example: Fluent API with file patterns
func TestArchRule_fileNaming(t *testing.T) {
	// Define rule: All Go files should follow snake_case
	rule := api.NewArchRule().
		That(api.Files().Matching("**/*.go")).
		Should().
		HaveNamingConvention("snake_case")

	// In a real test, you would check the rule
	_ = rule
}

// Example: Complex architectural rule
func TestArchRule_complex(t *testing.T) {
	// Test files should not be empty
	rule1 := api.NewArchRule().
		That(api.Files().Matching("**/*_test.go")).
		ShouldNot().
		BeEmpty()

	// Controllers should only depend on services, not directly on repositories
	rule2 := api.NewArchRule().
		That(api.Layers().Matching("controller")).
		ShouldNot().
		DependOn(api.Layers().Matching("repository"))

	_, _ = rule1, rule2
}

// Example: Listing available rules
func ExampleAvailableRules() {
	rules := api.AvailableRules()

	fmt.Println("Available rules:")
	for _, rule := range rules {
		fixable := ""
		if rule.Fixable {
			fixable = " (fixable)"
		}
		fmt.Printf("  - %s: %s%s\n", rule.Name, rule.Description, fixable)
	}
}

// Example: Loading configuration from file
func ExampleLoadConfig() {
	// Load configuration from .structurelint.yml
	cfg, err := api.LoadConfig(".")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Use the loaded configuration
	linter := api.NewLinter().WithConfig(cfg)
	violations, err := linter.Lint(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d violation(s)\n", len(violations))
}

// Example: Checking if a rule is fixable
func ExampleIsFixable() {
	if api.IsFixable("naming-convention") {
		fmt.Println("naming-convention rule supports auto-fix")
	}

	if api.IsFixable("disallow-unused-exports") {
		fmt.Println("disallow-unused-exports rule supports auto-fix")
	}

	if !api.IsFixable("no-empty-files") {
		fmt.Println("no-empty-files rule does not support auto-fix")
	}
}
