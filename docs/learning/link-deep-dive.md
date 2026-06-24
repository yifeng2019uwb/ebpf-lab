# eBPF Link Deep Dive

## What is a Link?
A Link = the attachment between a BPF program and a kernel hook.
```
BPF program (the code) + Link (the hook) = BPF running in kernel
```

## All Link Types

### Tracing/Probing links
```
PerfEventType     → kprobe, kretprobe, uprobe, uretprobe, tracepoint
                    uses Linux perf subsystem (perf_event_open syscall)
KprobeMultiType   → attach to many kprobes at once (newer bpf_link API)
UprobeMultiType   → attach to many uprobes at once
RawTracepointType → raw tracepoint (lower overhead than tracepoint)
TracingType       → BTF-based tracing: fentry/fexit/LSM
```

### Network links
```
XDPType      → packet processing at driver level (fastest, before network stack)
TCXType      → traffic control (tc) hook
NetkitType   → virtual network device hook
NetNsType    → network namespace hook
NetfilterType → netfilter/iptables replacement hook
```

### Other links
```
CgroupType   → cgroup hooks (resource control per container/process group)
IterType     → iterate kernel objects (task, bpf_map, socket...)
StructOpsType → replace kernel struct function pointers
```

---

## PerfEvent Subtypes
`PerfEventType` link covers multiple subtypes:

```
PerfEventKprobe      → kernel function ENTRY hook
PerfEventKretprobe   → kernel function RETURN hook
PerfEventUprobe      → userspace function ENTRY hook
PerfEventUretprobe   → userspace function RETURN hook
PerfEventTracepoint  → kernel tracepoint
PerfEventEvent       → hardware/software perf counters (CPU cycles, cache misses)
PerfEventUnspecified → unknown subtype
```

## Perf Event Arrays: Kernel→Userspace Transport

Ring buffer mechanism for sending event data from BPF to Go.

**Flow:**
```
BPF program → bpf_perf_event_output() → Per-CPU kernel ring buffer → perf.NewReader() → Go app
                                         (lockfree, parallel)          (userspace)
```

**BPF side (C):**
```
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));  // CPU ID
    // omit value_size — kernel manages internally
} events SEC(".maps");

int probe(struct trace_event_raw_sys_enter *ctx) {
    struct event_data ev = {.pid = ..., .filename = ...};
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &ev, sizeof(ev));
    return 0;
}
```

**Go side (userspace reader):**
```
reader, _ := perf.NewReader(objs.Maps.Events, 4096)  // 4096 = buffer size
record, _ := reader.Read()
binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event)
```

**Key differences from ringbuf:**
- Perf events = per-CPU buffers (scales to many cores, simpler)
- Ringbuf = single shared buffer (flexible data format, modern)
- Perf = fixed struct only; ringbuf = any serialization format
- Perf = older kernels (5.0+); ringbuf = newer (5.8+)

**Common pitfall:** struct alignment. C and Go struct layouts must match exactly — field order, padding, all matter.

**Memory reading in BPF:**
- `bpf_probe_read_kernel_str()` → kernel memory (kernel code, kernel data structures)
- `bpf_probe_read_user_str()` → user-space memory (syscall arguments often point here)

## kprobe vs kretprobe
```
kprobe    = fires when kernel function is CALLED (entry)
            → see function arguments
            → "what was passed in?"

kretprobe = fires when kernel function RETURNS (exit)
            → see return value
            → "did it succeed or fail?"

example: sys_execve
    kprobe    → fires when process calls exec → see filename, argv
    kretprobe → fires when exec returns → see success/error code
```

## uprobe vs uretprobe — same concept but userspace
```
kprobe/kretprobe  → hooks KERNEL functions (sys_execve, vfs_open...)
uprobe/uretprobe  → hooks USERSPACE functions (nginx, python, your app)

uprobe    = fires when userspace function is CALLED
uretprobe = fires when userspace function RETURNS

example: /usr/bin/nginx
    uprobe on main()    → fires when nginx starts
    uretprobe on main() → fires when nginx exits
```

## Kernel space vs Userspace hooks
```
kernel hooks (kprobe/LSM/tracepoint):
    runs in HOST kernel
    sees ALL processes: containers, k8s pods, bare metal
    one hook catches everything on the machine

userspace hooks (uprobe):
    targets a specific binary path (/usr/bin/nginx)
    only fires for that specific binary
    needs to know which binary to watch
```

Your EDR uses LSM (TracingType) → catches ALL processes from one hook.
Tetragon/Falco use same approach — kernel-level = universal coverage.

## Netfilter link
```
NetfilterInet = WHERE in the network path to hook:
    PRE_ROUTING   ← packet arrived, before routing decision
    LOCAL_IN      ← packet destined for this machine
    FORWARD       ← packet being forwarded to another host
    LOCAL_OUT     ← packet leaving this machine
    POST_ROUTING  ← packet leaving, after routing decision

NetfilterProto = WHICH protocol family:
    NFPROTO_IPV4  ← only IPv4
    NFPROTO_IPV6  ← only IPv6
    NFPROTO_INET  ← both IPv4 and IPv6
    NFPROTO_ARP   ← ARP packets
```
Netfilter = iptables replacement using BPF. Your EDR does not use this.

## fentry/fexit — The Better kprobe/kretprobe

**fentry** = function entry (like kprobe, but better)
**fexit** = function exit (like kretprobe, but better)

**Key advantages:**
- No confusion between entry/exit — syntax makes it explicit
- Direct kernel memory access (no `bpf_probe_read_kernel` needed)
- Requires BTF (kernel 5.8+)
- Future direction: fprobes expected to mostly replace kprobe/kretprobe

**Example:**
```c
SEC("fentry/tcp_connect")
int BPF_PROG(tcp_connect, struct sock *sk) {
    // Direct access to sk — BTF guarantees memory layout
    __be32 daddr = sk->__sk_common.skc_daddr;
    return 0;
}
```

**Why preferred over kprobe/kretprobe:**
- Entry vs exit is **syntactically clear** (no risk of mixing them up)
- Works with ringbuf (modern transport)
- Kernel direction is toward BTF-based tracing

**Issue #1497 Status:** Marked "won't fix" — maintainer prefers ecosystem to move toward fentry/fexit rather than patch kprobe/kretprobe confusion

## Reference
https://pkg.go.dev/github.com/cilium/ebpf@v0.21.0/link

## All Link Files (cilium/ebpf/link/)

### Probe-based (✅ Learned)
- **kprobe.go** - Kprobe() / Kretprobe() + validation
- **kprobe_multi.go** - Batch attach to many kernel functions
- **uprobe.go** - Uprobe() / Uretprobe() 
- **uprobe_multi.go** - Batch attach to many user functions
- **perf_event.go** - Performance monitoring events
- **raw_tracepoint.go** - Raw kernel tracepoints
- **tracepoint.go** - Kernel static tracepoints

### Network-related (For EDR)
- **netfilter.go** - iptables/nftables hook
- **socket_filter.go** - Socket-level packet filtering
- **xdp.go** - Express data path (NIC driver level)
- **tcx.go** - Traffic control hooks
- **netkit.go** - Virtual network device
- **netns.go** - Network namespace hooks

### System & Observability
- **tracing.go** - fentry/fexit/LSM hooks
- **syscalls.go** - Syscall-specific monitoring
- **cgroup.go** - Cgroup hooks
- **struct_ops.go** - Kernel struct operations
- **iter.go** - Iterator links (your PR #2040 ✨)

### Infrastructure
- **link.go** - Link interface, Info, RawLink base
- **link_other.go** - Link type constants, Info structs
- **anchor.go** - Attach point anchors
- **program.go** - Program linking
- **query.go** - Query utilities
- **doc.go** - Documentation

## Learning Path for EDR

**Completed:**
- ✅ kprobe / kretprobe / uprobe / uretprobe
- ✅ kprobe_multi / uprobe_multi
- ✅ Method vs function patterns
- ✅ Perf event arrays (per-CPU kernel→userspace transport)
- ✅ Tracepoint attachment (sys_enter_* / sys_exit_*)
- ✅ Struct alignment issues in C↔Go binary serialization
- ✅ Discussed issue #1497 (won't fix; future is fentry/fexit)
- ✅ **PR #2040 approved** (Iter.Info implementation)

**Current focus:**
- ebpf-lab/experiments/link-perf-lab/kprobe-test: Multi-syscall tracepoint monitoring with perf events

**Next learning priorities:**
1. **tracing.go** (fentry/fexit) - Preferred over kprobe/kretprobe for new code
2. **syscalls.go** - Syscall-specific monitoring (core EDR)
3. **netfilter.go** / **socket_filter.go** - Network monitoring
4. **cgroup.go** - Container/process group resource control
5. **ringbuf** vs perf event comparison - Modern data transport

**Key insights:**
- Perf events are good for syscall tracing (per-CPU, automatic)
- fentry/fexit are the future (no confusion, direct kernel memory)
- Struct alignment must match C↔Go exactly (explicit padding needed)
- Signal handling for clean shutdown with blocking I/O (use context cancellation)
