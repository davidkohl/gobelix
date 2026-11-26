// cat/cat062/fusion_i380_decode_test.go
package cat062_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestDecodeI380FromFusionMessage decodes I062/380 out of a real fusion
// message and checks it consumes exactly the item's length.
//
// The item starts at offset 42 with the primary subfield C1 01 01 00:
// octet 1 announces ADR and ID, octets 2 and 3 carry only FX, octet 4 is
// empty. That makes the item 4 + 3 + 6 = 13 octets. The bytes that follow
// (0EF6 03010130 ...) are I062/040, I062/080 and the rest of the record, not
// part of I062/380; an earlier version of this test assumed 31 octets.
func TestDecodeI380FromFusionMessage(t *testing.T) {
	hexData := "3E0056BFFFED066401016D32EE008271240023976605F3F7EEAD5603C8FC5114EA0A900014C673C78820C10101007380C714C673C788200EF603010130B0020202400006170617FFFDC401720171DCC4B7B784622A28"

	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	const i380Offset = 42
	const i380Len = 13
	want := data[i380Offset : i380Offset+i380Len]

	item := &v120.AircraftDerivedData{}
	buf := bytes.NewBuffer(data[i380Offset:])
	n, err := item.Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if n != i380Len {
		t.Fatalf("Decode() consumed %d bytes, want %d (%x)", n, i380Len, want)
	}

	// The next item in the record is I062/040 (track number 0x0EF6).
	if next := buf.Bytes()[:2]; !bytes.Equal(next, []byte{0x0E, 0xF6}) {
		t.Errorf("bytes after I062/380 = %x, want 0ef6 (I062/040)", next)
	}

	if item.TargetAddress == nil || *item.TargetAddress != 0x7380C7 {
		t.Errorf("TargetAddress = %v, want 0x7380C7", item.TargetAddress)
	}
	if item.TargetIdentification == nil || *item.TargetIdentification != "ELY318" {
		t.Errorf("TargetIdentification = %v, want ELY318", item.TargetIdentification)
	}

	// Re-encoding gives the canonical form: the empty trailing primary
	// subfield octets are dropped, the subfield data is unchanged.
	var out bytes.Buffer
	if _, err := item.Encode(&out); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	canonical := append([]byte{0xC0}, want[4:]...)
	if !bytes.Equal(out.Bytes(), canonical) {
		t.Errorf("Encode() = %x, want %x", out.Bytes(), canonical)
	}
}
