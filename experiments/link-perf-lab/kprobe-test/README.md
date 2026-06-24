# Lab: Multi-Syscall Tracepoint Monitor with Perf Events

**Status:** ✅ Working

---

## What It Does

Monitors 4 Linux syscalls using BPF tracepoints + perf event arrays:
- `openat` — file open operations
- `close` — file close operations
- `read` — file read operations
- `write` — file write operations

**Transport:** Perf event arrays (per-CPU ringbuffers) → userspace Go receiver

---

## Build & Run

```bash
# Generate BPF bindings (Linux/VM only)
make gen

# Build the program
make build

# Run (requires root for BPF)
sudo ./kprobe-test
```

---

## Key Learnings

See [NOTES.md](NOTES.md) for:
- Struct alignment issues (C ↔ Go binary serialization)
- Perf event arrays vs ringbuf comparison
- Tracepoint context structure and syscall arguments
- Signal handling with blocked I/O goroutines

---

## Files

- `file_tracepoint.bpf.c` — eBPF program (4 tracepoints)
- `main.go` — Go userspace reader + signal handling
- `Makefile` — Build automation
- `NOTES.md` — Technical implementation details

---

## Known Limitations

- `excludeTIDs` filtering is hardcoded (should be removed/improved)
- Only prints events from files in `/root/workspace/` path
- Need explicit signal handler for clean Ctrl+C exit

---

## Next Steps

- [ ] Remove hardcoded TID filtering
- [ ] Add file path filtering via command-line flags
- [ ] Test with larger syscall monitoring scope
