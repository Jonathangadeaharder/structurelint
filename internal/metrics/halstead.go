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

// goBuiltins contains all Go built-in types, functions, and constants
var goBuiltins = map[string]bool{
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
	operators := make(map[string]int)     // operator -> count
	operands := make(map[string]int)      // operand -> count
	totalOperators := 0
	totalOperands := 0

	// Traverse the AST and count operators and operands
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch node := n.(type) {
		// Operators
		case *ast.BinaryExpr:
			op := node.Op.String()
			operators[op]++
			totalOperators++
			return true

		case *ast.UnaryExpr:
			op := node.Op.String()
			operators[op]++
			totalOperators++
			return true

		case *ast.AssignStmt:
			op := node.Tok.String()
			operators[op]++
			totalOperators++
			return true

		case *ast.IncDecStmt:
			op := node.Tok.String()
			operators[op]++
			totalOperators++
			return true

		case *ast.IfStmt:
			operators["if"]++
			totalOperators++
			return true

		case *ast.ForStmt:
			operators["for"]++
			totalOperators++
			return true

		case *ast.RangeStmt:
			operators["range"]++
			totalOperators++
			return true

		case *ast.SwitchStmt:
			operators["switch"]++
			totalOperators++
			return true

		case *ast.TypeSwitchStmt:
			operators["type-switch"]++
			totalOperators++
			return true

		case *ast.SelectStmt:
			operators["select"]++
			totalOperators++
			return true

		case *ast.CaseClause:
			if len(node.List) > 0 {
				operators["case"]++
				totalOperators++
			} else {
				operators["default"]++
				totalOperators++
			}
			return true

		case *ast.CommClause:
			operators["case"]++
			totalOperators++
			return true

		case *ast.BranchStmt:
			op := node.Tok.String()
			operators[op]++
			totalOperators++
			return true

		case *ast.ReturnStmt:
			operators["return"]++
			totalOperators++
			return true

		case *ast.DeferStmt:
			operators["defer"]++
			totalOperators++
			return true

		case *ast.GoStmt:
			operators["go"]++
			totalOperators++
			return true

		case *ast.CallExpr:
			operators["()"]++ // Function call operator
			totalOperators++
			return true

		case *ast.IndexExpr:
			operators["[]"]++ // Array/slice index operator
			totalOperators++
			return true

		case *ast.SliceExpr:
			operators["[:]"]++ // Slice operator
			totalOperators++
			return true

		case *ast.StarExpr:
			operators["*"]++ // Pointer dereference or type
			totalOperators++
			return true

		case *ast.TypeAssertExpr:
			operators[".(type)"]++ // Type assertion
			totalOperators++
			return true

		case *ast.SendStmt:
			operators["<-"]++ // Channel send
			totalOperators++
			return true

		// Operands
		case *ast.Ident:
			// Skip keywords that we've already counted as operators
			if !token.IsKeyword(node.Name) {
				// Skip built-in types and predeclared identifiers in some contexts
				if !isBuiltin(node.Name) {
					operands[node.Name]++
					totalOperands++
				}
			}
			return true

		case *ast.BasicLit:
			// Literal values (numbers, strings, etc.)
			operands[node.Value]++
			totalOperands++
			return true
		}

		return true
	})

	// Calculate Halstead metrics
	n1 := len(operators)
	n2 := len(operands)
	N1 := totalOperators
	N2 := totalOperands

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
	return goBuiltins[name]
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
