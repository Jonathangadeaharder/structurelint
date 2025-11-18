"""
Main CLI entry point for the semantic clone detection system.

Implements the command-line interface for both Blueprint A (Batch Ingestion)
and Blueprint B (Query Pipeline).
"""

import logging
import sys
from pathlib import Path
from typing import List, Optional

import click
import numpy as np
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table

from clone_detection.embeddings.graphcodebert import GraphCodeBERTEmbedder
from clone_detection.indexing.faiss_index import FAISSIndexBuilder, IndexType
from clone_detection.parsers.tree_sitter_parser import TreeSitterParser
from clone_detection.query.metadata import MetadataStore
from clone_detection.query.search import CloneSearcher

console = Console()
logger = logging.getLogger(__name__)


def setup_logging(verbose: bool) -> None:
    """Configure logging based on verbosity level."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        handlers=[logging.StreamHandler()],
    )


@click.group()
@click.version_option(version="0.1.0")
def cli():
    """
    Semantic Code Clone Detection using GraphCodeBERT and FAISS.

    A production-grade system for detecting Type-4 (semantic) code clones
    across multiple programming languages.
    """
    pass


@cli.command()
@click.option(
    "--source-dir",
    required=True,
    type=click.Path(exists=True, file_okay=False),
    help="Root directory of the codebase to index",
)
@click.option(
    "--index-output",
    required=True,
    type=click.Path(),
    help="Output path for the FAISS index file",
)
@click.option(
    "--metadata-db",
    required=True,
    type=click.Path(),
    help="Output path for the metadata SQLite database",
)
@click.option(
    "--languages",
    default="python,javascript,java,go",
    help="Comma-separated list of languages to parse (default: python,javascript,java,go)",
)
@click.option(
    "--exclude",
    multiple=True,
    help="Glob patterns to exclude (e.g., --exclude '**/*test*' --exclude '**/node_modules/**')",
)
@click.option(
    "--model",
    default="microsoft/graphcodebert-base",
    help="GraphCodeBERT model name or path (default: microsoft/graphcodebert-base)",
)
@click.option(
    "--device",
    default=None,
    type=click.Choice(["cpu", "cuda"], case_sensitive=False),
    help="Device to use for embedding (default: auto-detect)",
)
@click.option(
    "--batch-size",
    default=32,
    type=int,
    help="Batch size for embedding generation (default: 32)",
)
@click.option(
    "--nlist",
    default=4096,
    type=int,
    help="Number of IVF clusters (default: 4096)",
)
@click.option(
    "--nprobe",
    default=16,
    type=int,
    help="Number of clusters to probe at query time (default: 16)",
)
@click.option(
    "--use-gpu",
    is_flag=True,
    help="Use GPU for FAISS index (requires faiss-gpu)",
)
@click.option(
    "--max-files",
    type=int,
    help="Maximum number of files to process (for testing)",
)
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Enable verbose logging",
)
def ingest(
    source_dir: str,
    index_output: str,
    metadata_db: str,
    languages: str,
    exclude: tuple,
    model: str,
    device: Optional[str],
    batch_size: int,
    nlist: int,
    nprobe: int,
    use_gpu: bool,
    max_files: Optional[int],
    verbose: bool,
):
    """
    Ingest and index a codebase (Blueprint A: Batch Ingestion Pipeline).

    This command implements the complete batch ingestion pipeline:
    1. Parse the codebase with Tree-sitter
    2. Generate GraphCodeBERT embeddings
    3. Build the FAISS index
    4. Save index and metadata

    Example:
        clone-detect ingest --source-dir /path/to/repo --index-output clones.index --metadata-db clones.db
    """
    setup_logging(verbose)

    console.print("[bold blue]Semantic Clone Detection - Batch Ingestion[/bold blue]")
    console.print()

    # Parse language list
    lang_list = [lang.strip() for lang in languages.split(",")]

    try:
        # Step 1: Parse codebase
        console.print("[bold]Step 1/4:[/bold] Parsing codebase with Tree-sitter")
        parser = TreeSitterParser(languages=lang_list)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task("Parsing source files...", total=None)
            snippets = parser.parse_directory(
                source_dir, exclude_patterns=list(exclude), max_files=max_files
            )

        console.print(f"✓ Parsed {len(snippets)} function snippets")
        console.print()

        if len(snippets) == 0:
            console.print("[red]No code snippets found. Check language settings and exclusions.[/red]")
            sys.exit(1)

        # Step 2: Generate embeddings
        console.print("[bold]Step 2/4:[/bold] Generating GraphCodeBERT embeddings")
        embedder = GraphCodeBERTEmbedder(
            model_name=model, device=device, batch_size=batch_size
        )

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task(f"Embedding {len(snippets)} snippets...", total=None)
            embeddings = embedder.embed_batch(snippets)

        console.print(f"✓ Generated {embeddings.shape[0]} embeddings (dim={embeddings.shape[1]})")
        console.print()

        # Step 3: Build FAISS index
        console.print("[bold]Step 3/4:[/bold] Building FAISS index")

        # Generate IDs
        ids = np.arange(len(snippets), dtype=np.int64)

        # Create and train index
        index_builder = FAISSIndexBuilder(
            dimension=768,
            index_type=IndexType.IVF_PQ,
            nlist=nlist,
            m=64,
            nbits=8,
            nprobe=nprobe,
            use_gpu=use_gpu,
        )

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task("Training index...", total=None)

            # Use first 100k vectors for training (or all if fewer)
            train_size = min(100000, len(embeddings))
            train_indices = np.random.choice(len(embeddings), train_size, replace=False)
            train_vectors = embeddings[train_indices]

            index_builder.train(train_vectors)

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task("Adding vectors to index...", total=None)
            index_builder.add(embeddings, ids)

        console.print(f"✓ Built index with {index_builder.index.ntotal} vectors")
        console.print()

        # Step 4: Save index and metadata
        console.print("[bold]Step 4/4:[/bold] Saving index and metadata")

        # Save FAISS index
        index_builder.save(index_output)
        console.print(f"✓ Saved FAISS index to {index_output}")

        # Save metadata
        metadata_store = MetadataStore(metadata_db)
        snippet_tuples = [(i, snippet) for i, snippet in enumerate(snippets)]
        metadata_store.add_snippets_batch(snippet_tuples)
        metadata_store.close()
        console.print(f"✓ Saved metadata to {metadata_db}")
        console.print()

        # Summary
        console.print("[bold green]✓ Ingestion complete![/bold green]")
        console.print(f"  Total snippets: {len(snippets)}")
        console.print(f"  Languages: {', '.join(lang_list)}")
        console.print(f"  Index type: IndexIVFPQ (nlist={nlist}, nprobe={nprobe})")

    except Exception as e:
        console.print(f"[red]Error during ingestion: {e}[/red]")
        if verbose:
            console.print_exception()
        sys.exit(1)


@cli.command()
@click.option(
    "--index",
    required=True,
    type=click.Path(exists=True),
    help="Path to the FAISS index file",
)
@click.option(
    "--metadata-db",
    required=True,
    type=click.Path(exists=True),
    help="Path to the metadata SQLite database",
)
@click.option(
    "--query-code",
    type=str,
    help="Source code to search for clones",
)
@click.option(
    "--query-file",
    type=click.Path(exists=True),
    help="File containing the code to search",
)
@click.option(
    "--line-number",
    type=int,
    help="Line number within the file (used with --query-file)",
)
@click.option(
    "--similarity",
    default=0.95,
    type=float,
    help="Minimum cosine similarity threshold (default: 0.95)",
)
@click.option(
    "--max-results",
    default=100,
    type=int,
    help="Maximum number of results to return (default: 100)",
)
@click.option(
    "--model",
    default="microsoft/graphcodebert-base",
    help="GraphCodeBERT model name (must match the one used for ingestion)",
)
@click.option(
    "--device",
    default=None,
    type=click.Choice(["cpu", "cuda"], case_sensitive=False),
    help="Device to use for embedding (default: auto-detect)",
)
@click.option(
    "--use-gpu",
    is_flag=True,
    help="Use GPU for FAISS index",
)
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Enable verbose logging",
)
def search(
    index: str,
    metadata_db: str,
    query_code: Optional[str],
    query_file: Optional[str],
    line_number: Optional[int],
    similarity: float,
    max_results: int,
    model: str,
    device: Optional[str],
    use_gpu: bool,
    verbose: bool,
):
    """
    Search for semantic clones (Blueprint B: Query Pipeline).

    Example:
        clone-detect search --index clones.index --metadata-db clones.db --query-code "def add(a, b): return a + b"
    """
    setup_logging(verbose)

    console.print("[bold blue]Semantic Clone Detection - Search[/bold blue]")
    console.print()

    try:
        # Validate inputs
        if not query_code and not query_file:
            console.print("[red]Error: Either --query-code or --query-file must be provided[/red]")
            sys.exit(1)

        # Load components
        console.print("Loading index and model...")

        # Load FAISS index
        index_builder = FAISSIndexBuilder.load(index, use_gpu=use_gpu)
        faiss_index = index_builder.index

        # Load embedder
        embedder = GraphCodeBERTEmbedder(model_name=model, device=device)

        # Load metadata
        metadata_store = MetadataStore(metadata_db)

        # Create searcher
        searcher = CloneSearcher(faiss_index, embedder, metadata_store)

        console.print(f"✓ Loaded index with {faiss_index.ntotal} code snippets")
        console.print()

        # Determine query code
        if query_file:
            if line_number is not None:
                # Search by file location
                console.print(f"Searching for clones of function at {query_file}:{line_number}")
                clones = searcher.find_clones_by_location(
                    query_file, line_number, similarity, max_results
                )
            else:
                # Read entire file as query
                with open(query_file, "r") as f:
                    query_code = f.read()
                console.print(f"Searching for clones of code in {query_file}")
                clones = searcher.find_clones(query_code, similarity, max_results)
        else:
            console.print("Searching for clones...")
            clones = searcher.find_clones(query_code, similarity, max_results)

        console.print()

        # Display results
        if not clones:
            console.print(f"[yellow]No clones found with similarity >= {similarity:.2f}[/yellow]")
        else:
            console.print(f"[bold green]Found {len(clones)} clones:[/bold green]")
            console.print()

            table = Table(show_header=True, header_style="bold magenta")
            table.add_column("#", style="dim", width=4)
            table.add_column("File", style="cyan")
            table.add_column("Lines", justify="right")
            table.add_column("Similarity", justify="right")
            table.add_column("Function", style="green")

            for i, clone in enumerate(clones, 1):
                table.add_row(
                    str(i),
                    clone.file_path,
                    f"{clone.start_line}-{clone.end_line}",
                    f"{clone.similarity:.3f}",
                    clone.function_name or "N/A",
                )

            console.print(table)

        metadata_store.close()

    except Exception as e:
        console.print(f"[red]Error during search: {e}[/red]")
        if verbose:
            console.print_exception()
        sys.exit(1)


@cli.command()
@click.option(
    "--index",
    required=True,
    type=click.Path(exists=True),
    help="Path to the FAISS index file",
)
@click.option(
    "--metadata-db",
    required=True,
    type=click.Path(exists=True),
    help="Path to the metadata database",
)
def info(index: str, metadata_db: str):
    """
    Display information about a built index.

    Example:
        clone-detect info --index clones.index --metadata-db clones.db
    """
    console.print("[bold blue]Index Information[/bold blue]")
    console.print()

    try:
        # Load index
        index_builder = FAISSIndexBuilder.load(index)
        stats = index_builder.get_stats()

        # Load metadata
        metadata_store = MetadataStore(metadata_db)

        table = Table(show_header=False)
        table.add_column("Property", style="cyan")
        table.add_column("Value", style="green")

        table.add_row("Index Type", stats["index_type"])
        table.add_row("Dimension", str(stats["dimension"]))
        table.add_row("Total Vectors", str(stats["num_vectors"]))
        table.add_row("Is Trained", str(stats["is_trained"]))

        if stats["nlist"]:
            table.add_row("IVF Clusters (nlist)", str(stats["nlist"]))
            table.add_row("Probe Count (nprobe)", str(stats["nprobe"]))

        if stats["m"]:
            table.add_row("PQ Sub-vectors (m)", str(stats["m"]))
            table.add_row("PQ Bits (nbits)", str(stats["nbits"]))

        table.add_row("Metadata Count", str(metadata_store.count()))
        table.add_row("Languages", ", ".join(metadata_store.get_languages()))

        console.print(table)

        metadata_store.close()

    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        sys.exit(1)


if __name__ == "__main__":
    cli()
