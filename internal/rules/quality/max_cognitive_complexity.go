// Max Cognitive Complexity Rule
//
// Evidence: Meta-analysis (Schnappinger et al., 2020) shows r=0.54 correlation
// with comprehension time, making it superior to Cyclomatic Complexity for
// measuring code understandability.
//
// @structurelint:ignore test-adjacency Covered by max_cognitive_complexity_test.go
package quality

import (
	"fmt"
	"go/parser"
	"go/token"

	"github.com/Jonathangadeaharder/structurelint/internal/metrics"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// MaxCognitiveComplexityRule checks that functions don't exceed maximum cognitive complexity
type MaxCognitiveComplexityRule struct {
	Max          int
	TestMax      int      // Optional separate max for test files (0 = skip tests)
	FilePatterns []string // Optional: only check files matching these patterns
}

// Name returns the name of the rule
func (r *MaxCognitiveComplexityRule) Name() string {
	return "max-cognitive-complexity"
}

// Check validates that functions don't exceed the maximum cognitive complexity
func (r *MaxCognitiveComplexityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	files = rules.FilterIgnoredFiles(files, r.Name())

	goAnalyzer := metrics.NewCognitiveComplexityAnalyzer()
	multiLangAnalyzer := metrics.NewCognitiveComplexityAnalyzerV2()

	var violations []rules.Violation
	for _, file := range files {
		if file.IsDir {
			continue
		}
		violations = append(violations, r.checkFile(file, goAnalyzer, multiLangAnalyzer)...)
	}
	return violations
}

func (r *MaxCognitiveComplexityRule) checkFile(
	file walker.FileInfo,
	goAnalyzer *metrics.CognitiveComplexityAnalyzer,
	multiLangAnalyzer *metrics.AnalyzerV2,
) []rules.Violation {
	fileType := rules.DetectFileType(file.Path)

	if fileType == rules.FileTypeUnknown {
		return nil
	}

	// Apply file-patterns filter
	if len(r.FilePatterns) > 0 && !rules.MatchesAnyGlob(file.Path, r.FilePatterns) {
		return nil
	}

	isTest := rules.IsTestFile(file.Path, fileType)
	maxComplexity := r.Max

	if isTest {
		if r.TestMax == 0 {
			return nil // Skip test files
		}
		maxComplexity = r.TestMax
	}

	if fileType == rules.FileTypeGo {
		return r.analyzeGoFile(file, goAnalyzer, maxComplexity)
	}
	return r.analyzeMultiLangFile(file, multiLangAnalyzer, maxComplexity)
}

func (r *MaxCognitiveComplexityRule) analyzeGoFile(
	file walker.FileInfo,
	analyzer *metrics.CognitiveComplexityAnalyzer,
	maxComplexity int,
) []rules.Violation {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file.AbsPath, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	fileMetrics := analyzer.AnalyzeFile(node)
	return r.violationsFromMetrics(file, fileMetrics, maxComplexity)
}

func (r *MaxCognitiveComplexityRule) analyzeMultiLangFile(
	file walker.FileInfo,
	analyzer *metrics.AnalyzerV2,
	maxComplexity int,
) []rules.Violation {
	fileMetrics, err := analyzer.AnalyzeFileByPath(file.AbsPath)
	if err != nil {
		return nil
	}

	if complexity, ok := fileMetrics.FileLevel["cognitive_complexity"]; ok {
		if int(complexity) > maxComplexity {
			return []rules.Violation{{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("file has cognitive complexity %d, exceeds max %d", int(complexity), maxComplexity),
			}}
		}
	}
	return nil
}

func (r *MaxCognitiveComplexityRule) violationsFromMetrics(
	file walker.FileInfo,
	fileMetrics metrics.FileMetrics,
	maxComplexity int,
) []rules.Violation {
	var violations []rules.Violation
	for _, funcMetric := range fileMetrics.Functions {
		if funcMetric.Complexity > maxComplexity {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cognitive complexity %d, exceeds max %d", funcMetric.Name, funcMetric.Complexity, maxComplexity),
			})
		}
	}
	return violations
}

// NewMaxCognitiveComplexityRule creates a new MaxCognitiveComplexityRule
func NewMaxCognitiveComplexityRule(max int, filePatterns []string) *MaxCognitiveComplexityRule {
	return &MaxCognitiveComplexityRule{
		Max:          max,
		TestMax:      0,
		FilePatterns: filePatterns,
	}
}

// WithTestMax sets a different maximum for test files
func (r *MaxCognitiveComplexityRule) WithTestMax(testMax int) *MaxCognitiveComplexityRule {
	r.TestMax = testMax
	return r
}
