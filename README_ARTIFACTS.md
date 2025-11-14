# 🎯 Complete Guide: CI Artifact Upload Demo

## 🚀 What's Happening Right Now

Your GitHub Actions CI workflow is **currently running** with an intentional test failure to demonstrate the automatic artifact upload feature.

### ⏱️ Timeline (approximately)

```
[00:00] ✅ Push successful
[00:30] 🔄 CI workflow starts
[01:00] 🏃 Tests begin running
[01:30] ❌ Demo test fails
[01:35] 📦 Artifact uploaded
[02:00] ✅ Workflow completes (with failures)
[02:00] 🎁 Artifacts ready for download!
```

**Current Status**: Go check now! → https://github.com/Jonathangadeaharder/structurelint/actions

---

## 📋 Quick Start: See Your Artifacts in 60 Seconds

### Step 1: Open Actions Page
Click here: **https://github.com/Jonathangadeaharder/structurelint/actions**

### Step 2: Find Your Workflow Run
Look for: **"DEMO: Add failing test to trigger CI artifact upload"**

### Step 3: Click It
Click on that workflow run to open details

### Step 4: Scroll Down
Scroll all the way to the **BOTTOM** of the page

### Step 5: See Artifacts! 🎉
You should see:
```
┌─────────────────────────────────────────────┐
│ Artifacts                                   │
│ Produced during runtime                     │
├─────────────────────────────────────────────┤
│ 📦 test-logs-Linux                          │
│    1.2 KB · Will be deleted after 7 days   │
│    [ Download ]                             │
└─────────────────────────────────────────────┘
```

### Step 6: Download
Click on **"test-logs-Linux"** to download the ZIP file

### Step 7: Extract and View
```bash
unzip test-logs-Linux.zip
cat test.log
```

You should see our intentional failure message:
```
🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!
```

---

## 🎓 Understanding What You're Seeing

### The Workflow Structure

```
CI Workflow
├── test (❌ Will FAIL - intentional)
│   └── Creates: test-logs-Linux artifact
├── lint (✅ Should PASS)
├── complexity (✅ Should PASS)
├── build (✅ Should PASS)
└── self-lint (✅ Should PASS)
```

### Why Test Fails

We added this code temporarily:
```go
func TestArtifactDemo_IntentionalFailure(t *testing.T) {
    t.Error("🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!")
    // ... more error messages ...
}
```

### The Upload Trigger

In `.github/workflows/ci.yml`:
```yaml
- name: Run tests
  run: |
    go test -v -race -coverprofile=coverage.out ./... 2>&1 | tee test.log

- name: Upload test logs on failure
  if: failure()  # ← Only runs if previous step failed!
  uses: actions/upload-artifact@v4
  with:
    name: test-logs-Linux
    path: test.log
    retention-days: 7
```

**Magic happens here**:
1. `tee test.log` saves output to file while showing in console
2. `if: failure()` only triggers when test fails
3. `actions/upload-artifact@v4` uploads the file
4. Artifact appears at bottom of workflow run

---

## 📊 What Each File Does

### 📁 Helper Scripts

| File | Purpose | Usage |
|------|---------|-------|
| `check_ci_artifacts.sh` | Detailed status guide | `./check_ci_artifacts.sh` |
| `cleanup_demo.sh` | Cleanup after demo | `./cleanup_demo.sh` |

### 📄 Documentation

| File | Purpose |
|------|---------|
| `ARTIFACT_VERIFICATION_CHECKLIST.md` | Step-by-step verification |
| `EXPECTED_ARTIFACT_EXAMPLE.md` | Shows expected log contents |
| `docs/ARTIFACT_DEMO_GUIDE.md` | Complete demo instructions |
| `docs/CI_LOG_ARTIFACTS.md` | Full technical documentation |

### 🧪 Demo Files (Temporary)

| File | Purpose | Status |
|------|---------|--------|
| `internal/metrics/artifact_test_demo.go` | Failing test | Remove after demo |
| `.github/workflows/ci.yml` (modified) | Branch trigger | Revert after demo |

---

## ✅ Verification Checklist

Use this to confirm everything works:

### Basic Verification
- [ ] Workflow run appears in Actions tab
- [ ] Test job shows red X (failed)
- [ ] Artifacts section visible at bottom
- [ ] Can click and download artifact
- [ ] ZIP file contains test.log
- [ ] Log file is readable and complete

### Detailed Verification
- [ ] Log shows our demo failure with emoji 🎯
- [ ] All three error messages are present
- [ ] File size is reasonable (1-10 KB)
- [ ] Retention shows "7 days"
- [ ] Artifact name includes OS (test-logs-Linux)

### Advanced Verification
- [ ] Other jobs (lint, build) show their status
- [ ] If other jobs failed, their artifacts exist too
- [ ] Can download multiple artifacts
- [ ] Logs match what was shown in console output

---

## 🎨 Visual Guide

### What You'll See (Step by Step)

#### 1. Actions Page
```
Recent workflow runs
┌────────────────────────────────────────────────┐
│ ❌ CI · DEMO: Add failing test... · 2 min ago │ ← Click this!
└────────────────────────────────────────────────┘
```

#### 2. Workflow Run Details (Top)
```
CI
❌ Failed in 2m 15s

Jobs
├── ❌ test (ubuntu-latest)          1m 23s ← Failed!
├── ✅ lint (ubuntu-latest)          45s
├── ✅ complexity (ubuntu-latest)    12s
├── ✅ build (ubuntu-latest)         58s
└── ✅ self-lint (ubuntu-latest)     34s
```

#### 3. Artifacts Section (BOTTOM of page)
```
Artifacts                               👈 SCROLL DOWN TO HERE!
Produced during runtime
┌──────────────────────────────────────────────┐
│ 📦 test-logs-Linux                           │
│    1.2 KB · Will be deleted after 7 days    │
│    [Download] ←────────────────── Click!     │
└──────────────────────────────────────────────┘
```

#### 4. Downloaded File
```
Downloads/
└── test-logs-Linux.zip  ← Extract this
    └── test.log         ← Read this!
```

#### 5. Log File Contents
```bash
$ cat test.log

=== RUN   TestArtifactDemo_IntentionalFailure
    artifact_test_demo.go:9: 🎯 DEMO: This intentional test failure...
    artifact_test_demo.go:10: Check the 'Artifacts' section...
    artifact_test_demo.go:11: The artifact will contain this complete output...
--- FAIL: TestArtifactDemo_IntentionalFailure (0.00s)
FAIL
```

---

## 🧹 Cleanup (After Verification)

### ⚠️ Important: Remove Demo After Testing!

The demo includes an intentional failure that will cause CI to fail.
Once you've verified artifacts work, clean up:

### Option 1: Automated (Recommended)
```bash
cd /home/user/structurelint
./cleanup_demo.sh
```

This will:
1. ✅ Remove the failing test file
2. ✅ Revert CI workflow to original
3. ✅ Clean up demo scripts
4. ✅ Commit and push changes

### Option 2: Manual Revert
```bash
git revert HEAD
git push
```

This reverts the entire demo commit.

### Option 3: Manual Cleanup
```bash
# Remove demo test
git rm internal/metrics/artifact_test_demo.go

# Edit .github/workflows/ci.yml
# Remove line: - claude/software-quality-metrics-framework-01WfgH2HcdXAPN1gViMK79QV

# Commit
git add .github/workflows/ci.yml
git commit -m "Cleanup: Remove artifact demo"
git push
```

---

## 🎯 Real-World Usage After Cleanup

Once demo is cleaned up, the artifact system continues working automatically:

### Scenario 1: Real Test Failure
```yaml
Developer pushes code with bug
    ↓
CI runs tests
    ↓
Test fails (genuine failure)
    ↓
Artifact automatically created: test-logs-Linux
    ↓
Developer downloads artifact
    ↓
Complete logs available for debugging
    ↓
Bug fixed!
```

### Scenario 2: Build Failure
```yaml
Code doesn't compile
    ↓
Build job fails
    ↓
Artifact created: build-logs-ubuntu-latest
    ↓
Download to see exact compiler errors
    ↓
Fix compilation issue
```

### Scenario 3: Lint Errors
```yaml
Code violates linting rules
    ↓
Lint job fails
    ↓
Artifact created: lint-logs-Linux
    ↓
Download to see all violations
    ↓
Fix linting issues
```

### No Manual Steps Needed!

✅ Just download the artifact
✅ All logs are complete
✅ Debug offline with full context
✅ No need to re-run CI just to see logs

---

## 🔍 Troubleshooting

### Problem: No Artifacts Section

**Likely Cause**: Workflow still running or all tests passed

**Solution**:
1. Wait for workflow to complete
2. Check if test job actually failed (red X)
3. Artifacts only appear on failure!

### Problem: Can't Download Artifact

**Likely Cause**: Browser blocking or permissions

**Solution**:
```bash
# Use GitHub CLI instead
gh run list --limit 1
gh run download <run-id>
```

### Problem: Artifact Empty

**Likely Cause**: Log file wasn't created

**Solution**: Check workflow logs for errors in the `tee` command

### Problem: Workflow Didn't Trigger

**Likely Cause**: Branch not in trigger list

**Solution**: We already added it! Should work automatically.
If not, check `.github/workflows/ci.yml` line 8.

---

## 📚 Further Reading

- **Full Documentation**: `docs/CI_LOG_ARTIFACTS.md`
- **Demo Guide**: `docs/ARTIFACT_DEMO_GUIDE.md`
- **Verification Steps**: `ARTIFACT_VERIFICATION_CHECKLIST.md`
- **Expected Contents**: `EXPECTED_ARTIFACT_EXAMPLE.md`

---

## 🎉 Success Criteria

Mark the demo as successful when you can:

✅ See the workflow run in GitHub Actions
✅ Confirm test job failed (red X)
✅ Find Artifacts section at bottom of page
✅ Download test-logs-Linux.zip
✅ Extract and read test.log
✅ See our demo failure messages
✅ Understand how it works
✅ Clean up demo files successfully

---

## 🚀 What Happens Next

After cleanup:

1. ✅ CI triggers only on main/master pushes or PRs
2. ✅ Tests pass normally (no intentional failures)
3. ✅ Artifact system remains active
4. ✅ Future real failures → automatic artifacts
5. ✅ No configuration needed

**The artifact upload feature is permanent and production-ready!**

It will automatically capture logs whenever:
- Tests fail
- Builds fail
- Linting fails
- Any job fails

No manual intervention needed. Just download and debug! 🎊

---

## 📞 Quick Reference

| Need | Command/Link |
|------|-------------|
| **View workflow** | https://github.com/Jonathangadeaharder/structurelint/actions |
| **Status guide** | `./check_ci_artifacts.sh` |
| **Cleanup** | `./cleanup_demo.sh` |
| **Verification** | Read `ARTIFACT_VERIFICATION_CHECKLIST.md` |
| **Expected output** | Read `EXPECTED_ARTIFACT_EXAMPLE.md` |

---

**🎯 Current Status**: CI should be running NOW!

**👉 Go check**: https://github.com/Jonathangadeaharder/structurelint/actions

Have fun exploring the artifact system! 🚀
