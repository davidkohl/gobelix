// encoding/encoder.go
package asterix

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/davidkohl/gobelix/encoding"
)

// EncoderOption defines a functional option for configuring the Encoder
type EncoderOption func(*Encoder)

// Encoder provides optimized encoding of ASTERIX messages.
//
// Thread Safety:
//   - Encode, EncodeTo, EncodeItems, EncodeItemsTo: Safe for concurrent use
//   - EncodeParallel, EncodeStream: Safe for concurrent use
//   - Batch operations (StartBatch, AddToBatch, FinishBatch, FinishBatchTo):
//     NOT safe for concurrent use. The encoder maintains batch state that is
//     protected by batchMu, but batch operations should be serialized per
//     encoder instance. For concurrent batch encoding, create multiple encoder
//     instances.
//
// The encoder maintains internal buffer pool which is safe for concurrent access.
type Encoder struct {
	// Configuration options
	maxBatchSize int                  // Maximum batch size in bytes
	parallelism  int                  // Number of parallel encoding goroutines
	pool         *encoding.BufferPool // Buffer pool for reusing memory
	uapCache     map[Category]UAP     // Cache of UAPs for categories

	// Internal state
	batchMu       sync.Mutex // Mutex for batch operations
	batchCategory Category   // Category of the current batch
	batchUAP      UAP        // UAP of the current batch
	batchRecords  []*Record  // Records in the current batch
}

// DefaultMaxBatchSize is the default maximum batch size in bytes
const DefaultMaxBatchSize = 4096

// DefaultParallelism is the default number of parallel encoding goroutines
const DefaultParallelism = 1 // Default to single-threaded

// WithMaxBatchSize sets the maximum batch size in bytes
func WithMaxBatchSize(size int) EncoderOption {
	return func(e *Encoder) {
		if size > 0 {
			e.maxBatchSize = size
		}
	}
}

// WithParallelism sets the number of parallel encoding goroutines
func WithParallelism(n int) EncoderOption {
	return func(e *Encoder) {
		if n > 0 {
			e.parallelism = n
		}
	}
}

// WithBufferPool sets the buffer pool to use for encoding
func WithBufferPool(pool *encoding.BufferPool) EncoderOption {
	return func(e *Encoder) {
		if pool != nil {
			e.pool = pool
		}
	}
}

// NewEncoder creates a new ASTERIX encoder with the given options
func NewEncoder(opts ...EncoderOption) *Encoder {
	encoder := &Encoder{
		maxBatchSize: DefaultMaxBatchSize,
		parallelism:  DefaultParallelism,
		pool:         encoding.DefaultBufferPool, // Use the package-level default
		uapCache:     make(map[Category]UAP),
	}

	// Apply options
	for _, opt := range opts {
		opt(encoder)
	}

	return encoder
}

// Encode encodes a DataBlock to bytes
// The returned byte slice is a fresh allocation and safe to use/modify.
func (e *Encoder) Encode(dataBlock *DataBlock) ([]byte, error) {
	// Get buffer from pool
	estimatedSize := dataBlock.EstimateSize()
	poolBuf := e.pool.Get(estimatedSize)
	defer e.pool.Put(poolBuf)

	// Create a bytes.Buffer wrapper around the pooled slice
	buf := bytes.NewBuffer(poolBuf[:0])

	// Encode to buffer
	_, err := dataBlock.EncodeWithBuffer(buf)
	if err != nil {
		return nil, fmt.Errorf("encoding data block: %w", err)
	}

	// Copy to a fresh allocation for the caller
	// (pool buffer will be returned and reused)
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
}

// EncodeTo encodes a DataBlock to an io.Writer
func (e *Encoder) EncodeTo(dataBlock *DataBlock, w io.Writer) (int, error) {
	// Estimate size and get buffer from pool
	estimatedSize := dataBlock.EstimateSize()
	buf := e.pool.Get(estimatedSize)
	defer e.pool.Put(buf)

	// Encode to buffer
	bufWriter := bytes.NewBuffer(buf[:0])
	_, err := dataBlock.EncodeWithBuffer(bufWriter)
	if err != nil {
		return 0, fmt.Errorf("encoding data block: %w", err)
	}

	// Write to the provided writer
	n, err := w.Write(bufWriter.Bytes())
	if err != nil {
		return n, fmt.Errorf("writing to output: %w", err)
	}

	return n, nil
}

// EncodeItems encodes a map of items directly to bytes
func (e *Encoder) EncodeItems(cat Category, uap UAP, items map[string]DataItem) ([]byte, error) {
	// Create a data block
	dataBlock, err := NewDataBlock(cat, uap)
	if err != nil {
		return nil, fmt.Errorf("creating data block: %w", err)
	}

	// Add items as a record
	if err := dataBlock.EncodeRecord(items); err != nil {
		return nil, fmt.Errorf("encoding record: %w", err)
	}

	// Encode the data block
	return e.Encode(dataBlock)
}

// EncodeItemsTo encodes a map of items directly to an io.Writer
func (e *Encoder) EncodeItemsTo(cat Category, uap UAP, items map[string]DataItem, w io.Writer) (int, error) {
	// Create a data block
	dataBlock, err := NewDataBlock(cat, uap)
	if err != nil {
		return 0, fmt.Errorf("creating data block: %w", err)
	}

	// Add items as a record
	if err := dataBlock.EncodeRecord(items); err != nil {
		return 0, fmt.Errorf("encoding record: %w", err)
	}

	// Encode the data block
	return e.EncodeTo(dataBlock, w)
}

// StartBatch begins a new batch encoding operation
func (e *Encoder) StartBatch(cat Category, uap UAP) {
	e.batchMu.Lock()
	defer e.batchMu.Unlock()

	// Clear any existing batch
	e.batchCategory = cat
	e.batchUAP = uap
	e.batchRecords = nil
}

// AddToBatch adds a map of items to the current batch
func (e *Encoder) AddToBatch(items map[string]DataItem) error {
	e.batchMu.Lock()
	defer e.batchMu.Unlock()

	if e.batchUAP == nil {
		return fmt.Errorf("no batch in progress, call StartBatch first")
	}

	// Create a record
	record, err := NewRecord(e.batchCategory, e.batchUAP)
	if err != nil {
		return fmt.Errorf("creating record: %w", err)
	}

	// Add items to the record
	for id, item := range items {
		if err := record.SetDataItem(id, item); err != nil {
			return fmt.Errorf("setting data item %s: %w", id, err)
		}
	}

	// Add the record to the batch
	e.batchRecords = append(e.batchRecords, record)

	return nil
}

// FinishBatch completes a batch encoding operation
func (e *Encoder) FinishBatch() ([]byte, error) {
	e.batchMu.Lock()
	defer e.batchMu.Unlock()

	if e.batchUAP == nil {
		return nil, fmt.Errorf("no batch in progress, call StartBatch first")
	}

	if len(e.batchRecords) == 0 {
		return nil, fmt.Errorf("batch contains no records")
	}

	// Create a data block
	dataBlock, err := NewDataBlock(e.batchCategory, e.batchUAP)
	if err != nil {
		return nil, fmt.Errorf("creating data block: %w", err)
	}

	// Add records to the data block
	for _, record := range e.batchRecords {
		if err := dataBlock.AddRecord(record); err != nil {
			return nil, fmt.Errorf("adding record to batch: %w", err)
		}
	}

	// Encode the data block
	data, err := e.Encode(dataBlock)
	if err != nil {
		return nil, fmt.Errorf("encoding batch: %w", err)
	}

	// Clear the batch
	e.batchCategory = 0
	e.batchUAP = nil
	e.batchRecords = nil

	return data, nil
}

// FinishBatchTo completes a batch encoding operation and writes to an io.Writer
func (e *Encoder) FinishBatchTo(w io.Writer) (int, error) {
	e.batchMu.Lock()
	defer e.batchMu.Unlock()

	if e.batchUAP == nil {
		return 0, fmt.Errorf("no batch in progress, call StartBatch first")
	}

	if len(e.batchRecords) == 0 {
		return 0, fmt.Errorf("batch contains no records")
	}

	// Create a data block
	dataBlock, err := NewDataBlock(e.batchCategory, e.batchUAP)
	if err != nil {
		return 0, fmt.Errorf("creating data block: %w", err)
	}

	// Add records to the data block
	for _, record := range e.batchRecords {
		if err := dataBlock.AddRecord(record); err != nil {
			return 0, fmt.Errorf("adding record to batch: %w", err)
		}
	}

	// Encode the data block
	n, err := e.EncodeTo(dataBlock, w)
	if err != nil {
		return n, fmt.Errorf("encoding batch: %w", err)
	}

	// Clear the batch
	e.batchCategory = 0
	e.batchUAP = nil
	e.batchRecords = nil

	return n, nil
}

// EncodeParallel encodes multiple data blocks in parallel
func (e *Encoder) EncodeParallel(dataBlocks []*DataBlock) ([][]byte, error) {
	if e.parallelism <= 1 || len(dataBlocks) <= 1 {
		// Use sequential encoding for small batches or when parallelism is disabled
		results := make([][]byte, len(dataBlocks))
		for i, block := range dataBlocks {
			data, err := e.Encode(block)
			if err != nil {
				return nil, fmt.Errorf("encoding block %d: %w", i, err)
			}
			results[i] = data
		}
		return results, nil
	}

	// Use parallel encoding
	results := make([][]byte, len(dataBlocks))
	errs := make(chan error, len(dataBlocks))

	// Create a worker pool
	var wg sync.WaitGroup
	workCh := make(chan int, len(dataBlocks))

	// Start workers
	workerCount := min(e.parallelism, len(dataBlocks))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workCh {
				data, err := e.Encode(dataBlocks[idx])
				if err != nil {
					errs <- fmt.Errorf("encoding block %d: %w", idx, err)
					return
				}
				results[idx] = data
			}
		}()
	}

	// Send work to workers
	for i := range dataBlocks {
		workCh <- i
	}
	close(workCh)

	// Wait for all workers to finish
	wg.Wait()
	close(errs)

	// Check for errors
	if err := <-errs; err != nil {
		return nil, err
	}

	return results, nil
}

// EncodeStream encodes items from a channel and writes to an io.Writer
func (e *Encoder) EncodeStream(cat Category, uap UAP,
	itemsCh <-chan map[string]DataItem, w io.Writer) error {

	// Create a data block
	dataBlock, err := NewDataBlock(cat, uap)
	if err != nil {
		return fmt.Errorf("creating data block: %w", err)
	}

	// Set block to allow multiple records if supported
	dataBlock.SetBlockable(cat.IsBlockable())

	// Process items from the channel
	currentSize := 0
	for items := range itemsCh {
		// Create a new record for these items
		record, err := NewRecord(cat, uap)
		if err != nil {
			return fmt.Errorf("creating record: %w", err)
		}

		// Add items to the record
		for id, item := range items {
			if err := record.SetDataItem(id, item); err != nil {
				return fmt.Errorf("setting data item %s: %w", id, err)
			}
		}

		// Add the record to the data block
		if err := dataBlock.AddRecord(record); err != nil {
			return fmt.Errorf("adding record to data block: %w", err)
		}

		// Update current size
		currentSize += record.EstimateSize()

		// Flush if batch is full
		if currentSize >= e.maxBatchSize {
			if _, err := e.EncodeTo(dataBlock, w); err != nil {
				return fmt.Errorf("encoding data block: %w", err)
			}

			// Reset for next batch
			dataBlock.Clear()
			currentSize = 0
		}
	}

	// Flush any remaining records
	if dataBlock.RecordCount() > 0 {
		if _, err := e.EncodeTo(dataBlock, w); err != nil {
			return fmt.Errorf("encoding final data block: %w", err)
		}
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
