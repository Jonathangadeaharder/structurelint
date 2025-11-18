# Multi-Language Extension Plan: C#, C++, and Java

## Executive Summary

This plan outlines the steps to extend structurelint's functionality to fully support C#, C++, and Java across all phases. Currently, these languages have partial support:

- **Java**: Test validation only (Tier 2) + Clone detection (Tier 3)
- **C++**: Test validation only (Tier 2) + Clone detection (Tier 3)
- **C#**: Clone detection only (Tier 3)

**Goal**: Elevate all three languages to **Tier 1** (Full Support) with import/export analysis and metrics.

---

## Current Support Matrix

| Feature | Java | C++ | C# | Required For |
|---------|------|-----|-----|--------------|
| **Phase 0: Filesystem Linting** | ✅ | ✅ | ✅ | Universal |
| **Phase 1: Architectural Layers** | ❌ | ❌ | ❌ | Import parsing |
| **Phase 2: Dead Code Detection** | ❌ | ❌ | ❌ | Import/export parsing |
| **Phase 3: Test Validation** | ✅ | ✅ | ❌ | Test patterns |
| **Phase 4: File Content Templates** | ✅ | ✅ | ✅ | Universal |
| **Phase 5: Quality Metrics** | ❌ | ❌ | ❌ | Metrics scripts |
| **Phase 8: GitHub Workflows** | ✅ | ✅ | ✅ | Universal |
| **Clone Detection** | ✅ | ✅ | ✅ | Already supported |

---

## Implementation Plan

### Task 1: Extend Test Pattern Support for C#

**Files to modify:**
- `internal/init/detector.go`

**Changes:**

1. **Add C# extension mapping** (line 152):
```go
".cs": "csharp",
```

2. **Add C# test patterns** (line 189):
```go
"csharp": {"Test", "Tests", ".test"},  // FooTest.cs, FooTests.cs, Foo.test.cs
```

3. **Add C# source patterns** (line 205):
```go
"csharp": {"**/*.cs"},
```

**Testing:**
- Create test fixtures with C# test files
- Verify test pattern detection for:
  - `CalculatorTest.cs` (suffix pattern)
  - `Calculator.test.cs` (infix pattern)
  - Adjacent vs separate test directory detection

---

### Task 2: Implement Java Import/Export Parsing

**Files to modify:**
- `internal/parser/parser.go`

**Changes:**

1. **Add Java case to ParseFile()** (line 42):
```go
case ".java":
    return p.parseJava(filePath)
```

2. **Add Java case to ParseExports()** (line 59):
```go
case ".java":
    return p.parseJavaExports(filePath)
```

3. **Implement parseJava() function**:
```go
// parseJava extracts imports from Java files
func (p *Parser) parseJava(filePath string) ([]Import, error) {
    // Patterns to match:
    // import com.example.MyClass;
    // import com.example.*;
    // import static com.example.MyClass.staticMethod;
    // Ignore: package declarations

    importRegex := regexp.MustCompile(`^\s*import\s+(?:static\s+)?([a-zA-Z0-9_.]+)(?:\.\*)?;`)

    // Parse line by line
    // Extract package paths
    // Determine if relative (same package prefix)
}
```

4. **Implement parseJavaExports() function**:
```go
// parseJavaExports extracts public classes/interfaces/methods from Java
func (p *Parser) parseJavaExports(filePath string) ([]Export, error) {
    // Patterns to match:
    // public class ClassName
    // public interface InterfaceName
    // public enum EnumName
    // public static void methodName()

    classRegex := regexp.MustCompile(`^\s*public\s+(?:class|interface|enum)\s+(\w+)`)
    methodRegex := regexp.MustCompile(`^\s*public\s+.*?\s+(\w+)\s*\(`)
}
```

**Java Import Resolution Rules:**
- Absolute imports: `com.example.MyClass` → `com/example/MyClass.java`
- Wildcard imports: `com.example.*` → all classes in `com/example/`
- Same package = relative import
- Different package = external/library import

**Testing:**
- Create test fixtures with various Java import styles
- Test static imports
- Test wildcard imports
- Test package-relative detection

---

### Task 3: Implement C++ Import/Export Parsing

**Files to modify:**
- `internal/parser/parser.go`

**Changes:**

1. **Add C++ cases to ParseFile()** (line 42):
```go
case ".cpp", ".cc", ".cxx", ".hpp", ".h":
    return p.parseCpp(filePath)
```

2. **Add C++ cases to ParseExports()** (line 59):
```go
case ".cpp", ".cc", ".cxx", ".hpp", ".h":
    return p.parseCppExports(filePath)
```

3. **Implement parseCpp() function**:
```go
// parseCpp extracts includes from C++ files
func (p *Parser) parseCpp(filePath string) ([]Import, error) {
    // Patterns to match:
    // #include "myheader.h"       (relative, local)
    // #include <iostream>         (system/library)
    // #include <boost/algorithm.hpp>  (external library)

    includeRegex := regexp.MustCompile(`^\s*#include\s+[<"]([^>"]+)[>"]`)

    // Quote includes ("") = relative/local
    // Angle bracket includes (<>) = system/external
}
```

4. **Implement parseCppExports() function**:
```go
// parseCppExports extracts public symbols from C++ headers
func (p *Parser) parseCppExports(filePath string) ([]Export, error) {
    // Only parse header files (.h, .hpp)
    // Patterns to match:
    // class ClassName
    // struct StructName
    // namespace NamespaceName
    // extern declarations
    // Function declarations

    // Note: Only parse headers, not .cpp implementation files
}
```

**C++ Import Resolution Rules:**
- `#include "header.h"` → relative import, search local directory first
- `#include <header>` → system/library import, search include paths
- Path resolution: `#include "subdir/header.h"` → resolve relative to current file

**Challenges:**
- Header guards don't affect imports
- Preprocessor directives may conditionally include files
- Forward declarations vs includes
- **Recommendation**: Start with simple regex, note limitations in docs

**Testing:**
- Test quoted includes (relative)
- Test angle bracket includes (system)
- Test subdirectory includes
- Test header-only libraries

---

### Task 4: Implement C# Import/Export Parsing

**Files to modify:**
- `internal/parser/parser.go`

**Changes:**

1. **Add C# case to ParseFile()** (line 42):
```go
case ".cs":
    return p.parseCSharp(filePath)
```

2. **Add C# case to ParseExports()** (line 59):
```go
case ".cs":
    return p.parseCSharpExports(filePath)
```

3. **Implement parseCSharp() function**:
```go
// parseCSharp extracts using statements from C# files
func (p *Parser) parseCSharp(filePath string) ([]Import, error) {
    // Patterns to match:
    // using System;
    // using System.Collections.Generic;
    // using MyNamespace.SubNamespace;
    // using static System.Math;
    // using Alias = System.Text.StringBuilder;

    usingRegex := regexp.MustCompile(`^\s*using\s+(?:static\s+)?(?:\w+\s+=\s+)?([a-zA-Z0-9_.]+);`)

    // Determine relative vs external:
    // - Parse namespace from file
    // - If using starts with same namespace = relative
    // - Otherwise = external/library
}
```

4. **Implement parseCSharpExports() function**:
```go
// parseCSharpExports extracts public types from C# files
func (p *Parser) parseCSharpExports(filePath string) ([]Export, error) {
    // Patterns to match:
    // public class ClassName
    // public interface IInterfaceName
    // public struct StructName
    // public enum EnumName
    // public delegate ...
    // public static void MethodName()

    typeRegex := regexp.MustCompile(`^\s*public\s+(?:class|interface|struct|enum|delegate)\s+(\w+)`)
}
```

**C# Import Resolution Rules:**
- Parse `namespace MyNamespace.SubNamespace { }` from file
- Using statements in same namespace hierarchy = relative
- Using statements in different namespace = external
- Using static = import static members

**Testing:**
- Test standard using statements
- Test using static
- Test using aliases
- Test namespace-relative detection
- Test nested namespaces

---

### Task 5: Implement Java Metrics Support

**Files to create:**
- `internal/metrics/scripts/java_metrics.py`

**Files to modify:**
- `internal/metrics/multilang_analyzer.go`

**Implementation:**

1. **Add Java to language detection** (multilang_analyzer.go, line 60):
```go
".java": "java",
```

2. **Add Java case to AnalyzeFileByPath()** (line 48):
```go
case "java":
    return a.analyzeJavaFile(filePath)
```

3. **Implement analyzeJavaFile() method**:
```go
func (a *MultiLanguageAnalyzer) analyzeJavaFile(filePath string) (FileMetrics, error) {
    scriptPath, err := getScriptPath("java_metrics.py")
    if err != nil {
        return FileMetrics{}, err
    }

    cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
    // ... similar to Python/JS analysis
}
```

4. **Create `java_metrics.py`** (based on `python_metrics.py`):

```python
#!/usr/bin/env python3
"""
Java metrics calculator using tree-sitter-java
Calculates cognitive complexity and Halstead metrics
"""

import sys
import json
from tree_sitter import Language, Parser
import tree_sitter_java

def calculate_cognitive_complexity(tree, source_code):
    """
    Calculate cognitive complexity for Java code

    Increments for:
    - if, else if, ternary operators (+1, +nesting)
    - switch, case (+1, +nesting)
    - for, while, do-while (+1, +nesting)
    - catch (+1, +nesting)
    - break/continue with label (+1)
    - logical operators && || in conditions (+1)
    - recursion (+1)
    - nested functions (+nesting)
    """
    # Implementation similar to python_metrics.py
    # Use tree-sitter queries for Java AST nodes
    pass

def calculate_halstead_metrics(tree, source_code):
    """
    Calculate Halstead metrics for Java code

    Operators: +, -, *, /, %, &&, ||, !, ==, !=, <, >, etc.
    Operands: variables, constants, method calls, literals
    """
    # Implementation similar to python_metrics.py
    pass

# Main execution
if __name__ == "__main__":
    metric_type = sys.argv[1]  # "cognitive-complexity" or "halstead"
    file_path = sys.argv[2]

    # Parse Java file
    parser = Parser()
    parser.set_language(Language(tree_sitter_java.language()))

    # Calculate metrics
    # Output JSON: {"functions": [...], "file_summary": {...}}
```

**Dependencies:**
- `tree-sitter` Python package
- `tree-sitter-java` grammar

**Testing:**
- Test with various Java control structures
- Test with nested conditionals
- Test with lambdas and streams
- Compare results with known complexity values

---

### Task 6: Implement C++ Metrics Support

**Files to create:**
- `internal/metrics/scripts/cpp_metrics.py`

**Files to modify:**
- `internal/metrics/multilang_analyzer.go`

**Implementation:**

1. **Add C++ to language detection** (multilang_analyzer.go, line 60):
```go
".cpp": "cpp",
".cc":  "cpp",
".cxx": "cpp",
".hpp": "cpp",
".h":   "cpp",
```

2. **Add C++ case to AnalyzeFileByPath()** (line 48):
```go
case "cpp", "c":
    return a.analyzeCppFile(filePath)
```

3. **Implement analyzeCppFile() method**:
```go
func (a *MultiLanguageAnalyzer) analyzeCppFile(filePath string) (FileMetrics, error) {
    scriptPath, err := getScriptPath("cpp_metrics.py")
    if err != nil {
        return FileMetrics{}, err
    }

    cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
    // ... similar to Python/JS analysis
}
```

4. **Create `cpp_metrics.py`** (based on `python_metrics.py`):

```python
#!/usr/bin/env python3
"""
C++ metrics calculator using tree-sitter-cpp
Calculates cognitive complexity and Halstead metrics
"""

import sys
import json
from tree_sitter import Language, Parser
import tree_sitter_cpp

def calculate_cognitive_complexity(tree, source_code):
    """
    Calculate cognitive complexity for C++ code

    Increments for:
    - if, else if, ternary operators (+1, +nesting)
    - switch, case (+1, +nesting)
    - for, while, do-while (+1, +nesting)
    - catch (+1, +nesting)
    - goto statements (+1)
    - logical operators && || in conditions (+1)
    - recursion (+1)
    - nested functions/lambdas (+nesting)
    - template metaprogramming constructs
    """
    # Implementation similar to python_metrics.py
    # Handle C++ specific constructs:
    # - Templates
    # - Lambdas
    # - Range-based for loops
    # - Exception handling
    pass

def calculate_halstead_metrics(tree, source_code):
    """
    Calculate Halstead metrics for C++ code

    Operators: +, -, *, /, %, &&, ||, !, ==, !=, <, >, ->, ::, etc.
    Operands: variables, constants, function calls, literals
    """
    # Implementation similar to python_metrics.py
    pass

# Main execution
```

**Dependencies:**
- `tree-sitter` Python package
- `tree-sitter-cpp` grammar

**Challenges:**
- C++ preprocessor macros affect complexity
- Template metaprogramming complexity
- Header vs implementation file metrics
- **Recommendation**: Focus on implementation files (.cpp), note limitations

**Testing:**
- Test with templates
- Test with lambdas
- Test with modern C++ features (range-for, auto)
- Test with exception handling

---

### Task 7: Implement C# Metrics Support

**Files to create:**
- `internal/metrics/scripts/csharp_metrics.py`

**Files to modify:**
- `internal/metrics/multilang_analyzer.go`

**Implementation:**

1. **Add C# to language detection** (multilang_analyzer.go, line 60):
```go
".cs": "csharp",
```

2. **Add C# case to AnalyzeFileByPath()** (line 48):
```go
case "csharp":
    return a.analyzeCSharpFile(filePath)
```

3. **Implement analyzeCSharpFile() method**:
```go
func (a *MultiLanguageAnalyzer) analyzeCSharpFile(filePath string) (FileMetrics, error) {
    scriptPath, err := getScriptPath("csharp_metrics.py")
    if err != nil {
        return FileMetrics{}, err
    }

    cmd := exec.Command("python3", scriptPath, a.metricType, filePath)
    // ... similar to Python/JS analysis
}
```

4. **Create `csharp_metrics.py`** (based on `python_metrics.py`):

```python
#!/usr/bin/env python3
"""
C# metrics calculator using tree-sitter-c-sharp
Calculates cognitive complexity and Halstead metrics
"""

import sys
import json
from tree_sitter import Language, Parser
import tree_sitter_c_sharp

def calculate_cognitive_complexity(tree, source_code):
    """
    Calculate cognitive complexity for C# code

    Increments for:
    - if, else if, ternary operators (+1, +nesting)
    - switch, case (+1, +nesting)
    - for, foreach, while, do-while (+1, +nesting)
    - catch (+1, +nesting)
    - LINQ query expressions (+1, +nesting)
    - null-coalescing operators ?? (+1)
    - logical operators && || in conditions (+1)
    - recursion (+1)
    - nested functions/lambdas (+nesting)
    - async/await patterns
    """
    # Implementation similar to python_metrics.py
    # Handle C# specific constructs:
    # - Properties
    # - Events
    # - LINQ
    # - Async/await
    # - Pattern matching
    pass

def calculate_halstead_metrics(tree, source_code):
    """
    Calculate Halstead metrics for C# code

    Operators: +, -, *, /, %, &&, ||, !, ==, !=, <, >, ??, ?., etc.
    Operands: variables, constants, method calls, literals, LINQ
    """
    # Implementation similar to python_metrics.py
    pass

# Main execution
```

**Dependencies:**
- `tree-sitter` Python package
- `tree-sitter-c-sharp` grammar

**C# Specific Considerations:**
- LINQ queries add complexity
- Null-coalescing operators (??, ?.)
- Pattern matching (switch expressions)
- Async/await patterns
- Properties vs fields

**Testing:**
- Test with LINQ queries
- Test with async/await
- Test with pattern matching
- Test with null-coalescing operators
- Test with properties and events

---

### Task 8: Update Clone Detection (Verification)

**Files to verify:**
- `clone_detection/clone_detection/parsers/language_configs.py`

**Action:**
Verify that Java, C++, and C# are already configured in the clone detection system.

**Expected configuration:**
```python
LANGUAGE_CONFIGS = {
    "java": LanguageConfig(
        name="java",
        extensions=[".java"],
        grammar_module="tree_sitter_java",
        function_query="...",
    ),
    "cpp": LanguageConfig(
        name="cpp",
        extensions=[".cpp", ".cc", ".cxx", ".h", ".hpp"],
        grammar_module="tree_sitter_cpp",
        function_query="...",
    ),
    "csharp": LanguageConfig(
        name="csharp",
        extensions=[".cs"],
        grammar_module="tree_sitter_c_sharp",
        function_query="...",
    ),
}
```

**If missing:** Add configurations following the pattern of existing languages.

---

### Task 9: Update Documentation

**Files to create/update:**
- `docs/multi-language-support.md` (new)
- `README.md` (update supported languages section)

**Content:**

1. **Update README.md**:
   - Add Java, C++, C# to "Supported Languages" section
   - Update feature matrix showing Tier 1 support
   - Add note about test pattern configuration

2. **Create comprehensive multi-language guide**:
   - Import/export parsing rules for each language
   - Test pattern conventions
   - Metrics calculation methodology
   - Known limitations (regex vs AST parsing)
   - Configuration examples for mixed-language projects

3. **Add migration guide**:
   - For projects upgrading from Tier 2 → Tier 1 support
   - Configuration changes required
   - Breaking changes (if any)

---

### Task 10: Integration Testing

**Files to create:**
- `testdata/java_project/` (test fixtures)
- `testdata/cpp_project/` (test fixtures)
- `testdata/csharp_project/` (test fixtures)
- `testdata/mixed_language_project/` (multi-language)

**Test scenarios:**

1. **Java Project:**
   - Maven/Gradle structure
   - Package-based imports
   - Test adjacency (JUnit conventions)
   - Layer boundaries (Spring MVC structure)
   - Dead code detection
   - Metrics validation

2. **C++ Project:**
   - Header/implementation separation
   - Include paths (relative vs system)
   - Test adjacency (Google Test conventions)
   - Namespace-based organization
   - Metrics validation

3. **C# Project:**
   - .NET project structure
   - Namespace-based imports
   - Test adjacency (xUnit/NUnit conventions)
   - Layer boundaries (Clean Architecture)
   - Metrics validation

4. **Mixed-Language Project:**
   - Go + Java (microservices)
   - TypeScript + C++ (Node.js native modules)
   - Python + C# (IronPython interop)
   - File-pattern configuration
   - No cross-contamination of test patterns

**Test Coverage:**
- Unit tests for each parser
- Integration tests for each rule
- Metrics accuracy validation
- Performance benchmarks

---

## Implementation Order

### Phase 1: Foundation (Week 1)
1. ✅ Task 1: C# test patterns
2. ✅ Task 8: Verify clone detection

### Phase 2: Import/Export Parsing (Week 2-3)
3. ✅ Task 2: Java import/export parsing
4. ✅ Task 3: C++ import/export parsing
5. ✅ Task 4: C# import/export parsing
6. ✅ Integration testing for parsing

### Phase 3: Metrics Support (Week 4-5)
7. ✅ Task 5: Java metrics
8. ✅ Task 6: C++ metrics
9. ✅ Task 7: C# metrics
10. ✅ Integration testing for metrics

### Phase 4: Documentation & Testing (Week 6)
11. ✅ Task 9: Documentation
12. ✅ Task 10: Comprehensive testing
13. ✅ Performance benchmarking

---

## Dependencies & Prerequisites

### Runtime Dependencies:
- Python 3.7+ (for metrics scripts)
- `tree-sitter` Python package
- `tree-sitter-java` grammar
- `tree-sitter-cpp` grammar
- `tree-sitter-c-sharp` grammar

### Development Dependencies:
- Go 1.19+ (existing)
- Test fixtures for each language
- Sample projects for integration testing

### Installation:
```bash
pip3 install tree-sitter tree-sitter-java tree-sitter-cpp tree-sitter-c-sharp
```

---

## Risk Assessment & Mitigation

### Risks:

1. **Regex Parsing Limitations**
   - Risk: Regex-based parsing may miss complex import patterns
   - Mitigation: Document known limitations, consider tree-sitter upgrade path
   - Impact: Medium (affects accuracy of import graph)

2. **C++ Preprocessor Complexity**
   - Risk: Conditional includes, macros affect parsing
   - Mitigation: Start simple, note limitations in docs
   - Impact: Low (most projects use straightforward includes)

3. **Metrics Accuracy**
   - Risk: Tree-sitter parsing may differ from compiler
   - Mitigation: Validate against known test cases, accept ±10% variance
   - Impact: Medium (affects metric reliability)

4. **Cross-Platform Testing**
   - Risk: Path separators, file encoding differences
   - Mitigation: Test on Windows, Linux, macOS
   - Impact: High (affects reliability)

5. **Performance**
   - Risk: External Python scripts slow down analysis
   - Mitigation: Parallel processing, caching, benchmarking
   - Impact: Medium (affects user experience)

### Mitigation Strategies:
- Comprehensive test coverage (>80%)
- Performance benchmarks in CI/CD
- User feedback loop for edge cases
- Incremental rollout (beta testing)

---

## Success Criteria

### Functional Requirements:
- ✅ All three languages have import/export parsing
- ✅ All three languages have metrics calculation
- ✅ Test validation works for all three
- ✅ Layer boundaries enforceable
- ✅ Dead code detection functional
- ✅ Clone detection verified

### Quality Requirements:
- ✅ Test coverage >80%
- ✅ Import parsing accuracy >95%
- ✅ Metrics accuracy within ±10% of manual calculation
- ✅ Performance: <5s for projects with 1000 files
- ✅ Documentation complete and accurate

### User Experience:
- ✅ Auto-detection works for all three languages
- ✅ Error messages are clear and actionable
- ✅ Configuration examples provided
- ✅ Migration path documented

---

## Post-Implementation Tasks

1. **Performance Optimization:**
   - Profile metrics calculation
   - Implement caching for large projects
   - Parallelize file analysis

2. **Tree-Sitter Migration (Future):**
   - Replace regex parsers with tree-sitter
   - Improve parsing accuracy
   - Add type-aware analysis

3. **IDE Integration:**
   - VS Code extension
   - IntelliJ IDEA plugin
   - Visual Studio extension

4. **Community Feedback:**
   - Beta testing program
   - Issue tracker for edge cases
   - Feature request prioritization

---

## Appendix: Code Examples

### Example: Java Import Parsing

**Input (`Calculator.java`):**
```java
package com.example.calculator;

import java.util.List;
import java.util.ArrayList;
import com.example.utils.MathHelper;
import static java.lang.Math.PI;

public class Calculator {
    // ...
}
```

**Expected Output:**
```go
[]Import{
    {SourceFile: "Calculator.java", ImportPath: "java.util.List", IsRelative: false},
    {SourceFile: "Calculator.java", ImportPath: "java.util.ArrayList", IsRelative: false},
    {SourceFile: "Calculator.java", ImportPath: "com.example.utils.MathHelper", IsRelative: true},
    {SourceFile: "Calculator.java", ImportPath: "java.lang.Math", IsRelative: false},
}
```

### Example: C++ Include Parsing

**Input (`calculator.cpp`):**
```cpp
#include <iostream>
#include <vector>
#include "calculator.h"
#include "utils/math_helper.h"

void calculate() {
    // ...
}
```

**Expected Output:**
```go
[]Import{
    {SourceFile: "calculator.cpp", ImportPath: "iostream", IsRelative: false},
    {SourceFile: "calculator.cpp", ImportPath: "vector", IsRelative: false},
    {SourceFile: "calculator.cpp", ImportPath: "calculator.h", IsRelative: true},
    {SourceFile: "calculator.cpp", ImportPath: "utils/math_helper.h", IsRelative: true},
}
```

### Example: C# Using Parsing

**Input (`Calculator.cs`):**
```csharp
using System;
using System.Collections.Generic;
using MyApp.Utils;
using static System.Math;

namespace MyApp.Calculators
{
    public class Calculator
    {
        // ...
    }
}
```

**Expected Output:**
```go
[]Import{
    {SourceFile: "Calculator.cs", ImportPath: "System", IsRelative: false},
    {SourceFile: "Calculator.cs", ImportPath: "System.Collections.Generic", IsRelative: false},
    {SourceFile: "Calculator.cs", ImportPath: "MyApp.Utils", IsRelative: true},
    {SourceFile: "Calculator.cs", ImportPath: "System.Math", IsRelative: false},
}
```

---

## Timeline Summary

| Week | Tasks | Deliverables |
|------|-------|--------------|
| 1 | Foundation setup | C# test patterns, clone detection verified |
| 2-3 | Import/Export parsing | All 3 languages parsing imports/exports |
| 4-5 | Metrics implementation | All 3 languages calculating metrics |
| 6 | Documentation & testing | Complete docs, test suite, benchmarks |

**Total Estimated Effort:** 6 weeks (1 developer full-time)

**Complexity Assessment:**
- Low: Test patterns, clone detection verification (Task 1, 8)
- Medium: Import/export parsing (Task 2-4)
- High: Metrics implementation (Task 5-7)
- Medium: Documentation & testing (Task 9-10)

---

## Conclusion

This plan provides a comprehensive roadmap to extend structurelint's functionality to C#, C++, and Java. The phased approach ensures incremental progress with validation at each step. The modular architecture makes it straightforward to add new languages following established patterns.

**Key Success Factor:** Maintaining consistency with existing Go/Python/TypeScript implementations while respecting language-specific idioms.

**Next Steps:**
1. Review and approve this plan
2. Set up development environment with all dependencies
3. Create feature branch: `feature/multi-language-csharp-cpp-java`
4. Begin with Task 1 (C# test patterns) as proof of concept
5. Iterate with stakeholder feedback
