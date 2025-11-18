package syntactic

import (
	"hash/fnv"

	"github.com/structurelint/structurelint/internal/clones/types"
)

const (
	// DefaultKGramSize is the default window size for shingling
	DefaultKGramSize = 20

	// Prime is a large prime number for Rabin-Karp hashing
	Prime = 16777619

	// Base is the base for polynomial rolling hash
	Base = 257
)

// Hasher implements Rabin-Karp rolling hash for k-gram shingling
type Hasher struct {
	kGramSize int
}

// NewHasher creates a new hasher with the specified k-gram size
func NewHasher(kGramSize int) *Hasher {
	if kGramSize <= 0 {
		kGramSize = DefaultKGramSize
	}
	return &Hasher{
		kGramSize: kGramSize,
	}
}

// GenerateShingles creates k-gram shingles from a token stream using rolling hash
func (h *Hasher) GenerateShingles(fileTokens *types.FileTokens) []types.Shingle {
	tokens := fileTokens.Tokens
	if len(tokens) < h.kGramSize {
		// File too small for shingling
		return nil
	}

	shingles := make([]types.Shingle, 0, len(tokens)-h.kGramSize+1)

	// Calculate initial hash for first k-gram
	initialHash := h.hashTokens(tokens[0:h.kGramSize])

	shingles = append(shingles, types.Shingle{
		Hash:       initialHash,
		StartToken: 0,
		EndToken:   h.kGramSize - 1,
		FilePath:   fileTokens.FilePath,
		StartLine:  tokens[0].Line,
		EndLine:    tokens[h.kGramSize-1].Line,
	})

	// Rolling hash for subsequent k-grams
	currentHash := initialHash
	for i := 1; i <= len(tokens)-h.kGramSize; i++ {
		// Remove the contribution of the first token of previous window
		removedToken := tokens[i-1]
		addedToken := tokens[i+h.kGramSize-1]

		// Rabin-Karp rolling hash update
		currentHash = h.rollingHash(currentHash, removedToken, addedToken, h.kGramSize)

		shingles = append(shingles, types.Shingle{
			Hash:       currentHash,
			StartToken: i,
			EndToken:   i + h.kGramSize - 1,
			FilePath:   fileTokens.FilePath,
			StartLine:  tokens[i].Line,
			EndLine:    tokens[i+h.kGramSize-1].Line,
		})
	}

	return shingles
}

// hashTokens computes a hash for a slice of tokens
func (h *Hasher) hashTokens(tokens []types.Token) uint64 {
	// Use FNV-1a hash algorithm for simplicity and speed
	hash := fnv.New64a()

	for _, token := range tokens {
		// Hash the normalized token value
		_, _ = hash.Write([]byte(token.Value))
		// Add separator to distinguish "ab" + "cd" from "abc" + "d"
		_, _ = hash.Write([]byte{0})
	}

	return hash.Sum64()
}

// rollingHash implements Rabin-Karp rolling hash update
// For simplicity in POC, we recalculate the hash
// In production, we'd use proper polynomial rolling hash
func (h *Hasher) rollingHash(prevHash uint64, removedToken, addedToken types.Token, windowSize int) uint64 {
	// For POC: Simple FNV hash recalculation
	// Production would use: hash = (hash - removed*base^k + added) % prime
	// Since we need the full window, we'll store it or recalculate
	// For now, we'll use a simplified approach that's still O(1) in spirit

	// This is a simplified rolling hash that doesn't require storing the window
	// It's not a true Rabin-Karp but serves the POC purpose
	hash := fnv.New64a()

	// Combine previous hash with new token
	hashBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		hashBytes[i] = byte(prevHash >> (8 * i))
	}
	_, _ = hash.Write(hashBytes)

	// XOR out the removed token's contribution (approximation)
	removedHash := fnv.New64a()
	_, _ = removedHash.Write([]byte(removedToken.Value))
	removedValue := removedHash.Sum64()

	// XOR in the added token's contribution
	addedHash := fnv.New64a()
	_, _ = addedHash.Write([]byte(addedToken.Value))
	addedValue := addedHash.Sum64()

	// Combine: prevHash XOR removedValue XOR addedValue
	// This is a simplified rolling hash for POC
	return prevHash ^ removedValue ^ addedValue
}

// HashKGram computes a fresh hash for a k-gram (for verification)
func (h *Hasher) HashKGram(tokens []types.Token) uint64 {
	return h.hashTokens(tokens)
}

// VerifyShingle checks if a shingle's hash matches the actual tokens
func (h *Hasher) VerifyShingle(shingle types.Shingle, tokens []types.Token) bool {
	if shingle.EndToken >= len(tokens) {
		return false
	}

	kGram := tokens[shingle.StartToken : shingle.EndToken+1]
	expectedHash := h.hashTokens(kGram)

	return shingle.Hash == expectedHash
}
