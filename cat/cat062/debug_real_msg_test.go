package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestDebugRealMessage decodes the actual failing message
func TestDebugRealMessage(t *testing.T) {
	// Real fusion message that has 1-byte length mismatch
	hexData := "3e0058bfffed066401014377090086cb860015d714faa6d2f3d617fd4bfd6b00010649005da6b5d92120c1210100471db95da6b5d921200000f5718b010130b008080a000005280528ffdcc4010d010cf9817c7c84620e180a"
	
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}
	
	t.Logf("Message length: %d bytes", len(data))
	if len(data) >= 3 {
		cat := data[0]
		length := int(data[1])<<8 | int(data[2])
		t.Logf("CAT=%d, LENGTH=%d, actual=%d", cat, length, len(data))
		if length != len(data) {
			t.Logf("⚠️  LENGTH MISMATCH: %d bytes", len(data)-length)
		}
	}
	
	// Try to decode
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)
	
	_, err = decoder.DecodeAll(data)
	if err != nil {
		t.Logf("❌ Decode error: %v", err)
		t.Fatal("Decode failed as expected - this proves the 1-byte bug exists in the message")
	}
	
	t.Logf("✅ Decoded successfully (unexpected!)")
}
