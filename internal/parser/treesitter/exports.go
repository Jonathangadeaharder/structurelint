package treesitter

import (
	"fmt"
	"os"

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
	// In Go, exported symbols start with uppercase letters - simplified version without regex
	var exports []Export
	
	// Walk the tree looking for top-level declarations
	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child == nil {
			continue
		}
		
		nodeType := child.Type()
		var nameNode *sitter.Node
		
		switch nodeType {
		case "function_declaration":
			nameNode = child.ChildByFieldName("name")
		case "type_declaration":
			// Get type_spec -> name
			for j := 0; j < int(child.ChildCount()); j++ {
				typeSpec := child.Child(j)
				if typeSpec != nil && typeSpec.Type() == "type_spec" {
					nameNode = typeSpec.ChildByFieldName("name")
					break
				}
			}
		}
		
		if nameNode != nil {
			name := string(source[nameNode.StartByte():nameNode.EndByte()])
			// Check if capitalized (exported in Go)
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				exports = append(exports, Export{
					SourceFile: filePath,
					Names:      []string{name},
					IsDefault:  false,
					Line:       nameNode.StartPoint().Row,
				})
			}
		}
	}
	
	return exports, nil
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
		
		nodeType := child.Type()
		if nodeType != "function_definition" && nodeType != "class_definition" {
			continue
		}
		
		nameNode := child.ChildByFieldName("name")
		if nameNode != nil {
			name := string(source[nameNode.StartByte():nameNode.EndByte()])
			// Skip private (starting with _)
			if len(name) > 0 && name[0] != '_' {
				exports = append(exports, Export{
					SourceFile: filePath,
					Names:      []string{name},
					IsDefault:  false,
					Line:       nameNode.StartPoint().Row,
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
