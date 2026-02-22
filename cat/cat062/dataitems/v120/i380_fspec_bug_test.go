package v120_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestI380FSPECBug reproduces the bug where I062/380 writes a non-zero FSPEC
// but no actual subfield data.
//
// EXPECTED: If no subfields are present, FSPEC should be 0x00
// ACTUAL: FSPEC is non-zero (e.g., 0x60) even when no data is written
func TestI380FSPECBug(t *testing.T) {
	// Create empty I062/380
	i380 := &v120.AircraftDerivedData{}

	// Encode
	var buf bytes.Buffer
	n, err := i380.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	encoded := buf.Bytes()
	t.Logf("Encoded %d bytes: %s", n, hex.EncodeToString(encoded))

	if len(encoded) == 0 {
		t.Fatal("Encoded 0 bytes - should have at least FSPEC")
	}

	fspec := encoded[0]
	t.Logf("FSPEC: 0x%02x = 0b%08b", fspec, fspec)

	// Check: if FSPEC has bits set, there should be corresponding data
	if fspec != 0x00 {
		if len(encoded) == 1 {
			t.Errorf("BUG: FSPEC 0x%02x has bits set, but no subfield data encoded!", fspec)
			t.Errorf("This will cause the decoder to read from wrong positions!")
		} else {
			t.Logf("FSPEC has bits set and %d data bytes present", len(encoded)-1)
		}
	}
}

// TestI380WithNilFields tests what happens when we create I062/380
// with all fields as nil (which is what happens in the failing message).
func TestI380WithNilFields(t *testing.T) {
	i380 := &v120.AircraftDerivedData{
		// All fields nil
	}

	var buf bytes.Buffer
	n, err := i380.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded %d bytes with all fields nil", n)

	if n == 0 {
		t.Error("Encoded 0 bytes - should at least encode FSPEC 0x00")
		return
	}

	encoded := buf.Bytes()
	fspec := encoded[0]

	if fspec != 0x00 {
		t.Errorf("BUG: FSPEC should be 0x00 when all fields are nil, got 0x%02x", fspec)
	}

	// Decode it back
	decoded := &v120.AircraftDerivedData{}
	decBuf := bytes.NewBuffer(encoded)
	nDec, err := decoded.Decode(decBuf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	t.Logf("Decoded %d bytes successfully", nDec)

	if nDec != n {
		t.Errorf("Byte count mismatch: encoded %d, decoded %d", n, nDec)
	}
}
