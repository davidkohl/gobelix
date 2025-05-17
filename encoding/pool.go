// encoding/pool.go
package encoding

import (
	"sync"
)

// BufferPool provides reusable byte slices to reduce memory allocations.
// Different pools are used for different sizes to avoid wasting memory.
type BufferPool struct {
	small  sync.Pool // For buffers up to 64 bytes
	medium sync.Pool // For buffers up to 1024 bytes
	large  sync.Pool // For buffers up to 8192 bytes
}

// Default pool sizes
const (
	smallBufferSize  = 64
	mediumBufferSize = 1024
	largeBufferSize  = 8192
)

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		small: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, smallBufferSize)
				return &buf
			},
		},
		medium: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, mediumBufferSize)
				return &buf
			},
		},
		large: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, largeBufferSize)
				return &buf
			},
		},
	}
}

// Get retrieves a buffer with at least the specified capacity
func (p *BufferPool) Get(capacity int) []byte {
	var buf *[]byte

	switch {
	case capacity <= smallBufferSize:
		buf = p.small.Get().(*[]byte)
		if cap(*buf) < capacity {
			*buf = make([]byte, 0, smallBufferSize)
		}
	case capacity <= mediumBufferSize:
		buf = p.medium.Get().(*[]byte)
		if cap(*buf) < capacity {
			*buf = make([]byte, 0, mediumBufferSize)
		}
	case capacity <= largeBufferSize:
		buf = p.large.Get().(*[]byte)
		if cap(*buf) < capacity {
			*buf = make([]byte, 0, largeBufferSize)
		}
	default:
		// For very large buffers, don't use the pool
		slice := make([]byte, 0, capacity)
		return slice
	}

	// Reset length but keep capacity
	*buf = (*buf)[:0]
	return *buf
}

// Put returns a buffer to the pool
func (p *BufferPool) Put(buf []byte) {
	if buf == nil {
		return
	}

	// Return to the appropriate pool based on capacity
	switch cap(buf) {
	case 0:
		// Don't store empty buffers
		return
	case smallBufferSize:
		p.small.Put(&buf)
	case mediumBufferSize:
		p.medium.Put(&buf)
	case largeBufferSize:
		p.large.Put(&buf)
	default:
		// Don't keep non-standard sized buffers
		// They'll be garbage collected
	}
}

// GetWithSize retrieves a buffer with the specified capacity and pre-sets its length
func (p *BufferPool) GetWithSize(size int) []byte {
	buf := p.Get(size)
	// Grow the buffer to the requested size
	if cap(buf) < size {
		buf = make([]byte, size)
	} else {
		buf = buf[:size]
	}
	return buf
}

// GetExact retrieves a buffer with exactly the specified size and capacity
// This is useful when you need a buffer of a specific size that won't be resized
func (p *BufferPool) GetExact(size int) []byte {
	buf := make([]byte, size)
	return buf
}

// DefaultBufferPool is the package-level buffer pool that can be used
// by all components in the package.
var DefaultBufferPool = NewBufferPool()

// GetBuffer retrieves a buffer from the default pool with at least the specified capacity
func GetBuffer(capacity int) []byte {
	return DefaultBufferPool.Get(capacity)
}

// PutBuffer returns a buffer to the default pool
func PutBuffer(buf []byte) {
	DefaultBufferPool.Put(buf)
}

// GetBufferWithSize retrieves a buffer from the default pool with the specified size
func GetBufferWithSize(size int) []byte {
	return DefaultBufferPool.GetWithSize(size)
}
