package init

// LanguagePatterns defines file patterns for supported languages
var LanguagePatterns = map[string]struct {
	Source []string
	Test   []string
}{
	"go": {
		Source: []string{"**/*.go"},
		Test:   []string{"_test"},
	},
	"python": {
		Source: []string{"**/*.py"},
		Test:   []string{"test_", "_test"},
	},
	"typescript": {
		Source: []string{"**/*.ts", "**/*.tsx"},
		Test:   []string{".test", ".spec"},
	},
	"javascript": {
		Source: []string{"**/*.js", "**/*.jsx"},
		Test:   []string{".test", ".spec"},
	},
	"java": {
		Source: []string{"src/main/java/**/*.java"},
		Test:   []string{"Test", "IT"},
	},
	"rust": {
		Source: []string{"src/**/*.rs"},
		Test:   []string{"_test"},
	},
	"ruby": {
		Source: []string{"**/*.rb"},
		Test:   []string{"_spec"},
	},
	"cpp": {
		Source: []string{"**/*.cpp", "**/*.cc", "**/*.cxx"},
		Test:   []string{"test_", "_test"},
	},
	"c": {
		Source: []string{"**/*.c"},
		Test:   []string{"test_", "_test"},
	},
	"csharp": {
		Source: []string{"**/*.cs"},
		Test:   []string{"Test", "Tests", ".test"},
	},
}

// ExtensionMap maps file extensions to language names
var ExtensionMap = map[string]string{
	".go":   "go",
	".py":   "python",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".jsx":  "javascript",
	".java": "java",
	".rs":   "rust",
	".rb":   "ruby",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".c":    "c",
	".h":    "c",
	".hpp":  "cpp",
	".cs":   "csharp",
}
