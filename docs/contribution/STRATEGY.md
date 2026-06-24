# Open Source Contribution Direction

Open discussion — no final decision. Updated as thinking evolves.

## Current State
- First PR #2040 submitted to cilium/ebpf (Iter.Info() implementation)
- Deep learning cilium/ebpf library internals
- Building ebpf-edr-demo as real project foundation
- Understanding: link/, prog.go, elf_sections.go, internal/sys/

## Core Principle
PRs are the real standard, not time.
```
PR-based:    "3 merged PRs" → proves you understood, implemented, got reviewed
```

## Milestone Plan (no fixed timeline)

### Stage 1 — cilium/ebpf foundation
```
□ PR #2040 merged                           ← in progress
□ 1 more meaningful PR (issue #1497 or other)
□ 1 PR found yourself (no issue guiding)   ← most important signal
□ review another contributor's PR,
  maintainer agrees with your comment
```

The "self-found PR" is the strongest signal:
```
guided PR (issue exists) → shows you can implement
self-found PR            → shows you truly understand the codebase
                         → maintainers notice this most
```

### Stage 2 — expand to Tetragon
```
□ understand Tetragon codebase enough to open a PR
□ 1-2 merged PRs
□ recognized by their maintainers
```

Why Tetragon before others:
```
focus:    runtime security enforcement, eBPF-based EDR
language: Go + C (BPF) — same stack as cilium/ebpf
size:     smaller, more focused than Cilium main
fit:      very strong — ebpf-edr-demo IS a mini-Tetragon
          direct overlap with security background
```

### Stage 3 — go deeper
```
□ fix something hard without any guidance
□ OR contribute to Linux kernel BPF subsystem itself
□ OR build something original on top of this foundation
□ OR become recognized at maintainer level in one project
```

## Project Landscape

```
cilium/ebpf        ← current, library foundation, Go
                      used by everyone below internally

Tetragon           ← next natural step
                      runtime security + eBPF, Go + C
                      direct match to EDR background

Falco              ← good security match
                      C++ core (barrier), Go plugins
                      after Tetragon gives broader context

Cilium (main)      ← later or skip
                      networking/CNI focus, not security first
                      very large codebase, steep entry

Inspektor Gadget   ← observability angle
                      good for broader eBPF understanding

Linux kernel BPF   ← long term
                      deepest level, C, kernel maintainers
                      when foundation is very solid
```

## What NOT to do
```
× spread across multiple projects simultaneously
× chase milestones without deep understanding
× jump to Cilium main before security-focused projects
× measure progress by hours or time spent
```

## Key Insight
```
depth in one project > shallow in many
the work speaks for itself when done with understanding
Tetragon/Falco match background more than Cilium networking
```
