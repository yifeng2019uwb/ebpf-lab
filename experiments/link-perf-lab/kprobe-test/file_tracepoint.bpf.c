//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define MAX_FILENAME_LEN 256
#define O_RDONLY   0     // (read only)
#define O_WRONLY   1     // (write only)
#define O_RDWR     2     // (read + write)
#define O_CREAT    64    // (create if not exists)
#define O_APPEND   1024  // (append mode)

struct file_open_event {
    // Process info
    __u32 pid;              // process ID
    __u32 tid;              // thread ID
    __u32 uid;              // user ID
    __u32 gid;              // group ID
    char comm[16];          // process name
    
    // File operation details
    int dfd;                // directory fd
    char filename[256];     // file path
    int flags;              // open flags
    int mode;               // file mode
    
    // Timing
    __u64 timestamp;        // when event occurred
    
    // Context
    // __u64 kernel_stack_id;  // kernel call stack
    // __u64 stack_trace[128]; // kernel call stack trace, this is too large
}; // 4+4+4+4+16+4+256+4+4+4 + 8  = 312 bytes <= 512 bytes

struct file_close_event {
        // Process info
        __u32 pid;              // process ID
        __u32 tid;              // thread ID
        __u32 uid;              // user ID
        __u32 gid;              // group ID

        // File operation details
        int fd;               // directory fd  ← change to: // file descriptor
        __u32 pad;        
        // Timing
        __u64 timestamp;        // when event occurred
};

struct file_read_write_event {
    // Process info
    __u32 pid;              // process ID
    __u32 tid;              // thread ID
    __u32 uid;              // user ID
    __u32 gid;              // group ID
    
    // File operation details
    int fd;                // directory fd  ← change to: // file descriptor
    __u32 pad;
    __u64 size;              // bytes to read/write
    
    // Timing
    __u64 timestamp;        // when event occurred
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    // __uint(value_size, sizeof(struct file_event)); // if self defined event, need use ring buf
    __uint(max_entries, 1024);
//    __type(value, struct file_event);
} open_events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    // __uint(value_size, sizeof(struct file_event)); // if self defined event, need use ring buf
    __uint(max_entries, 1024);
//    __type(value, struct file_event);
} close_events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    // __uint(value_size, sizeof(struct file_event)); // if self defined event, need use ring buf
    __uint(max_entries, 1024);
//    __type(value, struct file_event);
} read_events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    // __uint(value_size, sizeof(struct file_event)); // if self defined event, need use ring buf
    __uint(max_entries, 1024);
//    __type(value, struct file_event);
} write_events SEC(".maps");

SEC("tracepoint/syscalls/sys_enter_openat") 
int handle_sys_open(struct trace_event_raw_sys_enter *ctx)
{
    struct file_open_event ev = {};

    u64 id = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;
    u32 tid = (u32)id;
    ev.pid = tgid;
    ev.tid = tid;

    __u64 uid_gid = bpf_get_current_uid_gid();
    ev.uid = uid_gid >> 32; // get user ID
    ev.gid = uid_gid & 0xFFFFFFFF; // get group ID
    bpf_get_current_comm(&ev.comm, sizeof(ev.comm)); // read process name

    ev.dfd = ctx->args[0];

    // args[0] = dirfd (int) - the directory file descriptor
    // args[1] = pathname (pointer) - address of filename string in user-space
    // args[2] = flags (int) - file flags
    // args[3] = mode (int) - file mode
    // ctx->args[1] is the filename pointer, we need to read the filename from the kernel space
    // bpf_probe_read_kernel_str(&ev.filename, sizeof(ev.filename), (const char *)ctx->args[1]);
    bpf_probe_read_user_str(&ev.filename, sizeof(ev.filename), (const char *)ctx->args[1]);

    ev.flags = ctx->args[2];
    ev.mode = ctx->args[3];
    ev.timestamp = bpf_ktime_get_ns();

    bpf_perf_event_output(ctx, &open_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev)); //save event to current cpu(0)BPF_F_CURRENT_CPU
    return 0;
}  

SEC("tracepoint/syscalls/sys_enter_close") 
int handle_sys_close(struct trace_event_raw_sys_enter *ctx)
{
    struct file_close_event ev = {};

    u64 id = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;
    u32 tid = (u32)id;
    ev.pid = tgid;
    ev.tid = tid;

    __u64 uid_gid = bpf_get_current_uid_gid();
    ev.uid = uid_gid >> 32; // get user ID
    ev.gid = uid_gid & 0xFFFFFFFF; // get group ID
  
    ev.fd = ctx->args[0];
    ev.timestamp = bpf_ktime_get_ns();

    bpf_perf_event_output(ctx, &close_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev)); //save event to current cpu(0)BPF_F_CURRENT_CPU
    return 0;
}  

SEC("tracepoint/syscalls/sys_enter_read") 
int handle_sys_read(struct trace_event_raw_sys_enter *ctx)
{
    struct file_read_write_event ev = {};

    u64 id = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;
    u32 tid = (u32)id;
    ev.pid = tgid;
    ev.tid = tid;

    __u64 uid_gid = bpf_get_current_uid_gid();
    ev.uid = uid_gid >> 32; // get user ID
    ev.gid = uid_gid & 0xFFFFFFFF; // get group ID

    //ssize_t read(int fd, void *buf, size_t count)
    //                  ↓      ↓        ↓
    //               args[0] args[1] args[2]
    ev.fd = ctx->args[0];
    ev.size = ctx->args[2];  // count (bytes to read)

    ev.timestamp = bpf_ktime_get_ns();

    bpf_perf_event_output(ctx, &read_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev)); //save event to current cpu(0)BPF_F_CURRENT_CPU
    return 0;
} 

SEC("tracepoint/syscalls/sys_enter_write") 
int handle_sys_write(struct trace_event_raw_sys_enter *ctx)
{
    struct file_read_write_event ev = {};

    u64 id = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;
    u32 tid = (u32)id;
    ev.pid = tgid;
    ev.tid = tid;

    __u64 uid_gid = bpf_get_current_uid_gid();
    ev.uid = uid_gid >> 32; // get user ID
    ev.gid = uid_gid & 0xFFFFFFFF; // get group ID

    // ssize_t write(int fd, const void *buf, size_t count)
    //                  ↓      ↓               ↓
    //              args[0] args[1]       args[2]
    ev.fd = ctx->args[0];
    ev.size = ctx->args[2];  // count (bytes to write)
    ev.timestamp = bpf_ktime_get_ns();

    bpf_perf_event_output(ctx, &write_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev)); //save event to current cpu(0)BPF_F_CURRENT_CPU
    return 0;
} 


char __license[] SEC("license") = "GPL";
