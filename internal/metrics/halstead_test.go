package metrics

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"
)

func TestHalstead_Simple(t *testing.T) {
	// Arrange
	code := `
package main

func add(a, b int) int {
	return a + b
}
`

	// Act
	metrics := analyzeHalstead(t, code)

	// Assert
	// Operators: func, return, +, int (type)
	// Operands: add, a, b
	// This is a simplified check - exact counts depend on AST traversal
	if metrics.Effort == 0 {
		t.Error("Expected non-zero Halstead effort")
	}

	if metrics.Volume == 0 {
		t.Error("Expected non-zero Halstead volume")
	}
}

func TestHalstead_IfStatement(t *testing.T) {
	// Arrange
	code := `
package main

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
`

	// Act
	metrics := analyzeHalstead(t, code)

	// Assert
	// Should have some operators and operands
	// Exact counts depend on AST traversal, so we just check they're non-zero
	if metrics.DistinctOperators == 0 {
		t.Error("Expected non-zero distinct operators")
	}

	if metrics.DistinctOperands == 0 {
		t.Error("Expected non-zero distinct operands")
	}

	// Effort should be non-zero
	if metrics.Effort == 0 {
		t.Error("Expected non-zero Halstead effort")
	}
}

func TestHalstead_Loop(t *testing.T) {
	// Arrange
	code := `
package main

func sum(numbers []int) int {
	total := 0
	for _, num := range numbers {
		total = total + num
	}
	return total
}
`

	// Act
	metrics := analyzeHalstead(t, code)

	// Assert
	// Operators: func, :=, for, range, =, +, return
	// Operands: sum, numbers, total, num, 0
	if metrics.TotalOperators == 0 {
		t.Error("Expected non-zero total operators")
	}

	if metrics.TotalOperands == 0 {
		t.Error("Expected non-zero total operands")
	}

	// Volume = Length * log2(Vocabulary)
	expectedLength := metrics.TotalOperators + metrics.TotalOperands
	expectedVocabulary := metrics.DistinctOperators + metrics.DistinctOperands
	if expectedVocabulary > 0 {
		expectedVolume := float64(expectedLength) * math.Log2(float64(expectedVocabulary))
		if math.Abs(metrics.Volume-expectedVolume) > 0.01 {
			t.Errorf("Expected volume %.2f, got %.2f", expectedVolume, metrics.Volume)
		}
	}
}

func TestHalstead_ComplexFunction(t *testing.T) {
	// Arrange
	code := `
package main

func process(items []string, threshold int) ([]string, error) {
	var result []string
	count := 0

	for i, item := range items {
		if len(item) > threshold {
			result = append(result, item)
			count++
		} else if i > 10 {
			break
		}
	}

	if count == 0 {
		return nil, errors.New("no items")
	}

	return result, nil
}
`

	// Act
	metrics := analyzeHalstead(t, code)

	// Assert
	// This function has many operators and operands
	// Verify basic properties
	if metrics.DistinctOperators == 0 {
		t.Error("Expected distinct operators")
	}

	if metrics.DistinctOperands == 0 {
		t.Error("Expected distinct operands")
	}

	if metrics.Vocabulary != metrics.DistinctOperators+metrics.DistinctOperands {
		t.Errorf("Vocabulary should equal sum of distinct operators and operands")
	}

	if metrics.Length != metrics.TotalOperators+metrics.TotalOperands {
		t.Errorf("Length should equal sum of total operators and operands")
	}

	// Effort = Difficulty * Volume
	if metrics.Difficulty > 0 && metrics.Volume > 0 {
		expectedEffort := metrics.Difficulty * metrics.Volume
		if math.Abs(metrics.Effort-expectedEffort) > 0.01 {
			t.Errorf("Effort calculation incorrect: expected %.2f, got %.2f",
				expectedEffort, metrics.Effort)
		}
	}
}

func TestHalstead_BinaryOperators(t *testing.T) {
	// Arrange
	code := `
package main

func calculate(a, b, c int) int {
	x := a + b
	y := a - b
	z := a * b
	w := a / b
	return x + y + z + w
}
`

	// Act
	metrics := analyzeHalstead(t, code)

	// Assert
	// Should have binary operators: +, -, *, /, :=, return
	if metrics.DistinctOperators < 5 {
		t.Errorf("Expected at least 5 distinct operators, got %d", metrics.DistinctOperators)
	}

	// Multiple uses of + operator should increase total
	if metrics.TotalOperators <= metrics.DistinctOperators {
		t.Error("Total operators should be greater than distinct operators")
	}
}

// Helper function to analyze code and return Halstead metrics
func analyzeHalstead(t *testing.T, code string) HalsteadMetrics {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewHalsteadAnalyzer()
	fileMetrics := analyzer.AnalyzeFile(node)

	if len(fileMetrics.Functions) == 0 {
		t.Fatal("No functions found in code")
	}

	// Return the metrics for the first function
	// To get detailed metrics, we need to call calculateHalstead directly
	// For now, just verify that effort was calculated
	funcMetric := fileMetrics.Functions[0]

	// We need to re-parse to get detailed metrics
	// This is a simplification - in real use, we'd expose more details
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil {
				return calculateHalstead(fn.Body)
			}
		}
	}

	return HalsteadMetrics{Effort: funcMetric.Value}
}

func TestHalsteadAnalyzer_AnalyzeFile(t *testing.T) {
	// Arrange
	code := `
package main

func simple() {
	x := 1
}

func add(a, b int) int {
	return a + b
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewHalsteadAnalyzer()

	// Act
	metrics := analyzer.AnalyzeFile(node)

	// Assert
	if len(metrics.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(metrics.Functions))
	}

	// Check that file-level metrics are calculated
	if metrics.FileLevel["function_count"] != 2 {
		t.Errorf("Expected function_count 2, got %v", metrics.FileLevel["function_count"])
	}

	if metrics.FileLevel["total_effort"] == 0 {
		t.Error("Expected non-zero total effort")
	}

	if metrics.FileLevel["average_effort"] == 0 {
		t.Error("Expected non-zero average effort")
	}
}
