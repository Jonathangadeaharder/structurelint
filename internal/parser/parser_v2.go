//go:build cgo

package parser

import (
	"path/filepath"

	"github.com/Jonathangadeaharder/structurelint/internal/parser/treesitter"
)

// ParserV2 is the new tree-sitter based parser
type ParserV2 struct {
	rootPath string
}

// NewV2 creates a new tree-sitter based parser
func NewV2(rootPath string) *ParserV2 {
	return &ParserV2{
		rootPath: rootPath,
	}
}

// ParseFile extracts imports from a single file using tree-sitter
func (p *ParserV2) ParseFile(filePath string) ([]Import, error) {
	ext := filepath.Ext(filePath)

	// Detect language from extension
	lang, err := treesitter.DetectLanguageFromExtension(ext)
	if err != nil {
		// Fall back to empty for unsupported languages
		return []Import{}, nil
	}

	// Create extractor for this language
	extractor, err := treesitter.NewImportExtractor(lang)
	if err != nil {
		return nil, err
	}

	// Extract imports
	tsImports, err := extractor.ExtractFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert to our Import type
	imports := make([]Import, len(tsImports))
	for i, tsImp := range tsImports {
		imports[i] = Import{
			SourceFile:    tsImp.SourceFile,
			ImportPath:    tsImp.ImportPath,
			IsRelative:    tsImp.IsRelative,
			ImportedNames: tsImp.ImportedNames,
		}
	}

	return imports, nil
}

// ParseExports extracts exports from a single file using tree-sitter
func (p *ParserV2) ParseExports(filePath string) ([]Export, error) {
	ext := filepath.Ext(filePath)

	// Detect language from extension
	lang, err := treesitter.DetectLanguageFromExtension(ext)
	if err != nil {
		// Fall back to empty for unsupported languages
		return []Export{}, nil
	}

	// Create extractor for this language
	extractor, err := treesitter.NewExportExtractor(lang)
	if err != nil {
		return nil, err
	}

	// Extract exports
	tsExports, err := extractor.ExtractFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert to our Export type
	exports := make([]Export, len(tsExports))
	for i, tsExp := range tsExports {
		exports[i] = Export{
			SourceFile: tsExp.SourceFile,
			Names:      tsExp.Names,
			IsDefault:  tsExp.IsDefault,
		}
	}

	return exports, nil
}

// ResolveImportPath resolves a relative import path to a path within the project
func (p *ParserV2) ResolveImportPath(sourceFile, importPath string) string {
	return treesitter.ResolveImportPath(sourceFile, importPath)
}
