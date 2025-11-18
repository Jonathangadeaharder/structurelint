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
	TestMax      int      // Optional separate max for test files (0 = skip tests)
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
	multiLangAnalyzer := metrics.NewCognitiveComplexityAnalyzerV2()

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
	multiLangAnalyzer *metrics.AnalyzerV2,
) []Violation {
	fileType := detectFileType(file.Path)

	// Check if file type is supported
	if fileType == FileTypeUnknown {
		return nil
	}

	// Check if file matches any of the patterns (if specified)
	if len(r.FilePatterns) > 0 && !matchesAnyGlob(file.Path, r.FilePatterns) {
		return nil
	}

	// Determine if this is a test file and get appropriate threshold
	isTest := isTestFile(file.Path, fileType)
	maxComplexity := r.Max

	if isTest {
		// If TestMax is 0, skip test files (backward compatible behavior)
		if r.TestMax == 0 {
			return nil
		}
		maxComplexity = r.TestMax
	}

	// Use appropriate analyzer based on file type
	if fileType == FileTypeGo {
		return r.analyzeFileWithMax(file, goAnalyzer, maxComplexity)
	}
	return r.analyzeMultiLangFileWithMax(file, multiLangAnalyzer, maxComplexity)
}

// analyzeFileWithMax analyzes a single Go file for cognitive complexity with a specific max
func (r *MaxCognitiveComplexityRule) analyzeFileWithMax(file walker.FileInfo, analyzer *metrics.CognitiveComplexityAnalyzer, maxComplexity int) []Violation {
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
		if funcMetric.Complexity > maxComplexity {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cognitive complexity %d, exceeds max %d (evidence: CoC correlates with comprehension time, r=0.54)",
					funcMetric.Name, funcMetric.Complexity, maxComplexity),
			})
		}
	}

	return violations
}

// analyzeMultiLangFileWithMax analyzes a Python/JavaScript/TypeScript file for cognitive complexity with a specific max
func (r *MaxCognitiveComplexityRule) analyzeMultiLangFileWithMax(file walker.FileInfo, analyzer *metrics.AnalyzerV2, maxComplexity int) []Violation {
	var violations []Violation

	// Analyze the file using the multi-language analyzer
	fileMetrics, err := analyzer.AnalyzeFileByPath(file.AbsPath)
	if err != nil {
		// Skip files that can't be analyzed
		return violations
	}

	// V2 analyzer returns file-level metrics in FileLevel map
	if complexity, ok := fileMetrics.FileLevel["cognitive_complexity"]; ok {
		if int(complexity) > maxComplexity {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("file has cognitive complexity %d, exceeds max %d (evidence: CoC correlates with comprehension time, r=0.54)",
					int(complexity), maxComplexity),
			})
		}
	}

	// Also check function-level metrics if available
	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Complexity > maxComplexity {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cognitive complexity %d, exceeds max %d (evidence: CoC correlates with comprehension time, r=0.54)",
					funcMetric.Name, funcMetric.Complexity, maxComplexity),
			})
		}
	}

	return violations
}

// NewMaxCognitiveComplexityRule creates a new MaxCognitiveComplexityRule
func NewMaxCognitiveComplexityRule(max int, filePatterns []string) *MaxCognitiveComplexityRule {
	return &MaxCognitiveComplexityRule{
		Max:          max,
		TestMax:      0, // Default: skip test files (backward compatible)
		FilePatterns: filePatterns,
	}
}

// WithTestMax sets a different maximum for test files
func (r *MaxCognitiveComplexityRule) WithTestMax(testMax int) *MaxCognitiveComplexityRule {
	r.TestMax = testMax
	return r
}
