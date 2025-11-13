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
	IsDir      bool   // Whether this is a directory
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
	rootPath        string
	files           []FileInfo
	dirs            map[string]*DirInfo
	excludePatterns []string
}

// New creates a new Walker
func New(rootPath string) *Walker {
	return &Walker{
		rootPath: rootPath,
		files:    []FileInfo{},
		dirs:     make(map[string]*DirInfo),
	}
}

// WithExclude sets exclude patterns for the walker
func (w *Walker) WithExclude(patterns []string) *Walker {
	w.excludePatterns = patterns
	return w
}

// isExcluded checks if a path matches any exclude pattern
func (w *Walker) isExcluded(relPath string) bool {
	for _, pattern := range w.excludePatterns {
		if MatchesPattern(relPath, pattern) {
			return true
		}
	}
	return false
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

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		return w.processPath(relPath, path, d)
	})
}

// processPath handles a single path entry during the walk
func (w *Walker) processPath(relPath, absPath string, d fs.DirEntry) error {
	// Skip the root itself
	if relPath == "." {
		return nil
	}

	// Check for exclusions and skippable directories
	if skipAction := w.shouldSkip(relPath, d); skipAction != nil {
		return skipAction
	}

	depth := w.calculateDepth(relPath, d.IsDir())
	parentPath := w.normalizeParentPath(relPath)

	info := FileInfo{
		Path:       relPath,
		AbsPath:    absPath,
		IsDir:      d.IsDir(),
		Depth:      depth,
		ParentPath: parentPath,
	}

	w.files = append(w.files, info)
	w.updateDirectoryStats(relPath, parentPath, depth, d.IsDir())

	return nil
}

// shouldSkip determines if a path should be skipped and returns the appropriate action
func (w *Walker) shouldSkip(relPath string, d fs.DirEntry) error {
	if w.isExcluded(relPath) {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if d.IsDir() && w.isIgnoredDir(relPath) {
		return filepath.SkipDir
	}

	return nil
}

// isIgnoredDir checks if a directory should be ignored
func (w *Walker) isIgnoredDir(relPath string) bool {
	baseName := filepath.Base(relPath)
	return baseName == ".git" || baseName == "node_modules" || baseName == "vendor"
}

// calculateDepth calculates the depth of a path
func (w *Walker) calculateDepth(relPath string, isDir bool) int {
	depth := strings.Count(relPath, string(filepath.Separator))
	if isDir {
		depth++ // Directories count themselves
	}
	return depth
}

// normalizeParentPath returns the normalized parent path
func (w *Walker) normalizeParentPath(relPath string) string {
	parentPath := filepath.Dir(relPath)
	if parentPath == "." {
		return ""
	}
	return parentPath
}

// updateDirectoryStats updates statistics for both the current directory and its parent
func (w *Walker) updateDirectoryStats(relPath, parentPath string, depth int, isDir bool) {
	if isDir {
		w.ensureDirExists(relPath, depth)
	}

	if parentPath != "" {
		w.ensureDirExists(parentPath, depth-1)
		w.updateParentCounts(parentPath, isDir)
	}
}

// ensureDirExists ensures a directory entry exists in the stats map
func (w *Walker) ensureDirExists(dirPath string, depth int) {
	if _, exists := w.dirs[dirPath]; !exists {
		w.dirs[dirPath] = &DirInfo{
			Path:  dirPath,
			Depth: depth,
		}
	}
}

// updateParentCounts updates file or subdir count for a parent directory
func (w *Walker) updateParentCounts(parentPath string, isDir bool) {
	if isDir {
		w.dirs[parentPath].SubdirCount++
	} else {
		w.dirs[parentPath].FileCount++
	}
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
