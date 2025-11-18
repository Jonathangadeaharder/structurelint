// Package metrics provides multi-language metrics analysis
package metrics

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// MetricCognitiveComplexity is the metric type for cognitive complexity
	MetricCognitiveComplexity = "cognitive-complexity"
	// MetricHalstead is the metric type for Halstead effort
	MetricHalstead = "halstead"
)

// MultiLanguageAnalyzer provides metrics for multiple programming languages
type MultiLanguageAnalyzer struct {
	metricType string // "cognitive-complexity" or "halstead"
}

// NewMultiLanguageCognitiveComplexityAnalyzer creates an analyzer for cognitive complexity
func NewMultiLanguageCognitiveComplexityAnalyzer() *MultiLanguageAnalyzer {
	return &MultiLanguageAnalyzer{
		metricType: MetricCognitiveComplexity,
	}
}

// NewMultiLanguageHalsteadAnalyzer creates an analyzer for Halstead metrics
func NewMultiLanguageHalsteadAnalyzer() *MultiLanguageAnalyzer {
	return &MultiLanguageAnalyzer{
		metricType: MetricHalstead,
	}
}

// Name returns the metric name
func (a *MultiLanguageAnalyzer) Name() string {
	return a.metricType
}

// AnalyzeFileByPath analyzes a file and returns metrics
func (a *MultiLanguageAnalyzer) AnalyzeFileByPath(filePath string) (FileMetrics, error) {
	language := detectLanguage(filePath)

	switch language {
	case "python":
		return a.analyzePythonFile(filePath)
	case "javascript", "typescript":
		return a.analyzeJavaScriptFile(filePath)
	case "java":
		return a.analyzeJavaFile(filePath)
	case "cpp":
		return a.analyzeCppFile(filePath)
	case "csharp":
		return a.analyzeCSharpFile(filePath)
	default:
		return FileMetrics{}, fmt.Errorf("unsupported language: %s", language)
	}
}

// detectLanguage returns the programming language based on file extension
func detectLanguage(filePath string) string {
	extensionToLanguage := map[string]string{
		".py":   "python",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".java": "java",
		".cpp":  "cpp",
		".cc":   "cpp",
		".cxx":  "cpp",
		".h":    "cpp",
		".hpp":  "cpp",
		".cs":   "csharp",
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if lang, ok := extensionToLanguage[ext]; ok {
		return lang
	}
	return "unknown"
}

// analyzePythonFile analyzes a Python file using the Python script
func (a *MultiLanguageAnalyzer) analyzePythonFile(filePath string) (FileMetrics, error) {
	// Get the script path
	scriptPath, err := getScriptPath("python_metrics.py")
	if err != nil {
		return FileMetrics{}, err
	}

	// Execute Python script
	cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try 'python' if 'python3' fails
		cmd = exec.Command("python", scriptPath, a.metricType, filePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return FileMetrics{}, fmt.Errorf("failed to execute Python metrics script: %w\nOutput: %s", err, string(output))
		}
	}

	// Parse JSON output
	var result FileMetrics
	if err := json.Unmarshal(output, &result); err != nil {
		return FileMetrics{}, fmt.Errorf("failed to parse Python metrics output: %w\nOutput: %s", err, string(output))
	}

	result.FilePath = filePath
	return result, nil
}

// analyzeJavaScriptFile analyzes a JavaScript/TypeScript file using the Node.js script
func (a *MultiLanguageAnalyzer) analyzeJavaScriptFile(filePath string) (FileMetrics, error) {
	// Get the script path
	scriptPath, err := getScriptPath("js_metrics.js")
	if err != nil {
		return FileMetrics{}, err
	}

	// Execute Node.js script
	cmd := exec.Command("node", scriptPath, a.metricType, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return FileMetrics{}, fmt.Errorf("failed to execute Node.js metrics script: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result FileMetrics
	if err := json.Unmarshal(output, &result); err != nil {
		return FileMetrics{}, fmt.Errorf("failed to parse JavaScript metrics output: %w\nOutput: %s", err, string(output))
	}

	result.FilePath = filePath
	return result, nil
}

// analyzeJavaFile analyzes a Java file using the Java script
func (a *MultiLanguageAnalyzer) analyzeJavaFile(filePath string) (FileMetrics, error) {
	// Get the script path
	scriptPath, err := getScriptPath("java_metrics.py")
	if err != nil {
		return FileMetrics{}, err
	}

	// Execute Python script
	cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try 'python' if 'python3' fails
		cmd = exec.Command("python", scriptPath, a.metricType, filePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return FileMetrics{}, fmt.Errorf("failed to execute Java metrics script: %w\nOutput: %s", err, string(output))
		}
	}

	// Parse JSON output
	var result FileMetrics
	if err := json.Unmarshal(output, &result); err != nil {
		return FileMetrics{}, fmt.Errorf("failed to parse Java metrics output: %w\nOutput: %s", err, string(output))
	}

	result.FilePath = filePath
	return result, nil
}

// analyzeCppFile analyzes a C++ file using the C++ script
func (a *MultiLanguageAnalyzer) analyzeCppFile(filePath string) (FileMetrics, error) {
	// Get the script path
	scriptPath, err := getScriptPath("cpp_metrics.py")
	if err != nil {
		return FileMetrics{}, err
	}

	// Execute Python script
	cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try 'python' if 'python3' fails
		cmd = exec.Command("python", scriptPath, a.metricType, filePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return FileMetrics{}, fmt.Errorf("failed to execute C++ metrics script: %w\nOutput: %s", err, string(output))
		}
	}

	// Parse JSON output
	var result FileMetrics
	if err := json.Unmarshal(output, &result); err != nil {
		return FileMetrics{}, fmt.Errorf("failed to parse C++ metrics output: %w\nOutput: %s", err, string(output))
	}

	result.FilePath = filePath
	return result, nil
}

// analyzeCSharpFile analyzes a C# file using the C# script
func (a *MultiLanguageAnalyzer) analyzeCSharpFile(filePath string) (FileMetrics, error) {
	// Get the script path
	scriptPath, err := getScriptPath("csharp_metrics.py")
	if err != nil {
		return FileMetrics{}, err
	}

	// Execute Python script
	cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try 'python' if 'python3' fails
		cmd = exec.Command("python", scriptPath, a.metricType, filePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return FileMetrics{}, fmt.Errorf("failed to execute C# metrics script: %w\nOutput: %s", err, string(output))
		}
	}

	// Parse JSON output
	var result FileMetrics
	if err := json.Unmarshal(output, &result); err != nil {
		return FileMetrics{}, fmt.Errorf("failed to parse C# metrics output: %w\nOutput: %s", err, string(output))
	}

	result.FilePath = filePath
	return result, nil
}

// getScriptPath returns the absolute path to a metrics script
func getScriptPath(scriptName string) (string, error) {
	// Get the current file's directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}

	// Scripts are in the same package directory under scripts/
	dir := filepath.Dir(filename)
	scriptPath := filepath.Join(dir, "scripts", scriptName)

	return scriptPath, nil
}
