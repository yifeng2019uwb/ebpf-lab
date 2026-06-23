# Include path for vmlinux.h (located at ebpf-lab root)
# Without this: #include "vmlinux.h" would fail
# With this: clang finds vmlinux.h in parent directories
# Include path for vmlinux.h (located at ebpf-lab root)
CFLAGS += -I$(CURDIR)/../../..

gen:
	go generate

build:
	go build
	
run:
	./kprobe-test

# Compile + Build + Run
all: gen build run

# Just Build + Run (for after gen is done)
test: build run

clean:
	rm -f kprobe-test

.PHONY: gen build run all test clean