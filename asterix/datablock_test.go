// asterix/datablock_test.go
package asterix

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func setupTestDataBlock() (*DataBlock, *MockUAP, error) {
	// Same UAP as in record_test.go
	uap := &MockUAP{
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

	dataBlock, err := NewDataBlock(Cat021, uap)
	if err != nil {
		return nil, nil, err
	}

	return dataBlock, uap, nil
}

func createTestRecord(t *testing.T, dataBlock *DataBlock) *Record {
	record, err := NewRecord(dataBlock.Category(), dataBlock.UAP())
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Add required items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	return record
}

func TestNewDataBlock(t *testing.T) {
	// Valid case
	_, uap, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	dataBlock, err := NewDataBlock(Cat021, uap)
	if err != nil {
		t.Errorf("NewDataBlock with valid parameters failed: %v", err)
	}
	if dataBlock == nil {
		t.Fatal("NewDataBlock returned nil")
	}

	// Invalid category
	_, err = NewDataBlock(Category(0), uap)
	if err == nil {
		t.Error("NewDataBlock with invalid category should fail")
	}

	// Nil UAP
	_, err = NewDataBlock(Cat021, nil)
	if err == nil {
		t.Error("NewDataBlock with nil UAP should fail")
	}

	// Mismatched category
	_, err = NewDataBlock(Cat048, uap) // UAP is for Cat021
	if err == nil {
		t.Error("NewDataBlock with mismatched category should fail")
	}
}

func TestDataBlockAddRecord(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add a valid record
	record := createTestRecord(t, dataBlock)
	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Errorf("AddRecord failed: %v", err)
	}

	// Add nil record
	err = dataBlock.AddRecord(nil)
	if err == nil {
		t.Error("AddRecord with nil record should fail")
	}

	// Add record with mismatched category
	wrongCatUAP := &MockUAP{
		category: Cat048,
		version:  "1.0",
		fields:   []DataField{},
	}
	wrongCatRecord, err := NewRecord(Cat048, wrongCatUAP)
	if err != nil {
		t.Fatalf("Failed to create record with wrong category: %v", err)
	}
	err = dataBlock.AddRecord(wrongCatRecord)
	if err == nil {
		t.Error("AddRecord with mismatched category should fail")
	}
}

func TestDataBlockRecords(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some records
	for i := 0; i < 3; i++ {
		record := createTestRecord(t, dataBlock)
		err = dataBlock.AddRecord(record)
		if err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}
	}

	// Get records
	records := dataBlock.Records()
	if len(records) != 3 {
		t.Errorf("Records() returned %d records, want 3", len(records))
	}

	// Verify Records() returns a copy
	originalCount := dataBlock.RecordCount()
	records = append(records, records[0]) // Modify the slice
	if dataBlock.RecordCount() != originalCount {
		t.Error("Records() should return a copy, not the original")
	}
}

func TestDataBlockEncodeDecode(t *testing.T) {
	testCases := []struct {
		name       string
		numRecords int
		blockable  bool
	}{
		{"Single record, non-blockable", 1, false},
		{"Single record, blockable", 1, true},
		{"Multiple records, non-blockable", 3, false},
		{"Multiple records, blockable", 3, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataBlock, _, err := setupTestDataBlock()
			if err != nil {
				t.Fatalf("Failed to set up test: %v", err)
			}
			dataBlock.SetBlockable(tc.blockable)

			// Add records
			for i := 0; i < tc.numRecords; i++ {
				record := createTestRecord(t, dataBlock)
				// Add optional item to make records different
				if i%2 == 0 {
					err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, byte(i), 0xFF}, fixedLen: 3})
					if err != nil {
						t.Fatalf("Failed to set data item: %v", err)
					}
				}
				err = dataBlock.AddRecord(record)
				if err != nil {
					t.Fatalf("Failed to add record: %v", err)
				}
			}

			// Encode
			data, err := dataBlock.Encode()
			if err != nil {
				t.Errorf("Encode failed: %v", err)
			}

			// Basic checks on encoded data
			if len(data) < 3 {
				t.Fatalf("Encoded data too short: %d bytes", len(data))
			}
			if Category(data[0]) != Cat021 {
				t.Errorf("Encoded category = %d, want %d", data[0], Cat021)
			}
			length := binary.BigEndian.Uint16(data[1:3])
			if int(length) != len(data) {
				t.Errorf("Encoded length = %d, actual length = %d", length, len(data))
			}

			// Decode into a new data block
			newDataBlock, _, err := setupTestDataBlock()
			if err != nil {
				t.Fatalf("Failed to set up test: %v", err)
			}
			err = newDataBlock.Decode(data)
			if err != nil {
				t.Errorf("Decode failed: %v", err)
			}

			// Check record count
			if newDataBlock.RecordCount() != tc.numRecords {
				t.Errorf("Decoded block has %d records, want %d", newDataBlock.RecordCount(), tc.numRecords)
			}

			// Check records contain expected data
			for i, record := range newDataBlock.Records() {
				if !record.HasDataItem("I021/010") {
					t.Errorf("Record %d missing I021/010", i)
				}
				if !record.HasDataItem("I021/040") {
					t.Errorf("Record %d missing I021/040", i)
				}
				// Check optional item
				if i%2 == 0 && !record.HasDataItem("I021/030") {
					t.Errorf("Record %d missing optional item I021/030", i)
				}
			}
		})
	}
}

func TestDataBlockEncodeToDecodefrom(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add a record
	record := createTestRecord(t, dataBlock)
	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Encode to buffer
	buf := new(bytes.Buffer)
	err = dataBlock.EncodeTo(buf)
	if err != nil {
		t.Errorf("EncodeTo failed: %v", err)
	}

	// Decode from buffer
	newDataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	err = newDataBlock.DecodeFrom(buf)
	if err != nil {
		t.Errorf("DecodeFrom failed: %v", err)
	}

	// Check record count
	if newDataBlock.RecordCount() != 1 {
		t.Errorf("Decoded block has %d records, want 1", newDataBlock.RecordCount())
	}
}

func TestDataBlockDecodeErrors(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Data too short
	err = dataBlock.Decode([]byte{0x15})
	if err == nil {
		t.Error("Decode with too short data should fail")
	}

	// Wrong category
	err = dataBlock.Decode([]byte{0x10, 0x00, 0x03})
	if err == nil {
		t.Error("Decode with wrong category should fail")
	}

	// Wrong length
	err = dataBlock.Decode([]byte{0x15, 0x00, 0x20, 0x01, 0x02})
	if err == nil {
		t.Error("Decode with wrong length should fail")
	}

	// Invalid record
	// Create valid header but invalid record content
	data := []byte{0x15, 0x00, 0x05, 0x01, 0x01} // Cat021, 5 bytes, invalid FSPEC
	err = dataBlock.Decode(data)
	if err == nil {
		t.Error("Decode with invalid record should fail")
	}
}

func TestDataBlockDecodeFromErrors(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// EOF on header
	err = dataBlock.DecodeFrom(strings.NewReader(""))
	if err == nil {
		t.Error("DecodeFrom with EOF on header should fail")
	}

	// Wrong category
	buf := bytes.NewBuffer([]byte{0x10, 0x00, 0x03}) // Cat016, 3 bytes
	err = dataBlock.DecodeFrom(buf)
	if err == nil {
		t.Error("DecodeFrom with wrong category should fail")
	}

	// Invalid length
	buf = bytes.NewBuffer([]byte{0x15, 0x00, 0x02}) // Cat021, 2 bytes (too small)
	err = dataBlock.DecodeFrom(buf)
	if err == nil {
		t.Error("DecodeFrom with invalid length should fail")
	}

	// EOF on body
	buf = bytes.NewBuffer([]byte{0x15, 0x00, 0x10}) // Cat021, 16 bytes, but only 3 available
	err = dataBlock.DecodeFrom(buf)
	if err == nil {
		t.Error("DecodeFrom with EOF on body should fail")
	}
}

func TestDataBlockClear(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some records
	for i := 0; i < 3; i++ {
		record := createTestRecord(t, dataBlock)
		err = dataBlock.AddRecord(record)
		if err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}
	}

	// Clear the data block
	dataBlock.Clear()

	// Check record count
	if dataBlock.RecordCount() != 0 {
		t.Errorf("After Clear(), record count = %d, want 0", dataBlock.RecordCount())
	}
}

func TestDataBlockEstimateSize(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Empty data block
	size := dataBlock.EstimateSize()
	if size != 3 {
		t.Errorf("Empty data block size estimate = %d, want 3", size)
	}

	// Add some records
	for i := 0; i < 3; i++ {
		record := createTestRecord(t, dataBlock)
		err = dataBlock.AddRecord(record)
		if err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}
	}

	// Check size estimate
	size = dataBlock.EstimateSize()
	if size <= 3 {
		t.Errorf("Size estimate = %d, should be > 3 for non-empty block", size)
	}
}

func TestDataBlockEncodeRecord(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Prepare items
	items := map[string]DataItem{
		"I021/010": &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2},
		"I021/040": &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1},
		"I021/030": &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3},
	}

	// Encode the record
	err = dataBlock.EncodeRecord(items)
	if err != nil {
		t.Errorf("EncodeRecord failed: %v", err)
	}

	// Check record count
	if dataBlock.RecordCount() != 1 {
		t.Errorf("After EncodeRecord(), record count = %d, want 1", dataBlock.RecordCount())
	}

	// Check record contains expected items
	record := dataBlock.Records()[0]
	if !record.HasDataItem("I021/010") {
		t.Error("Record missing I021/010")
	}
	if !record.HasDataItem("I021/040") {
		t.Error("Record missing I021/040")
	}
	if !record.HasDataItem("I021/030") {
		t.Error("Record missing I021/030")
	}

	// Test with invalid item
	invalidItems := map[string]DataItem{
		"I021/999": &MockDataItem{id: "I021/999", data: []byte{0xAA, 0xBB}, fixedLen: 2}, // Invalid ID
	}
	err = dataBlock.EncodeRecord(invalidItems)
	if err == nil {
		t.Error("EncodeRecord with invalid item should fail")
	}
}

func TestDataBlockIsASRS(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Empty data block should return true
	if !dataBlock.IsASRS() {
		t.Error("IsASRS() should return true for empty data block")
	}

	// Add identical records
	for i := 0; i < 3; i++ {
		record := createTestRecord(t, dataBlock)
		err = dataBlock.AddRecord(record)
		if err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}
	}

	// Should be all same record structure
	if !dataBlock.IsASRS() {
		t.Error("IsASRS() should return true for identical records")
	}

	// Add a different record
	record := createTestRecord(t, dataBlock)
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Should no longer be all same record structure
	if dataBlock.IsASRS() {
		t.Error("IsASRS() should return false for different records")
	}
}

func TestDataBlockClone(t *testing.T) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some records
	for i := 0; i < 3; i++ {
		record := createTestRecord(t, dataBlock)
		if i%2 == 0 {
			err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, byte(i), 0xFF}, fixedLen: 3})
			if err != nil {
				t.Fatalf("Failed to set data item: %v", err)
			}
		}
		err = dataBlock.AddRecord(record)
		if err != nil {
			t.Fatalf("Failed to add record: %v", err)
		}
	}

	// Clone the data block
	clone, err := dataBlock.Clone()
	if err != nil {
		t.Errorf("Clone failed: %v", err)
	}

	// Check properties
	if clone.Category() != dataBlock.Category() {
		t.Errorf("Clone category = %v, want %v", clone.Category(), dataBlock.Category())
	}
	if clone.RecordCount() != dataBlock.RecordCount() {
		t.Errorf("Clone record count = %d, want %d", clone.RecordCount(), dataBlock.RecordCount())
	}
	if clone.Blockable() != dataBlock.Blockable() {
		t.Errorf("Clone blockable = %v, want %v", clone.Blockable(), dataBlock.Blockable())
	}

	// Verify independence
	record := createTestRecord(t, dataBlock)
	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}
	if clone.RecordCount() == dataBlock.RecordCount() {
		t.Error("Modifying original should not affect clone")
	}
}

func TestDataBlockGetters(t *testing.T) {
	dataBlock, uap, err := setupTestDataBlock()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Test Category()
	if dataBlock.Category() != Cat021 {
		t.Errorf("Category() = %v, want %v", dataBlock.Category(), Cat021)
	}

	// Test UAP()
	if dataBlock.UAP() != uap {
		t.Errorf("UAP() = %v, want %v", dataBlock.UAP(), uap)
	}

	// Test Blockable() and SetBlockable()
	if !dataBlock.Blockable() {
		t.Error("Blockable() should be true for Cat021")
	}
	dataBlock.SetBlockable(false)
	if dataBlock.Blockable() {
		t.Error("Blockable() should be false after SetBlockable(false)")
	}
}

func BenchmarkDataBlockEncode(b *testing.B) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
	}

	// Add some records
	for i := 0; i < 10; i++ {
		record, err := NewRecord(dataBlock.Category(), dataBlock.UAP())
		if err != nil {
			b.Fatalf("Failed to create record: %v", err)
		}

		// Add required items
		err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
		if err != nil {
			b.Fatalf("Failed to set data item: %v", err)
		}
		err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
		if err != nil {
			b.Fatalf("Failed to set data item: %v", err)
		}

		// Add optional item
		if i%2 == 0 {
			err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, byte(i), 0xFF}, fixedLen: 3})
			if err != nil {
				b.Fatalf("Failed to set data item: %v", err)
			}
		}

		err = dataBlock.AddRecord(record)
		if err != nil {
			b.Fatalf("Failed to add record: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := dataBlock.Encode()
		if err != nil {
			b.Fatalf("Encode failed: %v", err)
		}
	}
}

func BenchmarkDataBlockDecode(b *testing.B) {
	// Prepare a data block for encoding
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
	}

	// Add some records
	for i := 0; i < 10; i++ {
		record, err := NewRecord(dataBlock.Category(), dataBlock.UAP())
		if err != nil {
			b.Fatalf("Failed to create record: %v", err)
		}

		// Add required items
		err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
		if err != nil {
			b.Fatalf("Failed to set data item: %v", err)
		}
		err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
		if err != nil {
			b.Fatalf("Failed to set data item: %v", err)
		}

		// Add optional item
		if i%2 == 0 {
			err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, byte(i), 0xFF}, fixedLen: 3})
			if err != nil {
				b.Fatalf("Failed to set data item: %v", err)
			}
		}

		err = dataBlock.AddRecord(record)
		if err != nil {
			b.Fatalf("Failed to add record: %v", err)
		}
	}

	// Encode to get bytes
	data, err := dataBlock.Encode()
	if err != nil {
		b.Fatalf("Failed to encode data block: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a new data block for each iteration
		newDataBlock, _, err := setupTestDataBlock()
		if err != nil {
			b.Fatalf("Failed to set up test: %v", err)
		}

		// Decode the data block
		err = newDataBlock.Decode(data)
		if err != nil {
			b.Fatalf("Decode failed: %v", err)
		}
	}
}

func BenchmarkDataBlockEncodeRecord(b *testing.B) {
	dataBlock, _, err := setupTestDataBlock()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
	}

	// Prepare items
	items := map[string]DataItem{
		"I021/010": &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2},
		"I021/040": &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1},
		"I021/030": &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset for each iteration
		dataBlock.Clear()

		// Encode the record
		err = dataBlock.EncodeRecord(items)
		if err != nil {
			b.Fatalf("EncodeRecord failed: %v", err)
		}
	}
}
