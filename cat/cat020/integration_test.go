// cat/cat020/integration_test.go
package cat020

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat020/dataitems/v10"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestEdition10_FullEncodeDecode(t *testing.T) {
	// Create UAP for Edition 1.0
	uap, err := NewUAP(Version10)
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Create a data block
	dataBlock, err := asterix.NewDataBlock(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("Failed to create data block: %v", err)
	}

	// Create a record with typical MLT target report data
	record, err := asterix.NewRecord(asterix.Cat020, uap)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Set mandatory fields

	// I020/010 - Data Source Identifier
	record.SetDataItem("I020/010", &common.DataSourceIdentifier{
		SAC: 100,
		SIC: 50,
	})

	// I020/020 - Target Report Descriptor
	record.SetDataItem("I020/020", &v10.TargetReportDescriptor{
		SSR: true,
		MS:  true,
		SPI: false,
		RAB: false,
	})

	// I020/140 - Time of Day
	record.SetDataItem("I020/140", &common.TimeOfDay{
		TimeOfDay: 43200.5, // 12:00:00.5
	})

	// I020/041 - Position in WGS-84
	record.SetDataItem("I020/041", &v10.PositionWGS84{
		Latitude:  48.856614,  // Paris
		Longitude: 2.352222,
	})

	// Optional fields

	// I020/161 - Track Number
	record.SetDataItem("I020/161", &v10.TrackNumber{
		TrackNumber: 1234,
	})

	// I020/070 - Mode-3/A Code
	record.SetDataItem("I020/070", &v10.Mode3ACode{
		V:    false,
		G:    false,
		L:    false,
		Code: 01234, // Octal
	})

	// I020/220 - Target Address
	record.SetDataItem("I020/220", &v10.TargetAddress{
		Address: 0x3C6789,
	})

	// Add record to data block
	err = dataBlock.AddRecord(record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// Encode
	encoder := asterix.NewEncoder()
	encoded, err := encoder.Encode(dataBlock)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	t.Logf("Encoded %d bytes: % X", len(encoded), encoded)

	// Decode
	decoder := asterix.NewDecoder(asterix.WithPreloadedUAPs(uap))
	decoded, err := decoder.Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify
	if decoded.Category() != asterix.Cat020 {
		t.Errorf("Category mismatch: got %d, want %d", decoded.Category(), asterix.Cat020)
	}

	if decoded.RecordCount() != 1 {
		t.Errorf("Record count mismatch: got %d, want 1", decoded.RecordCount())
	}

	decodedRecord := decoded.Records()[0]

	// Check Data Source Identifier
	if item, exists := decodedRecord.GetDataItem("I020/010"); exists {
		dsi := item.(*common.DataSourceIdentifier)
		if dsi.SAC != 100 || dsi.SIC != 50 {
			t.Errorf("DSI mismatch: got SAC=%d SIC=%d, want SAC=100 SIC=50", dsi.SAC, dsi.SIC)
		}
	} else {
		t.Error("I020/010 not found in decoded record")
	}

	// Check Position WGS-84
	if item, exists := decodedRecord.GetDataItem("I020/041"); exists {
		pos := item.(*v10.PositionWGS84)
		tolerance := 0.00001
		if abs(pos.Latitude-48.856614) > tolerance || abs(pos.Longitude-2.352222) > tolerance {
			t.Errorf("Position mismatch: got Lat=%.6f Lon=%.6f, want Lat=48.856614 Lon=2.352222",
				pos.Latitude, pos.Longitude)
		}
	} else {
		t.Error("I020/041 not found in decoded record")
	}

	// Check Track Number
	if item, exists := decodedRecord.GetDataItem("I020/161"); exists {
		tn := item.(*v10.TrackNumber)
		if tn.TrackNumber != 1234 {
			t.Errorf("Track number mismatch: got %d, want 1234", tn.TrackNumber)
		}
	} else {
		t.Error("I020/161 not found in decoded record")
	}

	// Check Mode-3/A Code
	if item, exists := decodedRecord.GetDataItem("I020/070"); exists {
		mode3a := item.(*v10.Mode3ACode)
		if mode3a.Code != 01234 {
			t.Errorf("Mode-3/A code mismatch: got %04o, want 01234", mode3a.Code)
		}
	} else {
		t.Error("I020/070 not found in decoded record")
	}

	t.Log("Edition 1.0 full encode/decode test passed successfully")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
