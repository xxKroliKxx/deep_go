package main

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

func Defragment(memory []byte, pointers []unsafe.Pointer) {
	if len(memory) == 0 || len(pointers) == 0 {
		return
	}

	base := unsafe.Pointer(&memory[0])

	locMap := make(map[int][]int, len(pointers))
	for i, p := range pointers {
		offset := int(uintptr(p) - uintptr(base))
		locMap[offset] = append(locMap[offset], i)
	}

	nextIdx := 0
	for i := 0; i < len(memory); i++ {
		if ptrIdx, ok := locMap[i]; ok {
			if nextIdx != i {
				memory[nextIdx] = memory[i]
			}

			for _, pi := range ptrIdx {
				pointers[pi] = unsafe.Pointer(&memory[nextIdx])
			}

			nextIdx++
		}
	}

	for i := nextIdx; i < len(memory); i++ {
		memory[i] = 0
	}
}

func TestDefragmentation(t *testing.T) {
	var fragmentedMemory = []byte{
		0xFF, 0x00, 0x00, 0x00,
		0x00, 0xFF, 0x00, 0x00,
		0x00, 0x00, 0xFF, 0x00,
		0x00, 0x00, 0x00, 0xFF,
	}

	var fragmentedPointers = []unsafe.Pointer{
		unsafe.Pointer(&fragmentedMemory[0]),
		unsafe.Pointer(&fragmentedMemory[5]),
		unsafe.Pointer(&fragmentedMemory[10]),
		unsafe.Pointer(&fragmentedMemory[15]),
	}

	var defragmentedPointers = []unsafe.Pointer{
		unsafe.Pointer(&fragmentedMemory[0]),
		unsafe.Pointer(&fragmentedMemory[1]),
		unsafe.Pointer(&fragmentedMemory[2]),
		unsafe.Pointer(&fragmentedMemory[3]),
	}

	var defragmentedMemory = []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	Defragment(fragmentedMemory, fragmentedPointers)
	assert.True(t, reflect.DeepEqual(defragmentedMemory, fragmentedMemory))
	assert.True(t, reflect.DeepEqual(defragmentedPointers, fragmentedPointers))
}
