package idxd

// #cgo CFLAGS: -g -Wall -O0
// #include "dsa.h"
// #include "idxd_device.h"
import "C"
import (
	"fmt"
	"unsafe"
	// "github.com/edsrzf/mmap-go"
)

func DSA_setup_c(path string) {
	C.dsa_setup(C.CString(path))
}

func DSA_close_c() {
	C.dsa_close()
}

func DSA_memmove_sync_c(dst []byte, src []byte, size uint32) uint32 {
	ret := C.dsa_memmove_sync(unsafe.Pointer(&dst[0]), unsafe.Pointer(&src[0]), C.u_int32_t(size))
	return uint32(ret)
}

// var (
// 	work_queue mmap.MMap
// 	err        error
// 	fd         *os.File
// )

type DSA_hw_desc_go struct {
	unused1         uint32
	Flags           [3]uint8
	Opcode          uint8
	Completion_addr *uint64
	Src_addr        *uint64
	Dst_addr        *uint64
	Xfer_size       uint32
	unused2         uint32
	unused3         [3]uint64
}

type DSA_completion_record_go struct {
	Status          uint8
	res1            [3]uint8
	Bytes_completed uint32
	unused          [24]uint8
}

const (
	DSA_OPCODE_MEMMOVE       = 3
	MAX_COMP_RETRY           = 2000000000
	IDXD_FLAG_BLOCK_ON_FAULT = 1 << 1
	IDXD_FLAG_CRAV           = 1 << 2
	IDXD_FLAG_RCR            = 1 << 3
)

// func DSA_setup_go(path string) {
// 	fmt.Println("Entering dsa_setup")
// 	fd, err = os.OpenFile(path, os.O_RDWR, 0666)
// 	if err != nil {
// 		fmt.Println("File open error %s: %s", path, err)
// 		return
// 	}
// 	defer fd.Close()

// 	work_queue, err = mmap.MapRegion(fd, 0x1000, mmap.RDWR, unix.MAP_SHARED|unix.MAP_POPULATE, 0)
// 	if err != nil {
// 		fmt.Println("File mapping error %s", err)
// 		return
// 	}
// 	fmt.Println("Exiting dsa_setup")
// }

// func DSA_close_go() {
// 	fmt.Println("Entering dsa_close")
// 	work_queue.Unmap()
// 	fmt.Println("Exiting dsa_close")
// }

func DSA_memmove_sync_go(output []byte, input []byte, size uint32) uint32 {
	// fmt.Println("Entering dsa_memmove_go")
	// var desc *dsa_hw_desc_go = new(dsa_hw_desc_go)
	// desc := make([]byte, 64)

	// desc := (*dsa_hw_desc_go)(C.malloc(64))
	desc_mem := AlignedBlock(64, 64)
	desc := (*DSA_hw_desc_go)(unsafe.Pointer(&desc_mem[0]))

	// comp := (*dsa_completion_record_go)(C.malloc(32))
	comp_mem := AlignedBlock(32, 32)
	comp := (*DSA_completion_record_go)(unsafe.Pointer(&comp_mem[0]))

	desc.Src_addr = (*uint64)(unsafe.Pointer(&input[0]))
	desc.Dst_addr = (*uint64)(unsafe.Pointer(&output[0]))
	desc.Xfer_size = size
	desc.Opcode = DSA_OPCODE_MEMMOVE
	desc.Completion_addr = (*uint64)(unsafe.Pointer(comp))
	desc.Flags[0] = IDXD_FLAG_BLOCK_ON_FAULT | IDXD_FLAG_CRAV | IDXD_FLAG_RCR

	// C.dsa_desc_submit((unsafe.Pointer)(&work_queue[0]), 0, (unsafe.Pointer)(desc))
	status := DSA_memmove_desc_go(desc, 1)

	// fmt.Println("Exiting dsa_memmove_go")
	return status
}

func DSA_memmove_desc_go(hw_desc *DSA_hw_desc_go, sync uint) uint32 {

	C.dsa_memmove_desc_wrapper(unsafe.Pointer(hw_desc), C.u_int32_t(sync))

	if sync != 0 {
		comp := (*DSA_completion_record_go)((unsafe.Pointer)(hw_desc.Completion_addr))

		DSA_wait_for_comp_go(hw_desc)
		if comp.Status != 1 {
			fmt.Printf("comp.Status = %d\n", comp.Status)
			return uint32(comp.Status)
		} else {
			return 0
		}
	}
	return 0
}

func DSA_wait_for_comp_go(hw_desc *DSA_hw_desc_go) uint32 {
	var retry uint32 = 0
	comp := (*DSA_completion_record_go)((unsafe.Pointer)(hw_desc.Completion_addr))
	for retry := 0; comp.Status == 0 && retry < MAX_COMP_RETRY; retry++ {
		C.umonitor(unsafe.Pointer(comp))
	}

	if retry >= MAX_COMP_RETRY {
		fmt.Println("Desc timeout!!")
		return 1
	}
	return 0
}
