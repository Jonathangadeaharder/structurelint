# syntactic

⬆️ **[Parent Directory](../README.md)**

## Overview

The `syntactic` package implements Strategy A: Syntactic Clone Detection using Rabin-Karp rolling hash, inverted indexing, and greedy match expansion.

## Components

### hasher.go
Rabin-Karp rolling hash for k-gram shingling.

**Algorithm**: O(N) complexity using rolling hash
**Hash Function**: FNV-1a (64-bit)
**Window Size**: Configurable k-gram (default: 20 tokens)

#### Key Functions

**`GenerateShingles(fileTokens *FileTokens) []Shingle`**
- Creates overlapping k-gram windows
- Applies rolling hash to each window
- Returns shingles with hash values

**`hashTokens(tokens []Token) uint64`**
- Computes FNV-1a hash for token sequence
- Separates tokens to avoid collisions

### index.go
Inverted index mapping hash values to locations.

**Data Structure**: `map[uint64][]Shingle`
**Thread Safety**: sync.RWMutex for concurrent access
**Operations**: O(1) add, O(1) lookup

#### Key Functions

**`Add(shingle Shingle)`**
- Adds shingle to index

**`FindCollisions() map[uint64][]Shingle`**
- Returns all hashes with multiple locations
- Identifies potential clone seeds

**`FindCrossFileCollisions() map[uint64][]Shingle`**
- Returns only cross-file collisions
- Filters within-file duplicates

### expander.go
Greedy bidirectional match expansion.

**Algorithm**: Expand seed matches forward and backward until mismatch
**Complexity**: O(M) where M = length of expanded clone

#### Key Functions

**`ExpandClone(shingle1, shingle2 Shingle) *Clone`**
- Verifies seed match (eliminate hash collisions)
- Expands backward until mismatch
- Expands forward until mismatch
- Returns full clone with accurate boundaries

**`ExpandAllCollisions(collisions map[uint64][]Shingle) []*Clone`**
- Processes all hash collisions
- Generates all clone pairs
- Deduplicates overlapping reports

## Algorithm Flow

```
1. Normalize tokens for all files
2. Generate k-gram shingles (rolling hash)
3. Build inverted index (hash → locations)
4. Find hash collisions (potential clones)
5. Verify and expand each collision
6. Filter by minimum size
7. Report clones
```

## Performance

- **Hashing**: ~200MB/s (lightweight)
- **Indexing**: O(1) per shingle
- **Expansion**: Depends on collision count
- **Memory**: ~50 bytes per shingle
