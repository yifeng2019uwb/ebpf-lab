module github.com/yifeng2019uwb/ebpf-lab

go 1.25.0

require github.com/cilium/ebpf v0.17.3

require golang.org/x/sys v0.43.0 // indirect

// Test against local cilium/ebpf changes:
replace github.com/cilium/ebpf => /Users/yifengzh/workspace/ebpf
