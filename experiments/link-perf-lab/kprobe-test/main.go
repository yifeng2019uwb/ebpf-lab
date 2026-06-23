package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"unsafe"

	"encoding/binary"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
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

	var objs fileTraceObjects
	if err := loadFileTraceObjects(&objs, nil); err != nil {
		log.Fatalf("load BPF objects: %v", err)
	}
	defer objs.Close()

	fmt.Println("BPF program loaded successfully!")

	fmt.Printf("Programs: %+v\n", objs.fileTracePrograms)
	fmt.Printf("Maps: %+v\n", objs.fileTraceMaps)
	fmt.Printf("Variables: %+v\n", objs.fileTraceVariables)
	fmt.Printf("  Map events: %v\n", objs.fileTraceMaps.Events)

	// After loading BPF objects
	tp, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.fileTracePrograms.HandleSysOpen, nil)
	if err != nil {
		log.Fatalf("attach tracepoint: %v", err)
	}
	defer tp.Close()

	fmt.Println("Tracepoint attached!")

	// Create perf event reader
	reader, err := perf.NewReader(objs.fileTraceMaps.Events, 4096)
	if err != nil {
		log.Fatalf("create perf reader: %v", err)
	}
	fmt.Println("BPF Event NewReader successfully!")
	defer reader.Close()

	fmt.Println("Listening for events... (open a file to trigger)")

	// Read events in a loop
	for {
		record, err := reader.Read()
		if err != nil {
			log.Fatalf("read: %v", err)
		}

		// Parse raw bytes into FileEvent
		var event FileEvent
		if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parse event: %v", err)
			continue
		}

		filename := strings.TrimRight(string(event.Filename[:]), "\x00")
		fmt.Printf("Event: PID=%d, GID=%d, UID=%d, TID=%d, File=%s Flags=%d, timestamp=%d\n",
			event.PID, event.GID, event.UID, event.TID, filename, event.Flags, event.Timestamp)

	}

}
