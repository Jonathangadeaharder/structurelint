package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/structurelint/structurelint/internal/clones/detector"
	"github.com/structurelint/structurelint/internal/plugin"
)

// runClones handles the 'clones' subcommand
func runClones(args []string) error {
	// Create flagset for clones subcommand
	fs := flag.NewFlagSet("clones", flag.ExitOnError)

	// Clone detection modes
	modeFlag := fs.String("mode", "syntactic", "Detection mode: syntactic, semantic, both")

	// Syntactic detection options
	minTokens := fs.Int("min-tokens", 20, "Minimum clone size in tokens")
	minLines := fs.Int("min-lines", 3, "Minimum clone size in lines")
	kGramSize := fs.Int("k-gram", 20, "Window size for shingling")
	crossFileOnly := fs.Bool("cross-file-only", true, "Only report cross-file clones")
	workers := fs.Int("workers", 4, "Number of parallel workers")

	// Semantic detection options (plugin)
	pluginURL := fs.String("plugin-url", "http://localhost:8765", "URL of semantic clone detection plugin")
	similarityThreshold := fs.Float64("similarity", 0.85, "Similarity threshold for semantic clones (0.0-1.0)")

	// Output options
	formatFlag := fs.String("format", "console", "Output format: console, json, sarif")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get path argument
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Determine which detection modes to run
	runSyntactic := *modeFlag == "syntactic" || *modeFlag == "both"
	runSemantic := *modeFlag == "semantic" || *modeFlag == "both"

	var totalClones int

	// Run syntactic clone detection
	if runSyntactic {
		fmt.Printf("🔍 Detecting syntactic clones in %s...\n\n", path)

		config := detector.Config{
			MinTokens:      *minTokens,
			MinLines:       *minLines,
			KGramSize:      *kGramSize,
			ExcludePattern: []string{"*_test.go", "**/*_gen.go", "**/vendor/**"},
			CrossFileOnly:  *crossFileOnly,
			NumWorkers:     *workers,
		}

		d := detector.NewDetector(config)
		clones, err := d.DetectClones(path)
		if err != nil {
			return fmt.Errorf("syntactic clone detection failed: %w", err)
		}

		if len(clones) > 0 {
			reporter := detector.NewReporter(*formatFlag)
			output := reporter.Report(clones)
			fmt.Print(output)

			if *formatFlag == "console" {
				fmt.Println("\n" + reporter.Summary(clones))
			}

			totalClones += len(clones)
		} else {
			fmt.Println("✓ No syntactic clones found")
		}
		fmt.Println()
	}

	// Run semantic clone detection (via plugin)
	if runSemantic {
		fmt.Printf("🧠 Detecting semantic clones via plugin at %s...\n", *pluginURL)

		// Get absolute path
		absPath, err := os.Getwd()
		if err != nil {
			absPath = path
		}
		if path != "." {
			absPath = path
		}

		// Try to connect to plugin
		client := plugin.NewHTTPPluginClient(*pluginURL)

		if !client.IsAvailable() {
			fmt.Fprintf(os.Stderr, "⚠ Warning: Semantic clone detection plugin not available at %s\n", *pluginURL)
			fmt.Fprintf(os.Stderr, "  To enable semantic detection:\n")
			fmt.Fprintf(os.Stderr, "    1. cd clone_detection\n")
			fmt.Fprintf(os.Stderr, "    2. pip install -r requirements.txt\n")
			fmt.Fprintf(os.Stderr, "    3. python plugin_server.py\n\n")

			if *modeFlag == "semantic" {
				return fmt.Errorf("semantic clone detection plugin required but not available")
			}
			// Graceful degradation for "both" mode
			fmt.Println("Continuing with syntactic detection only...")
		} else {
			// Plugin is available - run semantic detection
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			req := &plugin.SemanticCloneRequest{
				SourceDir:           absPath,
				Languages:           []string{"go", "python", "javascript"},
				ExcludePatterns:     []string{"**/*_test.*", "**/vendor/**", "**/node_modules/**"},
				SimilarityThreshold: *similarityThreshold,
				MaxResults:          100,
			}

			resp, err := client.DetectClones(ctx, req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠ Warning: Semantic detection failed: %v\n\n", err)
				if *modeFlag == "semantic" {
					return err
				}
				// Graceful degradation for "both" mode
			} else if resp.Error != "" {
				fmt.Fprintf(os.Stderr, "⚠ Warning: %s\n\n", resp.Error)
			} else {
				// Report semantic clones
				if len(resp.Clones) > 0 {
					fmt.Printf("\nFound %d semantic clone pairs:\n\n", len(resp.Clones))
					for i, clone := range resp.Clones {
						fmt.Printf("%d. Similarity: %.2f%%\n", i+1, clone.Similarity*100)
						fmt.Printf("   %s:%d-%d\n", clone.SourceFile, clone.SourceStartLine, clone.SourceEndLine)
						fmt.Printf("   %s:%d-%d\n", clone.TargetFile, clone.TargetStartLine, clone.TargetEndLine)
						if clone.Explanation != "" {
							fmt.Printf("   %s\n", clone.Explanation)
						}
						fmt.Println()
					}

					totalClones += len(resp.Clones)
				} else {
					fmt.Println("✓ No semantic clones found")
				}

				// Print stats
				fmt.Printf("Analyzed %d files, %d functions in %dms\n",
					resp.Stats.FilesAnalyzed,
					resp.Stats.FunctionsAnalyzed,
					resp.Stats.DurationMs)
			}
		}
	}

	// Return error if clones found
	if totalClones > 0 {
		return fmt.Errorf("found %d total clone(s)", totalClones)
	}

	return nil
}

// printClonesHelp prints help for the clones subcommand
func printClonesHelp() {
	fmt.Println(`structurelint clones - Detect code clones (duplicated code)

Usage:
  structurelint clones [options] [path]

Description:
  Detects code clones using syntactic analysis (built-in) and optional
  semantic analysis (via plugin). Supports:

  Syntactic Detection (Built-in, Fast):
    - Type-1: Exact copies (ignoring whitespace/comments)
    - Type-2: Renamed variables/functions
    - Type-3: Minor modifications (added/deleted statements)

  Semantic Detection (Plugin, ML-based):
    - Type-4: Semantically similar code with different syntax
    - Uses GraphCodeBERT embeddings + FAISS similarity search

Detection Modes:
  --mode <mode>           Detection mode: syntactic, semantic, both (default: syntactic)

Syntactic Options (Built-in):
  --min-tokens <n>        Minimum clone size in tokens (default: 20)
  --min-lines <n>         Minimum clone size in lines (default: 3)
  --k-gram <n>            Window size for shingling (default: 20)
  --cross-file-only       Only report cross-file clones (default: true)
  --workers <n>           Number of parallel workers (default: 4)

Semantic Options (Plugin):
  --plugin-url <url>      URL of semantic plugin (default: http://localhost:8765)
  --similarity <n>        Similarity threshold 0.0-1.0 (default: 0.85)

Output Options:
  --format <format>       Output format: console, json, sarif (default: console)

Examples:
  # Detect syntactic clones (default, fast)
  structurelint clones

  # Detect semantic clones (requires plugin)
  structurelint clones --mode semantic

  # Detect both syntactic and semantic clones
  structurelint clones --mode both

  # Custom thresholds for syntactic detection
  structurelint clones --min-tokens 30 --min-lines 5

  # Custom similarity for semantic detection
  structurelint clones --mode semantic --similarity 0.90

Output Formats:
  console  - Human-readable output (default)
  json     - Machine-readable JSON
  sarif    - SARIF format for IDE integration

Semantic Plugin Setup (Optional):
  The semantic clone detection plugin provides advanced ML-based
  clone detection but requires Python dependencies:

  1. Install dependencies:
     cd clone_detection
     pip install -r requirements.txt

  2. Start the plugin server:
     python plugin_server.py

  3. Run semantic detection:
     structurelint clones --mode semantic

  The plugin is completely optional. If not available, the tool
  gracefully degrades to syntactic detection only.

Detection Algorithms:
  Syntactic:
    1. Parse and normalize source files (AST-based)
    2. Generate k-gram shingles using rolling hash
    3. Build inverted index of hash values
    4. Find hash collisions (potential clones)
    5. Expand matches greedily (forward/backward)
    6. Filter by minimum size and report

  Semantic (Plugin):
    1. Parse code into functions using tree-sitter
    2. Generate embeddings using GraphCodeBERT
    3. Build FAISS similarity index
    4. Find semantically similar code via cosine similarity
    5. Filter by similarity threshold and report

Configuration:
  Clone detection settings can be added to .structurelint.yml:

  clone-detection:
    mode: syntactic  # or semantic, or both
    min-tokens: 20
    min-lines: 3
    k-gram-size: 20
    similarity-threshold: 0.85
    plugin-url: http://localhost:8765
    exclude-patterns:
      - "**/*_test.go"
      - "**/*_gen.go"`)
}
