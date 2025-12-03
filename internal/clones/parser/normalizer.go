package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
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
		token := n.processNode(node, pos, position)
		
		if token != nil {
			tokens = append(tokens, *token)
			position++
		}

		return true
	})

	return tokens
}

func (n *Normalizer) processNode(node ast.Node, pos token.Position, position int) *types.Token {
	switch node := node.(type) {
	case *ast.Ident:
		return n.processIdent(node, pos, position)
	case *ast.BasicLit:
		return &types.Token{
			Type:     types.TokenLiteral,
			Value:    "_LIT_",
			Line:     pos.Line,
			Column:   pos.Column,
			Position: position,
		}
	case *ast.BinaryExpr:
		return &types.Token{
			Type:     types.TokenOperator,
			Value:    node.Op.String(),
			Line:     pos.Line,
			Column:   pos.Column,
			Position: position,
		}
	case *ast.UnaryExpr:
		return &types.Token{
			Type:     types.TokenOperator,
			Value:    node.Op.String(),
			Line:     pos.Line,
			Column:   pos.Column,
			Position: position,
		}
	case *ast.AssignStmt:
		return &types.Token{
			Type:     types.TokenOperator,
			Value:    node.Tok.String(),
			Line:     pos.Line,
			Column:   pos.Column,
			Position: position,
		}
	default:
		return n.processKeywords(node, pos, position)
	}
}

func (n *Normalizer) processIdent(ident *ast.Ident, pos token.Position, position int) *types.Token {
	if token.Lookup(ident.Name).IsKeyword() {
		return &types.Token{
			Type:     types.TokenKeyword,
			Value:    ident.Name,
			Line:     pos.Line,
			Column:   pos.Column,
			Position: position,
		}
	}
	return &types.Token{
		Type:     types.TokenIdentifier,
		Value:    "_ID_",
		Line:     pos.Line,
		Column:   pos.Column,
		Position: position,
	}
}

func (n *Normalizer) processKeywords(node ast.Node, pos token.Position, position int) *types.Token {
	var value string
	
	switch node.(type) {
	case *ast.IfStmt:
		value = "if"
	case *ast.ForStmt, *ast.RangeStmt:
		value = "for"
	case *ast.ReturnStmt:
		value = "return"
	case *ast.FuncDecl:
		value = "func"
	default:
		return nil
	}

	return &types.Token{
		Type:     types.TokenKeyword,
		Value:    value,
		Line:     pos.Line,
		Column:   pos.Column,
		Position: position,
	}
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
