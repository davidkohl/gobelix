// cat/cat062/test_real_msg_converter_path.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestRealMessageConverterPath uses EXACT same decoding path as track-converter
func TestRealMessageConverterPath(t *testing.T) {
	// Real message from NATS (captured 2026-02-22 10:35)
	hexData := "3e0058bfffed066401014377090086cb860015d714faa6d2f3d617fd4bfd6b00010649005da6b5d92120c1210100471db95da6b5d921200000f5718b010130b008080a000005280528ffdcc4010d010cf9817c7c84620e180a"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	t.Logf("Message length: %d bytes", len(data))
	lengthField := int(data[1])<<8 | int(data[2])
	t.Logf("LENGTH field: %d", lengthField)

	if lengthField != len(data) {
		t.Logf("⚠️  LENGTH MISMATCH: header says %d, actual is %d (diff: %d)",
			lengthField, len(data), len(data)-lengthField)
	}

	// Use EXACT same code as track-converter (decoder.go line 64-76)
	decoder := asterix.NewDecoder()

	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create CAT062 UAP: %v", err)
	}
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(data)
	if err != nil {
		t.Logf("❌ DecodeAll failed (same as track-converter): %v", err)

		// This is the EXACT error track-converter sees!
		t.Fatalf("Decode failed with same error as track-converter")
	}

	t.Logf("✅ Decode succeeded with %d data blocks", len(dataBlocks))

	if len(dataBlocks) > 0 {
		db := dataBlocks[0]
		records := db.Records()
		if len(records) > 0 {
			items := records[0].Items()
			t.Logf("Decoded %d items", len(items))
		}
	}
}
