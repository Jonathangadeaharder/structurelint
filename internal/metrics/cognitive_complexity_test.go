package metrics

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestCognitiveComplexity_Simple(t *testing.T) {
	// Arrange
	code := `
package main

func simple() {
	x := 1
	y := 2
	z := x + y
}
`

	// Act
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 0 {
		t.Errorf("Expected complexity 0 for simple function, got %d", complexity)
	}
}

func TestCognitiveComplexity_SingleIf(t *testing.T) {
	// Arrange
	code := `
package main

func singleIf(x int) {
	if x > 0 {
		println(x)
	}
}
`

	// Act
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 1 {
		t.Errorf("Expected complexity 1 for single if, got %d", complexity)
	}
}

func TestCognitiveComplexity_NestedIf(t *testing.T) {
	// Arrange
	code := `
package main

func nestedIf(x, y int) {
	if x > 0 {          // +1
		if y > 0 {      // +1 (if) + 1 (nesting) = +2
			println(x, y)
		}
	}
}
`

	// Act
	// Expected: 1 + 2 = 3
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 3 {
		t.Errorf("Expected complexity 3 for nested if, got %d", complexity)
	}
}

func TestCognitiveComplexity_ForLoop(t *testing.T) {
	// Arrange
	code := `
package main

func forLoop(items []int) {
	for _, item := range items {
		println(item)
	}
}
`

	// Act
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 1 {
		t.Errorf("Expected complexity 1 for simple for loop, got %d", complexity)
	}
}

func TestCognitiveComplexity_NestedLoopAndIf(t *testing.T) {
	// Arrange
	code := `
package main

func nestedLoopAndIf(items []int) {
	for _, item := range items {        // +1
		if item > 0 {                   // +1 (if) + 1 (nesting) = +2
			println(item)
		}
	}
}
`

	// Act
	// Expected: 1 + 2 = 3
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 3 {
		t.Errorf("Expected complexity 3 for nested loop and if, got %d", complexity)
	}
}

func TestCognitiveComplexity_Switch(t *testing.T) {
	// Arrange
	code := `
package main

func switchStmt(x int) {
	switch x {        // +1
	case 1:           // +1
		println("one")
	case 2:           // +1
		println("two")
	case 3:           // +1
		println("three")
	default:          // +0 (default doesn't count)
		println("other")
	}
}
`

	// Act
	// Expected: 1 (switch) + 3 (cases) = 4
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 4 {
		t.Errorf("Expected complexity 4 for switch with 3 cases, got %d", complexity)
	}
}

func TestCognitiveComplexity_ElseIf(t *testing.T) {
	// Arrange
	code := `
package main

func elseIf(x int) {
	if x == 1 {           // +1
		println("one")
	} else if x == 2 {    // +1 (else-if doesn't add nesting)
		println("two")
	} else {
		println("other")
	}
}
`

	// Act
	// Expected: 1 + 1 = 2
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 2 {
		t.Errorf("Expected complexity 2 for if-else-if, got %d", complexity)
	}
}

func TestCognitiveComplexity_DeeplyNested(t *testing.T) {
	// Arrange
	code := `
package main

func deeplyNested(x, y, z int) {
	if x > 0 {              // +1 (nesting 0)
		if y > 0 {          // +1 + 1 (nesting 1) = +2
			if z > 0 {      // +1 + 2 (nesting 2) = +3
				println(x, y, z)
			}
		}
	}
}
`

	// Act
	// Expected: 1 + 2 + 3 = 6
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 6 {
		t.Errorf("Expected complexity 6 for deeply nested, got %d", complexity)
	}
}

func TestCognitiveComplexity_BranchStatements(t *testing.T) {
	// Arrange
	code := `
package main

func withBreak(items []int) {
	for _, item := range items {  // +1
		if item == 0 {            // +1 + 1 (nesting level 1) = +2
			break                 // +1 + 2 (nesting level 2) = +3
		}
		println(item)
	}
}
`

	// Act
	// Expected: 1 + 2 + 3 = 6
	complexity := analyzeCode(t, code)

	// Assert
	if complexity != 6 {
		t.Errorf("Expected complexity 6 with break statement, got %d", complexity)
	}
}

// Helper function to analyze code and return cognitive complexity
func analyzeCode(t *testing.T, code string) int {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCognitiveComplexityAnalyzer()
	fileMetrics := analyzer.AnalyzeFile(node)

	if len(fileMetrics.Functions) == 0 {
		t.Fatal("No functions found in code")
	}

	return fileMetrics.Functions[0].Complexity
}

func TestCognitiveComplexityAnalyzer_AnalyzeFile(t *testing.T) {
	// Arrange
	code := `
package main

func simple() {
	x := 1
}

func withIf(x int) {
	if x > 0 {
		println(x)
	}
}

func nested(x, y int) {
	if x > 0 {
		if y > 0 {
			println(x, y)
		}
	}
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCognitiveComplexityAnalyzer()

	// Act
	metrics := analyzer.AnalyzeFile(node)

	// Assert
	if len(metrics.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(metrics.Functions))
	}

	// Check individual function complexities
	expectedComplexities := map[string]int{
		"simple": 0,
		"withIf": 1,
		"nested": 3,
	}

	for _, funcMetric := range metrics.Functions {
		expected, ok := expectedComplexities[funcMetric.Name]
		if !ok {
			t.Errorf("Unexpected function: %s", funcMetric.Name)
			continue
		}

		if funcMetric.Complexity != expected {
			t.Errorf("Function %s: expected complexity %d, got %d",
				funcMetric.Name, expected, funcMetric.Complexity)
		}
	}

	// Check file-level metrics
	if metrics.FileLevel["function_count"] != 3 {
		t.Errorf("Expected function_count 3, got %v", metrics.FileLevel["function_count"])
	}

	expectedTotal := 0.0 + 1.0 + 3.0
	if metrics.FileLevel["total"] != expectedTotal {
		t.Errorf("Expected total %v, got %v", expectedTotal, metrics.FileLevel["total"])
	}

	expectedAvg := expectedTotal / 3.0
	if metrics.FileLevel["average"] != expectedAvg {
		t.Errorf("Expected average %v, got %v", expectedAvg, metrics.FileLevel["average"])
	}

	if metrics.FileLevel["max"] != 3 {
		t.Errorf("Expected max 3, got %v", metrics.FileLevel["max"])
	}
}
