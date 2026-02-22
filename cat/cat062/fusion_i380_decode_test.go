// cat/cat062/fusion_i380_decode_test.go
package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestDecodeI380FromFusionMessage extracts and decodes I062/380 from real fusion message
func TestDecodeI380FromFusionMessage(t *testing.T) {
	// Real message from fusion
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	// I062/380 starts at offset 42 (confirmed by manual decode)
	// I062/380 bytes (31 total): c10101007380c714c673c788200ef603010130b0020202400006170617fffd
	i380Offset := 42
	i380Bytes := data[i380Offset : i380Offset+31]

	t.Logf("I062/380 hex: %s", hex.EncodeToString(i380Bytes))
	t.Logf("I062/380 length: %d bytes", len(i380Bytes))

	// Manually decode FSPEC
	t.Logf("\n=== FSPEC Analysis ===")
	t.Logf("Byte 1: 0x%02X (%08b)", i380Bytes[0], i380Bytes[0])
	t.Logf("Byte 2: 0x%02X (%08b)", i380Bytes[1], i380Bytes[1])
	t.Logf("Byte 3: 0x%02X (%08b)", i380Bytes[2], i380Bytes[2])
	t.Logf("Byte 4: 0x%02X (%08b)", i380Bytes[3], i380Bytes[3])

	// Decode with gobelix
	t.Logf("\n=== Decoding with gobelix ===")
	item := &v120.AircraftDerivedData{}
	buf := bytes.NewBuffer(i380Bytes)

	bytesRead, err := item.Decode(buf)
	if err != nil {
		t.Logf("❌ Decode error: %v", err)
		t.Logf("Bytes read before error: %d", bytesRead)
	} else {
		t.Logf("✅ Decode succeeded! Read %d bytes", bytesRead)

		if item.TargetAddress != nil {
			t.Logf("Target Address: 0x%06X", *item.TargetAddress)
		}
		if item.TargetIdentification != nil {
			t.Logf("Target Identification: %s", *item.TargetIdentification)
		}
		if item.MagneticHeading != nil {
			t.Logf("Magnetic Heading: %.2f°", *item.MagneticHeading)
		}
		if item.GroundSpeed != nil {
			t.Logf("Ground Speed: %.1f kt", *item.GroundSpeed)
		}
	}

	// Check if all 31 bytes were consumed
	if bytesRead != 31 {
		t.Errorf("Expected to consume 31 bytes, but consumed %d bytes", bytesRead)
		t.Logf("Remaining bytes in buffer: %d", buf.Len())
		if buf.Len() > 0 {
			remaining := make([]byte, buf.Len())
			buf.Read(remaining)
			t.Logf("Remaining bytes: %s", hex.EncodeToString(remaining))
		}
	}

	// Manual subfield decode based on FSPEC
	t.Logf("\n=== Manual Subfield Decode ===")

	// FSPEC: C1 01 01 00
	// Byte 1 (C1 = 11000001): bits 7,6 set + FX
	// → Subfield #1 (Target Address) and Subfield #2 (Target ID) present

	offset := 4 // After FSPEC

	// Subfield #1: Target Address (3 bytes)
	if (i380Bytes[0] & 0x80) != 0 {
		addr := uint32(i380Bytes[offset])<<16 | uint32(i380Bytes[offset+1])<<8 | uint32(i380Bytes[offset+2])
		t.Logf("Subfield #1 (Target Address): 0x%06X (%d bytes)", addr, 3)
		offset += 3
	}

	// Subfield #2: Target Identification (6 bytes)
	if (i380Bytes[0] & 0x40) != 0 {
		idBytes := i380Bytes[offset : offset+6]
		t.Logf("Subfield #2 (Target ID): %s (%d bytes)", hex.EncodeToString(idBytes), 6)
		offset += 6
	}

	t.Logf("\nBytes consumed after subfields #1 and #2: %d", offset)
	t.Logf("Expected total: 31 bytes")
	t.Logf("Discrepancy: %d bytes", 31-offset)

	if offset < 31 {
		t.Logf("\nRemaining bytes after subfield #2:")
		t.Logf("%s", hex.EncodeToString(i380Bytes[offset:]))

		// These might be additional subfields that shouldn't be there
		// OR the FSPEC is wrong
		t.Logf("\nHypothesis: The FSPEC 'C1 01 01 00' might be WRONG.")
		t.Logf("The actual FSPEC might be longer or different.")
		t.Logf("Let's check if byte 4 is actually part of subfield data, not FSPEC...")
	}
}
