// Package init provides auto-configuration functionality for structurelint.
package init

import (
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// LanguageInfo represents detected language information
type LanguageInfo struct {
	Language          string   // "go", "python", "typescript", "javascript", "java", "rust", "ruby", "cpp"
	FileCount         int      // Number of source files
	TestPattern       string   // Detected test pattern: "adjacent", "separate"
	TestDir           string   // Test directory if separate pattern
	TestFilePatterns  []string // Detected test file naming patterns
	SourcePatterns    []string // Source file patterns (e.g., "**/*.go")
	HasIntegrationDir bool     // Has a separate integration test directory
	IntegrationDir    string   // Integration test directory name
}

// ProjectInfo represents detected project information
type ProjectInfo struct {
	Languages          []LanguageInfo
	PrimaryLanguage    *LanguageInfo
	HasMonorepo        bool
	MaxDepth           int
	MaxFilesInDir      int
	MaxSubdirs         int
	DocumentationStyle string // "comprehensive", "minimal", "none"
}

// DetectProject analyzes a project and returns configuration recommendations
func DetectProject(rootPath string) (*ProjectInfo, error) {
	// Walk the filesystem
	w := walker.New(rootPath).WithExclude([]string{
		"node_modules/**",
		"vendor/**",
		".git/**",
		"target/**",    // Rust/Java
		"build/**",     // Various
		"dist/**",      // JavaScript
		"__pycache__/**", // Python
		".pytest_cache/**",
		"coverage/**",
	})

	if err := w.Walk(); err != nil {
		return nil, err
	}

	files := w.GetFiles()
	dirs := w.GetDirs()

	info := &ProjectInfo{
		Languages: make([]LanguageInfo, 0),
	}

	// Detect languages
	languages := detectLanguages(files)
	info.Languages = languages

	// Determine primary language
	if len(languages) > 0 {
		info.PrimaryLanguage = &languages[0]
	}

	// Detect project structure metrics
	info.MaxDepth = calculateMaxDepth(files)
	info.MaxFilesInDir = calculateMaxFilesInDir(dirs)
	info.MaxSubdirs = calculateMaxSubdirs(dirs)

	// Detect documentation style
	info.DocumentationStyle = detectDocumentationStyle(files)

	return info, nil
}

// detectLanguages identifies programming languages and their testing patterns
func detectLanguages(files []walker.FileInfo) []LanguageInfo {
	languageCounts := make(map[string]*LanguageInfo)

	// Count files by language
	for _, file := range files {
		if file.IsDir {
			continue
		}

		ext := filepath.Ext(file.Path)
		lang := extensionToLanguage(ext)
		if lang == "" {
			continue
		}

		info, exists := languageCounts[lang]
		if !exists {
			info = &LanguageInfo{
				Language:         lang,
				TestFilePatterns: make([]string, 0),
				SourcePatterns:   make([]string, 0),
			}
			languageCounts[lang] = info
		}

		info.FileCount++
		analyzeTestFile(file, files, lang, info)
	}

	// Build source patterns and test file patterns
	for lang, info := range languageCounts {
		info.SourcePatterns = getSourcePatterns(lang)
		info.TestFilePatterns = getTestFilePatterns(lang)
	}

	return sortLanguages(languageCounts)
}

func analyzeTestFile(file walker.FileInfo, files []walker.FileInfo, lang string, info *LanguageInfo) {
	if !isTestFile(file.Path, lang) {
		return
	}

	// Analyze test pattern (adjacent vs separate)
	if isAdjacentTest(file.Path, files, lang) {
		info.TestPattern = "adjacent"
	} else if testDir := findTestDirectory(file.Path); testDir != "" {
		info.TestPattern = "separate"
		if info.TestDir == "" {
			info.TestDir = testDir
		}
	}

	// Check for integration test directory
	if isIntegrationTestDir(file.Path) {
		info.HasIntegrationDir = true
		info.IntegrationDir = extractIntegrationDir(file.Path)
	}
}

func sortLanguages(counts map[string]*LanguageInfo) []LanguageInfo {
	result := make([]LanguageInfo, 0, len(counts))
	for _, info := range counts {
		result = append(result, *info)
	}

	// Sort by file count (descending)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].FileCount > result[i].FileCount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// extensionToLanguage maps file extensions to language names
func extensionToLanguage(ext string) string {
	return ExtensionMap[ext]
}

// isTestFile checks if a file is a test file based on naming conventions
func isTestFile(path, lang string) bool {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	patterns := getTestFilePatterns(lang)
	for _, pattern := range patterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

// getTestFilePatterns returns test file naming patterns for a language
func getTestFilePatterns(lang string) []string {
	if patterns, ok := LanguagePatterns[lang]; ok {
		return patterns.Test
	}
	return nil
}

// getSourcePatterns returns source file glob patterns for a language
func getSourcePatterns(lang string) []string {
	if patterns, ok := LanguagePatterns[lang]; ok {
		return patterns.Source
	}
	return nil
}

// isAdjacentTest checks if a test file is adjacent to its source
func isAdjacentTest(testPath string, files []walker.FileInfo, lang string) bool {
	dir := filepath.Dir(testPath)
	base := filepath.Base(testPath)

	// Get the source file name
	sourceName := testToSourceFilename(base, lang)
	if sourceName == "" {
		return false
	}

	// Check if source file exists in same directory
	for _, file := range files {
		if file.IsDir {
			continue
		}
		if filepath.Dir(file.Path) == dir && filepath.Base(file.Path) == sourceName {
			return true
		}
	}

	return false
}

// testToSourceFilename converts test filename to source filename
func testToSourceFilename(testFile, lang string) string {
	ext := filepath.Ext(testFile)
	name := strings.TrimSuffix(testFile, ext)

	// Remove test patterns
	patterns := getTestFilePatterns(lang)
	for _, pattern := range patterns {
		if strings.Contains(name, pattern) {
			name = strings.ReplaceAll(name, pattern, "")
			// Clean up potential double dots
			name = strings.ReplaceAll(name, "..", ".")
			name = strings.TrimPrefix(name, ".")
			return name + ext
		}
	}

	return ""
}

// findTestDirectory extracts test directory from path
func findTestDirectory(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if part == "tests" || part == "test" || part == "__tests__" ||
			part == "spec" || part == "src" && filepath.Dir(path) == "src/test" {
			return part
		}
	}
	return ""
}

// isIntegrationTestDir checks if path contains integration test indicators
func isIntegrationTestDir(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "integration") ||
		strings.Contains(lower, "e2e") ||
		strings.Contains(lower, "functional")
}

// extractIntegrationDir extracts integration test directory name
func extractIntegrationDir(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	for i, part := range parts {
		lower := strings.ToLower(part)
		if strings.Contains(lower, "integration") ||
			strings.Contains(lower, "e2e") ||
			strings.Contains(lower, "functional") {
			// Return path up to and including this directory
			return filepath.Join(parts[:i+1]...)
		}
	}
	return ""
}

// calculateMaxDepth calculates reasonable max depth limit
func calculateMaxDepth(files []walker.FileInfo) int {
	maxDepth := 0
	for _, file := range files {
		if file.Depth > maxDepth {
			maxDepth = file.Depth
		}
	}

	// Add some buffer
	recommended := maxDepth + 2
	if recommended < 4 {
		recommended = 4
	}
	if recommended > 10 {
		recommended = 10
	}

	return recommended
}

// calculateMaxFilesInDir calculates reasonable max files per directory
func calculateMaxFilesInDir(dirs map[string]*walker.DirInfo) int {
	maxFiles := 0
	for _, dir := range dirs {
		if dir.FileCount > maxFiles {
			maxFiles = dir.FileCount
		}
	}

	// Add 20% buffer
	recommended := int(float64(maxFiles) * 1.2)
	if recommended < 20 {
		recommended = 20
	}
	if recommended > 100 {
		recommended = 100
	}

	return recommended
}

// calculateMaxSubdirs calculates reasonable max subdirectories
func calculateMaxSubdirs(dirs map[string]*walker.DirInfo) int {
	maxSubdirs := 0
	for _, dir := range dirs {
		if dir.SubdirCount > maxSubdirs {
			maxSubdirs = dir.SubdirCount
		}
	}

	// Add buffer
	recommended := maxSubdirs + 3
	if recommended < 10 {
		recommended = 10
	}
	if recommended > 30 {
		recommended = 30
	}

	return recommended
}

// detectDocumentationStyle analyzes documentation completeness
func detectDocumentationStyle(files []walker.FileInfo) string {
	readmeCount := 0
	totalDirs := make(map[string]bool)

	for _, file := range files {
		if file.IsDir {
			totalDirs[file.Path] = true
		} else if strings.ToLower(filepath.Base(file.Path)) == "readme.md" {
			readmeCount++
		}
	}

	if readmeCount == 0 {
		return "none"
	}

	// Calculate ratio
	ratio := float64(readmeCount) / float64(len(totalDirs))

	if ratio > 0.5 {
		return "comprehensive"
	}
	return "minimal"
}
