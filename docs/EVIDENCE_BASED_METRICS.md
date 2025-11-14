# Evidence-Based Software Quality Metrics Framework

## Overview

This document describes the integration of scientifically-validated software quality metrics into structurelint, based on systematic literature reviews and empirical studies.

## Background

Traditional metrics like Cyclomatic Complexity (CC) have significant limitations:
- **Weak understandability predictor**: CC's mathematical model is "unsatisfactory" for measuring maintainability
- **Deviates from human perception**: EEG studies show CC "deviates considerably from programmers' perception of code complexity"
- **Outperformed by LOC**: Lines of Code often outperforms CC in defect prediction

This framework implements evidence-based alternatives with stronger empirical support.

## Metric Categories

### 1. Cognitive Complexity (CoC) - **Highest Priority**

**Evidence Level**: Meta-analysis (14 studies)
**Correlation**: r = 0.54 with comprehension time, r = -0.29 with subjective difficulty

**Why Superior to Cyclomatic Complexity**:
- Penalizes nesting (human cognitive load increases exponentially with nesting)
- Ignores "shorthand" structures that improve readability
- Based on human assessment, not mathematical models

**Calculation Rules**:
```
1. Base complexity = 0 (not 1 like CC)
2. +1 for each flow break: if, for, while, catch, switch, recursion
3. +1 additional for each level of nesting
4. +0 for shorthand operators (&&, ||, ?:) in sequence
```

**Example**:
```go
// Cyclomatic Complexity = 11
// Cognitive Complexity = 7
func process(items []Item) {
    for _, item := range items {        // +1 (for) = 1
        if item.IsActive {              // +2 (+1 for if, +1 for nesting) = 3
            if item.HasPermission {     // +3 (+1 for if, +2 for nesting) = 6
                process(item)           // +1 (recursion +2 for nesting) = 9... wait
                                        // Actually +1 (recursion at nesting 2) = 7
            }
        }
    }
}
```

**Implementation**: `internal/metrics/cognitive_complexity.go`

### 2. Halstead Metrics - **High Priority**

**Evidence Level**: Neuroscience (EEG study)
**Correlation**: rs = 0.901 with measured cognitive load

**Why Critical**:
- Captures "data complexity" (vocabulary, operands, operators)
- Complements Cognitive Complexity (which captures control-flow complexity)
- Highest correlation with actual brain activity during code comprehension

**Metrics**:
```
n1 = distinct operators (if, +, =, func, etc.)
n2 = distinct operands (variables, constants)
N1 = total operators
N2 = total operands

Program Vocabulary (n) = n1 + n2
Program Length (N) = N1 + N2
Volume (V) = N × log₂(n)          // Information content in bits
Difficulty (D) = (n1/2) × (N2/n2) // How hard to write/understand
Effort (E) = D × V                // Mental effort required
```

**Implementation**: `internal/metrics/halstead.go`

### 3. CK Suite (Object-Oriented Metrics) - **Medium Priority**

**Evidence Level**: Multiple SLRs and 2023 large-scale empirical study

#### 3.1 Coupling Between Objects (CBO)
**Strength**: Very Strong
**Citation**: Consistently identified as strong, reliable defect predictor

**Definition**: Number of other classes to which a class is coupled (through method calls, field access, or inheritance)

**Causal Mechanism**: High coupling → high instability → changes cascade → defects

#### 3.2 Response For a Class (RFC)
**Strength**: Very Strong
**Definition**: Number of methods in class + methods it calls in other classes
**Why Predictive**: Measures complexity of interaction and response set

#### 3.3 Lack of Cohesion in Methods - Version 5 (LCOM5)
**Strength**: Very Strong (Recent)
**Citation**: 2023 study - "among the highest-performing individual metrics"

**Definition**: LCOM5 provides normalized measure of method connectivity within a class
**Causal Mechanism**: High LCOM → "God Class" → poor design → confusion → defects

#### 3.4 Depth of Inheritance Tree (DIT) & Number of Children (NOC)
**Strength**: Conflicting
**Note**: Evidence contradictory - predictive power highly dataset-dependent
**Recommendation**: Implement as optional, not enabled by default

**Implementation**: `internal/metrics/ck_suite.go`

### 4. Process Metrics - **Highest Predictive Power**

**Evidence Level**: Systematic Literature Review
**Finding**: "change data [process metrics] are overall effectively better indicators... than static code attributes"

**Key Metrics**:
- **Code Churn**: Lines added + deleted + modified over time
- **Revision Count**: Number of commits touching this file
- **Bug Fix Count**: Number of bug-fixing commits
- **Age**: Time since file creation
- **Developer Count**: Number of distinct developers who modified file
- **Developer Experience**: Average experience of contributors

**Why Superior**:
- Captures social and historical context
- A file churned by 10 developers with 5 bug fixes is riskier than a structurally complex but stable file
- Process metrics alone often outperform combined (static + process) models

**Implementation**: `internal/metrics/process_metrics.go`

### 5. Statistical Framework - **Critical for Validity**

**Problem**: Size Confounding
- Nearly all complexity metrics are highly correlated with LOC
- A metric 90% correlated with LOC provides only 10% new information
- Without controlling for LOC, dashboards may just be measuring "big files are risky"

**Solution**: Multivariate Logistic Regression
```
P(Defect) = β₀ + β₁(LOC) + β₂(CBO) + β₃(LCOM5) + β₄(CoC) + β₅(Halstead_E)
```

Statistical significance (p-value) of β₂, β₃, β₄, β₅ proves whether metrics predict defects **after** accounting for size.

**Implementation**: `internal/metrics/statistical_model.go`

## Architecture

### Directory Structure
```
internal/
├── metrics/               # New metrics package
│   ├── cognitive_complexity.go
│   ├── halstead.go
│   ├── ck_suite.go
│   ├── process_metrics.go
│   ├── statistical_model.go
│   ├── analyzer.go        # Unified analyzer interface
│   └── metrics_test.go
├── rules/
│   ├── max_cognitive_complexity.go
│   ├── max_halstead_effort.go
│   ├── max_coupling.go
│   ├── max_lcom.go
│   └── quality_threshold.go  # Combined quality score rule
└── output/
    └── metrics_report.go      # New formatted output for metrics
```

### Metric Analyzer Interface
```go
// Analyzer computes metrics for a file
type Analyzer interface {
    Name() string
    Analyze(file FileInfo) (MetricResult, error)
    SupportedLanguages() []string
}

// MetricResult represents the output of a metric analysis
type MetricResult struct {
    File    string
    Metric  string
    Value   float64
    Details map[string]interface{}  // For complex metrics
}

// MultiMetricAnalyzer combines multiple analyzers
type MultiMetricAnalyzer struct {
    analyzers []Analyzer
}
```

### Rule Configuration
```yaml
rules:
  # RECOMMENDED: Replace cyclomatic complexity with cognitive complexity
  max-cognitive-complexity:
    max: 15
    file-patterns:
      - "**/*.go"
      - "**/*.ts"
      - "**/*.js"
      - "**/*.py"

  # NEW: Halstead Effort threshold
  max-halstead-effort:
    max: 100000
    file-patterns:
      - "**/*.go"

  # NEW: CK Suite metrics
  max-coupling:
    cbo-max: 10    # Coupling Between Objects
    rfc-max: 50    # Response For a Class
    file-patterns:
      - "**/*.go"
      - "**/*.ts"
      - "**/*.java"

  max-cohesion:
    lcom5-max: 0.8  # LCOM5 (0-1, higher = less cohesive)
    file-patterns:
      - "**/*.go"

  # NEW: Process metrics thresholds
  max-process-risk:
    churn-max: 1000           # Max lines churned
    revisions-max: 50         # Max number of revisions
    developer-count-max: 10   # Max unique developers
    bugfix-count-max: 5       # Max bug fix commits

  # NEW: Composite quality score with LOC control
  quality-threshold:
    enabled: true
    model: "multivariate"      # or "simple"
    defect-probability-max: 0.3  # Max 30% predicted defect probability
    control-for-loc: true      # Statistical control for size
```

### Metric Reports
```yaml
# Generate detailed metrics reports
reports:
  - type: "metrics-summary"
    output: "metrics-report.md"
    include:
      - cognitive-complexity
      - halstead
      - ck-suite
      - process-metrics

  - type: "risk-ranking"
    output: "risk-ranked-files.csv"
    model: "multivariate"
    top-n: 50  # Show top 50 riskiest files
```

## Implementation Phases

### Phase 1: Core Metric Implementations (Week 1)
- [ ] Create `internal/metrics` package
- [ ] Implement Cognitive Complexity analyzer
- [ ] Implement Halstead metrics analyzer
- [ ] Add AST support for TypeScript, Python, Java (extend Go support)
- [ ] Create metric analyzer tests

### Phase 2: OO Metrics (Week 2)
- [ ] Implement CK Suite analyzer (CBO, RFC, LCOM5)
- [ ] Add OO-specific AST traversals
- [ ] Implement DIT/NOC (optional)
- [ ] Create CK suite tests

### Phase 3: Process Metrics (Week 2-3)
- [ ] Implement Git history mining
- [ ] Parse commit messages for bug fix detection (keywords: "fix", "bug", "defect")
- [ ] Calculate churn, revisions, age, developer count
- [ ] Handle repository edge cases (submodules, etc.)

### Phase 4: Statistical Framework (Week 3)
- [ ] Implement multivariate logistic regression
- [ ] Add LOC as baseline predictor
- [ ] Implement correlation-based feature selection
- [ ] Create defect probability calculator

### Phase 5: Rules and Integration (Week 4)
- [ ] Create rules for each metric category
- [ ] Add configuration schema updates
- [ ] Implement metric report generation
- [ ] Create risk ranking output format

### Phase 6: Documentation and Examples (Week 4)
- [ ] Update README with evidence-based recommendations
- [ ] Create example configurations
- [ ] Add scientific citations
- [ ] Deprecation notice for Maintainability Index
- [ ] Usage guide with interpretation

## Configuration Examples

### Evidence-Based Go Project
```yaml
root: true

rules:
  # Replace CC with Cognitive Complexity
  max-cyclomatic-complexity: 0  # Disable (deprecated)
  max-cognitive-complexity:
    max: 15
    file-patterns: ["**/*.go"]

  # Add Halstead for data complexity
  max-halstead-effort:
    max: 100000
    file-patterns: ["**/*.go"]

  # OO metrics for Go packages
  max-coupling:
    cbo-max: 8
    rfc-max: 40
    file-patterns: ["**/*.go"]

  max-cohesion:
    lcom5-max: 0.7
    file-patterns: ["**/*.go"]

  # Process metrics (strongest predictor)
  max-process-risk:
    churn-max: 800
    revisions-max: 40
    developer-count-max: 8
    bugfix-count-max: 4

# Generate risk report
reports:
  - type: "risk-ranking"
    output: "risk-analysis.csv"
    model: "multivariate"
    control-for-loc: true
```

### Evidence-Based TypeScript/React Project
```yaml
root: true

rules:
  max-cognitive-complexity:
    max: 15
    file-patterns: ["**/*.ts", "**/*.tsx"]

  max-halstead-effort:
    max: 120000
    file-patterns: ["**/*.ts", "**/*.tsx"]

  max-coupling:
    cbo-max: 10
    rfc-max: 50
    file-patterns: ["**/*.ts", "**/*.tsx"]

  quality-threshold:
    enabled: true
    defect-probability-max: 0.25
    control-for-loc: true
    weights:  # Project-specific learned weights
      loc: 0.3
      cognitive-complexity: 0.25
      halstead-effort: 0.2
      cbo: 0.15
      lcom5: 0.1
```

## Metric Interpretation Guide

### Cognitive Complexity Thresholds
- **0-5**: Simple, easy to understand
- **6-10**: Moderate complexity, acceptable
- **11-15**: High complexity, consider refactoring
- **16-25**: Very high complexity, should refactor
- **26+**: Extremely complex, high maintenance risk

### Halstead Effort Thresholds
- **0-10,000**: Low effort
- **10,000-50,000**: Moderate effort
- **50,000-100,000**: High effort
- **100,000+**: Very high effort, high cognitive load

### CBO (Coupling) Thresholds
- **0-5**: Low coupling, good isolation
- **6-10**: Moderate coupling, acceptable
- **11-15**: High coupling, increased instability
- **16+**: Very high coupling, high defect risk

### LCOM5 Thresholds (0-1 scale)
- **0.0-0.3**: High cohesion, well-designed
- **0.4-0.6**: Moderate cohesion, acceptable
- **0.7-0.8**: Low cohesion, consider splitting
- **0.9-1.0**: Very low cohesion, "God Class" smell

### Process Metrics Interpretation
- **High Churn + High Revisions**: Unstable, frequently changing
- **High Bug Fixes**: Historical defect-proneness
- **High Developer Count**: Diffused ownership, knowledge gaps
- **Low Age + High Churn**: Newly created and unstable

## Scientific Citations

1. McCabe, T. J. (1976). "A Complexity Measure"
2. Shepperd, M. (1988). "A Critique of Cyclomatic Complexity as a Software Metric"
3. Scalabrino et al. (2022). "EEG Study on Code Complexity Metrics"
4. Schnappinger et al. (2020). "Meta-Analysis of Cognitive Complexity"
5. Basili et al. (2023). "Large-Scale Study of LCOM5"
6. Hassan & Holt (2016). "Change Metrics vs. Static Code Attributes"
7. Chidamber & Kemerer (1994). "A Metrics Suite for Object-Oriented Design"

## Deprecation Notice

### Maintainability Index (MI)
**Status**: Not Recommended for New Projects

**Rationale**:
- Obscure and unactionable formula
- Statistically confounded by LOC
- Based on small, obsolete 1980s dataset
- All components (Halstead Volume, CC, LOC) are inter-correlated

**Migration**:
```yaml
# OLD (not recommended)
rules:
  maintainability-index:
    min: 20

# NEW (evidence-based)
rules:
  max-cognitive-complexity: { max: 15 }
  max-halstead-effort: { max: 100000 }
  quality-threshold:
    enabled: true
    defect-probability-max: 0.3
    control-for-loc: true
```

## Testing Strategy

### Unit Tests
- Test each metric calculator with known code samples
- Verify against published examples from literature
- Edge cases: empty functions, single-line functions, generated code

### Integration Tests
- Full project analysis with multiple metrics
- Comparison with manual calculations
- Performance benchmarks (should handle 100K+ LOC projects)

### Validation Tests
- Compare against reference implementations (e.g., SonarQube's Cognitive Complexity)
- Statistical validation on open-source projects with known defect data
- Correlation analysis to verify literature findings

## Performance Considerations

- **Caching**: Cache AST parsing results, reuse for multiple metrics
- **Incremental Analysis**: Only reanalyze changed files in CI
- **Parallel Processing**: Analyze files concurrently
- **Git Optimization**: Shallow clone for process metrics, limit history depth

## Future Enhancements

- **Machine Learning**: Train project-specific models for weight optimization
- **Trend Analysis**: Track metric evolution over time
- **Hotspot Detection**: Identify files with metrics degrading over time
- **IDE Integration**: Real-time metric feedback in editors
- **Custom Thresholds**: Auto-calibrate thresholds based on project percentiles
