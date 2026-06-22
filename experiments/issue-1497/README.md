# Experiment: issue #1497 — kprobe/kretprobe mismatch

https://github.com/cilium/ebpf/issues/1497

## What this tests

The cilium/ebpf library does not return an error when a probe program is attached
via the wrong link function. All four probe types share BPF_PROG_TYPE_KPROBE, so
the existing `prog.Type() != ebpf.Kprobe` check cannot distinguish them:

```
SEC("kprobe/...")    → BPF_PROG_TYPE_KPROBE  ← same
SEC("kretprobe/...") → BPF_PROG_TYPE_KPROBE  ← same
SEC("uprobe/...")    → BPF_PROG_TYPE_KPROBE  ← same
SEC("uretprobe/...") → BPF_PROG_TYPE_KPROBE  ← same
```

## Real-world bug reference

Issue #1490: https://github.com/cilium/ebpf/issues/1490
Developer used `SEC("kretprobe/inet_csk_accept")` but called `link.Kprobe()` in Go.
Result: garbage output, no error from library.

## Two-phase experiment

### Phase 1 — reproduce (unmodified cilium/ebpf)

Run 10 tests across 4 pairs + 2 cross-domain.
All attach() calls succeed (no error). Compare CORRECT vs WRONG values:
- CORRECT: program fires at intended hook, values make sense
- WRONG:   program fires at wrong hook, values are garbage

Expected output for Pair 2 (issue #1490 exact case):
```
kretprobe prog + Kretprobe() [CORRECT]  pid=1234  family=10  sport=22  dport=54321
kretprobe prog + Kprobe()    [WRONG]    pid=1234  family=47213 sport=0  dport=0
→ values differ: WRONG is garbage (proves bug)
```

### Phase 2 — fix (modify cilium/ebpf, use go.mod replace)

Uncomment the replace directive in go.mod:
```
replace github.com/cilium/ebpf => /Users/yifengzh/workspace/ebpf
```

After implementing sectionName fix in cilium/ebpf:
- Tests labeled CORRECT → still no error (fix doesn't break correct usage)
- Tests labeled WRONG   → attach returns error (fix catches the mismatch)

## Test matrix (main.go)

```
Pair 1 — kprobe program (key 0, inet_csk_accept):
  CORRECT: kprobe prog + Kprobe()       fires at entry  → reads PT_REGS_RC (garbage)
  WRONG:   kprobe prog + Kretprobe()    fires at return → reads PT_REGS_RC (real sk)

Pair 2 — kretprobe program (key 1, inet_csk_accept)  ← issue #1490 exact case:
  CORRECT: kretprobe prog + Kretprobe() fires at return → real family/sport/dport
  WRONG:   kretprobe prog + Kprobe()    fires at entry  → garbage values

Pair 3 — uprobe program (key 2, libc write()):
  CORRECT: uprobe prog + Uprobe()       fires at entry
  WRONG:   uprobe prog + Uretprobe()    fires at return
  Note: uprobe_write only captures pid — values won't differ.
        Bug proven by: no attach error returned for wrong combo.

Pair 4 — uretprobe program (key 3, libc write()):
  CORRECT: uretprobe prog + Uretprobe() fires at return → retval=bytes written
  WRONG:   uretprobe prog + Uprobe()    fires at entry  → retval=garbage

Cross-domain (key 0 and key 2):
  kprobe prog + Uprobe()   — kernel program on userspace hook
  uprobe prog + Kprobe()   — userspace program on kernel hook
  Empirical result: observe whether kernel rejects or silently accepts.
```

## Files

```
probe.bpf.c  BPF C programs (4 probe types, 1 array map with 4 slots)
gen.go       go:generate directive for bpf2go
main.go      Go loader: runs all 10 tests, prints side-by-side comparison
```

### struct event layout (40 bytes)

```
offset  0  __u32 pid           (4 bytes)
offset  4  __u16 family        (2 bytes)
offset  6  __u16 sport         (2 bytes)
offset  8  __u16 dport         (2 bytes)
offset 10  [6 bytes padding]   (to align __s64 to 8-byte boundary)
offset 16  __s64 retval        (8 bytes)
offset 24  char  probe_type[16](16 bytes)
total  40
```

## Phase 1 results — volatility is proof of the bug

**Pair 3 (uprobe program):** Captures fd and count from registers.
- CORRECT (entry): relatively stable (fd=4, count=64 from Go runtime write)
- WRONG (return):  **unpredictable** — sometimes fd=4,count=64 (lucky same write), sometimes garbage (different write caught by uretprobe)

Running multiple times shows: garbage values change, cross-domain B values shift slightly. This volatility is **more proof** than fixed garbage:
- Correct usage: deterministic (always captures the right data)
- Wrong usage: non-deterministic (reads random kernel state or clobbered registers)

**Cross-domain B rcb1 (uprobe prog + Kprobe):**
- uprobe reads argument registers rdi/rdx at inet_csk_accept entry
- rdi = struct sock *sk (kernel address, low 16 bits = ~37888, stable)
- rdx = 2nd kernel arg (varies slightly each run: 2161261428 vs 2161261948)
- The program "works" (no attach error) but reads completely wrong data

## How to run

Requires Linux with kernel 5.8+, root, clang/llvm, libbpf-dev.
Run on DigitalOcean droplet (x86_64). DO NOT run go generate on Mac (ARM64 headers).

```bash
# 1. generate BPF Go bindings (Linux only)
go generate .

# 2. check compile errors (no root needed)
go build .

# 3. run Phase 1 (try multiple times to see volatility)
sudo go run .
sudo go run .   # run again — values in wrong cases will differ

# 4. to run Phase 2: uncomment replace in go.mod, implement fix, then:
sudo go run .
```
