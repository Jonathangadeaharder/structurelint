//go:build !cgo

package parser

import (
	"fmt"
)

// ParserV2 is the new tree-sitter based parser (Stub implementation for non-CGO builds)
type ParserV2 struct {
	rootPath string
	v1       *Parser // Fallback to V1
}

// NewV2 creates a new parser. Without CGO, this falls back to V1 behavior or returns empty results for unsupported languages.
func NewV2(rootPath string) *ParserV2 {
	return &ParserV2{
		rootPath: rootPath,
		v1:       New(rootPath),
	}
}

// ParseFile extracts imports from a single file
func (p *ParserV2) ParseFile(filePath string) ([]Import, error) {
	// Fallback to V1 regex parser
	return p.v1.ParseFile(filePath)
}

// ParseExports extracts exports from a single file
func (p *ParserV2) ParseExports(filePath string) ([]Export, error) {
	// Fallback to V1 regex parser
	return p.v1.ParseExports(filePath)
}

// ResolveImportPath resolves a relative import path to a path within the project
func (p *ParserV2) ResolveImportPath(sourceFile, importPath string) string {
	return p.v1.ResolveImportPath(sourceFile, importPath)
}
