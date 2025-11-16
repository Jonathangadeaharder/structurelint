// Package metrics provides evidence-based software quality metrics
// based on systematic literature reviews and empirical studies.
//
// @structurelint:no-test Interface definitions only, tested through implementations
package metrics

import (
	"go/ast"
)

// MetricResult represents the output of a metric analysis
type MetricResult struct {
	File    string                 // File path
	Metric  string                 // Metric name
	Value   float64                // Primary metric value
	Details map[string]interface{} // Additional details
}

// FunctionMetric represents metrics for a single function
type FunctionMetric struct {
	Name       string  // Function name (including receiver for methods)
	StartLine  int     // Starting line number
	EndLine    int     // Ending line number
	Value      float64 // Metric value
	Complexity int     // For complexity metrics
}

// FileMetrics represents all metrics for a file
type FileMetrics struct {
	FilePath  string
	Functions []FunctionMetric
	FileLevel map[string]float64 // File-level aggregations
}

// Analyzer computes metrics for code
type Analyzer interface {
	// Name returns the metric name
	Name() string

	// AnalyzeFunction computes the metric for a function
	AnalyzeFunction(funcDecl *ast.FuncDecl) FunctionMetric

	// AnalyzeFile computes the metric for an entire file
	AnalyzeFile(node *ast.File) FileMetrics
}
