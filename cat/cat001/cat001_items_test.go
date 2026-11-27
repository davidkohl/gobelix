package cat001_test

import (
	"bytes"
	"math/bits"
	"reflect"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat001/dataitems/v12"
	"github.com/davidkohl/gobelix/cat/cat001/uap"
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

func TestCat001ItemsKnownBytes(t *testing.T) {
	runItemCases(t, []itemCase{
		{"I001/020 plot first part", []byte{0x6C},
			&v12.TargetReportDescriptor{SIM: true, SSR: true, ANT: true, SPI: true},
			func() asterix.DataItem { return &v12.TargetReportDescriptor{} }},
		{"I001/020 track with extent", []byte{0xB3, 0xB8},
			&v12.TargetReportDescriptor{TYP: v12.TYPTrack, SSR: true, PSR: true, RAB: true, TST: true, DS1DS2: 1, ME: true, MI: true},
			func() asterix.DataItem { return &v12.TargetReportDescriptor{} }},
		{"I001/030 two values", []byte{0x03, 0x04},
			&v12.WarningErrorConditions{W: 1, Values: []uint8{1, 2}},
			func() asterix.DataItem { return &v12.WarningErrorConditions{} }},
		{"I001/040", []byte{0x19, 0x40, 0x40, 0x00},
			&v12.PositionPolar{RHO: 50.5, THETA: 90},
			func() asterix.DataItem { return &v12.PositionPolar{} }},
		{"I001/042", []byte{0xFF, 0xC0, 0x00, 0x40},
			&v12.PositionCartesian{X: -1, Y: 1},
			func() asterix.DataItem { return &v12.PositionCartesian{} }},
		{"I001/050", []byte{0xE5, 0x55},
			&v12.Mode2Code{V: true, G: true, L: true, Code: 0o2525},
			func() asterix.DataItem { return &v12.Mode2Code{} }},
		{"I001/060", []byte{0x0A, 0x81},
			&v12.Mode2CodeConfidence{QA4: true, QA1: true, QB2: true, QD1: true},
			func() asterix.DataItem { return &v12.Mode2CodeConfidence{} }},
		{"I001/070", []byte{0x0F, 0xFF},
			&v12.Mode3ACode{Mode: 0o7777},
			func() asterix.DataItem { return &v12.Mode3ACode{} }},
		{"I001/080", []byte{0x0A, 0x81},
			&v12.Mode3ACodeConfidence{QA4: true, QA1: true, QB2: true, QD1: true},
			func() asterix.DataItem { return &v12.Mode3ACodeConfidence{} }},
		{"I001/090 negative", []byte{0x3F, 0xFC},
			&v12.ModeCCode{FlightLevel: -1},
			func() asterix.DataItem { return &v12.ModeCCode{} }},
		{"I001/100 -1200 ft (C4 only)", []byte{0x80, 0x80, 0x0F, 0xFF},
			&v12.ModeCCodeConfidence{V: true, GrayCode: 0x080, FlightLevel: -12, AltitudeValid: true,
				QC1: true, QA1: true, QC2: true, QA2: true, QC4: true, QA4: true,
				QB1: true, QD1: true, QB2: true, QD2: true, QB4: true, QD4: true},
			func() asterix.DataItem { return &v12.ModeCCodeConfidence{} }},
		{"I001/100 invalid Gray code kept raw", []byte{0x00, 0x10, 0x00, 0x00},
			&v12.ModeCCodeConfidence{GrayCode: 0x010},
			func() asterix.DataItem { return &v12.ModeCCodeConfidence{} }},
		{"I001/120 -1 LSB", []byte{0xFF},
			&v12.MeasuredRadialDopplerSpeed{DopplerSpeed: -14.0625},
			func() asterix.DataItem { return &v12.MeasuredRadialDopplerSpeed{} }},
		{"I001/130 two indicators", []byte{0x05, 0x06},
			&v12.RadarPlotCharacteristics{SRL: 2, SRR: 3, Indicators: []uint8{2, 3}},
			func() asterix.DataItem { return &v12.RadarPlotCharacteristics{} }},
		{"I001/131", []byte{0xF6},
			&v12.ReceivedPower{Power: -10},
			func() asterix.DataItem { return &v12.ReceivedPower{} }},
		{"I001/141", []byte{0x80, 0x00},
			&v12.TruncatedTimeOfDay{Time: 256},
			func() asterix.DataItem { return &v12.TruncatedTimeOfDay{} }},
		{"I001/150", []byte{0xA4},
			&v12.PresenceXPulse{XA: true, XC: true, X2: true},
			func() asterix.DataItem { return &v12.PresenceXPulse{} }},
		{"I001/161", []byte{0x12, 0x34},
			&v12.TrackPlotNumber{Number: 0x1234},
			func() asterix.DataItem { return &v12.TrackPlotNumber{} }},
		{"I001/170 with TRE extent", []byte{0x81, 0x80},
			&v12.TrackStatus{CON: true, TRE: true},
			func() asterix.DataItem { return &v12.TrackStatus{} }},
		{"I001/200", []byte{0x10, 0x00, 0x40, 0x00},
			&v12.TrackVelocity{GroundSpeed: 900, Heading: 90},
			func() asterix.DataItem { return &v12.TrackVelocity{} }},
		{"I001/210 two octets", []byte{0x03, 0x04},
			&v12.TrackQuality{Quality: []uint8{1, 2}},
			func() asterix.DataItem { return &v12.TrackQuality{} }},
		{"SP001", []byte{0x03, 0xAA, 0xBB},
			&common.ExplicitField{Payload: []byte{0xAA, 0xBB}},
			func() asterix.DataItem { return &common.ExplicitField{} }},
	})
}

func TestCat001TargetReportDescriptorLegacyFieldsEncode(t *testing.T) {
	// Callers that only set W / the SRL..APD view keep working.
	var buf bytes.Buffer
	if _, err := (&v12.WarningErrorConditions{W: 5}).Encode(&buf); err != nil || !bytes.Equal(buf.Bytes(), []byte{0x0A}) {
		t.Fatalf("W-only encode: % X, %v", buf.Bytes(), err)
	}
	buf.Reset()
	if _, err := (&v12.RadarPlotCharacteristics{SRL: 2, SAM: -1}).Encode(&buf); err != nil || !bytes.Equal(buf.Bytes(), []byte{0x05, 0x01, 0xFE}) {
		t.Fatalf("legacy I001/130 encode: % X, %v", buf.Bytes(), err)
	}
}

func TestCat001GillhamGrayProperty(t *testing.T) {
	// Every 100 ft step from -1200 ft to 126 700 ft is representable, round
	// trips, and changes exactly one pulse (the Gray property).
	prev, ok := v12.AltitudeToGillham(-1200)
	if !ok || prev != 0x080 {
		t.Fatalf("-1200 ft: code %#x ok=%v, want C4 only (0x080)", prev, ok)
	}
	for alt := -1100; alt <= 126700; alt += 100 {
		code, ok := v12.AltitudeToGillham(alt)
		if !ok {
			t.Fatalf("%d ft not encodable", alt)
		}
		back, ok := v12.GillhamToAltitude(code)
		if !ok || back != alt {
			t.Fatalf("%d ft -> %#x -> %d (ok=%v)", alt, code, back, ok)
		}
		if d := bits.OnesCount16(code ^ prev); d != 1 {
			t.Fatalf("%d ft: %d pulses changed from previous step", alt, d)
		}
		prev = code
	}
	if _, ok := v12.GillhamToAltitude(0x010); ok {
		t.Fatal("D1 set must not decode")
	}
	if _, ok := v12.GillhamToAltitude(0x000); ok {
		t.Fatal("no C pulse must not decode")
	}
}

func TestCat001ModeCCodeConfidenceEncodesFromFlightLevel(t *testing.T) {
	var buf bytes.Buffer
	if _, err := (&v12.ModeCCodeConfidence{FlightLevel: -12}).Encode(&buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{0x00, 0x80, 0x00, 0x00}) {
		t.Fatalf("got % X", buf.Bytes())
	}
	if _, err := (&v12.ModeCCodeConfidence{FlightLevel: 100.5}).Encode(&buf); err == nil {
		t.Fatal("FL 100.5 is not a 100 ft step and must not encode")
	}
}

func TestCat001ItemValidation(t *testing.T) {
	for name, item := range map[string]asterix.DataItem{
		"TYP 2":         &v12.TargetReportDescriptor{TYP: 2},
		"DS1DS2 4":      &v12.TargetReportDescriptor{DS1DS2: 4},
		"W/E 128":       &v12.WarningErrorConditions{Values: []uint8{128}},
		"Mode-2 13 bit": &v12.Mode2Code{Code: 0x1000},
		"Doppler range": &v12.MeasuredRadialDopplerSpeed{DopplerSpeed: 2000},
		"Gray 13 bit":   &v12.ModeCCodeConfidence{GrayCode: 0x1000},
		"SP too long":   &common.ExplicitField{Payload: make([]byte, 255)},
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
	// FRN 1-5: 010, 020 (plot, SSR), 040, 070, 090.
	cat001PlotRecord = []byte{0xF8, 0x32, 0x02, 0x20, 0x19, 0x40, 0x40, 0x00, 0x0F, 0xFF, 0x00, 0x64}
	// FRN 1-4 + 13: 010, 020 (track, SSR+PSR), 161, 040, 170 (CON + TRE).
	cat001TrackRecord = []byte{0xF1, 0x04, 0x32, 0x02, 0xB0, 0x12, 0x34, 0x19, 0x40, 0x40, 0x00, 0x81, 0x80}
	// FRN 1, 2 and 20: 010, 020 (plot), SP with two payload octets.
	cat001SPRecord = []byte{0xC1, 0x01, 0x04, 0x32, 0x02, 0x00, 0x03, 0xAA, 0xBB}
)

func decodeCat001(t *testing.T, data []byte) *asterix.DataBlock {
	t.Helper()
	u, err := uap.NewUAP12()
	if err != nil {
		t.Fatal(err)
	}
	db, err := asterix.NewDecoder(asterix.WithPreloadedUAPs(u)).Decode(data)
	if err != nil {
		t.Fatalf("decode % X: %v", data, err)
	}
	return db
}

func TestCat001RecordSelectsUAPByTYP(t *testing.T) {
	data := block(1, cat001PlotRecord, cat001TrackRecord, cat001SPRecord)
	db := decodeCat001(t, data)
	recs := db.Records()
	if len(recs) != 3 {
		t.Fatalf("got %d records", len(recs))
	}

	if _, ok := recs[0].GetDataItem("I001/040"); !ok {
		t.Error("plot record: I001/040 missing")
	}
	if it, ok := recs[0].GetDataItem("I001/090"); !ok || it.(*v12.ModeCCode).FlightLevel != 25 {
		t.Errorf("plot record: I001/090 = %+v", it)
	}

	tn, ok := recs[1].GetDataItem("I001/161")
	if !ok || tn.(*v12.TrackPlotNumber).Number != 0x1234 {
		t.Fatalf("track record: I001/161 = %+v (track UAP not selected)", tn)
	}
	pos, ok := recs[1].GetDataItem("I001/040")
	if !ok || pos.(*v12.PositionPolar).RHO != 50.5 {
		t.Errorf("track record: I001/040 = %+v", pos)
	}
	ts, ok := recs[1].GetDataItem("I001/170")
	if !ok || !ts.(*v12.TrackStatus).TRE {
		t.Errorf("track record: I001/170 = %+v", ts)
	}

	sp, ok := recs[2].GetDataItem("SP001")
	if !ok || !bytes.Equal(sp.(*common.ExplicitField).Payload, []byte{0xAA, 0xBB}) {
		t.Errorf("SP record: SP001 = %+v", sp)
	}

	out, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("re-encode:\n got % X\nwant % X", out, data)
	}
}

func TestCat001TrackRecordEncodesWithTrackFRNs(t *testing.T) {
	u, err := uap.NewUAP12()
	if err != nil {
		t.Fatal(err)
	}
	rec, err := asterix.NewRecord(asterix.Cat001, u)
	if err != nil {
		t.Fatal(err)
	}
	// Set the track-only items before I001/020 to show that FRNs do not
	// depend on the order of SetDataItem.
	for id, it := range map[string]asterix.DataItem{
		"I001/161": &v12.TrackPlotNumber{Number: 0x1234},
		"I001/170": &v12.TrackStatus{CON: true, TRE: true},
		"I001/040": &v12.PositionPolar{RHO: 50.5, THETA: 90},
		"I001/010": &common.DataSourceIdentifier{SAC: 0x32, SIC: 0x02},
		"I001/020": &v12.TargetReportDescriptor{TYP: v12.TYPTrack, SSR: true, PSR: true},
	} {
		if err := rec.SetDataItem(id, it); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	var buf bytes.Buffer
	if _, err := rec.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), cat001TrackRecord) {
		t.Fatalf("got % X\nwant % X", buf.Bytes(), cat001TrackRecord)
	}
}

func TestCat001PlotRecordRefusesTrackOnlyItem(t *testing.T) {
	u, _ := uap.NewUAP12()
	rec, _ := asterix.NewRecord(asterix.Cat001, u)
	_ = rec.SetDataItem("I001/020", &v12.TargetReportDescriptor{})
	if err := rec.SetDataItem("I001/161", &v12.TrackPlotNumber{Number: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := rec.Encode(new(bytes.Buffer)); err == nil {
		t.Fatal("plot record with I001/161 encoded")
	}
}

func TestCat001MalformedRecordsFail(t *testing.T) {
	u, _ := uap.NewUAP12()
	d := asterix.NewDecoder(asterix.WithPreloadedUAPs(u))
	for _, rec := range [][]byte{cat001PlotRecord, cat001TrackRecord, cat001SPRecord} {
		for i := 1; i < len(rec); i++ {
			data := block(1, rec[:i])
			if _, err := d.Decode(data); err == nil {
				t.Errorf("truncated record % X decoded without error", rec[:i])
			}
		}
	}
	for name, rec := range map[string][]byte{
		"SP LEN 0":        {0xC1, 0x01, 0x04, 0x32, 0x02, 0x00, 0x00},
		"SP LEN too long": {0xC1, 0x01, 0x04, 0x32, 0x02, 0x00, 0x09, 0xAA},
		"020 FX dangling": {0x40, 0x01},
	} {
		if _, err := d.Decode(block(1, rec)); err == nil {
			t.Errorf("%s: decoded without error", name)
		}
	}
}

func FuzzCat001Decode(f *testing.F) {
	f.Add(block(1, cat001PlotRecord))
	f.Add(block(1, cat001TrackRecord))
	f.Add(block(1, cat001SPRecord))
	f.Add(block(1, cat001PlotRecord, cat001TrackRecord))
	u, err := uap.NewUAP12()
	if err != nil {
		f.Fatal(err)
	}
	d := asterix.NewDecoder(asterix.WithPreloadedUAPs(u))
	f.Fuzz(func(t *testing.T, data []byte) {
		db, err := d.Decode(data)
		if err != nil {
			return
		}
		// Whatever decodes must not panic on re-encode.
		_, _ = db.Encode()
	})
}
