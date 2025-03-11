// reader.go
package asterix

import (
	"fmt"
	"io"
)

const (
	// DefaultBufferSize is the initial size of the read buffer
	DefaultBufferSize = 16384 // 16KB

	// MaxBufferSize is the maximum allowed size for the buffer
	MaxBufferSize = 1024 * 1024 // 1MB

	// DefaultReadSize is the size of chunks to read from the source
	DefaultReadSize = 4096 // 4KB

	// MaxInvalidMessages is the maximum number of consecutive invalid messages
	// before returning an error (to prevent infinite loops)
	MaxInvalidMessages = 10
)

// Reader reads and decodes ASTERIX messages from an io.Reader
type Reader struct {
	decoder *Decoder
	buffer  []byte
	source  io.Reader

	// Configuration options
	maxBufferSize   int
	readSize        int
	maxInvalidCount int
}

// ReaderOption configures a Reader
type ReaderOption func(*Reader)

// WithMaxBufferSize sets the maximum buffer size
func WithMaxBufferSize(size int) ReaderOption {
	return func(r *Reader) {
		r.maxBufferSize = size
	}
}

// WithReadSize sets the size of chunks to read
func WithReadSize(size int) ReaderOption {
	return func(r *Reader) {
		r.readSize = size
	}
}

// WithMaxInvalidCount sets the maximum number of consecutive invalid messages
func WithMaxInvalidCount(count int) ReaderOption {
	return func(r *Reader) {
		r.maxInvalidCount = count
	}
}

// NewReader creates a new ASTERIX reader with optional configuration
func NewReader(source io.Reader, decoder *Decoder, opts ...ReaderOption) *Reader {
	r := &Reader{
		decoder:         decoder,
		buffer:          make([]byte, 0, DefaultBufferSize),
		source:          source,
		maxBufferSize:   MaxBufferSize,
		readSize:        DefaultReadSize,
		maxInvalidCount: MaxInvalidMessages,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Close releases resources associated with the reader
// It does not close the underlying io.Reader, as that's the caller's responsibility
func (r *Reader) Close() error {
	// Clear the buffer to allow garbage collection
	r.buffer = nil
	return nil
}

// ReadMessage reads and decodes a single ASTERIX message
func (r *Reader) ReadMessage() (*AsterixMessage, error) {
	// Reusable read buffer to avoid allocations on each call
	tempBuf := make([]byte, r.readSize)
	invalidCount := 0

	for {
		// First, check if we already have a complete message
		if len(r.buffer) >= 3 {
			msgLength := uint16(r.buffer[1])<<8 | uint16(r.buffer[2])

			// Validate length
			if msgLength < 3 {
				// Invalid length, skip first byte and try again
				r.buffer = r.buffer[1:]
				invalidCount++

				if invalidCount >= r.maxInvalidCount {
					return nil, fmt.Errorf("%w: too many consecutive invalid messages (>%d)",
						ErrInvalidMessage, r.maxInvalidCount)
				}
				continue
			}

			// Reset invalid count since we found a valid message length
			invalidCount = 0

			// If we have a complete message
			if len(r.buffer) >= int(msgLength) {
				message := r.buffer[:msgLength]

				// Update buffer - slide remaining data to beginning
				r.buffer = r.buffer[msgLength:]

				// Compact buffer if it's mostly empty to save memory
				// This ensures we don't hold onto a large buffer if we don't need it
				if cap(r.buffer) > r.readSize*2 && len(r.buffer) < cap(r.buffer)/4 {
					newBuf := make([]byte, len(r.buffer), r.readSize)
					copy(newBuf, r.buffer)
					r.buffer = newBuf
				}

				// Decode and return
				decodedMsg, err := r.decoder.Decode(message)
				if err != nil {
					return nil, err
				}

				return decodedMsg, nil
			}
		}

		// Check if buffer is getting too large
		if len(r.buffer) >= r.maxBufferSize {
			return nil, fmt.Errorf("%w: buffer size exceeded maximum allowed (%d bytes)",
				ErrInvalidMessage, r.maxBufferSize)
		}

		// Need more data
		n, err := r.source.Read(tempBuf)
		if err != nil {
			if err == io.EOF && len(r.buffer) > 0 {
				// If we have partial data but hit EOF, return what we have
				continue
			}
			return nil, err
		}

		// No data read (shouldn't normally happen with blocking readers)
		if n == 0 {
			continue
		}

		// Append new data to our buffer
		r.buffer = append(r.buffer, tempBuf[:n]...)
	}
}
