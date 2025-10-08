// cat/cat020/cat020_test.go
package cat020

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestCat020UAP(t *testing.T) {
	uap, err := NewUAP(Version110)
	if err != nil {
		t.Fatalf("NewUAP() failed: %v", err)
	}

	if uap == nil {
		t.Fatal("UAP is nil")
	}

	if uap.Category() != asterix.Cat020 {
		t.Errorf("Category() = %v, want %v", uap.Category(), asterix.Cat020)
	}

	if uap.Version() != Version110 {
		t.Errorf("Version() = %v, want %v", uap.Version(), Version110)
	}
}

func TestCat020EncodeDecode(t *testing.T) {
	uap, err := NewUAP(Version110)
	if err != nil {
		t.Fatalf("NewUAP() failed: %v", err)
	}

	// Create a data block
	dataBlock, err := asterix.NewDataBlock(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("NewDataBlock() failed: %v", err)
	}

	// Create a record with mandatory fields
	record, err := asterix.NewRecord(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("NewRecord() failed: %v", err)
	}

	// Add mandatory fields: I020/010 (Data Source Identifier) and I020/140 (Time of Day)
	record.SetDataItem("I020/010", &common.DataSourceIdentifier{
		SAC: 5,
		SIC: 10,
	})

	record.SetDataItem("I020/140", &common.TimeOfDay{
		TimeOfDay: 12345.678,
	})

	// Add the record to the data block
	if err := dataBlock.AddRecord(record); err != nil {
		t.Fatalf("AddRecord() failed: %v", err)
	}

	// Encode the data block
	encoded, err := dataBlock.Encode()
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Decode the data block
	decodedBlock, err := asterix.NewDataBlock(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("NewDataBlock() for decode failed: %v", err)
	}

	if err := decodedBlock.Decode(encoded); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	// Verify the decoded data
	if decodedBlock.Category() != asterix.Cat020 {
		t.Errorf("Decoded Category = %v, want %v", decodedBlock.Category(), asterix.Cat020)
	}

	if decodedBlock.RecordCount() != 1 {
		t.Errorf("Decoded RecordCount = %d, want 1", decodedBlock.RecordCount())
	}

	records := decodedBlock.Records()
	if len(records) != 1 {
		t.Fatalf("Decoded records length = %d, want 1", len(records))
	}

	// Check Data Source Identifier
	dsi, exists := records[0].GetDataItem("I020/010")
	if !exists {
		t.Error("I020/010 not found in decoded record")
	} else {
		dsiTyped := dsi.(*common.DataSourceIdentifier)
		if dsiTyped.SAC != 5 || dsiTyped.SIC != 10 {
			t.Errorf("DSI = SAC:%d SIC:%d, want SAC:5 SIC:10", dsiTyped.SAC, dsiTyped.SIC)
		}
	}

	// Check Time of Day (allow for encoding precision loss due to 1/128s resolution)
	tod, exists := records[0].GetDataItem("I020/140")
	if !exists {
		t.Error("I020/140 not found in decoded record")
	} else {
		todTyped := tod.(*common.TimeOfDay)
		diff := todTyped.TimeOfDay - 12345.678
		if diff < -0.01 || diff > 0.01 {
			t.Errorf("TimeOfDay = %f, want ~12345.678 (diff: %f)", todTyped.TimeOfDay, diff)
		}
	}
}
