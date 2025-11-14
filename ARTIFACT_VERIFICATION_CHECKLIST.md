# Artifact Upload Verification Checklist

Use this checklist to verify the artifact upload feature is working correctly.

## 🎯 Quick Access

**GitHub Actions Page**: https://github.com/Jonathangadeaharder/structurelint/actions

**Expected Workflow Run**: Look for "DEMO: Add failing test to trigger CI artifact upload"

---

## ✅ Verification Steps

### Step 1: Confirm Workflow Started
- [ ] Go to GitHub Actions page
- [ ] See workflow run with commit message "DEMO: Add failing test..."
- [ ] Workflow status is "In progress" or "Failed" (not "Success")
- [ ] Timestamp shows recent run (within last few minutes)

### Step 2: Check Job Execution
- [ ] Click on the workflow run
- [ ] See 5 jobs listed: test, lint, complexity, build, self-lint
- [ ] "test" job has a **red X** (failed) ✅ Expected!
- [ ] Other jobs may pass or fail (doesn't matter for demo)

### Step 3: Verify Artifact Creation
- [ ] Scroll to **bottom** of workflow run page
- [ ] See "Artifacts" section
- [ ] See artifact named: **test-logs-Linux** (or test-logs-ubuntu-latest)
- [ ] Artifact shows size (e.g., "1.2 KB")
- [ ] Artifact shows retention: "7 days"

### Step 4: Download and Inspect Artifact
- [ ] Click on "test-logs-Linux" artifact name
- [ ] ZIP file downloads (test-logs-Linux.zip)
- [ ] Extract ZIP file
- [ ] See file named: **test.log** inside
- [ ] Open test.log in text editor

### Step 5: Verify Log Contents
The test.log file should contain:

- [ ] Test output header (go test -v -race ...)
- [ ] Our intentional failure message:
  ```
  🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!
  ```
- [ ] Second failure message:
  ```
  Check the 'Artifacts' section at the bottom of the workflow run to download test.log
  ```
- [ ] Third failure message:
  ```
  The artifact will contain this complete output plus all other test results
  ```
- [ ] Test summary (e.g., "FAIL github.com/structurelint/structurelint/internal/metrics")
- [ ] Complete stack traces and details

### Step 6: Verify Other Features

**Test other artifacts (if available)**:
- [ ] Check if lint-logs-Linux exists (if lint failed)
- [ ] Check if build-logs exist (if build failed)
- [ ] Download and verify they contain expected content

**Check retention**:
- [ ] Artifact shows "Will be deleted after 7 days"
- [ ] This matches our `retention-days: 7` configuration

**Check cross-platform**:
- [ ] If build job ran on multiple OS, check for:
  - build-logs-ubuntu-latest
  - build-logs-macos-latest
  - build-logs-windows-latest
- [ ] Windows logs should use Tee-Object (PowerShell syntax)

---

## 🎨 Visual Guide

### What You Should See:

#### 1. Actions Page
```
┌─────────────────────────────────────────────────────┐
│ All workflows                                       │
├─────────────────────────────────────────────────────┤
│ ❌ CI · DEMO: Add failing test... · 2 minutes ago  │
│ ❌ CI · Add automatic log artifact... · 1 hour ago │
│ ✅ CI · Fix all PR review issues · 2 days ago      │
└─────────────────────────────────────────────────────┘
```

#### 2. Workflow Run Page (Top)
```
┌─────────────────────────────────────────────────────┐
│ CI                                                  │
│ ❌ Failed in 2m 34s                                │
│                                                     │
│ Jobs:                                               │
│   ❌ test (ubuntu-latest)          1m 23s          │
│   ✅ lint (ubuntu-latest)          45s             │
│   ✅ complexity (ubuntu-latest)    12s             │
│   ✅ build (ubuntu-latest)         1m 2s           │
│   ✅ self-lint (ubuntu-latest)     34s             │
└─────────────────────────────────────────────────────┘
```

#### 3. Workflow Run Page (Bottom - THE IMPORTANT PART!)
```
┌─────────────────────────────────────────────────────┐
│ Artifacts                                           │
│ Produced during runtime                             │
├─────────────────────────────────────────────────────┤
│ 📦 test-logs-Linux                                  │
│    1.2 KB · Will be deleted after 7 days           │
│    [ Download ]                                     │
└─────────────────────────────────────────────────────┘
```

#### 4. Downloaded Artifact Contents
```
test-logs-Linux.zip
  └── test.log  (the complete log file!)
```

#### 5. test.log Contents (Example)
```
=== RUN   TestCognitiveComplexity_Simple
--- PASS: TestCognitiveComplexity_Simple (0.00s)
=== RUN   TestCognitiveComplexity_SingleIf
--- PASS: TestCognitiveComplexity_SingleIf (0.00s)
...
=== RUN   TestArtifactDemo_IntentionalFailure
    artifact_test_demo.go:9: 🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!
    artifact_test_demo.go:10: Check the 'Artifacts' section at the bottom of the workflow run to download test.log
    artifact_test_demo.go:11: The artifact will contain this complete output plus all other test results
--- FAIL: TestArtifactDemo_IntentionalFailure (0.00s)
...
FAIL
FAIL    github.com/structurelint/structurelint/internal/metrics 0.123s
```

---

## ❌ Common Issues & Solutions

### Issue 1: No Artifacts Section
**Problem**: Artifacts section doesn't appear at bottom

**Possible Causes**:
1. ❌ Workflow still running → Wait for completion
2. ❌ All jobs passed → Expected! Artifacts only on failure
3. ❌ `if: failure()` not working → Check workflow syntax

**Solution**: Ensure test job actually failed (red X)

---

### Issue 2: Artifact Empty or Missing
**Problem**: Artifact exists but log file is empty

**Possible Causes**:
1. ❌ `tee` command failed
2. ❌ File path incorrect
3. ❌ Command never ran

**Solution**: Check workflow logs for `tee` output errors

---

### Issue 3: Cannot Download Artifact
**Problem**: Click artifact but nothing downloads

**Possible Causes**:
1. ❌ Browser blocking download
2. ❌ Artifact expired (> 7 days old)
3. ❌ Permissions issue

**Solution**:
- Try different browser
- Use GitHub CLI: `gh run download <run-id>`
- Check artifact age

---

## 🎓 Understanding the Results

### What "Success" Means:

✅ **Artifact Created**: System works! Logs are being captured
✅ **Complete Output**: Log file has all test details
✅ **Automatic Upload**: No manual intervention needed
✅ **Conditional Trigger**: Only failed jobs create artifacts

### What Each File Contains:

| Artifact | File | Contents |
|----------|------|----------|
| test-logs-Linux | test.log | Full `go test -v -race` output |
| lint-logs-Linux | lint.log | golangci-lint issues and errors |
| build-logs-ubuntu | build.log | Go build output and compile errors |
| complexity-logs-Linux | complexity.log | gocyclo analysis results |
| self-lint-logs-Linux | Multiple | Build, run, and verify logs |

### Retention Policy:

- **CI Artifacts**: 7 days (quick debugging)
- **Release Artifacts**: 30 days (production issues)
- **Maximum**: 90 days (GitHub limit)

---

## 🧹 After Verification

Once you've confirmed the artifact system works:

### Option 1: Automatic Cleanup (Recommended)
```bash
cd /home/user/structurelint
./cleanup_demo.sh
```

### Option 2: Manual Cleanup
```bash
# Remove demo files
git rm internal/metrics/artifact_test_demo.go

# Revert CI workflow
# Edit .github/workflows/ci.yml
# Remove the line: - claude/software-quality-metrics-framework-01WfgH2HcdXAPN1gViMK79QV

# Commit
git add .github/workflows/ci.yml
git commit -m "Cleanup: Revert artifact demo"
git push
```

### Option 3: Revert Commit
```bash
git revert HEAD
git push
```

---

## 📊 Success Criteria Summary

Mark this checklist complete when:

- [x] Workflow triggered automatically on push
- [x] Test job failed as expected
- [x] Artifact appeared in Artifacts section
- [x] Downloaded artifact contains test.log
- [x] test.log has complete output including failures
- [x] Artifact retention shows 7 days
- [x] Can download and extract artifact successfully
- [x] Log file is readable and complete

**If all boxes checked**: ✅ **Artifact Upload System is WORKING!**

---

## 🚀 Real-World Usage

After cleanup, the system continues to work automatically:

1. ✅ Any PR that fails tests → Artifact created
2. ✅ Any build that fails → Artifact created
3. ✅ Any lint errors → Artifact created
4. ✅ Download logs for easy debugging
5. ✅ No artifacts for successful builds (efficient!)

**No configuration needed** - it just works! 🎉

---

## 📞 Need Help?

If artifacts aren't appearing:

1. Check workflow file syntax
2. Verify `if: failure()` is present
3. Confirm `tee` command ran
4. Look for upload-artifact step in logs
5. Check GitHub Actions quota/limits

The system is designed to be automatic and reliable.
If it's not working, double-check the workflow syntax in `.github/workflows/ci.yml`.
