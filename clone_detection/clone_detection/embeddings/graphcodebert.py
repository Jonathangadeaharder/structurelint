"""
GraphCodeBERT embedding generator for semantic code representation.

This module implements Part II of the blueprint: Vectorization.
It uses the GraphCodeBERT model to transform code snippets into 768-dimensional
semantic vectors, following the "Path A" implementation (no explicit DFG).
"""

import logging
from typing import List, Optional, Union

import numpy as np
import torch
from transformers import RobertaModel, RobertaTokenizer

from clone_detection.parsers.tree_sitter_parser import CodeSnippet

logger = logging.getLogger(__name__)


class GraphCodeBERTEmbedder:
    """
    Semantic code embedder using GraphCodeBERT.

    This class implements the embedding extraction strategy from Section 2.3:
    - Uses the <s> token's last hidden state as the code representation
    - Supports batch inference for efficiency
    - Can run on CPU or GPU

    The model is pre-trained with Data Flow Graph awareness, providing superior
    semantic understanding compared to standard CodeBERT, even without explicit
    DFG input at inference time (Path A implementation).

    Example:
        >>> embedder = GraphCodeBERTEmbedder(device="cuda")
        >>> code = ["def add(a, b): return a + b"]
        >>> embeddings = embedder.embed_batch(code)
        >>> print(embeddings.shape)  # (1, 768)
    """

    def __init__(
        self,
        model_name: str = "microsoft/graphcodebert-base",
        device: Optional[str] = None,
        max_length: int = 512,
        batch_size: int = 32,
    ):
        """
        Initialize the GraphCodeBERT embedder.

        Args:
            model_name: HuggingFace model identifier. Can be:
                       - "microsoft/graphcodebert-base" (default, pre-trained)
                       - Path to a fine-tuned model checkpoint
            device: Device to run on ("cuda", "cpu", or None for auto-detect)
            max_length: Maximum sequence length (GraphCodeBERT limit: 512)
            batch_size: Batch size for inference
        """
        self.model_name = model_name
        self.max_length = max_length
        self.batch_size = batch_size

        # Auto-detect device if not specified
        if device is None:
            self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        else:
            self.device = torch.device(device)

        logger.info(f"Initializing GraphCodeBERT on device: {self.device}")

        # Load tokenizer and model
        self.tokenizer = RobertaTokenizer.from_pretrained(model_name)
        self.model = RobertaModel.from_pretrained(model_name)

        # Move model to device and set to evaluation mode
        self.model.to(self.device)
        self.model.eval()

        logger.info(f"Loaded model: {model_name}")
        logger.info(f"Embedding dimension: 768")

    @torch.no_grad()
    def embed_batch(
        self, code_snippets: Union[List[str], List[CodeSnippet]]
    ) -> np.ndarray:
        """
        Generate embeddings for a batch of code snippets.

        This implements the batch inference strategy from Section 2.3:
        1. Tokenize all snippets (truncate to max_length)
        2. Forward pass through GraphCodeBERT
        3. Extract the <s> token's last hidden state

        Args:
            code_snippets: List of code strings or CodeSnippet objects

        Returns:
            NumPy array of shape (N, 768) containing the embeddings

        Example:
            >>> embedder = GraphCodeBERTEmbedder()
            >>> code = ["def foo(): pass", "def bar(): return 1"]
            >>> embeddings = embedder.embed_batch(code)
            >>> print(embeddings.shape)  # (2, 768)
        """
        # Extract code strings if CodeSnippet objects were provided
        if code_snippets and isinstance(code_snippets[0], CodeSnippet):
            code_strings = [snippet.code for snippet in code_snippets]
        else:
            code_strings = code_snippets

        if not code_strings:
            return np.zeros((0, 768), dtype=np.float32)

        # Process in batches
        all_embeddings = []

        for i in range(0, len(code_strings), self.batch_size):
            batch = code_strings[i : i + self.batch_size]
            batch_embeddings = self._embed_single_batch(batch)
            all_embeddings.append(batch_embeddings)

            if (i + self.batch_size) % (self.batch_size * 10) == 0:
                logger.debug(f"Processed {i + len(batch)}/{len(code_strings)} snippets")

        # Concatenate all batch embeddings
        embeddings = np.vstack(all_embeddings)

        logger.info(f"Generated {embeddings.shape[0]} embeddings")
        return embeddings

    def _embed_single_batch(self, code_batch: List[str]) -> np.ndarray:
        """
        Embed a single batch of code snippets.

        This is the core inference logic from Section 2.3:
        - Tokenize with padding and truncation
        - Forward pass through the model
        - Extract <s> token representation

        Args:
            code_batch: List of code strings (size <= batch_size)

        Returns:
            NumPy array of shape (batch_size, 768)
        """
        # Tokenize the batch
        # - padding="max_length": Pad all sequences to max_length
        # - truncation=True: Truncate sequences longer than max_length
        # - return_tensors="pt": Return PyTorch tensors
        inputs = self.tokenizer(
            code_batch,
            padding="max_length",
            truncation=True,
            max_length=self.max_length,
            return_tensors="pt",
        )

        # Move inputs to the same device as the model
        inputs = {k: v.to(self.device) for k, v in inputs.items()}

        # Forward pass through GraphCodeBERT
        outputs = self.model(**inputs)

        # Extract the <s> token's last hidden state
        # outputs.last_hidden_state shape: (batch_size, seq_len, 768)
        # The <s> token is always at position 0
        # Shape: (batch_size, 768)
        embeddings = outputs.last_hidden_state[:, 0, :]

        # Move to CPU and convert to numpy
        embeddings_np = embeddings.cpu().numpy()

        return embeddings_np

    def embed_single(self, code: str) -> np.ndarray:
        """
        Generate embedding for a single code snippet.

        Args:
            code: Source code string

        Returns:
            NumPy array of shape (768,)

        Example:
            >>> embedder = GraphCodeBERTEmbedder()
            >>> embedding = embedder.embed_single("def foo(): pass")
            >>> print(embedding.shape)  # (768,)
        """
        embeddings = self.embed_batch([code])
        return embeddings[0]

    def get_embedding_dimension(self) -> int:
        """Get the dimensionality of the embeddings (always 768 for GraphCodeBERT)."""
        return 768

    def get_model_info(self) -> dict:
        """Get information about the loaded model."""
        return {
            "model_name": self.model_name,
            "device": str(self.device),
            "max_length": self.max_length,
            "batch_size": self.batch_size,
            "embedding_dim": 768,
            "parameters": sum(p.numel() for p in self.model.parameters()),
        }

    def save_model(self, output_dir: str) -> None:
        """
        Save the model and tokenizer to a directory.

        Useful for saving fine-tuned models.

        Args:
            output_dir: Directory to save the model
        """
        logger.info(f"Saving model to {output_dir}")
        self.model.save_pretrained(output_dir)
        self.tokenizer.save_pretrained(output_dir)
        logger.info("Model saved successfully")

    @classmethod
    def load_model(cls, model_dir: str, device: Optional[str] = None) -> "GraphCodeBERTEmbedder":
        """
        Load a saved model from a directory.

        Args:
            model_dir: Directory containing the saved model
            device: Device to load on ("cuda", "cpu", or None)

        Returns:
            GraphCodeBERTEmbedder instance with loaded model
        """
        logger.info(f"Loading model from {model_dir}")
        return cls(model_name=model_dir, device=device)


def calculate_cosine_similarity(embedding1: np.ndarray, embedding2: np.ndarray) -> float:
    """
    Calculate cosine similarity between two embeddings.

    Args:
        embedding1: First embedding vector (768,)
        embedding2: Second embedding vector (768,)

    Returns:
        Cosine similarity in range [-1, 1]

    Example:
        >>> emb1 = np.random.randn(768)
        >>> emb2 = np.random.randn(768)
        >>> sim = calculate_cosine_similarity(emb1, emb2)
    """
    # Normalize vectors
    norm1 = np.linalg.norm(embedding1)
    norm2 = np.linalg.norm(embedding2)

    if norm1 == 0 or norm2 == 0:
        return 0.0

    # Cosine similarity = dot product of normalized vectors
    similarity = np.dot(embedding1, embedding2) / (norm1 * norm2)

    return float(similarity)
