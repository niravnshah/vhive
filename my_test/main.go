package main

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/intel/idxd"
)

var (
	copyLen uint32 = 4096
	elem    uint32 = 100
)

func via_go(input []byte, output []byte, size uint32) {
	idxd.DSA_setup_c("/dev/dsa/wq0.0")

	result := idxd.DSA_memmove_sync_go(output, input, size*elem)

	if result != 0 {
		fmt.Println("Copy didnot complete")
	}

	res := bytes.Compare(input, output)
	if res == 0 {
		fmt.Println("dsa_memmove_sync succeeded via go..!!\n")
	} else {
		fmt.Println("dsa_memmove_sync failed via go..!!\n")
	}

	for i := uint32(0); i < copyLen*elem; i++ {
		input[i] = 'A'
		output[i] = '0'
	}

	desc_mem := idxd.AlignedBlock(64*int(elem), 64)
	comp_mem := idxd.AlignedBlock(32*int(elem), 32)

	for idx := uint32(0); idx < elem; idx++ {
		desc := (*idxd.DSA_hw_desc_go)(unsafe.Pointer(&desc_mem[64*idx]))
		comp := (*idxd.DSA_completion_record_go)(unsafe.Pointer(&comp_mem[32*idx]))

		desc.Src_addr = (*uint64)(unsafe.Pointer(&input[size*idx]))
		desc.Dst_addr = (*uint64)(unsafe.Pointer(&output[size*idx]))
		desc.Xfer_size = size
		desc.Opcode = idxd.DSA_OPCODE_MEMMOVE
		desc.Completion_addr = (*uint64)(unsafe.Pointer(comp))
		desc.Flags[0] = idxd.IDXD_OP_FLAG_CRAV | idxd.IDXD_OP_FLAG_RCR

		// idxd.DSA_memmove_sync_go(output[size*idx:], input[size*idx:], size)
		// res = bytes.Compare(input[size*idx:(size+1)*idx], output[size*idx:(size+1)*idx])
		// if res == 0 {
		// 	fmt.Printf("dsa_memmove_sync succeeded via go.. input =%p output=%p!!\n", &input[size*idx], &output[size*idx])
		// } else {
		// 	fmt.Printf("dsa_memmove_sync failed via go..!!\n")
		// }
		idxd.DSA_memmove_desc_go(desc, 0)
	}

	for idx := uint32(0); idx < elem; idx++ {
		desc := (*idxd.DSA_hw_desc_go)(unsafe.Pointer(&desc_mem[64*idx]))
		idxd.DSA_wait_for_comp_go(desc)
	}

	res = bytes.Compare(input, output)
	if res == 0 {
		fmt.Println("dsa_memmove_sync succeeded via go..!!\n")
	} else {
		fmt.Println("dsa_memmove_sync failed via go..!!\n")
	}

	idxd.DSA_close_c()
}

func via_c(input []byte, output []byte, size uint32) {

	// input := C.malloc(C.sizeof_char * copyLen)
	// C.memset(input, 'A', copyLen)
	// output := C.malloc(C.sizeof_char * copyLen)
	// C.memset(output, 0, copyLen)

	idxd.DSA_setup_c("/dev/dsa/wq0.0")

	// C.dsa_memmove_sync(unsafe.Pointer(output), unsafe.Pointer(input), C.u_int32_t(copyLen))
	result := idxd.DSA_memmove_sync_c(output, input, size)

	if result != 0 {
		fmt.Println("Copy didnot complete. Result = ", result)
	}

	// inputSlice := C.GoBytes(input, copyLen)
	// outputSlice := C.GoBytes(output, copyLen)

	// res := bytes.Compare(inputSlice, outputSlice)
	res := bytes.Compare(input, output)
	if res == 0 {
		fmt.Println("dsa_memmove_sync succeeded via c..!!\n")
	} else {
		fmt.Println("dsa_memmove_sync failed via c..!!\n")
	}

	idxd.DSA_close_c()
}

func do_memmove() {

	input := make([]byte, copyLen*uint32(elem))
	output := make([]byte, copyLen*uint32(elem))

	for i := uint32(0); i < copyLen*uint32(elem); i++ {
		input[i] = 'A'
		output[i] = '0'
	}

	via_c(input, output, copyLen*uint32(elem))

	for i := uint32(0); i < copyLen*uint32(elem); i++ {
		input[i] = 'A'
		output[i] = '0'
	}

	via_go(input, output, copyLen)
}

func do_move_pages()
{
	// filepath := s.GuestMemPath         //"/home/nshah5/Linux-test/uffdio/abc.bin" //
	// filesize := uint64(s.GuestMemSize) //uint64(26 * 4096)                        //

	// fd, err := os.OpenFile(filepath, os.O_RDONLY, 0444)
	// if err != nil {
	// 	log.Errorf("Failed to open guest memory file: %v", err)
	// 	return err
	// }
	// defer fd.Close()

	// guestMem, err := unix.Mmap(int(fd.Fd()), 0, int(filesize), unix.PROT_READ, unix.MAP_PRIVATE)
	// if err != nil {
	// 	log.Errorf("Failed to mmap guest memory file: %v", err)
	// 	return err
	// }
	// log.Infof("GuestMemFile size = %d against %d", filesize, len(guestMem))

	// str := "Contents -> "
	// for i := uint64(0); i < nb_pages; i++ {
	// 	str += string(guestMem[i*page_size])
	// }
	// log.Infof(str)

	// if err := unix.Munmap(guestMem); err != nil {
	// 	log.Errorf("Failed to munmap guest memory file: %v", err)
	// 	return err
	// }
}

func main() {
	do_memmove()
}
