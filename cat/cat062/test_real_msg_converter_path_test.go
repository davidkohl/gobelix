// cat/cat062/test_real_msg_converter_path.go
package cat062_test

import (
	"encoding/hex"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestRealMessageConverterPath uses the same decoding path as track-converter
// (NewDecoder, RegisterUAP, DecodeAll) on a message captured from NATS on
// 2026-02-22 10:35. The capture carries a trailing newline after the 88-byte
// data block, which is not part of the ASTERIX message and is cut off here.
func TestRealMessageConverterPath(t *testing.T) {
	hexData := "3e0058bfffed066401014377090086cb860015d714faa6d2f3d617fd4bfd6b00010649005da6b5d92120c1210100471db95da6b5d921200000f5718b010130b008080a000005280528ffdcc4010d010cf9817c7c84620e180a"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}
	lengthField := int(data[1])<<8 | int(data[2])
	if lengthField != len(data)-1 || data[lengthField] != '\n' {
		t.Fatalf("fixture changed: LENGTH=%d, have %d bytes", lengthField, len(data))
	}

	decoder := asterix.NewDecoder()
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("Failed to create CAT062 UAP: %v", err)
	}
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(data[:lengthField])
	if err != nil {
		t.Fatalf("DecodeAll() error = %v", err)
	}
	if len(dataBlocks) != 1 {
		t.Fatalf("decoded %d data blocks, want 1", len(dataBlocks))
	}
	records := dataBlocks[0].Records()
	if len(records) != 1 {
		t.Fatalf("decoded %d records, want 1", len(records))
	}
	if n := len(records[0].Items()); n != len(realFusionItems) {
		t.Errorf("decoded %d items, want %d", n, len(realFusionItems))
	}
}
