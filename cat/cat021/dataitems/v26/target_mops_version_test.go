// dataitems/cat021/mops_version_test.go
package v26

import (
	"bytes"
	"testing"
)

// Reserved VN values (4-7) are transmitted by real transponders; the encoder
// must carry them faithfully rather than reject the record.
func TestMOPSVersionReservedVNRoundtrip(t *testing.T) {
	for vn := uint8(0); vn <= 7; vn++ {
		m := &MOPSVersion{VN: vn, LTT: LTT1090ES}
		var buf bytes.Buffer
		if _, err := m.Encode(&buf); err != nil {
			t.Fatalf("VN=%d: encode: %v", vn, err)
		}
		want := byte(vn<<3 | LTT1090ES)
		if got := buf.Bytes()[0]; got != want {
			t.Fatalf("VN=%d: encoded 0x%02X, want 0x%02X", vn, got, want)
		}
		var d MOPSVersion
		if _, err := d.Decode(&buf); err != nil {
			t.Fatalf("VN=%d: decode: %v", vn, err)
		}
		if d.VN != vn || d.LTT != LTT1090ES || d.VNS {
			t.Fatalf("VN=%d: decoded %+v", vn, d)
		}
	}
}

func TestMOPSVersionValidateRejectsOverwideVN(t *testing.T) {
	m := &MOPSVersion{VN: 8, LTT: LTT1090ES}
	if err := m.Validate(); err == nil {
		t.Fatal("VN=8 (doesn't fit 3 bits) must fail validation")
	}
}
