// dataitems/cat021/acas_resolution_advisory_test.go
package v26

import (
	"bytes"
	"testing"
)

// Hand-computed vector against CAT021 v2.6 §5.2.39: the item is the verbatim
// TC28/ST2 (BDS 6,1) image — TYP(5)+STYP(3)+ARA(14)+RAC(4)+RAT+MTE+TTI(2)+TID(26).
func TestACASResolutionAdvisoryVector(t *testing.T) {
	a := &ACASResolutionAdvisory{
		TYP:  28,
		STYP: 2,
		ARA:  0x2A55,
		RAC:  0x0A,
		RAT:  true,
		MTE:  false,
		TTI:  1,
		TID:  0x2ABCDEF,
	}
	var buf bytes.Buffer
	if _, err := a.Encode(&buf); err != nil {
		t.Fatalf("encode: %v", err)
	}
	want := []byte{0xE2, 0xA9, 0x56, 0xA6, 0xAB, 0xCD, 0xEF}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("encoded % X, want % X", buf.Bytes(), want)
	}

	var d ACASResolutionAdvisory
	if _, err := d.Decode(&buf); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if d != *a {
		t.Fatalf("roundtrip mismatch: %+v != %+v", d, *a)
	}
}

func TestServiceManagementDataDriven(t *testing.T) {
	s := &ServiceManagement{RP: 0} // data-driven mode, ED-129B REQ 90
	var buf bytes.Buffer
	if _, err := s.Encode(&buf); err != nil {
		t.Fatalf("RP=0 (data-driven) must encode: %v", err)
	}
	if buf.Bytes()[0] != 0x00 {
		t.Fatalf("encoded %#x, want 0x00", buf.Bytes()[0])
	}
	// 4-second periodic reporting = raw 8 (LSB 0.5 s).
	s = &ServiceManagement{RP: 8}
	buf.Reset()
	if _, err := s.Encode(&buf); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if buf.Bytes()[0] != 0x08 {
		t.Fatalf("encoded %#x, want 0x08", buf.Bytes()[0])
	}
}
