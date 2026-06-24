# Interesting eBPF Projects

## bpfcompat
https://github.com/Kernel-Guard/bpfcompat

**What it does:**
Validates whether compiled eBPF artifacts (.bpf.o) will load and attach
correctly on different kernel/distro versions before shipping.

**Problem it solves:**
CO-RE (Compile Once, Run Everywhere) doesn't guarantee runtime compatibility.
Real failures still happen:
- map type not supported (ringbuf requires kernel ≥5.8)
- LSM hooks require ≥5.7
- missing kernel BTF
- program/attach type unsupported on old kernel

**How it works:**
```
1. takes your .bpf.o file
2. spins up disposable VMs (QEMU/KVM) per kernel version
3. tries load + attach inside each VM
4. outputs compatibility matrix:
    kernel 5.4 → FAIL (ringbuf not supported)
    kernel 5.8 → PASS
    kernel 6.1 → PASS
```

**Usage:**
```bash
# single artifact
bpfcompat test ./build/probe.bpf.o --kernel ubuntu-24.04

# suite
bpfcompat suite run suite.yaml --kernels kernels.yaml

# GitHub Actions
- uses: Kernel-Guard/bpfcompat@v0.1.5
  with:
    suite: ./bpf/suite.yaml
    kernels: ubuntu-lts, rhel-9
    gate: load-attach
```

**Tech stack:** Go (84%), C (4%), Shell — QEMU/KVM/Firecracker backends

**Relevant to your EDR project:**
ebpf-edr-demo uses ringbuf (≥5.8), LSM (≥5.7), LPM trie (≥4.11)
→ bpfcompat could validate before deploying to GKE

**Note:** bpfcompat = pre-ship validation layer
cilium/ebpf feature detection = runtime adaptation
Both solve compatibility but at different stages.

**Found via:** #ebpf-go-dev Slack, 2026-06-20
**Status:** v0.1.5, serious MVP, Apache-2.0
**Priority:** low for now — revisit when deploying EDR to new environments
