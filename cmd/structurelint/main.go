package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	init_pkg "github.com/structurelint/structurelint/internal/init"
	"github.com/structurelint/structurelint/internal/linter"
	"github.com/structurelint/structurelint/internal/output"
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
	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "clones":
			return runClones(os.Args[2:])
		case "help":
			if len(os.Args) > 2 && os.Args[2] == "clones" {
				printClonesHelp()
				return nil
			}
		}
	}

	// Define flags
	fs := flag.NewFlagSet("structurelint", flag.ContinueOnError)
	formatFlag := fs.String("format", "text", "Output format: text, json, junit")
	versionFlag := fs.Bool("version", false, "Show version information")
	versionFlagShort := fs.Bool("v", false, "Show version information (shorthand)")
	helpFlag := fs.Bool("help", false, "Show help message")
	helpFlagShort := fs.Bool("h", false, "Show help message (shorthand)")
	initFlag := fs.Bool("init", false, "Initialize configuration")

	// Parse flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// Extract path argument once
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Handle version flag
	if *versionFlag || *versionFlagShort {
		fmt.Printf("structurelint version %s\n", Version)
		return nil
	}

	// Handle help flag
	if *helpFlag || *helpFlagShort {
		printHelp()
		return nil
	}

	// Handle init flag
	if *initFlag {
		return runInit(path)
	}

	// Get output formatter
	formatter, err := output.GetFormatter(*formatFlag, Version)
	if err != nil {
		return err
	}

	// Create and run linter
	l := linter.New()
	violations, err := l.Lint(path)
	if err != nil {
		return err
	}

	// Format and report violations
	if len(violations) > 0 {
		formatted, err := formatter.Format(violations)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(formatted)

		// For text format, print error message to stderr
		// For JSON/JUnit, just exit with error code (violations are in formatted output)
		if *formatFlag == "text" || *formatFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: found %d violation(s)\n", len(violations))
		}
		os.Exit(1)
	}

	// Only print success message for text format
	if *formatFlag == "text" || *formatFlag == "" {
		fmt.Println("✓ All checks passed")
	} else {
		// For JSON/JUnit, output empty success structure
		formatted, err := formatter.Format(violations)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(formatted)
	}

	return nil
}

func runInit(path string) error {

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
	configPath := filepath.Join(path, ".structurelint.yml")
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
  structurelint [options] [path]   Lint the project at path (default: current directory)
  structurelint clones [options]   Detect code clones (duplicated code)
  structurelint --init [path]      Generate configuration by analyzing project
  structurelint --version          Show version information
  structurelint --help             Show this help message

Commands:
  (default)                    Lint project structure and architecture
  clones                       Detect code clones (see 'structurelint help clones')

Options:
  -v, --version                Show version information
  -h, --help                   Show help message
      --init                   Initialize configuration for your project
      --format <format>        Output format: text, json, junit (default: text)

Configuration:
  structurelint looks for .structurelint.yml or .structurelint.yaml files
  in the current directory and parent directories.

Examples:
  structurelint                     Lint current directory
  structurelint --init              Generate config based on current project
  structurelint --format json .     Output violations as JSON
  structurelint --format junit ./src  Output violations as JUnit XML
  structurelint /path/to/project    Lint specific directory

Output Formats:
  text    - Human-readable text output (default)
  json    - JSON format for machine parsing and CI/CD integration
  junit   - JUnit XML format for Jenkins, GitHub Actions, etc.

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
