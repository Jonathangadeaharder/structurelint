package syntactic

import (
	"sync"

	"github.com/structurelint/structurelint/internal/clones/types"
)

// Index is an inverted index mapping hash values to shingle locations
type Index struct {
	index map[uint64][]types.Shingle
	mu    sync.RWMutex
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		index: make(map[uint64][]types.Shingle),
	}
}

// Add adds a shingle to the index
func (idx *Index) Add(shingle types.Shingle) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.index[shingle.Hash] = append(idx.index[shingle.Hash], shingle)
}

// AddBatch adds multiple shingles to the index efficiently
func (idx *Index) AddBatch(shingles []types.Shingle) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, shingle := range shingles {
		idx.index[shingle.Hash] = append(idx.index[shingle.Hash], shingle)
	}
}

// GetCandidates returns all shingles with the given hash
func (idx *Index) GetCandidates(hash uint64) []types.Shingle {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	shingles := idx.index[hash]
	if shingles == nil {
		return nil
	}
	
	// Return a copy to prevent concurrent modification
	result := make([]types.Shingle, len(shingles))
	copy(result, shingles)
	return result
}

// FindCollisions returns all hash values that have multiple locations (potential clones)
func (idx *Index) FindCollisions() map[uint64][]types.Shingle {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	collisions := make(map[uint64][]types.Shingle)

	for hash, shingles := range idx.index {
		if len(shingles) > 1 {
			// Copy the slice to prevent concurrent modification
			shinglesCopy := make([]types.Shingle, len(shingles))
			copy(shinglesCopy, shingles)
			collisions[hash] = shinglesCopy
		}
	}

	return collisions
}

// FindCrossFileCollisions returns collisions that span multiple files
func (idx *Index) FindCrossFileCollisions() map[uint64][]types.Shingle {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	crossFileCollisions := make(map[uint64][]types.Shingle)

	for hash, shingles := range idx.index {
		if len(shingles) < 2 {
			continue
		}

		// Check if shingles are from different files
		fileSet := make(map[string]bool)
		for _, shingle := range shingles {
			fileSet[shingle.FilePath] = true
		}

		// Only include if multiple files are involved
		if len(fileSet) > 1 {
			// Copy the slice to prevent concurrent modification
			shinglesCopy := make([]types.Shingle, len(shingles))
			copy(shinglesCopy, shingles)
			crossFileCollisions[hash] = shinglesCopy
		}
	}

	return crossFileCollisions
}

// Stats returns statistics about the index
func (idx *Index) Stats() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	totalShingles := 0
	collisionCount := 0

	for _, shingles := range idx.index {
		totalShingles += len(shingles)
		if len(shingles) > 1 {
			collisionCount++
		}
	}

	return IndexStats{
		TotalHashes:    len(idx.index),
		TotalShingles:  totalShingles,
		CollisionCount: collisionCount,
	}
}

// IndexStats represents statistics about the index
type IndexStats struct {
	TotalHashes    int // Number of unique hash values
	TotalShingles  int // Total number of shingles indexed
	CollisionCount int // Number of hashes with multiple shingles
}

// Clear removes all entries from the index
func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.index = make(map[uint64][]types.Shingle)
}

// Size returns the number of unique hashes in the index
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.index)
}
