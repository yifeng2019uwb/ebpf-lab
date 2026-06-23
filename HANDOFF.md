# Phase 2 Handoff — implementing sectionName fix for issue #1497

## Phase 1 Status ✓ COMPLETE

Experiment successfully reproduces issue #1497 (kprobe/kretprobe mismatch).
See: `/Users/yifengzh/workspace/ebpf-lab/experiments/issue-1497/`

**Key findings:**
- All 10 probe tests attach **without error** — library does NOT catch mismatches
- Pair 2 (kretprobe via Kprobe) reproduces issue #1490 exactly: garbage family/sport/dport
- Pair 4 (uretprobe via Uprobe): garbage retval (95221441342864 vs 64)
- Cross-domain tests show kernel accepts wrong attachments, kernel memory read as function args
- Volatility across runs proves the bug: wrong hook points read non-deterministic kernel state

Run: `sudo ./issue-1497` multiple times to see non-determinism.

## Phase 2 — What needs to be done

**Current behavior (Phase 1):** Library returns NO ERROR for all mismatches.
```
All 10 tests → attach() succeeds → no error → bug proven by garbage data
```

**Desired behavior (Phase 2):** Library returns ERROR for mismatches.
```
Kprobe() with kretprobe prog  → error: "program is kretprobe/*, cannot attach via Kprobe()"
Kretprobe() with kprobe prog  → error: "program is kprobe/*, cannot attach via Kretprobe()"
Uprobe() with uretprobe prog  → error: "program is uretprobe/*, cannot attach via Uprobe()"
Uretprobe() with uprobe prog  → error: "program is uprobe/*, cannot attach via Uretprobe()"
```

**Goal:** Make cilium/ebpf library return an error when:
- kprobe program attached via Kretprobe()
- kretprobe program attached via Kprobe()
- uprobe program attached via Uretprobe()
- uretprobe program attached via Uprobe()

**Design decision (already made):**
```
sectionName field on *Program struct (from ELF ProgramSpec.SectionName)

Validation logic:
  if sectionName != "" {  // ELF load path (90%+ of usage)
    validate: kprobe prog must use Kprobe(), etc
  }
  // if sectionName == "" (pin load path), skip validation (acceptable limitation)
```

**Reference:** PR #2011 (merged May 2026) added `btf *btf.Handle` field to Program struct
→ proves maintainers accept adding fields + methods to *Program

## Files to modify in cilium/ebpf

### 1. prog.go (root package)

**Add sectionName field to Program struct (line 236-247):**

Current:
```go
type Program struct {
    VerifierLog string
    fd          *sys.FD
    name        string
    pinnedPath  string
    typ         ProgramType
    btf         *btf.Handle  // PR #2011
}
```

Change to:
```go
type Program struct {
    VerifierLog string
    fd          *sys.FD
    name        string
    pinnedPath  string
    typ         ProgramType
    btf         *btf.Handle
    sectionName string  // NEW: from ProgramSpec at ELF load time
}
```

**Populate sectionName in NewProgramWithOptions() (around line 249-280):**

Find where Program is created from ProgramSpec (in NewProgramWithOptions).
Pass spec.SectionName to the constructor:
```go
prog := &Program{
    ...,
    sectionName: spec.SectionName,  // capture from spec
}
```

**Add SectionName() method (after line 650):**

```go
// SectionName returns the ELF section name this program originated from,
// or "" if loaded via pin (no section info available).
func (p *Program) SectionName() string {
    return p.sectionName
}
```

### 2. link/kprobe.go (line 148-160)

**Update kprobe() internal function validation:**

Current (line 158-159):
```go
if prog.Type() != ebpf.Kprobe {
    return nil, fmt.Errorf("invalid program type %s, expected Kprobe", prog.Type())
}
```

Add AFTER existing Type() check:
```go
// Validate entry/return hook type matches program section
if ret {
    // Kretprobe() was called, expect kretprobe program
    if sn := prog.SectionName(); sn != "" && !strings.HasPrefix(sn, "kretprobe/") && !strings.HasPrefix(sn, "uretprobe/") {
        return nil, fmt.Errorf("program is %s (entry hook), cannot attach via Kretprobe() (expects return hook); use Kprobe()", sn)
    }
} else {
    // Kprobe() was called, expect kprobe program
    if sn := prog.SectionName(); sn != "" && (strings.HasPrefix(sn, "kretprobe/") || strings.HasPrefix(sn, "uretprobe/")) {
        return nil, fmt.Errorf("program is %s (return hook), cannot attach via Kprobe() (expects entry hook); use Kretprobe() or Kretprobe()", sn)
    }
}
```

### 3. link/uprobe.go (line 333-340)

**Update uprobe() internal function validation:**

Same pattern as kprobe.go. Add after Type() check:
```go
// Validate entry/return hook type matches program section
if ret {
    // Uretprobe() was called
    if sn := prog.SectionName(); sn != "" && !strings.HasPrefix(sn, "uretprobe/") {
        return nil, fmt.Errorf("program is %s (entry hook), cannot attach via Uretprobe(); use Uretprobe()", sn)
    }
} else {
    // Uprobe() was called
    if sn := prog.SectionName(); sn != "" && strings.HasPrefix(sn, "uretprobe/") {
        return nil, fmt.Errorf("program is %s (return hook), cannot attach via Uprobe(); use Uretprobe()", sn)
    }
}
```

## Testing the fix

### Option A: Local cilium/ebpf fork (fastest)

1. Clone cilium/ebpf to `/Users/yifengzh/workspace/ebpf` (already exists)
2. Make the changes above
3. In `/Users/yifengzh/workspace/ebpf-lab/go.mod`, uncomment:
   ```
   replace github.com/cilium/ebpf => /Users/yifengzh/workspace/ebpf
   ```
4. On DO droplet:
   ```bash
   cd ~/workspace/ebpf-lab/experiments/issue-1497
   go mod tidy
   go generate .
   go build .
   sudo ./issue-1497
   ```

Expected Phase 2 results (after fix):
```
Pair 1 CORRECT (Kprobe + Kprobe):      no error   ✓
Pair 1 WRONG   (Kprobe + Kretprobe):   ERROR      ✓
Pair 2 CORRECT (Kretprobe + Kretprobe):no error   ✓
Pair 2 WRONG   (Kretprobe + Kprobe):   ERROR      ✓
Pair 3 CORRECT (Uprobe + Uprobe):      no error   ✓
Pair 3 WRONG   (Uprobe + Uretprobe):   ERROR      ✓
Pair 4 CORRECT (Uretprobe + Uretprobe):no error   ✓
Pair 4 WRONG   (Uretprobe + Uprobe):   ERROR      ✓
```

WRONG cases will now print: `attach error: program is kprobe/...` instead of attaching.

Cross-domain tests: may still attach (different prog.Type paths) or may error (depends on whether kernel accepts cross-domain, which is informative either way).

### Option B: Upstream PR (later)

After local testing passes, create PR against cilium/ebpf main.

## Important constraints

**Do NOT modify:**
- `probe.bpf.c` — correct as-is
- `main.go` — only add attach-error handling if needed for cross-domain tests
- Anything in `ebpf-edr-demo` — keep this repo clean

**Pin limitation (intentional):**
- Programs loaded via LoadPinnedProgram() have sectionName="" → validation skipped
- This is acceptable: pin users are advanced, they know what they pinned
- Document this clearly in any PR

## References

**Learning notes:**
- [issue-1497-plan.md](/Users/yifengzh/workspace/learnGo/notes/ebpf/issue-1497-plan.md) — full design analysis
- [link-deep-dive.md](/Users/yifengzh/workspace/learnGo/notes/ebpf/link-deep-dive.md) — eBPF link types reference

**Cilium/ebpf code:**
- `/Users/yifengzh/workspace/ebpf/prog.go:140` — ProgramSpec.SectionName field
- `/Users/yifengzh/workspace/ebpf/prog.go:236-247` — Program struct
- `/Users/yifengzh/workspace/ebpf/prog.go:249+` — NewProgramWithOptions
- `/Users/yifengzh/workspace/ebpf/link/kprobe.go:148-160` — kprobe() validation
- `/Users/yifengzh/workspace/ebpf/link/uprobe.go:333-340` — uprobe() validation
- `/Users/yifengzh/workspace/ebpf/elf_sections.go:13-27` — all 4 probe types → BPF_PROG_TYPE_KPROBE

**PR reference:**
- PR #2011 (merged) — added btf field to Program struct, same structural pattern

## Next steps for Phase 2 session

1. Read the files listed above to understand current state
2. Implement changes in sections 1-3 above
3. Run go generate on DO droplet
4. Test locally with go.mod replace
5. If passes, prepare for upstream PR

Good luck! The fix is straightforward — just adding field, populating at load time, validating at attach time.
