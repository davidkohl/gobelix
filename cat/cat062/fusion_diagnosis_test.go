// cat/cat062/fusion_diagnosis_test.go
package cat062_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionDiagnosis provides detailed byte-by-byte diagnosis of decode failures
func TestFusionDiagnosis(t *testing.T) {
	tests := []struct {
		name    string
		hexData string
		desc    string
	}{
		{
			name:    "Message #1 - I062/220 EOF",
			hexData: "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528",
			desc:    "Reported I062/220 EOF error at byte 83/83",
		},
		{
			name:    "Message #2 - I062/500 ATV",
			hexData: "3E0056BFFFED066401015D1A2F008E9F4B001D3BC0009AE7FD1579FCDBFF830E000200005161B2D995A0C101010039DE475161B2D995A0CB2483010130B12000040400000005C805C8000084001400142A2484622528",
			desc:    "Reported I062/500 'buffer too short for ATV' at byte 78/78",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex to bytes
			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to parse hex: %v", err)
			}

			t.Logf("Message length: %d bytes", len(data))
			t.Logf("Description: %s", tt.desc)
			t.Logf("Full hex: %s", hex.EncodeToString(data))

			// Manual decode to understand structure
			t.Log("\n=== Manual Structure Analysis ===")

			if len(data) < 3 {
				t.Fatal("Message too short for ASTERIX header")
			}

			cat := data[0]
			length := uint16(data[1])<<8 | uint16(data[2])
			t.Logf("CAT: %d", cat)
			t.Logf("Block length: %d bytes", length)

			if cat != 62 {
				t.Fatalf("Expected CAT062, got CAT%03d", cat)
			}

			// Decode FSPEC
			fspecStart := 3
			fspecBytes := []byte{}
			for i := fspecStart; i < len(data); i++ {
				fspecBytes = append(fspecBytes, data[i])
				t.Logf("FSPEC byte %d: 0x%02X (%08b)", len(fspecBytes), data[i], data[i])
				if data[i]&0x01 == 0 {
					// No FX bit - end of FSPEC
					break
				}
			}

			t.Logf("\nFSPEC: %d bytes: %s", len(fspecBytes), hex.EncodeToString(fspecBytes))

			// Show which data items are present
			t.Log("\n=== Data Items Present ===")
			dataItemMap := map[int]string{
				1:  "I062/010 - Data Source Identifier",
				2:  "I062/015 - Service Identification",
				3:  "I062/070 - Time of Track Information",
				4:  "I062/105 - Calculated Position (WGS-84)",
				5:  "I062/100 - Calculated Position (Cartesian)",
				6:  "I062/185 - Calculated Track Velocity (Cartesian)",
				7:  "I062/210 - Calculated Acceleration (Cartesian)",
				8:  "I062/060 - Track Mode 3/A Code",
				9:  "I062/245 - Target Identification",
				10: "I062/380 - Aircraft Derived Data",
				11: "I062/040 - Track Number",
				12: "I062/080 - Track Status",
				13: "I062/290 - System Track Update Ages",
				14: "I062/200 - Mode of Movement",
				// FX - bit 1 extends to next byte
				// Byte 2:
				15: "I062/295 - Track Data Ages",
				16: "I062/136 - Measured Flight Level",
				17: "I062/130 - Calculated Track Geometric Altitude",
				18: "I062/135 - Calculated Track Barometric Altitude",
				19: "I062/220 - Calculated Rate of Climb/Descent",
				20: "I062/390 - Flight Plan Related Data",
				21: "I062/270 - Target Size & Orientation",
				22: "I062/300 - Vehicle Fleet Identification",
				// FX - bit 15 extends to byte 3
				// Byte 3:
				23: "I062/110 - Mode 5 Data",
				24: "I062/120 - Track Mode 2 Code",
				25: "I062/510 - Composed Track Number",
				26: "I062/500 - Estimated Accuracies",
				27: "I062/340 - Measured Information",
				// ... more
			}

			bitNum := 0
			for byteIdx, fspecByte := range fspecBytes {
				for bitIdx := 7; bitIdx >= 0; bitIdx-- {
					bitNum++
					if bitIdx == 0 && byteIdx < len(fspecBytes)-1 {
						// FX bit (skip)
						continue
					}
					if fspecByte&(1<<bitIdx) != 0 {
						itemName := dataItemMap[bitNum]
						if itemName == "" {
							itemName = fmt.Sprintf("Unknown item #%d", bitNum)
						}
						t.Logf("  Bit %2d (byte %d, bit %d): %s", bitNum, byteIdx+1, bitIdx, itemName)
					}
				}
			}

			// Now try to decode with gobelix
			t.Log("\n=== gobelix Decode Attempt ===")
			decoder := asterix.NewDecoder()
			uap062, err := uap.NewUAP120()
			if err != nil {
				t.Fatalf("Failed to create UAP: %v", err)
			}
			decoder.RegisterUAP(uap062)

			_, err = decoder.DecodeAll(data)
			if err != nil {
				t.Logf("Decode error: %v", err)
				t.Logf("This confirms the decode failure")
			} else {
				t.Log("Decode succeeded!")
			}

			// Try partial decode to see how far we get
			t.Log("\n=== Partial Decode Test ===")
			for tryLen := len(data); tryLen >= 10; tryLen -= 5 {
				partial := data[:tryLen]
				_, err := decoder.DecodeAll(partial)
				if err == nil {
					t.Logf("Decode succeeded with %d bytes (full is %d)", tryLen, len(data))
					break
				}
				if tryLen == len(data) {
					t.Logf("Full message (%d bytes): FAILED - %v", tryLen, err)
				}
			}
		})
	}
}

// TestManualDataItemDecode attempts to decode specific data items in isolation
func TestManualDataItemDecode(t *testing.T) {
	// Extract I062/220 (Calculated Rate of Climb/Descent) from message #1
	// After manual analysis, find where I062/220 should be in the message

	t.Log("=== I062/220 Isolation Test ===")

	// I062/220 is a simple 2-byte field
	// From the error "EOF at byte 83/83", the message ends exactly where I062/220 should be
	// This suggests I062/220 is present in FSPEC but data is missing

	msg1Hex := "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528"
	data, _ := hex.DecodeString(msg1Hex)

	t.Logf("Message #1 total length: %d bytes", len(data))
	t.Logf("Expected length from header: %d", (uint16(data[1])<<8)|uint16(data[2]))

	// The FSPEC shows I062/220 is present (bit 19)
	// But when we reach byte 83, there's no more data
	// This means either:
	// 1. Fusion miscalculated the length field, OR
	// 2. Fusion set the I062/220 bit but didn't write the data, OR
	// 3. gobelix is consuming too many bytes before reaching I062/220

	expectedLen := (uint16(data[1])<<8) | uint16(data[2])
	actualLen := len(data)

	if expectedLen != uint16(actualLen) {
		t.Logf("LENGTH MISMATCH: Header says %d bytes, actual data is %d bytes", expectedLen, actualLen)
		t.Logf("Difference: %d bytes", int(actualLen)-int(expectedLen))
	} else {
		t.Logf("Length field matches actual data: %d bytes", actualLen)
	}
}

// TestFSPECDecoding verifies FSPEC bit interpretation
func TestFSPECDecoding(t *testing.T) {
	msg1Hex := "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528"
	data, _ := hex.DecodeString(msg1Hex)

	// FSPEC starts at byte 3
	fspec := []byte{}
	for i := 3; i < len(data); i++ {
		fspec = append(fspec, data[i])
		if data[i]&0x01 == 0 {
			break
		}
	}

	t.Logf("FSPEC: %s (%d bytes)", hex.EncodeToString(fspec), len(fspec))

	// Decode bit by bit
	presentItems := []int{}
	bitNum := 0
	for byteIdx, b := range fspec {
		for bitIdx := 7; bitIdx >= 0; bitIdx-- {
			bitNum++
			if bitIdx == 0 && byteIdx < len(fspec)-1 {
				// FX bit, skip
				continue
			}
			if b&(1<<bitIdx) != 0 {
				presentItems = append(presentItems, bitNum)
			}
		}
	}

	t.Logf("Present data items (bit numbers): %v", presentItems)
	t.Logf("Total items present: %d", len(presentItems))

	// Key question: Is bit 19 (I062/220) really set?
	hasBit19 := false
	for _, bit := range presentItems {
		if bit == 19 {
			hasBit19 = true
			break
		}
	}

	if hasBit19 {
		t.Log("I062/220 (bit 19) IS present in FSPEC")
	} else {
		t.Log("I062/220 (bit 19) is NOT present in FSPEC")
	}
}
