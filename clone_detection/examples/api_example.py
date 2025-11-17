"""
Example: Using the Python API for semantic clone detection.

This example demonstrates how to use the clone detection system programmatically
without the CLI.
"""

from clone_detection.parsers import TreeSitterParser
from clone_detection.embeddings import GraphCodeBERTEmbedder
from clone_detection.indexing import FAISSIndexBuilder, IndexType
from clone_detection.query import CloneSearcher, MetadataStore


def main():
    print("=" * 70)
    print("Semantic Clone Detection - Python API Example")
    print("=" * 70)
    print()

    # Step 1: Parse code snippets
    print("Step 1: Parsing code with Tree-sitter...")
    parser = TreeSitterParser(languages=["python"])

    # Example: Parse a single file or directory
    snippets = parser.parse_file("example_code.py")
    # Or: snippets = parser.parse_directory("/path/to/codebase")

    print(f"  Found {len(snippets)} code snippets")
    print()

    # Step 2: Generate embeddings
    print("Step 2: Generating GraphCodeBERT embeddings...")
    embedder = GraphCodeBERTEmbedder(
        model_name="microsoft/graphcodebert-base",
        device="cpu",  # Use "cuda" for GPU
        batch_size=8,
    )

    embeddings = embedder.embed_batch(snippets)
    print(f"  Generated embeddings: shape={embeddings.shape}")
    print()

    # Step 3: Build FAISS index
    print("Step 3: Building FAISS index...")
    index_builder = FAISSIndexBuilder(
        dimension=768,
        index_type=IndexType.IVF_PQ,
        nlist=256,  # Smaller for this example
        nprobe=8,
    )

    # Generate IDs for snippets
    import numpy as np

    snippet_ids = np.arange(len(snippets), dtype=np.int64)

    # Build the index (train + add)
    index = index_builder.build(vectors=embeddings, ids=snippet_ids)
    print(f"  Built index with {index.ntotal} vectors")
    print()

    # Step 4: Save index and metadata
    print("Step 4: Saving index and metadata...")
    index_builder.save("example_clones.index")

    # Save metadata
    metadata_store = MetadataStore("example_clones.db")
    for i, snippet in enumerate(snippets):
        metadata_store.add_snippet(i, snippet)
    print("  Saved to example_clones.index and example_clones.db")
    print()

    # Step 5: Query for clones
    print("Step 5: Searching for clones...")

    # Reload for demonstration (in practice, you'd do this in a separate session)
    index_builder_loaded = FAISSIndexBuilder.load("example_clones.index")
    searcher = CloneSearcher(
        index=index_builder_loaded.index, embedder=embedder, metadata_store=metadata_store
    )

    # Example query
    query_code = """
def calculate_sum(numbers):
    total = 0
    for num in numbers:
        total += num
    return total
"""

    clones = searcher.find_clones(
        query_code=query_code, similarity_threshold=0.80, max_results=10
    )

    print(f"  Found {len(clones)} potential clones")
    for i, clone in enumerate(clones[:5], 1):  # Show top 5
        print(f"    {i}. {clone.file_path}:{clone.start_line} (sim={clone.similarity:.3f})")

    print()
    print("=" * 70)
    print("Example complete!")
    print("=" * 70)

    metadata_store.close()


if __name__ == "__main__":
    # Note: This is a minimal example. In practice, you'd have actual code files to parse.
    # For this example to run, create an 'example_code.py' file with some Python functions.

    # Create a sample file for demonstration
    sample_code = '''
def add(a, b):
    """Add two numbers."""
    return a + b

def subtract(x, y):
    """Subtract y from x."""
    return x - y

def multiply(m, n):
    """Multiply two numbers."""
    result = m * n
    return result

def calculate_sum(numbers):
    """Sum a list of numbers."""
    total = 0
    for num in numbers:
        total += num
    return total

def compute_total(values):
    """Another way to sum numbers."""
    sum_val = 0
    for v in values:
        sum_val = sum_val + v
    return sum_val
'''

    with open("example_code.py", "w") as f:
        f.write(sample_code)

    try:
        main()
    finally:
        # Cleanup
        import os

        for file in ["example_code.py", "example_clones.index", "example_clones.db"]:
            if os.path.exists(file):
                os.remove(file)
