package rules

import "testing"

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		path     string
		expected FileType
	}{
		{"main.go", FileTypeGo},
		{"src/app.go", FileTypeGo},
		{"script.py", FileTypePython},
		{"module.py", FileTypePython},
		{"app.js", FileTypeJavaScript},
		{"component.jsx", FileTypeJavaScript},
		{"app.ts", FileTypeTypeScript},
		{"component.tsx", FileTypeTypeScript},
		{"README.md", FileTypeUnknown},
		{"data.json", FileTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectFileType(tt.path)
			if result != tt.expected {
				t.Errorf("detectFileType(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path     string
		fileType FileType
		expected bool
	}{
		// Go test files
		{"main_test.go", FileTypeGo, true},
		{"main.go", FileTypeGo, false},
		{"src/app_test.go", FileTypeGo, true},

		// Python test files
		{"test_module.py", FileTypePython, false}, // Python uses different convention
		{"module_test.py", FileTypePython, true},
		{"module.py", FileTypePython, false},

		// JavaScript test files
		{"app.test.js", FileTypeJavaScript, true},
		{"app.spec.js", FileTypeJavaScript, true},
		{"component.test.jsx", FileTypeJavaScript, true},
		{"component.spec.jsx", FileTypeJavaScript, true},
		{"app.js", FileTypeJavaScript, false},

		// TypeScript test files
		{"app.test.ts", FileTypeTypeScript, true},
		{"app.spec.ts", FileTypeTypeScript, true},
		{"component.test.tsx", FileTypeTypeScript, true},
		{"component.spec.tsx", FileTypeTypeScript, true},
		{"app.ts", FileTypeTypeScript, false},

		// Unknown type
		{"test.md", FileTypeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isTestFile(tt.path, tt.fileType)
			if result != tt.expected {
				t.Errorf("isTestFile(%q, %v) = %v, want %v", tt.path, tt.fileType, result, tt.expected)
			}
		})
	}
}

func TestMatchesAnyGlob(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		expected bool
	}{
		{
			name:     "matches first pattern",
			path:     "src/main.go",
			patterns: []string{"src/**", "lib/**"},
			expected: true,
		},
		{
			name:     "matches second pattern",
			path:     "lib/util.go",
			patterns: []string{"src/**", "lib/**"},
			expected: true,
		},
		{
			name:     "no match",
			path:     "test/main.go",
			patterns: []string{"src/**", "lib/**"},
			expected: false,
		},
		{
			name:     "empty patterns",
			path:     "main.go",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "exact match",
			path:     "main.go",
			patterns: []string{"main.go"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAnyGlob(tt.path, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesAnyGlob(%q, %v) = %v, want %v", tt.path, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestShouldAnalyzeFile(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		fileType     FileType
		filePatterns []string
		expected     bool
	}{
		{
			name:         "Go file, no patterns, not test",
			path:         "main.go",
			fileType:     FileTypeGo,
			filePatterns: nil,
			expected:     true,
		},
		{
			name:         "Go test file",
			path:         "main_test.go",
			fileType:     FileTypeGo,
			filePatterns: nil,
			expected:     false,
		},
		{
			name:         "Python file matches pattern",
			path:         "src/app.py",
			fileType:     FileTypePython,
			filePatterns: []string{"src/**"},
			expected:     true,
		},
		{
			name:         "Python file doesn't match pattern",
			path:         "lib/util.py",
			fileType:     FileTypePython,
			filePatterns: []string{"src/**"},
			expected:     false,
		},
		{
			name:         "JavaScript test file",
			path:         "app.test.js",
			fileType:     FileTypeJavaScript,
			filePatterns: nil,
			expected:     false,
		},
		{
			name:         "TypeScript file",
			path:         "app.ts",
			fileType:     FileTypeTypeScript,
			filePatterns: nil,
			expected:     true,
		},
		{
			name:         "Unknown file type",
			path:         "README.md",
			fileType:     FileTypeUnknown,
			filePatterns: nil,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldAnalyzeFile(tt.path, tt.fileType, tt.filePatterns)
			if result != tt.expected {
				t.Errorf("shouldAnalyzeFile(%q, %v, %v) = %v, want %v",
					tt.path, tt.fileType, tt.filePatterns, result, tt.expected)
			}
		})
	}
}
