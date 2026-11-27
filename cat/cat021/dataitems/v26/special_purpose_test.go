package v26

import (
	"bytes"
	"testing"
)

// TestSpecialPurposeLen checks that LEN of SP021 counts the length octet
// itself: LEN = 3 takes 2 more octets and leaves the next item alone.
func TestSpecialPurposeLen(t *testing.T) {
	var s SpecialPurpose
	buf := bytes.NewBuffer([]byte{3, 0xAA, 0xBB, 0xCC})
	n, err := s.Decode(buf)
	if err != nil || n != 3 || buf.Len() != 1 || !bytes.Equal(s.Data, []byte{3, 0xAA, 0xBB}) {
		t.Fatalf("n = %d, err %v, left %d, Data % X; want 3 octets read, 1 left", n, err, buf.Len(), s.Data)
	}
	var out bytes.Buffer
	if _, err := s.Encode(&out); err != nil || !bytes.Equal(out.Bytes(), []byte{3, 0xAA, 0xBB}) {
		t.Errorf("encode = % X (%v)", out.Bytes(), err)
	}
	if _, err := (&SpecialPurpose{}).Decode(bytes.NewBuffer([]byte{0})); err == nil {
		t.Error("LEN 0 accepted")
	}
	if _, err := (&SpecialPurpose{}).Decode(bytes.NewBuffer([]byte{4, 1, 2})); err == nil {
		t.Error("LEN beyond the data accepted")
	}
	out.Reset()
	if _, err := (&SpecialPurpose{}).Encode(&out); err != nil || !bytes.Equal(out.Bytes(), []byte{1}) {
		t.Errorf("empty SP encodes as % X (%v), want 01", out.Bytes(), err)
	}
	if err := (&SpecialPurpose{Data: []byte{2, 1, 2}}).Validate(); err == nil {
		t.Error("LEN 2 with 3 octets accepted")
	}
}
