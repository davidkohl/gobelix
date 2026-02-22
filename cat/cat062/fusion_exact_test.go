package cat062_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// TestFusionExactConfiguration tests the EXACT data items fusion encodes
func TestFusionExactConfiguration(t *testing.T) {
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	
	db, err := asterix.NewDataBlock(asterix.Cat062, uap062)
	if err != nil {
		t.Fatalf("Failed to create datablock: %v", err)
	}
	
	items := make(map[string]asterix.DataItem)
	
	// Minimal fusion configuration
	items["I062/010"] = &common.DataSourceIdentifier{SAC: 6, SIC: 100}
	items["I062/040"] = &cat062.TrackNumber{Value: 12345}
	items["I062/070"] = &cat062.TimeOfTrackInformation{Time: 43377.09}
	items["I062/080"] = &cat062.TrackStatus{}
	items["I062/105"] = &cat062.CalculatedPositionWGS84{Latitude: 51.0, Longitude: 10.0}
	items["I062/500"] = func() *cat062.EstimatedAccuracies {
		apcX := uint16(100)
		apcY := uint16(120)
		return &cat062.EstimatedAccuracies{
			APCX: &apcX,
			APCY: &apcY,
		}
	}()
	
	// Encode
	if err := db.EncodeRecord(items); err != nil {
		t.Fatalf("EncodeRecord failed: %v", err)
	}
	
	data, err := db.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes", len(data))
	
	// Check LENGTH field
	if len(data) >= 3 {
		cat := data[0]
		length := int(data[1])<<8 | int(data[2])
		t.Logf("CAT=%d, LENGTH=%d, actual=%d", cat, length, len(data))
		
		if length != len(data) {
			t.Errorf("❌ LENGTH MISMATCH: header says %d, actual %d (diff=%d)",
				length, len(data), len(data)-length)
			t.Fatalf("FOUND THE BUG!")
		}
	}
	
	// Decode
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)
	
	_, err = decoder.DecodeAll(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	t.Logf("✅ Roundtrip succeeded")
}
