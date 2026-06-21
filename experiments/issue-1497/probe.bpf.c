//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

char __license[] SEC("license") = "GPL";

// Event sent to userspace via map
struct event {
    __u32 pid;
    __u16 family;         // kretprobe: socket family (AF_INET=2, AF_INET6=10)
    __u16 sport;          // kretprobe: source port
    __u16 dport;          // kretprobe: dest port
    __s64 retval;         // uretprobe: bytes written (or -errno)
    char  probe_type[16]; // which probe fired: kprobe/kretprobe/uprobe/uretprobe
};

// Array map — 4 slots, one per probe type
// index 0 = kprobe
// index 1 = kretprobe
// index 2 = uprobe
// index 3 = uretprobe
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 4);
    __type(key, __u32);
    __type(value, struct event);
} results SEC(".maps");

// ── kprobe: fires at inet_csk_accept ENTRY ──────────────────────────────────
// At entry: inet_csk_accept has NOT returned yet.
// Trying to read the socket return value = reading garbage from rax register.
// This demonstrates what happens when Kprobe() is used with a kretprobe program.
SEC("kprobe/inet_csk_accept")
int kprobe_accept(struct pt_regs *ctx)
{
    __u32 key = 0;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid = bpf_get_current_pid_tgid() >> 32;

    // Reading "return value" at entry = garbage (rax not set yet)
    struct sock *sk = (struct sock *)PT_REGS_RC(ctx);
    e->family = BPF_CORE_READ(sk, __sk_common.skc_family);
    e->sport  = BPF_CORE_READ(sk, __sk_common.skc_num);
    e->dport  = 0;

    __builtin_memcpy(e->probe_type, "kprobe", 7);
    return 0;
}

// ── kretprobe: fires at inet_csk_accept RETURN ──────────────────────────────
// At return: sk = the accepted socket (real return value from rax).
// Reading family/sport/dport here gives correct values.
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
// At entry: fd and count arguments are available in registers.
// This is the correct usage for uprobe.
// Demonstrates: if called via Uretprobe() instead → reads garbage return value
SEC("uprobe/write")
int BPF_UPROBE(uprobe_write)
{
    __u32 key = 2;
    struct event *e = bpf_map_lookup_elem(&results, &key);
    if (!e)
        return 0;

    e->pid = bpf_get_current_pid_tgid() >> 32;

    __builtin_memcpy(e->probe_type, "uprobe", 7);
    return 0;
}

// ── uretprobe: fires at write() RETURN in libc ──────────────────────────────
// At return: retval = bytes written (or -errno on error).
// If called via Uprobe() instead → reads garbage (rax not set at entry).
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
