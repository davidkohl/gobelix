// cat/cat062/i380_encoder_test.go
package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestI380EncoderWithCommsStat tests that I062/380 encoder correctly handles CommunicationsSTAT
func TestI380EncoderWithCommsStat(t *testing.T) {
	// Create I062/380 with same fields fusion sets
	addr := uint32(0x7380C7)
	callsign := "ELY318"
	stat := uint8(1) // No alert, no SPI, aircraft on ground
	ecat := uint8(10)

	item := &v120.AircraftDerivedData{
		TargetAddress:        &addr,
		TargetIdentification: &callsign,
		CommunicationsSTAT:   &stat,
		EmitterCategory:      &ecat,
	}

	// Encode
	var buf bytes.Buffer
	bytesWritten, err := item.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Bytes written: %d", bytesWritten)
	t.Logf("Encoded hex: %s", hex.EncodeToString(buf.Bytes()))

	// Decode it back
	item2 := &v120.AircraftDerivedData{}
	buf2 := bytes.NewBuffer(buf.Bytes())
	bytesRead, err := item2.Decode(buf2)
	if err != nil {
		t.Fatalf("Decode failed: %v (read %d bytes)", err, bytesRead)
	}

	t.Logf("Bytes read: %d", bytesRead)

	// Verify roundtrip
	if bytesRead != bytesWritten {
		t.Errorf("Bytes mismatch: wrote %d, read %d", bytesWritten, bytesRead)
	}

	if item2.TargetAddress == nil || *item2.TargetAddress != addr {
		t.Errorf("TargetAddress mismatch: got %v, want 0x%06X", item2.TargetAddress, addr)
	}

	if item2.TargetIdentification == nil || *item2.TargetIdentification != callsign {
		t.Errorf("TargetIdentification mismatch: got %v, want %s", item2.TargetIdentification, callsign)
	}

	if item2.CommunicationsSTAT == nil || *item2.CommunicationsSTAT != stat {
		t.Errorf("CommunicationsSTAT mismatch: got %v, want %d", item2.CommunicationsSTAT, stat)
	}

	if item2.EmitterCategory == nil || *item2.EmitterCategory != ecat {
		t.Errorf("EmitterCategory mismatch: got %v, want %d", item2.EmitterCategory, ecat)
	}

	t.Logf("✅ Roundtrip successful!")
}
