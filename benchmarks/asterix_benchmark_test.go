// benchmarks/asterix_benchmark_test.go
package benchmarks

import (
	"bytes"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021"
	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
	"github.com/davidkohl/gobelix/encoding"
)

// createTestMessage creates a simple ASTERIX message for testing
func createTestMessage(t *testing.B) []byte {
	// Get the Category 021 UAP
	uap, err := cat021.NewUAP(cat021.Version26)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Create a data block for encoding
	dataBlock, err := asterix.NewDataBlock(asterix.Cat021, uap)
	if err != nil {
		t.Fatalf("Failed to create data block: %v", err)
	}

	// Create a record
	record, err := asterix.NewRecord(asterix.Cat021, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Add required data items
	dsi := &common.DataSourceIdentifier{
		SAC: 25,
		SIC: 100,
	}
	if err := record.SetDataItem("I021/010", dsi); err != nil {
		t.Fatalf("Failed to set DSI: %v", err)
	}

	trd := &v26.TargetReportDescriptor{
		ATP: 1,
		ARC: 0,
	}
	if err := record.SetDataItem("I021/040", trd); err != nil {
		t.Fatalf("Failed to set TRD: %v", err)
	}

	ta := &v26.TargetAddress{
		Address: 0xABCDEF,
	}
	if err := record.SetDataItem("I021/080", ta); err != nil {
		t.Fatalf("Failed to set TA: %v", err)
	}

	// Add the record to the data block
	if err := dataBlock.AddRecord(record); err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Encode the data block
	data, err := dataBlock.Encode()
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	return data
}

// createMultipleTestMessages creates a batch of test messages
func createMultipleTestMessages(t *testing.B, count int) [][]byte {
	messages := make([][]byte, count)

	// Get the Category 021 UAP
	uap, err := cat021.NewUAP(cat021.Version26)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	for i := 0; i < count; i++ {
		// Create a data block for encoding
		dataBlock, err := asterix.NewDataBlock(asterix.Cat021, uap)
		if err != nil {
			t.Fatalf("Failed to create data block: %v", err)
		}

		// Create a record
		record, err := asterix.NewRecord(asterix.Cat021, uap)
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}

		// Add required data items with slight variations
		dsi := &common.DataSourceIdentifier{
			SAC: 25,
			SIC: uint8(100 + i%50),
		}
		if err := record.SetDataItem("I021/010", dsi); err != nil {
			t.Fatalf("Failed to set DSI: %v", err)
		}

		trd := &v26.TargetReportDescriptor{
			ATP: 1,
			ARC: 0,
		}
		if err := record.SetDataItem("I021/040", trd); err != nil {
			t.Fatalf("Failed to set TRD: %v", err)
		}

		ta := &v26.TargetAddress{
			Address: uint32(0xAA0000 + i),
		}
		if err := record.SetDataItem("I021/080", ta); err != nil {
			t.Fatalf("Failed to set TA: %v", err)
		}

		// Add the record to the data block
		if err := dataBlock.AddRecord(record); err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}

		// Encode the data block
		data, err := dataBlock.Encode()
		if err != nil {
			t.Fatalf("Failed to encode: %v", err)
		}

		messages[i] = data
	}

	return messages
}

// BenchmarkDirectDecode tests decoding using the direct method
func BenchmarkDirectDecode(b *testing.B) {
	data := createTestMessage(b)

	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dataBlock, _ := asterix.NewDataBlock(asterix.Cat021, uap)
		dataBlock.Decode(data)
	}
}

// BenchmarkDecoderDecode tests decoding using our decoder implementation
func BenchmarkDecoderDecode(b *testing.B) {
	data := createTestMessage(b)

	// Create decoder and register UAP
	uap, _ := cat021.NewUAP(cat021.Version26)
	decoder := encoding.NewDecoder(encoding.WithPreloadedUAPs(uap))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder.Decode(data)
	}
}

// BenchmarkDecoderWithPoolDecode tests decoding using our decoder with a custom buffer pool
func BenchmarkDecoderWithPoolDecode(b *testing.B) {
	data := createTestMessage(b)

	// Create a custom pool
	pool := encoding.NewBufferPool()

	// Create decoder with pool and register UAP
	uap, _ := cat021.NewUAP(cat021.Version26)
	decoder := encoding.NewDecoder(
		encoding.WithPreloadedUAPs(uap),
		encoding.WithDecoderBufferPool(pool),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder.Decode(data)
	}
}

// BenchmarkDirectEncode tests encoding using the direct method
func BenchmarkDirectEncode(b *testing.B) {
	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a data block
		dataBlock, _ := asterix.NewDataBlock(asterix.Cat021, uap)

		// Create a record
		record, _ := asterix.NewRecord(asterix.Cat021, uap)

		// Add required data items
		dsi := &common.DataSourceIdentifier{
			SAC: 25,
			SIC: 100,
		}
		record.SetDataItem("I021/010", dsi)

		trd := &v26.TargetReportDescriptor{
			ATP: 1,
			ARC: 0,
		}
		record.SetDataItem("I021/040", trd)

		ta := &v26.TargetAddress{
			Address: 0xABCDEF,
		}
		record.SetDataItem("I021/080", ta)

		// Add the record to the data block
		dataBlock.AddRecord(record)

		// Encode the data block
		dataBlock.Encode()
	}
}

// BenchmarkEncoderEncode tests encoding using our encoder implementation
func BenchmarkEncoderEncode(b *testing.B) {
	// Create encoder
	encoder := encoding.NewEncoder()

	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a data block
		dataBlock, _ := asterix.NewDataBlock(asterix.Cat021, uap)

		// Create a record
		record, _ := asterix.NewRecord(asterix.Cat021, uap)

		// Add required data items
		dsi := &common.DataSourceIdentifier{
			SAC: 25,
			SIC: 100,
		}
		record.SetDataItem("I021/010", dsi)

		trd := &v26.TargetReportDescriptor{
			ATP: 1,
			ARC: 0,
		}
		record.SetDataItem("I021/040", trd)

		ta := &v26.TargetAddress{
			Address: 0xABCDEF,
		}
		record.SetDataItem("I021/080", ta)

		// Add the record to the data block
		dataBlock.AddRecord(record)

		// Encode the data block
		encoder.Encode(dataBlock)
	}
}

// BenchmarkEncoderWithPoolEncode tests encoding using our encoder with a custom buffer pool
func BenchmarkEncoderWithPoolEncode(b *testing.B) {
	// Create a custom pool
	pool := encoding.NewBufferPool()

	// Create encoder with pool
	encoder := encoding.NewEncoder(encoding.WithBufferPool(pool))

	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a data block
		dataBlock, _ := asterix.NewDataBlock(asterix.Cat021, uap)

		// Create a record
		record, _ := asterix.NewRecord(asterix.Cat021, uap)

		// Add required data items
		dsi := &common.DataSourceIdentifier{
			SAC: 25,
			SIC: 100,
		}
		record.SetDataItem("I021/010", dsi)

		trd := &v26.TargetReportDescriptor{
			ATP: 1,
			ARC: 0,
		}
		record.SetDataItem("I021/040", trd)

		ta := &v26.TargetAddress{
			Address: 0xABCDEF,
		}
		record.SetDataItem("I021/080", ta)

		// Add the record to the data block
		dataBlock.AddRecord(record)

		// Encode the data block
		encoder.Encode(dataBlock)
	}
}

// BenchmarkEncoderBatchEncode tests batch encoding using our encoder implementation
func BenchmarkEncoderBatchEncode(b *testing.B) {
	// Create encoder
	encoder := encoding.NewEncoder()

	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	// Create items for 10 records
	items := make([]map[string]asterix.DataItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = map[string]asterix.DataItem{
			"I021/010": &common.DataSourceIdentifier{
				SAC: 25,
				SIC: uint8(100 + i),
			},
			"I021/040": &v26.TargetReportDescriptor{
				ATP: 1,
				ARC: 0,
			},
			"I021/080": &v26.TargetAddress{
				Address: uint32(0xAA0000 + i),
			},
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Start a new batch
		encoder.StartBatch(asterix.Cat021, uap)

		// Add items to batch
		for _, itemMap := range items {
			encoder.AddToBatch(itemMap)
		}

		// Finish batch
		encoder.FinishBatch()
	}
}

// BenchmarkParallelDecode tests decoding multiple messages in parallel
func BenchmarkParallelDecode(b *testing.B) {
	// Create 50 test messages
	messages := createMultipleTestMessages(b, 50)

	// Create decoder with 4 parallel workers
	uap, _ := cat021.NewUAP(cat021.Version26)
	decoder := encoding.NewDecoder(
		encoding.WithPreloadedUAPs(uap),
		encoding.WithDecoderParallelism(4),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder.DecodeParallel(messages)
	}
}

// BenchmarkStreamDecode tests decoding a stream of messages
func BenchmarkStreamDecode(b *testing.B) {
	// Create 100 test messages and concatenate them
	messages := createMultipleTestMessages(b, 100)

	// Concatenate messages into a stream
	var streamData bytes.Buffer
	for _, msg := range messages {
		streamData.Write(msg)
	}

	// Create decoder
	uap, _ := cat021.NewUAP(cat021.Version26)
	decoder := encoding.NewDecoder(encoding.WithPreloadedUAPs(uap))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset stream position
		stream := bytes.NewReader(streamData.Bytes())

		// Process stream
		decoder.StreamDecode(stream, func(block *asterix.DataBlock) error {
			return nil
		})
	}
}

// BenchmarkEncoderWithPoolReuse tests encoding with pool reuse
func BenchmarkEncoderWithPoolReuse(b *testing.B) {
	// Create a custom pool
	pool := encoding.NewBufferPool()

	// Create encoder with pool
	encoder := encoding.NewEncoder(encoding.WithBufferPool(pool))

	// Get the UAP
	uap, _ := cat021.NewUAP(cat021.Version26)

	// Create a data block
	dataBlock, _ := asterix.NewDataBlock(asterix.Cat021, uap)

	// Create a record
	record, _ := asterix.NewRecord(asterix.Cat021, uap)

	// Add required data items
	dsi := &common.DataSourceIdentifier{
		SAC: 25,
		SIC: 100,
	}
	record.SetDataItem("I021/010", dsi)

	trd := &v26.TargetReportDescriptor{
		ATP: 1,
		ARC: 0,
	}
	record.SetDataItem("I021/040", trd)

	ta := &v26.TargetAddress{
		Address: 0xABCDEF,
	}
	record.SetDataItem("I021/080", ta)

	// Add the record to the data block
	dataBlock.AddRecord(record)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Encode the data block - same data block used repeatedly
		encoder.Encode(dataBlock)
	}
}

// BenchmarkMemoryPressure simulates high memory pressure conditions
func BenchmarkMemoryPressure(b *testing.B) {
	// Create large number of messages (50)
	messages := createMultipleTestMessages(b, 50)

	// Create decoders
	uap, _ := cat021.NewUAP(cat021.Version26)
	standardDecoder := encoding.NewDecoder(encoding.WithPreloadedUAPs(uap))
	pooledDecoder := encoding.NewDecoder(
		encoding.WithPreloadedUAPs(uap),
		encoding.WithDecoderBufferPool(encoding.NewBufferPool()),
	)

	b.Run("StandardDecoder", func(b *testing.B) {
		// Allocate memory to create pressure
		garbage := make([][]byte, 0, b.N*10)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Create memory pressure by allocating random buffers
			if i%10 == 0 {
				garbage = append(garbage, make([]byte, 1024*1024))
			}

			// Decode all messages in a loop
			for _, msg := range messages {
				standardDecoder.Decode(msg)
			}
		}
	})

	b.Run("PooledDecoder", func(b *testing.B) {
		// Allocate memory to create pressure
		garbage := make([][]byte, 0, b.N*10)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Create memory pressure by allocating random buffers
			if i%10 == 0 {
				garbage = append(garbage, make([]byte, 1024*1024))
			}

			// Decode all messages in a loop
			for _, msg := range messages {
				pooledDecoder.Decode(msg)
			}
		}
	})
}
