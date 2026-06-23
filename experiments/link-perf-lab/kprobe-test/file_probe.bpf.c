// go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char __license[] SEC("license") = "GPL";

struct event {
    __u64 timestamp;
    __u32 pid;
    __u32 tid;
    __u64 stack_id;
    __u64 stack_trace[128];
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __type(value, struct event);
} events SEC(".maps");

SEC("kprobe/sys_open")
int BPF_KPROBE(do_sys_open, const char *filename)
{
    bpf_printk("filename: %s\n", filename);
    return 0;
}
