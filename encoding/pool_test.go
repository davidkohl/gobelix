// encoding/pool_test.go
package encoding

import (
	"strconv"
	"sync"
	"testing"
)

func TestBufferPoolGet(t *testing.T) {
	pool := NewBufferPool()

	testCases := []struct {
		name     string
		capacity int
		minCap   int
	}{
		{"Small buffer", 32, smallBufferSize},
		{"Medium buffer", 512, mediumBufferSize},
		{"Large buffer", 4096, largeBufferSize},
		{"Extra large buffer", 10000, 10000},
		{"Zero capacity", 0, smallBufferSize},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := pool.Get(tc.capacity)

			// Buffer should be empty
			if len(buf) != 0 {
				t.Errorf("Buffer length = %d, want 0", len(buf))
			}

			// Buffer should have at least the requested capacity
			if cap(buf) < tc.capacity {
				t.Errorf("Buffer capacity = %d, want at least %d", cap(buf), tc.capacity)
			}

			// For standard sizes, buffer should have exact pool size
			if tc.capacity <= largeBufferSize && cap(buf) != tc.minCap {
				t.Errorf("Buffer capacity = %d, want %d", cap(buf), tc.minCap)
			}

			// Return to pool
			pool.Put(buf)
		})
	}
}

func TestBufferPoolPut(t *testing.T) {
	pool := NewBufferPool()

	// Put should handle nil buffers without panicking
	t.Run("Put nil buffer", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Put(nil) panicked: %v", r)
			}
		}()
		pool.Put(nil)
	})

	// Put and get should work together
	t.Run("Put and get", func(t *testing.T) {
		buf1 := pool.Get(smallBufferSize)
		copy(buf1, []byte("test data"))
		buf1 = buf1[:9] // resize to length of data
		pool.Put(buf1)

		// Get the buffer back
		buf2 := pool.Get(smallBufferSize)

		// Buffer should be empty due to reset
		if len(buf2) != 0 {
			t.Errorf("Buffer length = %d, want 0", len(buf2))
		}

		// Capacity should remain the same
		if cap(buf2) != smallBufferSize {
			t.Errorf("Buffer capacity = %d, want %d", cap(buf2), smallBufferSize)
		}
	})
}

func TestBufferGetWithSize(t *testing.T) {
	pool := NewBufferPool()

	testCases := []struct {
		name string
		size int
	}{
		{"Small size", 32},
		{"Medium size", 512},
		{"Large size", 4096},
		{"Zero size", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := pool.GetWithSize(tc.size)

			// Buffer should have the requested length
			if len(buf) != tc.size {
				t.Errorf("Buffer length = %d, want %d", len(buf), tc.size)
			}

			// Buffer should have at least the requested capacity
			if cap(buf) < tc.size {
				t.Errorf("Buffer capacity = %d, want at least %d", cap(buf), tc.size)
			}

			// Return to pool
			pool.Put(buf)
		})
	}
}

func TestBufferGetExact(t *testing.T) {
	pool := NewBufferPool()

	testCases := []struct {
		name string
		size int
	}{
		{"Small exact", 33},
		{"Medium exact", 555},
		{"Large exact", 4321},
		{"Zero exact", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := pool.GetExact(tc.size)

			// Buffer should have exactly the requested length
			if len(buf) != tc.size {
				t.Errorf("Buffer length = %d, want %d", len(buf), tc.size)
			}

			// Buffer should have exactly the requested capacity
			if cap(buf) != tc.size {
				t.Errorf("Buffer capacity = %d, want exactly %d", cap(buf), tc.size)
			}

			// These buffers are not returned to the pool
		})
	}
}

func TestDefaultBufferPool(t *testing.T) {
	// Test the package-level functions
	buf := GetBuffer(100)
	if cap(buf) < 100 {
		t.Errorf("GetBuffer capacity = %d, want at least 100", cap(buf))
	}
	PutBuffer(buf)

	buf = GetBufferWithSize(50)
	if len(buf) != 50 {
		t.Errorf("GetBufferWithSize length = %d, want 50", len(buf))
	}
	PutBuffer(buf)
}

func TestConcurrentUse(t *testing.T) {
	pool := NewBufferPool()
	var wg sync.WaitGroup

	// Number of goroutines to spin up
	workers := 10
	// Operations per goroutine
	opsPerWorker := 100

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				// Get different sized buffers
				size := (id * j) % 2000

				// Get a buffer
				buf := pool.Get(size)

				// Do some work with the buffer
				if cap(buf) < size {
					t.Errorf("Worker %d: Buffer capacity too small: %d < %d", id, cap(buf), size)
				}

				// Resize the buffer
				buf = buf[:size]

				// Clear it out and fill it
				for k := range buf {
					buf[k] = byte(k % 256)
				}

				// Return to pool
				pool.Put(buf)
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkBufferPoolGet(b *testing.B) {
	pool := NewBufferPool()
	sizes := []int{32, 512, 4096}

	for _, size := range sizes {
		b.Run(string(strconv.Itoa(size)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf := pool.Get(size)
				pool.Put(buf)
			}
		})
	}
}

func BenchmarkBufferAllocation(b *testing.B) {
	sizes := []int{32, 512, 4096}

	for _, size := range sizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]byte, 0, size)
				// Note: no cleanup as we're measuring allocation
			}
		})
	}
}

func BenchmarkBufferPoolVsDirectAllocation(b *testing.B) {
	pool := NewBufferPool()
	sizes := []int{32, 512, 4096}

	for _, size := range sizes {
		b.Run("Pool-"+strconv.Itoa(size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf := pool.Get(size)
				buf = buf[:size] // Resize to simulate use
				pool.Put(buf)
			}
		})

		b.Run("Direct-"+strconv.Itoa(size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]byte, size)
				// No cleanup as we're comparing to pooling
			}
		})
	}
}
