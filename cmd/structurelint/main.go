package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/export"
	"github.com/structurelint/structurelint/internal/fixer"
	"github.com/structurelint/structurelint/internal/graph"
	init_pkg "github.com/structurelint/structurelint/internal/init"
	"github.com/structurelint/structurelint/internal/linter"
	"github.com/structurelint/structurelint/internal/output"
	"github.com/structurelint/structurelint/internal/rules"
	"github.com/structurelint/structurelint/internal/walker"
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
	// Define flags
	fs := flag.NewFlagSet("structurelint", flag.ContinueOnError)
	formatFlag := fs.String("format", "text", "Output format: text, json, junit")
	versionFlag := fs.Bool("version", false, "Show version information")
	versionFlagShort := fs.Bool("v", false, "Show version information (shorthand)")
	helpFlag := fs.Bool("help", false, "Show help message")
	helpFlagShort := fs.Bool("h", false, "Show help message (shorthand)")
	initFlag := fs.Bool("init", false, "Initialize configuration")
	exportGraphFlag := fs.String("export-graph", "", "Export dependency graph: dot, mermaid, json")
	fixFlag := fs.Bool("fix", false, "Automatically fix violations")
	dryRunFlag := fs.Bool("dry-run", false, "Show what would be fixed without making changes")
	productionFlag := fs.Bool("production", false, "Analyze only production code (exclude test files from graph)")

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

	// Handle export-graph flag
	if *exportGraphFlag != "" {
		return runExportGraph(path, *exportGraphFlag)
	}

	// Handle fix mode
	if *fixFlag || *dryRunFlag {
		return runFix(path, *dryRunFlag)
	}

	// Get output formatter
	formatter, err := output.GetFormatter(*formatFlag, Version)
	if err != nil {
		return err
	}

	// Create and run linter
	l := linter.New().WithProductionMode(*productionFlag)
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

func runFix(path string, dryRun bool) error {
	// Load configuration
	configs, err := config.FindConfigs(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg := config.Merge(configs...)

	// Walk the filesystem
	w := walker.New(path).WithExclude(cfg.Exclude)
	if err := w.Walk(); err != nil {
		return fmt.Errorf("failed to walk filesystem: %w", err)
	}

	files := w.GetFiles()
	dirs := w.GetDirs()

	// Collect all fixes from fixable rules
	var allFixes []rules.Fix

	// For now, we only support fixing naming-convention violations
	// TODO: Add support for more fixable rules
	if namingConfig, ok := cfg.Rules["naming-convention"].(map[string]interface{}); ok {
		patterns := make(map[string]string)
		for k, v := range namingConfig {
			if str, ok := v.(string); ok {
				patterns[k] = str
			}
		}

		if len(patterns) > 0 {
			rule := rules.NewNamingConventionRule(patterns)
			if fixable, ok := interface{}(rule).(rules.FixableRule); ok {
				fixes := fixable.Fix(files, dirs)
				allFixes = append(allFixes, fixes...)
			}
		}
	}

	if len(allFixes) == 0 {
		fmt.Println("No fixable violations found.")
		return nil
	}

	// Apply fixes
	f := fixer.New(dryRun, true)
	return f.Apply(allFixes)
}

func runExportGraph(path string, format string) error {
	// Load configuration
	configs, err := config.FindConfigs(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg := config.Merge(configs...)

	// Walk the filesystem
	w := walker.New(path).WithExclude(cfg.Exclude)
	if err := w.Walk(); err != nil {
		return fmt.Errorf("failed to walk filesystem: %w", err)
	}

	files := w.GetFiles()

	// Build import graph
	builder := graph.NewBuilder(path, cfg.Layers)
	importGraph, err := builder.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build import graph: %w", err)
	}

	// Export graph in requested format
	exporter := export.NewGraphExporter(importGraph)

	var output string
	switch format {
	case "dot":
		output = exporter.ExportDOT()
	case "mermaid":
		output = exporter.ExportMermaid()
	case "json":
		output = exporter.ExportJSON()
	default:
		return fmt.Errorf("unknown export format: %s (supported: dot, mermaid, json)", format)
	}

	fmt.Print(output)
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
  structurelint [options] [path]        Lint the project at path (default: current directory)
  structurelint --init [path]           Generate configuration by analyzing project
  structurelint --export-graph <fmt>    Export dependency graph visualization
  structurelint --version               Show version information
  structurelint --help                  Show this help message

Options:
  -v, --version                Show version information
  -h, --help                   Show help message
      --init                   Initialize configuration for your project
      --format <format>        Output format: text, json, junit (default: text)
      --export-graph <format>  Export dependency graph: dot, mermaid, json
      --fix                    Automatically fix violations when possible
      --dry-run                Show what would be fixed without making changes
      --production             Analyze only production code (exclude test files)

Configuration:
  structurelint looks for .structurelint.yml or .structurelint.yaml files
  in the current directory and parent directories.

Examples:
  structurelint                          Lint current directory
  structurelint --init                   Generate config based on current project
  structurelint --format json .          Output violations as JSON
  structurelint --format junit ./src     Output violations as JUnit XML
  structurelint /path/to/project         Lint specific directory
  structurelint --export-graph dot .     Export dependency graph in DOT format
  structurelint --export-graph mermaid . Export dependency graph in Mermaid format
  structurelint --fix .                  Automatically fix violations
  structurelint --dry-run .              Preview fixes without applying them
  structurelint --production .           Find dead code in production (excludes tests)

Output Formats:
  text    - Human-readable text output (default)
  json    - JSON format for machine parsing and CI/CD integration
  junit   - JUnit XML format for Jenkins, GitHub Actions, etc.

Graph Export Formats:
  dot     - Graphviz DOT format (pipe to: dot -Tpng -o graph.png)
  mermaid - Mermaid diagram format (use in GitHub/GitLab markdown)
  json    - JSON format with nodes and edges

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
