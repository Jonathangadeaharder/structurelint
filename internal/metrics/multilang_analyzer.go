// Package metrics provides multi-language metrics analysis
package metrics

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/parser/treesitter"
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
	// Detect language from file extension
	lang, err := a.detectLanguageFromPath(filePath)
	if err != nil {
		return FileMetrics{}, err
	}

	// Create tree-sitter language constant
	tsLang, err := a.convertToTreeSitterLanguage(lang)
	if err != nil {
		return FileMetrics{}, err
	}

	// Create metrics calculator
	calculator, err := treesitter.NewMetricsCalculator(tsLang)
	if err != nil {
		return FileMetrics{}, fmt.Errorf("failed to create metrics calculator: %w", err)
	}

	// Calculate metrics using tree-sitter
	tsMetrics, err := calculator.CalculateFromFile(filePath)
	if err != nil {
		return FileMetrics{}, fmt.Errorf("failed to calculate metrics for %s: %w", filePath, err)
	}

	// Convert to FileMetrics format
	result := FileMetrics{
		FilePath:  filePath,
		Functions: []FunctionMetric{}, // Multi-language analysis provides file-level metrics only
		FileLevel: make(map[string]float64),
	}

	// Store metrics based on analyzer type
	switch a.metricType {
	case MetricCognitiveComplexity:
		result.FileLevel["cognitive_complexity"] = float64(tsMetrics.CognitiveComplexity)
	case MetricHalstead:
		result.FileLevel["halstead_effort"] = tsMetrics.HalsteadEffort
		result.FileLevel["halstead_volume"] = tsMetrics.HalsteadVolume
		result.FileLevel["halstead_difficulty"] = tsMetrics.HalsteadDifficulty
	}

	return result, nil
}

// detectLanguageFromPath returns the programming language based on file path
func (a *MultiLanguageAnalyzer) detectLanguageFromPath(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	extensionToLanguage := map[string]string{
		".py":   "python",
		".js":   "javascript",
		".jsx":  "javascript",
		".mjs":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".java": "java",
		".cpp":  "cpp",
		".cc":   "cpp",
		".cxx":  "cpp",
		".c":    "c",
		".h":    "cpp",
		".hpp":  "cpp",
		".cs":   "csharp",
	}

	if lang, ok := extensionToLanguage[ext]; ok {
		return lang, nil
	}

	return "", fmt.Errorf("unsupported file extension: %s", ext)
}

// convertToTreeSitterLanguage converts language string to tree-sitter Language type
func (a *MultiLanguageAnalyzer) convertToTreeSitterLanguage(lang string) (treesitter.Language, error) {
	languageMap := map[string]treesitter.Language{
		"python":     treesitter.LanguagePython,
		"javascript": treesitter.LanguageJavaScript,
		"typescript": treesitter.LanguageTypeScript,
		"java":       treesitter.LanguageJava,
		"cpp":        treesitter.LanguageCpp,
		"c":          treesitter.LanguageCpp, // Use C++ parser for C files
		"csharp":     treesitter.LanguageCSharp,
	}

	if tsLang, ok := languageMap[lang]; ok {
		return tsLang, nil
	}

	return "", fmt.Errorf("unsupported language: %s", lang)
}
