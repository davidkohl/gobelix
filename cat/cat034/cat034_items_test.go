package cat034_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat034"
	"github.com/davidkohl/gobelix/cat/cat034/dataitems/v129"
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

func u8(v uint8) *uint8    { return &v }
func u16(v uint16) *uint16 { return &v }

func TestCat034ItemsKnownBytes(t *testing.T) {
	runItemCases(t, []itemCase{
		{"I034/000 solar storm", []byte{0x05}, &v129.MessageType{MessageType: 5},
			func() asterix.DataItem { return v129.NewMessageType() }},
		{"I034/010", []byte{0x19, 0x05}, &common.DataSourceIdentifier{SAC: 25, SIC: 5},
			func() asterix.DataItem { return &common.DataSourceIdentifier{} }},
		{"I034/020", []byte{0x40}, &common.SectorNumber{SectorNumber: 90},
			func() asterix.DataItem { return &common.SectorNumber{} }},
		{"I034/030", []byte{0x54, 0x60, 0x00}, &common.TimeOfDay{Time: 43200},
			func() asterix.DataItem { return &common.TimeOfDay{} }},
		{"I034/041", []byte{0x02, 0x00}, &v129.AntennaRotationPeriod{Period: 4},
			func() asterix.DataItem { return v129.NewAntennaRotationPeriod() }},
		{"I034/050 all subfields", []byte{0x9C, 0xFE, 0xD8, 0x20, 0xFF, 0x80},
			&v129.SystemConfigurationStatus{COM: 0xFE, PSR: u8(0xD8), SSR: u8(0x20), MDS: u16(0xFF80)},
			func() asterix.DataItem { return v129.NewSystemConfigurationStatus() }},
		{"I034/050 PSR only", []byte{0x10, 0x00},
			&v129.SystemConfigurationStatus{OmitCOM: true, PSR: u8(0)},
			func() asterix.DataItem { return v129.NewSystemConfigurationStatus() }},
		{"I034/060 all subfields", []byte{0x9C, 0x7E, 0xFC, 0xE0, 0xF0},
			&v129.SystemProcessingMode{COM: u8(0x7E), PSR: u8(0xFC), SSR: u8(0xE0), MDS: u8(0xF0)},
			func() asterix.DataItem { return v129.NewSystemProcessingMode() }},
		{"I034/060 empty", []byte{0x00}, &v129.SystemProcessingMode{},
			func() asterix.DataItem { return v129.NewSystemProcessingMode() }},
		{"I034/070", []byte{0x01, 0xA7, 0xFF},
			&v129.MessageCountValues{Values: []v129.MessageCountValue{{TYP: 20, Counter: 2047}}},
			func() asterix.DataItem { return &v129.MessageCountValues{} }},
		{"I034/090", []byte{0x80, 0x7F},
			&v129.CollimationError{RangeError: -1, AzimuthError: 127 * 360.0 / 16384},
			func() asterix.DataItem { return &v129.CollimationError{} }},
		{"I034/100", []byte{0x00, 0x00, 0x80, 0x00, 0x40, 0x00, 0xC0, 0x00},
			&v129.GenericPolarWindow{RhoStart: 0, RhoEnd: 128, ThetaStart: 90, ThetaEnd: 270},
			func() asterix.DataItem { return &v129.GenericPolarWindow{} }},
		{"I034/110", []byte{0x02}, &v129.DataFilter{TYP: 2},
			func() asterix.DataItem { return &v129.DataFilter{} }},
		{"I034/120", []byte{0xFF, 0x9C, 0x20, 0x00, 0x00, 0xE0, 0x00, 0x00},
			&v129.DataSourcePosition{Height: -100, Latitude: 45, Longitude: -45},
			func() asterix.DataItem { return v129.NewDataSourcePosition() }},
		{"RE empty", []byte{0x01}, &common.ExplicitField{Payload: []byte{}},
			func() asterix.DataItem { return &common.ExplicitField{} }},
		{"SP", []byte{0x03, 0x01, 0x02}, &common.ExplicitField{Payload: []byte{0x01, 0x02}},
			func() asterix.DataItem { return &common.ExplicitField{} }},
	})
}

func TestCat034CompoundSubfieldBits(t *testing.T) {
	s := &v129.SystemConfigurationStatus{COM: 0xFE, PSR: u8(0xD8), SSR: u8(0x20), MDS: u16(0xFF80)}
	if c, ok := s.Common(); !ok || c != (v129.SystemStatusCOM{NOGO: true, RDPC: true, RDPR: true, OVLRDP: true, OVLXMT: true, MSC: true, TSV: true}) {
		t.Errorf("COM = %+v, %v", c, ok)
	}
	if p, ok := s.PSRStatus(); !ok || p != (v129.SensorStatus{ANT: true, CHAB: 2, OVL: true, MSC: true}) {
		t.Errorf("PSR = %+v", p)
	}
	if p, ok := s.SSRStatus(); !ok || p != (v129.SensorStatus{CHAB: 1}) {
		t.Errorf("SSR = %+v", p)
	}
	want := v129.ModeSStatus{ANT: true, CHAB: 3, OVLSUR: true, MSC: true, SCF: true, DLF: true, OVLSCF: true, OVLDLF: true}
	if m, ok := s.MDSStatus(); !ok || m != want {
		t.Errorf("MDS = %+v", m)
	}
	if want.Value() != 0xFF80 || (v129.SensorStatus{ANT: true, CHAB: 2, OVL: true, MSC: true}).Octet() != 0xD8 {
		t.Error("status encoders do not invert the parsers")
	}

	p := &v129.SystemProcessingMode{COM: u8(0x7E), PSR: u8(0xFC), SSR: u8(0xE0), MDS: u8(0xF0)}
	if c, _ := p.Common(); c != (v129.ProcessingCOM{REDRDP: 7, REDXMT: 7}) || c.Octet() != 0x7E {
		t.Errorf("060 COM = %+v", c)
	}
	if m, _ := p.PSRMode(); m != (v129.PSRProcessing{POL: true, REDRAD: 7, STC: 3}) || m.Octet() != 0xFC {
		t.Errorf("060 PSR = %+v", m)
	}
	if m, _ := p.SSRMode(); m != (v129.SSRProcessing{REDRAD: 7}) || m.Octet() != 0xE0 {
		t.Errorf("060 SSR = %+v", m)
	}
	if m, _ := p.MDSMode(); m != (v129.MDSProcessing{REDRAD: 7, CLU: true}) || m.Octet() != 0xF0 {
		t.Errorf("060 MDS = %+v", m)
	}
	if _, ok := (&v129.SystemProcessingMode{}).PSRMode(); ok {
		t.Error("absent PSR reported present")
	}
}

func TestCat034CompoundRejectsUndefinedSubfields(t *testing.T) {
	for name, wire := range map[string][]byte{
		"spare bit 7":           {0x40, 0x00},
		"spare bit 2":           {0x02, 0x00},
		"extension with a bit":  {0x81, 0x80, 0x00},
		"extension FX dangling": {0x01},
	} {
		for _, it := range []asterix.DataItem{v129.NewSystemConfigurationStatus(), v129.NewSystemProcessingMode()} {
			if _, err := it.Decode(bytes.NewBuffer(wire)); err == nil {
				t.Errorf("%T %s: decoded without error", it, name)
			}
		}
	}
	// An empty extension octet is harmless.
	it := v129.NewSystemProcessingMode()
	if n, err := it.Decode(bytes.NewBuffer([]byte{0x81, 0x00, 0x07})); err != nil || n != 3 || *it.COM != 0x07 {
		t.Errorf("empty extension octet: n=%d err=%v", n, err)
	}
}

func TestCat034ItemValidation(t *testing.T) {
	for name, item := range map[string]asterix.DataItem{
		"000 type 8":    &v129.MessageType{MessageType: 8},
		"041 too long":  &v129.AntennaRotationPeriod{Period: 512},
		"090 range":     &v129.CollimationError{RangeError: 1},
		"100 rho 256":   &v129.GenericPolarWindow{RhoEnd: 256},
		"120 height":    &v129.DataSourcePosition{Height: 40000},
		"120 longitude": &v129.DataSourcePosition{Longitude: 180},
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

var (
	// North marker: FRN 1, 2, 3, 5, 6, 7, 8 (070), 11 (120), 12 (090), 13 (RE), 14 (SP).
	cat034NorthMarker = []byte{
		0xEF, 0x9E,
		0x19, 0x05,
		0x01,
		0x54, 0x60, 0x00,
		0x02, 0x00,
		0x9C, 0xFE, 0xD8, 0x20, 0xFF, 0x80,
		0x9C, 0x7E, 0xFC, 0xE0, 0xF0,
		0x01, 0xA7, 0xFF,
		0xFF, 0x9C, 0x20, 0x00, 0x00, 0xE0, 0x00, 0x00,
		0x80, 0x7F,
		0x01,
		0x03, 0x01, 0x02,
	}
	// Geographical filtering: FRN 1, 2, 3, 9 (100), 10 (110).
	cat034GeoFilter = []byte{
		0xE1, 0x60,
		0x19, 0x05,
		0x03,
		0x54, 0x60, 0x00,
		0x00, 0x00, 0x80, 0x00, 0x40, 0x00, 0xC0, 0x00,
		0x02,
	}
)

func newCat034Decoder(t testing.TB) *asterix.Decoder {
	u, err := cat034.NewUAP(cat034.Version129)
	if err != nil {
		t.Fatal(err)
	}
	return asterix.NewDecoder(asterix.WithPreloadedUAPs(u))
}

func TestCat034RecordsWithAllItems(t *testing.T) {
	data := block(34, cat034NorthMarker, cat034GeoFilter)
	db, err := newCat034Decoder(t).Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	recs := db.Records()
	if len(recs) != 2 {
		t.Fatalf("got %d records", len(recs))
	}
	nm := recs[0]
	for _, id := range []string{"I034/010", "I034/000", "I034/030", "I034/041", "I034/050", "I034/060", "I034/070", "I034/120", "I034/090", "RE", "SP"} {
		if !nm.HasDataItem(id) {
			t.Errorf("north marker: %s missing", id)
		}
	}
	if it, _ := nm.GetDataItem("I034/120"); !reflect.DeepEqual(it, &v129.DataSourcePosition{Height: -100, Latitude: 45, Longitude: -45}) {
		t.Errorf("I034/120 = %+v", it)
	}
	gf := recs[1]
	if it, ok := gf.GetDataItem("I034/100"); !ok || it.(*v129.GenericPolarWindow).ThetaEnd != 270 {
		t.Errorf("I034/100 = %+v", it)
	}
	if it, ok := gf.GetDataItem("I034/110"); !ok || it.(*v129.DataFilter).TYP != 2 {
		t.Errorf("I034/110 = %+v", it)
	}
	out, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("re-encode:\n got % X\nwant % X", out, data)
	}
}

func TestCat034MalformedRecordsFail(t *testing.T) {
	d := newCat034Decoder(t)
	for _, rec := range [][]byte{cat034NorthMarker, cat034GeoFilter} {
		for i := 1; i < len(rec); i++ {
			if _, err := d.Decode(block(34, rec[:i])); err == nil {
				t.Errorf("truncated record % X decoded without error", rec[:i])
			}
		}
	}
	for name, rec := range map[string][]byte{
		"050 spare subfield": {0x04, 0x40, 0x00},
		"RE LEN 0":           {0x01, 0x04, 0x00},
		"SP LEN too long":    {0x01, 0x02, 0x05, 0x00},
	} {
		if _, err := d.Decode(block(34, rec)); err == nil {
			t.Errorf("%s: decoded without error", name)
		}
	}
}

func FuzzCat034Decode(f *testing.F) {
	f.Add(block(34, cat034NorthMarker))
	f.Add(block(34, cat034GeoFilter))
	f.Add(block(34, cat034NorthMarker, cat034GeoFilter))
	d := newCat034Decoder(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		db, err := d.Decode(data)
		if err != nil {
			return
		}
		_, _ = db.Encode()
	})
}
