// Cognitive Complexity implementation based on:
// - Schnappinger et al. (2020) Meta-Analysis: r=0.54 correlation with comprehension time
// - Superior to Cyclomatic Complexity for measuring understandability
//
// Key differences from Cyclomatic Complexity:
// 1. Penalizes nesting (human cognitive load increases with nesting)
// 2. Ignores shorthand structures (they improve readability)
// 3. Based on human assessment, not mathematical models
//
// @structurelint:ignore test-adjacency Covered by cognitive_complexity_test.go
package metrics

import (
	"fmt"
	"go/ast"
	"go/token"
)

// CognitiveComplexityAnalyzer calculates cognitive complexity
type CognitiveComplexityAnalyzer struct{}

// NewCognitiveComplexityAnalyzer creates a new analyzer
func NewCognitiveComplexityAnalyzer() *CognitiveComplexityAnalyzer {
	return &CognitiveComplexityAnalyzer{}
}

// Name returns the metric name
func (a *CognitiveComplexityAnalyzer) Name() string {
	return "cognitive-complexity"
}

// AnalyzeFunction computes cognitive complexity for a function
func (a *CognitiveComplexityAnalyzer) AnalyzeFunction(funcDecl *ast.FuncDecl) FunctionMetric {
	if funcDecl.Body == nil {
		return FunctionMetric{
			Name:       getFunctionName(funcDecl),
			Value:      0,
			Complexity: 0,
		}
	}

	complexity := calculateCognitiveComplexity(funcDecl.Body, 0)

	return FunctionMetric{
		Name:       getFunctionName(funcDecl),
		StartLine:  int(funcDecl.Pos()),
		EndLine:    int(funcDecl.End()),
		Value:      float64(complexity),
		Complexity: complexity,
	}
}

// AnalyzeFile computes cognitive complexity for all functions in a file
func (a *CognitiveComplexityAnalyzer) AnalyzeFile(node *ast.File) FileMetrics {
	metrics := FileMetrics{
		Functions: []FunctionMetric{},
		FileLevel: make(map[string]float64),
	}

	var totalComplexity int
	var functionCount int

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		funcMetric := a.AnalyzeFunction(funcDecl)
		metrics.Functions = append(metrics.Functions, funcMetric)

		totalComplexity += funcMetric.Complexity
		functionCount++

		return true
	})

	if functionCount > 0 {
		metrics.FileLevel["total"] = float64(totalComplexity)
		metrics.FileLevel["average"] = float64(totalComplexity) / float64(functionCount)
		metrics.FileLevel["max"] = findMaxComplexity(metrics.Functions)
		metrics.FileLevel["function_count"] = float64(functionCount)
	}

	return metrics
}

// calculateCognitiveComplexity computes cognitive complexity using a single-pass algorithm
// It manually tracks nesting level as it traverses the AST
func calculateCognitiveComplexity(body *ast.BlockStmt, initialNesting int) int {
	complexity := 0

	// Use a custom visitor to track nesting level
	var visit func(ast.Node, int)
	visit = func(node ast.Node, nestingLevel int) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.IfStmt:
			// +1 for the if statement + nesting penalty
			complexity += 1 + nestingLevel

			// Visit the if body at increased nesting
			if n.Body != nil {
				for _, stmt := range n.Body.List {
					visit(stmt, nestingLevel+1)
				}
			}

			// Handle else/else-if
			if n.Else != nil {
				// Check if it's an else-if (no extra increment, just continue)
				if elseIf, ok := n.Else.(*ast.IfStmt); ok {
					// else-if doesn't add nesting, treated as continuation
					visit(elseIf, nestingLevel)
				} else if elseBlock, ok := n.Else.(*ast.BlockStmt); ok {
					// Regular else block
					for _, stmt := range elseBlock.List {
						visit(stmt, nestingLevel+1)
					}
				}
			}

		case *ast.ForStmt:
			// +1 for the loop + nesting penalty
			complexity += 1 + nestingLevel

			// Visit the loop body at increased nesting
			if n.Body != nil {
				for _, stmt := range n.Body.List {
					visit(stmt, nestingLevel+1)
				}
			}

		case *ast.RangeStmt:
			// +1 for the loop + nesting penalty
			complexity += 1 + nestingLevel

			// Visit the loop body at increased nesting
			if n.Body != nil {
				for _, stmt := range n.Body.List {
					visit(stmt, nestingLevel+1)
				}
			}

		case *ast.SwitchStmt:
			// +1 for the switch + nesting penalty
			complexity += 1 + nestingLevel

			// Each case adds +1 (but at the switch's nesting level)
			if n.Body != nil {
				for _, stmt := range n.Body.List {
					if caseClause, ok := stmt.(*ast.CaseClause); ok {
						// Don't count default case
						if len(caseClause.List) > 0 {
							// Case adds complexity but no extra nesting penalty
							complexity++
						}
						// Visit case body at increased nesting
						for _, caseStmt := range caseClause.Body {
							visit(caseStmt, nestingLevel+1)
						}
					}
				}
			}

		case *ast.TypeSwitchStmt:
			// Similar to regular switch
			complexity += 1 + nestingLevel

			if n.Body != nil {
				for _, stmt := range n.Body.List {
					if caseClause, ok := stmt.(*ast.CaseClause); ok {
						if len(caseClause.List) > 0 {
							complexity++
						}
						for _, caseStmt := range caseClause.Body {
							visit(caseStmt, nestingLevel+1)
						}
					}
				}
			}

		case *ast.SelectStmt:
			// +1 for select + nesting penalty
			complexity += 1 + nestingLevel

			if n.Body != nil {
				for _, stmt := range n.Body.List {
					if commClause, ok := stmt.(*ast.CommClause); ok {
						// Each comm case adds complexity
						if commClause.Comm != nil {
							complexity++
						}
						for _, caseStmt := range commClause.Body {
							visit(caseStmt, nestingLevel+1)
						}
					}
				}
			}

		case *ast.BranchStmt:
			// break, continue, goto add complexity (flow breaks)
			// return does not (it's the natural exit)
			if n.Tok != token.RETURN {
				complexity += 1 + nestingLevel
			}

		case *ast.BinaryExpr:
			// Cognitive Complexity does NOT penalize logical operators
			// The parent if/for will handle the increment
			// This is a key difference from Cyclomatic Complexity

		case *ast.BlockStmt:
			// Just traverse the block's statements
			for _, stmt := range n.List {
				visit(stmt, nestingLevel)
			}

		case *ast.ExprStmt:
			// Traverse the expression
			visit(n.X, nestingLevel)

		case *ast.AssignStmt:
			// Traverse the RHS expressions
			for _, expr := range n.Rhs {
				visit(expr, nestingLevel)
			}

		case *ast.ReturnStmt:
			// Traverse return expressions
			for _, expr := range n.Results {
				visit(expr, nestingLevel)
			}

		case *ast.DeclStmt:
			// Variable declarations inside functions
			visit(n.Decl, nestingLevel)

		case *ast.DeferStmt:
			// Don't add complexity for defer
			visit(n.Call, nestingLevel)

		case *ast.GoStmt:
			// +1 for goroutine creation (concurrency adds complexity)
			complexity += 1 + nestingLevel
			visit(n.Call, nestingLevel)

		// For other statement types, we could add more cases
		// but the main complexity drivers are covered above
		}
	}

	// Start visiting from the function body
	for _, stmt := range body.List {
		visit(stmt, initialNesting)
	}

	return complexity
}

// getFunctionName extracts the function name including receiver for methods
func getFunctionName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if len(funcDecl.Recv.List[0].Names) > 0 {
			receiverName := funcDecl.Recv.List[0].Names[0].Name
			return fmt.Sprintf("(%s).%s", receiverName, funcDecl.Name.Name)
		}
	}
	return funcDecl.Name.Name
}

// findMaxComplexity finds the maximum complexity among functions
func findMaxComplexity(functions []FunctionMetric) float64 {
	var max float64
	for _, f := range functions {
		if f.Value > max {
			max = f.Value
		}
	}
	return max
}
