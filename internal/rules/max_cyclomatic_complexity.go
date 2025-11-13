// @structurelint:ignore test-adjacency AST-based complexity analysis is indirectly tested through integration tests
package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// MaxCyclomaticComplexityRule checks that functions don't exceed a maximum cyclomatic complexity
type MaxCyclomaticComplexityRule struct {
	Max          int
	FilePatterns []string
}

// Name returns the name of the rule
func (r *MaxCyclomaticComplexityRule) Name() string {
	return "max-cyclomatic-complexity"
}

// Check validates that functions don't exceed the maximum cyclomatic complexity
func (r *MaxCyclomaticComplexityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Filter files that should be ignored
	files = FilterIgnoredFiles(files, r.Name())

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Only check Go files
		if !strings.HasSuffix(file.Path, ".go") {
			continue
		}

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

		// Parse and analyze the file
		fileViolations := r.analyzeFile(file)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// analyzeFile analyzes a single Go file for cyclomatic complexity
func (r *MaxCyclomaticComplexityRule) analyzeFile(file walker.FileInfo) []Violation {
	var violations []Violation

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file.AbsPath, nil, 0)
	if err != nil {
		// Skip files that can't be parsed (might not be valid Go)
		return violations
	}

	// Walk the AST and calculate complexity for each function
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Calculate cyclomatic complexity
		complexity := calculateCyclomaticComplexity(funcDecl)

		if complexity > r.Max {
			funcName := funcDecl.Name.Name

			// Include receiver name for methods
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if len(funcDecl.Recv.List[0].Names) > 0 {
					receiverName := funcDecl.Recv.List[0].Names[0].Name
					funcName = fmt.Sprintf("(%s).%s", receiverName, funcName)
				}
			}

			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("function '%s' has cyclomatic complexity %d, exceeds max %d", funcName, complexity, r.Max),
			})
		}

		return true
	})

	return violations
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function
// Cyclomatic complexity = 1 + number of decision points
// Decision points: if, for, switch case, select case, &&, ||
func calculateCyclomaticComplexity(funcDecl *ast.FuncDecl) int {
	if funcDecl.Body == nil {
		return 0
	}

	complexity := 1 // Base complexity

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++

		case *ast.ForStmt, *ast.RangeStmt:
			complexity++

		case *ast.CaseClause:
			// Don't count default case
			if len(node.List) > 0 {
				complexity++
			}

		case *ast.CommClause:
			// Select case
			if node.Comm != nil {
				complexity++
			}

		case *ast.BinaryExpr:
			// Count logical operators
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}

		return true
	})

	return complexity
}

// NewMaxCyclomaticComplexityRule creates a new MaxCyclomaticComplexityRule
func NewMaxCyclomaticComplexityRule(max int, filePatterns []string) *MaxCyclomaticComplexityRule {
	return &MaxCyclomaticComplexityRule{
		Max:          max,
		FilePatterns: filePatterns,
	}
}
