package syntactic

import (
	"strconv"

	"github.com/structurelint/structurelint/internal/clones/types"
)

// Expander implements greedy match expansion for clone detection
type Expander struct {
	tokenCache map[string][]types.Token // Cache of file -> tokens
}

// NewExpander creates a new expander instance
func NewExpander() *Expander {
	return &Expander{
		tokenCache: make(map[string][]types.Token),
	}
}

// SetTokenCache sets the token cache for file lookups
func (e *Expander) SetTokenCache(cache map[string][]types.Token) {
	e.tokenCache = cache
}

// ExpandClone takes two matching shingles and expands them into a full clone
func (e *Expander) ExpandClone(shingle1, shingle2 types.Shingle) *types.Clone {
	// Get token streams for both files
	tokens1, ok1 := e.tokenCache[shingle1.FilePath]
	tokens2, ok2 := e.tokenCache[shingle2.FilePath]

	if !ok1 || !ok2 {
		return nil
	}

	// Verify the seed match (hash collision could be spurious)
	if !e.verifySeedMatch(tokens1, tokens2, shingle1, shingle2) {
		return nil
	}

	// Expand backward from the seed
	startToken1, startToken2 := e.expandBackward(tokens1, tokens2, shingle1.StartToken, shingle2.StartToken)

	// Expand forward from the seed
	endToken1, endToken2 := e.expandForward(tokens1, tokens2, shingle1.EndToken, shingle2.EndToken)

	// Calculate token count
	tokenCount := endToken1 - startToken1 + 1

	// Determine clone type based on exact match
	cloneType := e.determineCloneType(tokens1[startToken1:endToken1+1], tokens2[startToken2:endToken2+1])

	// Create clone object
	clone := &types.Clone{
		Type:       cloneType,
		TokenCount: tokenCount,
		LineCount:  tokens1[endToken1].Line - tokens1[startToken1].Line + 1,
		Hash:       shingle1.Hash,
		Similarity: 1.0, // Syntactic clones are 100% similar after normalization
		Locations: []types.Location{
			{
				FilePath:   shingle1.FilePath,
				StartLine:  tokens1[startToken1].Line,
				EndLine:    tokens1[endToken1].Line,
				StartToken: startToken1,
				EndToken:   endToken1,
			},
			{
				FilePath:   shingle2.FilePath,
				StartLine:  tokens2[startToken2].Line,
				EndLine:    tokens2[endToken2].Line,
				StartToken: startToken2,
				EndToken:   endToken2,
			},
		},
	}

	return clone
}

// verifySeedMatch checks if the seed shingles actually match token-by-token
func (e *Expander) verifySeedMatch(tokens1, tokens2 []types.Token, shingle1, shingle2 types.Shingle) bool {
	// Check bounds
	if shingle1.EndToken >= len(tokens1) || shingle2.EndToken >= len(tokens2) {
		return false
	}

	// Verify each token matches
	seedLen := shingle1.EndToken - shingle1.StartToken + 1
	for i := 0; i < seedLen; i++ {
		tok1 := tokens1[shingle1.StartToken+i]
		tok2 := tokens2[shingle2.StartToken+i]

		if tok1.Value != tok2.Value {
			return false
		}
	}

	return true
}

// expandBackward expands the match backward until tokens differ
func (e *Expander) expandBackward(tokens1, tokens2 []types.Token, start1, start2 int) (int, int) {
	// Greedily match backward
	for start1 > 0 && start2 > 0 {
		if tokens1[start1-1].Value != tokens2[start2-1].Value {
			break
		}
		start1--
		start2--
	}

	return start1, start2
}

// expandForward expands the match forward until tokens differ
func (e *Expander) expandForward(tokens1, tokens2 []types.Token, end1, end2 int) (int, int) {
	// Greedily match forward
	maxLen1 := len(tokens1)
	maxLen2 := len(tokens2)

	for end1+1 < maxLen1 && end2+1 < maxLen2 {
		if tokens1[end1+1].Value != tokens2[end2+1].Value {
			break
		}
		end1++
		end2++
	}

	return end1, end2
}

// determineCloneType classifies the clone as Type-1, Type-2, or Type-3
func (e *Expander) determineCloneType(tokens1, tokens2 []types.Token) types.CloneType {
	// After normalization, all identifier/literal differences are abstracted
	// If the normalized tokens match exactly, it could be Type-1 or Type-2
	// For POC, we classify all syntactic matches as Type-2 (conservative)
	// To distinguish Type-1 from Type-2, we'd need to check pre-normalization

	// Type-3 detection would require fuzzy matching (handled by min-token threshold)
	// For now, all expanded clones are classified as Type-2
	return types.Type2
}

// ExpandAllCollisions processes all hash collisions and returns unique clones
func (e *Expander) ExpandAllCollisions(collisions map[uint64][]types.Shingle) []*types.Clone {
	var clones []*types.Clone
	processed := make(map[string]bool) // Track processed clone pairs to avoid duplicates

	for _, shingles := range collisions {
		// Generate all pairs of shingles with the same hash
		for i := 0; i < len(shingles); i++ {
			for j := i + 1; j < len(shingles); j++ {
				// Create unique key for this pair
				key := e.clonePairKey(shingles[i], shingles[j])
				if processed[key] {
					continue
				}

				// Expand the clone
				clone := e.ExpandClone(shingles[i], shingles[j])
				if clone != nil {
					clones = append(clones, clone)
					processed[key] = true
				}
			}
		}
	}

	return clones
}

// clonePairKey generates a unique key for a clone pair
func (e *Expander) clonePairKey(s1, s2 types.Shingle) string {
	// Ensure consistent ordering (smaller file path first)
	if s1.FilePath < s2.FilePath {
		return s1.FilePath + ":" + strconv.Itoa(s1.StartToken) + "-" + s2.FilePath + ":" + strconv.Itoa(s2.StartToken)
	}
	return s2.FilePath + ":" + strconv.Itoa(s2.StartToken) + "-" + s1.FilePath + ":" + strconv.Itoa(s1.StartToken)
}
