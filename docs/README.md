# eBPF Lab Documentation

Index of learning materials, contribution guides, and archived plans.

---

## 📚 Learning Materials

Resources for understanding eBPF concepts and cilium/ebpf library internals.

### [link-deep-dive.md](learning/link-deep-dive.md)
Comprehensive reference guide to all eBPF link types in cilium/ebpf.

**Contents:**
- What is a Link? (attachment between BPF program and kernel hook)
- All link types: Tracing, Network, System & Observability
- PerfEvent subtypes (kprobe, kretprobe, uprobe, uretprobe, tracepoint)
- Perf event arrays (kernel→userspace transport mechanism)
- kprobe vs kretprobe vs fentry vs fexit
- Memory reading patterns (bpf_probe_read_kernel vs bpf_probe_read_user)
- Learning path for EDR development

---

## 🤝 Contribution Guidelines

How to contribute to cilium/ebpf and related projects.

### [contributing.md](contribution/contributing.md)
Best practices for submitting PRs and working with the cilium/ebpf maintainers.

### [STRATEGY.md](contribution/STRATEGY.md)
Strategy for choosing high-impact issues to work on, from issue triage to PR submission.

---

## 💡 Ideas & References

Brainstorming and reference materials for future work.

### [interesting-projects.md](reference/interesting-projects.md)
Interesting open source eBPF projects worth learning from.

### [proposed-ideas.md](reference/proposed-ideas.md)
Ideas for new features or improvements in cilium/ebpf.

---

## 🧪 Lab Experiments

Active learning labs in `/experiments/`:

### link-perf-lab/kprobe-test
Multi-syscall tracepoint monitor using BPF perf event arrays.

**Status:** Working — captures openat, close, read, write syscalls
- Fixed: struct alignment issues (C↔Go binary serialization)
- SIZE values now correct
- Signal handling: context cancellation for clean shutdown

**Key learnings:**
- Struct padding must match C and Go exactly
- Perf event arrays route automatically to CPU
- Tracepoint context (sys_enter_*) provides syscall args in user-space

See: [../HANDOFF.md](../HANDOFF.md) for current session details.

---

## 📖 How to Use

1. **New to eBPF links?** Start with [link-deep-dive.md](learning/link-deep-dive.md)
2. **Contributing to cilium/ebpf?** Read [STRATEGY.md](contribution/STRATEGY.md) then [contributing.md](contribution/contributing.md)
3. **Looking for ideas?** Check [proposed-ideas.md](reference/proposed-ideas.md)

---

## 🔗 Related Files

- [../HANDOFF.md](../HANDOFF.md) — Current session status and next priorities
- [../README.md](../README.md) — Lab overview
- [../experiments/](../experiments/) — Active learning experiments
