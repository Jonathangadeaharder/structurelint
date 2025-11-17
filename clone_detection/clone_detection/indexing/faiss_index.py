"""
FAISS index builder for high-performance vector similarity search.

This module implements Part III of the blueprint: High-Scale Vector Indexing.
It provides IndexIVFPQ with proper L2 normalization to ensure mathematically
correct cosine similarity search using L2 distance metrics.

Critical Implementation Notes:
1. All vectors MUST be L2-normalized before adding to the index
2. This ensures L2 distance is equivalent to cosine similarity
3. Mathematical proof: D_L2² = 2 - 2·cos_sim (for normalized vectors)
"""

import logging
from enum import Enum
from pathlib import Path
from typing import Optional, Tuple

import faiss
import numpy as np

logger = logging.getLogger(__name__)


class IndexType(Enum):
    """
    FAISS index types as defined in Table 3.1 of the blueprint.
    """

    FLAT = "Flat"  # Brute-force exact search (baseline)
    IVF_FLAT = "IVF,Flat"  # IVF with full-precision vectors
    IVF_PQ = "IVF,PQ"  # IVF with Product Quantization (production)


class FAISSIndexBuilder:
    """
    Builder for FAISS indexes with proper configuration for code clone detection.

    This class implements the index architecture from Section 3.3, including:
    - IndexIVFPQ for production scale
    - L2 normalization for cosine similarity equivalence
    - Training and population phases
    - Query-time parameter tuning (nprobe)

    Example:
        >>> builder = FAISSIndexBuilder(dimension=768, nlist=4096)
        >>> builder.train(training_vectors)
        >>> builder.add(all_vectors, all_ids)
        >>> builder.save("clones.index")
    """

    def __init__(
        self,
        dimension: int = 768,
        index_type: IndexType = IndexType.IVF_PQ,
        nlist: int = 4096,
        m: int = 64,
        nbits: int = 8,
        nprobe: int = 16,
        use_gpu: bool = False,
    ):
        """
        Initialize the index builder.

        Args:
            dimension: Vector dimension (768 for GraphCodeBERT)
            index_type: Type of FAISS index to build
            nlist: Number of IVF clusters (Table 3.2: 4*sqrt(N) to 16*sqrt(N))
            m: Number of PQ sub-vectors (must divide dimension evenly)
            nbits: Bits per PQ code (typically 8 for 256 centroids)
            nprobe: Number of clusters to probe at query time (trade-off parameter)
            use_gpu: Whether to use GPU acceleration

        Parameter Guidelines (Table 3.2):
            - nlist: For 10M vectors, use 4096-65536
            - m: Must divide 768 (e.g., 32, 64, 96)
            - nbits: 8 is standard (2^8 = 256 centroids)
            - nprobe: 16-32 for good accuracy/speed balance
        """
        self.dimension = dimension
        self.index_type = index_type
        self.nlist = nlist
        self.m = m
        self.nbits = nbits
        self.nprobe = nprobe
        self.use_gpu = use_gpu

        self.index: Optional[faiss.Index] = None
        self.is_trained = False

        # Validate parameters
        if dimension % m != 0:
            raise ValueError(
                f"m ({m}) must evenly divide dimension ({dimension}). "
                f"Valid values for d=768: 32, 64, 96, 128, 192, 256, 384, 768"
            )

        logger.info(f"Initialized FAISS index builder: {self._get_index_description()}")

    def _get_index_description(self) -> str:
        """Get a human-readable description of the index configuration."""
        if self.index_type == IndexType.IVF_PQ:
            return (
                f"IndexIVFPQ(d={self.dimension}, nlist={self.nlist}, "
                f"m={self.m}, nbits={self.nbits}, nprobe={self.nprobe})"
            )
        elif self.index_type == IndexType.IVF_FLAT:
            return f"IndexIVFFlat(d={self.dimension}, nlist={self.nlist}, nprobe={self.nprobe})"
        else:
            return f"IndexFlatL2(d={self.dimension})"

    def build(
        self,
        vectors: Optional[np.ndarray] = None,
        ids: Optional[np.ndarray] = None,
        train_vectors: Optional[np.ndarray] = None,
    ) -> faiss.Index:
        """
        Build the FAISS index (train + add data).

        This is a convenience method that combines training and population.
        For large-scale applications, call train() and add() separately.

        Args:
            vectors: All vectors to index (N, 768)
            ids: Corresponding IDs for each vector (N,)
            train_vectors: Optional separate training set. If None, uses vectors.

        Returns:
            Trained and populated FAISS index

        Example:
            >>> builder = FAISSIndexBuilder()
            >>> index = builder.build(embeddings, snippet_ids)
        """
        # Create the index
        self._create_index()

        # Train
        train_data = train_vectors if train_vectors is not None else vectors
        if train_data is not None:
            self.train(train_data)

        # Add vectors
        if vectors is not None:
            self.add(vectors, ids)

        return self.index

    def _create_index(self) -> None:
        """
        Create the FAISS index structure.

        Implements Step 1 from Section 3.3: Instantiate the Index.
        """
        if self.index_type == IndexType.FLAT:
            # Brute-force exact search
            self.index = faiss.IndexFlatL2(self.dimension)
            self.is_trained = True  # Flat index doesn't require training

        elif self.index_type == IndexType.IVF_FLAT:
            # IVF with full-precision vectors
            quantizer = faiss.IndexFlatL2(self.dimension)
            self.index = faiss.IndexIVFFlat(quantizer, self.dimension, self.nlist)

        elif self.index_type == IndexType.IVF_PQ:
            # IVF with Product Quantization (production)
            quantizer = faiss.IndexFlatL2(self.dimension)
            self.index = faiss.IndexIVFPQ(
                quantizer, self.dimension, self.nlist, self.m, self.nbits
            )

        else:
            raise ValueError(f"Unknown index type: {self.index_type}")

        # Wrap with IDMap to support custom IDs
        # This MUST be done before adding any data
        self.index = faiss.IndexIDMap(self.index)

        # Apply GPU acceleration if requested
        if self.use_gpu:
            if not hasattr(faiss, "StandardGpuResources"):
                logger.warning(
                    "GPU requested but faiss-gpu not available. "
                    "Install with: conda install -c conda-forge faiss-gpu"
                )
            else:
                res = faiss.StandardGpuResources()
                self.index = faiss.index_cpu_to_gpu(res, 0, self.index)
                logger.info("Moved index to GPU")

        logger.info(f"Created index: {self._get_index_description()}")

    def train(self, train_vectors: np.ndarray) -> None:
        """
        Train the index on a representative sample.

        Implements Step 2 from Section 3.3: Train the Index.

        The training performs k-means clustering for:
        1. IVF: nlist cluster centroids
        2. PQ: m × 2^nbits sub-vector centroids

        Args:
            train_vectors: Training vectors (typically 1M-2M samples, shape: (N, 768))

        Note:
            - Training is expensive but done only once (offline)
            - Use a representative sample, not necessarily all data
            - For billion-scale indexes, 1-2M samples are sufficient
        """
        if self.is_trained:
            logger.warning("Index is already trained, skipping")
            return

        if self.index is None:
            self._create_index()

        # CRITICAL STEP: L2-normalize training vectors
        # This ensures L2 distance = cosine similarity (Section 4.1)
        logger.info("L2-normalizing training vectors...")
        train_vectors = train_vectors.astype(np.float32)
        faiss.normalize_L2(train_vectors)

        logger.info(f"Training index on {len(train_vectors)} vectors...")
        self.index.train(train_vectors)
        self.is_trained = True
        logger.info("Index training complete")

    def add(self, vectors: np.ndarray, ids: Optional[np.ndarray] = None) -> None:
        """
        Add vectors to the trained index.

        Implements Step 3 from Section 3.3: Populate the Index.

        Args:
            vectors: Vectors to add (N, 768)
            ids: Custom IDs for each vector. If None, uses sequential IDs.

        Note:
            - Vectors are automatically L2-normalized (critical!)
            - Can be called multiple times to add in batches
            - All vectors must be added AFTER training
        """
        if not self.is_trained:
            raise RuntimeError("Index must be trained before adding vectors. Call train() first.")

        if self.index is None:
            raise RuntimeError("Index not created. Call build() or _create_index() first.")

        # Generate sequential IDs if not provided
        if ids is None:
            current_size = self.index.ntotal
            ids = np.arange(current_size, current_size + len(vectors), dtype=np.int64)
        else:
            ids = np.asarray(ids, dtype=np.int64)

        # CRITICAL STEP: L2-normalize vectors in-place
        # This is the linchpin that enables cosine similarity via L2 distance (Section 4.1)
        vectors = vectors.astype(np.float32).copy()  # Copy to avoid modifying original
        faiss.normalize_L2(vectors)

        logger.info(f"Adding {len(vectors)} vectors to index...")
        self.index.add_with_ids(vectors, ids)
        logger.info(f"Index now contains {self.index.ntotal} vectors")

    def set_nprobe(self, nprobe: int) -> None:
        """
        Set the number of clusters to probe at query time.

        This is the most important runtime parameter for tuning speed/accuracy.

        Args:
            nprobe: Number of clusters to search (1 = fastest, nlist = exact)

        Guidelines (Table 3.2):
            - nprobe=1: Very fast, lower accuracy
            - nprobe=16: Good balance (default)
            - nprobe=32: Higher accuracy, slower
            - nprobe=nlist: Exact search (defeats IVF purpose)
        """
        if self.index_type == IndexType.FLAT:
            logger.warning("nprobe has no effect on Flat index (already exact search)")
            return

        # Access the underlying IVF index (unwrap IDMap if needed)
        index = self.index
        if isinstance(index, faiss.IndexIDMap):
            index = faiss.downcast_index(index.index)

        if hasattr(index, "nprobe"):
            index.nprobe = nprobe
            self.nprobe = nprobe
            logger.info(f"Set nprobe = {nprobe}")
        else:
            logger.warning(f"Index type {type(index)} does not support nprobe")

    def save(self, file_path: str) -> None:
        """
        Save the index to disk.

        Args:
            file_path: Path to save the index file

        Example:
            >>> builder.save("clones.index")
        """
        if self.index is None:
            raise RuntimeError("No index to save. Build the index first.")

        file_path = str(Path(file_path).resolve())
        logger.info(f"Saving index to {file_path}")

        # If using GPU, move to CPU before saving
        index_to_save = self.index
        if self.use_gpu:
            index_to_save = faiss.index_gpu_to_cpu(self.index)

        faiss.write_index(index_to_save, file_path)
        logger.info(f"Index saved ({self.index.ntotal} vectors)")

    @classmethod
    def load(cls, file_path: str, use_gpu: bool = False) -> "FAISSIndexBuilder":
        """
        Load a saved index from disk.

        Args:
            file_path: Path to the index file
            use_gpu: Whether to move index to GPU after loading

        Returns:
            FAISSIndexBuilder instance with loaded index

        Example:
            >>> builder = FAISSIndexBuilder.load("clones.index")
        """
        file_path = str(Path(file_path).resolve())
        logger.info(f"Loading index from {file_path}")

        # Create a new instance
        instance = cls.__new__(cls)

        # Load the index
        index = faiss.read_index(file_path)

        # Move to GPU if requested
        if use_gpu:
            if not hasattr(faiss, "StandardGpuResources"):
                logger.warning("GPU requested but faiss-gpu not available")
            else:
                res = faiss.StandardGpuResources()
                index = faiss.index_cpu_to_gpu(res, 0, index)
                logger.info("Moved index to GPU")

        instance.index = index
        instance.is_trained = True
        instance.use_gpu = use_gpu

        # Try to infer parameters from the index
        instance.dimension = index.d

        logger.info(f"Loaded index with {index.ntotal} vectors (dimension: {index.d})")
        return instance

    def get_stats(self) -> dict:
        """Get statistics about the current index."""
        if self.index is None:
            return {"status": "not_created"}

        return {
            "index_type": self.index_type.value,
            "dimension": self.dimension,
            "num_vectors": self.index.ntotal,
            "is_trained": self.is_trained,
            "nlist": self.nlist if self.index_type != IndexType.FLAT else None,
            "nprobe": self.nprobe if self.index_type != IndexType.FLAT else None,
            "m": self.m if self.index_type == IndexType.IVF_PQ else None,
            "nbits": self.nbits if self.index_type == IndexType.IVF_PQ else None,
            "use_gpu": self.use_gpu,
        }


def cosine_to_l2_threshold(cosine_similarity: float) -> float:
    """
    Convert cosine similarity threshold to L2 distance threshold.

    Implements the conversion formula from Section 4.3:
        D_L2 = sqrt(2 - 2 * cos_sim)

    This is valid ONLY for L2-normalized vectors.

    Args:
        cosine_similarity: Desired minimum cosine similarity (0 to 1)

    Returns:
        Corresponding L2 distance threshold

    Example:
        >>> threshold = cosine_to_l2_threshold(0.95)  # Returns 0.316
        >>> # Use with: index.range_search(query, threshold)
    """
    if not 0 <= cosine_similarity <= 1:
        raise ValueError(f"cosine_similarity must be in [0, 1], got {cosine_similarity}")

    l2_distance = np.sqrt(2 - 2 * cosine_similarity)
    return float(l2_distance)


def l2_to_cosine_similarity(l2_distance: float) -> float:
    """
    Convert L2 distance to cosine similarity.

    Inverse of cosine_to_l2_threshold:
        cos_sim = 1 - (D_L2² / 2)

    Args:
        l2_distance: L2 distance between normalized vectors

    Returns:
        Corresponding cosine similarity

    Example:
        >>> sim = l2_to_cosine_similarity(0.316)  # Returns ~0.95
    """
    cosine_sim = 1 - (l2_distance ** 2) / 2
    return float(cosine_sim)
