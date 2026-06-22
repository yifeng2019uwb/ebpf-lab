# Issue #1497 Implementation Notes

## Problem
Kprobe and Kretprobe share the same BPF_PROG_TYPE_KPROBE, causing silent mismatches when programs are attached via wrong link function (e.g., kretprobe program attached via Kprobe()).

## Solution
Added `sectionName` field to Program struct to track ELF section origin (kprobe/, kretprobe/, uprobe/, uretprobe/) and validate correct attachment.

## Implementation

### Files Modified
- `prog.go`: Added sectionName field, SectionName() getter
- `link/kprobe.go`: Validation logic for kprobe/kretprobe
- `link/uprobe.go`: Validation logic for uprobe/uretprobe

### Validation Behavior
- **ELF-loaded programs** (sectionName != ""): Full validation, error on mismatch
- **Pin/FD-loaded programs** (sectionName == ""): No validation possible
- **Error messages**: Clear, actionable (e.g., "program is kretprobe/inet_csk_accept, cannot attach via Kprobe(); use appropriate link")

## Test Results

All 20 validation cases passing:
- 4 correct attachments: succeed
- 4 wrong attachments: error with message
- 8 cross-domain: properly rejected
- Pin/FD: no validation (sectionName unavailable)

## Remaining Work (Required to Complete)

### 1. Add Unit Tests
Must add tests for all changes:
- **prog.go**: Test sectionName field and SectionName() method
- **link/kprobe.go**: Test validation logic for kprobe/kretprobe mismatches
- **link/uprobe.go**: Test validation logic for uprobe/uretprobe mismatches

### 2. Pass All CI Checks
Before submitting PR (currently not done):
- [ ] staticcheck passes
- [ ] golangci-lint passes  
- [ ] go build succeeds
- [ ] All unit tests pass
- [ ] No new test failures introduced

## References
- Real-world bug: #1490 (developer used wrong attachment hook)
- Test evidence: TestResult.png (20/20 validation cases)
- GitHub discussion: https://github.com/cilium/ebpf/discussions/1490#discussioncomment-9821116
