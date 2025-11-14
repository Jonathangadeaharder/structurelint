// @structurelint:ignore test-adjacency Hash validation is tested through integration tests
package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/structurelint/structurelint/internal/walker"
)

// FileHashRule validates that files match expected hash values
type FileHashRule struct {
	Hashes map[string]string // Map of file path pattern to expected SHA256 hash
}

// NewFileHashRule creates a new FileHashRule
func NewFileHashRule(hashes map[string]string) *FileHashRule {
	return &FileHashRule{
		Hashes: hashes,
	}
}

// Name returns the rule name
func (r *FileHashRule) Name() string {
	return "file-hash"
}

// Check validates files against their expected hashes
func (r *FileHashRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for pattern, expectedHash := range r.Hashes {
		found := false

		for _, file := range files {
			if file.IsDir {
				continue
			}

			// Check if file matches pattern
			matched, err := filepath.Match(pattern, file.Path)
			if err != nil {
				continue
			}

			if !matched && !matchesGlob(file.Path, pattern) {
				continue
			}

			found = true

			// Calculate file hash
			actualHash, err := calculateFileHash(file.AbsPath)
			if err != nil {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("failed to calculate hash: %v", err),
				})
				continue
			}

			// Compare hashes
			if actualHash != expectedHash {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("file hash mismatch: expected %s, got %s", expectedHash, actualHash),
				})
			}
		}

		if !found {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    pattern,
				Message: fmt.Sprintf("no file found matching pattern '%s'", pattern),
			})
		}
	}

	return violations
}

// calculateFileHash computes the SHA256 hash of a file
func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// matchesGlob checks if a path matches a glob pattern (simple ** support)
func matchesGlob(path, pattern string) bool {
	// Simple glob matching - for full implementation, use a glob library
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}
