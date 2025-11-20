package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/structurelint/structurelint/internal/clones/detector"
	"github.com/structurelint/structurelint/internal/plugin"
)

// runClones handles the 'clones' subcommand
func runClones(args []string) error {
	config, err := parseClonesFlags(args)
	if err != nil {
		return err
	}

	var totalClones int

	if config.runSyntactic {
		count, err := runSyntacticDetection(config)
		if err != nil {
			return err
		}
		totalClones += count
	}

	if config.runSemantic {
		count, err := runSemanticDetection(config)
		if err != nil {
			return err
		}
		totalClones += count
	}

	if totalClones > 0 {
		return fmt.Errorf("found %d total clone(s)", totalClones)
	}

	return nil
}

// clonesConfig holds configuration for clone detection
type clonesConfig struct {
	path                string
	mode                string
	runSyntactic        bool
	runSemantic         bool
	minTokens           int
	minLines            int
	kGramSize           int
	crossFileOnly       bool
	workers             int
	pluginURL           string
	similarityThreshold float64
	format              string
}

// parseClonesFlags parses command-line flags for clone detection
func parseClonesFlags(args []string) (*clonesConfig, error) {
	fs := flag.NewFlagSet("clones", flag.ExitOnError)

	config := &clonesConfig{}

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

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Get path argument
	config.path = "."
	if fs.NArg() > 0 {
		config.path = fs.Arg(0)
	}

	// Set config values
	config.mode = *modeFlag

	// Validate mode flag
	validModes := []string{"syntactic", "semantic", "both"}
	isValidMode := false
	for _, validMode := range validModes {
		if *modeFlag == validMode {
			isValidMode = true
			break
		}
	}
	if !isValidMode {
		return nil, fmt.Errorf("invalid mode '%s': must be one of: syntactic, semantic, both", *modeFlag)
	}

	config.runSyntactic = *modeFlag == "syntactic" || *modeFlag == "both"
	config.runSemantic = *modeFlag == "semantic" || *modeFlag == "both"
	config.minTokens = *minTokens
	config.minLines = *minLines
	config.kGramSize = *kGramSize
	config.crossFileOnly = *crossFileOnly
	config.workers = *workers
	config.pluginURL = *pluginURL
	config.similarityThreshold = *similarityThreshold
	config.format = *formatFlag

	return config, nil
}

// runSyntacticDetection performs syntactic clone detection
func runSyntacticDetection(config *clonesConfig) (int, error) {
	fmt.Printf("🔍 Detecting syntactic clones in %s...\n\n", config.path)

	detectorConfig := detector.Config{
		MinTokens:      config.minTokens,
		MinLines:       config.minLines,
		KGramSize:      config.kGramSize,
		ExcludePattern: []string{"*_test.go", "**/*_gen.go", "**/vendor/**"},
		CrossFileOnly:  config.crossFileOnly,
		NumWorkers:     config.workers,
	}

	d := detector.NewDetector(detectorConfig)
	clones, err := d.DetectClones(config.path)
	if err != nil {
		return 0, fmt.Errorf("syntactic clone detection failed: %w", err)
	}

	if len(clones) > 0 {
		reporter := detector.NewReporter(config.format)
		output := reporter.Report(clones)
		fmt.Print(output)

		if config.format == "console" {
			fmt.Println("\n" + reporter.Summary(clones))
		}
	} else {
		fmt.Println("✓ No syntactic clones found")
	}

	fmt.Println()
	return len(clones), nil
}

// runSemanticDetection performs semantic clone detection via plugin
func runSemanticDetection(config *clonesConfig) (int, error) {
	fmt.Printf("🧠 Detecting semantic clones via plugin at %s...\n", config.pluginURL)

	absPath := resolveAbsolutePath(config.path)
	client := plugin.NewHTTPPluginClient(config.pluginURL)

	if !client.IsAvailable() {
		return handlePluginUnavailable(config.mode)
	}

	return runSemanticDetectionWithPlugin(client, absPath, config)
}

// resolveAbsolutePath resolves the absolute path for semantic detection
func resolveAbsolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't resolve to absolute path, return original
		return path
	}
	return absPath
}

// handlePluginUnavailable handles the case when the plugin is not available
func handlePluginUnavailable(mode string) (int, error) {
	fmt.Fprintf(os.Stderr, "⚠ Warning: Semantic clone detection plugin not available\n")
	fmt.Fprintf(os.Stderr, "  To enable semantic detection:\n")
	fmt.Fprintf(os.Stderr, "    1. cd clone_detection\n")
	fmt.Fprintf(os.Stderr, "    2. pip install -r requirements.txt\n")
	fmt.Fprintf(os.Stderr, "    3. python plugin_server.py\n\n")

	if mode == "semantic" {
		return 0, fmt.Errorf("semantic clone detection plugin required but not available")
	}

	fmt.Println("Continuing with syntactic detection only...")
	return 0, nil
}

// runSemanticDetectionWithPlugin executes semantic detection with the available plugin
func runSemanticDetectionWithPlugin(client *plugin.HTTPPluginClient, absPath string, config *clonesConfig) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	req := &plugin.SemanticCloneRequest{
		SourceDir:           absPath,
		Languages:           []string{"go", "python", "javascript"},
		ExcludePatterns:     []string{"**/*_test.*", "**/vendor/**", "**/node_modules/**"},
		SimilarityThreshold: config.similarityThreshold,
		MaxResults:          100,
	}

	resp, err := client.DetectClones(ctx, req)
	if err != nil {
		return handleSemanticDetectionError(err, config.mode)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "⚠ Warning: %s\n\n", resp.Error)
		return 0, nil
	}

	return reportSemanticClones(resp)
}

// handleSemanticDetectionError handles errors from semantic detection
func handleSemanticDetectionError(err error, mode string) (int, error) {
	fmt.Fprintf(os.Stderr, "⚠ Warning: Semantic detection failed: %v\n\n", err)
	if mode == "semantic" {
		return 0, err
	}
	return 0, nil
}

// reportSemanticClones reports semantic clone detection results
func reportSemanticClones(resp *plugin.SemanticCloneResponse) (int, error) {
	if len(resp.Clones) == 0 {
		fmt.Println("✓ No semantic clones found")
		printSemanticStats(resp)
		return 0, nil
	}

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

	printSemanticStats(resp)
	return len(resp.Clones), nil
}

// printSemanticStats prints semantic detection statistics
func printSemanticStats(resp *plugin.SemanticCloneResponse) {
	fmt.Printf("Analyzed %d files, %d functions in %dms\n",
		resp.Stats.FilesAnalyzed,
		resp.Stats.FunctionsAnalyzed,
		resp.Stats.DurationMs)
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
