package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/cilium/ebpf/rlimit"
)

type FileEvent struct {
	PID       uint32
	TID       uint32
	UID       uint32
	GID       uint32
	Comm      [16]byte
	DFD       int32
	Filename  [256]byte
	Flags     int32
	Mode      int32
	Timestamp uint64
}

func main() {
	size := unsafe.Sizeof(FileEvent{})
	fmt.Printf("FileEvent size: %d bytes\n", size)

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("remove memlock: %v", err)
	}

	objs := fileTraceObjects{}
	if err := loadFileTrace(&objs, nil); err != nil {
		log.Fatalf("load BPF file objects: %v", err)
	}
	defer objs.Close()

}
