// Halstead Complexity Measures implementation based on:
// - Halstead, M. (1977) "Elements of Software Science"
// - Scalabrino et al. (2022) EEG Study: rs=0.901 correlation with cognitive load
//
// Key findings:
// - Halstead Effort has >90% correlation with measured brain activity during code comprehension
// - Captures "data complexity" (vocabulary, operators, operands)
// - Complements Cognitive Complexity (which captures control-flow complexity)
//
// Metrics:
// - Volume (V): Information content in bits
// - Difficulty (D): How difficult to write/understand
// - Effort (E): Mental effort required (D × V)
//
// @structurelint:ignore test-adjacency Covered by halstead_test.go
package metrics

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
)

// HalsteadMetrics represents the complete Halstead metrics for a code unit
type HalsteadMetrics struct {
	DistinctOperators int     // n1: unique operators
	DistinctOperands  int     // n2: unique operands
	TotalOperators    int     // N1: total operators
	TotalOperands     int     // N2: total operands
	Vocabulary        int     // n = n1 + n2
	Length            int     // N = N1 + N2
	Volume            float64 // V = N × log₂(n)
	Difficulty        float64 // D = (n1/2) × (N2/n2)
	Effort            float64 // E = D × V
}

// HalsteadAnalyzer calculates Halstead metrics
type HalsteadAnalyzer struct{}

// NewHalsteadAnalyzer creates a new analyzer
func NewHalsteadAnalyzer() *HalsteadAnalyzer {
	return &HalsteadAnalyzer{}
}

// Name returns the metric name
func (a *HalsteadAnalyzer) Name() string {
	return "halstead"
}

// AnalyzeFunction computes Halstead metrics for a function
func (a *HalsteadAnalyzer) AnalyzeFunction(funcDecl *ast.FuncDecl) FunctionMetric {
	if funcDecl.Body == nil {
		return FunctionMetric{
			Name:  getFunctionNameHalstead(funcDecl),
			Value: 0,
		}
	}

	metrics := calculateHalstead(funcDecl.Body)

	return FunctionMetric{
		Name:      getFunctionNameHalstead(funcDecl),
		StartLine: int(funcDecl.Pos()),
		EndLine:   int(funcDecl.End()),
		Value:     metrics.Effort,
	}
}

// AnalyzeFile computes Halstead metrics for all functions in a file
func (a *HalsteadAnalyzer) AnalyzeFile(node *ast.File) FileMetrics {
	fileMetrics := FileMetrics{
		Functions: []FunctionMetric{},
		FileLevel: make(map[string]float64),
	}

	var totalEffort float64
	var functionCount int

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		funcMetric := a.AnalyzeFunction(funcDecl)
		fileMetrics.Functions = append(fileMetrics.Functions, funcMetric)

		totalEffort += funcMetric.Value
		functionCount++

		return true
	})

	if functionCount > 0 {
		fileMetrics.FileLevel["total_effort"] = totalEffort
		fileMetrics.FileLevel["average_effort"] = totalEffort / float64(functionCount)
		fileMetrics.FileLevel["max_effort"] = findMaxComplexityHalstead(fileMetrics.Functions)
		fileMetrics.FileLevel["function_count"] = float64(functionCount)
	}

	return fileMetrics
}

// calculateHalstead computes the Halstead metrics for a code block
func calculateHalstead(body *ast.BlockStmt) HalsteadMetrics {
	counter := &halsteadCounter{
		operators:      make(map[string]int),
		operands:       make(map[string]int),
		totalOperators: 0,
		totalOperands:  0,
	}

	ast.Inspect(body, counter.visit)

	return counter.computeMetrics()
}

// halsteadCounter tracks operator and operand counts
type halsteadCounter struct {
	operators      map[string]int
	operands       map[string]int
	totalOperators int
	totalOperands  int
}

func (h *halsteadCounter) visit(n ast.Node) bool {
	if n == nil {
		return false
	}

	switch node := n.(type) {
	case *ast.BinaryExpr:
		h.addOperator(node.Op.String())
	case *ast.UnaryExpr:
		h.addOperator(node.Op.String())
	case *ast.AssignStmt:
		h.addOperator(node.Tok.String())
	case *ast.IncDecStmt:
		h.addOperator(node.Tok.String())
	case *ast.IfStmt:
		h.addOperator("if")
	case *ast.ForStmt:
		h.addOperator("for")
	case *ast.RangeStmt:
		h.addOperator("range")
	case *ast.SwitchStmt:
		h.addOperator("switch")
	case *ast.TypeSwitchStmt:
		h.addOperator("type-switch")
	case *ast.SelectStmt:
		h.addOperator("select")
	case *ast.CaseClause:
		h.handleCaseClause(node)
	case *ast.CommClause:
		h.addOperator("case")
	case *ast.BranchStmt:
		h.addOperator(node.Tok.String())
	case *ast.ReturnStmt:
		h.addOperator("return")
	case *ast.DeferStmt:
		h.addOperator("defer")
	case *ast.GoStmt:
		h.addOperator("go")
	case *ast.CallExpr:
		h.addOperator("()")
	case *ast.IndexExpr:
		h.addOperator("[]")
	case *ast.SliceExpr:
		h.addOperator("[:]")
	case *ast.StarExpr:
		h.addOperator("*")
	case *ast.TypeAssertExpr:
		h.addOperator(".(type)")
	case *ast.SendStmt:
		h.addOperator("<-")
	case *ast.Ident:
		h.handleIdent(node)
	case *ast.BasicLit:
		h.addOperand(node.Value)
	}

	return true
}

func (h *halsteadCounter) addOperator(op string) {
	h.operators[op]++
	h.totalOperators++
}

func (h *halsteadCounter) addOperand(operand string) {
	h.operands[operand]++
	h.totalOperands++
}

func (h *halsteadCounter) handleCaseClause(node *ast.CaseClause) {
	if len(node.List) > 0 {
		h.addOperator("case")
	} else {
		h.addOperator("default")
	}
}

func (h *halsteadCounter) handleIdent(node *ast.Ident) {
	if !token.IsKeyword(node.Name) && !isBuiltin(node.Name) {
		h.addOperand(node.Name)
	}
}

func (h *halsteadCounter) computeMetrics() HalsteadMetrics {
	n1 := len(h.operators)
	n2 := len(h.operands)
	N1 := h.totalOperators
	N2 := h.totalOperands

	vocabulary := n1 + n2
	length := N1 + N2

	var volume, difficulty, effort float64

	if vocabulary > 0 {
		volume = float64(length) * math.Log2(float64(vocabulary))
	}

	if n2 > 0 {
		difficulty = (float64(n1) / 2.0) * (float64(N2) / float64(n2))
	}

	effort = difficulty * volume

	return HalsteadMetrics{
		DistinctOperators: n1,
		DistinctOperands:  n2,
		TotalOperators:    N1,
		TotalOperands:     N2,
		Vocabulary:        vocabulary,
		Length:            length,
		Volume:            volume,
		Difficulty:        difficulty,
		Effort:            effort,
	}
}

// isBuiltin checks if an identifier is a built-in type or function
func isBuiltin(name string) bool {
	builtins := map[string]bool{
		// Built-in types
		"bool": true, "byte": true, "complex64": true, "complex128": true,
		"error": true, "float32": true, "float64": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"rune": true, "string": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,

		// Built-in functions
		"append": true, "cap": true, "close": true, "complex": true, "copy": true,
		"delete": true, "imag": true, "len": true, "make": true, "new": true,
		"panic": true, "print": true, "println": true, "real": true, "recover": true,

		// Constants
		"true": true, "false": true, "iota": true, "nil": true,
	}
	return builtins[name]
}

// getFunctionNameHalstead extracts the function name for Halstead metrics
func getFunctionNameHalstead(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if len(funcDecl.Recv.List[0].Names) > 0 {
			receiverName := funcDecl.Recv.List[0].Names[0].Name
			return fmt.Sprintf("(%s).%s", receiverName, funcDecl.Name.Name)
		}
	}
	return funcDecl.Name.Name
}

// findMaxComplexityHalstead finds the maximum Halstead effort
func findMaxComplexityHalstead(functions []FunctionMetric) float64 {
	var max float64
	for _, f := range functions {
		if f.Value > max {
			max = f.Value
		}
	}
	return max
}
