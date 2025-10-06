// cat/cat002/cat002_test.go
package cat002_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat002"
	v10 "github.com/davidkohl/gobelix/cat/cat002/dataitems/v10"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestCat002UAP(t *testing.T) {
	uap, err := cat002.NewUAP(cat002.Version10)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	if uap.Category() != asterix.Cat002 {
		t.Errorf("Expected category 2, got %d", uap.Category())
	}

	if uap.Version() != "1.0" {
		t.Errorf("Expected version 1.0, got %s", uap.Version())
	}
}

func TestCat002EncodeDecode(t *testing.T) {
	uap, err := cat002.NewUAP(cat002.Version10)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	record, err := asterix.NewRecord(asterix.Cat002, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Add data items
	record.SetDataItem("I002/010", &common.DataSourceIdentifier{SAC: 25, SIC: 5})
	record.SetDataItem("I002/000", &v10.MessageType{MessageType: 1}) // North marker
	record.SetDataItem("I002/020", &v10.SectorNumber{SectorNumber: 90.0})
	record.SetDataItem("I002/030", &common.TimeOfDay{TimeOfDay: 43200.0}) // 12:00:00
	record.SetDataItem("I002/041", &v10.AntennaRotationSpeed{RotationPeriod: 4.0})

	dataBlock, err := asterix.NewDataBlock(asterix.Cat002, uap)
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

	// Decode
	decoder := asterix.NewDecoder(asterix.WithPreloadedUAPs(uap))
	decoded, err := decoder.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.Category() != asterix.Cat002 {
		t.Errorf("Expected category 2, got %d", decoded.Category())
	}
}
