package main

import (
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type bufferData struct {
	ref  int32
	data []byte
}

type COWBuffer struct {
	d *bufferData
}

func NewCOWBuffer(data []byte) COWBuffer {
	return COWBuffer{&bufferData{
		ref:  1,
		data: data,
	}}
}

func (b *COWBuffer) Clone() COWBuffer {
	if b.d != nil {
		atomic.AddInt32(&b.d.ref, 1)
	}

	return COWBuffer{b.d}
}

func (b *COWBuffer) Close() {
	if b.d == nil {
		return
	}

	if atomic.AddInt32(&b.d.ref, -1) == 0 {
		b.d.data = nil
	}

	b.d = nil
}

func (b *COWBuffer) Update(index int, value byte) bool {
	if b.d == nil || index < 0 || index >= len(b.d.data) {
		return false
	}

	if atomic.LoadInt32(&b.d.ref) > 1 {
		newData := make([]byte, len(b.d.data))

		copy(newData, b.d.data)

		atomic.AddInt32(&b.d.ref, -1)

		b.d = &bufferData{ref: 1, data: newData}
	}

	b.d.data[index] = value

	return true
}

func (b *COWBuffer) String() string {
	if b.d == nil {
		return ""
	}

	data := b.d.data
	if len(data) == 0 {
		return ""
	}

	ptr := unsafe.SliceData(data)
	return unsafe.String(ptr, len(data))
}

func TestCOWBuffer(t *testing.T) {
	data := []byte{'a', 'b', 'c', 'd'}
	buffer := NewCOWBuffer(data)
	defer buffer.Close()

	copy1 := buffer.Clone()
	copy2 := buffer.Clone()

	assert.Equal(t, unsafe.SliceData(data), unsafe.SliceData(buffer.d.data))
	assert.Equal(t, unsafe.SliceData(buffer.d.data), unsafe.SliceData(copy1.d.data))
	assert.Equal(t, unsafe.SliceData(copy1.d.data), unsafe.SliceData(copy2.d.data))

	assert.True(t, (*byte)(unsafe.SliceData(data)) == unsafe.StringData(buffer.String()))
	assert.True(t, (*byte)(unsafe.StringData(buffer.String())) == unsafe.StringData(copy1.String()))
	assert.True(t, (*byte)(unsafe.StringData(copy1.String())) == unsafe.StringData(copy2.String()))

	assert.True(t, buffer.Update(0, 'g'))
	assert.False(t, buffer.Update(-1, 'g'))
	assert.False(t, buffer.Update(4, 'g'))

	assert.True(t, reflect.DeepEqual([]byte{'g', 'b', 'c', 'd'}, buffer.d.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy1.d.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy2.d.data))

	assert.NotEqual(t, unsafe.SliceData(buffer.d.data), unsafe.SliceData(copy1.d.data))
	assert.Equal(t, unsafe.SliceData(copy1.d.data), unsafe.SliceData(copy2.d.data))

	copy1.Close()

	previous := copy2.d.data
	copy2.Update(0, 'f')
	current := copy2.d.data

	// 1 reference - don't need to copy buffer during update
	assert.Equal(t, unsafe.SliceData(previous), unsafe.SliceData(current))

	copy2.Close()
}
