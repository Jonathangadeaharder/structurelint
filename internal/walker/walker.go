package walker

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Path       string // Relative path from root
	AbsPath    string // Absolute path
	IsDir      bool
	Depth      int    // Nesting depth from root
	ParentPath string // Path of parent directory
}

// DirInfo represents aggregated information about a directory
type DirInfo struct {
	Path        string
	FileCount   int
	SubdirCount int
	Depth       int
}

// Walker walks a filesystem and collects information
type Walker struct {
	rootPath string
	files    []FileInfo
	dirs     map[string]*DirInfo
}

// New creates a new Walker
func New(rootPath string) *Walker {
	return &Walker{
		rootPath: rootPath,
		files:    []FileInfo{},
		dirs:     make(map[string]*DirInfo),
	}
}

// Walk traverses the filesystem starting from the root path
func (w *Walker) Walk() error {
	absRoot, err := filepath.Abs(w.rootPath)
	if err != nil {
		return err
	}

	return filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		// Skip the root itself
		if relPath == "." {
			return nil
		}

		// Calculate depth
		depth := strings.Count(relPath, string(filepath.Separator))
		if d.IsDir() {
			depth++ // Directories count themselves
		}

		// Get parent path
		parentPath := filepath.Dir(relPath)
		if parentPath == "." {
			parentPath = ""
		}

		// Create FileInfo
		info := FileInfo{
			Path:       relPath,
			AbsPath:    path,
			IsDir:      d.IsDir(),
			Depth:      depth,
			ParentPath: parentPath,
		}

		w.files = append(w.files, info)

		// Update directory statistics
		if d.IsDir() {
			// Initialize this directory
			if _, exists := w.dirs[relPath]; !exists {
				w.dirs[relPath] = &DirInfo{
					Path:  relPath,
					Depth: depth,
				}
			}
		}

		// Update parent directory statistics
		if parentPath != "" || parentPath == "." {
			parent := parentPath
			if parent == "." {
				parent = ""
			}

			if _, exists := w.dirs[parent]; !exists {
				w.dirs[parent] = &DirInfo{
					Path:  parent,
					Depth: depth - 1,
				}
			}

			if d.IsDir() {
				w.dirs[parent].SubdirCount++
			} else {
				w.dirs[parent].FileCount++
			}
		}

		return nil
	})
}

// GetFiles returns all files found during the walk
func (w *Walker) GetFiles() []FileInfo {
	return w.files
}

// GetDirs returns directory statistics
func (w *Walker) GetDirs() map[string]*DirInfo {
	return w.dirs
}

// GetMaxDepth returns the maximum depth found in the filesystem
func (w *Walker) GetMaxDepth() int {
	maxDepth := 0
	for _, info := range w.files {
		if info.Depth > maxDepth {
			maxDepth = info.Depth
		}
	}
	return maxDepth
}

// MatchesPattern checks if a path matches a glob pattern
func MatchesPattern(path, pattern string) bool {
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		// For directory patterns, check if the path starts with the pattern
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}

	// Use filepath.Match for glob patterns
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// For patterns with path separators, try matching the full path
	if strings.Contains(pattern, "/") {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Handle ** for recursive matching
		if strings.Contains(pattern, "**") {
			return matchGlob(path, pattern)
		}
	}

	return false
}

// matchGlob provides more sophisticated glob matching including **
func matchGlob(path, pattern string) bool {
	// Simple implementation of ** matching
	// This is a basic version; production would use a library like doublestar
	parts := strings.Split(pattern, "**")
	if len(parts) == 1 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// For patterns like "src/**/*.ts"
	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// Check prefix
		if prefix != "" && !strings.HasPrefix(path, prefix) {
			return false
		}

		// Check suffix
		if suffix != "" {
			matched, _ := filepath.Match(suffix, filepath.Base(path))
			if !matched {
				return false
			}
		}

		return true
	}

	return false
}
