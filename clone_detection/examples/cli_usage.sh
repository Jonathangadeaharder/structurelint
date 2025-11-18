#!/bin/bash
# Example CLI usage for semantic clone detection

# This script demonstrates how to use the clone-detect CLI tool

set -e

echo "=========================================="
echo "Semantic Clone Detection - CLI Example"
echo "=========================================="
echo

# 1. Ingest a codebase
echo "Step 1: Ingesting codebase..."
clone-detect ingest \
    --source-dir /path/to/your/codebase \
    --index-output clones.index \
    --metadata-db clones.db \
    --languages python,javascript,java \
    --exclude '**/node_modules/**' \
    --exclude '**/*test*' \
    --device cuda \
    --batch-size 64 \
    --nlist 4096 \
    --verbose

echo
echo "Step 2: Viewing index information..."
clone-detect info \
    --index clones.index \
    --metadata-db clones.db

echo
echo "Step 3: Searching for clones by code..."
clone-detect search \
    --index clones.index \
    --metadata-db clones.db \
    --query-code "def calculate_total(items): return sum(item.price for item in items)" \
    --similarity 0.95 \
    --max-results 20 \
    --device cuda

echo
echo "Step 4: Searching for clones by file location..."
clone-detect search \
    --index clones.index \
    --metadata-db clones.db \
    --query-file src/utils/helpers.py \
    --line-number 42 \
    --similarity 0.90 \
    --max-results 50

echo
echo "=========================================="
echo "Example complete!"
echo "=========================================="
