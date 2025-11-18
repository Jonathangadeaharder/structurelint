"""
Clone search engine with threshold-based range search.

This module implements Part IV of the blueprint: The Clone-Finding Query Pipeline.
It provides threshold-based similarity search using FAISS range_search with proper
cosine-to-L2 conversion.
"""

import logging
from dataclasses import dataclass
from typing import List, Optional, Tuple

import faiss
import numpy as np

from clone_detection.embeddings.graphcodebert import GraphCodeBERTEmbedder
from clone_detection.indexing.faiss_index import cosine_to_l2_threshold, l2_to_cosine_similarity
from clone_detection.query.metadata import MetadataStore

logger = logging.getLogger(__name__)


@dataclass
class CloneMatch:
    """
    Represents a detected code clone.

    Attributes:
        snippet_id: ID of the matching snippet
        file_path: Path to the source file
        start_line: Starting line number
        end_line: Ending line number
        language: Programming language
        function_name: Name of the function (if available)
        similarity: Cosine similarity score (0-1)
        code: The actual code snippet
    """

    snippet_id: int
    file_path: str
    start_line: int
    end_line: int
    language: str
    function_name: Optional[str]
    similarity: float
    code: str

    def __repr__(self) -> str:
        return (
            f"CloneMatch(file={self.file_path}, "
            f"lines={self.start_line}-{self.end_line}, "
            f"similarity={self.similarity:.3f})"
        )

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        return {
            "snippet_id": self.snippet_id,
            "file_path": self.file_path,
            "start_line": self.start_line,
            "end_line": self.end_line,
            "language": self.language,
            "function_name": self.function_name,
            "similarity": self.similarity,
            "code": self.code,
        }


class CloneSearcher:
    """
    High-level interface for semantic clone detection.

    This class implements Blueprint B (Query Pipeline) from Section 4.4:
    1. Embed query code using GraphCodeBERT
    2. L2-normalize the query vector
    3. Convert cosine similarity to L2 threshold
    4. Run FAISS range_search
    5. Hydrate results with metadata

    Example:
        >>> searcher = CloneSearcher(index, embedder, metadata_store)
        >>> clones = searcher.find_clones(
        ...     query_code="def foo(): pass",
        ...     similarity_threshold=0.95
        ... )
    """

    def __init__(
        self,
        index: faiss.Index,
        embedder: GraphCodeBERTEmbedder,
        metadata_store: MetadataStore,
    ):
        """
        Initialize the clone searcher.

        Args:
            index: Trained and populated FAISS index
            embedder: GraphCodeBERT embedder (same model used for indexing)
            metadata_store: Database with code snippet metadata
        """
        self.index = index
        self.embedder = embedder
        self.metadata_store = metadata_store

        logger.info("Initialized CloneSearcher")
        logger.info(f"Index contains {self.index.ntotal} code snippets")

    def find_clones(
        self,
        query_code: str,
        similarity_threshold: float = 0.95,
        max_results: Optional[int] = None,
        exclude_self: bool = True,
    ) -> List[CloneMatch]:
        """
        Find all semantic clones of the given code.

        This implements the complete query pipeline from Section 4.4 (Blueprint B):
        1. Embed the query code
        2. Normalize the query vector
        3. Convert similarity threshold to L2 distance
        4. Execute range_search
        5. Retrieve and return metadata

        Args:
            query_code: Source code to find clones of
            similarity_threshold: Minimum cosine similarity (0-1)
            max_results: Maximum number of results to return
            exclude_self: Whether to exclude exact self-matches (sim=1.0)

        Returns:
            List of CloneMatch objects, sorted by similarity (descending)

        Example:
            >>> clones = searcher.find_clones(
            ...     "def add(a, b): return a + b",
            ...     similarity_threshold=0.95
            ... )
            >>> for clone in clones:
            ...     print(f"{clone.file_path}:{clone.start_line} - {clone.similarity:.3f}")
        """
        # Step 1: Generate embedding for query code
        query_embedding = self.embedder.embed_single(query_code)
        query_embedding = query_embedding.reshape(1, -1).astype(np.float32)

        # Step 2: L2-normalize the query vector (CRITICAL!)
        faiss.normalize_L2(query_embedding)

        # Step 3: Convert cosine similarity to L2 distance threshold (Table 4.1)
        l2_threshold = cosine_to_l2_threshold(similarity_threshold)

        logger.debug(
            f"Searching with cosine_similarity >= {similarity_threshold:.3f} "
            f"(L2 distance <= {l2_threshold:.3f})"
        )

        # Step 4: Execute range_search
        # Returns: lims, D (distances), I (IDs)
        lims, distances, ids = self.index.range_search(query_embedding, l2_threshold)

        # Extract results for the single query
        # lims[0]:lims[1] gives the range of results for query 0
        start_idx = lims[0]
        end_idx = lims[1]

        result_distances = distances[start_idx:end_idx]
        result_ids = ids[start_idx:end_idx]

        logger.info(f"Found {len(result_ids)} potential clones")

        # Step 5: Convert L2 distances back to cosine similarities
        similarities = [l2_to_cosine_similarity(d) for d in result_distances]

        # Step 6: Retrieve metadata for all matches
        clones = self._hydrate_results(result_ids, similarities, exclude_self)

        # Sort by similarity (descending)
        clones.sort(key=lambda x: x.similarity, reverse=True)

        # Apply max_results limit
        if max_results:
            clones = clones[:max_results]

        return clones

    def find_clones_by_location(
        self,
        file_path: str,
        line_number: int,
        similarity_threshold: float = 0.95,
        max_results: Optional[int] = None,
    ) -> List[CloneMatch]:
        """
        Find clones of a function at a specific file location.

        Args:
            file_path: Path to the source file
            line_number: Line number within the function
            similarity_threshold: Minimum cosine similarity
            max_results: Maximum number of results

        Returns:
            List of CloneMatch objects

        Example:
            >>> clones = searcher.find_clones_by_location(
            ...     "src/utils/helper.py",
            ...     line_number=42,
            ...     similarity_threshold=0.95
            ... )
        """
        # Look up the snippet in the metadata store
        snippet_metadata = self.metadata_store.get_snippet_by_location(file_path, line_number)

        if snippet_metadata is None:
            logger.warning(f"No snippet found at {file_path}:{line_number}")
            return []

        # Use the code from the metadata
        query_code = snippet_metadata["code"]

        logger.info(
            f"Found snippet: {snippet_metadata.get('function_name', 'unknown')} "
            f"at lines {snippet_metadata['start_line']}-{snippet_metadata['end_line']}"
        )

        # Find clones of this code
        return self.find_clones(
            query_code=query_code,
            similarity_threshold=similarity_threshold,
            max_results=max_results,
        )

    def find_clones_batch(
        self,
        query_codes: List[str],
        similarity_threshold: float = 0.95,
    ) -> List[List[CloneMatch]]:
        """
        Find clones for multiple queries in a single batch.

        Args:
            query_codes: List of source code strings
            similarity_threshold: Minimum cosine similarity

        Returns:
            List of lists (one per query) of CloneMatch objects
        """
        if not query_codes:
            return []

        # Generate embeddings for all queries
        query_embeddings = self.embedder.embed_batch(query_codes)
        query_embeddings = query_embeddings.astype(np.float32)

        # L2-normalize all queries
        faiss.normalize_L2(query_embeddings)

        # Convert threshold
        l2_threshold = cosine_to_l2_threshold(similarity_threshold)

        # Batch range search
        lims, distances, ids = self.index.range_search(query_embeddings, l2_threshold)

        # Process results for each query
        all_results = []
        for i in range(len(query_codes)):
            start_idx = lims[i]
            end_idx = lims[i + 1]

            result_distances = distances[start_idx:end_idx]
            result_ids = ids[start_idx:end_idx]

            similarities = [l2_to_cosine_similarity(d) for d in result_distances]
            clones = self._hydrate_results(result_ids, similarities)
            clones.sort(key=lambda x: x.similarity, reverse=True)

            all_results.append(clones)

        return all_results

    def _hydrate_results(
        self,
        snippet_ids: np.ndarray,
        similarities: List[float],
        exclude_self: bool = True,
    ) -> List[CloneMatch]:
        """
        Retrieve metadata for search results and create CloneMatch objects.

        This implements the "ResultHydration" step from Blueprint B.

        Args:
            snippet_ids: Array of snippet IDs from FAISS
            similarities: Corresponding similarity scores
            exclude_self: Whether to exclude exact matches (similarity = 1.0)

        Returns:
            List of CloneMatch objects
        """
        if len(snippet_ids) == 0:
            return []

        # Retrieve metadata from database
        metadata_list = self.metadata_store.get_snippets(snippet_ids.tolist())

        # Create a mapping from ID to metadata
        metadata_map = {m["id"]: m for m in metadata_list}

        # Build CloneMatch objects
        clones = []
        for snippet_id, similarity in zip(snippet_ids, similarities):
            # Skip if metadata not found
            if snippet_id not in metadata_map:
                logger.warning(f"Metadata not found for snippet ID: {snippet_id}")
                continue

            # Optionally exclude exact self-matches
            if exclude_self and similarity >= 0.9999:
                continue

            metadata = metadata_map[snippet_id]

            clone = CloneMatch(
                snippet_id=int(snippet_id),
                file_path=metadata["file_path"],
                start_line=metadata["start_line"],
                end_line=metadata["end_line"],
                language=metadata["language"],
                function_name=metadata["function_name"],
                similarity=similarity,
                code=metadata["code"],
            )
            clones.append(clone)

        return clones

    def get_statistics(self) -> dict:
        """Get statistics about the search index and metadata."""
        return {
            "index_size": self.index.ntotal,
            "metadata_count": self.metadata_store.count(),
            "languages": self.metadata_store.get_languages(),
            "embedding_dimension": self.embedder.get_embedding_dimension(),
        }
