// cat/cat062/fusion_bug_hunt_feb22_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionMessageFeb22 decodes real failing message from fusion (2026-02-22 09:22)
func TestFusionMessageFeb22(t *testing.T) {
	// Real message from fusion that fails to decode in track-converter
	hexData := "3e0059bfffed0664010141f93f0094d7b700115ecef7e7f704b884fe070130fffe0200003534b7d77820c121010001024535" +
		"34b7d77820" // ... (truncated in log, but let's work with what we have)

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	t.Logf("Message length: %d bytes", len(data))
	t.Logf("Full hex: %s\n", hex.EncodeToString(data))

	// Decode with gobelix
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}

	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(data)
	if err != nil {
		t.Logf("❌ Decode error: %v", err)
		t.Logf("This is the bug we need to find!")

		// Manual FSPEC analysis
		t.Logf("\n=== Manual Analysis ===")
		t.Logf("CAT: 0x%02X (%d)", data[0], data[0])
		length := int(data[1])<<8 | int(data[2])
		t.Logf("LEN: %d bytes", length)

		t.Logf("\nFSPEC:")
		offset := 3
		for i := 0; i < 8 && offset < len(data); i++ {
			fspecByte := data[offset]
			t.Logf("  Byte %d: 0x%02X (%08b)", i+1, fspecByte, fspecByte)
			offset++
			hasFX := (fspecByte & 0x01) != 0
			if !hasFX {
				break
			}
		}

		t.Logf("\nData starts at offset %d", offset)
		t.Fatalf("Decode failed as expected")
	}

	t.Logf("✅ Decode succeeded! (unexpected - message was supposed to fail)")
	if len(dataBlocks) > 0 {
		t.Logf("Decoded %d data blocks", len(dataBlocks))
	}
}
