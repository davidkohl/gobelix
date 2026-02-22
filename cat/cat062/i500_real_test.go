package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestI500RealFailure tests with the actual I062/500 data that's causing failures
// Error message: "decoding data item: decoding APW: buffer too short for APW"
// This suggests the encoder is writing fewer bytes than it claims
func TestI500RealFailure(t *testing.T) {
	// Create I062/500 with APC and ATV (no APW)
	// This is what fusion creates based on createEstimatedAccuracies
	i500 := &cat062.EstimatedAccuracies{}
	
	// APC: Position accuracy (APCX, APCY)
	apcX := uint16(100) // 50m std dev (100 * 0.5m LSB)
	apcY := uint16(120) // 60m std dev
	i500.APCX = &apcX
	i500.APCY = &apcY
	
	// ATV: Velocity accuracy (VelocityX, VelocityY)
	vx := uint8(20) // 5 m/s std dev (20 * 0.25 m/s LSB)
	vy := uint8(24) // 6 m/s std dev
	i500.VelocityX = &vx
	i500.VelocityY = &vy
	
	// Encode
	var encBuf bytes.Buffer
	n, err := i500.Encode(&encBuf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	encoded := encBuf.Bytes()
	t.Logf("Encoded %d bytes: %s", n, hex.EncodeToString(encoded))
	t.Logf("Expected FSPEC: 0x84 (APC=bit8, ATV=bit3)")
	t.Logf("Expected length: 1 (FSPEC) + 4 (APC) + 2 (ATV) = 7 bytes")
	
	// Check byte count matches
	if n != len(encoded) {
		t.Errorf("CRITICAL: Encode() returned n=%d but wrote %d bytes", n, len(encoded))
	}
	
	// Decode it back
	decBuf := bytes.NewBuffer(encoded)
	var decoded cat062.EstimatedAccuracies
	nDec, err := decoded.Decode(decBuf)
	if err != nil {
		t.Fatalf("Decode failed: %v (read %d bytes)", err, nDec)
	}
	
	t.Logf("Decode succeeded: read %d bytes", nDec)
	
	// Verify roundtrip
	if *decoded.APCX != *i500.APCX {
		t.Errorf("APCX mismatch: got %d, want %d", *decoded.APCX, *i500.APCX)
	}
	if *decoded.APCY != *i500.APCY {
		t.Errorf("APCY mismatch: got %d, want %d", *decoded.APCY, *i500.APCY)
	}
	if *decoded.VelocityX != *i500.VelocityX {
		t.Errorf("VelocityX mismatch: got %d, want %d", *decoded.VelocityX, *i500.VelocityX)
	}
	if *decoded.VelocityY != *i500.VelocityY {
		t.Errorf("VelocityY mismatch: got %d, want %d", *decoded.VelocityY, *i500.VelocityY)
	}
}

// TestI500InFullRecord tests I062/500 within a complete CAT062 record
func TestI500InFullRecord(t *testing.T) {
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	
	db, err := asterix.NewDataBlock(asterix.Cat062, uap062)
	if err != nil {
		t.Fatalf("Failed to create data block: %v", err)
	}
	
	// Create minimal record with I062/500
	items := make(map[string]asterix.DataItem)
	
	// I062/040: Track Number
	items["I062/040"] = &cat062.TrackNumber{Value: 12345}
	
	// I062/500: Estimated Accuracies
	i500 := &cat062.EstimatedAccuracies{}
	apcX := uint16(100)
	apcY := uint16(120)
	i500.APCX = &apcX
	i500.APCY = &apcY
	vx := uint8(20)
	vy := uint8(24)
	i500.VelocityX = &vx
	i500.VelocityY = &vy
	items["I062/500"] = i500
	
	// Encode record
	if err := db.EncodeRecord(items); err != nil {
		t.Fatalf("EncodeRecord failed: %v", err)
	}
	
	// Encode datablock
	data, err := db.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	t.Logf("Encoded CAT062 message: %d bytes", len(data))
	t.Logf("Hex: %s", hex.EncodeToString(data))
	
	// Check LENGTH field
	if len(data) >= 3 {
		cat := data[0]
		length := int(data[1])<<8 | int(data[2])
		t.Logf("CAT=%d, LENGTH=%d, actual=%d", cat, length, len(data))
		if length != len(data) {
			t.Errorf("LENGTH MISMATCH: header says %d, actual %d (diff: %d)", length, len(data), len(data)-length)
		}
	}
	
	// Decode it back
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)
	
	dataBlocks, err := decoder.DecodeAll(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	if len(dataBlocks) == 0 {
		t.Fatal("No data blocks decoded")
	}
	
	records := dataBlocks[0].Records()
	if len(records) == 0 {
		t.Fatal("No records decoded")
	}
	
	decodedItems := records[0].Items()
	t.Logf("Decoded %d data items", len(decodedItems))
	
	// Check I062/500
	if i500Dec, ok := decodedItems["I062/500"].(*cat062.EstimatedAccuracies); ok {
		t.Logf("I062/500 decoded: %s", i500Dec.String())
		if *i500Dec.APCX != *i500.APCX {
			t.Errorf("APCX mismatch")
		}
	} else {
		t.Error("I062/500 not found in decoded record")
	}
}
