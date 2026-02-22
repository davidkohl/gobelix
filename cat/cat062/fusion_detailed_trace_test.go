// cat/cat062/fusion_detailed_trace_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"
)

// TestFusionMessageTrace manually traces through a failing message byte by byte
func TestFusionMessageTrace(t *testing.T) {
	// Real message that fails with I062/500 error at byte 73
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, _ := hex.DecodeString(hexData)

	t.Logf("Message length: %d bytes", len(data))
	t.Logf("Header: CAT=%d, Length=%d", data[0], (uint16(data[1])<<8)|uint16(data[2]))

	// FSPEC: BFFFED06
	// Byte 1: BF = 10111111 → FRN 1,3,4,5,6,7 + FX
	// Byte 2: FF = 11111111 → FRN 8,9,10,11,12,13,14 + FX
	// Byte 3: ED = 11101101 → FRN 15,16,17,19,20,22,23 + FX
	// Byte 4: 06 = 00000110 → FRN 27,28 (no FX)

	offset := 3 // After CAT+LEN

	// FSPEC
	fspecLen := 4
	t.Logf("\n[%d] FSPEC: %s (%d bytes)", offset, hex.EncodeToString(data[offset:offset+fspecLen]), fspecLen)
	offset += fspecLen

	// Data items in order:
	// FRN 1: I062/010 - Data Source ID (2 bytes fixed)
	t.Logf("[%d] I062/010 (FRN 1): %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 3: I062/015 - Service ID (1 byte fixed)
	t.Logf("[%d] I062/015 (FRN 3): %02X (1 byte)", offset, data[offset])
	offset += 1

	// FRN 4: I062/070 - Time (3 bytes fixed)
	t.Logf("[%d] I062/070 (FRN 4): %s (3 bytes)", offset, hex.EncodeToString(data[offset:offset+3]))
	offset += 3

	// FRN 5: I062/105 - WGS-84 Position (8 bytes fixed)
	t.Logf("[%d] I062/105 (FRN 5): %s (8 bytes)", offset, hex.EncodeToString(data[offset:offset+8]))
	offset += 8

	// FRN 6: I062/100 - Cartesian Position (6 bytes fixed)
	t.Logf("[%d] I062/100 (FRN 6): %s (6 bytes)", offset, hex.EncodeToString(data[offset:offset+6]))
	offset += 6

	// FRN 7: I062/185 - Velocity (4 bytes fixed)
	t.Logf("[%d] I062/185 (FRN 7): %s (4 bytes)", offset, hex.EncodeToString(data[offset:offset+4]))
	offset += 4

	// FRN 8: I062/210 - Acceleration (2 bytes fixed)
	t.Logf("[%d] I062/210 (FRN 8): %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 9: I062/060 - Mode 3/A (2 bytes fixed)
	t.Logf("[%d] I062/060 (FRN 9): %s (2 bytes)", offset, hex.EncodeToString(data[offset:offset+2]))
	offset += 2

	// FRN 10: I062/245 - Target ID (7 bytes fixed)
	t.Logf("[%d] I062/245 (FRN 10): %s (7 bytes)", offset, hex.EncodeToString(data[offset:offset+7]))
	offset += 7

	// FRN 11: I062/380 - Aircraft Derived Data (COMPOUND - variable)
	t.Logf("[%d] I062/380 (FRN 11) - COMPOUND:", offset)
	i380_fspec := data[offset]
	t.Logf("  FSPEC: %02X (%08b)", i380_fspec, i380_fspec)
	_ = 1 // i380_len
	_ = offset // offset_before_380

	// Count how many subfields in I062/380
	// Bit 8 = ADR (24 bits = 3 bytes)
	// Bit 7 = ID (48 bits = 6 bytes)
	// ... etc
	// For now, let's just scan until we hit something that looks like the next item

	// Actually, let's check if FX is set
	if i380_fspec&0x01 != 0 {
		t.Logf("  I062/380 has FX bit - has extension")
	}

	// Try to figure out the length by looking ahead
	// Next item should be I062/040 (FRN 12) which is 2 bytes and should have track number
	// Let's search for a pattern

	// Skip for now and manually calculate from the error
	t.Logf("  [Skipping detailed I062/380 decode - will calculate from total]")

	// Since we know the error occurs at byte 73 for I062/500,
	// and we're currently at offset %d, let's see how much is left
	t.Logf("\nCurrent offset: %d", offset)
	t.Logf("Error will occur at byte 73 trying to decode I062/500")
	t.Logf("That means %d bytes consumed by items FRN 11-26", 73-offset)

	// Let's jump to where gobelix THINKS I062/500 is
	t.Logf("\n=== What gobelix sees at byte 73 ===")
	t.Logf("Bytes 73-82: %s", hex.EncodeToString(data[73:83]))
	t.Logf("First byte (FSPEC): 0x%02X (%08b)", data[73], data[73])

	// And what we know is the ACTUAL I062/500 location by decoding in isolation
	t.Logf("\n=== The ACTUAL I062/500 (which decodes successfully) ===")
	t.Logf("Should be somewhere near byte 73, but offset is wrong")

	// The key is: which data item between offset %d and byte 73 has wrong length?
	t.Logf("\n=== Remaining data items to analyze ===")
	t.Logf("FRN 11: I062/380 - Aircraft Derived Data (compound)")
	t.Logf("FRN 12: I062/040 - Track Number (2 bytes)")
	t.Logf("FRN 13: I062/080 - Track Status (extended)")
	t.Logf("FRN 14: I062/290 - System Track Update Ages (compound)")
	t.Logf("FRN 15: I062/200 - Mode of Movement (1 byte)")
	t.Logf("FRN 16: I062/295 - Track Data Ages (compound)")
	t.Logf("FRN 17: I062/136 - Measured FL (2 bytes)")
	t.Logf("FRN 19: I062/135 - Barometric Alt (2 bytes)")
	t.Logf("FRN 20: I062/220 - Rate of C/D (2 bytes)")
	t.Logf("FRN 22: I062/270 - Target Size (extended)")
	t.Logf("FRN 23: I062/300 - Vehicle Fleet (1 byte)")
	t.Logf("FRN 27: I062/500 - Estimated Accuracies (compound) ← ERROR HERE")

	t.Logf("\nOne of the compound items (380, 290, 295, or extended 080/270) has wrong length!")
}
