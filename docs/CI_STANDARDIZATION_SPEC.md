# CI Standardization Spec

## Goal

Every VidiomTM project gets a standardized set of 10 CI gates in its PR and
merge workflows. structurelint enforces that the required config files and
workflows exist with correct settings.

## Project Types

| Type | Detection | Repos |
|------|-----------|-------|
| Python | `pyproject.toml` exists | Gretel, gh-wait, sitcom-pilot, timelog, mac-control-mcp, scriptforge |
| SvelteKit | `svelte.config.js` or `svelte.config.ts` exists | ObstetraScroll |
| JS/TS | `package.json` exists, no svelte config | Vidiom, svelteuml, scriptforge |
| Rust | `Cargo.toml` exists | pytest-linter, vitest-linter, cargo-test-lint |
| Go | `go.mod` exists | structurelint |
| Other | none of the above | PPSN_FOGA_GECCO, lean-runtime-analysis |

## The 10 Gates

Ordered fast-to-slow. First failure stops the PR gate (fail-fast).

### Gate 1: Format

Ensures consistent code style without human review.

| Type | Command |
|------|---------|
| Python | `uvx ruff format --check .` |
| SvelteKit | `pnpm dlx prettier --check .` |
| JS/TS | `pnpm dlx @biomejs/biome check --formatter-enabled .` |
| Rust | `cargo fmt --check` |
| Go | n/a (gofmt is not a gate, formatting is implicit in build) |

### Gate 2: Lint

Code quality rules — no unhandled errors, unused variables, etc.

| Type | Command |
|------|---------|
| Python | `uvx ruff check .` |
| SvelteKit | `pnpm dlx eslint .` |
| JS/TS | `pnpm dlx @biomejs/biome check --linter-enabled .` |
| Rust | `cargo clippy -- -D warnings` |
| Go | `golangci-lint run` |

### Gate 3: Type Check

Full-project type correctness.

| Type | Command |
|------|---------|
| Python | `uvx pyright` |
| SvelteKit | `pnpm exec svelte-check --tsconfig ./tsconfig.json` |
| JS/TS | `pnpm exec tsc --noEmit` |
| Rust | `cargo check` |
| Go | `go vet ./...` |

### Gate 4: Build

Project compiles successfully.

| Type | Command |
|------|---------|
| Python | `uv build` (check only, no publish) |
| SvelteKit | `pnpm build` |
| JS/TS | `pnpm build` |
| Rust | `cargo build` |
| Go | `go build ./...` |

### Gate 5: Unit Tests

Isolated unit tests with coverage threshold (90% branch).

| Type | Command |
|------|---------|
| Python | `uv run pytest --cov --cov-branch --cov-fail-under=90` |
| SvelteKit | `pnpm exec vitest run --coverage --coverage.thresholds.branches=90` |
| JS/TS | `pnpm exec vitest run --coverage` (90% where applicable) |
| Rust | `cargo test --lib` |
| Go | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'` |

### Gate 6: Mutation Tests (changed files only)

Runs mutation testing only on files changed in the PR. Uses
`tj-actions/changed-files` to compute the file set, then passes only those to
the mutation runner.

| Type | Command |
|------|---------|
| Python | `uv run mutmut run --paths-to-mutate <changed_files>` |
| SvelteKit | `pnpm exec stryker run --mutate <changed_files>` |
| JS/TS | `pnpm exec stryker run --mutate <changed_files>` |
| Rust | `cargo test` with `tarpaulin` (or `cargo mutant` if configured) |
| Go | `go test` (mutation testing not standard in Go CI) |

**Skip conditions:**
- No changed source files → skip
- Only non-source changes (docs, config) → skip
- `any_changed == false` → skip

Implementation pattern:

```yaml
- name: Get changed files
  id: changed-files
  uses: tj-actions/changed-files@v44
  with:
    files: |
      src/**/*.py
      src/**/*.ts
      src/**/*.svelte
    separator: ','

- name: Mutation test
  if: steps.changed-files.outputs.any_changed == 'true'
  run: mutmut run --paths-to-mutate ${{ steps.changed-files.outputs.all_changed_files }}
```

### Gate 7: Dead Code

Finds unused exports, files, and dependencies.

| Type | Command |
|------|---------|
| Python | `uvx vulture src/` |
| SvelteKit | `pnpm exec knip` |
| JS/TS | `pnpm exec knip` |
| Rust | `cargo +nightly udeps` (nightly only) |
| Go | skip (Go compiler eliminates dead code at build) |

### Gate 8: Duplication — jscpd (PR Blocker)

Literal copy-paste detection. Blocks PR if duplication exceeds threshold (3-5%).
Fast token-based matching, no false positives from string/variable abstraction.

All types: `jscpd --threshold 5 --pattern "src/**/*" .`

Config at `.jscpd.json`:
```json
{
  "threshold": 5,
  "pattern": ["src/**/*"],
  "ignore": ["**/*.test.*", "**/*.spec.*", "**/node_modules/**"]
}
```

### Gate 9: Dependency Vulnerability Audit

Check dependency tree for known CVEs.

| Type | Command |
|------|---------|
| Python | `uvx pip-audit` |
| SvelteKit | `pnpm audit --audit-level=high` |
| JS/TS | `pnpm audit --audit-level=high` |
| Rust | `cargo deny check advisories` |
| Go | `trivy filesystem --scanners vuln .` |

### Gate 10: Security — Semgrep (PR Blocker)

Fast syntax-tree-based security scan. Blocks the PR within seconds.
Runs on every `pull_request`. Every project type runs semgrep.

All types: `semgrep --config=auto .` (or `returntocorp/semgrep-action@v1` with `p/default`)

Semgrep config at `.semgrep.yml` in repo root.

### Gate 10: Integration Tests + E2E

Cross-service/in-browser tests. Run only when service deps change or on merge.

| Type | Command |
|------|---------|
| Python | `uv run pytest -m integration` |
| SvelteKit | `pnpm exec vitest run --config vitest.integration.ts` |
| JS/TS | `pnpm exec vitest run --config vitest.integration.ts` |
| Rust | `cargo test --test '*'` |
| Go | `go test -tags=integration ./...` |

**E2E (SvelteKit only):**
`pnpm exec playwright test`

Placed in merge gate (not PR gate) due to speed.

## CI Workflow Structure

### pr-gate.yml

Runs on `pull_request` to `main`/`master`. Fail-fast through gates 1-9.

```yaml
name: PR Gate

on:
  pull_request:
    branches: [main, master]

jobs:
  format:
    runs-on: ${{ matrix.runner }}
    steps: [...]

  lint:
    needs: [format]
    runs-on: ${{ matrix.runner }}

  typecheck:
    needs: [lint]

  build:
    needs: [typecheck]

  unit:
    needs: [build]

  mutation:
    needs: [build]
    if: steps.changed-files.outputs.any_changed == 'true'

  deadcode:
    needs: [build]

  deps:
    needs: [build]

  semgrep:
    needs: [build]

  integration:
    needs: [unit, build]
    # only when tests/src/migrations/deps change
```

### merge-gate.yml

Runs on `push` to `main`/`master`. Full suite including E2E.

```yaml
name: Merge Gate

on:
  push:
    branches: [main, master]

jobs:
  # same as PR gate + E2E at the end
  e2e:
    needs: [build, integration]
    # SvelteKit only
```

### (Post-merge) codeql.yml — Deep Security Audit

Separate workflow. Runs on push to main (auto after merge) and weekly cron.
Catches cross-file taint-tracking vulnerabilities that Semgrep can't see.

```yaml
name: CodeQL Advanced Security

on:
  push:
    branches: [main, master]
  schedule:
    - cron: '0 2 * * 1'

jobs:
  analyze:
    name: Analyze (${{ matrix.language }})
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      actions: read
      contents: read
    strategy:
      fail-fast: false
      matrix:
        language: ['python', 'javascript']
    steps:
      - uses: actions/checkout@v4
      - uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}
          queries: security-extended,security-and-quality
      - uses: github/codeql-action/autobuild@v3
      - uses: github/codeql-action/analyze@v3
```

Does NOT block PRs. Opens Security Alerts in GitHub's Security tab.

### SonarCloud — Semantic Quality Dashboard

Separate from jscpd (literal duplication). SonarCloud handles deep semantic
analysis: code smells, security hotspots, cyclomatic complexity, test coverage
trends.

**CPD exclusions** — mute noisy duplication in areas where jscpd already
guards the gate:

```properties
# sonar-project.properties
sonar.cpd.exclusions=**/*.json,**/locales/**,src/components/ui/**
```

SonarCloud decorates PRs with inline comments on security hotspots and coverage
drops. Acts as the "Quality Gate" that shows pass/fail on each PR.

Already mandated per project setup rules. Config file:
`sonar-project.properties` at repo root.

## Config Files Required Per Type

### Python

`pyproject.toml` additions:

```toml
[tool.ruff]
target-version = "py312"
line-length = 100

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "SIM", "ARG", "PL"]

[tool.pyright]
typeCheckingMode = "strict"

[tool.vulture]
paths = ["src/"]

[tool.pytest.ini_options]
addopts = "--cov --cov-branch --cov-fail-under=90"
```

Pre-commit hook (`.pre-commit-config.yaml`):

```yaml
repos:
  - repo: https://github.com/astral-sh/ty
    rev: v0.0.36
    hooks:
      - id: ty
```

### SvelteKit / JS/TS

`tsconfig.json`:

```json
{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

`package.json` scripts:

```json
{
  "scripts": {
    "check": "svelte-check --tsconfig ./tsconfig.json",
    "knip": "knip",
    "audit": "pnpm audit --audit-level=high"
  }
}
```

### All Projects

`.semgrep.yml` at repo root (minimal starter config):

```yaml
rules:
  - id: no-hardcoded-secrets
    pattern-either:
      - pattern: password = "..."
      - pattern: api_key = "..."
    severity: ERROR
```

## structurelint Rules

### New rule: `linter-config` (revived from removal)

Checks that each project type has its required config files with the correct
settings. This rule was previously removed — it's being revived and enhanced.

For Python projects:
- `pyproject.toml` contains `[tool.pyright]` with `typeCheckingMode = "strict"`
- `pyproject.toml` contains `[tool.ruff]` section
- `.pre-commit-config.yaml` (if exists) contains `ty` hook
- `.semgrep.yml` exists

For SvelteKit projects:
- `tsconfig.json` has `"strict": true`
- `package.json` has `"check"` script
- `.semgrep.yml` exists

For JS/TS projects:
- `tsconfig.json` has `"strict": true` (if tsconfig exists)
- `.semgrep.yml` exists

### New rule: `ci-gates`

Ensures `.github/workflows/pr-gate.yml` and `.github/workflows/merge-gate.yml`
exist and contain the required jobs for the project's language.

Checks:
- Gate files exist
- Each required gate has a job
- `runs-on` uses org runner labels
- Merge gate includes E2E for SvelteKit

### Preset updates

Update existing `sveltekit` and `python-monorepo` presets to include the new
rules.

## Mutation Gate Details

### Python (mutmut)

Uses `tj-actions/changed-files@v44` to detect changed `.py` files, then passes
them to `mutmut run --paths-to-mutate <csv>`.

Threshold: mutation score ≥ 80% on changed code.
If score drops below 80%, the gate fails.

Format:

```yaml
- name: Mutation test
  if: steps.changed-files.outputs.any_changed == 'true'
  env:
    MUT_TEST_RESULT: ${{ steps.mutation.outputs.score }}
  run: |
    mutmut run --paths-to-mutate ${{ steps.changed-files.outputs.all_changed_files }}
    # parse output for score threshold
```

### SvelteKit/JS/TS (Stryker)

Same pattern with `--mutate` flag instead of `--paths-to-mutate`.

Config file: `stryker.config.json` (per-project, generated by structurelint
if missing).

## Project-Specific Notes

### scriptforge (polyglot — Python + JS)

Both Python and JS/TS gates apply. The `package.json` gates run on the JS
portion, `pyproject.toml` gates on the Python portion.

### timelog, mac-control-mcp

Default branch is `feat/quality-gates` not `main` or `master`. Both `pr-gate.yml`
and `merge-gate.yml` must match the actual default branch name.

### lean-runtime-analysis, PPSN_FOGA_GECCO

Limited applicability — only gates 2 (lint), 4 (build), 8 (semgrep) where
tooling exists. Format and type-check gates are skipped.

## Implementation Order

1. structurelint: add `linter-config` rule (config file validation)
2. structurelint: add `ci-gates` rule (workflow existence + job validation)
3. structurelint: update presets
4. Generate config files for each project type
5. Generate pr-gate.yml + merge-gate.yml per project (parallel per repo)
6. Push all changes, verify CI passes on next PR
