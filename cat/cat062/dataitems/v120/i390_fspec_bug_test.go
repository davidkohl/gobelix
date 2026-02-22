package v120_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestI390FSPECBug reproduces the bug where I062/390 writes a non-zero FSPEC
// with spare bits set when no subfields are present.
//
// EXPECTED: If no subfields are present, FSPEC should be 0x00
// ACTUAL: FSPEC is 0x0130 (spare bits set in byte 2)
func TestI390FSPECBug(t *testing.T) {
	// Create empty I062/390
	i390 := &v120.FlightPlanRelatedData{}

	// Encode
	var buf bytes.Buffer
	n, err := i390.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	encoded := buf.Bytes()
	t.Logf("Encoded %d bytes: %s", n, hex.EncodeToString(encoded))

	if len(encoded) == 0 {
		t.Fatal("Encoded 0 bytes - should have at least FSPEC")
	}

	if len(encoded) == 1 {
		fspec := encoded[0]
		if fspec != 0x00 {
			t.Errorf("BUG: FSPEC should be 0x00 when all fields are nil, got 0x%02x", fspec)
		} else {
			t.Logf("✅ Correct: empty I062/390 encodes as single 0x00 byte")
		}
	} else {
		t.Errorf("BUG: Empty I062/390 encoded as %d bytes (%s), should be 1 byte (0x00)", len(encoded), hex.EncodeToString(encoded))
		if len(encoded) >= 2 {
			t.Errorf("  Byte 1: 0x%02x = 0b%08b", encoded[0], encoded[0])
			t.Errorf("  Byte 2: 0x%02x = 0b%08b (should not exist!)", encoded[1], encoded[1])
		}
	}
}

// TestI390WithNilFields tests roundtrip with all fields nil
func TestI390WithNilFields(t *testing.T) {
	i390 := &v120.FlightPlanRelatedData{
		// All fields nil
	}

	var buf bytes.Buffer
	n, err := i390.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded %d bytes with all fields nil", n)

	if n != 1 {
		t.Errorf("Expected 1 byte, got %d bytes", n)
	}

	encoded := buf.Bytes()
	if len(encoded) > 0 && encoded[0] != 0x00 {
		t.Errorf("BUG: FSPEC should be 0x00 when all fields are nil, got 0x%02x", encoded[0])
	}

	// Decode it back
	decoded := &v120.FlightPlanRelatedData{}
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
