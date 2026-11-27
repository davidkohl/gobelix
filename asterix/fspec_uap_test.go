// asterix/fspec_uap_test.go
package asterix

import (
	"bytes"
	"errors"
	"testing"
)

// maskedUAP is a BaseUAP-based UAP (precomputed FRN mask path).
type maskedUAP struct {
	*BaseUAP
	mock *MockUAP
}

func (u *maskedUAP) CreateDataItem(id string) (DataItem, error) { return u.mock.CreateDataItem(id) }

// testUAPs returns the same layout as a generic UAP (MockUAP, lookup path)
// and as a BaseUAP (mask path): FRN 1 I021/010 (2 octets), FRN 2 spare,
// FRN 3 I021/030 (3 octets), FRN 4 I021/040 (1 octet).
func testUAPs(t *testing.T) map[string]UAP {
	t.Helper()
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 2, DataItem: "", Description: "Spare", Type: Fixed, Length: 0},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3},
		{FRN: 4, DataItem: "I021/040", Type: Fixed, Length: 1},
	}
	mock := &MockUAP{category: Cat021, version: "1.0", fields: fields}
	base, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]UAP{"generic": mock, "base": &maskedUAP{BaseUAP: base, mock: mock}}
}

// A set FSPEC bit for an FRN that the UAP cannot read (spare, beyond the
// UAP, or in an extension octet) is a decode error, never a silent skip
// that shifts the later items.
func TestRecordDecodeRejectsUnknownFSPECBits(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		frn  string
	}{
		{"spare FRN 2", []byte{0xC0, 0x01, 0x02}, "FRN 2"},
		{"FRN 5 beyond UAP", []byte{0x88, 0x01, 0x02, 0xAA}, "FRN 5"},
		{"FRN 7 beyond UAP", []byte{0x82, 0x01, 0x02, 0xAA}, "FRN 7"},
		{"FRN 8 in extension octet", []byte{0x81, 0x80, 0x01, 0x02, 0xAA}, "FRN 8"},
		{"FRN 21 in third octet", []byte{0x81, 0x01, 0x02, 0x01, 0x02}, "FRN 21"},
	}
	for name, uap := range testUAPs(t) {
		for _, tc := range cases {
			rec, err := NewRecord(Cat021, uap)
			if err != nil {
				t.Fatal(err)
			}
			_, err = rec.Decode(bytes.NewBuffer(tc.data))
			if !errors.Is(err, ErrInvalidFSPEC) {
				t.Errorf("%s/%s: err = %v, want ErrInvalidFSPEC", name, tc.name, err)
				continue
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tc.frn+" ")) {
				t.Errorf("%s/%s: error %q does not name %s", name, tc.name, err, tc.frn)
			}
		}
	}
}

// Valid FSPECs, including a trailing all-zero extension octet, still decode.
func TestRecordDecodeAcceptsKnownFSPECBits(t *testing.T) {
	cases := [][]byte{
		{0xB0, 0x01, 0x02, 0x03, 0x04, 0x05, 0xFF}, // FRN 1, 3, 4
		{0x81, 0x00, 0x01, 0x02},                   // FRN 1, empty 2nd octet
	}
	for name, uap := range testUAPs(t) {
		for i, data := range cases {
			rec, _ := NewRecord(Cat021, uap)
			n, err := rec.Decode(bytes.NewBuffer(data))
			if err != nil || n != len(data) {
				t.Errorf("%s/%d: n=%d err=%v, want n=%d", name, i, n, err, len(data))
			}
		}
	}
}

func rfsUAP(t *testing.T) *maskedUAP {
	t.Helper()
	fields := []DataField{
		{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2, Mandatory: true},
		{FRN: 3, DataItem: "I021/030", Type: Fixed, Length: 3},
		{FRN: 4, DataItem: "I021/040", Type: Fixed, Length: 1},
	}
	base, err := NewBaseUAP(Cat021, "1.0", fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := base.EnableRFS(2); err != nil {
		t.Fatal(err)
	}
	return &maskedUAP{BaseUAP: base, mock: &MockUAP{category: Cat021, fields: fields}}
}

func TestBaseUAPEnableRFSRejectsUsedFRN(t *testing.T) {
	base, err := NewBaseUAP(Cat021, "1.0", []DataField{{FRN: 1, DataItem: "I021/010", Type: Fixed, Length: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if err := base.EnableRFS(1); err == nil {
		t.Error("EnableRFS on a used FRN: want error")
	}
	if err := base.EnableRFS(0); err == nil {
		t.Error("EnableRFS(0): want error")
	}
	if err := base.EnableRFS(57); err == nil {
		t.Error("EnableRFS(57): want error (beyond an 8-octet FSPEC)")
	}
	for _, x := range []uint8{0, 2, 57} {
		if err := base.EnableRFS(2, x); err == nil {
			t.Errorf("EnableRFS(2, %d): want error (out of range or the RS FRN)", x)
		}
	}
}

// An FRN that EnableRFS excludes from the RFS field is a decode error when
// an RFS entry names it, and still decodes at its FSPEC position.
func TestRecordDecodeRFSExcludedFRN(t *testing.T) {
	u := rfsUAP(t)
	if err := u.BaseUAP.EnableRFS(2, 4); err != nil {
		t.Fatal(err)
	}
	// FSPEC FRN 1, 2 (RS) | I021/010 | RFS: N=1, FRN 4, I021/040
	rec, _ := NewRecord(Cat021, u)
	_, err := rec.Decode(bytes.NewBuffer([]byte{0xC0, 0x01, 0x02, 0x01, 0x04, 0xAB}))
	if !errors.Is(err, ErrInvalidFSPEC) {
		t.Errorf("excluded FRN 4 inside RFS: err = %v, want ErrInvalidFSPEC", err)
	}
	// FSPEC FRN 1, 4 | I021/010 | I021/040
	rec, _ = NewRecord(Cat021, u)
	if _, err := rec.Decode(bytes.NewBuffer([]byte{0x90, 0x01, 0x02, 0xAB})); err != nil {
		t.Errorf("excluded FRN 4 at its FSPEC position: %v", err)
	}
}

// The RFS field sits at the FSPEC position of its RS bit (FRN 2 here,
// between FRN 1 and FRN 3); its items join the record and the FSPEC is
// rebuilt to the standard form.
func TestRecordDecodeRFS(t *testing.T) {
	uap := rfsUAP(t)
	// FSPEC FRN 1, 2 (RS), 3 | I021/010 | RFS: N=1, FRN 4, I021/040 | I021/030
	data := []byte{0xE0, 0x01, 0x02, 0x01, 0x04, 0xAB, 0x03, 0x04, 0x05}
	rec, _ := NewRecord(Cat021, uap)
	n, err := rec.Decode(bytes.NewBuffer(data))
	if err != nil || n != len(data) {
		t.Fatalf("n=%d err=%v, want n=%d", n, err, len(data))
	}
	want := map[string][]byte{"I021/010": {1, 2}, "I021/030": {3, 4, 5}, "I021/040": {0xAB}}
	for id, b := range want {
		it, ok := rec.GetDataItem(id)
		if !ok || !bytes.Equal(it.(*MockDataItem).data, b) {
			t.Errorf("%s = %v, want % X", id, it, b)
		}
	}
	if rec.FSPEC().GetFRN(2) || !rec.FSPEC().GetFRN(4) {
		t.Errorf("FSPEC %v: want RS bit cleared and FRN 4 set", rec.FSPEC())
	}
	var buf bytes.Buffer
	if _, err := rec.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.Bytes(), []byte{0xB0, 0x01, 0x02, 0x03, 0x04, 0x05, 0xAB}; !bytes.Equal(got, want) {
		t.Errorf("re-encoded % X, want % X", got, want)
	}

	// N = 0 is accepted.
	rec, _ = NewRecord(Cat021, uap)
	if _, err := rec.Decode(bytes.NewBuffer([]byte{0xC0, 0x01, 0x02, 0x00})); err != nil {
		t.Errorf("RFS with N=0: %v", err)
	}
}

func TestRecordDecodeRFSErrors(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"truncated count", []byte{0xC0, 0x01, 0x02}},
		{"truncated FRN", []byte{0xC0, 0x01, 0x02, 0x01}},
		{"truncated item", []byte{0xC0, 0x01, 0x02, 0x01, 0x03, 0x01}},
		{"unknown FRN", []byte{0xC0, 0x01, 0x02, 0x01, 0x05, 0x00}},
		{"RS FRN inside RFS", []byte{0xC0, 0x01, 0x02, 0x01, 0x02, 0x00}},
		{"FRN also in FSPEC", []byte{0xE0, 0x01, 0x02, 0x01, 0x03, 0x03, 0x04, 0x05, 0x03, 0x04, 0x05}},
		{"FRN twice in RFS", []byte{0xC0, 0x01, 0x02, 0x02, 0x04, 0xAA, 0x04, 0xBB}},
	}
	for _, tc := range cases {
		rec, _ := NewRecord(Cat021, rfsUAP(t))
		if _, err := rec.Decode(bytes.NewBuffer(tc.data)); err == nil {
			t.Errorf("%s: want error", tc.name)
		}
	}
}
