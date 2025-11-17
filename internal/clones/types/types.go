package types

// Token represents a normalized token from source code
type Token struct {
	Type     TokenType // Type of token (keyword, identifier, literal, operator)
	Value    string    // Normalized value (_ID_, _LIT_, or raw keyword/operator)
	Line     int       // Line number in source file
	Column   int       // Column number in source file
	Position int       // Position in token stream (for indexing)
}

// TokenType represents the category of a token
type TokenType int

const (
	TokenKeyword    TokenType = iota // Keywords: if, for, func, etc.
	TokenIdentifier                  // Variable/function names (normalized to _ID_)
	TokenLiteral                     // String/number literals (normalized to _LIT_)
	TokenOperator                    // Operators: +, -, =, ==, etc.
	TokenPunctuation                 // Punctuation: (, ), {, }, ;, etc.
)

// Location represents a position in a source file
type Location struct {
	FilePath   string // Path to the file
	StartLine  int    // Starting line number (1-indexed)
	EndLine    int    // Ending line number (1-indexed)
	StartToken int    // Starting token index in normalized stream
	EndToken   int    // Ending token index in normalized stream
}

// Clone represents a detected code clone
type Clone struct {
	Type       CloneType   // Type of clone (Type-1, Type-2, Type-3)
	Locations  []Location  // All locations of this clone
	TokenCount int         // Number of tokens in the clone
	LineCount  int         // Approximate number of lines
	Hash       uint64      // Hash value (for syntactic clones)
	Similarity float64     // Similarity score (for semantic clones, 0.0-1.0)
}

// CloneType represents the classification of a code clone
type CloneType int

const (
	Type1 CloneType = iota // Exact copy-paste (whitespace/comments differ)
	Type2                  // Renamed identifiers/literals
	Type3                  // Modified statements (additions/deletions)
	Type4                  // Semantic equivalence (different implementation)
)

func (ct CloneType) String() string {
	switch ct {
	case Type1:
		return "Type-1 (exact copy)"
	case Type2:
		return "Type-2 (renamed)"
	case Type3:
		return "Type-3 (modified)"
	case Type4:
		return "Type-4 (semantic)"
	default:
		return "Unknown"
	}
}

// Shingle represents a k-gram window with its hash
type Shingle struct {
	Hash       uint64   // Rabin-Karp rolling hash
	StartToken int      // Starting position in token stream
	EndToken   int      // Ending position in token stream
	FilePath   string   // Source file path
	StartLine  int      // Starting line number
	EndLine    int      // Ending line number
}

// ClonePair represents a pair of clone locations for reporting
type ClonePair struct {
	LocationA  Location
	LocationB  Location
	Clone      *Clone
	Confidence float64 // Confidence score (1.0 = exact match)
}

// FileTokens represents a file's normalized token stream
type FileTokens struct {
	FilePath string
	Tokens   []Token
}
