// cat/cat062/opentrack_profile_test.go
//
// Round trip of full CAT 062 records with the items an SDPS track output
// sends (system SAC/SIC, service, time, WGS-84 and Cartesian position,
// velocity, Mode 3/A, identification, I062/380 address and callsign, track
// number and status, update and data ages, mode of movement, altitudes,
// rate of climb, accuracies, measured information and SP062), several
// records in one data block: a wrong length of any item misaligns the
// record after it.
package cat062_test

import (
	"bytes"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	v117 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v117"
	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func u8(v uint8) *uint8    { return &v }
func u16(v uint16) *uint16 { return &v }
func u32(v uint32) *uint32 { return &v }
func str(v string) *string { return &v }

// sdpsRecordV120 returns the items of one SDPS track record, edition 1.20.
func sdpsRecordV120(tn uint16, tod float64, sp []byte) map[string]asterix.DataItem {
	ts := &v120.TrackStatus{MON: tn%2 == 0, SRC: 3, TSB: tn == 1, AFF: true, CST: tn == 3, ADS: true, EMS: 1, SFC: false, MLAT: tn == 2}
	ts.SetHasExtension()
	return map[string]asterix.DataItem{
		"I062/010": &common.DataSourceIdentifier{SAC: 0x3E, SIC: 0x01},
		"I062/015": &common.ServiceIdentification{Value: 7},
		"I062/070": &v120.TimeOfTrackInformation{Time: tod},
		"I062/105": &v120.CalculatedPositionWGS84{Latitude: 50.0379, Longitude: 8.5622},
		"I062/100": &v120.CalculatedTrackPositionCartesian{X: -12345.5, Y: 67890},
		"I062/185": &v120.CalculatedTrackVelocity{Vx: 220.25, Vy: -30.5},
		"I062/060": &v120.TrackMode3ACode{Changed: tn == 2, Code: 1234},
		"I062/245": &v120.TargetIdentification{IdentType: v120.CallsignRegistration, Ident: "DLH4AB"},
		"I062/380": &v120.AircraftDerivedData{TargetAddress: u32(0x3C6544), TargetIdentification: str("DLH4AB")},
		"I062/040": &v120.TrackNumber{Value: tn},
		"I062/080": ts,
		// SSR age 0.5 s, ADS age 1.25 s, MLT age 2 s (LSB 1/4 s).
		"I062/290": &v120.SystemTrackUpdateAges{Data: []byte{0x29, 0x02, 2, 0, 5, 8}},
		"I062/200": &v120.ModeOfMovement{Trans: v120.TransLeftTurn, Long: v120.LongDecreasingGroundspeed, Vert: v120.VertDescent},
		// FL age 1 s (MFL, octet 1 bit 8), LSB 1/4 s.
		"I062/295": &v120.TrackDataAges{Data: []byte{0x80, 4}},
		"I062/136": &v120.MeasuredFlightLevel{FlightLevel: 350.25},
		"I062/135": &v120.CalculatedTrackBarometricAltitude{Altitude: 350.5},
		"I062/220": &v120.CalculatedRateOfClimbDescent{Rate: -1500},
		"I062/500": &v120.EstimatedAccuracies{APCX: u16(80), APCY: u16(90), BarometricAltitudeAccuracy: u8(4),
			VelocityX: u8(12), VelocityY: u8(13), RateOfClimbDescentAccuracy: u8(16)},
		"I062/340": &v120.MeasuredInformation{SensorSAC: u8(0x3E), SensorSIC: u8(0x21)},
		"SP062":    &v120.SpecialPurpose{Data: sp},
	}
}

// TestSDPSProfileRoundTripV120 encodes three records into one data block,
// decodes it and compares every item.
func TestSDPSProfileRoundTripV120(t *testing.T) {
	u, err := uap.NewUAP120()
	if err != nil {
		t.Fatal(err)
	}
	sps := [][]byte{{3, 0x10, 0x80}, {3, 0x18, 0x00}, {1}}
	db, err := asterix.NewDataBlock(asterix.Cat062, u)
	if err != nil {
		t.Fatal(err)
	}
	var want []map[string]asterix.DataItem
	for i, sp := range sps {
		items := sdpsRecordV120(uint16(i+1), 43200+float64(i)/128, sp)
		want = append(want, items)
		if err := db.EncodeRecord(items); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	data, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	dec := asterix.NewDecoder()
	dec.RegisterUAP(u)
	got, err := dec.Decode(data)
	if err != nil {
		t.Fatalf("decode: %v (% X)", err, data)
	}
	if got.RecordCount() != len(sps) {
		t.Fatalf("records = %d, want %d", got.RecordCount(), len(sps))
	}
	for i, rec := range got.Records() {
		for id, w := range want[i] {
			g, ok := rec.GetDataItem(id)
			if !ok {
				t.Errorf("record %d: %s missing", i, id)
				continue
			}
			var wb, gb bytes.Buffer
			if _, err := w.Encode(&wb); err != nil {
				t.Fatal(err)
			}
			if _, err := g.Encode(&gb); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wb.Bytes(), gb.Bytes()) {
				t.Errorf("record %d %s: decoded % X, sent % X", i, id, gb.Bytes(), wb.Bytes())
			}
		}
		if n := rec.ItemCount(); n != len(want[i]) {
			t.Errorf("record %d: %d items, want %d", i, n, len(want[i]))
		}
	}
	// Spot checks of decoded values against what was sent.
	r0 := got.Records()[0]
	if it, _ := r0.GetDataItem("SP062"); !bytes.Equal(it.(*v120.SpecialPurpose).Data, sps[0]) {
		t.Errorf("SP062 = % X, want % X", it.(*v120.SpecialPurpose).Data, sps[0])
	}
	if it, _ := got.Records()[2].GetDataItem("I062/040"); it.(*v120.TrackNumber).Value != 3 {
		t.Errorf("track number of record 3 = %d, want 3", it.(*v120.TrackNumber).Value)
	}
	if it, _ := r0.GetDataItem("I062/380"); *it.(*v120.AircraftDerivedData).TargetAddress != 0x3C6544 ||
		*it.(*v120.AircraftDerivedData).TargetIdentification != "DLH4AB" {
		t.Errorf("I062/380 = %v", it)
	}
	if it, _ := got.Records()[1].GetDataItem("I062/080"); !it.(*v120.TrackStatus).MLAT || !it.(*v120.TrackStatus).AFF {
		t.Errorf("I062/080 of record 2 = %v, want MLAT and AFF", it)
	}
	if it, _ := r0.GetDataItem("I062/105"); it.(*v120.CalculatedPositionWGS84).Latitude-50.0379 > 1e-5 {
		t.Errorf("I062/105 = %v", it)
	}
}

// TestSpecialPurposeLenV120 checks that LEN of SP062 and RE062 counts the
// length octet itself (ASTERIX Part 1 §4.3.5): LEN = 3 takes 2 more octets.
func TestSpecialPurposeLenV120(t *testing.T) {
	for _, it := range []interface {
		asterix.DataItem
	}{&v120.SpecialPurpose{}, &v120.ReservedExpansion{}, &v117.ReservedExpansion{}} {
		buf := bytes.NewBuffer([]byte{3, 0xAA, 0xBB, 0xCC})
		n, err := it.Decode(buf)
		if err != nil || n != 3 || buf.Len() != 1 {
			t.Errorf("%T: n = %d, err %v, left %d; want 3 octets read, 1 left", it, n, err, buf.Len())
		}
		if _, err := it.Decode(bytes.NewBuffer([]byte{0})); err == nil {
			t.Errorf("%T: LEN 0 accepted", it)
		}
		if _, err := it.Decode(bytes.NewBuffer([]byte{4, 1, 2})); err == nil {
			t.Errorf("%T: LEN beyond the data accepted", it)
		}
	}
	var b bytes.Buffer
	if _, err := (&v120.SpecialPurpose{}).Encode(&b); err != nil || !bytes.Equal(b.Bytes(), []byte{1}) {
		t.Errorf("empty SP062 encodes as % X (%v), want 01", b.Bytes(), err)
	}
	if err := (&v120.SpecialPurpose{Data: []byte{2, 1, 2}}).Validate(); err == nil {
		t.Error("SP062 with LEN 2 and 3 octets accepted")
	}
}

// TestSDPSProfileRoundTripV117 is the edition 1.17 round trip of the same
// profile without SP062 and the 1.20-only items.
func TestSDPSProfileRoundTripV117(t *testing.T) {
	u, err := uap.NewUAP117()
	if err != nil {
		t.Fatal(err)
	}
	mk := func(tn uint16) map[string]asterix.DataItem {
		ts := &v117.TrackStatus{MON: true, SRC: 3, TSE: tn == 2, CST: true}
		ts.SetHasExtension()
		ax, vx := 40.0, 3.0
		return map[string]asterix.DataItem{
			"I062/010": &common.DataSourceIdentifier{SAC: 0x3E, SIC: 0x01},
			"I062/015": &common.ServiceIdentification{Value: 7},
			"I062/070": &v117.TimeOfTrackInformation{Time: 1000.5},
			"I062/105": &v117.CalculatedPositionWGS84{Latitude: 50.0379, Longitude: 8.5622},
			"I062/185": &v117.CalculatedTrackVelocity{Vx: 220.25, Vy: -30.5},
			"I062/060": &v117.TrackMode3ACode{Code: 0o1234},
			"I062/245": &v117.TargetIdentification{IdentType: v117.CallsignRegistration, Ident: "DLH4AB"},
			"I062/380": &v117.AircraftDerivedData{TargetAddress: u32(0x3C6544)},
			"I062/040": &v117.TrackNumber{Value: tn},
			"I062/080": ts,
			"I062/135": &v117.CalculatedTrackBarometricAltitude{Altitude: 350.5},
			"I062/220": &v117.CalculatedRateOfClimbDescent{Rate: -1500},
			"I062/500": &v117.EstimatedAccuracies{PositionAccuracyX: &ax, PositionAccuracyY: &ax, VelocityAccuracyX: &vx, VelocityAccuracyY: &vx},
			"RE062":    &v117.ReservedExpansion{Data: []byte{2, 0x55}},
		}
	}
	db, err := asterix.NewDataBlock(asterix.Cat062, u)
	if err != nil {
		t.Fatal(err)
	}
	want := []map[string]asterix.DataItem{mk(1), mk(2)}
	for _, w := range want {
		if err := db.EncodeRecord(w); err != nil {
			t.Fatal(err)
		}
	}
	data, err := db.Encode()
	if err != nil {
		t.Fatal(err)
	}
	dec := asterix.NewDecoder()
	dec.RegisterUAP(u)
	got, err := dec.Decode(data)
	if err != nil {
		t.Fatalf("decode: %v (% X)", err, data)
	}
	if got.RecordCount() != 2 {
		t.Fatalf("records = %d, want 2", got.RecordCount())
	}
	for i, rec := range got.Records() {
		for id, w := range want[i] {
			g, ok := rec.GetDataItem(id)
			if !ok {
				t.Errorf("record %d: %s missing", i, id)
				continue
			}
			var wb, gb bytes.Buffer
			if _, err := w.Encode(&wb); err != nil {
				t.Fatal(err)
			}
			if _, err := g.Encode(&gb); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wb.Bytes(), gb.Bytes()) {
				t.Errorf("record %d %s: decoded % X, sent % X", i, id, gb.Bytes(), wb.Bytes())
			}
		}
	}
	if it, _ := got.Records()[1].GetDataItem("I062/080"); !it.(*v117.TrackStatus).TSE {
		t.Errorf("TSE lost: %v", it)
	}
}
