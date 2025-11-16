# How to Test the Artifact Upload Feature

## Option 1: Create a Pull Request (Recommended)

This will trigger the CI automatically and you can see the artifacts in action.

### Steps:

1. **Go to GitHub Repository**:
   ```
   https://github.com/Jonathangadeaharder/structurelint
   ```

2. **Create Pull Request**:
   - Click "Pull requests" tab
   - Click "New pull request"
   - Select base: `main` (or your default branch)
   - Select compare: `claude/software-quality-metrics-framework-01WfgH2HcdXAPN1gViMK79QV`
   - Click "Create pull request"

3. **Watch CI Run**:
   - The CI workflow will automatically start
   - Click "Actions" tab to watch progress

4. **To Test Artifact Upload** (optional):
   - In the PR, uncomment the failing test in `internal/metrics/artifact_test_demo.go`
   - Commit and push
   - CI will fail and create artifacts

## Option 2: Manual Workflow Trigger

If the workflow has `workflow_dispatch`, you can trigger it manually:

```bash
gh workflow run ci.yml --ref claude/software-quality-metrics-framework-01WfgH2HcdXAPN1gViMK79QV
```

## Option 3: Temporary Branch Trigger (For Testing)

Temporarily modify `.github/workflows/ci.yml` to trigger on your branch:

```yaml
on:
  push:
    branches:
      - main
      - master
      - claude/software-quality-metrics-framework-01WfgH2HcdXAPN1gViMK79QV  # Add this
  pull_request:
    branches: [ main, master ]
```

Then push any change to trigger CI.

## How to Trigger a Failure (For Artifact Testing)

### Method 1: Failing Test

Edit `internal/metrics/artifact_test_demo.go` and uncomment the error line:

```go
func TestArtifactDemo_IntentionalFailure(t *testing.T) {
    t.Error("DEMO: This intentional failure will trigger log artifact upload in CI")
}
```

### Method 2: Syntax Error

Temporarily introduce a syntax error in any Go file:

```go
// Add this invalid syntax to any file
this will cause a build failure
```

### Method 3: Lint Error

Add code that violates linting rules:

```go
func badFunction() {
    unused := "this will trigger unused variable lint error"
}
```

## What to Expect

### When CI Fails:

1. **Workflow Completes with Failure**:
   - Red X next to the workflow run
   - Failed job will be marked

2. **Artifacts Appear**:
   - Scroll to bottom of workflow run page
   - "Artifacts" section appears
   - Download ZIP file with logs

3. **Artifact Contents**:
   - `test.log` - Full test output
   - `lint.log` - Linting errors (if lint failed)
   - `build.log` - Build errors (if build failed)
   - etc.

### Example Artifact Download:

```bash
# Using GitHub CLI
gh run list --limit 1
gh run download <run-id> -n test-logs-Linux

# Extract and view
unzip test-logs-Linux.zip
cat test.log
```

### In GitHub UI:

1. Go to Actions tab
2. Click on failed workflow run
3. Scroll to bottom
4. See "Artifacts" section
5. Click artifact name to download

## Demonstration Scenario

Let's create a complete demonstration:

### Step 1: Enable the Demo Test

```bash
cd /home/user/structurelint

# Edit the demo test file
sed -i 's|// t.Error|t.Error|' internal/metrics/artifact_test_demo.go

# Commit and push
git add internal/metrics/artifact_test_demo.go
git commit -m "Demo: Enable failing test to demonstrate artifacts"
git push
```

### Step 2: Create PR (Triggers CI)

```bash
# Using gh CLI
gh pr create \
  --title "Test: Demonstrate Artifact Upload Feature" \
  --body "This PR intentionally includes a failing test to demonstrate the automatic artifact upload feature on CI failures."

# Or via GitHub UI (easier)
# Just go to: https://github.com/Jonathangadeaharder/structurelint/compare
```

### Step 3: Watch CI Fail

- CI will run automatically
- Test job will fail (due to our intentional error)
- Artifact will be uploaded

### Step 4: Download Artifact

1. Go to failed workflow run
2. Scroll to "Artifacts" section at bottom
3. Click "test-logs-Linux" to download
4. Extract ZIP and open `test.log`
5. See complete test output including our failure

### Step 5: Fix and Re-run

```bash
# Comment out the error again
sed -i 's|t.Error|// t.Error|' internal/metrics/artifact_test_demo.go

git add internal/metrics/artifact_test_demo.go
git commit -m "Fix: Disable demo test, tests should pass now"
git push
```

Now CI will pass and no artifacts will be created (as intended).

## Verifying the Feature Works

### Success Criteria:

✅ **No artifacts when tests pass**
- Successful runs should NOT create artifacts
- This saves storage and keeps things clean

✅ **Artifacts created only on failure**
- Failed runs MUST create artifacts
- Artifact name includes OS info
- Retention period set correctly

✅ **Log files are complete**
- Download and extract artifact
- Log file contains full output
- Output matches what was shown in CI console

✅ **Cross-platform support**
- Works on Linux (uses `tee`)
- Works on macOS (uses `tee`)
- Works on Windows (uses `Tee-Object`)

## Cleanup

After testing, remember to:

1. **Remove demo test file** (optional):
   ```bash
   git rm internal/metrics/artifact_test_demo.go
   git commit -m "Cleanup: Remove artifact demo test"
   ```

2. **Merge or close PR**:
   - If PR was created for testing, close it or merge it

3. **Revert any temporary workflow changes**:
   - Remove any branch triggers added for testing

## Real-World Usage

In production, you don't need to do anything special:

1. ✅ CI runs on every PR automatically
2. ✅ If any job fails, logs are captured
3. ✅ Artifacts appear at bottom of workflow run
4. ✅ Download and debug offline

**That's it!** The artifact system is now fully automated and production-ready.
