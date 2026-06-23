# Include path for vmlinux.h (located at ebpf-lab root)
# Without this: #include "vmlinux.h" would fail
# With this: clang finds vmlinux.h in parent directories
CFLAGS += -I$(CURDIR)/../../..

gen:
	go generate

build:
	go build
	
run:
	./kprobe-test

clean:
	rm -f kprobe-test

.PHONY: gen build run clean
