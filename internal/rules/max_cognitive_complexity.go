// Max Cognitive Complexity Rule
//
// Evidence: Meta-analysis (Schnappinger et al., 2020) shows r=0.54 correlation
// with comprehension time, making it superior to Cyclomatic Complexity for
// measuring code understandability.
//
// @structurelint:ignore test-adjacency Covered by max_cognitive_complexity_test.go
package rules

import (
	"fmt"
	"go/parser"
	"go/token"

	"github.com/structurelint/structurelint/internal/metrics"
	"github.com/structurelint/structurelint/internal/walker"
)

// MaxCognitiveComplexityRule checks that functions don't exceed maximum cognitive complexity
type MaxCognitiveComplexityRule struct {
	Max          int
	FilePatterns []string
}

// Name returns the name of the rule
func (r *MaxCognitiveComplexityRule) Name() string {
	return "max-cognitive-complexity"
}

// Check validates that functions don't exceed the maximum cognitive complexity
func (r *MaxCognitiveComplexityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Filter files that should be ignored
	files = FilterIgnoredFiles(files, r.Name())

	goAnalyzer := metrics.NewCognitiveComplexityAnalyzer()
	multiLangAnalyzer := metrics.NewMultiLanguageCognitiveComplexityAnalyzer()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		fileViolations := r.checkFile(file, goAnalyzer, multiLangAnalyzer)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// checkFile checks a single file for cognitive complexity violations
func (r *MaxCognitiveComplexityRule) checkFile(
	file walker.FileInfo,
	goAnalyzer *metrics.CognitiveComplexityAnalyzer,
	multiLangAnalyzer *metrics.MultiLanguageAnalyzer,
) []Violation {
	fileType := detectFileType(file.Path)

	// Check if file should be analyzed
	if !shouldAnalyzeFile(file.Path, fileType, r.FilePatterns) {
		return nil
	}

	// Use appropriate analyzer based on file type
	if fileType == FileTypeGo {
		return r.analyzeFile(file, goAnalyzer)
	}
	return r.analyzeMultiLangFile(file, multiLangAnalyzer)
}

// analyzeFile analyzes a single Go file for cognitive complexity
func (r *MaxCognitiveComplexityRule) analyzeFile(file walker.FileInfo, analyzer *metrics.CognitiveComplexityAnalyzer) []Violation {
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
		if funcMetric.Complexity > r.Max {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cognitive complexity %d, exceeds max %d (evidence: CoC correlates with comprehension time, r=0.54)",
					funcMetric.Name, funcMetric.Complexity, r.Max),
			})
		}
	}

	return violations
}

// analyzeMultiLangFile analyzes a Python/JavaScript/TypeScript file for cognitive complexity
func (r *MaxCognitiveComplexityRule) analyzeMultiLangFile(file walker.FileInfo, analyzer *metrics.MultiLanguageAnalyzer) []Violation {
	var violations []Violation

	// Analyze the file using the multi-language analyzer
	fileMetrics, err := analyzer.AnalyzeFileByPath(file.AbsPath)
	if err != nil {
		// Skip files that can't be analyzed (e.g., missing interpreter)
		return violations
	}

	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Complexity > r.Max {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cognitive complexity %d, exceeds max %d (evidence: CoC correlates with comprehension time, r=0.54)",
					funcMetric.Name, funcMetric.Complexity, r.Max),
			})
		}
	}

	return violations
}

// NewMaxCognitiveComplexityRule creates a new MaxCognitiveComplexityRule
func NewMaxCognitiveComplexityRule(max int, filePatterns []string) *MaxCognitiveComplexityRule {
	return &MaxCognitiveComplexityRule{
		Max:          max,
		FilePatterns: filePatterns,
	}
}
