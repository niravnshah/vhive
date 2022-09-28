package main

// #cgo CFLAGS: -g -Wall -O2
// #include "idxd/dsa.h"
// #include <linux/idxd.h>
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/edsrzf/mmap-go"
	"golang.org/x/sys/unix"
)

var (
	work_queue mmap.MMap
	err        error
	fd         *os.File
)

type dsa_hw_desc_go struct {
	unused1         uint32
	flags           [3]uint8
	opcode          uint8
	completion_addr *uint64
	src_addr        *uint64
	dst_addr        *uint64
	xfer_size       uint32
	unused2         uint32
	unused3         [3]uint64
}

type dsa_completion_record_go struct {
	status uint8
	unused [31]uint8
}

const (
	DSA_OPCODE_MEMMOVE = 3
	MAX_COMP_RETRY     = 2000000000
	IDXD_OP_FLAG_CRAV  = 0x4
	IDXD_OP_FLAG_RCR   = 0x8
)

func dsa_setup(path string) {
	fmt.Println("Entering dsa_setup")
	fd, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("File open error %s: %s", path, err)
		return
	}
	defer fd.Close()

	work_queue, err = mmap.MapRegion(fd, 0x1000, mmap.RDWR, unix.MAP_SHARED|unix.MAP_POPULATE, 0)
	if err != nil {
		fmt.Println("File mapping error %s", err)
		return
	}
	fmt.Println("Exiting dsa_setup")
}

func dsa_close() {
	fmt.Println("Entering dsa_close")
	work_queue.Unmap()
	fmt.Println("Exiting dsa_close")
}

func dsa_memmove(input []byte, output []byte, size uint32) int {
	fmt.Println("Entering dsa_memmove")
	// var desc *dsa_hw_desc_go = new(dsa_hw_desc_go)
	// desc := make([]byte, 64)
	desc_mem := AlignedBlock(64, 64)
	comp_mem := AlignedBlock(32, 32)

	// desc := (*dsa_hw_desc_go)(C.malloc(64))
	desc := (*dsa_hw_desc_go)(unsafe.Pointer(&desc_mem[0]))

	// comp := (*dsa_completion_record_go)(C.malloc(32))
	comp := (*dsa_completion_record_go)(unsafe.Pointer(&comp_mem[0]))

	desc.src_addr = (*uint64)(unsafe.Pointer(&input[0]))
	desc.dst_addr = (*uint64)(unsafe.Pointer(&output[0]))
	desc.xfer_size = size
	desc.opcode = DSA_OPCODE_MEMMOVE
	desc.completion_addr = (*uint64)(unsafe.Pointer(comp))
	desc.flags[0] = IDXD_OP_FLAG_CRAV | IDXD_OP_FLAG_RCR

	C.dsa_desc_submit((unsafe.Pointer)(&work_queue[0]), 0, (unsafe.Pointer)(desc))

	for retry := 0; comp.status == 0 && retry < MAX_COMP_RETRY; retry++ {
		C.umonitor(unsafe.Pointer(comp))
	}

	fmt.Println("Exiting dsa_memmove")
	return 0
}
