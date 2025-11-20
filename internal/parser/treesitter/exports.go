package treesitter

import (
	"fmt"
	"os"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
)

// Export represents an exported symbol from source code
type Export struct {
	SourceFile string
	Names      []string
	IsDefault  bool
	Line       uint32
}

// ExportExtractor extracts exports from source code using tree-sitter
type ExportExtractor struct {
	parser   *Parser
	language Language
}

// NewExportExtractor creates a new export extractor for the given language
func NewExportExtractor(lang Language) (*ExportExtractor, error) {
	parser, err := New(lang)
	if err != nil {
		return nil, err
	}

	return &ExportExtractor{
		parser:   parser,
		language: lang,
	}, nil
}

// ExtractFromFile extracts exports from a source file
func (e *ExportExtractor) ExtractFromFile(filePath string) ([]Export, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	tree, err := e.parser.Parse(source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	return e.extractExports(tree, source, filePath)
}

// extractExports extracts exports based on the language
func (e *ExportExtractor) extractExports(tree *sitter.Tree, source []byte, filePath string) ([]Export, error) {
	switch e.language {
	case LanguageGo:
		return e.extractGoExports(tree, source, filePath)
	case LanguagePython:
		return e.extractPythonExports(tree, source, filePath)
	case LanguageJavaScript, LanguageTypeScript:
		return e.extractJSExports(tree, source, filePath)
	case LanguageJava:
		return e.extractJavaExports(tree, source, filePath)
	default:
		return nil, fmt.Errorf("unsupported language: %s", e.language)
	}
}

// extractGoExports extracts public (capitalized) symbols from Go
func (e *ExportExtractor) extractGoExports(tree *sitter.Tree, source []byte, filePath string) ([]Export, error) {
	var exports []Export

	// Walk the tree looking for top-level declarations
	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "function_declaration":
			if exp := e.extractGoFunction(child, source, filePath); exp != nil {
				exports = append(exports, *exp)
			}
		case "type_declaration":
			exports = append(exports, e.extractGoTypes(child, source, filePath)...)
		}
	}

	return exports, nil
}

// extractGoFunction extracts a function export if it's exported
func (e *ExportExtractor) extractGoFunction(node *sitter.Node, source []byte, filePath string) *Export {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(source[nameNode.StartByte():nameNode.EndByte()])
	if !isGoExported(name) {
		return nil
	}

	return &Export{
		SourceFile: filePath,
		Names:      []string{name},
		IsDefault:  false,
		Line:       nameNode.StartPoint().Row + 1, // Convert 0-based to 1-based
	}
}

// extractGoTypes extracts all exported types from a type_declaration
func (e *ExportExtractor) extractGoTypes(node *sitter.Node, source []byte, filePath string) []Export {
	var exports []Export

	// Handle grouped type declarations: type ( Foo struct {}; Bar struct {} )
	for j := 0; j < int(node.ChildCount()); j++ {
		typeSpec := node.Child(j)
		if typeSpec == nil || typeSpec.Type() != "type_spec" {
			continue
		}

		if exp := e.extractGoTypeSpec(typeSpec, source, filePath); exp != nil {
			exports = append(exports, *exp)
		}
	}

	return exports
}

// extractGoTypeSpec extracts a single type spec if it's exported
func (e *ExportExtractor) extractGoTypeSpec(typeSpec *sitter.Node, source []byte, filePath string) *Export {
	nameNode := typeSpec.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(source[nameNode.StartByte():nameNode.EndByte()])
	if !isGoExported(name) {
		return nil
	}

	return &Export{
		SourceFile: filePath,
		Names:      []string{name},
		IsDefault:  false,
		Line:       nameNode.StartPoint().Row + 1, // Convert 0-based to 1-based
	}
}

// isGoExported checks if a Go identifier is exported (starts with uppercase Unicode letter)
func isGoExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	firstRune := []rune(name)[0]
	return unicode.IsUpper(firstRune)
}

// extractPythonExports extracts symbols from Python
func (e *ExportExtractor) extractPythonExports(tree *sitter.Tree, source []byte, filePath string) ([]Export, error) {
	// Python: top-level functions and classes (not starting with _)
	var exports []Export

	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child == nil {
			continue
		}

		// Handle decorated definitions (@decorator syntax)
		// In tree-sitter-python, decorated functions/classes are wrapped in decorated_definition
		definitionNode := child
		if child.Type() == "decorated_definition" {
			// Find the inner function_definition or class_definition
			for j := 0; j < int(child.ChildCount()); j++ {
				innerNode := child.Child(j)
				if innerNode != nil && (innerNode.Type() == "function_definition" || innerNode.Type() == "class_definition") {
					definitionNode = innerNode
					break
				}
			}
		}

		nodeType := definitionNode.Type()
		if nodeType != "function_definition" && nodeType != "class_definition" {
			continue
		}

		nameNode := definitionNode.ChildByFieldName("name")
		if nameNode != nil {
			name := string(source[nameNode.StartByte():nameNode.EndByte()])
			// Skip private (starting with _)
			if len(name) > 0 && name[0] != '_' {
				exports = append(exports, Export{
					SourceFile: filePath,
					Names:      []string{name},
					IsDefault:  false,
					Line:       nameNode.StartPoint().Row + 1, // Convert 0-based to 1-based
				})
			}
		}
	}

	return exports, nil
}

// extractJSExports extracts exports from JavaScript/TypeScript
func (e *ExportExtractor) extractJSExports(tree *sitter.Tree, source []byte, filePath string) ([]Export, error) {
	// For now, return empty - full implementation would walk export_statement nodes
	return []Export{}, nil
}

// extractJavaExports extracts public classes/interfaces from Java
func (e *ExportExtractor) extractJavaExports(tree *sitter.Tree, source []byte, filePath string) ([]Export, error) {
	// For now, return empty - full implementation would check modifiers
	return []Export{}, nil
}
