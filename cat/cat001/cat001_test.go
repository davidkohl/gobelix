// cat/cat001/cat001_test.go
package cat001_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat001"
	v12 "github.com/davidkohl/gobelix/cat/cat001/dataitems/v12"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestCat001UAP(t *testing.T) {
	uap, err := cat001.NewUAP(cat001.Version12)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	if uap.Category() != asterix.Cat001 {
		t.Errorf("Expected category 1, got %d", uap.Category())
	}

	if uap.Version() != "1.2" {
		t.Errorf("Expected version 1.2, got %s", uap.Version())
	}
}

func TestCat001EncodeDecode(t *testing.T) {
	uap, err := cat001.NewUAP(cat001.Version12)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	record, err := asterix.NewRecord(asterix.Cat001, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Add data items
	record.SetDataItem("I001/010", &common.DataSourceIdentifier{SAC: 50, SIC: 2})
	record.SetDataItem("I001/020", &v12.TargetReportDescriptor{TYP: 1, SSR: true})
	record.SetDataItem("I001/040", &v12.PositionPolar{RHO: 50.5, THETA: 90.0})
	record.SetDataItem("I001/070", &v12.Mode3ACode{V: true, Mode: 0o7777})

	dataBlock, err := asterix.NewDataBlock(asterix.Cat001, uap)
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

	if decoded.Category() != asterix.Cat001 {
		t.Errorf("Expected category 1, got %d", decoded.Category())
	}
}
