// Max Halstead Effort Rule
//
// Evidence: EEG Study (Scalabrino et al., 2022) shows rs=0.901 correlation
// between Halstead Effort and measured cognitive load, making it the best
// predictor of actual brain activity during code comprehension.
//
// Halstead captures "data complexity" (vocabulary, operators, operands) which
// complements Cognitive Complexity (control-flow complexity).
//
// @structurelint:ignore test-adjacency Covered by max_halstead_effort_test.go
package rules

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/structurelint/structurelint/internal/metrics"
	"github.com/structurelint/structurelint/internal/walker"
)

// MaxHalsteadEffortRule checks that functions don't exceed maximum Halstead effort
type MaxHalsteadEffortRule struct {
	Max          float64
	FilePatterns []string
}

// Name returns the name of the rule
func (r *MaxHalsteadEffortRule) Name() string {
	return "max-halstead-effort"
}

// Check validates that functions don't exceed the maximum Halstead effort
func (r *MaxHalsteadEffortRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Filter files that should be ignored
	files = FilterIgnoredFiles(files, r.Name())

	goAnalyzer := metrics.NewHalsteadAnalyzer()
	multiLangAnalyzer := metrics.NewMultiLanguageHalsteadAnalyzer()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Determine file type and analyzer to use
		var fileViolations []Violation
		if strings.HasSuffix(file.Path, ".go") {
			// Check if file matches any of the patterns (if specified)
			if len(r.FilePatterns) > 0 {
				matched := false
				for _, pattern := range r.FilePatterns {
					if matchesGlobPattern(file.Path, pattern) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			// Skip test files
			if strings.HasSuffix(file.Path, "_test.go") {
				continue
			}

			fileViolations = r.analyzeFile(file, goAnalyzer)
		} else if strings.HasSuffix(file.Path, ".py") ||
			strings.HasSuffix(file.Path, ".js") ||
			strings.HasSuffix(file.Path, ".jsx") ||
			strings.HasSuffix(file.Path, ".ts") ||
			strings.HasSuffix(file.Path, ".tsx") {

			// Check if file matches any of the patterns (if specified)
			if len(r.FilePatterns) > 0 {
				matched := false
				for _, pattern := range r.FilePatterns {
					if matchesGlobPattern(file.Path, pattern) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			// Skip test files
			if strings.Contains(file.Path, "_test.py") ||
				strings.Contains(file.Path, ".test.js") ||
				strings.Contains(file.Path, ".test.ts") ||
				strings.Contains(file.Path, ".spec.js") ||
				strings.Contains(file.Path, ".spec.ts") {
				continue
			}

			fileViolations = r.analyzeMultiLangFile(file, multiLangAnalyzer)
		} else {
			// Unsupported file type
			continue
		}

		violations = append(violations, fileViolations...)
	}

	return violations
}

// analyzeFile analyzes a single Go file for Halstead effort
func (r *MaxHalsteadEffortRule) analyzeFile(file walker.FileInfo, analyzer *metrics.HalsteadAnalyzer) []Violation {
	var violations []Violation

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file.AbsPath, nil, 0)
	if err != nil {
		// Skip files that can't be parsed
		return violations
	}

	// Analyze all functions
	fileMetrics := analyzer.AnalyzeFile(node)

	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Value > r.Max {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has Halstead effort %.0f, exceeds max %.0f (evidence: rs=0.901 correlation with cognitive load)",
					funcMetric.Name, funcMetric.Value, r.Max),
			})
		}
	}

	return violations
}

// analyzeMultiLangFile analyzes a Python/JavaScript/TypeScript file for Halstead effort
func (r *MaxHalsteadEffortRule) analyzeMultiLangFile(file walker.FileInfo, analyzer *metrics.MultiLanguageAnalyzer) []Violation {
	var violations []Violation

	// Analyze the file using the multi-language analyzer
	fileMetrics, err := analyzer.AnalyzeFileByPath(file.AbsPath)
	if err != nil {
		// Skip files that can't be analyzed (e.g., missing interpreter)
		return violations
	}

	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Value > r.Max {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has Halstead effort %.0f, exceeds max %.0f (evidence: rs=0.901 correlation with cognitive load)",
					funcMetric.Name, funcMetric.Value, r.Max),
			})
		}
	}

	return violations
}

// NewMaxHalsteadEffortRule creates a new MaxHalsteadEffortRule
func NewMaxHalsteadEffortRule(max float64, filePatterns []string) *MaxHalsteadEffortRule {
	return &MaxHalsteadEffortRule{
		Max:          max,
		FilePatterns: filePatterns,
	}
}
