package treesitter

import (
	"math"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
)

// FileMetrics contains metrics for a source file
type FileMetrics struct {
	FilePath            string
	CognitiveComplexity int
	HalsteadVolume      float64
	HalsteadDifficulty  float64
	HalsteadEffort      float64
}

// MetricsCalculator calculates code metrics using tree-sitter
type MetricsCalculator struct {
	parser   *Parser
	language Language
}

// NewMetricsCalculator creates a new metrics calculator for the given language
func NewMetricsCalculator(lang Language) (*MetricsCalculator, error) {
	parser, err := New(lang)
	if err != nil {
		return nil, err
	}

	return &MetricsCalculator{
		parser:   parser,
		language: lang,
	}, nil
}

// CalculateFromFile calculates metrics for a source file
func (m *MetricsCalculator) CalculateFromFile(filePath string) (*FileMetrics, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := m.parser.Parse(source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	metrics := &FileMetrics{
		FilePath: filePath,
	}

	// Calculate cognitive complexity
	metrics.CognitiveComplexity = m.calculateCognitiveComplexity(tree, source)

	// Calculate Halstead metrics
	halstead := m.calculateHalsteadMetrics(tree, source)
	metrics.HalsteadVolume = halstead.Volume
	metrics.HalsteadDifficulty = halstead.Difficulty
	metrics.HalsteadEffort = halstead.Effort

	return metrics, nil
}

// calculateCognitiveComplexity calculates cognitive complexity for a tree
func (m *MetricsCalculator) calculateCognitiveComplexity(tree *sitter.Tree, source []byte) int {
	complexity := 0
	nestingLevel := 0

	// Walk the tree and calculate complexity
	var walk func(*sitter.Node)
	walk = func(node *sitter.Node) {
		nodeType := node.Type()

		// Structures that increase nesting
		nestingIncrease := false
		if m.isNestingStructure(nodeType) {
			nestingLevel++
			nestingIncrease = true
		}

		// Structures that add complexity
		if m.isComplexityNode(nodeType) {
			// Base complexity + nesting increment
			complexity += 1 + nestingLevel
		}

		// Recursively process children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				walk(child)
			}
		}

		// Decrease nesting on exit
		if nestingIncrease {
			nestingLevel--
		}
	}

	walk(tree.RootNode())
	return complexity
}

// isNestingStructure checks if a node type increases nesting
func (m *MetricsCalculator) isNestingStructure(nodeType string) bool {
	nestingStructures := map[string]bool{
		// Common across languages
		"if_statement":         true,
		"for_statement":        true,
		"while_statement":      true,
		"do_statement":         true,
		"switch_statement":     true,
		"case_clause":          true,
		"function_declaration": true,
		"method_declaration":   true,
		"lambda":               true,

		// Language-specific variants
		"if_expression":    true,
		"for_in_statement": true,
		"for_of_statement": true,
		"while_expression": true,
		"match_statement":  true, // Rust, Python 3.10+
		"with_statement":   true, // Python
		"try_statement":    true,
		"catch_clause":     true,
	}

	return nestingStructures[nodeType]
}

// isComplexityNode checks if a node adds to complexity
func (m *MetricsCalculator) isComplexityNode(nodeType string) bool {
	complexityNodes := map[string]bool{
		// Control flow
		"if_statement":    true,
		"else_clause":     true,
		"for_statement":   true,
		"while_statement": true,
		"do_statement":    true,
		"switch_statement": true,
		"case_clause":     true,

		// Logical operators (short-circuit)
		"binary_expression": true, // Will need refinement
		"boolean_operator":  true,

		// Exception handling
		"try_statement":    true,
		"catch_clause":     true,
		"except_clause":    true,

		// Jumps/breaks in flow
		"break_statement":    true,
		"continue_statement": true,
		"return_statement":   true,
		"goto_statement":     true,

		// Recursion
		"call_expression": true, // Will need refinement
	}

	return complexityNodes[nodeType]
}

// HalsteadMetrics contains Halstead complexity measures
type HalsteadMetrics struct {
	Operators       int
	Operands        int
	UniqueOperators int
	UniqueOperands  int
	Volume          float64
	Difficulty      float64
	Effort          float64
}

// calculateHalsteadMetrics calculates Halstead metrics for a tree
func (m *MetricsCalculator) calculateHalsteadMetrics(tree *sitter.Tree, source []byte) HalsteadMetrics {
	operators := make(map[string]int)
	operands := make(map[string]int)

	var walk func(*sitter.Node)
	walk = func(node *sitter.Node) {
		nodeType := node.Type()

		if m.isOperator(nodeType) {
			content := string(source[node.StartByte():node.EndByte()])
			operators[content]++
		} else if m.isOperand(nodeType) {
			content := string(source[node.StartByte():node.EndByte()])
			operands[content]++
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				walk(child)
			}
		}
	}

	walk(tree.RootNode())

	// Calculate metrics
	n1 := len(operators) // Unique operators
	n2 := len(operands)  // Unique operands

	N1 := 0 // Total operators
	for _, count := range operators {
		N1 += count
	}

	N2 := 0 // Total operands
	for _, count := range operands {
		N2 += count
	}

	// Vocabulary
	n := float64(n1 + n2)

	// Length
	N := float64(N1 + N2)

	// Volume: V = N * log2(n)
	volume := 0.0
	if n > 0 {
		volume = N * math.Log2(n)
	}

	// Difficulty: D = (n1 / 2) * (N2 / n2)
	difficulty := 0.0
	if n2 > 0 {
		difficulty = (float64(n1) / 2.0) * (float64(N2) / float64(n2))
	}

	// Effort: E = D * V
	effort := difficulty * volume

	return HalsteadMetrics{
		Operators:       N1,
		Operands:        N2,
		UniqueOperators: n1,
		UniqueOperands:  n2,
		Volume:          volume,
		Difficulty:      difficulty,
		Effort:          effort,
	}
}

// isOperator checks if a node type is an operator
func (m *MetricsCalculator) isOperator(nodeType string) bool {
	operators := map[string]bool{
		// Arithmetic
		"+": true, "-": true, "*": true, "/": true, "%": true,

		// Comparison
		"==": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true,

		// Logical
		"&&": true, "||": true, "!": true, "and": true, "or": true, "not": true,

		// Assignment
		"=": true, "+=": true, "-=": true, "*=": true, "/=": true,

		// Node types
		"binary_expression":     true,
		"unary_expression":      true,
		"update_expression":     true,
		"assignment_expression": true,
		"comparison_operator":   true,
		"boolean_operator":      true,
		"arithmetic_operator":   true,
	}

	return operators[nodeType]
}

// isOperand checks if a node type is an operand
func (m *MetricsCalculator) isOperand(nodeType string) bool {
	operands := map[string]bool{
		"identifier":          true,
		"number":              true,
		"integer":             true,
		"float":               true,
		"string":              true,
		"true":                true,
		"false":               true,
		"null":                true,
		"undefined":           true,
		"type_identifier":     true,
		"field_identifier":    true,
		"property_identifier": true,
	}

	return operands[nodeType]
}
