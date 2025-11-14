package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/pkg/plugin"
)

// Import represents an import statement found in a source file
type Import struct {
	SourceFile    string   // The file containing the import
	ImportPath    string   // The imported path/module
	IsRelative    bool     // Whether this is a relative import
	ImportedNames []string // Specific symbols imported (for Phase 2)
}

// Export represents an export statement found in a source file (Phase 2)
type Export struct {
	SourceFile string   // The file containing the export
	Names      []string // Exported symbol names
	IsDefault  bool     // Whether this is a default export
}

// Parser extracts imports from source files
type Parser struct {
	rootPath string
}

// New creates a new Parser
func New(rootPath string) *Parser {
	return &Parser{
		rootPath: rootPath,
	}
}

// ParseFile extracts imports from a single file
func (p *Parser) ParseFile(filePath string) ([]Import, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return p.parseTypeScriptJavaScript(filePath)
	case ".go":
		return p.parseGo(filePath)
	case ".py":
		return p.parsePython(filePath)
	default:
		// Check if a plugin parser is registered for this extension
		if pluginParser, exists := plugin.GetRegistry().GetParser(ext); exists {
			pluginImports, err := pluginParser.ParseImports(filePath)
			if err != nil {
				return nil, err
			}

			// Convert plugin imports to internal format
			imports := make([]Import, len(pluginImports))
			for i, imp := range pluginImports {
				imports[i] = Import{
					SourceFile: filePath,
					ImportPath: imp.ImportPath,
					IsRelative: imp.IsRelative,
				}
			}
			return imports, nil
		}

		// Unsupported file type, return empty
		return []Import{}, nil
	}
}

// ParseExports extracts exports from a single file (Phase 2)
func (p *Parser) ParseExports(filePath string) ([]Export, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return p.parseTypeScriptJavaScriptExports(filePath)
	case ".go":
		return p.parseGoExports(filePath)
	case ".py":
		return p.parsePythonExports(filePath)
	default:
		// Check if a plugin parser is registered for this extension
		if pluginParser, exists := plugin.GetRegistry().GetParser(ext); exists {
			pluginExports, err := pluginParser.ParseExports(filePath)
			if err != nil {
				return nil, err
			}

			// Convert plugin exports to internal format
			var exports []Export
			for _, exp := range pluginExports {
				exports = append(exports, Export{
					SourceFile: filePath,
					Names:      []string{exp.Name},
					IsDefault:  exp.Kind == "default",
				})
			}
			return exports, nil
		}

		// Unsupported file type, return empty
		return []Export{}, nil
	}
}

// parseTypeScriptJavaScript extracts imports from TypeScript/JavaScript files
func (p *Parser) parseTypeScriptJavaScript(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Regex patterns for different import styles
	// import foo from 'bar'
	// import { foo } from 'bar'
	// import * as foo from 'bar'
	// import 'bar'
	// const foo = require('bar')
	importRegex := regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|import\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := importRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			// Extract the import path from the matched groups
			importPath := ""
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					importPath = match[i]
					break
				}
			}

			if importPath != "" {
				isRelative := strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/")
				imports = append(imports, Import{
					SourceFile: filePath,
					ImportPath: importPath,
					IsRelative: isRelative,
				})
			}
		}
	}

	return imports, scanner.Err()
}

// parseGo extracts imports from Go files
func (p *Parser) parseGo(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Track if we're in an import block
	inImportBlock := false

	// Single import: import "path"
	singleImportRegex := regexp.MustCompile(`^\s*import\s+"([^"]+)"`)
	// Import block start: import (
	importBlockStartRegex := regexp.MustCompile(`^\s*import\s+\(`)
	// Import in block: "path" or alias "path"
	importInBlockRegex := regexp.MustCompile(`^\s*(?:\w+\s+)?"([^"]+)"`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for single-line import
		if match := singleImportRegex.FindStringSubmatch(line); match != nil {
			imports = append(imports, Import{
				SourceFile: filePath,
				ImportPath: match[1],
				IsRelative: strings.HasPrefix(match[1], "."),
			})
			continue
		}

		// Check for import block start
		if importBlockStartRegex.MatchString(line) {
			inImportBlock = true
			continue
		}

		// Check for end of import block
		if inImportBlock && strings.TrimSpace(line) == ")" {
			inImportBlock = false
			continue
		}

		// Parse imports within block
		if inImportBlock {
			if match := importInBlockRegex.FindStringSubmatch(line); match != nil {
				imports = append(imports, Import{
					SourceFile: filePath,
					ImportPath: match[1],
					IsRelative: strings.HasPrefix(match[1], "."),
				})
			}
		}
	}

	return imports, scanner.Err()
}

// parsePython extracts imports from Python files
func (p *Parser) parsePython(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// from foo import bar
	// import foo
	// import foo.bar
	importRegex := regexp.MustCompile(`^\s*(?:from\s+([\w.]+)\s+import|import\s+([\w.]+))`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := importRegex.FindStringSubmatch(line); match != nil {
			importPath := ""
			if match[1] != "" {
				importPath = match[1]
			} else if match[2] != "" {
				importPath = match[2]
			}

			if importPath != "" {
				isRelative := strings.HasPrefix(importPath, ".")
				imports = append(imports, Import{
					SourceFile: filePath,
					ImportPath: importPath,
					IsRelative: isRelative,
				})
			}
		}
	}

	return imports, scanner.Err()
}

// ResolveImportPath resolves a relative import path to a path within the project
func (p *Parser) ResolveImportPath(sourceFile, importPath string) string {
	if !strings.HasPrefix(importPath, ".") {
		// Not a relative import, return as-is
		return importPath
	}

	// Get the directory of the source file
	sourceDir := filepath.Dir(sourceFile)

	// Resolve the import path relative to the source directory
	resolvedPath := filepath.Join(sourceDir, importPath)

	// Clean the path to resolve .. and .
	resolvedPath = filepath.Clean(resolvedPath)

	return resolvedPath
}
