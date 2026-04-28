package rules

import "strings"

// FileType represents the type of source code file
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeGo
	FileTypePython
	FileTypeJavaScript
	FileTypeTypeScript
)

// DetectFileType returns the type of a source code file
func DetectFileType(path string) FileType {
	if strings.HasSuffix(path, ".go") {
		return FileTypeGo
	}
	if strings.HasSuffix(path, ".py") {
		return FileTypePython
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") {
		return FileTypeJavaScript
	}
	if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return FileTypeTypeScript
	}
	return FileTypeUnknown
}

// IsTestFile returns true if the file is a test file
func IsTestFile(path string, fileType FileType) bool {
	switch fileType {
	case FileTypeGo:
		return strings.HasSuffix(path, "_test.go")
	case FileTypePython:
		return strings.HasSuffix(path, "_test.py")
	case FileTypeJavaScript:
		return strings.HasSuffix(path, ".test.js") ||
			strings.HasSuffix(path, ".spec.js") ||
			strings.HasSuffix(path, ".test.jsx") ||
			strings.HasSuffix(path, ".spec.jsx")
	case FileTypeTypeScript:
		return strings.HasSuffix(path, ".test.ts") ||
			strings.HasSuffix(path, ".spec.ts") ||
			strings.HasSuffix(path, ".test.tsx") ||
			strings.HasSuffix(path, ".spec.tsx")
	default:
		return false
	}
}

// MatchesAnyGlob returns true if path matches any of the patterns
func MatchesAnyGlob(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchesGlobPattern(path, pattern) {
			return true
		}
	}
	return false
}

// ShouldAnalyzeFile returns true if the file should be analyzed
func ShouldAnalyzeFile(path string, fileType FileType, filePatterns []string) bool {
	// Check file type is supported
	if fileType == FileTypeUnknown {
		return false
	}

	// Check if file matches any of the patterns (if specified)
	if len(filePatterns) > 0 && !MatchesAnyGlob(path, filePatterns) {
		return false
	}

	// Skip test files
	return !IsTestFile(path, fileType)
}
