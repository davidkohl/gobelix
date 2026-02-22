// cat/cat062/fusion_offset_hunt_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionOffsetHunt systematically scans the message to find where the real I062/500 is
func TestFusionOffsetHunt(t *testing.T) {
	// Real message from fusion
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	t.Logf("Message length: %d bytes", len(data))
	t.Logf("Full hex: %s\n", hexData)

	// FSPEC: BFFFED06
	// FRN 27 (I062/500) is SET

	// Expected data items in order (from FSPEC):
	// FRN 1:  I062/010 (2 bytes)
	// FRN 3:  I062/015 (1 byte)
	// FRN 4:  I062/070 (3 bytes)
	// FRN 5:  I062/105 (8 bytes)
	// FRN 6:  I062/100 (6 bytes)
	// FRN 7:  I062/185 (4 bytes)
	// FRN 8:  I062/210 (2 bytes)
	// FRN 9:  I062/060 (2 bytes)
	// FRN 10: I062/245 (7 bytes)
	// FRN 11: I062/380 (COMPOUND - variable) ← LIKELY CULPRIT
	// FRN 12: I062/040 (2 bytes)
	// FRN 13: I062/080 (extended)
	// FRN 14: I062/290 (COMPOUND - variable)
	// FRN 15: I062/200 (1 byte)
	// FRN 16: I062/295 (COMPOUND - variable)
	// FRN 17: I062/136 (2 bytes)
	// FRN 19: I062/135 (2 bytes)
	// FRN 20: I062/220 (2 bytes)
	// FRN 22: I062/270 (extended)
	// FRN 23: I062/300 (1 byte)
	// FRN 27: I062/500 (COMPOUND - variable)
	// FRN 28: I062/340 (COMPOUND - variable)

	offset := 3 // After CAT+LEN
	t.Logf("\n=== Manual decode ===")

	// FSPEC
	t.Logf("[%d] FSPEC: %s (4 bytes)", offset, hex.EncodeToString(data[offset:offset+4]))
	offset += 4

	// FRN 1: I062/010 (2 bytes)
	t.Logf("[%d] I062/010: %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 3: I062/015 (1 byte)
	t.Logf("[%d] I062/015: %02X (1 byte)", offset, data[offset])
	offset += 1

	// FRN 4: I062/070 (3 bytes)
	t.Logf("[%d] I062/070: %s (3 bytes)", offset, hex.EncodeToString(data[offset:offset+3]))
	offset += 3

	// FRN 5: I062/105 (8 bytes)
	t.Logf("[%d] I062/105: %s (8 bytes)", offset, hex.EncodeToString(data[offset:offset+8]))
	offset += 8

	// FRN 6: I062/100 (6 bytes)
	t.Logf("[%d] I062/100: %s (6 bytes)", offset, hex.EncodeToString(data[offset:offset+6]))
	offset += 6

	// FRN 7: I062/185 (4 bytes)
	t.Logf("[%d] I062/185: %s (4 bytes)", offset, hex.EncodeToString(data[offset:offset+4]))
	offset += 4

	// FRN 8: I062/210 (2 bytes)
	t.Logf("[%d] I062/210: %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 9: I062/060 (2 bytes)
	t.Logf("[%d] I062/060: %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 10: I062/245 (7 bytes)
	t.Logf("[%d] I062/245: %s (7 bytes)", offset, hex.EncodeToString(data[offset:offset+7]))
	offset += 7

	// FRN 11: I062/380 - COMPOUND (variable length!)
	// This is where the bug likely is
	t.Logf("\n[%d] I062/380 (COMPOUND - Aircraft Derived Data):", offset)
	i380_start := offset
	i380_fspec1 := data[offset]
	t.Logf("  FSPEC byte 1: 0x%02X (%08b)", i380_fspec1, i380_fspec1)
	offset++

	// Check if FX bit is set (bit 0)
	hasFX1 := (i380_fspec1 & 0x01) != 0
	t.Logf("  FX bit (byte 1): %v", hasFX1)

	if hasFX1 {
		i380_fspec2 := data[offset]
		t.Logf("  FSPEC byte 2: 0x%02X (%08b)", i380_fspec2, i380_fspec2)
		offset++

		hasFX2 := (i380_fspec2 & 0x01) != 0
		t.Logf("  FX bit (byte 2): %v", hasFX2)

		if hasFX2 {
			i380_fspec3 := data[offset]
			t.Logf("  FSPEC byte 3: 0x%02X (%08b)", i380_fspec3, i380_fspec3)
			offset++

			hasFX3 := (i380_fspec3 & 0x01) != 0
			t.Logf("  FX bit (byte 3): %v", hasFX3)

			if hasFX3 {
				i380_fspec4 := data[offset]
				t.Logf("  FSPEC byte 4: 0x%02X (%08b)", i380_fspec4, i380_fspec4)
				offset++
			}
		}
	}

	// Now we need to count how many subfields are present
	// Count bits set in FSPEC (excluding FX bits)
	// For now, let's try using the decoder

	t.Logf("\n=== Attempt to decode I062/380 with gobelix ===")

	// Create a dummy decoder for I062/380
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	t.Log("Scanning for I062/380 subfields...")

	// Instead of decoding, let's use gobelix full decoder and see WHERE it fails
	t.Logf("\n=== Full gobelix decode (to see exact failure point) ===")
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	_, err = decoder.DecodeAll(data)
	if err != nil {
		t.Logf("Decode error: %v", err)
		t.Logf("Error says I062/500 at byte 73, but manual analysis shows I062/380 starts at byte %d", i380_start)

		// Calculate expected I062/500 offset if all previous items are correct
		t.Logf("\nIf gobelix THINKS I062/500 is at byte 73, but it's actually somewhere else...")
		t.Logf("The difference is: %d - %d = %d bytes", 73, i380_start, 73-i380_start)

		t.Logf("\nLet's manually scan for the I062/500 pattern...")
		// I062/500 should start with FSPEC that has bits for APC/ATV
		// We know from the isolation test that it decodes at offset 73 successfully
		// So the actual I062/500 IS at offset 73, but gobelix is counting wrong before that

		t.Logf("\nByte 73 onward (confirmed I062/500):")
		t.Logf("  %s", hex.EncodeToString(data[73:82]))
		t.Logf("  Decodes successfully as: APCX=370, APCY=369, COV=-9020, VelocityX=183, VelocityY=183")

		t.Logf("\nConclusion: I062/500 IS at byte 73.")
		t.Logf("Problem: gobelix hasn't consumed %d bytes yet when it reaches byte 73.", 73-i380_start)
		t.Logf("This means something between byte %d and byte 73 is being decoded incorrectly.", i380_start)
	}

	// Calculate bytes consumed by fixed items before I062/380
	fixedBeforeI380 := 2 + 1 + 3 + 8 + 6 + 4 + 2 + 2 + 7
	t.Logf("\nFixed items before I062/380: %d bytes", fixedBeforeI380)
	t.Logf("I062/380 starts at offset: %d (CAT+LEN+FSPEC = 3+4 = 7, plus %d fixed items)", i380_start, fixedBeforeI380)
	t.Logf("Expected byte 73 is at: 7 + %d + I062/380_length", fixedBeforeI380)
	t.Logf("So I062/380 should be: 73 - 7 - %d = %d bytes", fixedBeforeI380, 73-7-fixedBeforeI380)

	expectedI380Len := 73 - 7 - fixedBeforeI380
	t.Logf("\nExpected I062/380 length: %d bytes", expectedI380Len)
	t.Logf("I062/380 bytes: %s", hex.EncodeToString(data[i380_start:i380_start+expectedI380Len]))
}
