package main

import (
	"fmt"
	"unsafe"
)

// alignment returns alignment of the block in memory
// with reference to alignSize
//
// Can't check alignment of a zero sized block as &block[0] is invalid
func alignment(block []byte, alignSize int) int {
	return int(uintptr(unsafe.Pointer(&block[0])) & uintptr(alignSize-1))
}

// AlignedBlock returns []byte of size BlockSize aligned to a multiple
// of alignSize in memory (must be power of two)
func AlignedBlock(blockSize int, alignSize int) []byte {
	if blockSize == 0 {
		return nil
	}

	block := make([]byte, blockSize+alignSize)

	a := alignment(block, alignSize)
	offset := 0
	if a != 0 {
		offset = alignSize - a
	}
	block = block[offset : offset+blockSize]

	// Check
	if blockSize != 0 {
		a = alignment(block, alignSize)
		if a != 0 {
			fmt.Println("Failed to align block")
		}
	}
	return block
}
