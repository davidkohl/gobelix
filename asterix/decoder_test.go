// encoding/decoder_test.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"

	"github.com/davidkohl/gobelix/encoding"
)

func setupTestDecoder() (*Decoder, *MockUAP) {
	// Create a mock UAP for testing
	mockUAP := &MockUAP{
		category: Cat021,
		version:  "1.0",
		fields: []DataField{
			{
				FRN:         1,
				DataItem:    "I021/010",
				Description: "Data Source Identifier",
				Type:        Fixed,
				Length:      2,
				Mandatory:   true,
			},
			{
				FRN:         2,
				DataItem:    "I021/040",
				Description: "Target Report Descriptor",
				Type:        Fixed,
				Length:      1,
				Mandatory:   true,
			},
			{
				FRN:         3,
				DataItem:    "I021/030",
				Description: "Time of Day",
				Type:        Fixed,
				Length:      3,
				Mandatory:   false,
			},
		},
	}

	decoder := NewDecoder(WithPreloadedUAPs(mockUAP))
	return decoder, mockUAP
}

// createTestMessage creates a simple ASTERIX message for testing
func createTestMessage(category Category, itemData map[string][]byte) []byte {
	// Create a buffer with initial capacity
	buf := bytes.NewBuffer(make([]byte, 0, 64))

	// Write category
	buf.WriteByte(byte(category))

	// Reserve space for length
	buf.Write([]byte{0, 0})

	// Write FSPEC
	fspec := byte(0)
	for id := range itemData {
		switch id {
		case "I021/010":
			fspec |= 0x80 // Bit 8 (FRN 1)
		case "I021/040":
			fspec |= 0x40 // Bit 7 (FRN 2)
		case "I021/030":
			fspec |= 0x20 // Bit 6 (FRN 3)
		}
	}
	buf.WriteByte(fspec)

	// Write data items in FRN order
	if data, ok := itemData["I021/010"]; ok {
		buf.Write(data)
	}
	if data, ok := itemData["I021/040"]; ok {
		buf.Write(data)
	}
	if data, ok := itemData["I021/030"]; ok {
		buf.Write(data)
	}

	// Update length
	binary.BigEndian.PutUint16(buf.Bytes()[1:3], uint16(buf.Len()))

	return buf.Bytes()
}

// createMultipleMessages creates a byte slice with multiple concatenated messages
func createMultipleMessages(count int) []byte {
	var buf bytes.Buffer
	for i := 0; i < count; i++ {
		// Create a message with different data for each iteration
		itemData := map[string][]byte{
			"I021/010": {byte(i), byte(i + 1)},
			"I021/040": {byte(i + 2)},
		}
		if i%2 == 0 {
			itemData["I021/030"] = []byte{byte(i + 3), byte(i + 4), byte(i + 5)}
		}
		msg := createTestMessage(Cat021, itemData)
		buf.Write(msg)
	}
	return buf.Bytes()
}

func TestNewDecoder(t *testing.T) {
	// Test with default options
	decoder := NewDecoder()
	if decoder == nil {
		t.Fatal("NewDecoder() returned nil")
	}
	if decoder.parallelism != DefaultParallelism {
		t.Errorf("parallelism = %d, want %d", decoder.parallelism, DefaultParallelism)
	}

	// Test with custom parallelism
	decoder = NewDecoder(WithDecoderParallelism(4))
	if decoder.parallelism != 4 {
		t.Errorf("parallelism = %d, want 4", decoder.parallelism)
	}

	// Test with custom buffer pool
	customPool := encoding.NewBufferPool()
	decoder = NewDecoder(WithDecoderBufferPool(customPool))
	if decoder.pool != customPool {
		t.Error("buffer pool not set correctly")
	}

	// Test with preloaded UAPs
	mockUAP := &MockUAP{category: Cat021}
	decoder = NewDecoder(WithPreloadedUAPs(mockUAP))
	if uap := decoder.GetUAP(Cat021); uap != mockUAP {
		t.Error("UAP not registered correctly")
	}
}

func TestRegisterUAP(t *testing.T) {
	decoder := NewDecoder()

	// Register a valid UAP
	mockUAP := &MockUAP{category: Cat021}
	decoder.RegisterUAP(mockUAP)
	if uap := decoder.GetUAP(Cat021); uap != mockUAP {
		t.Error("UAP not registered correctly")
	}

	// Register nil UAP should be a no-op
	decoder.RegisterUAP(nil)
	if uap := decoder.GetUAP(Cat021); uap != mockUAP {
		t.Error("Registering nil UAP modified existing UAP")
	}
}

func TestDecode(t *testing.T) {
	decoder, _ := setupTestDecoder()

	testCases := []struct {
		name     string
		itemData map[string][]byte
		wantErr  bool
	}{
		{
			"Valid message with mandatory items",
			map[string][]byte{
				"I021/010": {0xAA, 0xBB},
				"I021/040": {0xCC},
			},
			false,
		},
		{
			"Valid message with all items",
			map[string][]byte{
				"I021/010": {0xAA, 0xBB},
				"I021/040": {0xCC},
				"I021/030": {0xDD, 0xEE, 0xFF},
			},
			false,
		},
		{
			"Invalid message - missing mandatory item",
			map[string][]byte{
				"I021/010": {0xAA, 0xBB},
				// Missing I021/040
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := createTestMessage(Cat021, tc.itemData)

			// Decode the message
			block, err := decoder.Decode(msg)

			if tc.wantErr {
				if err == nil {
					t.Error("Decode() should have returned an error")
				}
			} else {
				if err != nil {
					t.Errorf("Decode() failed: %v", err)
				}
				if block == nil {
					t.Fatal("Decode() returned nil block")
				}
				if block.Category() != Cat021 {
					t.Errorf("Category = %v, want %v", block.Category(), Cat021)
				}
				if block.RecordCount() != 1 {
					t.Errorf("RecordCount = %d, want 1", block.RecordCount())
				}

				// Verify data items
				record := block.Records()[0]
				for id := range tc.itemData {
					if !record.HasDataItem(id) {
						t.Errorf("Record missing data item %s", id)
					}
				}
			}
		})
	}

	// Test error cases
	t.Run("Empty data", func(t *testing.T) {
		_, err := decoder.Decode([]byte{})
		if err == nil {
			t.Error("Decode() with empty data should return error")
		}
	})

	t.Run("Invalid category", func(t *testing.T) {
		// Create a message with an invalid category
		msg := createTestMessage(Category(0), map[string][]byte{
			"I021/010": {0xAA, 0xBB},
			"I021/040": {0xCC},
		})
		msg[0] = 0 // Set invalid category

		_, err := decoder.Decode(msg)
		if err == nil {
			t.Error("Decode() with invalid category should return error")
		}
	})

	t.Run("Unknown category", func(t *testing.T) {
		// Create a message with a category not registered with the decoder
		msg := createTestMessage(Cat048, map[string][]byte{
			"I021/010": {0xAA, 0xBB},
			"I021/040": {0xCC},
		})

		_, err := decoder.Decode(msg)
		if err == nil {
			t.Error("Decode() with unknown category should return error")
		}
	})
}

func TestDecodeFrom(t *testing.T) {
	decoder, _ := setupTestDecoder()

	// Create a test message
	itemData := map[string][]byte{
		"I021/010": {0xAA, 0xBB},
		"I021/040": {0xCC},
		"I021/030": {0xDD, 0xEE, 0xFF},
	}
	msg := createTestMessage(Cat021, itemData)

	// Test successful decoding
	t.Run("Success", func(t *testing.T) {
		r := bytes.NewReader(msg)
		block, err := decoder.DecodeFrom(r)
		if err != nil {
			t.Errorf("DecodeFrom() failed: %v", err)
		}
		if block == nil {
			t.Fatal("DecodeFrom() returned nil block")
		}
		if block.Category() != Cat021 {
			t.Errorf("Category = %v, want %v", block.Category(), Cat021)
		}
		if block.RecordCount() != 1 {
			t.Errorf("RecordCount = %d, want 1", block.RecordCount())
		}
	})

	// Test error cases
	t.Run("EOF on header", func(t *testing.T) {
		r := bytes.NewReader([]byte{})
		_, err := decoder.DecodeFrom(r)
		if err == nil {
			t.Error("DecodeFrom() with empty reader should return error")
		}
	})

	t.Run("EOF on body", func(t *testing.T) {
		// Create a truncated message (header only)
		r := bytes.NewReader(msg[:3])
		_, err := decoder.DecodeFrom(r)
		if err == nil {
			t.Error("DecodeFrom() with truncated message should return error")
		}
	})

	t.Run("Invalid category", func(t *testing.T) {
		// Create a message with an invalid category
		badMsg := make([]byte, len(msg))
		copy(badMsg, msg)
		badMsg[0] = 0 // Set invalid category

		r := bytes.NewReader(badMsg)
		_, err := decoder.DecodeFrom(r)
		if err == nil {
			t.Error("DecodeFrom() with invalid category should return error")
		}
	})

	t.Run("Invalid length", func(t *testing.T) {
		// Create a message with an invalid length
		badMsg := make([]byte, len(msg))
		copy(badMsg, msg)
		badMsg[1] = 0 // Set length to 0
		badMsg[2] = 0

		r := bytes.NewReader(badMsg)
		_, err := decoder.DecodeFrom(r)
		if err == nil {
			t.Error("DecodeFrom() with invalid length should return error")
		}
	})
}

func TestDecodeAll(t *testing.T) {
	decoder, _ := setupTestDecoder()

	// Test with multiple valid messages
	t.Run("Multiple valid messages", func(t *testing.T) {
		data := createMultipleMessages(5)
		blocks, err := decoder.DecodeAll(data)
		if err != nil {
			t.Errorf("DecodeAll() failed: %v", err)
		}
		if len(blocks) != 5 {
			t.Errorf("DecodeAll() returned %d blocks, want 5", len(blocks))
		}
		// Check each block
		for i, block := range blocks {
			if block.Category() != Cat021 {
				t.Errorf("Block %d: Category = %v, want %v", i, block.Category(), Cat021)
			}
			if block.RecordCount() != 1 {
				t.Errorf("Block %d: RecordCount = %d, want 1", i, block.RecordCount())
			}
		}
	})

	// Test with truncated last message
	t.Run("Truncated last message", func(t *testing.T) {
		data := createMultipleMessages(3)
		// Truncate the last message
		truncatedData := data[:len(data)-2]
		blocks, err := decoder.DecodeAll(truncatedData)
		if err == nil {
			t.Error("DecodeAll() with truncated data should return error")
		}
		// Should still decode the complete messages
		if len(blocks) != 2 {
			t.Errorf("DecodeAll() returned %d blocks, want 2", len(blocks))
		}
	})

	// Test with invalid message in the middle
	t.Run("Invalid message in the middle", func(t *testing.T) {
		data := createMultipleMessages(3)
		// Corrupt the second message's category
		data[len(createTestMessage(Cat021, nil))] = 0 // Set invalid category
		blocks, err := decoder.DecodeAll(data)
		if err == nil {
			t.Error("DecodeAll() with invalid message should return error")
		}
		// Should decode the first message
		if len(blocks) != 1 {
			t.Errorf("DecodeAll() returned %d blocks, want 1", len(blocks))
		}
	})
}

func TestDecodeParallel(t *testing.T) {
	decoder, _ := setupTestDecoder()
	decoder.parallelism = 4 // Use parallel processing

	// Create test messages
	messages := make([][]byte, 10)
	for i := range messages {
		itemData := map[string][]byte{
			"I021/010": {0xAA, byte(i)},
			"I021/040": {0xCC},
		}
		messages[i] = createTestMessage(Cat021, itemData)
	}

	// Test successful parallel decoding
	blocks, err := decoder.DecodeParallel(messages)
	if err != nil {
		t.Errorf("DecodeParallel() failed: %v", err)
	}
	if len(blocks) != len(messages) {
		t.Errorf("DecodeParallel() returned %d blocks, want %d", len(blocks), len(messages))
	}

	// Test with an invalid message
	messages[5] = []byte{0x15, 0x00, 0x03} // Invalid message (too short)
	blocks, err = decoder.DecodeParallel(messages)
	if err == nil {
		t.Error("DecodeParallel() with invalid message should return error")
	}
	// The valid messages should still be decoded
	for i, block := range blocks {
		if i == 5 {
			if block != nil {
				t.Errorf("Block %d should be nil due to error", i)
			}
		} else if block == nil {
			t.Errorf("Block %d should not be nil", i)
		}
	}

	// Test with single-threaded fallback
	decoder.parallelism = 1
	blocks, err = decoder.DecodeParallel(messages[:1]) // Single message
	if err != nil {
		t.Errorf("DecodeParallel() with single thread failed: %v", err)
	}
	if len(blocks) != 1 {
		t.Errorf("DecodeParallel() with single thread returned %d blocks, want 1", len(blocks))
	}
}

func TestStreamDecode(t *testing.T) {
	decoder, _ := setupTestDecoder()

	// Create multiple messages to simulate a stream
	data := createMultipleMessages(5)
	r := bytes.NewReader(data)

	// Test successful stream decoding
	var decodedCount int
	var mu sync.Mutex
	err := decoder.StreamDecode(r, func(block *DataBlock) error {
		mu.Lock()
		defer mu.Unlock()
		decodedCount++
		if block.Category() != Cat021 {
			return fmt.Errorf("unexpected category: %v", block.Category())
		}
		return nil
	})
	if err != nil {
		t.Errorf("StreamDecode() failed: %v", err)
	}
	if decodedCount != 5 {
		t.Errorf("StreamDecode() processed %d messages, want 5", decodedCount)
	}

	// Test with callback returning error
	r = bytes.NewReader(data)
	callbackErr := fmt.Errorf("callback error")
	err = decoder.StreamDecode(r, func(block *DataBlock) error {
		return callbackErr
	})
	if err == nil || err.Error() != "callback error: callback error" {
		t.Errorf("StreamDecode() should propagate callback error, got: %v", err)
	}

	// Test with invalid data in stream
	invalidData := append(createMultipleMessages(2), []byte{0x15, 0, 1}...) // Invalid length
	invalidData = append(invalidData, createMultipleMessages(1)...)
	r = bytes.NewReader(invalidData)
	decodedCount = 0
	err = decoder.StreamDecode(r, func(block *DataBlock) error {
		mu.Lock()
		defer mu.Unlock()
		decodedCount++
		return nil
	})
	if err == nil {
		t.Error("StreamDecode() with invalid data should return error")
	}
	// Should decode the valid messages before the error
	if decodedCount != 2 {
		t.Errorf("StreamDecode() processed %d messages before error, want 2", decodedCount)
	}

	// Test reset functionality
	decoder.ResetStream()
	// No real way to test this directly, but we can verify it doesn't throw an error
	if len(decoder.streamBuffer) != 0 {
		t.Errorf("ResetStream() did not clear buffer, length = %d", len(decoder.streamBuffer))
	}
}

func TestExtractMessages(t *testing.T) {
	decoder, _ := setupTestDecoder()

	// Create test data with valid messages surrounded by garbage
	validMessages := createMultipleMessages(3)
	data := append([]byte("garbage"), validMessages...)
	data = append(data, []byte("more garbage")...)

	// Extract valid messages
	messages, err := decoder.ExtractMessages(data)
	if err != nil {
		t.Errorf("ExtractMessages() failed: %v", err)
	}

	// Should find 3 valid messages
	if len(messages) != 3 {
		t.Errorf("ExtractMessages() found %d messages, want 3", len(messages))
	}

	// Verify each extracted message is valid
	for i, msg := range messages {
		// Category should be valid
		if Category(msg[0]) != Cat021 {
			t.Errorf("Message %d has invalid category %d", i, msg[0])
		}
		// Length should match
		length := int(msg[1])<<8 | int(msg[2])
		if length != len(msg) {
			t.Errorf("Message %d has incorrect length: got %d, want %d", i, len(msg), length)
		}
	}

	// Test with invalid data
	invalidData := []byte{0x15, 0xFF, 0xFF} // Valid category but invalid length
	messages, err = decoder.ExtractMessages(invalidData)
	if err != nil {
		t.Errorf("ExtractMessages() with invalid data failed: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("ExtractMessages() with invalid data found %d messages, want 0", len(messages))
	}
}

func BenchmarkDecode(b *testing.B) {
	decoder, _ := setupTestDecoder()

	// Create a test message
	itemData := map[string][]byte{
		"I021/010": {0xAA, 0xBB},
		"I021/040": {0xCC},
		"I021/030": {0xDD, 0xEE, 0xFF},
	}
	msg := createTestMessage(Cat021, itemData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.Decode(msg)
		if err != nil {
			b.Fatalf("Decode() failed: %v", err)
		}
	}
}

func BenchmarkDecodeAll(b *testing.B) {
	decoder, _ := setupTestDecoder()

	// Create multiple messages
	data := createMultipleMessages(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.DecodeAll(data)
		if err != nil {
			b.Fatalf("DecodeAll() failed: %v", err)
		}
	}
}

func BenchmarkDecodeParallel(b *testing.B) {
	decoder, _ := setupTestDecoder()
	decoder.parallelism = 4

	// Create test messages
	messages := make([][]byte, 10)
	for i := range messages {
		itemData := map[string][]byte{
			"I021/010": {0xAA, byte(i)},
			"I021/040": {0xCC},
		}
		messages[i] = createTestMessage(Cat021, itemData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.DecodeParallel(messages)
		if err != nil {
			b.Fatalf("DecodeParallel() failed: %v", err)
		}
	}
}

func BenchmarkStreamDecode(b *testing.B) {
	decoder, _ := setupTestDecoder()

	// Create multiple messages
	data := createMultipleMessages(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		err := decoder.StreamDecode(r, func(block *DataBlock) error {
			return nil
		})
		if err != nil {
			b.Fatalf("StreamDecode() failed: %v", err)
		}
	}
}
