# eBPF Lab — Work Status

**Last updated:** 2026-06-24

---

## Current Work

**Active:** Perf event tracing lab
- **Path:** `experiments/link-perf-lab/kprobe-test/`
- **Status:** ✅ Working — multi-syscall monitor (openat, close, read, write)
- **Recent:** Fixed struct alignment (C↔Go binary serialization), SIZE values correct
- **Next:** Add signal handler for clean shutdown, remove unused TID filtering, write README

---

## Stage 1 Progress

| Item | Status |
|------|--------|
| PR #2040 (Iter.Info) | ✅ Approved, awaiting merge |
| Issue #1497 | ✅ Discussed & understood (won't fix) |
| **Self-found issue** | ⏳ **TODO** |

**Blocker:** Need to find and start work on a self-found issue to complete Stage 1.

---

## Next Priorities

1. **Hunt for self-found issue** — browse cilium/ebpf, find something interesting
2. **Finish perf event lab** — signal handling, cleanup, README
3. **Learn remaining link types** — fentry/fexit, syscalls.go, netfilter.go

---

## Quick References

- **Docs:** See `docs/README.md` for learning materials, contribution strategy
- **Lab details:** See individual experiment READMEs
- **Link types reference:** `docs/learning/link-deep-dive.md`
