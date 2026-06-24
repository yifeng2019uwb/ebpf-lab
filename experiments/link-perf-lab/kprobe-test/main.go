package main

import (
	"bytes"
	"fmt"
	"log"
	"slices"
	"strings"

	// "unsafe"

	"encoding/binary"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
)

type OpenFileEvent struct {
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

type CloseFileEvent struct {
	PID       uint32
	TID       uint32
	UID       uint32
	GID       uint32
	FD        int32
	PAD	  uint32
	Timestamp uint64
}

type ReadWriteFileEvent struct {
	PID       uint32
	TID       uint32
	UID       uint32
	GID       uint32
	FD        int32
	PAD       uint32
	SIZE      uint64
	Timestamp uint64
}

var excludeTIDs = [6]uint32{15314, 285, 15312, 15316, 15318, 15320}

func main() {
	// size := unsafe.Sizeof(OpenFileEvent{})
	// fmt.Printf("FileEvent size: %d bytes\n", size)

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
	fmt.Printf("Map events: %v\n", objs.fileTraceMaps.OpenEvents)

	// After loading BPF sys_enter_openat objects
	tp_fo, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.fileTracePrograms.HandleSysOpen, nil)
	if err != nil {
		log.Fatalf("attach tracepoint: %v", err)
	}
	defer tp_fo.Close()

	fmt.Println("Tracepoint sys_enter_openat attached!")

	// Create perf file open event reader
	oe_reader, err := perf.NewReader(objs.fileTraceMaps.OpenEvents, 4096)
	if err != nil {
		log.Fatalf("create open perf event reader: %v", err)
	}
	fmt.Println("BPF Open Events NewReader successfully!")
	defer oe_reader.Close()

	fmt.Println("Listening for open events... (open a file to trigger)")

	// Create perf file close event reader
	tp_fc, err := link.Tracepoint("syscalls", "sys_enter_close", objs.fileTracePrograms.HandleSysClose, nil)
	if err != nil {
		log.Fatalf("attach tracepoint: %v", err)
	}
	defer tp_fc.Close()

	fmt.Println("Tracepoint sys_enter_close attached!")

	ce_reader, err := perf.NewReader(objs.fileTraceMaps.CloseEvents, 4096)
	if err != nil {
		log.Fatalf("create close event perf event reader: %v", err)
	}
	fmt.Println("BPF Close Event NewReader successfully!")
	defer ce_reader.Close()

	fmt.Println("Listening for close events... (close a file to trigger)")

	// Create perf file read event reader
	tp_rf, err := link.Tracepoint("syscalls", "sys_enter_read", objs.fileTracePrograms.HandleSysRead, nil)
	if err != nil {
		log.Fatalf("attach tracepoint: %v", err)
	}
	defer tp_rf.Close()

	fmt.Println("Tracepoint sys_enter_read attached!")

	rf_reader, err := perf.NewReader(objs.fileTraceMaps.ReadEvents, 4096)
	if err != nil {
		log.Fatalf("create read event perf event reader: %v", err)
	}
	fmt.Println("BPF Read Event NewReader successfully!")
	defer rf_reader.Close()

	fmt.Println("Listening for read events... (read a file to trigger)")

	// Create perf file read event reader
	tp_wf, err := link.Tracepoint("syscalls", "sys_enter_write", objs.fileTracePrograms.HandleSysWrite, nil)
	if err != nil {
		log.Fatalf("attach tracepoint: %v", err)
	}
	defer tp_wf.Close()

	fmt.Println("Tracepoint sys_enter_write attached!")

	wf_reader, err := perf.NewReader(objs.fileTraceMaps.WriteEvents, 4096)
	if err != nil {
		log.Fatalf("create write event perf event reader: %v", err)
	}
	fmt.Println("BPF Write Event NewReader successfully!")
	defer wf_reader.Close()

	fmt.Println("Listening for write events... (write a file to trigger)")

	// Open Event READ goroutine
	go func() {
		for {
			oe, err := oe_reader.Read()
			if err != nil {
				log.Fatalf("read: %v", err)
			}

			// Parse raw bytes into FileEvent
			var event_open OpenFileEvent
			if err := binary.Read(bytes.NewReader(oe.RawSample), binary.LittleEndian, &event_open); err != nil {
				log.Printf("parse event: %v", err)
				continue
			}

			filename := strings.TrimRight(string(event_open.Filename[:]), "\x00")
			if strings.HasPrefix(filename, "/root/workspace/") {
				fmt.Printf("Event: PID=%d, GID=%d, UID=%d, TID=%d, File=%s Flags=%d, timestamp=%d\n",
					event_open.PID, event_open.GID, event_open.UID, event_open.TID, filename, event_open.Flags, event_open.Timestamp)
			}
		}
	}()

	// Close Event READ goroutine
	go func() {
		for {
			ce, err := ce_reader.Read()
			if err != nil {
				log.Fatalf("read: %v", err)
			}

			// Parse raw bytes into FileEvent
			var event_close CloseFileEvent
			if err := binary.Read(bytes.NewReader(ce.RawSample), binary.LittleEndian, &event_close); err != nil {
				log.Printf("parse event: %v", err)
				continue
			}

			exist := slices.Contains(excludeTIDs[:], event_close.TID)
			if !exist {
				fmt.Printf("Event: PID=%d, GID=%d, UID=%d, TID=%d, fd=%d, timestamp=%d\n",
					event_close.PID, event_close.GID, event_close.UID, event_close.TID, event_close.FD, event_close.Timestamp)
			}
		}
	}()

	// Read Event READ goroutine
	go func() {
		for {
			rf, err := rf_reader.Read()
			if err != nil {
				log.Fatalf("read: %v", err)
			}

			// Parse raw bytes into FileEvent
			var read_event ReadWriteFileEvent
			if err := binary.Read(bytes.NewReader(rf.RawSample), binary.LittleEndian, &read_event); err != nil {
				log.Printf("parse event: %v", err)
				continue
			}

			exist := slices.Contains(excludeTIDs[:], read_event.TID)
			if !exist {
				fmt.Printf("Event: PID=%d, GID=%d, UID=%d, TID=%d, SIZE=%d, timestamp=%d\n",
					read_event.PID, read_event.GID, read_event.UID, read_event.TID, read_event.SIZE, read_event.Timestamp)
			}

		}
	}()

	// Write Event READ goroutine
	go func() {
		for {
			wf, err := wf_reader.Read()
			if err != nil {
				log.Fatalf("read: %v", err)
			}

			// Parse raw bytes into FileEvent
			var write_event ReadWriteFileEvent
			if err := binary.Read(bytes.NewReader(wf.RawSample), binary.LittleEndian, &write_event); err != nil {
				log.Printf("parse event: %v", err)
				continue
			}

			exist := slices.Contains(excludeTIDs[:], write_event.TID)
			if !exist {
				fmt.Printf("Event: PID=%d, GID=%d, UID=%d, TID=%d, SIZE=%d, timestamp=%d\n",
					write_event.PID, write_event.GID, write_event.UID, write_event.TID, write_event.SIZE, write_event.Timestamp)
			}
		}
	}()

	select {} // Block forever until Ctrl+C

}
