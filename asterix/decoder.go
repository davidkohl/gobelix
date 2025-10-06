// encoding/decoder.go
package asterix

import (
	"fmt"
	"io"
	"sync"

	"github.com/davidkohl/gobelix/encoding"
)

// DecoderOption defines a functional option for configuring the Decoder
type DecoderOption func(*Decoder)

// Decoder provides optimized decoding of ASTERIX messages.
//
// Thread Safety:
//   - Decode, DecodeFrom, DecodeAll, DecodeParallel: Safe for concurrent use
//   - RegisterUAP, GetUAP: Safe for concurrent use
//   - StreamDecode: NOT safe for concurrent calls on the same decoder instance
//     (uses internal state via streamMu and streamBuffer). Create separate
//     decoder instances for concurrent stream decoding.
//   - ExtractMessages, ResetStream: Thread-safe
//
// The decoder maintains internal UAP cache and buffer pool which are safe
// for concurrent access.
type Decoder struct {
	// Configuration options
	parallelism int                  // Number of parallel decoding goroutines
	pool        *encoding.BufferPool // Buffer pool for reusing memory
	uapCache    map[Category]UAP     // Cache of UAPs for categories

	// For tracking messages in stream mode
	streamMu     sync.Mutex
	streamBuffer []byte // Buffer for storing partial messages across reads
}

// WithParallelism sets the number of parallel decoding goroutines
func WithDecoderParallelism(n int) DecoderOption {
	return func(d *Decoder) {
		if n > 0 {
			d.parallelism = n
		}
	}
}

// WithDecoderBufferPool sets the buffer pool to use for decoding
func WithDecoderBufferPool(pool *encoding.BufferPool) DecoderOption {
	return func(d *Decoder) {
		if pool != nil {
			d.pool = pool
		}
	}
}

// WithPreloadedUAPs adds UAPs to the decoder's cache
func WithPreloadedUAPs(uaps ...UAP) DecoderOption {
	return func(d *Decoder) {
		for _, uap := range uaps {
			if uap != nil {
				d.uapCache[uap.Category()] = uap
			}
		}
	}
}

// NewDecoder creates a new ASTERIX decoder with the given options
func NewDecoder(opts ...DecoderOption) *Decoder {
	decoder := &Decoder{
		parallelism:  DefaultParallelism,
		pool:         encoding.DefaultBufferPool, // Use the package-level default
		uapCache:     make(map[Category]UAP),
		streamBuffer: make([]byte, 0, 4096), // Initial stream buffer capacity
	}

	// Apply options
	for _, opt := range opts {
		opt(decoder)
	}

	return decoder
}

// RegisterUAP adds a UAP to the decoder's cache
func (d *Decoder) RegisterUAP(uap UAP) {
	if uap == nil {
		return
	}
	d.uapCache[uap.Category()] = uap
}

// GetUAP retrieves a UAP from the cache or returns nil if not found
func (d *Decoder) GetUAP(category Category) UAP {
	return d.uapCache[category]
}

// Decode parses an ASTERIX data block from bytes.
// Returns nil, error for unknown/invalid categories (caller should skip/log).
func (d *Decoder) Decode(data []byte) (*DataBlock, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("data too short for ASTERIX message: %w", ErrInvalidMessage)
	}

	// Extract category from data
	cat := Category(data[0])
	if !cat.IsValid() {
		// Return error for invalid category - caller should skip this message
		return nil, fmt.Errorf("invalid ASTERIX category %d: %w", cat, ErrInvalidCategory)
	}

	// Get the UAP for this category
	uap := d.uapCache[cat]
	if uap == nil {
		// Return error for unknown category - caller should skip this message
		return nil, fmt.Errorf("no UAP registered for category %d: %w", cat, ErrUAPNotDefined)
	}

	// Create a new data block
	dataBlock, err := NewDataBlock(cat, uap)
	if err != nil {
		return nil, fmt.Errorf("creating data block: %w", err)
	}

	// Decode the data
	if err := dataBlock.Decode(data); err != nil {
		return nil, fmt.Errorf("decoding data block: %w", err)
	}

	return dataBlock, nil
}

// DecodeFrom decodes an ASTERIX data block from a reader
func (d *Decoder) DecodeFrom(r io.Reader) (*DataBlock, error) {
	// Read the header (3 bytes - CAT + LEN)
	header := d.pool.GetWithSize(3)
	defer d.pool.Put(header)

	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	// Get the message length first (so we can skip invalid messages)
	length := int(header[1])<<8 | int(header[2])
	if length < 3 {
		return nil, fmt.Errorf("invalid message length %d: %w", length, ErrInvalidLength)
	}

	// Check if the category is valid
	cat := Category(header[0])
	if !cat.IsValid() {
		// Skip the message body to maintain stream synchronization
		skipBuf := d.pool.GetWithSize(length - 3)
		defer d.pool.Put(skipBuf)
		if _, err := io.ReadFull(r, skipBuf); err != nil {
			return nil, fmt.Errorf("skipping invalid category %d message: %w", cat, err)
		}
		return nil, fmt.Errorf("invalid ASTERIX category %d: %w", cat, ErrInvalidCategory)
	}

	// Get the UAP for this category
	uap := d.uapCache[cat]
	if uap == nil {
		// Skip the message body to maintain stream synchronization
		// We already read the 3-byte header, so skip length-3 bytes
		skipBuf := d.pool.GetWithSize(length - 3)
		defer d.pool.Put(skipBuf)
		if _, err := io.ReadFull(r, skipBuf); err != nil {
			return nil, fmt.Errorf("skipping unknown category %d message: %w", cat, err)
		}
		return nil, fmt.Errorf("no UAP registered for category %d: %w", cat, ErrUAPNotDefined)
	}

	// Allocate a buffer for the full message
	data := d.pool.GetWithSize(length)
	defer d.pool.Put(data)

	// Copy the header we already read
	copy(data, header)

	// Read the rest of the message
	if _, err := io.ReadFull(r, data[3:]); err != nil {
		return nil, fmt.Errorf("reading message body: %w", err)
	}

	// Create and decode the data block
	dataBlock, err := NewDataBlock(cat, uap)
	if err != nil {
		return nil, fmt.Errorf("creating data block: %w", err)
	}

	if err := dataBlock.Decode(data); err != nil {
		return nil, fmt.Errorf("decoding data block: %w", err)
	}

	return dataBlock, nil
}

// DecodeAll decodes multiple ASTERIX data blocks from bytes
func (d *Decoder) DecodeAll(data []byte) ([]*DataBlock, error) {
	var results []*DataBlock
	offset := 0

	for offset < len(data) {
		// Ensure we have enough data for at least the header
		if offset+3 > len(data) {
			return results, fmt.Errorf("incomplete header at offset %d: %w", offset, ErrTruncatedMessage)
		}

		// Get the message length
		length := int(data[offset+1])<<8 | int(data[offset+2])
		if length < 3 {
			return results, fmt.Errorf("invalid message length %d at offset %d: %w", length, offset, ErrInvalidLength)
		}

		// Ensure we have enough data for the complete message
		if offset+length > len(data) {
			return results, fmt.Errorf("incomplete message at offset %d (need %d bytes, have %d): %w",
				offset, length, len(data)-offset, ErrTruncatedMessage)
		}

		// Decode this message
		block, err := d.Decode(data[offset : offset+length])
		if err != nil {
			return results, fmt.Errorf("decoding message at offset %d: %w", offset, err)
		}

		results = append(results, block)
		offset += length
	}

	return results, nil
}

// DecodeParallel decodes multiple ASTERIX data blocks in parallel
// Returns partial results even if some messages fail to decode (those will be nil)
// Returns an error if any message fails, but continues decoding others
func (d *Decoder) DecodeParallel(data [][]byte) ([]*DataBlock, error) {
	if d.parallelism <= 1 || len(data) <= 1 {
		// Use sequential decoding for small batches or when parallelism is disabled
		results := make([]*DataBlock, len(data))
		var firstErr error
		for i, msgData := range data {
			block, err := d.Decode(msgData)
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("decoding message %d: %w", i, err)
				}
				results[i] = nil // Leave as nil for failed message
				continue
			}
			results[i] = block
		}
		return results, firstErr
	}

	// Use parallel decoding
	results := make([]*DataBlock, len(data))
	var errMu sync.Mutex
	var firstErr error

	// Create a worker pool
	var wg sync.WaitGroup
	workCh := make(chan int, len(data))

	// Start workers
	workerCount := min(d.parallelism, len(data))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workCh {
				block, err := d.Decode(data[idx])
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("decoding message %d: %w", idx, err)
					}
					errMu.Unlock()
					results[idx] = nil // Leave as nil for failed message
					continue
				}
				results[idx] = block
			}
		}()
	}

	// Send all work to workers
	for i := range data {
		workCh <- i
	}

	close(workCh)
	wg.Wait()

	return results, firstErr
}

// StreamDecode processes a stream of ASTERIX messages and calls the callback for each one
// Note: The stream buffer is NOT automatically reset. Call ResetStream() before reusing the decoder
// for a new stream if you want to ensure a clean state.
func (d *Decoder) StreamDecode(r io.Reader, callback func(*DataBlock) error) error {
	d.streamMu.Lock()
	defer d.streamMu.Unlock()

	// Buffer for reading chunks from the stream
	readBuf := d.pool.GetWithSize(4096) // 4KB read buffer
	defer d.pool.Put(readBuf)

	for {
		// Read a chunk from the stream
		n, err := r.Read(readBuf)
		if err != nil {
			if err == io.EOF {
				return nil // Normal end of stream
			}
			return fmt.Errorf("reading from stream: %w", err)
		}

		// Append new data to the stream buffer
		d.streamBuffer = append(d.streamBuffer, readBuf[:n]...)

		// Process complete messages from the buffer
		offset := 0
		for offset+3 <= len(d.streamBuffer) { // Need at least 3 bytes for header
			// Get the message length
			length := int(d.streamBuffer[offset+1])<<8 | int(d.streamBuffer[offset+2])
			if length < 3 {
				return fmt.Errorf("invalid message length %d in stream: %w", length, ErrInvalidLength)
			}

			// If we don't have the complete message yet, wait for more data
			if offset+length > len(d.streamBuffer) {
				break
			}

			// Decode this message
			block, err := d.Decode(d.streamBuffer[offset : offset+length])
			if err != nil {
				return fmt.Errorf("decoding message in stream: %w", err)
			}

			// Call the callback with the decoded message
			if err := callback(block); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}

			offset += length
		}

		// Keep any partial message for the next iteration
		if offset < len(d.streamBuffer) {
			d.streamBuffer = append(d.streamBuffer[:0], d.streamBuffer[offset:]...)
		} else {
			d.streamBuffer = d.streamBuffer[:0]
		}
	}
}

// ExtractMessages extracts individual ASTERIX messages from a byte slice
// This is useful for protocols where ASTERIX messages may be embedded in another format
func (d *Decoder) ExtractMessages(data []byte) ([][]byte, error) {
	var messages [][]byte
	offset := 0

	for offset+3 <= len(data) { // Need at least 3 bytes for header
		// Check if this looks like the start of an ASTERIX message
		cat := Category(data[offset])
		if !cat.IsValid() {
			// Skip this byte and continue searching
			offset++
			continue
		}

		// Get the potential message length
		length := int(data[offset+1])<<8 | int(data[offset+2])
		if length < 3 || offset+length > len(data) {
			// Invalid length or incomplete message, skip this byte
			offset++
			continue
		}

		// This looks like a valid ASTERIX message
		msgData := make([]byte, length)
		copy(msgData, data[offset:offset+length])
		messages = append(messages, msgData)

		// Move to the next potential message
		offset += length
	}

	return messages, nil
}

// ResetStream clears any buffered data in the stream decoder
func (d *Decoder) ResetStream() {
	d.streamMu.Lock()
	defer d.streamMu.Unlock()
	d.streamBuffer = d.streamBuffer[:0]
}
