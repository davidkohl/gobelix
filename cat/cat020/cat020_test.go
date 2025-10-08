// cat/cat020/cat020_test.go
package cat020_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat020"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestCAT020_DecodeSAC_SIC(t *testing.T) {
	// Create UAP
	uap, err := cat020.NewUAP(cat020.Version15)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Create decoder
	decoder := asterix.NewDecoder(
		asterix.WithPreloadedUAPs(uap),
	)

	// Create a minimal CAT020 message with just mandatory fields:
	// - CAT: 020 (1 byte)
	// - LEN: 9 (2 bytes) - CAT(1) + LEN(2) + FSPEC(1) + I020/010(2) + I020/140(3) = 9 bytes
	// - FSPEC: 0xA0 (10100000) = FRN 1 and 3 set (I020/010 and I020/140)
	// - I020/010: SAC=25, SIC=10 (2 bytes: 0x19 0x0A)
	// - I020/140: Time of Day = 12345.0 seconds (3 bytes)
	//   12345.0 * 128 = 1580160 = 0x181880
	//   Bytes: 0x18 0x18 0x80

	data := []byte{
		0x14,       // CAT 020
		0x00, 0x09, // LEN = 9 bytes
		0xA0,             // FSPEC: FRN 1 and 3
		0x19, 0x0A,       // I020/010: SAC=25, SIC=10
		0x18, 0x18, 0x80, // I020/140: Time = 12345.0 seconds
	}

	// Decode
	msg, err := decoder.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify category
	if msg.Category() != asterix.Cat020 {
		t.Errorf("Expected category 020, got %d", msg.Category())
	}

	// Verify record count
	if msg.RecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", msg.RecordCount())
	}

	// Get the first record
	records := msg.Records()
	if len(records) == 0 {
		t.Fatal("No records found")
	}
	record := records[0]

	// Check I020/010 (SAC/SIC)
	item, exists := record.GetDataItem("I020/010")
	if !exists {
		t.Fatal("I020/010 not found in record")
	}

	dsi, ok := item.(*common.DataSourceIdentifier)
	if !ok {
		t.Fatalf("I020/010 is not a DataSourceIdentifier, got %T", item)
	}

	if dsi.SAC != 25 {
		t.Errorf("Expected SAC=25, got %d", dsi.SAC)
	}

	if dsi.SIC != 10 {
		t.Errorf("Expected SIC=10, got %d", dsi.SIC)
	}

	t.Logf("✓ Successfully decoded SAC=%d, SIC=%d", dsi.SAC, dsi.SIC)
}

func TestCAT020_EncodeDecodeSAC_SIC(t *testing.T) {
	// Create UAP
	uap, err := cat020.NewUAP(cat020.Version15)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Create a record
	record, err := asterix.NewRecord(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Set mandatory fields
	record.SetDataItem("I020/010", &common.DataSourceIdentifier{SAC: 42, SIC: 99})
	record.SetDataItem("I020/140", &common.TimeOfDay{TimeOfDay: 3600.5}) // 1 hour + 0.5 seconds

	// Create data block
	dataBlock, err := asterix.NewDataBlock(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("Failed to create data block: %v", err)
	}

	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Encode
	encoder := asterix.NewEncoder()
	data, err := encoder.Encode(dataBlock)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded %d bytes: % X", len(data), data)

	// Decode back
	decoder := asterix.NewDecoder(
		asterix.WithPreloadedUAPs(uap),
	)

	msg, err := decoder.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify
	records := msg.Records()
	if len(records) == 0 {
		t.Fatal("No records found after decode")
	}

	item, exists := records[0].GetDataItem("I020/010")
	if !exists {
		t.Fatal("I020/010 not found after decode")
	}

	dsi, ok := item.(*common.DataSourceIdentifier)
	if !ok {
		t.Fatalf("I020/010 is not a DataSourceIdentifier after decode, got %T", item)
	}

	if dsi.SAC != 42 {
		t.Errorf("Expected SAC=42 after round-trip, got %d", dsi.SAC)
	}

	if dsi.SIC != 99 {
		t.Errorf("Expected SIC=99 after round-trip, got %d", dsi.SIC)
	}

	t.Logf("✓ Round-trip successful: SAC=%d, SIC=%d", dsi.SAC, dsi.SIC)
}
