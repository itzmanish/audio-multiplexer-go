package multiplexer

import (
	"sync"
)

type FlushableBuffer struct {
	sync.RWMutex
	size     int
	writeBuf *[]int16
}

func NewFlushableBuffer(size int) FlushableBuffer {
	buf := int16BufferPool.Get()
	if buf == nil {
		bufSlice := make([]int16, 0, size)
		buf = &bufSlice
	}
	return FlushableBuffer{
		writeBuf: buf.(*[]int16),
		size:     size,
	}
}

func (buf *FlushableBuffer) Push(data []int16) {
	buf.Lock()
	temp := *buf.writeBuf
	temp = append(temp, data...)
	buf.writeBuf = &temp
	buf.Unlock()
}

func (buf *FlushableBuffer) Flush() []int16 {
	buffer := int16BufferPool.Get()
	if buffer == nil {
		bufferSlice := make([]int16, buf.size)
		buffer = &bufferSlice
	}
	buf.Lock()
	tempBuf := buf.writeBuf
	buf.writeBuf = buffer.(*[]int16)
	buf.Unlock()
	return *tempBuf
}
