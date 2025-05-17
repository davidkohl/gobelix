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

// Decoder provides optimized decoding of ASTERIX messages
type Decoder struct {
	// Configuration options
	parallelism int                  // Number of parallel decoding ∏
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

// Decode parses an ASTERIX data block from bytes
func (d *Decoder) Decode(data []byte) (*DataBlock, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("data too short for ASTERIX message: %w", ErrInvalidMessage)
	}

	// Extract category from data
	cat := Category(data[0])
	if !cat.IsValid() {
		return nil, fmt.Errorf("invalid ASTERIX category %d: %w", cat, ErrInvalidCategory)
	}

	// Get the UAP for this category
	uap := d.uapCache[cat]
	if uap == nil {
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

	// Check if the category is valid
	cat := Category(header[0])
	if !cat.IsValid() {
		return nil, fmt.Errorf("invalid ASTERIX category %d: %w", cat, ErrInvalidCategory)
	}

	// Get the UAP for this category
	uap := d.uapCache[cat]
	if uap == nil {
		return nil, fmt.Errorf("no UAP registered for category %d: %w", cat, ErrUAPNotDefined)
	}

	// Get the message length
	length := int(header[1])<<8 | int(header[2])
	if length < 3 {
		return nil, fmt.Errorf("invalid message length %d: %w", length, ErrInvalidLength)
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
func (d *Decoder) DecodeParallel(data [][]byte) ([]*DataBlock, error) {
	if d.parallelism <= 1 || len(data) <= 1 {
		// Use sequential decoding for small batches or when parallelism is disabled
		results := make([]*DataBlock, len(data))
		for i, msgData := range data {
			block, err := d.Decode(msgData)
			if err != nil {
				return results, fmt.Errorf("decoding message %d: %w", i, err)
			}
			results[i] = block
		}
		return results, nil
	}

	// Use parallel decoding
	results := make([]*DataBlock, len(data))
	errs := make(chan error, 1)

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
					select {
					case errs <- fmt.Errorf("decoding message %d: %w", idx, err):
					default:
						// Channel already has an error, don't block
					}
					return
				}
				results[idx] = block
			}
		}()
	}

	// Send work to workers
	for i := range data {
		select {
		case err := <-errs:
			// One of the workers encountered an error
			close(workCh) // Stop sending more work
			wg.Wait()     // Wait for all workers to finish
			return results, err
		case workCh <- i:
			// Work sent successfully
		}
	}

	close(workCh)
	wg.Wait()
	close(errs)

	// Check for errors
	select {
	case err := <-errs:
		return results, err
	default:
		// No errors
	}

	return results, nil
}

// StreamDecode processes a stream of ASTERIX messages and calls the callback for each one
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
