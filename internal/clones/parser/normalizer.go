package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/structurelint/structurelint/internal/clones/types"
)

// Normalizer converts source code into normalized token streams
type Normalizer struct {
	fset *token.FileSet
}

// NewNormalizer creates a new normalizer instance
func NewNormalizer() *Normalizer {
	return &Normalizer{
		fset: token.NewFileSet(),
	}
}

// NormalizeFile parses a Go source file and returns a normalized token stream
func (n *Normalizer) NormalizeFile(filePath string) (*types.FileTokens, error) {
	// Read the source file
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the file into an AST
	file, err := parser.ParseFile(n.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract and normalize tokens
	tokens := n.extractTokens(file, src)

	return &types.FileTokens{
		FilePath: filePath,
		Tokens:   tokens,
	}, nil
}

// extractTokens traverses the AST and extracts normalized tokens
func (n *Normalizer) extractTokens(file *ast.File, src []byte) []types.Token {
	var tokens []types.Token
	position := 0

	// Visitor function to traverse the AST
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		pos := n.fset.Position(node.Pos())

		switch n := node.(type) {
		case *ast.Ident:
			// Normalize identifiers to _ID_ (except keywords)
			if token.Lookup(n.Name).IsKeyword() {
				// It's a keyword, keep as-is
				tokens = append(tokens, types.Token{
					Type:     types.TokenKeyword,
					Value:    n.Name,
					Line:     pos.Line,
					Column:   pos.Column,
					Position: position,
				})
			} else {
				// It's an identifier, normalize
				tokens = append(tokens, types.Token{
					Type:     types.TokenIdentifier,
					Value:    "_ID_",
					Line:     pos.Line,
					Column:   pos.Column,
					Position: position,
				})
			}
			position++

		case *ast.BasicLit:
			// Normalize literals (strings, numbers, etc.) to _LIT_
			tokens = append(tokens, types.Token{
				Type:     types.TokenLiteral,
				Value:    "_LIT_",
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.BinaryExpr:
			// Add operator tokens
			tokens = append(tokens, types.Token{
				Type:     types.TokenOperator,
				Value:    n.Op.String(),
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.UnaryExpr:
			// Add unary operator tokens
			tokens = append(tokens, types.Token{
				Type:     types.TokenOperator,
				Value:    n.Op.String(),
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.AssignStmt:
			// Add assignment operator
			tokens = append(tokens, types.Token{
				Type:     types.TokenOperator,
				Value:    n.Tok.String(),
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.IfStmt:
			// Add 'if' keyword
			tokens = append(tokens, types.Token{
				Type:     types.TokenKeyword,
				Value:    "if",
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.ForStmt, *ast.RangeStmt:
			// Add 'for' keyword
			tokens = append(tokens, types.Token{
				Type:     types.TokenKeyword,
				Value:    "for",
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.ReturnStmt:
			// Add 'return' keyword
			tokens = append(tokens, types.Token{
				Type:     types.TokenKeyword,
				Value:    "return",
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++

		case *ast.FuncDecl:
			// Add 'func' keyword
			tokens = append(tokens, types.Token{
				Type:     types.TokenKeyword,
				Value:    "func",
				Line:     pos.Line,
				Column:   pos.Column,
				Position: position,
			})
			position++
		}

		return true
	})

	return tokens
}

// TokenStreamToString converts a token stream to a string for debugging
func TokenStreamToString(tokens []types.Token) string {
	var builder strings.Builder
	for _, tok := range tokens {
		builder.WriteString(tok.Value)
		builder.WriteString(" ")
	}
	return builder.String()
}
