package cat023_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat023"
	v13 "github.com/davidkohl/gobelix/cat/cat023/dataitems/v13"
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

var statsWire = []byte{0x02, 0x03, 0x80, 0x00, 0x01, 0x00, 0x00, 0x15, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}

func TestCat023ItemsKnownBytes(t *testing.T) {
	runItemCases(t, []itemCase{
		{"I023/000", []byte{0x02}, &v13.ReportType{ReportType: 2},
			func() asterix.DataItem { return &v13.ReportType{} }},
		{"I023/010", []byte{0x19, 0x05}, &v13.DataSourceIdentifier{SAC: 25, SIC: 5},
			func() asterix.DataItem { return &v13.DataSourceIdentifier{} }},
		{"I023/015", []byte{0x52}, &v13.ServiceTypeIdentification{SID: 5, STYPE: 2},
			func() asterix.DataItem { return &v13.ServiceTypeIdentification{} }},
		{"I023/070", []byte{0x54, 0x60, 0x00}, &v13.TimeOfDay{Time: 43200},
			func() asterix.DataItem { return &v13.TimeOfDay{} }},
		{"I023/100 with GSSP extension", []byte{0xA5, 0x14},
			&v13.GroundStationStatus{
				NOGO: &v13.NOGOStatus{NOGO: 1}, ODP: &v13.ODPStatus{}, OXT: &v13.OXTStatus{OXT: 1},
				MSC: &v13.MSCStatus{}, TSV: &v13.TSVStatus{}, SPO: &v13.SPOStatus{SPO: 1}, RN: &v13.RNStatus{},
				GSSP: &v13.GSSPValue{GSSP: 10},
			},
			func() asterix.DataItem { return &v13.GroundStationStatus{} }},
		{"I023/101 with SSRP extension", []byte{0x0A, 0x21, 0x0A},
			&v13.ServiceConfiguration{RP: 10, SC: 1, SSRP: 5},
			func() asterix.DataItem { return &v13.ServiceConfiguration{} }},
		{"I023/110 normal", []byte{0x08}, &v13.ServiceStatus{STAT: &v13.STATValue{STAT: 4}},
			func() asterix.DataItem { return &v13.ServiceStatus{} }},
		{"I023/110 initialisation", []byte{0x0A}, &v13.ServiceStatus{STAT: &v13.STATValue{STAT: 5}},
			func() asterix.DataItem { return &v13.ServiceStatus{} }},
		{"I023/120 two counters", statsWire,
			&v13.ServiceStatistics{
				Entries: []v13.ServiceStatisticsEntry{{Type: 3, Ref: true, Counter: 65536}, {Type: 21, Counter: 0xFFFFFFFF}},
				TYPE:    &v13.TYPEValue{TYPE: 3}, REF: &v13.REFValue{REF: true},
				CV: []v13.CounterValue{{Counter: 65536}, {Counter: 0xFFFFFFFF}},
			},
			func() asterix.DataItem { return &v13.ServiceStatistics{} }},
		{"I023/200", []byte{0xC8}, &v13.OperationalRange{Range: 200},
			func() asterix.DataItem { return &v13.OperationalRange{} }},
		{"RE023", []byte{0x02, 0xAB}, &v13.ReservedExpansion{Data: []byte{0x02, 0xAB}},
			func() asterix.DataItem { return &v13.ReservedExpansion{} }},
		{"SP023 empty", []byte{0x01}, &v13.SpecialPurpose{Data: []byte{0x01}},
			func() asterix.DataItem { return &v13.SpecialPurpose{} }},
	})
}

// I023/120 needs at least one counter block (ed 1.3 §5.2.8).
func TestCat023ServiceStatisticsREPZeroFails(t *testing.T) {
	if _, err := new(v13.ServiceStatistics).Decode(bytes.NewBuffer([]byte{0x00})); err == nil {
		t.Error("I023/120 REP 0 decoded")
	}
}

func TestCat023ServiceStatisticsLegacyEncode(t *testing.T) {
	// Callers that fill only TYPE/REF/CV get one block per counter.
	s := &v13.ServiceStatistics{TYPE: &v13.TYPEValue{TYPE: 3}, REF: &v13.REFValue{REF: true},
		CV: []v13.CounterValue{{Counter: 65536}}}
	var buf bytes.Buffer
	if _, err := s.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	if want := []byte{0x01, 0x03, 0x80, 0x00, 0x01, 0x00, 0x00}; !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("got % X, want % X", buf.Bytes(), want)
	}
}

func TestCat023ItemValidation(t *testing.T) {
	for name, item := range map[string]asterix.DataItem{
		"110 STAT 8":      &v13.ServiceStatus{STAT: &v13.STATValue{STAT: 8}},
		"120 256 entries": &v13.ServiceStatistics{Entries: make([]v13.ServiceStatisticsEntry, 256)},
		"120 no counters": &v13.ServiceStatistics{},
		"120 TYPE disagrees": &v13.ServiceStatistics{Entries: []v13.ServiceStatisticsEntry{{Type: 3}},
			TYPE: &v13.TYPEValue{TYPE: 4}},
		"120 REF disagrees": &v13.ServiceStatistics{Entries: []v13.ServiceStatisticsEntry{{Ref: true}},
			REF: &v13.REFValue{REF: false}},
		"120 CV count disagrees": &v13.ServiceStatistics{Entries: []v13.ServiceStatisticsEntry{{Counter: 1}},
			CV: []v13.CounterValue{{Counter: 1}, {Counter: 2}}},
		"120 CV value disagrees": &v13.ServiceStatistics{Entries: []v13.ServiceStatisticsEntry{{Counter: 1}},
			CV: []v13.CounterValue{{Counter: 2}}},
		"RE LEN mismatch": &v13.ReservedExpansion{Data: []byte{0x03, 0x00}},
		"SP LEN mismatch": &v13.SpecialPurpose{Data: []byte{0x01, 0x00}},
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
	// Ground station status: FRN 1, 2, 4, 5, 7.
	cat023GSStatus = []byte{0xDA, 0x19, 0x05, 0x01, 0x54, 0x60, 0x00, 0xA5, 0x14, 0xC8}
	// Service status: FRN 1, 2, 3, 4, 6, 8 (110), 13 (RE), 14 (SP).
	cat023ServiceStatus = []byte{0xF5, 0x86, 0x19, 0x05, 0x02, 0x52, 0x54, 0x60, 0x00, 0x0A, 0x21, 0x0A, 0x08, 0x02, 0xAB, 0x01}
	// Service statistics: FRN 1, 2, 3, 4, 9 (120).
	cat023Statistics = append([]byte{0xF1, 0x40, 0x19, 0x05, 0x03, 0x52, 0x54, 0x60, 0x00}, statsWire...)
)

func newCat023Decoder(t testing.TB) *asterix.Decoder {
	u, err := cat023.NewUAP(cat023.Version13)
	if err != nil {
		t.Fatal(err)
	}
	return asterix.NewDecoder(asterix.WithPreloadedUAPs(u))
}

func TestCat023RecordsUseStandardUAPFRNs(t *testing.T) {
	data := block(23, cat023GSStatus, cat023ServiceStatus, cat023Statistics)
	db, err := newCat023Decoder(t).Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	recs := db.Records()
	if len(recs) != 3 {
		t.Fatalf("got %d records", len(recs))
	}
	if it, ok := recs[0].GetDataItem("I023/100"); !ok || it.(*v13.GroundStationStatus).GSSP.GSSP != 10 {
		t.Errorf("I023/100 = %+v", it)
	}
	if it, ok := recs[1].GetDataItem("I023/110"); !ok || it.(*v13.ServiceStatus).STAT.STAT != 4 {
		t.Errorf("I023/110 (FRN 8) = %+v", it)
	}
	if it, ok := recs[1].GetDataItem("RE023"); !ok || !bytes.Equal(it.(*v13.ReservedExpansion).Data, []byte{0x02, 0xAB}) {
		t.Errorf("RE023 (FRN 13) = %+v", it)
	}
	if !recs[1].HasDataItem("SP023") {
		t.Error("SP023 (FRN 14) missing")
	}
	st, ok := recs[2].GetDataItem("I023/120")
	if !ok || !reflect.DeepEqual(st.(*v13.ServiceStatistics).Entries,
		[]v13.ServiceStatisticsEntry{{Type: 3, Ref: true, Counter: 65536}, {Type: 21, Counter: 0xFFFFFFFF}}) {
		t.Errorf("I023/120 (FRN 9) = %+v", st)
	}
	out, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("re-encode:\n got % X\nwant % X", out, data)
	}
}

func TestCat023MalformedRecordsFail(t *testing.T) {
	d := newCat023Decoder(t)
	for _, rec := range [][]byte{cat023GSStatus, cat023ServiceStatus, cat023Statistics} {
		for i := 1; i < len(rec); i++ {
			if _, err := d.Decode(block(23, rec[:i])); err == nil {
				t.Errorf("truncated record % X decoded without error", rec[:i])
			}
		}
	}
	for name, rec := range map[string][]byte{
		"RE LEN 0":         {0x01, 0x04, 0x00},
		"SP LEN too long":  {0x01, 0x02, 0x04, 0x00},
		"120 REP too long": {0x01, 0x40, 0x02, 0x03, 0x80, 0x00, 0x01, 0x00, 0x00},
		"110 FX dangling":  {0x01, 0x80, 0x01},
	} {
		if _, err := d.Decode(block(23, rec)); err == nil {
			t.Errorf("%s: decoded without error", name)
		}
	}
}

func FuzzCat023Decode(f *testing.F) {
	f.Add(block(23, cat023GSStatus))
	f.Add(block(23, cat023ServiceStatus))
	f.Add(block(23, cat023Statistics))
	d := newCat023Decoder(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		db, err := d.Decode(data)
		if err != nil {
			return
		}
		_, _ = db.Encode()
	})
}
