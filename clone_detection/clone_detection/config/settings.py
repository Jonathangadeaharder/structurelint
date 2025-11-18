"""
Configuration management using Pydantic for type safety and validation.
"""

from pathlib import Path
from typing import List, Optional

import yaml
from pydantic import BaseModel, Field, validator


class ModelConfig(BaseModel):
    """GraphCodeBERT model configuration."""

    name: str = Field(
        default="microsoft/graphcodebert-base",
        description="HuggingFace model name or local path",
    )
    device: Optional[str] = Field(
        default=None, description="Device to use (cuda/cpu, None for auto)"
    )
    batch_size: int = Field(default=32, description="Batch size for embedding generation")
    max_length: int = Field(default=512, description="Maximum sequence length")

    @validator("device")
    def validate_device(cls, v):
        if v is not None and v not in ["cpu", "cuda"]:
            raise ValueError("device must be 'cpu', 'cuda', or None")
        return v


class IndexConfig(BaseModel):
    """FAISS index configuration."""

    type: str = Field(default="IndexIVFPQ", description="FAISS index type")
    dimension: int = Field(default=768, description="Vector dimension")
    nlist: int = Field(default=4096, description="Number of IVF clusters")
    m: int = Field(default=64, description="PQ sub-vectors")
    nbits: int = Field(default=8, description="Bits per PQ code")
    nprobe: int = Field(default=16, description="Number of clusters to probe")
    use_gpu: bool = Field(default=False, description="Use GPU acceleration")

    @validator("m")
    def validate_m_divides_dimension(cls, v, values):
        if "dimension" in values and values["dimension"] % v != 0:
            raise ValueError(f"m ({v}) must evenly divide dimension ({values['dimension']})")
        return v


class ParsingConfig(BaseModel):
    """Tree-sitter parsing configuration."""

    languages: List[str] = Field(
        default=["python", "javascript", "java", "go"],
        description="Languages to parse",
    )
    chunk_size: int = Field(default=512, description="Max tokens per function")
    exclude_patterns: List[str] = Field(
        default_factory=lambda: [
            "**/node_modules/**",
            "**/__pycache__/**",
            "**/venv/**",
            "**/build/**",
            "**/dist/**",
            "**/*.min.js",
        ],
        description="Glob patterns to exclude",
    )


class QueryConfig(BaseModel):
    """Query/search configuration."""

    default_similarity: float = Field(default=0.95, description="Default similarity threshold")
    max_results: int = Field(default=100, description="Maximum results to return")

    @validator("default_similarity")
    def validate_similarity(cls, v):
        if not 0 <= v <= 1:
            raise ValueError("default_similarity must be between 0 and 1")
        return v


class CloneDetectionConfig(BaseModel):
    """Complete configuration for the clone detection system."""

    model: ModelConfig = Field(default_factory=ModelConfig)
    index: IndexConfig = Field(default_factory=IndexConfig)
    parsing: ParsingConfig = Field(default_factory=ParsingConfig)
    query: QueryConfig = Field(default_factory=QueryConfig)

    class Config:
        json_schema_extra = {
            "example": {
                "model": {
                    "name": "microsoft/graphcodebert-base",
                    "device": "cuda",
                    "batch_size": 32,
                },
                "index": {"nlist": 4096, "nprobe": 16},
                "parsing": {"languages": ["python", "java"]},
                "query": {"default_similarity": 0.95},
            }
        }

    def to_yaml(self, file_path: str) -> None:
        """Save configuration to YAML file."""
        data = self.model_dump()
        with open(file_path, "w") as f:
            yaml.dump(data, f, default_flow_style=False, sort_keys=False)

    @classmethod
    def from_yaml(cls, file_path: str) -> "CloneDetectionConfig":
        """Load configuration from YAML file."""
        with open(file_path, "r") as f:
            data = yaml.safe_load(f)
        return cls(**data)


def load_config(config_path: Optional[str] = None) -> CloneDetectionConfig:
    """
    Load configuration from file or use defaults.

    Args:
        config_path: Path to YAML config file. If None, uses defaults.

    Returns:
        CloneDetectionConfig instance
    """
    if config_path is None:
        return CloneDetectionConfig()

    path = Path(config_path)
    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {config_path}")

    return CloneDetectionConfig.from_yaml(str(path))
