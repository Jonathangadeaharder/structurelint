package main

import (
	"fmt"
	"os"

	init_pkg "github.com/structurelint/structurelint/internal/init"
	"github.com/structurelint/structurelint/internal/linter"
)

// Version is set during build via ldflags
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line arguments
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			fmt.Printf("structurelint version %s\n", Version)
			return nil
		}
		if arg == "--help" || arg == "-h" {
			printHelp()
			return nil
		}
		if arg == "--init" {
			return runInit()
		}
	}

	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	// Create and run linter
	l := linter.New()
	violations, err := l.Lint(path)
	if err != nil {
		return err
	}

	// Report violations
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Printf("%s: %s\n", v.Path, v.Message)
		}
		return fmt.Errorf("found %d violation(s)", len(violations))
	}

	fmt.Println("No violations found")
	return nil
}

func runInit() error {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	fmt.Println("Analyzing project structure...")

	// Detect project
	info, err := init_pkg.DetectProject(path)
	if err != nil {
		return fmt.Errorf("failed to analyze project: %w", err)
	}

	// Print summary
	fmt.Println()
	fmt.Print(init_pkg.GenerateSummary(info))
	fmt.Println()

	// Check if config already exists
	configPath := ".structurelint.yml"
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠ Warning: %s already exists\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response) // Ignore scan errors, default to "no"
		if response != "y" && response != "Y" {
			fmt.Println("Aborted. No changes made.")
			return nil
		}
	}

	// Generate config
	config := init_pkg.GenerateConfig(info)

	// Write config
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("✓ Created %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize .structurelint.yml")
	fmt.Println("  2. Run 'structurelint .' to validate your project")
	fmt.Println("  3. See docs/ for detailed rule documentation")

	return nil
}

func printHelp() {
	fmt.Println(`structurelint - Project structure and architecture linter

Usage:
  structurelint [path]         Lint the project at path (default: current directory)
  structurelint --init [path]  Generate configuration by analyzing project
  structurelint --version      Show version information
  structurelint --help         Show this help message

Options:
  -v, --version               Show version information
  -h, --help                  Show help message
      --init                  Initialize configuration for your project

Configuration:
  structurelint looks for .structurelint.yml or .structurelint.yaml files
  in the current directory and parent directories.

Examples:
  structurelint                Lint current directory
  structurelint --init         Generate config based on current project
  structurelint ./src          Lint src directory
  structurelint /path/to/project

Initialization:
  The --init command analyzes your project to detect:
  - Programming languages and test patterns
  - Project structure and organization
  - Documentation style

  It then generates an appropriate .structurelint.yml configuration
  with smart defaults based on detected patterns.

Documentation:
  https://github.com/structurelint/structurelint`)
}
