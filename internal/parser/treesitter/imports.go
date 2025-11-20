package treesitter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// Import represents an import statement found in source code
type Import struct {
	SourceFile    string
	ImportPath    string
	IsRelative    bool
	ImportedNames []string
	Line          uint32
}

// ImportExtractor extracts imports from source code using tree-sitter
type ImportExtractor struct {
	parser   *Parser
	language Language
}

// NewImportExtractor creates a new import extractor for the given language
func NewImportExtractor(lang Language) (*ImportExtractor, error) {
	parser, err := New(lang)
	if err != nil {
		return nil, err
	}

	return &ImportExtractor{
		parser:   parser,
		language: lang,
	}, nil
}

// ExtractFromFile extracts imports from a source file
func (e *ImportExtractor) ExtractFromFile(filePath string) ([]Import, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	tree, err := e.parser.Parse(source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	return e.extractImports(tree, source, filePath)
}

// ExtractFromSource extracts imports from source code
func (e *ImportExtractor) ExtractFromSource(sourceCode []byte, filePath string) ([]Import, error) {
	tree, err := e.parser.Parse(sourceCode)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	return e.extractImports(tree, sourceCode, filePath)
}

// extractImports extracts imports based on the language
func (e *ImportExtractor) extractImports(tree *sitter.Tree, source []byte, filePath string) ([]Import, error) {
	switch e.language {
	case LanguageGo:
		return e.extractGoImports(tree, source, filePath)
	case LanguagePython:
		return e.extractPythonImports(tree, source, filePath)
	case LanguageJavaScript, LanguageTypeScript:
		return e.extractJSImports(tree, source, filePath)
	case LanguageJava:
		return e.extractJavaImports(tree, source, filePath)
	default:
		return nil, fmt.Errorf("unsupported language: %s", e.language)
	}
}

// extractGoImports extracts imports from Go source code
func (e *ImportExtractor) extractGoImports(tree *sitter.Tree, source []byte, filePath string) ([]Import, error) {
	// Query for import declarations
	queryString := `(import_declaration
		(import_spec
			path: (interpreted_string_literal) @import.path
		)
	)
	(import_declaration
		(import_spec_list
			(import_spec
				path: (interpreted_string_literal) @import.path
			)
		)
	)`

	return e.executeQuery(tree, source, queryString, filePath, func(match *sitter.QueryMatch, src []byte) Import {
		importPathNode := match.Captures[0].Node
		importPath := string(src[importPathNode.StartByte():importPathNode.EndByte()])
		// Remove quotes
		importPath = strings.Trim(importPath, `"`)

		return Import{
			SourceFile: filePath,
			ImportPath: importPath,
			IsRelative: strings.HasPrefix(importPath, "."),
			Line:       importPathNode.StartPoint().Row,
		}
	})
}

// extractPythonImports extracts imports from Python source code
func (e *ImportExtractor) extractPythonImports(tree *sitter.Tree, source []byte, filePath string) ([]Import, error) {
	// Query for both 'import' and 'from ... import' statements
	queryString := `(import_statement
		name: (dotted_name) @import.module
	)
	(import_from_statement
		module_name: (dotted_name) @import.module
	)`

	return e.executeQuery(tree, source, queryString, filePath, func(match *sitter.QueryMatch, src []byte) Import {
		moduleNode := match.Captures[0].Node
		importPath := string(src[moduleNode.StartByte():moduleNode.EndByte()])

		return Import{
			SourceFile: filePath,
			ImportPath: importPath,
			IsRelative: strings.HasPrefix(importPath, "."),
			Line:       moduleNode.StartPoint().Row,
		}
	})
}

// extractJSImports extracts imports from JavaScript/TypeScript source code
func (e *ImportExtractor) extractJSImports(tree *sitter.Tree, source []byte, filePath string) ([]Import, error) {
	// Query for import statements and require calls
	queryString := `(import_statement
		source: (string) @import.source
	)
	(call_expression
		function: (identifier) @fn.name (#eq? @fn.name "require")
		arguments: (arguments (string) @import.source)
	)`

	return e.executeQuery(tree, source, queryString, filePath, func(match *sitter.QueryMatch, src []byte) Import {
		sourceNode := match.Captures[len(match.Captures)-1].Node
		importPath := string(src[sourceNode.StartByte():sourceNode.EndByte()])
		// Remove quotes
		importPath = strings.Trim(importPath, `"'`)

		return Import{
			SourceFile: filePath,
			ImportPath: importPath,
			IsRelative: strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/"),
			Line:       sourceNode.StartPoint().Row,
		}
	})
}

// extractJavaImports extracts imports from Java source code
func (e *ImportExtractor) extractJavaImports(tree *sitter.Tree, source []byte, filePath string) ([]Import, error) {
	// Query for import declarations
	queryString := `(import_declaration
		(scoped_identifier) @import.path
	)
	(import_declaration
		(identifier) @import.path
	)`

	// First, extract package name for relative import detection
	packageName := e.extractJavaPackage(tree, source)

	return e.executeQuery(tree, source, queryString, filePath, func(match *sitter.QueryMatch, src []byte) Import {
		importNode := match.Captures[0].Node
		importPath := string(src[importNode.StartByte():importNode.EndByte()])
		isRelative := packageName != "" && strings.HasPrefix(importPath, packageName+".")

		return Import{
			SourceFile: filePath,
			ImportPath: importPath,
			IsRelative: isRelative,
			Line:       importNode.StartPoint().Row,
		}
	})
}

// extractJavaPackage extracts the package name from Java source
func (e *ImportExtractor) extractJavaPackage(tree *sitter.Tree, source []byte) string {
	queryString := `(package_declaration
		(scoped_identifier) @package.name
	)`

	cursor, query, err := e.parser.Query(tree, queryString)
	if err != nil {
		return ""
	}
	defer cursor.Close()
	defer query.Close()

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		if len(match.Captures) > 0 {
			node := match.Captures[0].Node
			return string(source[node.StartByte():node.EndByte()])
		}
	}

	return ""
}

// executeQuery is a helper to execute a query and process results
func (e *ImportExtractor) executeQuery(
	tree *sitter.Tree,
	source []byte,
	queryString string,
	filePath string,
	processMatch func(*sitter.QueryMatch, []byte) Import,
) ([]Import, error) {
	cursor, query, err := e.parser.Query(tree, queryString)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	defer query.Close()

	var imports []Import

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		if len(match.Captures) > 0 {
			imp := processMatch(match, source)
			imports = append(imports, imp)
		}
	}

	return imports, nil
}

// ResolveImportPath resolves a relative import path to an absolute path
func ResolveImportPath(sourceFile, importPath string) string {
	if !strings.HasPrefix(importPath, ".") {
		return importPath
	}

	sourceDir := filepath.Dir(sourceFile)
	resolvedPath := filepath.Join(sourceDir, importPath)
	return filepath.Clean(resolvedPath)
}
