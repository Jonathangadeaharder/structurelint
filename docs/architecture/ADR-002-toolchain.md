# ADR-002: Toolchain — Go Toolchain and golangci-lint

**Status:** Accepted  
**Date:** 2026-05-17  
**Deciders:** Project team  
**References:** [.golangci.yml](../../.golangci.yml), [go.mod](../../go.mod), [ci.yml](../../.github/workflows/ci.yml)

---

## Context

The project needs a consistent, reproducible toolchain for building, linting, and formatting. The tooling must align with Go ecosystem conventions while supporting the project's multi-language analysis requirements via tree-sitter.

## Decision

### Primary Toolchain

1. **Go 1.24.7** as the runtime and compiler (`go.mod` specifies `go 1.24.7`). Selected for goroutine support, modern stdlib, and fast compilation.

2. **golangci-lint v2.1.0** as the linter aggregator. Configured in `.golangci.yml` with these enabled linters:
   - `errcheck` — missing error checks
   - `govet` — suspicious constructs
   - `ineffassign` — ineffective assignments
   - `staticcheck` — static analysis
   - `unused` — unused code detection
   - `gocyclo` — cyclomatic complexity (threshold: 26)
   - `gocognit` — cognitive complexity (threshold: 85)

3. **gocyclo** (standalone) for complexity gates in CI — separate from golangci-lint to allow per-function thresholds.

4. **Standard `go build` / `go test`** — no alternative build systems (no Makefile, no Bazel). Simplicity wins.

### Dependencies

- **tree-sitter** (`github.com/smacker/go-tree-sitter`) — CGo-based multi-language parsing. Necessary for non-Go language support (Python, TypeScript, etc.).
- **charmbracelet** stack (`bubbletea`, `bubbles`, `lipgloss`) — TUI for interactive output.
- **testify** — test assertions.
- **rapid** — property-based testing (PBT).
- **yaml.v3** — YAML config parsing.

### Dependency Management

- Go modules (`go.mod` / `go.sum`) — no vendoring.
- Dependencies pinned to specific versions, not ranges.

## Consequences

- **Positive:** Zero external runtime dependencies — single binary deployment. Fast CI (cached Go modules). Standard tooling familiar to all Go developers.
- **Negative:** CGo (tree-sitter) prevents cross-compilation for some target/OS combinations. Static linking requires `CGO_ENABLED=0` builds for pure-Go deployments.
- **Trade-off:** golangci-lint v2.1.0 pinned via curl install in CI rather than `go install` — ensures version consistency but adds a network dependency.
- **Risk:** tree-sitter CGo has caused nil pointer panics with Go 1.26+ stdlib changes. Monitor upstream compatibility.

## Compliance

`go mod tidy` must pass before any merge. `golangci-lint run --timeout=5m` must pass. All lint and complexity checks are enforced in CI via ci.yml.
