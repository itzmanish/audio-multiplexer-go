package multiplexer

import (
	"errors"
	"io"
	"sync/atomic"
)

var ErrEmptyBuffer = errors.New("empty buffer")
var ErrShortBuffer = io.ErrShortBuffer

type RingBuffer[T any] struct {
	buffer []T
	head   uint64
	tail   uint64
	size   uint64
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer: make([]T, capacity),
		size:   uint64(capacity),
	}
}

func (rb *RingBuffer[T]) Write(data []T) (int, error) {
	dataLen := len(data)
	if dataLen == 0 {
		return 0, ErrEmptyBuffer
	}

	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	// Calculate how much we can write, including overwriting old data
	writable := dataLen
	if uint64(dataLen) > rb.size {
		writable = int(rb.size)
	}

	for i := 0; i < writable; i++ {
		rb.buffer[(head+uint64(i))%rb.size] = data[i]
	}

	// Update head and possibly tail if we're overwriting
	atomic.StoreUint64(&rb.head, head+uint64(writable))
	if uint64(writable) > rb.size-(head-tail) {
		atomic.StoreUint64(&rb.tail, head+uint64(writable)-rb.size)
	}

	return writable, nil
}

func (rb *RingBuffer[T]) Read(buf []T) (int, error) {
	bufLen := len(buf)
	if bufLen == 0 {
		return 0, ErrShortBuffer
	}

	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	if head == tail {
		return 0, ErrEmptyBuffer // Buffer is empty
	}

	available := head - tail
	toRead := uint64(bufLen)
	if toRead > available {
		toRead = available
	}

	for i := 0; i < int(toRead); i++ {
		buf[i] = rb.buffer[(tail+uint64(i))%rb.size]
	}

	atomic.StoreUint64(&rb.tail, tail+toRead)

	return int(toRead), nil
}
