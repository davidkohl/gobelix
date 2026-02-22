// cat/cat062/minimal_encode_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// TestMinimalCAT062Encode tests that a minimal CAT062 message encodes correctly
func TestMinimalCAT062Encode(t *testing.T) {
	// Create UAP
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	// Build items map (minimal set - exactly what fusion encodes)
	items := make(map[string]asterix.DataItem)

	// I062/010: Data Source Identifier (MANDATORY)
	items["I062/010"] = &common.DataSourceIdentifier{
		SAC: 10,
		SIC: 1,
	}

	// I062/015: Service Identification
	items["I062/015"] = &common.ServiceIdentification{
		Value: 1,
	}

	// I062/040: Track Number (MANDATORY)
	items["I062/040"] = &cat062.TrackNumber{
		Value: 12345,
	}

	// I062/070: Time of Track Information (MANDATORY)
	items["I062/070"] = &cat062.TimeOfTrackInformation{
		Time: 30000.0, // 30 seconds
	}

	// I062/080: Track Status (MANDATORY)
	ts := &cat062.TrackStatus{
		MON: false,
		SPI: false,
		MRH: false,
		SRC: 3, // Fused track
		CNF: false,
	}
	ts.SetHasExtension()
	items["I062/080"] = ts

	// I062/105: WGS-84 Position
	items["I062/105"] = &cat062.CalculatedPositionWGS84{
		Latitude:  51.5,
		Longitude: 10.5,
	}

	// Encode using encoder
	encoder := asterix.NewEncoder()
	encoded, err := encoder.EncodeItems(asterix.Cat062, uap062, items)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded %d bytes: %s", len(encoded), hex.EncodeToString(encoded))

	// Decode it back
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(dataBlocks) != 1 {
		t.Fatalf("Expected 1 data block, got %d", len(dataBlocks))
	}

	db := dataBlocks[0]
	if db.RecordCount() != 1 {
		t.Fatalf("Expected 1 record, got %d", db.RecordCount())
	}

	rec := db.Records()[0]
	decodedItems := rec.Items()

	// Check I062/010 was decoded
	dsi, ok := decodedItems["I062/010"].(*common.DataSourceIdentifier)
	if !ok {
		t.Fatalf("I062/010 not found in decoded items! Items: %v", decodedItems)
	}

	if dsi.SAC != 10 || dsi.SIC != 1 {
		t.Errorf("I062/010 mismatch: got SAC=%d SIC=%d, want SAC=10 SIC=1", dsi.SAC, dsi.SIC)
	}

	t.Logf("✅ Minimal CAT062 encode/decode roundtrip successful!")
}
