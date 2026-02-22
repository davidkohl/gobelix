// dataitems/cat062/v120/roundtrip_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// encodeDecodeRoundTrip is a helper that encodes an item, decodes it, then re-encodes
// and verifies byte-for-byte equality. Returns the decoded item for further assertions.
func encodeDecodeRoundTrip(t *testing.T, item interface {
	Encode(*bytes.Buffer) (int, error)
	Decode(*bytes.Buffer) (int, error)
}, decoded interface {
	Encode(*bytes.Buffer) (int, error)
	Decode(*bytes.Buffer) (int, error)
}, name string) []byte {
	t.Helper()

	// Encode original
	encodeBuf := new(bytes.Buffer)
	n, err := item.Encode(encodeBuf)
	if err != nil {
		t.Fatalf("%s: Encode() error = %v", name, err)
	}
	encoded := encodeBuf.Bytes()
	if n != len(encoded) {
		t.Errorf("%s: Encode() returned %d but wrote %d bytes", name, n, len(encoded))
	}

	// Decode into fresh struct
	decodeBuf := bytes.NewBuffer(bytes.Clone(encoded))
	bytesRead, err := decoded.Decode(decodeBuf)
	if err != nil {
		t.Fatalf("%s: Decode() error = %v (encoded: %X)", name, err, encoded)
	}
	if bytesRead != len(encoded) {
		t.Errorf("%s: Decode() read %d bytes, want %d", name, bytesRead, len(encoded))
	}

	// Re-encode decoded to verify identity
	reEncodeBuf := new(bytes.Buffer)
	_, err = decoded.Encode(reEncodeBuf)
	if err != nil {
		t.Fatalf("%s: re-Encode() after decode error = %v", name, err)
	}
	if !bytes.Equal(reEncodeBuf.Bytes(), encoded) {
		t.Errorf("%s: round-trip mismatch:\n  original:   %X\n  re-encoded: %X", name, encoded, reEncodeBuf.Bytes())
	}

	return encoded
}

// --- I062/040 TrackNumber ---

func TestTrackNumber_RoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		value   uint16
		encoded []byte
	}{
		{"zero", 0, []byte{0x00, 0x00}},
		{"typical", 1234, []byte{0x04, 0xD2}},
		{"max", 0xFFFF, []byte{0xFF, 0xFF}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.TrackNumber{Value: tc.value}
			dec := &v120.TrackNumber{}
			encoded := encodeDecodeRoundTrip(t, orig, dec, "TrackNumber")
			if !bytes.Equal(encoded, tc.encoded) {
				t.Errorf("encoded = %X, want %X", encoded, tc.encoded)
			}
			if dec.Value != tc.value {
				t.Errorf("decoded Value = %d, want %d", dec.Value, tc.value)
			}
		})
	}
}

// --- I062/060 TrackMode3ACode ---

func TestTrackMode3ACode_RoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		item      v120.TrackMode3ACode
		encoded   []byte
	}{
		{
			name:    "zero code no flags",
			item:    v120.TrackMode3ACode{Code: 0},
			encoded: []byte{0x00, 0x00},
		},
		{
			// Code stores the squawk as a decimal number where each digit is an octal digit:
			// squawk 7000 → Code = 7000 (decimal). NOT 07000 (Go octal literal = 3584).
			name:    "squawk 7000 validated",
			item:    v120.TrackMode3ACode{Validated: true, Code: 7000},
			encoded: nil,
		},
		{
			name:    "squawk 7700 garbled changed",
			item:    v120.TrackMode3ACode{Garbled: true, Changed: true, Code: 7700},
			encoded: nil,
		},
		{
			name:    "squawk 1234",
			item:    v120.TrackMode3ACode{Code: 1234},
			encoded: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := tc.item
			dec := &v120.TrackMode3ACode{}
			encodeDecodeRoundTrip(t, &orig, dec, "TrackMode3ACode")
			if dec.Code != tc.item.Code {
				t.Errorf("decoded Code = 0%o, want 0%o", dec.Code, tc.item.Code)
			}
			if dec.Validated != tc.item.Validated {
				t.Errorf("decoded Validated = %v, want %v", dec.Validated, tc.item.Validated)
			}
			if dec.Garbled != tc.item.Garbled {
				t.Errorf("decoded Garbled = %v, want %v", dec.Garbled, tc.item.Garbled)
			}
			if dec.Changed != tc.item.Changed {
				t.Errorf("decoded Changed = %v, want %v", dec.Changed, tc.item.Changed)
			}
		})
	}
}

// --- I062/070 TimeOfTrackInformation ---

func TestTimeOfTrackInformation_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		time float64
	}{
		{"midnight", 0.0},
		{"one second", 1.0},
		{"noon", 43200.0},
		{"one LSB", 1.0 / 128.0},
		{"typical", 36000.0}, // 10:00:00 UTC
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.TimeOfTrackInformation{Time: tc.time}
			dec := &v120.TimeOfTrackInformation{}
			encodeDecodeRoundTrip(t, orig, dec, "TimeOfTrackInformation")
			// Allow 1 LSB of rounding
			diff := dec.Time - tc.time
			if diff < 0 {
				diff = -diff
			}
			if diff > 1.0/128.0 {
				t.Errorf("decoded Time = %f, want %f (diff %f > 1 LSB)", dec.Time, tc.time, diff)
			}
		})
	}
}

// --- I062/120 TrackMode2Code ---

func TestTrackMode2Code_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		code uint16
	}{
		{"zero", 0},
		{"typical 1234", 01234},
		{"max octal 7777", 07777},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.TrackMode2Code{Code: tc.code}
			dec := &v120.TrackMode2Code{}
			encodeDecodeRoundTrip(t, orig, dec, "TrackMode2Code")
			if dec.Code != tc.code {
				t.Errorf("decoded Code = 0%o, want 0%o", dec.Code, tc.code)
			}
		})
	}
}

// --- I062/130 CalculatedTrackGeometricAltitude ---

func TestCalculatedTrackGeometricAltitude_RoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		altitude float64
	}{
		{"sea level", 0.0},
		{"one LSB", 6.25},
		{"typical cruise 35000 ft", 35000.0},
		{"negative ground", -1500.0},
		{"high altitude", 60000.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedTrackGeometricAltitude{Altitude: tc.altitude}
			dec := &v120.CalculatedTrackGeometricAltitude{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedTrackGeometricAltitude")
			diff := dec.Altitude - tc.altitude
			if diff < 0 {
				diff = -diff
			}
			if diff > 6.25 {
				t.Errorf("decoded Altitude = %f, want %f", dec.Altitude, tc.altitude)
			}
		})
	}
}

// --- I062/135 CalculatedTrackBarometricAltitude ---

func TestCalculatedTrackBarometricAltitude_RoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		altitude float64
		qnh      bool
	}{
		{"FL000 no QNH", 0.0, false},
		{"FL350 with QNH", 350.0, true},
		{"FL350 no QNH", 350.0, false},
		{"negative FL", -15.0, false},
		{"one LSB", 0.25, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedTrackBarometricAltitude{Altitude: tc.altitude, QNH: tc.qnh}
			dec := &v120.CalculatedTrackBarometricAltitude{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedTrackBarometricAltitude")
			diff := dec.Altitude - tc.altitude
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.25 {
				t.Errorf("decoded Altitude = %f, want %f", dec.Altitude, tc.altitude)
			}
			if dec.QNH != tc.qnh {
				t.Errorf("decoded QNH = %v, want %v", dec.QNH, tc.qnh)
			}
		})
	}
}

// --- I062/136 MeasuredFlightLevel ---

func TestMeasuredFlightLevel_RoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		level float64
	}{
		{"FL000", 0.0},
		{"FL350", 350.0},
		{"FL010", 10.0},
		{"negative FL", -10.0},
		{"one LSB 0.25 FL", 0.25},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.MeasuredFlightLevel{FlightLevel: tc.level}
			dec := &v120.MeasuredFlightLevel{}
			encodeDecodeRoundTrip(t, orig, dec, "MeasuredFlightLevel")
			diff := dec.FlightLevel - tc.level
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.25 {
				t.Errorf("decoded FlightLevel = %f, want %f", dec.FlightLevel, tc.level)
			}
		})
	}
}

// --- I062/100 CalculatedTrackPositionCartesian ---

func TestCalculatedTrackPositionCartesian_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		x, y float64
	}{
		{"origin", 0, 0},
		{"east north", 50000.0, 80000.0},
		{"west south", -50000.0, -80000.0},
		{"one LSB", 0.5, 0.5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedTrackPositionCartesian{X: tc.x, Y: tc.y}
			dec := &v120.CalculatedTrackPositionCartesian{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedTrackPositionCartesian")
			if diff := dec.X - tc.x; diff < -0.5 || diff > 0.5 {
				t.Errorf("decoded X = %f, want %f", dec.X, tc.x)
			}
			if diff := dec.Y - tc.y; diff < -0.5 || diff > 0.5 {
				t.Errorf("decoded Y = %f, want %f", dec.Y, tc.y)
			}
		})
	}
}

// --- I062/105 CalculatedPositionWGS84 ---

func TestCalculatedPositionWGS84_RoundTrip(t *testing.T) {
	lsb := 180.0 / float64(int32(1)<<25) // approx 5.4e-6 degrees

	cases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"origin", 0.0, 0.0},
		{"London Heathrow", 51.4775, -0.4614},
		{"Sydney", -33.9461, 151.1772},
		{"North Pole", 89.9999, 0.0},
		{"South Pole", -89.9999, 0.0},
		{"date line east", 0.0, 179.9999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedPositionWGS84{Latitude: tc.lat, Longitude: tc.lon}
			dec := &v120.CalculatedPositionWGS84{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedPositionWGS84")
			if diff := dec.Latitude - tc.lat; diff < -lsb*2 || diff > lsb*2 {
				t.Errorf("decoded Latitude = %f, want %f", dec.Latitude, tc.lat)
			}
			if diff := dec.Longitude - tc.lon; diff < -lsb*2 || diff > lsb*2 {
				t.Errorf("decoded Longitude = %f, want %f", dec.Longitude, tc.lon)
			}
		})
	}
}

// --- I062/185 CalculatedTrackVelocity ---

func TestCalculatedTrackVelocity_RoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		vx, vy float64
	}{
		{"stationary", 0, 0},
		{"eastbound 250kt", 128.6, 0},
		{"typical cruise", 200.0, -50.0},
		{"one LSB", 0.25, 0.25},
		{"negative max", -8192.0, -8192.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedTrackVelocity{Vx: tc.vx, Vy: tc.vy}
			dec := &v120.CalculatedTrackVelocity{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedTrackVelocity")
			if diff := dec.Vx - tc.vx; diff < -0.25 || diff > 0.25 {
				t.Errorf("decoded Vx = %f, want %f", dec.Vx, tc.vx)
			}
			if diff := dec.Vy - tc.vy; diff < -0.25 || diff > 0.25 {
				t.Errorf("decoded Vy = %f, want %f", dec.Vy, tc.vy)
			}
		})
	}
}

// --- I062/210 CalculatedAcceleration ---

func TestCalculatedAcceleration_RoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		ax, ay float64
	}{
		{"zero", 0, 0},
		{"right turn", 2.0, -1.5},
		{"one LSB", 0.25, 0.25},
		{"max positive", 31.75, 31.75},
		{"max negative", -32.0, -32.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedAcceleration{Ax: tc.ax, Ay: tc.ay}
			dec := &v120.CalculatedAcceleration{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedAcceleration")
			if diff := dec.Ax - tc.ax; diff < -0.25 || diff > 0.25 {
				t.Errorf("decoded Ax = %f, want %f", dec.Ax, tc.ax)
			}
			if diff := dec.Ay - tc.ay; diff < -0.25 || diff > 0.25 {
				t.Errorf("decoded Ay = %f, want %f", dec.Ay, tc.ay)
			}
		})
	}
}

// --- I062/200 ModeOfMovement ---

func TestModeOfMovement_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		item v120.ModeOfMovement
	}{
		{
			"all zeros",
			v120.ModeOfMovement{
				Trans: v120.TransConstantCourse,
				Long:  v120.LongConstantGroundspeed,
				Vert:  v120.VertLevel,
			},
		},
		{
			"right turn climbing",
			v120.ModeOfMovement{
				Trans: v120.TransRightTurn,
				Long:  v120.LongIncreasingGroundspeed,
				Vert:  v120.VertClimb,
			},
		},
		{
			"left turn descending with alt discrepancy",
			v120.ModeOfMovement{
				Trans:          v120.TransLeftTurn,
				Long:           v120.LongDecreasingGroundspeed,
				Vert:           v120.VertDescent,
				AltDiscrepancy: true,
			},
		},
		{
			"all undetermined",
			v120.ModeOfMovement{
				Trans: v120.TransUndetermined,
				Long:  v120.LongUndetermined,
				Vert:  v120.VertUndetermined,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := tc.item
			dec := &v120.ModeOfMovement{}
			encodeDecodeRoundTrip(t, &orig, dec, "ModeOfMovement")
			if dec.Trans != tc.item.Trans {
				t.Errorf("decoded Trans = %d, want %d", dec.Trans, tc.item.Trans)
			}
			if dec.Long != tc.item.Long {
				t.Errorf("decoded Long = %d, want %d", dec.Long, tc.item.Long)
			}
			if dec.Vert != tc.item.Vert {
				t.Errorf("decoded Vert = %d, want %d", dec.Vert, tc.item.Vert)
			}
			if dec.AltDiscrepancy != tc.item.AltDiscrepancy {
				t.Errorf("decoded AltDiscrepancy = %v, want %v", dec.AltDiscrepancy, tc.item.AltDiscrepancy)
			}
		})
	}
}

// --- I062/080 TrackStatus ---

func TestTrackStatus_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		item v120.TrackStatus
		wantBytes int
	}{
		{
			"minimal - no extensions",
			v120.TrackStatus{MON: true, CNF: true},
			1,
		},
		{
			"with first extension - simulated flight plan correlated",
			v120.TrackStatus{
				MON: true, SPI: false, MRH: true, SRC: 3, CNF: false,
				SIM: true, FPC: true,
			},
			2,
		},
		{
			"with second extension - military",
			v120.TrackStatus{
				MON: false, SRC: 1, CNF: true,
				SIM: false, KOS: true,
				AMA: true, MD4: 2, ME: true,
			},
			3,
		},
		{
			"with third extension - coasting PSR SSR",
			v120.TrackStatus{
				CNF: true, SRC: 2,
				TSE: true,
				AMA: true,
				CST: true, PSR: true, SSR: true,
			},
			4,
		},
		{
			"with fourth extension",
			v120.TrackStatus{
				MON: true,
				SIM: true,
				AMA: true,
				CST: true,
				SDS: 2, EMS: 5, PFT: true,
			},
			5,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := tc.item
			orig.SetHasExtension()
			dec := &v120.TrackStatus{}
			encoded := encodeDecodeRoundTrip(t, &orig, dec, "TrackStatus")
			if len(encoded) != tc.wantBytes {
				t.Errorf("encoded length = %d, want %d (bytes: %X)", len(encoded), tc.wantBytes, encoded)
			}
		})
	}
}

// --- I062/245 TargetIdentification ---

func TestTargetIdentification_RoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		identType v120.TargetIdentificationType
		ident     string
	}{
		{"callsign SAS123", v120.CallsignRegistration, "SAS123"},
		{"callsign 8 chars", v120.CallsignRegistration, "EZY1234A"},
		{"registration", v120.RegistrationNotDownlinked, "LNRKA"},
		{"invalid ident", v120.InvalidIdentification, "ABC"},
		{"with spaces", v120.CallsignNotDownlinked, "BA 245"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.TargetIdentification{IdentType: tc.identType, Ident: tc.ident}
			dec := &v120.TargetIdentification{}
			encodeDecodeRoundTrip(t, orig, dec, "TargetIdentification")
			if dec.IdentType != tc.identType {
				t.Errorf("decoded IdentType = %d, want %d", dec.IdentType, tc.identType)
			}
			if dec.Ident != tc.ident {
				t.Errorf("decoded Ident = %q, want %q", dec.Ident, tc.ident)
			}
		})
	}
}

// --- I062/300 VehicleFleetIdentification ---

func TestVehicleFleetIdentification_RoundTrip(t *testing.T) {
	cases := []struct {
		name        string
		vehicleType v120.VehicleFleetType
	}{
		{"unknown", v120.UnknownVehicle},
		{"ATC equipment", v120.ATCEquipmentMaintenance},
		{"bus", v120.Bus},
		{"emergency", v120.Emergency},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.VehicleFleetIdentification{VehicleType: tc.vehicleType}
			dec := &v120.VehicleFleetIdentification{}
			encodeDecodeRoundTrip(t, orig, dec, "VehicleFleetIdentification")
			if dec.VehicleType != tc.vehicleType {
				t.Errorf("decoded VehicleType = %d, want %d", dec.VehicleType, tc.vehicleType)
			}
		})
	}
}

// --- I062/340 MeasuredInformation ---

func TestMeasuredInformation_RoundTrip(t *testing.T) {
	sac := uint8(100)
	sic := uint8(15)
	rho := int16(1000)
	theta := uint16(32768)
	height := int16(200)
	cv := false
	c := true
	cAlt := float64(35000)
	v3a := true
	g3a := false
	l3a := false
	code3a := uint16(07700)
	typ := uint8(3)
	sim := false
	rab := true
	tst := false

	cases := []struct {
		name string
		item v120.MeasuredInformation
	}{
		{
			"sensor ID only",
			v120.MeasuredInformation{
				SensorSAC: &sac,
				SensorSIC: &sic,
			},
		},
		{
			"position only",
			v120.MeasuredInformation{
				MeasuredPositionRho:   &rho,
				MeasuredPositionTheta: &theta,
			},
		},
		{
			"all subfields",
			v120.MeasuredInformation{
				SensorSAC:             &sac,
				SensorSIC:             &sic,
				MeasuredPositionRho:   &rho,
				MeasuredPositionTheta: &theta,
				Measured3DHeight:      &height,
				ModeCV:                &cv,
				ModeC:                 &c,
				ModeCAltitude:         &cAlt,
				Mode3AV:               &v3a,
				Mode3AG:               &g3a,
				Mode3AL:               &l3a,
				Mode3ACode:            &code3a,
				ReportTypeTYP:         &typ,
				ReportTypeSIM:         &sim,
				ReportTypeRAB:         &rab,
				ReportTypeTST:         &tst,
			},
		},
		{
			"empty",
			v120.MeasuredInformation{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := tc.item
			dec := &v120.MeasuredInformation{}
			encodeDecodeRoundTrip(t, &orig, dec, "MeasuredInformation")
		})
	}
}

// --- I062/500 EstimatedAccuracies (extending existing tests) ---

func TestEstimatedAccuracies_AllSubfields_RoundTrip(t *testing.T) {
	// All subfields present including the extension byte one (ARC)
	item := &v120.EstimatedAccuracies{
		APCX:                       uint16Ptr(100),
		APCY:                       uint16Ptr(200),
		XYCovariance:               int16Ptr(-50),
		APWLatitude:                uint16Ptr(75),
		APWLongitude:               uint16Ptr(75),
		GeometricAltitudeAccuracy:  uint8Ptr(8),
		BarometricAltitudeAccuracy: uint8Ptr(12),
		VelocityX:                  uint8Ptr(3),
		VelocityY:                  uint8Ptr(4),
		AccelerationX:              uint8Ptr(1),
		AccelerationY:              uint8Ptr(2),
		RateOfClimbDescentAccuracy: uint8Ptr(10),
	}
	dec := &v120.EstimatedAccuracies{}
	encodeDecodeRoundTrip(t, item, dec, "EstimatedAccuracies/AllSubfields")
}

// --- I062/220 CalculatedRateOfClimbDescent (extending existing tests) ---

func TestCalculatedRateOfClimbDescent_RoundTripExtended(t *testing.T) {
	cases := []struct{ name string; rate float64 }{
		{"zero", 0},
		{"climb 2000 ft/min", 2000},
		{"descent 3000 ft/min", -3000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &v120.CalculatedRateOfClimbDescent{Rate: tc.rate}
			dec := &v120.CalculatedRateOfClimbDescent{}
			encodeDecodeRoundTrip(t, orig, dec, "CalculatedRateOfClimbDescent")
		})
	}
}
