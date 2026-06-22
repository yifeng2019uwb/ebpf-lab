# ebpf-lab

eBPF experiments and issue reproductions for learning and contributing to cilium/ebpf.

Each experiment is self-contained in its own directory under `experiments/`.

## Structure

```
experiments/
    issue-1497/   → reproduce kprobe/kretprobe mismatch bug (cilium/ebpf#1497)
    pin-test/     → pin/unpin behavior and zero-downtime patterns
```

## Current work

**Issue #1497 — two-phase experiment:**
- Phase 1 ✓ COMPLETE: Reproduced bug in probe attachment validation
- Phase 2: Implement sectionName fix in cilium/ebpf

**Handoff document:** [HANDOFF.md](HANDOFF.md) — everything needed to start Phase 2

## Purpose

- Reproduce real bugs before fixing them in the library
- Test proposed fixes end-to-end using local cilium/ebpf
- Document learning through working code
- Keep ebpf-edr-demo clean from experimental code

## Related

- cilium/ebpf library: https://github.com/cilium/ebpf
- ebpf-edr-demo project: https://github.com/yifeng2019uwb/ebpf-edr-demo
- Issue #1497: https://github.com/cilium/ebpf/issues/1497
- Learning notes: `/Users/yifengzh/workspace/learnGo/notes/ebpf/issue-1497-plan.md`
