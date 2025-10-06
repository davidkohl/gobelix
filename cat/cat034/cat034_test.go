// cat/cat034/cat034_test.go
package cat034_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat034"
	v129 "github.com/davidkohl/gobelix/cat/cat034/dataitems/v129"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestCat034UAP(t *testing.T) {
	// Create UAP
	uap, err := cat034.NewUAP(cat034.Version129)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	if uap == nil {
		t.Fatal("UAP is nil")
	}

	// Verify it's for Cat034
	if uap.Category() != asterix.Cat034 {
		t.Errorf("Expected category 34, got %d", uap.Category())
	}

	// Verify version
	if uap.Version() != "1.29" {
		t.Errorf("Expected version 1.29, got %s", uap.Version())
	}
}

func TestCat034EncodeDecode(t *testing.T) {
	// Create UAP
	uap, err := cat034.NewUAP(cat034.Version129)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Create a record
	record, err := asterix.NewRecord(asterix.Cat034, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Add mandatory data items
	dsi := &common.DataSourceIdentifier{SAC: 100, SIC: 1}
	if err := record.SetDataItem("I034/010", dsi); err != nil {
		t.Fatalf("Failed to set I034/010: %v", err)
	}

	msgType := v129.NewMessageType()
	msgType.MessageType = 1 // North marker
	if err := record.SetDataItem("I034/000", msgType); err != nil {
		t.Fatalf("Failed to set I034/000: %v", err)
	}

	tod := &common.TimeOfDay{TimeOfDay: 43200.5} // Noon + 0.5s
	if err := record.SetDataItem("I034/030", tod); err != nil {
		t.Fatalf("Failed to set I034/030: %v", err)
	}

	// Add optional sector number
	sectorNum := v129.NewSectorNumber()
	sectorNum.SectorNumber = 45.0 // 45 degrees
	if err := record.SetDataItem("I034/020", sectorNum); err != nil {
		t.Fatalf("Failed to set I034/020: %v", err)
	}

	// Create a data block
	dataBlock, err := asterix.NewDataBlock(asterix.Cat034, uap)
	if err != nil {
		t.Fatalf("Failed to create data block: %v", err)
	}

	if err := dataBlock.AddRecord(record); err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Encode
	encoded, err := dataBlock.Encode()
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Decode
	decoder := asterix.NewDecoder(
		asterix.WithPreloadedUAPs(uap),
	)

	decoded, err := decoder.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Category() != asterix.Cat034 {
		t.Errorf("Expected category 34, got %d", decoded.Category())
	}

	if decoded.RecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", decoded.RecordCount())
	}

	// Verify decoded data
	decodedRecord := decoded.Records()[0]

	// Check Data Source Identifier
	if item, exists := decodedRecord.GetDataItem("I034/010"); exists {
		decodedDSI := item.(*common.DataSourceIdentifier)
		if decodedDSI.SAC != 100 || decodedDSI.SIC != 1 {
			t.Errorf("DSI mismatch: got SAC=%d SIC=%d, want SAC=100 SIC=1",
				decodedDSI.SAC, decodedDSI.SIC)
		}
	} else {
		t.Error("I034/010 not found in decoded record")
	}

	// Check Message Type
	if item, exists := decodedRecord.GetDataItem("I034/000"); exists {
		decodedMsgType := item.(*v129.MessageType)
		if decodedMsgType.MessageType != 1 {
			t.Errorf("Message type mismatch: got %d, want 1", decodedMsgType.MessageType)
		}
	} else {
		t.Error("I034/000 not found in decoded record")
	}

	// Check Sector Number
	if item, exists := decodedRecord.GetDataItem("I034/020"); exists {
		decodedSector := item.(*v129.SectorNumber)
		// Allow small rounding error due to 360/256 resolution
		diff := decodedSector.SectorNumber - 45.0
		if diff < -1.5 || diff > 1.5 {
			t.Errorf("Sector number mismatch: got %.2f°, want ~45°", decodedSector.SectorNumber)
		}
	} else {
		t.Error("I034/020 not found in decoded record")
	}
}
