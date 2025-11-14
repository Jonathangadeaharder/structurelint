# CI Log Artifacts - Automatic Failure Log Capture

## Overview

Our GitHub Actions workflows automatically capture and save logs as artifacts when jobs fail. This makes debugging CI failures much easier by preserving the full output of failed steps.

## How It Works

### The Pattern

Each critical step in our workflows follows this pattern:

1. **Capture Output**: Use `tee` to write output to both console AND a log file
   ```bash
   command 2>&1 | tee logfile.log
   ```

2. **Upload on Failure**: Add an artifact upload step with `if: failure()` condition
   ```yaml
   - name: Upload logs on failure
     if: failure()
     uses: actions/upload-artifact@v4
     with:
       name: descriptive-name
       path: logfile.log
       retention-days: 7
   ```

### Why `tee` Instead of Redirect?

```bash
# ❌ BAD: Output only goes to file, not visible in CI logs
./script.sh > build.log 2>&1

# ✅ GOOD: Output goes to BOTH console AND file
./script.sh 2>&1 | tee build.log
```

**Benefits of `tee`**:
- ✅ You can still see the output in real-time in GitHub Actions UI
- ✅ The log file is created for artifact upload
- ✅ Best of both worlds: immediate feedback + detailed logs

## Artifacts by Job

### CI Workflow (`.github/workflows/ci.yml`)

#### Test Job
- **Artifact Name**: `test-logs-{OS}` (e.g., `test-logs-Linux`)
- **File**: `test.log`
- **Contains**: Full test output including race detector and coverage
- **Retention**: 7 days

#### Lint Job
- **Artifact Name**: `lint-logs-{OS}`
- **File**: `lint.log`
- **Contains**: golangci-lint output with all linting issues
- **Retention**: 7 days

#### Complexity Job
- **Artifact Name**: `complexity-logs-{OS}`
- **File**: `complexity.log`
- **Contains**: gocyclo complexity analysis output
- **Retention**: 7 days

#### Build Job (Multi-OS Matrix)
- **Artifact Name**: `build-logs-{matrix.os}` (e.g., `build-logs-ubuntu-latest`)
- **File**: `build.log`
- **Contains**: Go build output for the specific OS
- **Retention**: 7 days

#### Self-Check Job (Multi-OS Matrix)
- **Artifact Name**: `self-check-logs-{matrix.os}`
- **File**: `structurelint-self-check.log`
- **Contains**: Output from running structurelint on its own codebase
- **Retention**: 7 days
- **Note**: Uses `Tee-Object` on Windows PowerShell

#### Self-Lint Job (Dogfooding)
- **Artifact Name**: `self-lint-logs-{OS}`
- **Files**:
  - `self-lint-build.log` - Build output
  - `self-lint-run.log` - First run output
  - `self-lint-verify.log` - Verification run output
- **Retention**: 7 days

### Release Workflow (`.github/workflows/release.yml`)

#### Test Step
- **Artifact Name**: `release-test-logs`
- **File**: `release-test.log`
- **Contains**: Test suite output before release
- **Retention**: 30 days (longer for releases)

#### Multi-Platform Build
- **Artifact Name**: `release-build-logs`
- **File**: `release-build.log`
- **Contains**:
  - Linux amd64 build output
  - Linux arm64 build output
  - macOS amd64 build output
  - macOS arm64 build output
  - Windows amd64 build output
  - Checksum generation output
- **Retention**: 30 days (longer for releases)

## Accessing Artifacts

### Via GitHub UI

1. **Navigate to the failed workflow run**:
   - Go to the "Actions" tab in your repository
   - Click on the failed workflow run

2. **Download artifacts**:
   - Scroll to the bottom of the workflow run page
   - Look for the "Artifacts" section
   - Click on the artifact name to download a ZIP file

### Via GitHub CLI

```bash
# List artifacts for a run
gh run view <run-id>

# Download specific artifact
gh run download <run-id> -n artifact-name

# Download all artifacts
gh run download <run-id>
```

### Via API

```bash
# Get artifacts for a workflow run
curl -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/repos/OWNER/REPO/actions/runs/RUN_ID/artifacts

# Download artifact
curl -L -H "Authorization: token YOUR_TOKEN" \
  -o artifact.zip \
  DOWNLOAD_URL
```

## Retention Policies

| Workflow Type | Retention Period | Rationale |
|--------------|------------------|-----------|
| **CI (PRs/Commits)** | 7 days | Short-lived debugging; issues should be fixed quickly |
| **Releases** | 30 days | Longer retention for production build troubleshooting |

**Note**: GitHub has a maximum artifact retention of 90 days on free plans.

## Platform-Specific Considerations

### Linux/macOS
```bash
# Standard tee usage
command 2>&1 | tee output.log
```

### Windows (PowerShell)
```powershell
# Use Tee-Object cmdlet
command 2>&1 | Tee-Object -FilePath output.log
```

Our workflows automatically detect the OS and use the appropriate command:

```yaml
- name: Run on Windows
  run: command 2>&1 | Tee-Object -FilePath log.log
  if: runner.os == 'Windows'
  shell: pwsh

- name: Run on Unix
  run: command 2>&1 | tee log.log
  if: runner.os != 'Windows'
```

## Troubleshooting Common Issues

### Artifact Not Created

**Problem**: The `if: failure()` step runs but no artifact is uploaded.

**Cause**: The log file wasn't created (command failed too early).

**Solution**: Use `tee -a` to append, or check that the file is created before the failure point:

```yaml
- name: Create log directory
  run: mkdir -p logs

- name: Run command
  run: command 2>&1 | tee logs/output.log || true

- name: Upload logs (even on failure)
  if: always()  # Or failure()
  uses: actions/upload-artifact@v4
  with:
    name: logs
    path: logs/
```

### Artifact Too Large

**Problem**: Artifact upload fails due to size limits.

**Cause**: Log files exceed GitHub's 500MB artifact size limit.

**Solution**: Compress logs or filter output:

```bash
# Compress before upload
gzip build.log

# Or filter verbose output
command 2>&1 | grep -v "DEBUG" | tee build.log
```

### Missing Output in Artifacts

**Problem**: Artifact contains only partial output.

**Cause**: Output buffering or process killed mid-stream.

**Solution**: Ensure output is flushed:

```bash
# Force unbuffered output
stdbuf -oL -eL command 2>&1 | tee output.log

# Or for Python scripts
python -u script.py 2>&1 | tee output.log
```

### Windows Encoding Issues

**Problem**: Log files have garbled characters on Windows.

**Cause**: PowerShell encoding mismatch.

**Solution**: Set UTF-8 encoding:

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
command 2>&1 | Tee-Object -FilePath output.log
```

## Best Practices

### 1. Use Descriptive Artifact Names

```yaml
# ❌ BAD: Generic names make it hard to find the right log
name: logs

# ✅ GOOD: Specific names with context
name: test-logs-${{ runner.os }}-${{ matrix.go-version }}
```

### 2. Group Related Logs

```yaml
# Upload multiple related files in one artifact
- name: Upload all build logs
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: build-artifacts
    path: |
      build.log
      test.log
      lint.log
```

### 3. Set Appropriate Retention

```yaml
# Short retention for frequent CI runs
retention-days: 7

# Longer retention for releases or important builds
retention-days: 30

# Maximum retention (default)
retention-days: 90
```

### 4. Add Timestamps to Logs

```bash
# Prefix each line with timestamp
command 2>&1 | ts '[%Y-%m-%d %H:%M:%S]' | tee output.log

# Or use a wrapper
{
  echo "=== Started at $(date -u +"%Y-%m-%d %H:%M:%S UTC") ==="
  command 2>&1
  echo "=== Finished at $(date -u +"%Y-%m-%d %H:%M:%S UTC") ==="
} | tee output.log
```

### 5. Sanitize Sensitive Data

```bash
# Filter out secrets before logging
command 2>&1 | sed 's/password=[^[:space:]]*/password=****/g' | tee output.log

# Or use environment variables
export GITHUB_TOKEN=***  # GitHub Actions does this automatically
```

## Examples from Our Workflows

### Example 1: Test Logs with Coverage

```yaml
- name: Run tests
  run: |
    go test -v -race -coverprofile=coverage.out ./... 2>&1 | tee test.log

- name: Upload test logs on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: test-logs-${{ runner.os }}
    path: test.log
    retention-days: 7
```

**What happens**:
1. Tests run with verbose output, race detector, and coverage
2. Output goes to both console (for real-time viewing) and `test.log`
3. If tests fail, `test.log` is uploaded as an artifact
4. Artifact is kept for 7 days

### Example 2: Multi-Platform Build Logs

```yaml
- name: Build
  run: |
    go build -v ./cmd/structurelint 2>&1 | tee build.log

- name: Upload build logs on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: build-logs-${{ matrix.os }}
    path: build.log
    retention-days: 7
```

**What happens**:
1. Build runs on multiple OS (via matrix strategy)
2. Each OS gets its own `build.log`
3. If build fails on any OS, that OS's log is uploaded
4. Artifact name includes the OS for easy identification

### Example 3: Complex Multi-File Upload

```yaml
- name: Run structurelint on itself
  run: |
    ./structurelint . 2>&1 | tee self-lint-run.log

- name: Verify no violations
  run: |
    if ./structurelint . | tee self-lint-verify.log | grep -q "violation"; then
      echo "structurelint found violations in its own codebase!"
      exit 1
    fi

- name: Upload self-lint logs on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: self-lint-logs-${{ runner.os }}
    path: |
      self-lint-build.log
      self-lint-run.log
      self-lint-verify.log
    retention-days: 7
```

**What happens**:
1. Multiple steps create different log files
2. All related logs are uploaded together in one artifact
3. Easy to correlate the full story of what went wrong

## Advanced Patterns

### Conditional Artifact Upload

```yaml
# Upload on failure OR when manually requested
- name: Upload logs
  if: failure() || github.event_name == 'workflow_dispatch'
  uses: actions/upload-artifact@v4
  with:
    name: logs
    path: output.log
```

### Upload Even on Success (for debugging)

```yaml
# Always upload, useful for debugging flaky tests
- name: Upload logs
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: debug-logs-${{ github.run_number }}
    path: debug.log
    retention-days: 3
```

### Incremental Logs

```yaml
# Append to log file across multiple steps
- name: Step 1
  run: echo "Step 1" | tee -a full.log

- name: Step 2
  run: echo "Step 2" | tee -a full.log

- name: Upload complete log
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: full-workflow-log
    path: full.log
```

## Related Documentation

- [GitHub Actions: Storing workflow data as artifacts](https://docs.github.com/en/actions/using-workflows/storing-workflow-data-as-artifacts)
- [GitHub Actions: Download artifacts](https://docs.github.com/en/actions/managing-workflow-runs/downloading-workflow-artifacts)
- [actions/upload-artifact](https://github.com/actions/upload-artifact)
- [Artifact and log retention policy](https://docs.github.com/en/organizations/managing-organization-settings/configuring-the-retention-period-for-github-actions-artifacts-and-logs-in-your-organization)

## Summary

Our CI workflows now automatically:
- ✅ Capture full output from all critical steps
- ✅ Upload logs only when failures occur (`if: failure()`)
- ✅ Use descriptive names with OS/platform info
- ✅ Set appropriate retention periods (7 days for CI, 30 for releases)
- ✅ Work cross-platform (Linux, macOS, Windows)
- ✅ Preserve logs for debugging without cluttering successful builds

This makes debugging CI failures **significantly easier** - just download the artifact and you have the complete log file!
