# Expected Artifact Contents - Reference

This file shows exactly what you should see when you download and extract the `test-logs-Linux.zip` artifact.

## 📦 Artifact Structure

```
test-logs-Linux.zip
└── test.log
```

## 📄 test.log Contents (Example)

The `test.log` file will contain output similar to this:

```
=== RUN   TestCognitiveComplexity_Simple
--- PASS: TestCognitiveComplexity_Simple (0.00s)
=== RUN   TestCognitiveComplexity_SingleIf
--- PASS: TestCognitiveComplexity_SingleIf (0.00s)
=== RUN   TestCognitiveComplexity_NestedIf
--- PASS: TestCognitiveComplexity_NestedIf (0.00s)
=== RUN   TestCognitiveComplexity_ForLoop
--- PASS: TestCognitiveComplexity_ForLoop (0.00s)
=== RUN   TestCognitiveComplexity_NestedLoopAndIf
--- PASS: TestCognitiveComplexity_NestedLoopAndIf (0.00s)
=== RUN   TestCognitiveComplexity_Switch
--- PASS: TestCognitiveComplexity_Switch (0.00s)
=== RUN   TestCognitiveComplexity_ElseIf
--- PASS: TestCognitiveComplexity_ElseIf (0.00s)
=== RUN   TestCognitiveComplexity_DeeplyNested
--- PASS: TestCognitiveComplexity_DeeplyNested (0.00s)
=== RUN   TestCognitiveComplexity_BranchStatements
--- PASS: TestCognitiveComplexity_BranchStatements (0.00s)
=== RUN   TestCognitiveComplexityAnalyzer_AnalyzeFile
--- PASS: TestCognitiveComplexityAnalyzer_AnalyzeFile (0.00s)

=== RUN   TestHalstead_Simple
--- PASS: TestHalstead_Simple (0.00s)
=== RUN   TestHalstead_IfStatement
--- PASS: TestHalstead_IfStatement (0.00s)
=== RUN   TestHalstead_Loop
--- PASS: TestHalstead_Loop (0.00s)
=== RUN   TestHalstead_ComplexFunction
--- PASS: TestHalstead_ComplexFunction (0.00s)
=== RUN   TestHalstead_BinaryOperators
--- PASS: TestHalstead_BinaryOperators (0.00s)
=== RUN   TestHalsteadAnalyzer_AnalyzeFile
--- PASS: TestHalsteadAnalyzer_AnalyzeFile (0.00s)

=== RUN   TestArtifactDemo_IntentionalFailure
    artifact_test_demo.go:9: 🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!
    artifact_test_demo.go:10: Check the 'Artifacts' section at the bottom of the workflow run to download test.log
    artifact_test_demo.go:11: The artifact will contain this complete output plus all other test results
--- FAIL: TestArtifactDemo_IntentionalFailure (0.00s)

FAIL
coverage: 85.7% of statements
FAIL    github.com/structurelint/structurelint/internal/metrics  0.234s
FAIL
```

## 🔍 What to Look For

### ✅ Success Indicators:

1. **File Exists**: `test.log` is present in the ZIP
2. **Complete Output**: All test results are shown
3. **Intentional Failure**: Our demo test failure appears:
   ```
   🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!
   ```
4. **Test Details**: Full stack trace and error details
5. **Coverage Info**: Coverage percentage (if enabled)
6. **Final Result**: `FAIL` status at the end

### 📊 Key Sections Explained:

#### Header Section (Not shown in example)
```
# Usually starts with:
go test -v -race -coverprofile=coverage.out ./... 2>&1
```

#### Test Execution
```
=== RUN   TestName
    Details...
--- PASS: TestName (0.00s)
```
or
```
=== RUN   TestName
    test_file.go:123: Error message here
--- FAIL: TestName (0.00s)
```

#### Our Demo Failure
```
=== RUN   TestArtifactDemo_IntentionalFailure
    artifact_test_demo.go:9: 🎯 DEMO: This intentional test failure...
    artifact_test_demo.go:10: Check the 'Artifacts' section...
    artifact_test_demo.go:11: The artifact will contain this complete output...
--- FAIL: TestArtifactDemo_IntentionalFailure (0.00s)
```

#### Summary
```
FAIL
coverage: 85.7% of statements
FAIL    github.com/structurelint/structurelint/internal/metrics  0.234s
FAIL
```

## 📏 File Size

Typical sizes:
- **Small test suite**: 1-5 KB
- **Medium test suite**: 5-50 KB
- **Large test suite**: 50-500 KB
- **Very large with race detector**: 500 KB - 2 MB

Our demo should be approximately **1-3 KB**.

## 🎨 Formatting

The log file is **plain text**, formatted for terminal output:
- ANSI color codes may be present (or stripped)
- UTF-8 encoding
- Unix line endings (LF)
- Readable in any text editor

## 🔎 How to Inspect

### Using Terminal:
```bash
# Extract
unzip test-logs-Linux.zip

# View entire file
cat test.log

# View with pagination
less test.log

# Search for our demo failure
grep "DEMO" test.log

# Count test results
grep -c "PASS" test.log
grep -c "FAIL" test.log

# View only failures
grep -A 5 "FAIL" test.log
```

### Using GUI:
- Open with any text editor (VS Code, Notepad++, etc.)
- Search for "DEMO" to find our intentional failure
- Look for "FAIL" to see all failures

## ✅ Verification Checklist

When you open `test.log`, verify:

- [ ] File is not empty
- [ ] Contains test output (=== RUN markers)
- [ ] Shows our demo failure with emoji: 🎯 DEMO
- [ ] Has complete error messages
- [ ] Ends with FAIL status
- [ ] Total file size is reasonable (1-10 KB expected)

## 🎯 The Point of This Demo

This demonstrates that:

1. ✅ **Logs are captured** - Complete test output is saved
2. ✅ **Upload works** - Artifact appears in GitHub UI
3. ✅ **Downloadable** - You can get the ZIP file
4. ✅ **Complete data** - Nothing is lost or truncated
5. ✅ **Automatic** - No manual steps needed

In real usage:
- Any genuine test failure → Artifact created
- Download log → Debug offline
- Complete context → Faster fixes

## 🔄 Real-World Example

### Genuine Failure Scenario:

```
=== RUN   TestUserAuthentication
    auth_test.go:42: Expected user ID 123, got 456
    auth_test.go:43: Authentication failed for user "john@example.com"
    auth_test.go:44: Database connection: OK
    auth_test.go:45: Cache status: MISS
--- FAIL: TestUserAuthentication (0.12s)
```

With artifacts:
1. ✅ Download `test-logs-Linux.zip`
2. ✅ See complete failure details
3. ✅ Have exact error messages
4. ✅ Know what to fix
5. ✅ No need to re-run CI for logs

## 📝 Summary

The artifact will contain:
- ✅ **Complete** test output
- ✅ **Readable** plain text format
- ✅ **Detailed** error messages
- ✅ **Searchable** with standard tools
- ✅ **Downloadable** via GitHub UI or CLI

This is exactly what you need to debug failures quickly and efficiently!

---

**Next**: After verifying the artifact, run `./cleanup_demo.sh` to remove the demo test and restore normal CI operation.
