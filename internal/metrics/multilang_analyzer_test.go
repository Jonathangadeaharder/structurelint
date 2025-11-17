package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"script.py", "python"},
		{"app.js", "javascript"},
		{"component.jsx", "javascript"},
		{"app.ts", "typescript"},
		{"component.tsx", "typescript"},
		{"main.go", "unknown"},
		{"README.md", "unknown"},
		{"/path/to/script.PY", "python"}, // Case insensitive
		{"/path/to/app.JS", "javascript"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := detectLanguage(tt.filePath)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestNewMultiLanguageCognitiveComplexityAnalyzer(t *testing.T) {
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	if analyzer == nil {
		t.Fatal("NewMultiLanguageCognitiveComplexityAnalyzer() returned nil")
	}

	if analyzer.metricType != MetricCognitiveComplexity {
		t.Errorf("metricType = %q, want %q", analyzer.metricType, MetricCognitiveComplexity)
	}

	if analyzer.Name() != MetricCognitiveComplexity {
		t.Errorf("Name() = %q, want %q", analyzer.Name(), MetricCognitiveComplexity)
	}
}

func TestNewMultiLanguageHalsteadAnalyzer(t *testing.T) {
	analyzer := NewMultiLanguageHalsteadAnalyzer()

	if analyzer == nil {
		t.Fatal("NewMultiLanguageHalsteadAnalyzer() returned nil")
	}

	if analyzer.metricType != MetricHalstead {
		t.Errorf("metricType = %q, want %q", analyzer.metricType, MetricHalstead)
	}

	if analyzer.Name() != MetricHalstead {
		t.Errorf("Name() = %q, want %q", analyzer.Name(), MetricHalstead)
	}
}

func TestGetScriptPath(t *testing.T) {
	scriptPath, err := getScriptPath("python_metrics.py")
	if err != nil {
		t.Fatalf("getScriptPath() error = %v", err)
	}

	// Check that path ends with expected structure
	expectedSuffix := filepath.Join("metrics", "scripts", "python_metrics.py")
	if !filepath.IsAbs(scriptPath) {
		t.Errorf("getScriptPath() returned relative path %q, want absolute path", scriptPath)
	}

	if !containsSuffix(scriptPath, expectedSuffix) {
		t.Errorf("getScriptPath() = %q, want path ending with %q", scriptPath, expectedSuffix)
	}
}

func TestAnalyzeFileByPath_UnsupportedLanguage(t *testing.T) {
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	// Create a temporary file with unsupported extension
	tmpFile, err := os.CreateTemp("", "test*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = analyzer.AnalyzeFileByPath(tmpFile.Name())
	if err == nil {
		t.Error("AnalyzeFileByPath() with .go file should return error, got nil")
	}

	expectedErrMsg := "unsupported language"
	if err != nil && !containsString(err.Error(), expectedErrMsg) {
		t.Errorf("AnalyzeFileByPath() error = %q, want error containing %q", err.Error(), expectedErrMsg)
	}
}

func TestAnalyzePythonFile_ValidPython(t *testing.T) {
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	// Create a temporary Python file
	tmpFile, err := os.CreateTemp("", "test*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write simple Python code
	pythonCode := `def simple_function(x):
    return x + 1

def complex_function(x):
    if x > 0:
        if x > 10:
            return x * 2
    return x
`
	if _, err := tmpFile.WriteString(pythonCode); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	metrics, err := analyzer.AnalyzeFileByPath(tmpFile.Name())

	// If Python is not available, skip the test
	if err != nil && containsString(err.Error(), "failed to execute") {
		t.Skip("Python interpreter not available, skipping test")
	}

	if err != nil {
		t.Fatalf("AnalyzeFileByPath() error = %v", err)
	}

	if metrics.FilePath != tmpFile.Name() {
		t.Errorf("FilePath = %q, want %q", metrics.FilePath, tmpFile.Name())
	}

	if len(metrics.Functions) == 0 {
		t.Error("Expected at least one function in metrics")
	}

	// Verify we got metrics for both functions
	if len(metrics.Functions) != 2 {
		t.Errorf("Got %d functions, want 2", len(metrics.Functions))
	}
}

func TestAnalyzeJavaScriptFile_ValidJS(t *testing.T) {
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	// Create a temporary JavaScript file
	tmpFile, err := os.CreateTemp("", "test*.js")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write simple JavaScript code
	jsCode := `function simpleFunction(x) {
    return x + 1;
}

function complexFunction(x) {
    if (x > 0) {
        if (x > 10) {
            return x * 2;
        }
    }
    return x;
}
`
	if _, err := tmpFile.WriteString(jsCode); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	metrics, err := analyzer.AnalyzeFileByPath(tmpFile.Name())

	// If Node.js is not available or @babel/parser is missing, skip the test
	if err != nil && (containsString(err.Error(), "failed to execute") || containsString(err.Error(), "@babel/parser")) {
		t.Skip("Node.js or @babel/parser not available, skipping test")
	}

	if err != nil {
		t.Fatalf("AnalyzeFileByPath() error = %v", err)
	}

	if metrics.FilePath != tmpFile.Name() {
		t.Errorf("FilePath = %q, want %q", metrics.FilePath, tmpFile.Name())
	}

	if len(metrics.Functions) == 0 {
		t.Error("Expected at least one function in metrics")
	}
}

// Helper functions for tests

func containsSuffix(path, suffix string) bool {
	// Normalize separators for cross-platform compatibility
	normalizedPath := filepath.ToSlash(path)
	normalizedSuffix := filepath.ToSlash(suffix)
	return len(normalizedPath) >= len(normalizedSuffix) &&
		normalizedPath[len(normalizedPath)-len(normalizedSuffix):] == normalizedSuffix
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
