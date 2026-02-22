// cat/cat062/fusion_bug_analysis_test.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionI062_500Bug analyzes the "unexpected second extension bit" error
func TestFusionI062_500Bug(t *testing.T) {
	// Real message from fusion that fails with:
	// "I062/500, at byte 73/81: unexpected second extension bit"
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	t.Logf("Message length: %d bytes", len(data))

	// Manual structure analysis
	if len(data) < 3 {
		t.Fatal("Message too short")
	}

	cat := data[0]
	length := uint16(data[1])<<8 | uint16(data[2])
	t.Logf("CAT: %d", cat)
	t.Logf("Block length: %d bytes", length)

	// Decode FSPEC
	fspecStart := 3
	fspecBytes := []byte{}
	for i := fspecStart; i < len(data); i++ {
		fspecBytes = append(fspecBytes, data[i])
		t.Logf("FSPEC byte %d: 0x%02X (%08b)", len(fspecBytes), data[i], data[i])
		if data[i]&0x01 == 0 {
			break
		}
	}

	t.Logf("\nFSPEC: %s (%d bytes)", hex.EncodeToString(fspecBytes), len(fspecBytes))

	// Try to decode with gobelix
	t.Log("\n=== gobelix Decode Attempt ===")
	decoder := asterix.NewDecoder()
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	decoder.RegisterUAP(uap062)

	_, err = decoder.DecodeAll(data)
	if err != nil {
		t.Logf("Decode error (expected): %v", err)
	} else {
		t.Log("Decode succeeded (unexpected!)")
	}

	// Find I062/500 in the message
	// Based on FSPEC analysis, we need to count bytes consumed by each data item
	// to find where I062/500 starts

	t.Log("\n=== Manual I062/500 Location ===")
	t.Log("Error says 'at byte 73/81', so I062/500 FSPEC is around byte 73")
	t.Logf("Byte 73: 0x%02X (%08b)", data[73], data[73])
	if len(data) > 74 {
		t.Logf("Byte 74: 0x%02X (%08b)", data[74], data[74])
	}

	// Check if byte 73 or 74 has FX bit set in bit position 0
	if data[73]&0x01 != 0 {
		t.Logf("Byte 73 HAS FX bit set (bit 0 = 1)")
		if len(data) > 74 && data[74]&0x01 != 0 {
			t.Logf("Byte 74 ALSO has FX bit set (bit 0 = 1) ← THIS IS THE BUG!")
		}
	}
}
