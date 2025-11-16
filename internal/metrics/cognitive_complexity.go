// Cognitive Complexity implementation based on:
// - Schnappinger et al. (2020) Meta-Analysis: r=0.54 correlation with comprehension time
// - Superior to Cyclomatic Complexity for measuring understandability
//
// Key differences from Cyclomatic Complexity:
// 1. Penalizes nesting (human cognitive load increases with nesting)
// 2. Ignores shorthand structures (they improve readability)
// 3. Based on human assessment, not mathematical models
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
	visitor := &cognitiveVisitor{complexity: 0}
	for _, stmt := range body.List {
		visitor.visit(stmt, initialNesting)
	}
	return visitor.complexity
}

// cognitiveVisitor handles cognitive complexity calculation
type cognitiveVisitor struct {
	complexity int
}

func (v *cognitiveVisitor) visit(node ast.Node, nestingLevel int) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.IfStmt:
		v.handleIfStmt(n, nestingLevel)
	case *ast.ForStmt:
		v.handleForStmt(n, nestingLevel)
	case *ast.RangeStmt:
		v.handleRangeStmt(n, nestingLevel)
	case *ast.SwitchStmt:
		v.handleSwitchStmt(n, nestingLevel)
	case *ast.TypeSwitchStmt:
		v.handleTypeSwitchStmt(n, nestingLevel)
	case *ast.SelectStmt:
		v.handleSelectStmt(n, nestingLevel)
	case *ast.BranchStmt:
		v.handleBranchStmt(n, nestingLevel)
	case *ast.BlockStmt:
		v.visitBlockStmtList(n.List, nestingLevel)
	case *ast.ExprStmt:
		v.visit(n.X, nestingLevel)
	case *ast.AssignStmt:
		v.visitExprList(n.Rhs, nestingLevel)
	case *ast.ReturnStmt:
		v.visitExprList(n.Results, nestingLevel)
	case *ast.DeclStmt:
		v.visit(n.Decl, nestingLevel)
	case *ast.DeferStmt:
		v.visit(n.Call, nestingLevel)
	case *ast.GoStmt:
		v.handleGoStmt(n, nestingLevel)
	}
}

func (v *cognitiveVisitor) handleIfStmt(n *ast.IfStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.visitBody(n.Body, nestingLevel+1)
	v.visitElse(n.Else, nestingLevel)
}

func (v *cognitiveVisitor) handleForStmt(n *ast.ForStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.visitBody(n.Body, nestingLevel+1)
}

func (v *cognitiveVisitor) handleRangeStmt(n *ast.RangeStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.visitBody(n.Body, nestingLevel+1)
}

func (v *cognitiveVisitor) handleSwitchStmt(n *ast.SwitchStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.processSwitchCases(n.Body, nestingLevel)
}

func (v *cognitiveVisitor) handleTypeSwitchStmt(n *ast.TypeSwitchStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.processSwitchCases(n.Body, nestingLevel)
}

func (v *cognitiveVisitor) handleSelectStmt(n *ast.SelectStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.processSelectCases(n.Body, nestingLevel)
}

func (v *cognitiveVisitor) handleBranchStmt(n *ast.BranchStmt, nestingLevel int) {
	if n.Tok != token.RETURN {
		v.complexity += 1 + nestingLevel
	}
}

func (v *cognitiveVisitor) handleGoStmt(n *ast.GoStmt, nestingLevel int) {
	v.complexity += 1 + nestingLevel
	v.visit(n.Call, nestingLevel)
}

func (v *cognitiveVisitor) visitBody(body *ast.BlockStmt, nestingLevel int) {
	if body != nil {
		v.visitBlockStmtList(body.List, nestingLevel)
	}
}

func (v *cognitiveVisitor) visitElse(elseStmt ast.Stmt, nestingLevel int) {
	if elseStmt == nil {
		return
	}

	if elseIf, ok := elseStmt.(*ast.IfStmt); ok {
		v.visit(elseIf, nestingLevel)
	} else if elseBlock, ok := elseStmt.(*ast.BlockStmt); ok {
		v.visitBlockStmtList(elseBlock.List, nestingLevel+1)
	}
}

func (v *cognitiveVisitor) visitBlockStmtList(stmts []ast.Stmt, nestingLevel int) {
	for _, stmt := range stmts {
		v.visit(stmt, nestingLevel)
	}
}

func (v *cognitiveVisitor) visitExprList(exprs []ast.Expr, nestingLevel int) {
	for _, expr := range exprs {
		v.visit(expr, nestingLevel)
	}
}

func (v *cognitiveVisitor) processSwitchCases(body *ast.BlockStmt, nestingLevel int) {
	if body == nil {
		return
	}

	for _, stmt := range body.List {
		if caseClause, ok := stmt.(*ast.CaseClause); ok {
			if len(caseClause.List) > 0 {
				v.complexity++
			}
			v.visitBlockStmtList(caseClause.Body, nestingLevel+1)
		}
	}
}

func (v *cognitiveVisitor) processSelectCases(body *ast.BlockStmt, nestingLevel int) {
	if body == nil {
		return
	}

	for _, stmt := range body.List {
		if commClause, ok := stmt.(*ast.CommClause); ok {
			if commClause.Comm != nil {
				v.complexity++
			}
			v.visitBlockStmtList(commClause.Body, nestingLevel+1)
		}
	}
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
