package metrics

import (
	"path/filepath"

	"github.com/structurelint/structurelint/internal/parser/treesitter"
)

// AnalyzerV2 provides native tree-sitter based metrics analysis for all languages
type AnalyzerV2 struct {
	metricType string // "cognitive-complexity" or "halstead"
}

// NewCognitiveComplexityAnalyzerV2 creates a native cognitive complexity analyzer
func NewCognitiveComplexityAnalyzerV2() *AnalyzerV2 {
	return &AnalyzerV2{
		metricType: MetricCognitiveComplexity,
	}
}

// NewHalsteadAnalyzerV2 creates a native Halstead metrics analyzer
func NewHalsteadAnalyzerV2() *AnalyzerV2 {
	return &AnalyzerV2{
		metricType: MetricHalstead,
	}
}

// Name returns the metric name
func (a *AnalyzerV2) Name() string {
	return a.metricType
}

// AnalyzeFileByPath analyzes a file and returns metrics using tree-sitter
func (a *AnalyzerV2) AnalyzeFileByPath(filePath string) (FileMetrics, error) {
	ext := filepath.Ext(filePath)

	// Detect language from extension
	lang, err := treesitter.DetectLanguageFromExtension(ext)
	if err != nil {
		// For unsupported languages, return empty metrics
		return FileMetrics{FilePath: filePath}, nil
	}

	// Create metrics calculator for this language
	calculator, err := treesitter.NewMetricsCalculator(lang)
	if err != nil {
		return FileMetrics{}, err
	}

	// Calculate all metrics
	tsMetrics, err := calculator.CalculateFromFile(filePath)
	if err != nil {
		return FileMetrics{}, err
	}

	// Convert to our FileMetrics type
	metrics := FileMetrics{
		FilePath: filePath,
		Functions: []FunctionMetric{},
		FileLevel: make(map[string]float64),
	}

	switch a.metricType {
	case MetricCognitiveComplexity:
		metrics.FileLevel["cognitive_complexity"] = float64(tsMetrics.CognitiveComplexity)
	case MetricHalstead:
		metrics.FileLevel["halstead_effort"] = tsMetrics.HalsteadEffort
		metrics.FileLevel["halstead_volume"] = tsMetrics.HalsteadVolume
		metrics.FileLevel["halstead_difficulty"] = tsMetrics.HalsteadDifficulty
	}

	return metrics, nil
}
