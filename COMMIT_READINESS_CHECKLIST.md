# Issue #1497 Commit Readiness Checklist

**Status:** 🟡 IN PROGRESS - Unit tests pending

---

## Phase 1: Implementation ✅ COMPLETE

- [x] prog.go: Added sectionName field to Program struct
- [x] prog.go: Populated sectionName in NewProgramWithOptions()
- [x] prog.go: Added SectionName() method
- [x] kprobe.go: Added validation logic
- [x] uprobe.go: Added validation logic
- [x] uprobe.go: Added strings import

**Verification:** Issue-1497 experiment: 20/20 tests passing ✅

---

## Phase 2: Unit Tests ⏳ PENDING

### A. prog.go Unit Tests (4 tests needed)

**File:** `/Users/yifengzh/workspace/ebpf/prog_test.go`

#### Test 1: TestProgramSectionName
- [ ] Create program with sectionName
- [ ] Call SectionName()
- [ ] Assert value matches
- [ ] Status: ⏳ TODO

#### Test 2: TestProgramSectionNames (parameterized)
- [ ] Test multiple section names: kprobe/, kretprobe/, uprobe/, uretprobe/, ""
- [ ] Verify SectionName() returns correct value for each
- [ ] Status: ⏳ TODO

#### Test 3: TestProgramCloneSectionName
- [ ] Create program with sectionName
- [ ] Clone it
- [ ] Verify cloned program preserves sectionName
- [ ] Status: ⏳ TODO

#### Test 4: TestProgramStructSectionName
- [ ] Test NewProgram with sectionName set
- [ ] Test NewProgram without sectionName (default empty)
- [ ] Test all constructor paths
- [ ] Status: ⏳ TODO

### B. kprobe.go Unit Tests (6 tests needed)

**File:** `/Users/yifengzh/workspace/ebpf/link/kprobe_test.go`

#### Test 1: TestKprobeValidationCorrect
- [ ] Create program with "kprobe/" section
- [ ] Attach via Kprobe()
- [ ] Assert NO error
- [ ] Status: ⏳ TODO

#### Test 2: TestKprobeValidationWrong
- [ ] Create program with "kprobe/" section
- [ ] Attach via Kretprobe()
- [ ] Assert error: "cannot attach via Kretprobe"
- [ ] Status: ⏳ TODO

#### Test 3: TestKretprobeValidationCorrect
- [ ] Create program with "kretprobe/" section
- [ ] Attach via Kretprobe()
- [ ] Assert NO error
- [ ] Status: ⏳ TODO

#### Test 4: TestKretprobeValidationWrong
- [ ] Create program with "kretprobe/" section
- [ ] Attach via Kprobe()
- [ ] Assert error: "cannot attach via Kprobe"
- [ ] Status: ⏳ TODO

#### Test 5: TestCrossDomainUprobeOnKprobe
- [ ] Create program with "uprobe/" section
- [ ] Try to attach via Kprobe()
- [ ] Assert error (cross-domain)
- [ ] Status: ⏳ TODO

#### Test 6: TestCrossDomainUretprobeOnKretprobe
- [ ] Create program with "uretprobe/" section
- [ ] Try to attach via Kretprobe()
- [ ] Assert error (cross-domain)
- [ ] Status: ⏳ TODO

### C. uprobe.go Unit Tests (6 tests needed)

**File:** `/Users/yifengzh/workspace/ebpf/link/uprobe_test.go`

#### Test 1: TestUprobeValidationCorrect
- [ ] Create program with "uprobe/" section
- [ ] Attach via Uprobe()
- [ ] Assert NO error
- [ ] Status: ⏳ TODO

#### Test 2: TestUprobeValidationWrong
- [ ] Create program with "uprobe/" section
- [ ] Attach via Uretprobe()
- [ ] Assert error: "cannot attach via Uretprobe"
- [ ] Status: ⏳ TODO

#### Test 3: TestUretprobeValidationCorrect
- [ ] Create program with "uretprobe/" section
- [ ] Attach via Uretprobe()
- [ ] Assert NO error
- [ ] Status: ⏳ TODO

#### Test 4: TestUretprobeValidationWrong
- [ ] Create program with "uretprobe/" section
- [ ] Attach via Uprobe()
- [ ] Assert error: "cannot attach via Uprobe"
- [ ] Status: ⏳ TODO

#### Test 5: TestCrossDomainKprobeOnUprobe
- [ ] Create program with "kprobe/" section
- [ ] Try to attach via Uprobe()
- [ ] Assert error (cross-domain)
- [ ] Status: ⏳ TODO

#### Test 6: TestCrossDomainKretprobeOnUretprobe
- [ ] Create program with "kretprobe/" section
- [ ] Try to attach via Uretprobe()
- [ ] Assert error (cross-domain)
- [ ] Status: ⏳ TODO

### D. Test Helpers Needed

**File:** `/Users/yifengzh/workspace/ebpf/link/helpers_test.go`

- [ ] Add `loadProgramWithSection(tb testing.TB, progType ebpf.ProgramType, sectionName string) *ebpf.Program`
- [ ] Add stderr suppression helper for existing tests (optional)
- [ ] Status: ⏳ TODO

---

## Phase 3: Documentation ✅ COMPLETE

- [x] HANDOFF.md: Updated with test results
- [x] README.md (issue-1497): Updated with Phase 2 results
- [x] UNIT_TEST_PLAN.md: Created
- [x] PROG_GO_TEST_PLAN.md: Created
- [x] UPSTREAM_TEST_FAILURES_ANALYSIS.md: Created

---

## Phase 4: Verification ⏳ PENDING

- [ ] Run all new prog.go tests: `go test . -run TestProgram* -v`
- [ ] Run all new kprobe tests: `go test ./link -run TestKprobe* -v`
- [ ] Run all new uprobe tests: `go test ./link -run TestUprobe* -v`
- [ ] Verify no new test warnings/failures
- [ ] Verify issue-1497 experiment still passes: 20/20 tests
- [ ] Full test suite on DO droplet (with -short flag)

---

## Phase 5: Commit & PR ⏳ PENDING

- [ ] All 16 unit tests implemented and passing
- [ ] All 4 test helpers implemented
- [ ] Code review of new tests
- [ ] Create commit with message including:
  - Issue #1497 fix summary
  - Unit test coverage explanation
  - Reference to pre-existing upstream test failures
- [ ] Create PR to cilium/ebpf with full evidence
- [ ] Document in PR:
  - Test matrix results (20/20 passing)
  - Unit test coverage (16 new tests)
  - Known upstream issues (map tests)

---

## Summary

**Done:** Implementation + Documentation (20/20 validation tests proven)  
**Pending:** 16 unit tests + verification + commit  
**Blocked By:** Unit test implementation  

**Next Action:** Implement 16 unit tests across 3 files

---

## Files to Modify

1. `/Users/yifengzh/workspace/ebpf/prog_test.go` — Add 4 tests
2. `/Users/yifengzh/workspace/ebpf/link/helpers_test.go` — Add helper function
3. `/Users/yifengzh/workspace/ebpf/link/kprobe_test.go` — Add 6 tests
4. `/Users/yifengzh/workspace/ebpf/link/uprobe_test.go` — Add 6 tests

**Total:** 16 new unit tests + 1 helper function

