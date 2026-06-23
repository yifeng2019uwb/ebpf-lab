package main

import (
	"fmt"
	"unsafe"
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
}
