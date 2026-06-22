//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

char __license[] SEC("license") = "GPL";

// Struct layout (40 bytes total, same size as before — fd/count fill old padding):
//   offset  0  __u32 pid
//   offset  4  __u16 family   kprobe/kretprobe: socket family (AF_INET=2, AF_INET6=10)
//   offset  6  __u16 sport    kprobe/kretprobe: source port
//   offset  8  __u16 dport    kprobe/kretprobe: dest port
//   offset 10  __u16 fd       uprobe: write() fd at entry (e.g. 2=stderr); garbage at return
//   offset 12  __u32 count    uprobe: write() byte count at entry; garbage at return
//   offset 16  __s64 retval   uretprobe: bytes written at return; garbage at entry
//   offset 24  char  probe_type[16]
struct event {
    __u32 pid;
    __u16 family;
    __u16 sport;
    __u16 dport;
    __u16 fd;
    __u32 count;
    __s64 retval;
    char  probe_type[16];
};

// Array map — 4 slots, one per probe type
// index 0 = kprobe_accept
// index 1 = kretprobe_accept
// index 2 = uprobe_write
// index 3 = uretprobe_write
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 4);
    __type(key, __u32);
    __type(value, struct event);
} results SEC(".maps");

// ── kprobe: fires at inet_csk_accept ENTRY ──────────────────────────────────
// At entry: function has NOT returned yet. PT_REGS_RC reads rax = garbage.
// Demonstrates: wrong values when kretprobe program runs here instead.
SEC("kprobe/inet_csk_accept")
int kprobe_accept(struct pt_regs *ctx)
{
    __u32 key = 0;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid = bpf_get_current_pid_tgid() >> 32;

    // PT_REGS_RC = rax register, not yet set at entry → reads garbage
    struct sock *sk = (struct sock *)PT_REGS_RC(ctx);
    e->family = BPF_CORE_READ(sk, __sk_common.skc_family);
    e->sport  = BPF_CORE_READ(sk, __sk_common.skc_num);
    e->dport  = 0;

    __builtin_memcpy(e->probe_type, "kprobe", 7);
    return 0;
}

// ── kretprobe: fires at inet_csk_accept RETURN ──────────────────────────────
// At return: sk = accepted socket (real rax value). Valid family/sport/dport.
SEC("kretprobe/inet_csk_accept")
int BPF_KRETPROBE(kretprobe_accept, struct sock *sk)
{
    if (!sk)
        return 0;

    __u32 key = 1;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid    = bpf_get_current_pid_tgid() >> 32;
    e->family = BPF_CORE_READ(sk, __sk_common.skc_family);
    e->sport  = BPF_CORE_READ(sk, __sk_common.skc_num);
    e->dport  = BPF_CORE_READ(sk, __sk_common.skc_dport);

    __builtin_memcpy(e->probe_type, "kretprobe", 10);
    return 0;
}

// ── uprobe: fires at write() ENTRY in libc ──────────────────────────────────
// At entry: fd and count are in argument registers (rdi, rdx).
// fd=2 for stderr, count=bytes requested.
// If attached via Uretprobe() instead: registers clobbered → garbage fd/count.
SEC("uprobe/write")
int BPF_UPROBE(uprobe_write, int fd, const void *buf, size_t count)
{
    __u32 key = 2;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid   = bpf_get_current_pid_tgid() >> 32;
    e->fd    = (__u16)fd;
    e->count = (__u32)count;

    __builtin_memcpy(e->probe_type, "uprobe", 7);
    return 0;
}

// ── uretprobe: fires at write() RETURN in libc ──────────────────────────────
// At return: retval = bytes actually written (or -errno).
// If attached via Uprobe() instead: rax not set at entry → garbage retval.
SEC("uretprobe/write")
int BPF_URETPROBE(uretprobe_write, ssize_t retval)
{
    __u32 key = 3;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid    = bpf_get_current_pid_tgid() >> 32;
    e->retval = retval;

    __builtin_memcpy(e->probe_type, "uretprobe", 10);
    return 0;
}
