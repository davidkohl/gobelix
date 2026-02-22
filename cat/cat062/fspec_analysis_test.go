// cat/cat062/fspec_analysis_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"
)

// TestFSPECBitMapping analyzes FSPEC bit-to-FRN mapping
func TestFSPECBitMapping(t *testing.T) {
	msg1Hex := "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528"
	data, _ := hex.DecodeString(msg1Hex)

	fspec := data[3:7] // BFFFED06

	t.Logf("FSPEC: %s", hex.EncodeToString(fspec))
	t.Logf("Byte 1: 0x%02X = %08b", fspec[0], fspec[0])
	t.Logf("Byte 2: 0x%02X = %08b", fspec[1], fspec[1])
	t.Logf("Byte 3: 0x%02X = %08b", fspec[2], fspec[2])
	t.Logf("Byte 4: 0x%02X = %08b", fspec[3], fspec[3])

	// Manual FRN mapping (ignoring FX bits)
	// Each FSPEC byte has 7 data bits + 1 FX bit
	//
	// Byte 1: BF = 10111111
	//   Bit 7=1 → FRN 1  (I062/010)
	//   Bit 6=0 → FRN 2  (spare)
	//   Bit 5=1 → FRN 3  (I062/015) - wait, diagnostic said bit 3 is set
	//   Bit 4=1 → FRN 4  (I062/070)
	//   Bit 3=1 → FRN 5  (I062/105)
	//   Bit 2=1 → FRN 6  (I062/100)
	//   Bit 1=1 → FRN 7  (I062/185)
	//   Bit 0=1 → FX
	//
	// Diagnostic said bits [1 3 4 5 6 7 9 10 11 12 13 14 15 17 18 19 21 22 30 31]
	//
	// Wait - the diagnostic is using 1-based "bit number", not FRN!
	// Let me count correctly:
	//
	// Byte 1 contributes 7 bits: bit 1, 2, 3, 4, 5, 6, 7 (in diagnostic numbering)
	// Byte 2 contributes 7 bits: bit 9, 10, 11, 12, 13, 14, 15
	// Byte 3 contributes 7 bits: bit 17, 18, 19, 20, 21, 22, 23
	// Byte 4 contributes 7 bits: bit 25, 26, 27, 28, 29, 30, 31
	//
	// So "bit 30" in diagnostic = byte 4, bit position 2 (counting from bit 7)
	// And "bit 31" = byte 4, bit position 1
	//
	// Byte 4: 06 = 00000110
	//   Bit 7=0
	//   Bit 6=0
	//   Bit 5=0
	//   Bit 4=0
	//   Bit 3=0
	//   Bit 2=1 → diagnostic bit 30
	//   Bit 1=1 → diagnostic bit 31
	//   Bit 0=0 (no FX)
	//
	// Now map diagnostic bit number to FRN:
	// Diagnostic bit 1 → FRN 1
	// Diagnostic bit 2 (skipped, not set) → FRN 2 (spare)
	// Diagnostic bit 3 → FRN 3
	// ... and so on, BUT we have to account for spare FRN 2
	//
	// Actually, FRN and diagnostic bit don't align because of spares.
	// Let me use FRN directly from the UAP:

	t.Log("\n=== FRN Mapping (from UAP) ===")
	frnToItem := []string{
		"",           // FRN 0 (doesn't exist)
		"I062/010",   // FRN 1
		"Spare",      // FRN 2
		"I062/015",   // FRN 3
		"I062/070",   // FRN 4
		"I062/105",   // FRN 5
		"I062/100",   // FRN 6
		"I062/185",   // FRN 7
		"I062/210",   // FRN 8
		"I062/060",   // FRN 9
		"I062/245",   // FRN 10
		"I062/380",   // FRN 11
		"I062/040",   // FRN 12
		"I062/080",   // FRN 13
		"I062/290",   // FRN 14
		"I062/200",   // FRN 15
		"I062/295",   // FRN 16
		"I062/136",   // FRN 17
		"I062/130",   // FRN 18
		"I062/135",   // FRN 19
		"I062/220",   // FRN 20
		"I062/390",   // FRN 21
		"I062/270",   // FRN 22
		"I062/300",   // FRN 23
		"I062/110",   // FRN 24
		"I062/120",   // FRN 25
		"I062/510",   // FRN 26
		"I062/500",   // FRN 27
		"I062/340",   // FRN 28
		"Spare",      // FRN 29
		"Spare",      // FRN 30
		"Spare",      // FRN 31
		"Spare",      // FRN 32
		"Spare",      // FRN 33
		"RE062",      // FRN 34
		"SP062",      // FRN 35
	}

	// Now decode FSPEC byte by byte
	presentFRNs := []int{}
	frn := 1
	for byteIdx := 0; byteIdx < 4; byteIdx++ {
		b := fspec[byteIdx]
		for bitIdx := 7; bitIdx >= 1; bitIdx-- { // bits 7..1 (bit 0 is FX)
			if b&(1<<bitIdx) != 0 {
				presentFRNs = append(presentFRNs, frn)
				t.Logf("FRN %2d (%s): SET", frn, frnToItem[frn])
			}
			frn++
		}
		// Skip FX bit (bit 0)
	}

	t.Logf("\nPresent FRNs: %v", presentFRNs)

	// Check for problematic FRNs
	for _, frn := range presentFRNs {
		if frn == 27 {
			t.Log("⚠️ FRN 27 (I062/500 - Estimated Accuracies) is present")
		}
		if frn == 28 {
			t.Log("⚠️ FRN 28 (I062/340 - Measured Information) is present")
		}
		if frn >= 29 && frn <= 33 {
			t.Logf("❌ FRN %d (SPARE) is SET - THIS IS INVALID!", frn)
		}
	}
}
