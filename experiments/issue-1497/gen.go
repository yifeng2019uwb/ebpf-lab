// Run `go generate .` to rebuild BPF generated wrappers after modifying probe.bpf.c
//
// Requires on Linux host: clang, llvm, libbpf-dev
// Generated *_bpf*.go files should be committed so the experiment can build without clang.

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-I/usr/include/$(uname -m)-linux-gnu" probe ./probe.bpf.c

package main
