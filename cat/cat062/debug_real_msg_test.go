package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestDebugRealMessage decodes a fusion capture that is one byte longer than
// its LENGTH field. The extra byte is 0x0a, a newline appended by the capture
// tool after the 88-byte data block; the block itself is well formed. The
// earlier decode failure ("decoding APW: buffer too short") came from the
// I062/290 MDS misread, not from the length mismatch.
func TestDebugRealMessage(t *testing.T) {
	hexData := "3e0058bfffed066401014377090086cb860015d714faa6d2f3d617fd4bfd6b00010649005da6b5d92120c1210100471db95da6b5d921200000f5718b010130b008080a000005280528ffdcc4010d010cf9817c7c84620e180a"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	length := int(data[1])<<8 | int(data[2])
	if length != 88 || len(data) != 89 || data[88] != '\n' {
		t.Fatalf("fixture changed: LENGTH=%d, have %d bytes, last byte %#x", length, len(data), data[len(data)-1])
	}

	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create UAP: %v", err)
	}
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	blocks, err := decoder.DecodeAll(data[:length])
	if err != nil {
		t.Fatalf("DecodeAll() error = %v", err)
	}
	if len(blocks) != 1 || len(blocks[0].Records()) != 1 {
		t.Fatalf("decoded %d blocks, want 1 block with 1 record", len(blocks))
	}
}
