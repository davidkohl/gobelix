package cat002_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat002"
	v10 "github.com/davidkohl/gobelix/cat/cat002/dataitems/v10"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// itemCase is a known-byte vector for one data item: wire decodes to want,
// and want encodes back to wire.
type itemCase struct {
	name string
	wire []byte
	want asterix.DataItem
	newF func() asterix.DataItem
}

func runItemCases(t *testing.T, cases []itemCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.newF()
			n, err := got.Decode(bytes.NewBuffer(append([]byte(nil), c.wire...)))
			if err != nil {
				t.Fatalf("decode % X: %v", c.wire, err)
			}
			if n != len(c.wire) {
				t.Fatalf("decode consumed %d of %d octets", n, len(c.wire))
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("decode % X:\n got %+v\nwant %+v", c.wire, got, c.want)
			}
			var buf bytes.Buffer
			n, err = c.want.Encode(&buf)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if n != buf.Len() || !bytes.Equal(buf.Bytes(), c.wire) {
				t.Fatalf("encode: got % X (n=%d), want % X", buf.Bytes(), n, c.wire)
			}
			// Every strict prefix is a short buffer and must fail, not panic.
			for i := 0; i < len(c.wire); i++ {
				if _, err := c.newF().Decode(bytes.NewBuffer(append([]byte(nil), c.wire[:i]...))); err == nil {
					t.Errorf("decode of %d-octet prefix succeeded", i)
				}
			}
		})
	}
}

func TestCat002ItemsKnownBytes(t *testing.T) {
	runItemCases(t, []itemCase{
		{"I002/000", []byte{0x08}, &v10.MessageType{MessageType: 8},
			func() asterix.DataItem { return &v10.MessageType{} }},
		{"I002/010", []byte{0x19, 0x05}, &common.DataSourceIdentifier{SAC: 25, SIC: 5},
			func() asterix.DataItem { return &common.DataSourceIdentifier{} }},
		{"I002/020", []byte{0x40}, &common.SectorNumber{SectorNumber: 90},
			func() asterix.DataItem { return &common.SectorNumber{} }},
		{"I002/030", []byte{0x54, 0x60, 0x00}, &common.TimeOfDay{Time: 43200},
			func() asterix.DataItem { return &common.TimeOfDay{} }},
		{"I002/041", []byte{0x02, 0x00}, &v10.AntennaRotationSpeed{RotationPeriod: 4},
			func() asterix.DataItem { return &v10.AntennaRotationSpeed{} }},
		{"I002/050 two parts", []byte{0x03, 0x04}, &v10.StationConfigurationStatus{Status: 1, Parts: []uint8{1, 2}},
			func() asterix.DataItem { return &v10.StationConfigurationStatus{} }},
		{"I002/060 one part", []byte{0xFE}, &v10.StationProcessingMode{Mode: 0x7F, Parts: []uint8{0x7F}},
			func() asterix.DataItem { return &v10.StationProcessingMode{} }},
		{"I002/070 two counters", []byte{0x02, 0x84, 0x05, 0x0B, 0xFF},
			&v10.PlotCountValues{Values: []v10.PlotCountValue{{A: true, Ident: 1, Counter: 5}, {Ident: 2, Counter: 1023}}},
			func() asterix.DataItem { return &v10.PlotCountValues{} }},
		{"I002/080 two W/E values", []byte{0x81, 0x02}, &v10.WarningErrorConditions{Values: []uint8{64, 1}},
			func() asterix.DataItem { return &v10.WarningErrorConditions{} }},
		{"I002/090 negative", []byte{0xFF, 0x80}, &v10.CollimationError{RangeError: -1.0 / 128, AzimuthError: -2.8125},
			func() asterix.DataItem { return &v10.CollimationError{} }},
		{"I002/090 positive", []byte{0x7F, 0x01}, &v10.CollimationError{RangeError: 127.0 / 128, AzimuthError: 360.0 / 16384},
			func() asterix.DataItem { return &v10.CollimationError{} }},
		{"I002/100", []byte{0x00, 0x80, 0x40, 0x00, 0x40, 0x00, 0x80, 0x00},
			&v10.DynamicWindow{RhoStart: 1, RhoEnd: 128, ThetaStart: 90, ThetaEnd: 180},
			func() asterix.DataItem { return &v10.DynamicWindow{} }},
		{"SP002", []byte{0x02, 0x55}, &common.ExplicitField{Payload: []byte{0x55}},
			func() asterix.DataItem { return &common.ExplicitField{} }},
	})
}

func TestCat002LegacyFirstPartEncode(t *testing.T) {
	var buf bytes.Buffer
	if _, err := (&v10.StationConfigurationStatus{Status: 0x55}).Encode(&buf); err != nil || !bytes.Equal(buf.Bytes(), []byte{0xAA}) {
		t.Fatalf("Status-only encode: % X, %v", buf.Bytes(), err)
	}
}

func TestCat002ItemValidation(t *testing.T) {
	for name, item := range map[string]asterix.DataItem{
		"080 empty":       &v10.WarningErrorConditions{},
		"080 8-bit value": &v10.WarningErrorConditions{Values: []uint8{0x80}},
		"050 8-bit part":  &v10.StationConfigurationStatus{Parts: []uint8{0x80}},
		"090 range":       &v10.CollimationError{RangeError: 1.0},
		"090 azimuth":     &v10.CollimationError{AzimuthError: -3},
		"100 rho":         &v10.DynamicWindow{RhoEnd: 512},
		"100 negative":    &v10.DynamicWindow{ThetaStart: -1},
	} {
		if _, err := item.Encode(new(bytes.Buffer)); err == nil {
			t.Errorf("%s: encode accepted an invalid value", name)
		}
	}
}

func block(cat byte, records ...[]byte) []byte {
	body := bytes.Join(records, nil)
	n := len(body) + 3
	return append([]byte{cat, byte(n >> 8), byte(n)}, body...)
}

// Blind zone activation: FRN 1, 2, 9 (100), 10 (090), 11 (080), 13 (SP).
var cat002Record = []byte{
	0xC1, 0x74,
	0x19, 0x05,
	0x08,
	0x00, 0x80, 0x40, 0x00, 0x40, 0x00, 0x80, 0x00,
	0xFF, 0x80,
	0x81, 0x02,
	0x02, 0x55,
}

func newCat002Decoder(t testing.TB) *asterix.Decoder {
	u, err := cat002.NewUAP(cat002.Version10)
	if err != nil {
		t.Fatal(err)
	}
	return asterix.NewDecoder(asterix.WithPreloadedUAPs(u))
}

func TestCat002RecordWithAllNewItems(t *testing.T) {
	data := block(2, cat002Record)
	db, err := newCat002Decoder(t).Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	rec := db.Records()[0]
	if it, ok := rec.GetDataItem("I002/080"); !ok || !reflect.DeepEqual(it.(*v10.WarningErrorConditions).Values, []uint8{64, 1}) {
		t.Errorf("I002/080 = %+v", it)
	}
	if it, ok := rec.GetDataItem("I002/090"); !ok || it.(*v10.CollimationError).AzimuthError != -2.8125 {
		t.Errorf("I002/090 = %+v", it)
	}
	if it, ok := rec.GetDataItem("I002/100"); !ok || it.(*v10.DynamicWindow).ThetaEnd != 180 {
		t.Errorf("I002/100 = %+v", it)
	}
	if it, ok := rec.GetDataItem("SP002"); !ok || !bytes.Equal(it.(*common.ExplicitField).Payload, []byte{0x55}) {
		t.Errorf("SP002 = %+v", it)
	}
	out, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("re-encode:\n got % X\nwant % X", out, data)
	}
}

func TestCat002MalformedRecordsFail(t *testing.T) {
	d := newCat002Decoder(t)
	for i := 1; i < len(cat002Record); i++ {
		if _, err := d.Decode(block(2, cat002Record[:i])); err == nil {
			t.Errorf("truncated record % X decoded without error", cat002Record[:i])
		}
	}
	for name, rec := range map[string][]byte{
		"080 FX dangling":  {0x01, 0x10, 0x01},
		"070 REP too long": {0x01, 0x80, 0x03, 0x84, 0x05},
		"SP LEN 0":         {0x01, 0x04, 0x00},
	} {
		if _, err := d.Decode(block(2, rec)); err == nil {
			t.Errorf("%s: decoded without error", name)
		}
	}
}

func FuzzCat002Decode(f *testing.F) {
	f.Add(block(2, cat002Record))
	f.Add(block(2, []byte{0xF0, 0x19, 0x05, 0x01, 0x40, 0x54, 0x60, 0x00}))
	d := newCat002Decoder(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		db, err := d.Decode(data)
		if err != nil {
			return
		}
		_, _ = db.Encode()
	})
}

// FRN 14 is the RS bit (ed 1.2 §6): the RFS field is read and its items
// join the record; FRN 12 is spare (§5.3.2.13 note 1) and fails.
func TestCat002RFSFieldAndSpareBit(t *testing.T) {
	d := newCat002Decoder(t)
	// FRN 1, 2, 14 | 010 | 000 = 2 (sector crossing) | RFS N=1: FRN 3 (020)
	rfs := []byte{0xC1, 0x02, 0x19, 0x05, 0x02, 0x01, 0x03, 0x40}
	db, err := d.Decode(block(2, rfs))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if it, ok := db.Records()[0].GetDataItem("I002/020"); !ok || it.(*v10.SectorNumber).SectorNumber != 90 {
		t.Errorf("I002/020 = %+v", it)
	}
	out, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if want := block(2, []byte{0xE0, 0x19, 0x05, 0x02, 0x40}); !bytes.Equal(out, want) {
		t.Errorf("re-encode\n got % X\nwant % X", out, want)
	}

	_, err = d.Decode(block(2, []byte{0xC1, 0x08, 0x19, 0x05, 0x02}))
	if !errors.Is(err, asterix.ErrInvalidFSPEC) {
		t.Errorf("spare FRN 12: err = %v, want ErrInvalidFSPEC", err)
	}
}

// Legacy and new views of one item must agree when both are set; I002/070
// needs at least one counter (ed 1.2 §5.2.8).
func TestCat002LegacyAndNewFieldsMustAgreeAndREP(t *testing.T) {
	for name, it := range map[string]asterix.DataItem{
		"050 Status vs Parts": &v10.StationConfigurationStatus{Status: 1, Parts: []uint8{2}},
		"060 Mode vs Parts":   &v10.StationProcessingMode{Mode: 1, Parts: []uint8{2, 3}},
		"070 no counters":     &v10.PlotCountValues{},
	} {
		if _, err := it.Encode(new(bytes.Buffer)); err == nil {
			t.Errorf("%s: encoded", name)
		}
	}
	for name, it := range map[string]asterix.DataItem{
		"050 agree":      &v10.StationConfigurationStatus{Status: 2, Parts: []uint8{2, 1}},
		"060 Parts only": &v10.StationProcessingMode{Parts: []uint8{2}},
	} {
		if _, err := it.Encode(new(bytes.Buffer)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := new(v10.PlotCountValues).Decode(bytes.NewBuffer([]byte{0x00})); err == nil {
		t.Error("I002/070 REP 0 decoded")
	}
}
