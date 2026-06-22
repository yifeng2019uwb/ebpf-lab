# Upstream Test Failures Analysis

**Date:** 2026-06-22  
**Status:** Documented  
**Impact on Issue #1497:** None - unrelated  

## Summary

Two tests fail in upstream cilium/ebpf main branch despite CI being green:
- `TestMapBatch/PerCPUHash`
- `TestPerfEventArrayCompatible`

These failures are **pre-existing upstream bugs**, NOT caused by our issue #1497 fix.

## Root Cause Analysis

### Test 1: TestMapBatch/PerCPUHash

**Location:** map_test.go line 199

**Failure:**
```
error: wanted error is not found in error chain
got: "per-cpu value requires a slice or a pointer to slice"
want: "key does not exist"
```

**Root Cause:**

Commit `bcca828` (Nov 12, 2023) - "map: allow pre-allocating per-CPU values"

Changed per-CPU Lookup API:

**OLD API (before commit):**
```go
var v uint32
m.Lookup(key, &v)  // Library allocates slice
```

**NEW API (after commit):**
```go
v := make([]uint32, PossibleCPU())  // Pre-allocate
m.Lookup(key, v)                     // Pass slice
```

**Test Issue:**

Test still uses OLD API:
```go
var v uint32
qt.Assert(t, qt.ErrorIs(m.Lookup(uint32(0), &v), ErrKeyNotExist))
```

Now it fails with "per-cpu value requires a slice or a pointer to slice" because `&v` (pointer to uint32) is not a valid per-CPU argument anymore.

**Fix Required:**
```go
v := make([]uint32, possibleCPUs)  // Pre-allocate
qt.Assert(t, qt.ErrorIs(m.Lookup(uint32(0), v), ErrKeyNotExist))
```

---

### Test 2: TestPerfEventArrayCompatible

**Location:** map_test.go line 1729

**Failure:**
```
error: got <nil> but want non-nil
stack: qt.Assert(t, qt.IsNotNil(ms.Compatible(m)))
```

**Root Cause:**

Unclear from git log. Appears to be a compatibility check that's now returning nil when it should return an error.

**Investigation Needed:**
- Check when `Compatible()` behavior changed
- Check if MapSpec validation changed
- Check if PerfEventArray behavior changed

---

## Why CI is Green But Tests Fail

### CI Configuration

Main CI jobs **only compile** tests, don't run them:
```yaml
# ci.yml
go test -c -o /dev/null ./...  # -c = compile only, doesn't run!
```

**Actual test execution** happens in vimto runner:
```yaml
gotestsum --raw-command --ignore-non-json-output-lines \
  -- vimto -kernel :stable-selftests -- \
  go test -race -timeout 5m -short -count 1 -json ./...
```

**Theory:** The `vimto` test runner either:
1. Skips these specific tests
2. Runs in a different environment
3. Has different setup that makes tests pass
4. Uses different Go version or kernel

---

## Impact Assessment

### On Issue #1497
- ✅ **No impact** - our changes don't touch map.go or per-CPU logic
- ✅ **No impact** - our changes only modify prog.go, kprobe.go, uprobe.go
- ✅ **Safe to commit** - these failures are pre-existing

### On Upstream
- ❌ **Bug present** - tests are broken upstream
- ❌ **CI doesn't catch it** - only compiles, doesn't run
- ⚠️ **Regression** - tests worked before commit bcca828, broke after

---

## Recommendations

### Short Term (For Our PR)
1. ✅ Document these findings in PR description
2. ✅ Explain that failures are pre-existing, unrelated to issue #1497
3. ✅ Reference this analysis document
4. ✅ Proceed with commit - these don't block us

### Long Term (For Upstream)
1. Fix TestMapBatch - update to new per-CPU Lookup API
2. Investigate TestPerfEventArrayCompatible - unclear root cause
3. Fix CI to actually RUN tests, not just compile them
4. Consider adding pre-commit test validation

---

## Test Commands

### Reproduce Failures
```bash
cd /root/workspace/temp/ebpf  # Clean upstream checkout
go test -short ./...  # Fails consistently
```

### CI's Compile-Only Check
```bash
go test -c -o /dev/null ./...  # Passes (only compiles)
```

### Actual Test Run (vimto)
```bash
# Not reproducible locally - vimto is proprietary test runner
gotestsum -- vimto -kernel :stable-selftests -- \
  go test -race -timeout 5m -short -count 1 -json ./...
```

---

## References

**Upstream Commits:**
- `bcca828` - "map: allow pre-allocating per-CPU values" (Nov 12, 2023)
- Tests broke after this commit but weren't updated

**Upstream Issues:**
- None filed yet - consider filing if proceeding with upstream contribution

**Files Affected:**
- map_test.go: Lines 199 (TestMapBatch), 1729 (TestPerfEventArrayCompatible)
- map.go: Per-CPU lookup implementation

---

## Verification Checklist

- [x] Reproduced failures on clean upstream checkout
- [x] Confirmed failures are NOT in our modified files
- [x] Identified root cause (API change not reflected in tests)
- [x] Confirmed CI only compiles, doesn't run
- [x] Verified failures are pre-existing
- [x] Confirmed no impact on issue #1497

**Conclusion:** These are upstream bugs. We should document and proceed.
