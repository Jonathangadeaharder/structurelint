package init

import (
	"testing"
)

func TestLanguagePatterns(t *testing.T) {
	tests := []string{"go", "python", "typescript", "javascript", "java", "rust", "ruby", "cpp", "c", "csharp"}
	
	for _, lang := range tests {
		if patterns, ok := LanguagePatterns[lang]; !ok {
			t.Errorf("LanguagePatterns missing language: %s", lang)
		} else {
			if len(patterns.Source) == 0 {
				t.Errorf("LanguagePatterns[%s].Source is empty", lang)
			}
			if len(patterns.Test) == 0 {
				t.Errorf("LanguagePatterns[%s].Test is empty", lang)
			}
		}
	}
}

func TestExtensionMap(t *testing.T) {
	tests := map[string]string{
		".go":   "go",
		".py":   "python",
		".ts":   "typescript",
		".js":   "javascript",
		".java": "java",
	}
	
	for ext, expected := range tests {
		if lang, ok := ExtensionMap[ext]; !ok {
			t.Errorf("ExtensionMap missing extension: %s", ext)
		} else if lang != expected {
			t.Errorf("ExtensionMap[%s] = %s, want %s", ext, lang, expected)
		}
	}
}
