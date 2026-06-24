# Lab Experiment: Issue #1497 — kprobe/kretprobe mismatch

**Status: ❌ ARCHIVED (Won't Fix)**

**GitHub Issue:** https://github.com/cilium/ebpf/issues/1497

**Decision Date:** 2026-06-24

---

## Why This Experiment Was Abandoned

The cilium/ebpf maintainers decided to **close this issue as "won't fix"** with this reasoning:

> "I'd suggest we close this issue as 'won't fix' since this shouldn't happen that often in practice... I'd rather not invest too much time here, given that fprobes are expected to mostly take over what k(ret)probes did in the past."

**Key points:**
1. Problem is real but rare in practice
2. Proposed solution (sectionName) is incomplete (doesn't help with pin/FD cases)
3. **Better future:** ecosystem moving toward fentry/fexit (fprobes), not patching kprobe/kretprobe
4. fentry/fexit don't have this confusion by design

---

## What This Lab Showed

Successfully demonstrated that the cilium/ebpf library does NOT validate probe type mismatches:

```
SEC("kprobe/...")    → BPF_PROG_TYPE_KPROBE  ← all four map
SEC("kretprobe/...") → BPF_PROG_TYPE_KPROBE  ← to the same
SEC("uprobe/...")    → BPF_PROG_TYPE_KPROBE  ← program type
SEC("uretprobe/...") → BPF_PROG_TYPE_KPROBE  ← (issue #1490)
```

**Real-world impact:** Issue #1490 — developer attached kretprobe program via `link.Kprobe()`, got garbage data, no error from library.

---

## Solution That Was Proposed (Not Merged)

Added `sectionName` field to Program struct to validate correct attachment:
- **4 correct cases:** pass
- **4 wrong cases:** error with clear message
- **8 cross-domain:** properly rejected
- **Pin/FD limitation:** cannot validate (acceptable)

All 20 test cases passed. See NOTES.md for implementation details.

---

## Learning Value

This experiment demonstrates:
1. **Root cause analysis** — why different probe types share the same BPF_PROG_TYPE_KPROBE
2. **Design decisions** — tradeoffs between validation completeness and pin/FD limitations
3. **Why future is fentry/fexit** — they avoid this problem by design (entry vs exit is syntactically explicit)

---

## For Reference

- **Real-world bug:** https://github.com/cilium/ebpf/issues/1490
- **Original issue:** https://github.com/cilium/ebpf/issues/1497
- **Test evidence:** TestResult.png (20/20 validation cases passed)
- **Implementation notes:** NOTES.md (technical details of proposed solution)

---

## Takeaway

Don't fix broken kprobe/kretprobe — **use fentry/fexit instead!**

See: `docs/learning/link-deep-dive.md` → fentry/fexit section
