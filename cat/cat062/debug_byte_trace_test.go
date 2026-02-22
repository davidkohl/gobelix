package cat062_test

import (
	"encoding/hex"
	"testing"
)

// TestByteByByteTrace manually traces through the failing message to find
// EXACTLY where the byte misalignment happens.
func TestByteByByteTrace(t *testing.T) {
	hexData := "3e0058bfffed066401016da302008d3cfd001ff78d02ba90fb7453fc85014b00fe020000407539ccd460c12101004bb84d407539ccd4600000c34783010130b0000000000005f005f00001c4009d009e00a76e7284622728"

	data, _ := hex.DecodeString(hexData)

	t.Logf("Total: %d bytes", len(data))
	t.Logf("CAT: 0x%02x (%d)", data[0], data[0])
	length := int(data[1])<<8 | int(data[2])
	t.Logf("LENGTH: %d (matches actual: %v)", length, length == len(data))

	// FSPEC
	pos := 3
	t.Logf("\n=== FSPEC ===")
	fspec := []byte{}
	for {
		b := data[pos]
		fspec = append(fspec, b)
		t.Logf("Byte %d: 0x%02x = 0b%08b", pos, b, b)
		pos++
		if (b & 0x01) == 0 {
			break
		}
	}
	t.Logf("FSPEC: %x (%d bytes)", fspec, len(fspec))
	t.Logf("Current position: %d", pos)

	// Map FRN to data item
	items := []string{
		// Byte 1 (bf = 10111111): bits 8,6,5,4,3,2,1 set
		"I062/010", "", "I062/070", "I062/105", "I062/100", "I062/185", "I062/210",
		// Byte 2 (ff = 11111111): bits 8,7,6,5,4,3,2,1 set
		"I062/060", "I062/245", "I062/380", "I062/040", "I062/080", "I062/290", "I062/200",
		// Byte 3 (ed = 11101101): bits 8,7,6,5,3,2,1 set
		"I062/295", "I062/136", "I062/130", "I062/135", "", "I062/390", "I062/270",
		// Byte 4 (06 = 00000110): bits 2,1 set
		"", "", "", "", "I062/500", "I062/340", "",
	}

	present := []string{}
	itemIdx := 0
	for _, fb := range fspec {
		for bit := 7; bit >= 1; bit-- {
			if (fb & (1 << bit)) != 0 {
				if itemIdx < len(items) && items[itemIdx] != "" {
					present = append(present, items[itemIdx])
				}
			}
			itemIdx++
		}
	}

	t.Logf("\n=== Present items (%d) ===", len(present))
	for i, item := range present {
		t.Logf("%2d. %s", i+1, item)
	}

	// Known fixed sizes
	sizes := map[string]int{
		"I062/010": 2,  // SAC/SIC
		"I062/070": 3,  // Time of Track Info
		"I062/105": 8,  // WGS-84 Position
		"I062/100": 6,  // Cartesian Position
		"I062/185": 4,  // Cartesian Velocity
		"I062/210": 2,  // Cartesian Acceleration
		"I062/060": 2,  // Mode 3/A
		"I062/245": 7,  // Target Identification
		"I062/040": 2,  // Track Number
		"I062/080": 3,  // Track Status
		"I062/200": 1,  // Mode of Movement
		"I062/136": 2,  // Measured Flight Level
		"I062/130": 2,  // Calculated Position Accuracy
		"I062/135": 1,  // Calculated Geometric Altitude Accuracy
	}

	t.Logf("\n=== Parsing data items ===")
	for _, item := range present {
		if size, ok := sizes[item]; ok {
			if pos+size > len(data) {
				t.Errorf("ERROR at %s: need %d bytes, only %d available", item, size, len(data)-pos)
				break
			}
			itemData := data[pos : pos+size]
			t.Logf("%s @ byte %d: %x (%d bytes)", item, pos, itemData, size)
			pos += size
		} else {
			// Compound item
			t.Logf("%s @ byte %d: COMPOUND (unknown size)", item, pos)

			switch item {
			case "I062/380": // Aircraft Derived Data
				startPos := pos
				// Read FSPEC
				compoundFspec := []byte{}
				for {
					if pos >= len(data) {
						t.Fatalf("ERROR: buffer overrun reading %s FSPEC at byte %d", item, pos)
					}
					b := data[pos]
					compoundFspec = append(compoundFspec, b)
					pos++
					if (b & 0x01) == 0 {
						break
					}
				}
				t.Logf("  FSPEC: %x (%d bytes)", compoundFspec, len(compoundFspec))
				// I062/380 with empty FSPEC (0x00) has no subfields
				if len(compoundFspec) == 1 && compoundFspec[0] == 0x00 {
					t.Logf("  Empty FSPEC, no subfields")
				} else {
					t.Logf("  SKIPPING subfield parsing (complex)")
					// For now, we'll skip ahead - we know from the message this should work
				}
				bytesConsumed := pos - startPos
				t.Logf("  Total: %d bytes", bytesConsumed)

			case "I062/290": // System Track Update Ages
				startPos := pos
				// Read FSPEC
				fspecByte := data[pos]
				pos++
				t.Logf("  FSPEC byte 1: 0x%02x = 0b%08b", fspecByte, fspecByte)

				hasExtension := (fspecByte & 0x01) != 0
				if hasExtension {
					fspecByte2 := data[pos]
					pos++
					t.Logf("  FSPEC byte 2: 0x%02x = 0b%08b", fspecByte2, fspecByte2)
				}

				// Count subfields in FSPEC byte 1 (bits 7-1, skip bit 0 which is FX)
				subfieldCount := 0
				for bit := 7; bit >= 1; bit-- {
					if (fspecByte & (1 << bit)) != 0 {
						subfieldCount++
						// Subfield #5 (bit 4) is 2 bytes, others are 1 byte
						if bit == 4 {
							pos += 2
							t.Logf("  Subfield #5 (bit %d): 2 bytes", bit)
						} else {
							pos++
							t.Logf("  Subfield (bit %d): 1 byte", bit)
						}
					}
				}

				if hasExtension {
					// Read subfields from FSPEC byte 2 (bits 7-1)
					fspecByte2 := data[startPos+1]
					for bit := 7; bit >= 1; bit-- {
						if (fspecByte2 & (1 << bit)) != 0 {
							pos++
							t.Logf("  Subfield ext (bit %d): 1 byte", bit)
						}
					}
				}

				bytesConsumed := pos - startPos
				fspecSize := 1
				if hasExtension {
					fspecSize = 2
				}
				t.Logf("  Total: %d bytes (FSPEC + %d subfield bytes)", bytesConsumed, bytesConsumed-fspecSize)

			case "I062/295": // Track Data Ages
				startPos := pos
				// Read FSPEC (up to 5 bytes with FX extension)
				fspecBytes := []byte{}
				for i := 0; i < 5; i++ {
					if pos >= len(data) {
						t.Fatalf("ERROR: buffer overrun reading %s FSPEC", item)
					}
					b := data[pos]
					fspecBytes = append(fspecBytes, b)
					pos++
					if (b & 0x01) == 0 {
						break
					}
				}
				t.Logf("  FSPEC: %x (%d bytes)", fspecBytes, len(fspecBytes))

				// Count subfields (each set bit except FX = 1 byte)
				subfieldCount := 0
				for _, fb := range fspecBytes {
					for bit := 7; bit >= 1; bit-- {
						if (fb & (1 << bit)) != 0 {
							subfieldCount++
							pos++
						}
					}
				}
				t.Logf("  Subfields: %d bytes", subfieldCount)

				bytesConsumed := pos - startPos
				t.Logf("  Total: %d bytes", bytesConsumed)

			case "I062/390": // Flight Plan Related Data
				// Read FSPEC (variable length)
				fspecBytes := []byte{}
				for {
					if pos >= len(data) {
						t.Fatalf("ERROR: buffer overrun reading %s FSPEC", item)
					}
					b := data[pos]
					fspecBytes = append(fspecBytes, b)
					pos++
					if (b & 0x01) == 0 {
						break
					}
				}
				t.Logf("  FSPEC: %x (%d bytes)", fspecBytes, len(fspecBytes))
				t.Logf("  SKIPPING subfield parsing (too complex, would need full spec)")
				// Continue to next items instead of stopping
				t.Logf("  WARNING: I062/390 byte count cannot be verified - continuing anyway")

			case "I062/270": // Target Size & Orientation
				// Variable length item with extension bit
				startPos270 := pos
				for {
					if pos >= len(data) {
						t.Fatalf("ERROR: buffer overrun reading %s", item)
					}
					b := data[pos]
					pos++
					t.Logf("  Byte: 0x%02x", b)
					if (b & 0x01) == 0 {
						break
					}
				}
				bytesConsumed := pos - startPos270
				t.Logf("  Total: %d bytes", bytesConsumed)

			case "I062/500": // Estimated Accuracies
				t.Logf("*** REACHED I062/500 ***")
				t.Logf("Current position: %d", pos)
				t.Logf("Bytes remaining: %d", len(data)-pos)

				if pos >= len(data) {
					t.Fatalf("ERROR: No bytes left for I062/500 FSPEC!")
				}

				fspecByte1 := data[pos]
				t.Logf("FSPEC byte 1: 0x%02x = 0b%08b", fspecByte1, fspecByte1)
				pos++

				// Decode FSPEC
				hasAPC := (fspecByte1 & 0x80) != 0
				hasCOV := (fspecByte1 & 0x40) != 0
				hasAPW := (fspecByte1 & 0x20) != 0
				hasAGA := (fspecByte1 & 0x10) != 0
				hasABA := (fspecByte1 & 0x08) != 0
				hasATV := (fspecByte1 & 0x04) != 0
				hasAA := (fspecByte1 & 0x02) != 0
				hasFX := (fspecByte1 & 0x01) != 0

				t.Logf("  APC (pos cart): %v", hasAPC)
				t.Logf("  COV (xy cov):   %v", hasCOV)
				t.Logf("  APW (pos wgs):  %v", hasAPW)
				t.Logf("  AGA (geo alt):  %v", hasAGA)
				t.Logf("  ABA (baro alt): %v", hasABA)
				t.Logf("  ATV (velocity): %v", hasATV)
				t.Logf("  AA (accel):     %v", hasAA)
				t.Logf("  FX (extension): %v", hasFX)

				bytesNeeded := 0
				if hasAPC {
					bytesNeeded += 4
				}
				if hasCOV {
					bytesNeeded += 2
				}
				if hasAPW {
					bytesNeeded += 4
				}
				if hasAGA {
					bytesNeeded += 1
				}
				if hasABA {
					bytesNeeded += 1
				}
				if hasATV {
					bytesNeeded += 2
				}
				if hasAA {
					bytesNeeded += 2
				}

				if hasFX {
					if pos >= len(data) {
						t.Fatalf("ERROR: No bytes for I062/500 FSPEC byte 2!")
					}
					fspecByte2 := data[pos]
					t.Logf("FSPEC byte 2: 0x%02x = 0b%08b", fspecByte2, fspecByte2)
					pos++

					hasARC := (fspecByte2 & 0x80) != 0
					t.Logf("  ARC (rate climb): %v", hasARC)
					if hasARC {
						bytesNeeded += 1
					}
				}

				t.Logf("Total bytes needed: %d", bytesNeeded)
				t.Logf("Bytes available: %d", len(data)-pos)

				if len(data)-pos < bytesNeeded {
					t.Errorf("❌ ERROR: Not enough bytes for I062/500 data!")
					t.Errorf("  Need: %d, Have: %d, Short by: %d", bytesNeeded, len(data)-pos, bytesNeeded-(len(data)-pos))
					break
				} else {
					t.Logf("✅ Sufficient bytes available for I062/500")
				}

			case "I062/340": // Measured Information
				t.Logf("Reached I062/340 at byte %d", pos)
				break
			}
		}
	}

	t.Logf("\n=== Final position: %d / %d ===", pos, len(data))
}
