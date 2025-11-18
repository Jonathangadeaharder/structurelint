// Package treesitter provides tree-sitter based parsing for multiple languages
package treesitter

import (
	"fmt"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Language represents a supported programming language
type Language string

const (
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageTypeScript Language = "typescript"
	LanguageJava       Language = "java"
)

// Parser wraps tree-sitter parsing functionality
type Parser struct {
	parser     *sitter.Parser
	language   Language
	sitterLang *sitter.Language
}

// New creates a new Parser for the specified language
func New(lang Language) (*Parser, error) {
	parser := sitter.NewParser()

	var treeSitterLang *sitter.Language
	switch lang {
	case LanguageGo:
		treeSitterLang = golang.GetLanguage()
	case LanguagePython:
		treeSitterLang = python.GetLanguage()
	case LanguageJavaScript:
		treeSitterLang = javascript.GetLanguage()
	case LanguageTypeScript:
		treeSitterLang = typescript.GetLanguage()
	case LanguageJava:
		treeSitterLang = java.GetLanguage()
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	parser.SetLanguage(treeSitterLang)

	return &Parser{
		parser:     parser,
		language:   lang,
		sitterLang: treeSitterLang,
	}, nil
}

// Parse parses the source code and returns a syntax tree
func (p *Parser) Parse(sourceCode []byte) (*sitter.Tree, error) {
	tree, err := p.parser.ParseCtx(nil, nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return tree, nil
}

// ParseFile parses a source file and returns a syntax tree
func (p *Parser) ParseFile(filePath string) (*sitter.Tree, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.Parse(content)
}

// Query executes a tree-sitter query on the given tree
func (p *Parser) Query(tree *sitter.Tree, queryString string) (*sitter.QueryCursor, *sitter.Query, error) {
	query, err := sitter.NewQuery([]byte(queryString), p.sitterLang)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create query: %w", err)
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, tree.RootNode())

	return cursor, query, nil
}

// DetectLanguageFromExtension returns the Language based on file extension
func DetectLanguageFromExtension(ext string) (Language, error) {
	switch ext {
	case ".go":
		return LanguageGo, nil
	case ".py":
		return LanguagePython, nil
	case ".js", ".jsx", ".mjs":
		return LanguageJavaScript, nil
	case ".ts", ".tsx":
		return LanguageTypeScript, nil
	case ".java":
		return LanguageJava, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}
