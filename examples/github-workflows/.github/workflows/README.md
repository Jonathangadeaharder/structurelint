# GitHub Actions Workflows

## Overview

Complete example workflows for CI/CD pipelines that satisfy structurelint's github-workflows rule.

## Workflows

### test.yml
Comprehensive testing workflow with:
- Matrix testing across Go versions and operating systems
- Unit and integration test execution
- Coverage reporting and threshold enforcement
- Benchmark testing

### security.yml
Security scanning workflow with:
- CodeQL static analysis
- Go vulnerability scanning (govulncheck)
- Secret scanning with Trivy
- SAST with Gosec

### quality.yml
Code quality enforcement with:
- Linting with golangci-lint
- Format checking with gofmt
- Static analysis with staticcheck
- Coverage threshold validation
- Complexity analysis with gocyclo

## Usage

Copy these workflows to your `.github/workflows/` directory to implement comprehensive CI/CD automation.
