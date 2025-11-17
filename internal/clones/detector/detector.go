package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/structurelint/structurelint/internal/clones/parser"
	"github.com/structurelint/structurelint/internal/clones/syntactic"
	"github.com/structurelint/structurelint/internal/clones/types"
)

// Detector orchestrates the clone detection process
type Detector struct {
	normalizer *parser.Normalizer
	hasher     *syntactic.Hasher
	index      *syntactic.Index
	expander   *syntactic.Expander
	config     Config
}

// Config holds configuration for clone detection
type Config struct {
	MinTokens      int      // Minimum clone size in tokens
	MinLines       int      // Minimum clone size in lines
	KGramSize      int      // Window size for shingling
	ExcludePattern []string // Patterns to exclude
	CrossFileOnly  bool     // Only report cross-file clones
	NumWorkers     int      // Number of parallel workers
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MinTokens:      20,
		MinLines:       3,
		KGramSize:      20,
		ExcludePattern: []string{"*_test.go", "**/*_gen.go", "**/vendor/**"},
		CrossFileOnly:  true,
		NumWorkers:     4,
	}
}

// NewDetector creates a new clone detector
func NewDetector(config Config) *Detector {
	return &Detector{
		normalizer: parser.NewNormalizer(),
		hasher:     syntactic.NewHasher(config.KGramSize),
		index:      syntactic.NewIndex(),
		expander:   syntactic.NewExpander(),
		config:     config,
	}
}

// DetectClones finds all clones in the given directory
func (d *Detector) DetectClones(rootPath string) ([]*types.Clone, error) {
	// Step 1: Find all Go files
	files, err := d.findGoFiles(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found in %s", rootPath)
	}

	fmt.Printf("Found %d Go files\n", len(files))

	// Step 2: Normalize all files in parallel
	tokenCache, err := d.normalizeFiles(files)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize files: %w", err)
	}

	fmt.Printf("Normalized %d files\n", len(tokenCache))

	// Step 3: Generate shingles and build index
	err = d.buildIndex(tokenCache)
	if err != nil {
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	stats := d.index.Stats()
	fmt.Printf("Index: %d unique hashes, %d total shingles, %d collisions\n",
		stats.TotalHashes, stats.TotalShingles, stats.CollisionCount)

	// Step 4: Find hash collisions
	var collisions map[uint64][]types.Shingle
	if d.config.CrossFileOnly {
		collisions = d.index.FindCrossFileCollisions()
	} else {
		collisions = d.index.FindCollisions()
	}

	fmt.Printf("Found %d hash collisions\n", len(collisions))

	// Step 5: Expand collisions into full clones
	d.expander.SetTokenCache(tokenCache)
	clones := d.expander.ExpandAllCollisions(collisions)

	fmt.Printf("Expanded to %d clone pairs\n", len(clones))

	// Step 6: Filter clones by minimum size
	filteredClones := d.filterClones(clones)

	fmt.Printf("After filtering: %d clones (min %d tokens, %d lines)\n",
		len(filteredClones), d.config.MinTokens, d.config.MinLines)

	return filteredClones, nil
}

// findGoFiles recursively finds all .go files in the directory
func (d *Detector) findGoFiles(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Check exclude patterns
		for _, pattern := range d.config.ExcludePattern {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// normalizeFiles processes all files in parallel
func (d *Detector) normalizeFiles(files []string) (map[string][]types.Token, error) {
	tokenCache := make(map[string][]types.Token)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create worker pool
	jobs := make(chan string, len(files))
	errors := make(chan error, len(files))

	// Start workers
	for w := 0; w < d.config.NumWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				fileTokens, err := d.normalizer.NormalizeFile(filePath)
				if err != nil {
					errors <- fmt.Errorf("failed to normalize %s: %w", filePath, err)
					continue
				}

				mu.Lock()
				tokenCache[filePath] = fileTokens.Tokens
				mu.Unlock()
			}
		}()
	}

	// Submit jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return tokenCache, nil
}

// buildIndex generates shingles and adds them to the index
func (d *Detector) buildIndex(tokenCache map[string][]types.Token) error {
	for filePath, tokens := range tokenCache {
		fileTokens := &types.FileTokens{
			FilePath: filePath,
			Tokens:   tokens,
		}

		shingles := d.hasher.GenerateShingles(fileTokens)
		d.index.AddBatch(shingles)
	}

	return nil
}

// filterClones removes clones that don't meet minimum size requirements
func (d *Detector) filterClones(clones []*types.Clone) []*types.Clone {
	var filtered []*types.Clone

	for _, clone := range clones {
		if clone.TokenCount >= d.config.MinTokens && clone.LineCount >= d.config.MinLines {
			filtered = append(filtered, clone)
		}
	}

	return filtered
}

// GetIndexStats returns statistics about the index
func (d *Detector) GetIndexStats() syntactic.IndexStats {
	return d.index.Stats()
}
