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
package quality

import (
	"fmt"
	"go/parser"
	"go/token"

	"github.com/Jonathangadeaharder/structurelint/internal/metrics"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
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
func (r *MaxHalsteadEffortRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Filter files that should be ignored
	files = rules.FilterIgnoredFiles(files, r.Name())

	goAnalyzer := metrics.NewHalsteadAnalyzer()
	multiLangAnalyzer := metrics.NewHalsteadAnalyzerV2()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		fileViolations := r.checkFile(file, goAnalyzer, multiLangAnalyzer)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// checkFile checks a single file for Halstead effort violations
func (r *MaxHalsteadEffortRule) checkFile(
	file walker.FileInfo,
	goAnalyzer *metrics.HalsteadAnalyzer,
	multiLangAnalyzer *metrics.AnalyzerV2,
) []rules.Violation {
	fileType := rules.DetectFileType(file.Path)

	// Check if file should be analyzed
	if !rules.ShouldAnalyzeFile(file.Path, fileType, r.FilePatterns) {
		return nil
	}

	// Use appropriate analyzer based on file type
	if fileType == rules.FileTypeGo {
		return r.analyzeFile(file, goAnalyzer)
	}
	return r.analyzeMultiLangFile(file, multiLangAnalyzer)
}

// analyzeFile analyzes a single Go file for Halstead effort
func (r *MaxHalsteadEffortRule) analyzeFile(file walker.FileInfo, analyzer *metrics.HalsteadAnalyzer) []rules.Violation {
	var violations []rules.Violation

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
			violations = append(violations, rules.Violation{
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
func (r *MaxHalsteadEffortRule) analyzeMultiLangFile(file walker.FileInfo, analyzer *metrics.AnalyzerV2) []rules.Violation {
	var violations []rules.Violation

	// Analyze the file using the multi-language analyzer
	fileMetrics, err := analyzer.AnalyzeFileByPath(file.AbsPath)
	if err != nil {
		// Skip files that can't be analyzed
		return violations
	}

	// V2 analyzer returns file-level metrics in FileLevel map
	if effort, ok := fileMetrics.FileLevel["halstead_effort"]; ok {
		if effort > r.Max {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("file has Halstead effort %.0f, exceeds max %.0f (evidence: rs=0.901 correlation with cognitive load)",
					effort, r.Max),
			})
		}
	}

	// Also check function-level metrics if available
	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Value > r.Max {
			violations = append(violations, rules.Violation{
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
