package main

// #cgo CFLAGS: -g -Wall -O0
// #include <stdlib.h>
// #include <string.h>
// #include "idxd_device.h"
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
	wq  mmap.MMap
	err error
	fd  *os.File
)

type (
	my_dsa_hw_desc           C.struct_dsa_hw_desc
	my_dsa_completion_record C.struct_dsa_completion_record
)

func dsa_setup(path string) {
	fmt.Println("Entering dsa_setup")
	fd, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("File open error %s: %s", path, err)
		return
	}
	defer fd.Close()

	wq, err = mmap.MapRegion(fd, 0x1000, mmap.RDWR, unix.MAP_SHARED|unix.MAP_POPULATE, 0)
	if err != nil {
		fmt.Println("File mapping error %s", err)
		return
	}
	fmt.Println("Exiting dsa_setup")
}

func dsa_close() {
	fmt.Println("Entering dsa_close")
	wq.Unmap()
	fmt.Println("Exiting dsa_close")
}

func dsa_memmove(input []byte, output []byte, size uint32) int {
	fmt.Println("Entering dsa_memmove")
	// var desc my_dsa_hw_desc
	// var comp my_dsa_completion_record

	// desc.src_addr = (unsafe.Pointer(&input[0]))
	// desc.dst_addr = (unsafe.Pointer(&output[0]))
	// desc.xfer_size = size
	// desc.opcode = 3 //DSA_OPCODE_MEMMOVE
	// desc.completion_addr = &comp
	//desc.flags = IDXD_OP_FLAG_CRAV | IDXD_OP_FLAG_RCR

	// C.dsa_desc_submit(wq, 0, (void*)&desc);

	// while (comp.status == 0) {
	//         umonitor(&comp);
	// }

	// if (!memcmp(input, output, size))
	// 	printf("%s : dsa_memmove succeeeded!\n", __func__);
	// else
	// 	printf("%s : dsa_memmove failed!\n", __func__);

	fmt.Println("Exiting dsa_memmove")
	return 0
}

func via_c() {
	C.dsa_setup(C.CString("/dev/dsa/wq0.0"))

	input := C.malloc(C.sizeof_char * 4096)
	output := C.malloc(C.sizeof_char * 4096)

	C.memset(input, 'A', 4096)
	C.memset(output, 0, 4096)

	C.dsa_memmove(unsafe.Pointer(input), unsafe.Pointer(output), C.u_int32_t(4096))

	// inputSlice := C.GoBytes(input, 4096)
	// outputSlice := C.GoBytes(output, 4096)

	// res := bytes.Compare(inputSlice, outputSlice)
	// if res == 0 {
	// 	fmt.Println("dsa_memmove succeeded..!!\n")
	// } else {
	// 	fmt.Println("dsa_memmove failed..!!\n")
	// }

	C.dsa_close()
}

func via_go() {
	// dsa_setup("/dev/dsa/wq0.0")

	// input := make([]byte, 4096)
	// output := make([]byte, 4096)

	// for i := 0; i < 4096; i++ {
	// 	input[i] = 'A'
	// 	output[i] = '0'
	// }

	// dsa_memmove(input, output, 4096)

	// res := bytes.Compare(input, output)
	// if res == 0 {
	// 	fmt.Println("dsa_memmove succeeded..!!\n")
	// } else {
	// 	fmt.Println("dsa_memmove failed..!!\n")
	// }

	// dsa_close()
}

func main() {

	// name := C.CString("Nirav")
	// defer C.free(unsafe.Pointer(name))

	// year := C.int(2022)

	// ptr := C.malloc(C.sizeof_char * 1024)
	// // ptr2 := C.malloc(C.sizeof_struct_dsa_hw_desc)
	// defer C.free(unsafe.Pointer(ptr))

	// size := C.greet(name, year, (*C.char)(ptr))

	// //wq := C.wq_map(C.CString("dsa2"), 0, 0, 0)

	// //C.dsa_desc_submit(unsafe.Pointer(wq), 0, unsafe.Pointer(ptr2))

	// // ret := C.entry_point()
	// // fmt.Println((string(ret)))

	// b := C.GoBytes(ptr, size)
	// fmt.Println(string(b))

	via_c()

}
