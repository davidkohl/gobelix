// asterix/record_test.go
package asterix

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

// FixedLength returns the fixed length of the item
func (m *MockDataItem) FixedLength() int {
	return m.fixedLen
}

// Setup creates a test UAP and Record
func setupTestRecord() (*Record, *MockUAP, error) {
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

	record, err := NewRecord(Cat021, uap)
	if err != nil {
		return nil, nil, err
	}

	return record, uap, nil
}

func TestNewRecord(t *testing.T) {
	// Valid case
	_, uap, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	_, err = NewRecord(Cat021, uap)
	if err != nil {
		t.Errorf("NewRecord with valid parameters failed: %v", err)
	}

	// Invalid category
	_, err = NewRecord(Category(0), uap)
	if err == nil {
		t.Error("NewRecord with invalid category should fail")
	}

	// Nil UAP
	_, err = NewRecord(Cat021, nil)
	if err == nil {
		t.Error("NewRecord with nil UAP should fail")
	}

	// Mismatched category
	_, err = NewRecord(Cat048, uap) // UAP is for Cat021
	if err == nil {
		t.Error("NewRecord with mismatched category should fail")
	}
}

func TestRecordSetGetDataItem(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add a valid item
	item1 := &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2}
	err = record.SetDataItem("I021/010", item1)
	if err != nil {
		t.Errorf("SetDataItem failed: %v", err)
	}

	// Add another valid item
	item2 := &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1}
	err = record.SetDataItem("I021/040", item2)
	if err != nil {
		t.Errorf("SetDataItem failed: %v", err)
	}

	// Get item and verify
	gotItem1, exists := record.GetDataItem("I021/010")
	if !exists {
		t.Error("GetDataItem should find I021/010")
	}
	if gotItem1 != item1 {
		t.Error("GetDataItem returned wrong item")
	}

	// Test HasDataItem
	if !record.HasDataItem("I021/010") {
		t.Error("HasDataItem should return true for existing item")
	}
	if record.HasDataItem("I021/999") {
		t.Error("HasDataItem should return false for non-existent item")
	}

	// Test invalid item ID
	err = record.SetDataItem("I021/999", item1)
	if err == nil {
		t.Error("SetDataItem with unknown ID should fail")
	}

	// Test nil item
	err = record.SetDataItem("I021/010", nil)
	if err == nil {
		t.Error("SetDataItem with nil item should fail")
	}

	// Test item with validation error
	invalidItem := &MockDataItem{
		id:          "I021/010",
		validateErr: fmt.Errorf("validation failed"),
	}
	err = record.SetDataItem("I021/010", invalidItem)
	if err == nil {
		t.Error("SetDataItem with invalid item should fail")
	}
}

func TestRecordEncodeDecode(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
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

	// Optional item
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := record.Encode(buf)
	if err != nil {
		t.Errorf("Encode failed: %v", err)
	}

	expectedSize := 7 // FSPEC (1) + Item1 (2) + Item2 (1) + Item3 (3)
	if n != expectedSize {
		t.Errorf("Encode wrote %d bytes, expected %d", n, expectedSize)
	}

	// Create a new record for decoding
	newRecord, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Decode
	n, err = newRecord.Decode(buf)
	if err != nil {
		t.Errorf("Decode failed: %v", err)
	}
	if n != expectedSize {
		t.Errorf("Decode read %d bytes, expected %d", n, expectedSize)
	}

	// Verify items were decoded
	if !newRecord.HasDataItem("I021/010") {
		t.Error("Decoded record missing I021/010")
	}
	if !newRecord.HasDataItem("I021/040") {
		t.Error("Decoded record missing I021/040")
	}
	if !newRecord.HasDataItem("I021/030") {
		t.Error("Decoded record missing I021/030")
	}

	// Verify item count
	if newRecord.ItemCount() != 3 {
		t.Errorf("Decoded record has %d items, expected 3", newRecord.ItemCount())
	}
}

func TestRecordEncodeErrors(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Missing mandatory field
	buf := new(bytes.Buffer)
	_, err = record.Encode(buf)
	if err == nil {
		t.Error("Encode should fail with missing mandatory field")
	}
	if !IsMandatoryFieldMissing(err) {
		t.Errorf("Error should be of mandatory field type, got: %v", err)
	}

	// Item marked in FSPEC but not present
	// First ensure all mandatory fields are present
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set mandatory item 010: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set mandatory item 040: %v", err)
	}
	// Add a non-mandatory field
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xAA, 0xBB, 0xCC}, fixedLen: 3})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	// Then remove the non-mandatory item but leave the FSPEC bit set
	delete(record.items, "I021/030")
	_, err = record.Encode(buf)
	if err == nil {
		t.Error("Encode should fail with item in FSPEC but not present")
	}
	// This should be an encoding error (FSPEC mismatch), not a validation error
	if !IsEncodingError(err) {
		t.Errorf("Error should be encoding error, got: %v", err)
	}

	// Item with encoding error
	record, _, err = setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	err = record.SetDataItem("I021/010", &MockDataItem{
		id:        "I021/010",
		data:      []byte{0xAA, 0xBB},
		fixedLen:  2,
		encodeErr: fmt.Errorf("encode error"),
	})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	_, err = record.Encode(buf)
	if err == nil {
		t.Error("Encode should fail with item encoding error")
	}
	if !IsEncodingError(err) {
		t.Errorf("Error should be encoding error, got: %v", err)
	}
}

func TestRecordDecodeErrors(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Empty buffer
	buf := new(bytes.Buffer)
	_, err = record.Decode(buf)
	if err != io.EOF {
		t.Errorf("Decode with empty buffer should return EOF, got: %v", err)
	}

	// Invalid FSPEC
	buf = bytes.NewBuffer([]byte{0x01}) // FSPEC with extension bit set but no more bytes
	_, err = record.Decode(buf)
	if err == nil {
		t.Error("Decode with invalid FSPEC should fail")
	}
	if !IsDecodeError(err) {
		t.Errorf("Error should be decode error, got: %v", err)
	}

	// Item creation error
	buf = bytes.NewBuffer([]byte{0x80}) // FSPEC with bit 1 set (FRN 1)
	mockUAP := record.uap.(*MockUAP)
	mockUAP.fields[0].DataItem = "UnknownItem" // Change the item ID to cause creation error
	_, err = record.Decode(buf)
	if err == nil {
		t.Error("Decode with item creation error should fail")
	}
	if !IsDecodeError(err) {
		t.Errorf("Error should be decode error, got: %v", err)
	}

	// Restore UAP for further tests
	record, _, err = setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Buffer too short for fixed-length item
	buf = bytes.NewBuffer([]byte{0x80}) // FSPEC with bit 1 set (FRN 1)
	_, err = record.Decode(buf)
	if err == nil {
		t.Error("Decode with buffer too short should fail")
	}
	if !IsDecodeError(err) {
		t.Errorf("Error should be decode error, got: %v", err)
	}

	// Item with decode error
	record, _, err = setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	mockUAP = record.uap.(*MockUAP)
	mockUAP.createItemFunc = func(id string) (DataItem, error) {
		if id == "I021/010" {
			return &MockDataItem{
				id:        id,
				fixedLen:  2,
				decodeErr: fmt.Errorf("decode error"),
			}, nil
		}
		return nil, fmt.Errorf("unknown item: %s", id)
	}

	buf = bytes.NewBuffer([]byte{0x80, 0x01, 0x02}) // FSPEC with bit 1 set (FRN 1) + 2 bytes of data
	_, err = record.Decode(buf)
	if err == nil {
		t.Error("Decode with item decode error should fail")
	}
	if !IsDecodeError(err) {
		t.Errorf("Error should be decode error, got: %v", err)
	}
}

func TestRecordClone(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Clone the record
	clone, err := record.Clone()
	if err != nil {
		t.Errorf("Clone failed: %v", err)
	}

	// Verify properties
	if clone.Category() != record.Category() {
		t.Errorf("Clone category = %v, want %v", clone.Category(), record.Category())
	}
	if clone.ItemCount() != record.ItemCount() {
		t.Errorf("Clone item count = %d, want %d", clone.ItemCount(), record.ItemCount())
	}
	if !clone.HasDataItem("I021/010") || !clone.HasDataItem("I021/040") {
		t.Error("Clone missing expected items")
	}

	// Verify independence
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	if clone.HasDataItem("I021/030") {
		t.Error("Modifying original record should not affect clone")
	}
}

func TestRecordReset(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Reset the record
	record.Reset()

	// Verify it's empty
	if record.ItemCount() != 0 {
		t.Errorf("Reset record has %d items, expected 0", record.ItemCount())
	}
	if record.HasDataItem("I021/010") || record.HasDataItem("I021/040") {
		t.Error("Reset record should not have any items")
	}
}

func TestRecordEstimateSize(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Empty record
	size := record.EstimateSize()
	if size != 0 {
		t.Errorf("Empty record size estimate = %d, want 0", size)
	}

	// Add some items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Check size estimate
	size = record.EstimateSize()
	if size < 4 { // FSPEC (1) + Item1 (2) + Item2 (1)
		t.Errorf("Size estimate = %d, want at least 4", size)
	}
}

func TestRecordValidate(t *testing.T) {
	record, _, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Empty record should fail validation (missing mandatory items)
	err = record.Validate()
	if err == nil {
		t.Error("Validate should fail for empty record")
	}
	if !IsMandatoryFieldMissing(err) {
		t.Errorf("Error should be of mandatory field type, got: %v", err)
	}

	// Add mandatory items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Now validation should pass
	err = record.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestRecordGetters(t *testing.T) {
	record, uap, err := setupTestRecord()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Add some items
	err = record.SetDataItem("I021/010", &MockDataItem{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}
	err = record.SetDataItem("I021/040", &MockDataItem{id: "I021/040", data: []byte{0xCC}, fixedLen: 1})
	if err != nil {
		t.Fatalf("Failed to set data item: %v", err)
	}

	// Test Category()
	if record.Category() != Cat021 {
		t.Errorf("Category() = %v, want %v", record.Category(), Cat021)
	}

	// Test UAP()
	if record.UAP() != uap {
		t.Errorf("UAP() = %v, want %v", record.UAP(), uap)
	}

	// Test FSPEC()
	if record.FSPEC() == nil {
		t.Error("FSPEC() should not return nil")
	}

	// Test Items()
	items := record.Items()
	if len(items) != 2 {
		t.Errorf("Items() returned %d items, want 2", len(items))
	}
	if _, exists := items["I021/010"]; !exists {
		t.Error("Items() missing I021/010")
	}
	if _, exists := items["I021/040"]; !exists {
		t.Error("Items() missing I021/040")
	}

	// Test ItemCount()
	if record.ItemCount() != 2 {
		t.Errorf("ItemCount() = %d, want 2", record.ItemCount())
	}
}

func BenchmarkRecordEncode(b *testing.B) {
	record, _, err := setupTestRecord()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
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
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3})
	if err != nil {
		b.Fatalf("Failed to set data item: %v", err)
	}

	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		record.Encode(buf)
	}
}

func BenchmarkRecordDecode(b *testing.B) {
	// Prepare a record for encoding
	record, _, err := setupTestRecord()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
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
	err = record.SetDataItem("I021/030", &MockDataItem{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3})
	if err != nil {
		b.Fatalf("Failed to set data item: %v", err)
	}

	// Encode to get bytes
	buf := new(bytes.Buffer)
	_, err = record.Encode(buf)
	if err != nil {
		b.Fatalf("Failed to encode record: %v", err)
	}
	data := buf.Bytes()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a new record for each iteration
		newRecord, _, err := setupTestRecord()
		if err != nil {
			b.Fatalf("Failed to set up test: %v", err)
		}

		// Decode the record
		newBuf := bytes.NewBuffer(data)
		newRecord.Decode(newBuf)
	}
}

func BenchmarkRecordSetDataItem(b *testing.B) {
	record, _, err := setupTestRecord()
	if err != nil {
		b.Fatalf("Failed to set up test: %v", err)
	}

	items := []*MockDataItem{
		{id: "I021/010", data: []byte{0xAA, 0xBB}, fixedLen: 2},
		{id: "I021/040", data: []byte{0xCC}, fixedLen: 1},
		{id: "I021/030", data: []byte{0xDD, 0xEE, 0xFF}, fixedLen: 3},
	}
	ids := []string{"I021/010", "I021/040", "I021/030"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset for each iteration
		record.Reset()

		// Set all items
		for j := 0; j < len(items); j++ {
			record.SetDataItem(ids[j], items[j])
		}
	}
}
