# Issue #1497 Implementation Notes (ARCHIVED)

**Status:** Won't Fix — See README.md

---

## Problem Statement

Kprobe and Kretprobe share the same `BPF_PROG_TYPE_KPROBE`, causing silent mismatches when programs are attached via wrong link function (e.g., kretprobe program attached via `Kprobe()`).

---

## Proposed Solution (Not Merged)

Added `sectionName` field to Program struct to track ELF section origin (kprobe/, kretprobe/, uprobe/, uretprobe/) and validate correct attachment.

### Files Modified
- `prog.go`: Added sectionName field, SectionName() getter
- `link/kprobe.go`: Validation logic for kprobe/kretprobe
- `link/uprobe.go`: Validation logic for uprobe/uretprobe

### Validation Behavior
- **ELF-loaded programs** (sectionName != ""): Full validation, error on mismatch
- **Pin/FD-loaded programs** (sectionName == ""): No validation possible
- **Error messages**: Clear, actionable

---

## Test Results

All 20 validation cases passed:
- ✅ 4 correct attachments: succeed
- ✅ 4 wrong attachments: error with message
- ✅ 8 cross-domain: properly rejected
- ⚠️ Pin/FD: no validation (sectionName unavailable)

See TestResult.png for screenshot.

---

## Why Not Merged

1. **Incomplete solution** — doesn't help pin/FD cases (acceptable limitation but acknowledged)
2. **Not worth the complexity** — problem is rare in practice
3. **Better future exists** — fentry/fexit are the recommended direction

Maintainer: "I'd rather not invest too much time here, given that fprobes are expected to mostly take over what k(ret)probes did in the past."

---

## References

- Real-world bug: #1490 (developer used wrong attachment hook)
- Original GitHub issue: https://github.com/cilium/ebpf/issues/1497
- See: `docs/learning/link-deep-dive.md` for why fentry/fexit are better
