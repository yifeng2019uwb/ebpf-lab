# Pre-Commit Testing Guide for cilium/ebpf

**Principle:** ALL tests must pass before commit. Quality > speed.

---

## Overview

The cilium/ebpf CI pipeline runs 4 major test categories. We must run ALL of them locally before committing.

---

## Step 1: Format & Lint Checks ⏳

These catch code style issues, imports, and static analysis problems.

### 1.1 Staticcheck
```bash
go tool staticcheck ./...
```
**What it checks:** Code logic errors, unused variables, type issues
**Expected:** No errors
**If fails:** Fix the reported issues

### 1.2 Golangci-lint
```bash
golangci-lint run ./...
```
**What it checks:**
- goimports (import ordering/formatting)
- Various linters
- Code style compliance

**Expected:** No errors
**If fails:** Fix formatting, import order, style issues

### 1.3 Build Verification
```bash
go build -v ./...
```
**What it checks:** Code compiles
**Expected:** Successful build
**If fails:** Syntax errors or missing dependencies

---

## Step 2: Code Generation & Formatting ⏳

Some packages need code generation. We must ensure generated code is up-to-date.

### 2.1 Generate Code
```bash
make clean && make container-all
```
**What it does:** Regenerates code from templates/specs
**Expected:** No new changes to git
**If fails:** Generated files are out of date

### 2.2 Check for Diffs
```bash
git diff --exit-code
```
**What it checks:** No uncommitted changes from generation
**Expected:** Clean (no output)
**If fails:** Run `make` and commit generated files

### 2.3 Apply Go Fixes
```bash
./scripts/go-fix.sh
```
**What it does:** Applies standard Go fixes (imports, formatting)
**Expected:** No new changes
**If fails:** Check what changed and understand why

### 2.4 Verify Clean Again
```bash
git diff --exit-code
```
**Expected:** Clean (no output)

---

## Step 3: Unit Tests ⏳

Core functionality testing. This is where pre-existing failures show up.

### 3.1 Run Full Test Suite
```bash
go test -race -timeout 5m -short -count 1 ./... 2>&1 | tee test-results.log
```

**Flags explained:**
- `-race` — detect race conditions
- `-timeout 5m` — 5 minute timeout per test
- `-short` — skip long-running tests
- `-count 1` — run each test once
- `./...` — all packages

**Expected:** 
- Most tests pass
- Note any failures (pre-existing vs new)

### 3.2 Document Failures
```bash
# Show summary
grep -E "^(PASS|FAIL|ok|FAIL)" test-results.log | tail -20

# Count results
grep "^ok" test-results.log | wc -l  # passed
grep "^FAIL" test-results.log | wc -l  # failed
```

**Acceptable failures:**
- Pre-existing map test failures (TestMapBatch, TestPerfEventArrayCompatible)
- Tests that fail in upstream too

**NOT acceptable:**
- NEW failures caused by our changes
- Tests that passed before, fail now

---

## Step 4: Real-World Tests ⏳

Verify the actual bug fix works as intended.

### 4.1 Run Issue-Specific Tests
For issue #1497, run the experiment:
```bash
cd ~/workspace/ebpf-lab/experiments/issue-1497
sudo ./issue-1497
```

**Expected:** 20/20 validation tests passing

### 4.2 Verify the Fix
- ✅ Probe programs attach to correct link function
- ✅ Wrong attachment attempts are rejected
- ✅ Cross-domain attachment is prevented
- ✅ Pin/FD programs show appropriate warnings

---

## Full Testing Command Sequence

Run this sequence in order. Each step must pass before proceeding:

```bash
cd ~/workspace/ebpf

# Step 1: Format & Lint
echo "=== STEP 1: Staticcheck ==="
go tool staticcheck ./...
[ $? -eq 0 ] || exit 1

echo "=== STEP 2: Golangci-lint ==="
golangci-lint run ./...
[ $? -eq 0 ] || exit 1

echo "=== STEP 3: Build ==="
go build -v ./...
[ $? -eq 0 ] || exit 1

# Step 2: Code Generation
echo "=== STEP 4: Code Generation ==="
make clean && make container-all
[ $? -eq 0 ] || exit 1

git diff --exit-code
[ $? -eq 0 ] || (echo "Generated files changed"; exit 1)

echo "=== STEP 5: Go Fixes ==="
./scripts/go-fix.sh
[ $? -eq 0 ] || exit 1

git diff --exit-code
[ $? -eq 0 ] || (echo "Formatting changed"; exit 1)

# Step 3: Unit Tests
echo "=== STEP 6: Unit Tests ==="
go test -race -timeout 5m -short -count 1 ./... 2>&1 | tee test-results.log
# Check results manually

# Step 4: Real-World Tests
echo "=== STEP 7: Issue-Specific Tests ==="
cd ~/workspace/ebpf-lab/experiments/issue-1497
sudo ./issue-1497

echo "=== ALL TESTS COMPLETE ==="
```

---

## Interpreting Test Results

### Green Light (OK to commit)
- ✅ Staticcheck: 0 errors
- ✅ Golangci-lint: 0 errors
- ✅ Build: success
- ✅ Code generation: no changes
- ✅ Go fixes: no changes
- ✅ Unit tests: only pre-existing failures
- ✅ Issue-specific tests: all passing

### Red Light (Fix required)
- ❌ NEW errors in staticcheck/golangci-lint
- ❌ Build fails
- ❌ Code generation makes changes
- ❌ Go fixes make changes
- ❌ NEW test failures (didn't fail before)
- ❌ Issue-specific tests failing

---

## Pre-Existing Failures to Document

These are known issues in upstream (NOT our responsibility):

### Map Test Failures
- **TestMapBatch/PerCPUHash** — per-CPU API change not reflected in tests
- **TestPerfEventArrayCompatible** — unclear cause, pre-existing
- **Status:** Pre-existing upstream bug (PR #xxxx should fix)

### Goimports Cache Issues
- golangci-lint sometimes reports false goimports errors
- **Workaround:** `rm -rf ~/.cache/golangci-lint` and retry

---

## Commit Only When

All four steps are complete AND:
1. ✅ Format & lint checks pass
2. ✅ Code generation is clean
3. ✅ ALL unit tests result documented (pre-existing failures noted)
4. ✅ Issue-specific tests pass

**Then and ONLY THEN proceed to commit.**

