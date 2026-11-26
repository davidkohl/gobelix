// dataitems/cat062/v120/aircraft_derived_data_test.go
package v120_test

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// i380Subfields holds one canonical encoding per I062/380 subfield (FRN 1-28),
// with the subfield lengths of CAT062 v1.20: fixed lengths, TIS one octet
// (FX clear), TID and MB repetitive (REP + 15 / 8 octets per entry), MET 8
// octets, POS 6 octets.
var i380Subfields = []struct {
	frn  int
	name string
	data string
}{
	{1, "ADR", "123456"},
	{2, "ID", "14C673C78820"}, // ELY318
	{3, "MHG", "4000"},
	{4, "IAS (IAS)", "0123"},
	{4, "IAS (Mach)", "81C2"},
	{5, "TAS", "01C2"},
	{6, "SAL", "E0C8"},          // SAS=1, source=3, 5000 ft
	{6, "SAL negative", "9FFF"}, // SAS=1, source=0, -25 ft
	{7, "FSS", "A0C8"},          // MV=1, AH=0, AM=1, 5000 ft
	{8, "TIS", "80"},
	{8, "TIS NVB", "40"},
	{9, "TID one point", "01" + "85012C123456FEDCBA5B0123450064"},
	{9, "TID two points", "02" + "85012C123456FEDCBA5B0123450064" + "4AFFF6EDCBA9012345210000100032"},
	{10, "COM", "4CF5"},
	{11, "SAB", "5E05"},
	{12, "ACS", "11223344556677"},
	{13, "BVR", "FFC0"},
	{14, "GVR", "0040"},
	{15, "RAN", "FFF6"},
	{16, "TAR", "0010"},
	{17, "TAN", "4000"},
	{18, "GSP", "0100"},
	{19, "VUN", "05"},
	{20, "MET all", "F0" + "0014" + "005A" + "FFF0" + "02"},
	{20, "MET wind speed only", "80" + "0014" + "0000" + "0000" + "00"},
	{21, "EMC", "0D"},
	{22, "POS", "123456FEDCBA"},
	{23, "GAL", "0640"},
	{24, "PUN", "07"},
	{25, "MB one entry", "01" + "1122334455667740"},
	{25, "MB two entries", "02" + "1122334455667740" + "0102030405060750"},
	{26, "IAR", "0123"},
	{27, "MAC", "0064"},
	{28, "BPS", "03F5"},
}

// i380PrimarySubfield returns the shortest primary subfield announcing the
// given FRNs.
func i380PrimarySubfield(frns ...int) []byte {
	last := 0
	for _, frn := range frns {
		if octet := (frn - 1) / 7; octet > last {
			last = octet
		}
	}
	fspec := make([]byte, last+1)
	for _, frn := range frns {
		fspec[(frn-1)/7] |= 1 << (7 - (frn-1)%7)
	}
	for i := 0; i < last; i++ {
		fspec[i] |= 0x01
	}
	return fspec
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// assertI380ByteExact decodes input, checks every byte was consumed, and
// re-encodes it to the identical bytes.
func assertI380ByteExact(t *testing.T, input []byte) *v120.AircraftDerivedData {
	t.Helper()

	// A trailing sentinel byte catches over-reads into the next data item.
	buf := bytes.NewBuffer(append(append([]byte{}, input...), 0xA5))
	item := &v120.AircraftDerivedData{}
	n, err := item.Decode(buf)
	if err != nil {
		t.Fatalf("Decode(%x) error = %v", input, err)
	}
	if n != len(input) {
		t.Fatalf("Decode(%x) consumed %d bytes, want %d", input, n, len(input))
	}
	if buf.Len() != 1 {
		t.Fatalf("Decode(%x) left %d bytes, want only the sentinel", input, buf.Len())
	}

	var out bytes.Buffer
	n, err = item.Encode(&out)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if n != out.Len() {
		t.Errorf("Encode() returned %d but wrote %d bytes", n, out.Len())
	}
	if !bytes.Equal(out.Bytes(), input) {
		t.Errorf("round trip not byte-exact:\n  in:  %x\n  out: %x", input, out.Bytes())
	}
	return item
}

func TestAircraftDerivedData_SubfieldRoundTrip(t *testing.T) {
	for _, sf := range i380Subfields {
		t.Run(sf.name, func(t *testing.T) {
			input := append(i380PrimarySubfield(sf.frn), mustHex(t, sf.data)...)
			assertI380ByteExact(t, input)
		})
	}
}

// TestAircraftDerivedData_AllSubfields decodes an I062/380 with every one of
// the 28 subfields present (primary subfield FF FF FF FE).
func TestAircraftDerivedData_AllSubfields(t *testing.T) {
	var frns []int
	var payload strings.Builder
	seen := map[int]bool{}
	for _, sf := range i380Subfields {
		if seen[sf.frn] {
			continue
		}
		seen[sf.frn] = true
		frns = append(frns, sf.frn)
		payload.WriteString(sf.data)
	}
	if len(frns) != 28 {
		t.Fatalf("test table covers %d subfields, want 28", len(frns))
	}

	fspec := i380PrimarySubfield(frns...)
	if !bytes.Equal(fspec, []byte{0xFF, 0xFF, 0xFF, 0xFE}) {
		t.Fatalf("primary subfield = %x, want fffffffe", fspec)
	}
	input := append(fspec, mustHex(t, payload.String())...)
	if len(input) != 97 {
		t.Fatalf("all-subfield item is %d bytes, want 97", len(input))
	}

	item := assertI380ByteExact(t, input)

	if item.PositionLatitude == nil || item.PositionLongitude == nil {
		t.Fatal("position not decoded")
	}
	if item.MeteorologicalData == nil || item.MeteorologicalData.WindSpeed == nil ||
		*item.MeteorologicalData.WindSpeed != 20 || item.MeteorologicalData.Temperature == nil ||
		*item.MeteorologicalData.Temperature != -4 {
		t.Errorf("meteorological data = %+v", item.MeteorologicalData)
	}
	if len(item.TrajectoryIntentData) != 1 || item.TrajectoryIntentData[0].TCPNumber != 5 {
		t.Errorf("trajectory intent data = %+v", item.TrajectoryIntentData)
	}
	if len(item.ModeSMBData) != 1 {
		t.Errorf("mode S MB data = %x", item.ModeSMBData)
	}
	if item.BarometricPressureSetting == nil {
		t.Error("barometric pressure setting (FRN 28) not decoded")
	}
}

// TestAircraftDerivedData_EncodeAllFields encodes a value with every field
// set and checks it decodes back to the same value. Before the fix several
// encoders were stubs that set the presence bit but wrote no octets.
func TestAircraftDerivedData_EncodeAllFields(t *testing.T) {
	u8 := func(v uint8) *uint8 { return &v }
	u16 := func(v uint16) *uint16 { return &v }
	u32 := func(v uint32) *uint32 { return &v }
	i16 := func(v int16) *int16 { return &v }
	f64 := func(v float64) *float64 { return &v }
	b := func(v bool) *bool { return &v }
	s := func(v string) *string { return &v }
	acs := [7]byte{1, 2, 3, 4, 5, 6, 7}

	in := &v120.AircraftDerivedData{
		TargetAddress:          u32(0xABCDEF),
		TargetIdentification:   s("ELY318"),
		MagneticHeading:        f64(90),
		IndicatedAirspeedIMAch: b(false),
		IndicatedAirspeed:      f64(0.125),
		TrueAirspeed:           u16(450),
		SelectedAltitudeSAS:    b(true),
		SelectedAltitudeSource: u8(2),
		SelectedAltitude:       i16(25000),
		FinalStateAltitudeMV:   b(true),
		FinalStateAltitudeAH:   b(false),
		FinalStateAltitudeAM:   b(true),
		FinalStateAltitude:     i16(-1000),
		TrajectoryIntentNAV:    b(false),
		TrajectoryIntentNVB:    b(true),
		TrajectoryIntentData: []v120.TrajectoryIntentPoint{{
			TCA: true, NC: false, TCPNumber: 7, Altitude: 3000,
			Latitude: 45, Longitude: -90, PointType: 3, TD: 1,
			TRA: true, TOA: false, TOV: 3600, TTR: 1,
		}},
		CommunicationsCOM:         u8(1),
		CommunicationsSTAT:        u8(2),
		CommunicationsSSC:         b(true),
		CommunicationsARC:         b(false),
		CommunicationsAIC:         b(true),
		CommunicationsB1A:         b(false),
		CommunicationsB1B:         u8(9),
		ADSBStatusAC:              u8(1),
		ADSBStatusMN:              u8(2),
		ADSBStatusDC:              u8(3),
		ADSBStatusGBS:             b(true),
		ADSBStatusSTAT:            u8(4),
		ACASResolutionAdvisory:    &acs,
		BarometricVerticalRate:    i16(-160),
		GeometricVerticalRate:     i16(160),
		RollAngle:                 f64(-10),
		TrackAngleRate:            f64(0.5),
		TrackAngle:                f64(180),
		GroundSpeed:               f64(225),
		VelocityUncertainty:       u8(3),
		MeteorologicalData:        &v120.MeteorologicalData{WindSpeed: u16(35), WindDirection: u16(270), Temperature: f64(-56.5), Turbulence: u8(1)},
		EmitterCategory:           u8(5),
		PositionLatitude:          f64(-33.75),
		PositionLongitude:         f64(135),
		GeometricAltitude:         i16(5600),
		PositionUncertainty:       u8(8),
		ModeSMBData:               [][]byte{{1, 2, 3, 4, 5, 6, 7, 0x40}, {8, 9, 10, 11, 12, 13, 14, 0x50}},
		IndicatedAirspeedData:     u16(280),
		MachNumber:                f64(1),
		BarometricPressureSetting: f64(100),
	}

	var buf bytes.Buffer
	n, err := in.Encode(&buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	// 4 primary + 3+6+2+2+2+2+2 + 1 + (1+15) + 2+2+7+2+2 + 2+2+2+2+1+8+1 +
	// 6+2+1+(1+2*8)+2+2+2
	const wantLen = 4 + 19 + 1 + 16 + 15 + 18 + 32
	if n != wantLen || buf.Len() != wantLen {
		t.Fatalf("Encode() = %d bytes (buffer %d), want %d: %x", n, buf.Len(), wantLen, buf.Bytes())
	}
	encoded := append([]byte{}, buf.Bytes()...)

	out := &v120.AircraftDerivedData{}
	n, err = out.Decode(&buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if n != wantLen {
		t.Fatalf("Decode() consumed %d bytes, want %d", n, wantLen)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("decoded value differs:\n  in:  %+v\n  out: %+v", in, out)
	}

	assertI380ByteExact(t, encoded)
}

// TestAircraftDerivedData_TISExtents checks that FX extents of subfield #8
// are consumed, so the following subfield starts at the right octet.
func TestAircraftDerivedData_TISExtents(t *testing.T) {
	// Primary subfield: TIS (FRN 8) and COM (FRN 10). TIS = 81 01 00 (two
	// extents), COM = 4CF5.
	input := append(i380PrimarySubfield(8, 10), mustHex(t, "810100"+"4CF5")...)
	item := &v120.AircraftDerivedData{}
	n, err := item.Decode(bytes.NewBuffer(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if n != len(input) {
		t.Fatalf("Decode() consumed %d bytes, want %d", n, len(input))
	}
	if item.CommunicationsCOM == nil || *item.CommunicationsCOM != 2 {
		t.Errorf("COM decoded from the wrong offset: %v", item.CommunicationsCOM)
	}
}

func TestAircraftDerivedData_DecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"primary subfield FX beyond fourth octet", "01010101" + "00"},
		{"primary subfield truncated", "01"},
		{"MET shorter than 8 octets", "010104" + "F000140000FFF0"},
		{"POS shorter than 6 octets", "01010180" + "123456FEDC"},
		{"TID point truncated", "0140" + "01" + "85012C123456FEDCBA5B01234500"},
		{"MB entry truncated", "01010110" + "01" + "11223344556677"},
		{"TIS extent missing", "0180" + "81"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &v120.AircraftDerivedData{}
			if _, err := item.Decode(bytes.NewBuffer(mustHex(t, tt.input))); err == nil {
				t.Errorf("Decode(%s) succeeded, want error", tt.input)
			}
		})
	}
}

func TestAircraftDerivedData_EncodeErrors(t *testing.T) {
	lat := 10.0
	tests := []struct {
		name string
		item *v120.AircraftDerivedData
	}{
		{"position without longitude", &v120.AircraftDerivedData{PositionLatitude: &lat}},
		{"MB entry not 8 bytes", &v120.AircraftDerivedData{ModeSMBData: [][]byte{{1, 2, 3}}}},
		{"TID over 255 points", &v120.AircraftDerivedData{TrajectoryIntentData: make([]v120.TrajectoryIntentPoint, 256)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if _, err := tt.item.Encode(&buf); err == nil {
				t.Errorf("Encode() succeeded, want error (wrote %x)", buf.Bytes())
			}
		})
	}
}
