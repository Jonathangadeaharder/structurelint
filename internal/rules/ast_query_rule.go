package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/parser/treesitter"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// ASTQueryRule checks files using tree-sitter AST queries
type ASTQueryRule struct {
	name        string
	description string
	queries     map[treesitter.Language]string
	checkFunc   func(matches []*QueryMatch, file walker.FileInfo) []Violation
}

// QueryMatch represents a matched AST node
type QueryMatch struct {
	Text       string
	Line       int
	Column     int
	NodeType   string
	CaptureMap map[string]string
}

// NewASTQueryRule creates a new AST query-based rule
func NewASTQueryRule(name, description string, queries map[treesitter.Language]string, checkFunc func([]*QueryMatch, walker.FileInfo) []Violation) *ASTQueryRule {
	return &ASTQueryRule{
		name:        name,
		description: description,
		queries:     queries,
		checkFunc:   checkFunc,
	}
}

// Name returns the rule name
func (r *ASTQueryRule) Name() string {
	return r.name
}

// Check executes the rule and returns violations
func (r *ASTQueryRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Detect language from extension
		ext := filepath.Ext(file.Path)
		lang, err := treesitter.DetectLanguageFromExtension(ext)
		if err != nil {
			continue // Skip unsupported languages
		}

		// Get query for this language
		query, ok := r.queries[lang]
		if !ok {
			continue // No query defined for this language
		}

		// Parse file and execute query
		matches, err := r.executeQuery(file.AbsPath, lang, query)
		if err != nil {
			continue // Skip files with parse errors
		}

		// Check matches using custom function
		fileViolations := r.checkFunc(matches, file)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// executeQuery parses a file and executes a tree-sitter query
func (r *ASTQueryRule) executeQuery(filePath string, lang treesitter.Language, queryString string) ([]*QueryMatch, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Create parser
	parser, err := treesitter.New(lang)
	if err != nil {
		return nil, err
	}

	// Parse file
	tree, err := parser.Parse(content)
	if err != nil {
		return nil, err
	}

	// Execute query
	cursor, query, err := parser.Query(tree, queryString)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	defer query.Close()

	// Collect matches
	var matches []*QueryMatch
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			node := capture.Node
			text := string(content[node.StartByte():node.EndByte()])

			// Get capture name
			captureName := query.CaptureNameForId(capture.Index)

			matches = append(matches, &QueryMatch{
				Text:     text,
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				NodeType: node.Type(),
				CaptureMap: map[string]string{
					captureName: text,
				},
			})
		}
	}

	return matches, nil
}

// --- Example AST query rules ---

// ExampleRequireInterfaceRule checks that domain entities implement interfaces
func ExampleRequireInterfaceRule() Rule {
	queries := map[treesitter.Language]string{
		treesitter.LanguageGo: `
			(type_declaration
				(type_spec
					name: (type_identifier) @type.name
					type: (struct_type) @struct
				)
			) @decl
		`,
		treesitter.LanguageTypeScript: `
			(class_declaration
				name: (type_identifier) @type.name
			) @decl
		`,
	}

	return NewASTQueryRule(
		"require-interface",
		"Domain entities should implement interfaces",
		queries,
		func(matches []*QueryMatch, file walker.FileInfo) []Violation {
			var violations []Violation

			// Check if file is in domain layer
			if !strings.HasPrefix(filepath.ToSlash(file.Path), "internal/domain/") {
				return nil
			}

			// Check each struct definition
			for _, match := range matches {
				// This is a simplified example - real implementation would check
				// if the struct implements any interfaces
				if len(match.Text) > 1000 { // Large struct
					violations = append(violations, Violation{
						Rule:    "require-interface",
						Path:    file.Path,
						Message: fmt.Sprintf("Large domain entity at line %d should implement an interface", match.Line),
					})
				}
			}

			return violations
		},
	)
}

// ExampleDisallowDirectDBAccessRule checks for direct database access
func ExampleDisallowDirectDBAccessRule() Rule {
	queries := map[treesitter.Language]string{
		treesitter.LanguageGo: `
			(call_expression
				function: (selector_expression
					field: (field_identifier) @method
				)
			) @call
		`,
		treesitter.LanguagePython: `
			(call
				function: (attribute
					attribute: (identifier) @method
				)
			) @call
		`,
	}

	return NewASTQueryRule(
		"no-direct-db-access",
		"Domain layer should not access database directly",
		queries,
		func(matches []*QueryMatch, file walker.FileInfo) []Violation {
			var violations []Violation

			// Only check domain layer
			if !strings.HasPrefix(filepath.ToSlash(file.Path), "internal/domain/") {
				return nil
			}

			// Check for database-related method calls
			dbMethods := []string{"Query", "Exec", "QueryRow", "Prepare", "Begin"}
			for _, match := range matches {
				for _, dbMethod := range dbMethods {
					if match.CaptureMap["method"] == dbMethod {
						violations = append(violations, Violation{
							Rule:    "no-direct-db-access",
							Path:    file.Path,
							Message: fmt.Sprintf("Direct database access at line %d (method: %s) - use repository pattern", match.Line, dbMethod),
						})
					}
				}
			}

			return violations
		},
	)
}

// ExampleRequireDeprecationCommentRule checks for @deprecated annotations
func ExampleRequireDeprecationCommentRule() Rule {
	queries := map[treesitter.Language]string{
		treesitter.LanguageGo: `
			(function_declaration
				name: (identifier) @func.name
			) @func
		`,
		treesitter.LanguageTypeScript: `
			(function_declaration
				name: (identifier) @func.name
			) @func
		`,
	}

	return NewASTQueryRule(
		"require-deprecation-comment",
		"Deprecated functions must have // @deprecated comment",
		queries,
		func(matches []*QueryMatch, file walker.FileInfo) []Violation {
			var violations []Violation

			// This is a simplified example - real implementation would check
			// for @deprecated comment above the function
			for _, match := range matches {
				funcName := match.CaptureMap["func.name"]
				if funcName != "" && strings.HasPrefix(funcName, "Old") {
					// Check if there's a deprecation comment (simplified)
					violations = append(violations, Violation{
						Rule:    "require-deprecation-comment",
						Path:    file.Path,
						Message: fmt.Sprintf("Function '%s' at line %d appears deprecated but lacks @deprecated comment", funcName, match.Line),
					})
				}
			}

			return violations
		},
	)
}

// ExampleDisallowGlobalVariablesRule checks for global variables
func ExampleDisallowGlobalVariablesRule() Rule {
	queries := map[treesitter.Language]string{
		treesitter.LanguageGo: `
			(var_declaration
				(var_spec
					name: (identifier) @var.name
				)
			) @var
		`,
		treesitter.LanguagePython: `
			(module
				(expression_statement
					(assignment
						left: (identifier) @var.name
					)
				)
			) @var
		`,
	}

	return NewASTQueryRule(
		"no-global-variables",
		"Global variables are disallowed (use dependency injection)",
		queries,
		func(matches []*QueryMatch, file walker.FileInfo) []Violation {
			var violations []Violation

			// Skip test files
			if strings.HasSuffix(filepath.Base(file.Path), "_test.go") {
				return nil
			}

			for _, match := range matches {
				varName := match.CaptureMap["var.name"]
				// Allow certain patterns (constants, errors)
				if len(varName) > 0 && varName[0] >= 'a' && varName[0] <= 'z' {
					// Lowercase = likely a variable (not a constant)
					violations = append(violations, Violation{
						Rule:    "no-global-variables",
						Path:    file.Path,
						Message: fmt.Sprintf("Global variable '%s' at line %d (use dependency injection instead)", varName, match.Line),
					})
				}
			}

			return violations
		},
	)
}
