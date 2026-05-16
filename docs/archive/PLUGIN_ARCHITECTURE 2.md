# Plugin Architecture: Semantic Clone Detection

**Date**: November 18, 2025
**Status**: ✅ IMPLEMENTED (Phase 3.1)

---

## Overview

Structurelint uses a **plugin architecture** to keep the core binary small and fast while providing optional advanced features like semantic (ML-based) clone detection.

### Architecture

```
┌─────────────────────────────────────────┐
│   structurelint (Core Binary)           │
│   ├─ Syntactic Clone Detection (Built-in)
│   ├─ Architecture Linting (Built-in)    │
│   ├─ Graph Visualization (Built-in)     │
│   └─ Plugin Client (HTTP)               │
└──────────────┬──────────────────────────┘
               │ HTTP/JSON API
               │ (Optional)
┌──────────────┴──────────────────────────┐
│   Semantic Clone Detection Plugin       │
│   (Python HTTP Server)                  │
│   ├─ FastAPI HTTP Server                │
│   ├─ GraphCodeBERT Embedder             │
│   ├─ FAISS Similarity Search            │
│   └─ Tree-sitter Parser                 │
└─────────────────────────────────────────┘
```

---

## Benefits

### ✅ Small Core Binary
- **Without plugin**: ~20MB (no Python, no ML models)
- **With plugin**: Core stays ~20MB, plugin separate download
- **Comparison**: Monolithic approach would be >500MB

### ✅ Fast Installation
- Core: `go install` (30 seconds)
- Plugin: `pip install` (optional, 5-10 minutes)
- Users can opt-in to ML features

### ✅ Graceful Degradation
- If plugin not available → syntactic detection only
- No errors, just warnings
- Seamless user experience

### ✅ Language Flexibility
- Core: Pure Go (fast, cross-platform)
- Plugin: Python (rich ML ecosystem)
- Best of both worlds

---

## Plugin API

### HTTP Endpoints

#### `GET /health`
Check plugin health and availability.

**Response**:
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "capabilities": [
    "semantic-clone-detection",
    "graphcodebert-embeddings",
    "faiss-indexing"
  ],
  "message": "Semantic clone detection ready"
}
```

#### `POST /api/v1/detect`
Detect semantic clones.

**Request**:
```json
{
  "source_dir": "/path/to/code",
  "languages": ["go", "python", "javascript"],
  "exclude_patterns": ["**/*_test.*", "**/vendor/**"],
  "similarity_threshold": 0.85,
  "max_results": 100
}
```

**Response**:
```json
{
  "clones": [
    {
      "source_file": "internal/api/handler.go",
      "source_start_line": 10,
      "source_end_line": 25,
      "target_file": "internal/service/processor.go",
      "target_start_line": 50,
      "target_end_line": 65,
      "similarity": 0.92,
      "explanation": "Semantic similarity: 0.920"
    }
  ],
  "stats": {
    "files_analyzed": 150,
    "functions_analyzed": 1200,
    "duration_ms": 45000,
    "model_used": "microsoft/graphcodebert-base"
  },
  "error": null
}
```

---

## Usage

### 1. Syntactic Detection (Default, Built-in)

```bash
# No setup required - works out of the box
structurelint clones
```

**Performance**: <1 second for 1000 files
**Binary Size**: ~20MB
**Dependencies**: None

### 2. Semantic Detection (Optional Plugin)

#### Step 1: Install Plugin Dependencies

```bash
cd clone_detection
pip install -r requirements-plugin.txt
```

**Note**: This downloads ~500MB of ML models (GraphCodeBERT)

#### Step 2: Start Plugin Server

```bash
python plugin_server.py

# Or with custom host/port:
python plugin_server.py --host 0.0.0.0 --port 8765
```

**Output**:
```
INFO:     Starting plugin server on 127.0.0.1:8765
INFO:     Initializing GraphCodeBERT embedder...
INFO:     Semantic clone detection ready
```

#### Step 3: Run Semantic Detection

```bash
# In another terminal:
structurelint clones --mode semantic
```

**Performance**: ~30-60 seconds for 1000 functions
**Accuracy**: Detects Type-4 (semantic) clones

### 3. Both Modes (Comprehensive)

```bash
structurelint clones --mode both
```

Runs both syntactic (fast) and semantic (accurate) detection.

---

## Configuration

### Environment Variables

```bash
# Plugin URL (default: http://localhost:8765)
export STRUCTURELINT_PLUGIN_URL=http://localhost:8765

# Use in CLI:
structurelint clones --mode semantic --plugin-url $STRUCTURELINT_PLUGIN_URL
```

### YAML Configuration

```yaml
# .structurelint.yml
clone-detection:
  mode: both  # syntactic, semantic, or both

  # Syntactic options
  min-tokens: 20
  min-lines: 3
  k-gram-size: 20

  # Semantic options
  plugin-url: http://localhost:8765
  similarity-threshold: 0.85

  # Common options
  exclude-patterns:
    - "**/*_test.go"
    - "**/*_gen.go"
    - "**/vendor/**"
```

---

## Deployment Options

### Option 1: Local Development (Default)

**Setup**:
```bash
# Terminal 1: Start plugin
cd clone_detection
python plugin_server.py

# Terminal 2: Use structurelint
structurelint clones --mode semantic
```

**Best for**: Local development, testing

### Option 2: Docker Container

**Dockerfile** (in `clone_detection/`):
```dockerfile
FROM python:3.10-slim

WORKDIR /app
COPY requirements-plugin.txt .
RUN pip install --no-cache-dir -r requirements-plugin.txt

COPY . .
EXPOSE 8765

CMD ["python", "plugin_server.py", "--host", "0.0.0.0", "--port", "8765"]
```

**Run**:
```bash
# Build image
docker build -t structurelint-plugin clone_detection/

# Run container
docker run -d -p 8765:8765 structurelint-plugin

# Use from host
structurelint clones --mode semantic --plugin-url http://localhost:8765
```

**Best for**: CI/CD, shared team server

### Option 3: Remote Server

**Setup** (on server):
```bash
# Install dependencies
pip install -r requirements-plugin.txt

# Start with systemd/supervisor
python plugin_server.py --host 0.0.0.0 --port 8765
```

**Use** (from client):
```bash
structurelint clones --mode semantic --plugin-url http://server:8765
```

**Best for**: Large teams, centralized deployment

### Option 4: Serverless (Future)

Future enhancement: Deploy plugin as AWS Lambda / Google Cloud Function

---

## Graceful Degradation

The plugin architecture is designed to **never break** the user experience:

### Scenario 1: Plugin Not Running

```bash
$ structurelint clones --mode semantic

🧠 Detecting semantic clones via plugin at http://localhost:8765...
⚠ Warning: Semantic clone detection plugin not available at http://localhost:8765
  To enable semantic detection:
    1. cd clone_detection
    2. pip install -r requirements-plugin.txt
    3. python plugin_server.py

Error: semantic clone detection plugin required but not available
```

### Scenario 2: Plugin Fails (mode=both)

```bash
$ structurelint clones --mode both

🔍 Detecting syntactic clones in /path/to/code...
✓ No syntactic clones found

🧠 Detecting semantic clones via plugin at http://localhost:8765...
⚠ Warning: Semantic detection failed: connection refused

Continuing with syntactic detection only...
```

**Result**: Syntactic detection completes, semantic skipped

### Scenario 3: Plugin Succeeds

```bash
$ structurelint clones --mode both

🔍 Detecting syntactic clones in /path/to/code...
Found 2 syntactic clones...

🧠 Detecting semantic clones via plugin at http://localhost:8765...

Found 5 semantic clone pairs:

1. Similarity: 92.00%
   internal/api/handler.go:10-25
   internal/service/processor.go:50-65
   Semantic similarity: 0.920

...

Analyzed 150 files, 1200 functions in 45000ms
```

**Result**: Both modes complete successfully

---

## Performance Comparison

| Mode | Binary Size | Install Time | Analysis Time (1000 files) | Clone Types |
|------|-------------|--------------|---------------------------|-------------|
| **Syntactic only** | ~20MB | 30s | <1s | Type-1, 2, 3 |
| **Semantic only** | ~20MB + plugin | 30s + 10min | ~60s | Type-4 |
| **Both** | ~20MB + plugin | 30s + 10min | ~61s | All types |

---

## Security Considerations

### 1. Local Plugin (Default)
- Plugin runs on localhost only
- No external network access required
- Secure by default

### 2. Remote Plugin
- Use HTTPS for production
- Add authentication (API keys, OAuth)
- Network isolation (VPC, firewall)

### 3. Code Privacy
- Source code sent to plugin server
- Use local deployment for sensitive code
- Or run plugin in air-gapped environment

---

## Future Enhancements

### Phase 3.2: ONNX Runtime (Planned)

**Goal**: Embed lightweight ML model in core binary

**Approach**:
1. Export GraphCodeBERT to ONNX format
2. Quantize INT8: 500MB → 150MB
3. Integrate `onnxruntime_go`
4. Benchmark performance

**Decision Gate**:
- IF: <100MB binary increase AND <100ms/snippet
- THEN: Embed in core (flag: `--enable-semantic`)
- ELSE: Keep as plugin (current approach)

**Benefits**:
- No separate server required
- Single binary deployment
- Still smaller than monolithic approach

---

## Troubleshooting

### Problem: Plugin won't start

**Symptom**:
```
ImportError: No module named 'transformers'
```

**Solution**:
```bash
cd clone_detection
pip install -r requirements-plugin.txt
```

### Problem: Plugin slow

**Symptom**: Semantic detection takes >5 minutes

**Solution**: Use GPU
```bash
# Install GPU version of PyTorch and FAISS
pip install torch torchvision --index-url https://download.pytorch.org/whl/cu118
pip install faiss-gpu

# Restart plugin
python plugin_server.py
```

### Problem: Connection refused

**Symptom**:
```
⚠ Warning: plugin request failed: connection refused
```

**Solution**:
1. Check plugin is running: `curl http://localhost:8765/health`
2. Check firewall allows port 8765
3. Verify URL: `--plugin-url http://localhost:8765`

### Problem: Out of memory

**Symptom**: Plugin crashes analyzing large codebase

**Solution**:
1. Increase system memory
2. Reduce batch size (edit `plugin_server.py`)
3. Process in chunks (analyze subdirectories separately)

---

## API Clients

### Go Client (Built-in)

```go
import "github.com/structurelint/structurelint/internal/plugin"

// Create client
client := plugin.NewHTTPPluginClient("http://localhost:8765")

// Check availability
if !client.IsAvailable() {
    // Plugin not available - graceful degradation
    fmt.Println("Plugin not available")
    return
}

// Detect clones
req := &plugin.SemanticCloneRequest{
    SourceDir:           "/path/to/code",
    Languages:           []string{"go", "python"},
    SimilarityThreshold: 0.85,
    MaxResults:          100,
}

resp, err := client.DetectClones(context.Background(), req)
if err != nil {
    // Handle error
}

// Process results
for _, clone := range resp.Clones {
    fmt.Printf("Clone: %s:%d -> %s:%d (%.2f%%)\n",
        clone.SourceFile, clone.SourceStartLine,
        clone.TargetFile, clone.TargetStartLine,
        clone.Similarity*100)
}
```

### Python Client (For Testing)

```python
import requests

# Check health
response = requests.get("http://localhost:8765/health")
print(response.json())

# Detect clones
request = {
    "source_dir": "/path/to/code",
    "languages": ["go", "python"],
    "similarity_threshold": 0.85,
    "max_results": 100
}

response = requests.post("http://localhost:8765/api/v1/detect", json=request)
result = response.json()

print(f"Found {len(result['clones'])} semantic clones")
```

---

## Conclusion

The plugin architecture provides:

✅ **Small Core Binary** (~20MB vs >500MB monolithic)
✅ **Fast Installation** (30s vs 10+ minutes)
✅ **Optional ML Features** (opt-in semantic detection)
✅ **Graceful Degradation** (never breaks user experience)
✅ **Flexible Deployment** (local, Docker, remote)
✅ **Best of Both Worlds** (Go performance + Python ML ecosystem)

This architecture achieves the Phase 3 goal of **retaining semantic clone detection without bloating the core binary**.

---

**Author**: Claude (Sonnet 4.5)
**Date**: November 18, 2025
**Branch**: `claude/audit-structurelint-roadmap-01PYzjfTy7n7KF6kyKgFDEe1`
