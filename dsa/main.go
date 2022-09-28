package main

// #cgo CFLAGS: -g -Wall -O2
// // #cgo LDFLAGS: -L ./person_orig -lperson
// #include <stdlib.h>
// #include "idxd/idxd_device.h"
// #include "idxd/idxd_device.c"
// // #include "person_orig/person.h"
// // #include "greet.h"
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"
)

func via_go(input []byte, output []byte, size uint32) {
	dsa_setup("/dev/dsa/wq0.0")

	dsa_memmove(input, output, size)

	res := bytes.Compare(input, output)
	if res == 0 {
		fmt.Println("dsa_memmove succeeded via go..!!\n")
	} else {
		fmt.Println("dsa_memmove failed via go..!!\n")
	}

	dsa_close()
}

func via_c(input []byte, output []byte, size uint32) {

	// input := C.malloc(C.sizeof_char * 4096)
	// C.memset(input, 'A', 4096)
	// output := C.malloc(C.sizeof_char * 4096)
	// C.memset(output, 0, 4096)

	C.dsa_setup(C.CString("/dev/dsa/wq0.0"))

	// C.dsa_memmove(unsafe.Pointer(input), unsafe.Pointer(output), C.u_int32_t(4096))
	C.dsa_memmove(unsafe.Pointer(&input[0]), unsafe.Pointer(&output[0]), C.u_int32_t(size))

	// inputSlice := C.GoBytes(input, 4096)
	// outputSlice := C.GoBytes(output, 4096)

	// res := bytes.Compare(inputSlice, outputSlice)
	res := bytes.Compare(input, output)
	if res == 0 {
		fmt.Println("dsa_memmove succeeded via c..!!\n")
	} else {
		fmt.Println("dsa_memmove failed via c..!!\n")
	}

	C.dsa_close()
}

func do_memmove() {

	input := make([]byte, 4096)
	output := make([]byte, 4096)

	for i := 0; i < 4096; i++ {
		input[i] = 'A'
		output[i] = '0'
	}

	via_c(input, output, 4096)

	// via_go(input, output, 4096)
}
func main() {
	// greet()

	// person()

	do_memmove()

}

// func greet() {
// 	name := C.CString("Nirav")
// 	defer C.free(unsafe.Pointer(name))

// 	year := C.int(2022)

// 	ptr := C.malloc(C.sizeof_char * 1024)
// 	defer C.free(unsafe.Pointer(ptr))

// 	size := C.greet(name, year, (*C.char)(ptr))

// 	b := C.GoBytes(ptr, size)
// 	fmt.Println(string(b))
// }

// type (
// 	Person C.struct_APerson
// )

// func person() {
// 	var f *Person
// 	f = (*Person)(C.get_person(C.CString("tim"), C.CString("tim hughes")))
// 	fmt.Printf("Hello Go world: My name is %s, %s.\n", C.GoString(f.name), C.GoString(f.long_name))
// }
