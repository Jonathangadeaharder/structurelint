package main

import (
	"flag"
	"fmt"

	"github.com/structurelint/structurelint/internal/clones/detector"
)

// runClones handles the 'clones' subcommand
func runClones(args []string) error {
	// Create flagset for clones subcommand
	fs := flag.NewFlagSet("clones", flag.ExitOnError)
	minTokens := fs.Int("min-tokens", 20, "Minimum clone size in tokens")
	minLines := fs.Int("min-lines", 3, "Minimum clone size in lines")
	kGramSize := fs.Int("k-gram", 20, "Window size for shingling")
	formatFlag := fs.String("format", "console", "Output format: console, json, sarif")
	crossFileOnly := fs.Bool("cross-file-only", true, "Only report cross-file clones")
	workers := fs.Int("workers", 4, "Number of parallel workers")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get path argument
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Create detector configuration
	config := detector.Config{
		MinTokens:      *minTokens,
		MinLines:       *minLines,
		KGramSize:      *kGramSize,
		ExcludePattern: []string{"*_test.go", "**/*_gen.go", "**/vendor/**"},
		CrossFileOnly:  *crossFileOnly,
		NumWorkers:     *workers,
	}

	// Create detector
	d := detector.NewDetector(config)

	// Run clone detection
	fmt.Printf("🔍 Detecting code clones in %s...\n\n", path)
	clones, err := d.DetectClones(path)
	if err != nil {
		return fmt.Errorf("clone detection failed: %w", err)
	}

	// Create reporter and output results
	reporter := detector.NewReporter(*formatFlag)
	output := reporter.Report(clones)
	fmt.Print(output)

	// Print summary
	if *formatFlag == "console" {
		fmt.Println("\n" + reporter.Summary(clones))
	}

	// Return error if clones found
	if len(clones) > 0 {
		return fmt.Errorf("found %d code clones", len(clones))
	}

	return nil
}

// printClonesHelp prints help for the clones subcommand
func printClonesHelp() {
	fmt.Println(`structurelint clones - Detect code clones (duplicated code)

Usage:
  structurelint clones [options] [path]

Description:
  Detects code clones (duplicated code) using state-of-the-art
  syntactic analysis. Identifies Type-1, Type-2, and Type-3 clones:

  - Type-1: Exact copies (ignoring whitespace/comments)
  - Type-2: Renamed variables/functions
  - Type-3: Minor modifications (added/deleted statements)

Options:
  --min-tokens <n>        Minimum clone size in tokens (default: 20)
  --min-lines <n>         Minimum clone size in lines (default: 3)
  --k-gram <n>            Window size for shingling (default: 20)
  --format <format>       Output format: console, json, sarif (default: console)
  --cross-file-only       Only report cross-file clones (default: true)
  --workers <n>           Number of parallel workers (default: 4)

Examples:
  # Detect clones in current directory
  structurelint clones

  # Detect clones with custom thresholds
  structurelint clones --min-tokens 30 --min-lines 5

  # Output as JSON for CI/CD
  structurelint clones --format json .

  # Include within-file clones
  structurelint clones --cross-file-only=false

Output Formats:
  console  - Human-readable output (default)
  json     - Machine-readable JSON
  sarif    - SARIF format for IDE integration

Detection Algorithm:
  1. Parse and normalize Go source files (AST-based)
  2. Generate k-gram shingles using rolling hash
  3. Build inverted index of hash values
  4. Find hash collisions (potential clones)
  5. Expand matches greedily (forward/backward)
  6. Filter by minimum size and report

Configuration:
  Clone detection settings can be added to .structurelint.yml:

  clone-detection:
    min-tokens: 20
    min-lines: 3
    k-gram-size: 20
    exclude-patterns:
      - "**/*_test.go"
      - "**/*_gen.go"`)
}
