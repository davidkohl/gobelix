// cat/cat062/fullstack_roundtrip_test.go
//
// Full-stack round-trip test: build a CAT062 record with all data items,
// encode it to a raw ASTERIX DataBlock, decode it back through the real
// Decoder+UAP pipeline, and verify every field survives the journey.
package cat062_test

import (
	"fmt"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// buildFullCAT062Record returns a map of all concrete data items populated
// with representative values.  Every item that the UAP knows about is included
// so the test exercises the full FSPEC bitmap.
func buildFullCAT062Record() map[string]asterix.DataItem {
	// I062/080 TrackStatus – must call SetHasExtension() so the encoder
	// writes the correct number of extension bytes.
	ts := &v120.TrackStatus{
		MON: true,
		SPI: false,
		MRH: true,
		SRC: 3,
		CNF: false,
		// First extension
		SIM: false,
		FPC: true,
		KOS: true,
		// Second extension
		AMA: true,
		MD4: 2,
		ME:  false,
		MI:  true,
		MD5: 1,
		// Third extension
		CST: true,
		PSR: false,
		SSR: true,
		// Fourth extension
		SDS:  2,
		EMS:  3,
		PFT:  false,
		FPLT: true,
		// Fifth extension
		DUPT: false,
		DUPF: true,
		SFC:  false,
		IEC:  true,
	}
	ts.SetHasExtension()

	return map[string]asterix.DataItem{
		// Mandatory items
		"I062/010": &common.DataSourceIdentifier{SAC: 100, SIC: 15},
		"I062/070": &v120.TimeOfTrackInformation{Time: 43200.0}, // 12:00:00 UTC
		"I062/040": &v120.TrackNumber{Value: 4321},
		"I062/080": ts,

		// Optional fixed-length items
		"I062/060": &v120.TrackMode3ACode{Validated: true, Garbled: false, Changed: false, Code: 2300},
		"I062/105": &v120.CalculatedPositionWGS84{Latitude: 51.4775, Longitude: -0.4614},
		"I062/100": &v120.CalculatedTrackPositionCartesian{X: 12345.5, Y: -6789.0},
		"I062/185": &v120.CalculatedTrackVelocity{Vx: 120.25, Vy: -45.75},
		"I062/210": &v120.CalculatedAcceleration{Ax: 0.5, Ay: -0.25},
		"I062/245": &v120.TargetIdentification{IdentType: v120.CallsignRegistration, Ident: "SAS123"},
		"I062/200": &v120.ModeOfMovement{
			Trans: v120.TransRightTurn,
			Long:  v120.LongIncreasingGroundspeed,
			Vert:  v120.VertClimb,
		},
		"I062/136": &v120.MeasuredFlightLevel{FlightLevel: 350.0},
		"I062/130": &v120.CalculatedTrackGeometricAltitude{Altitude: 35000.0},
		"I062/135": &v120.CalculatedTrackBarometricAltitude{Altitude: 350.0, QNH: false},
		"I062/220": &v120.CalculatedRateOfClimbDescent{Rate: 1250.0},
		"I062/300": &v120.VehicleFleetIdentification{VehicleType: v120.UnknownVehicle},
		// Mode 2 code: stored as decimal where each digit is an octal digit (same as Mode3A).
		// 7777 octal = 4095 decimal which is fine, but 7777 decimal > 12 bits, so use octal 1234 = decimal 1234.
		"I062/120": &v120.TrackMode2Code{Code: 1234},

		// Optional compound/variable items
		"I062/500": &v120.EstimatedAccuracies{
			APCX:                      uint16ptr(100),
			APCY:                      uint16ptr(200),
			GeometricAltitudeAccuracy: uint8ptr(8),
			VelocityX:                 uint8ptr(3),
			VelocityY:                 uint8ptr(4),
		},
	}
}

// uint8ptr / uint16ptr are helpers for pointer fields used in EstimatedAccuracies.
func uint8ptr(v uint8) *uint8   { return &v }
func uint16ptr(v uint16) *uint16 { return &v }

// TestCAT062FullStackRoundTrip encodes a CAT062 record through the real
// ASTERIX Encoder then decodes it through the real Decoder and verifies
// every field value is preserved.
func TestCAT062FullStackRoundTrip(t *testing.T) {
	// --- Build UAP ---
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("NewUAP120: %v", err)
	}

	// --- Build encoder and encode ---
	encoder := asterix.NewEncoder()
	items := buildFullCAT062Record()

	encoded, err := encoder.EncodeItems(asterix.Cat062, uap062, items)
	if err != nil {
		t.Fatalf("EncodeItems: %v", err)
	}

	t.Logf("Encoded %d bytes: %X", len(encoded), encoded)

	if len(encoded) < 3 {
		t.Fatalf("encoded message too short: %d bytes", len(encoded))
	}
	// Verify ASTERIX framing: first byte = category, next 2 = total length
	if encoded[0] != byte(asterix.Cat062) {
		t.Errorf("category byte = 0x%02X, want 0x%02X", encoded[0], byte(asterix.Cat062))
	}
	msgLen := int(encoded[1])<<8 | int(encoded[2])
	if msgLen != len(encoded) {
		t.Errorf("length field = %d, actual = %d", msgLen, len(encoded))
	}

	// --- Decode ---
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(encoded)
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}
	if len(dataBlocks) != 1 {
		t.Fatalf("expected 1 DataBlock, got %d", len(dataBlocks))
	}
	db := dataBlocks[0]
	if db.Category() != asterix.Cat062 {
		t.Errorf("decoded category = %v, want Cat062", db.Category())
	}
	if db.RecordCount() != 1 {
		t.Fatalf("expected 1 record, got %d", db.RecordCount())
	}
	rec := db.Records()[0]

	// --- Verify each field ---

	// I062/010 DataSourceIdentifier
	assertField(t, rec, "I062/010", func(item asterix.DataItem) {
		dsi := item.(*common.DataSourceIdentifier)
		assertEqual(t, "SAC", uint8(100), dsi.SAC)
		assertEqual(t, "SIC", uint8(15), dsi.SIC)
	})

	// I062/070 TimeOfTrackInformation
	assertField(t, rec, "I062/070", func(item asterix.DataItem) {
		toi := item.(*v120.TimeOfTrackInformation)
		assertNearFloat(t, "Time", 43200.0, toi.Time, 1.0/128.0)
	})

	// I062/040 TrackNumber
	assertField(t, rec, "I062/040", func(item asterix.DataItem) {
		tn := item.(*v120.TrackNumber)
		assertEqual(t, "TrackNumber.Value", uint16(4321), tn.Value)
	})

	// I062/080 TrackStatus
	assertField(t, rec, "I062/080", func(item asterix.DataItem) {
		ts := item.(*v120.TrackStatus)
		assertEqual(t, "MON", true, ts.MON)
		assertEqual(t, "MRH", true, ts.MRH)
		assertEqual(t, "SRC", uint8(3), ts.SRC)
		assertEqual(t, "FPC", true, ts.FPC)
		assertEqual(t, "KOS", true, ts.KOS)
		assertEqual(t, "AMA", true, ts.AMA)
		assertEqual(t, "MD4", uint8(2), ts.MD4)
		assertEqual(t, "MI", true, ts.MI)
		assertEqual(t, "MD5", uint8(1), ts.MD5)
		assertEqual(t, "CST", true, ts.CST)
		assertEqual(t, "SSR", true, ts.SSR)
		assertEqual(t, "SDS", uint8(2), ts.SDS)
		assertEqual(t, "EMS", uint8(3), ts.EMS)
		assertEqual(t, "FPLT", true, ts.FPLT)
		assertEqual(t, "DUPF", true, ts.DUPF)
		assertEqual(t, "IEC", true, ts.IEC)
	})

	// I062/060 TrackMode3ACode
	assertField(t, rec, "I062/060", func(item asterix.DataItem) {
		mc := item.(*v120.TrackMode3ACode)
		assertEqual(t, "Validated", true, mc.Validated)
		assertEqual(t, "Code", uint16(2300), mc.Code)
	})

	// I062/105 CalculatedPositionWGS84
	assertField(t, rec, "I062/105", func(item asterix.DataItem) {
		pos := item.(*v120.CalculatedPositionWGS84)
		lsb := 180.0 / float64(int32(1)<<25)
		assertNearFloat(t, "Latitude", 51.4775, pos.Latitude, lsb*2)
		assertNearFloat(t, "Longitude", -0.4614, pos.Longitude, lsb*2)
	})

	// I062/100 CalculatedTrackPositionCartesian
	assertField(t, rec, "I062/100", func(item asterix.DataItem) {
		pos := item.(*v120.CalculatedTrackPositionCartesian)
		assertNearFloat(t, "X", 12345.5, pos.X, 0.5)
		assertNearFloat(t, "Y", -6789.0, pos.Y, 0.5)
	})

	// I062/185 CalculatedTrackVelocity
	assertField(t, rec, "I062/185", func(item asterix.DataItem) {
		vel := item.(*v120.CalculatedTrackVelocity)
		assertNearFloat(t, "Vx", 120.25, vel.Vx, 0.25)
		assertNearFloat(t, "Vy", -45.75, vel.Vy, 0.25)
	})

	// I062/210 CalculatedAcceleration
	assertField(t, rec, "I062/210", func(item asterix.DataItem) {
		acc := item.(*v120.CalculatedAcceleration)
		assertNearFloat(t, "Ax", 0.5, acc.Ax, 0.25)
		assertNearFloat(t, "Ay", -0.25, acc.Ay, 0.25)
	})

	// I062/245 TargetIdentification
	assertField(t, rec, "I062/245", func(item asterix.DataItem) {
		ti := item.(*v120.TargetIdentification)
		assertEqual(t, "IdentType", v120.CallsignRegistration, ti.IdentType)
		assertEqual(t, "Ident", "SAS123", ti.Ident)
	})

	// I062/200 ModeOfMovement
	assertField(t, rec, "I062/200", func(item asterix.DataItem) {
		mom := item.(*v120.ModeOfMovement)
		assertEqual(t, "Trans", v120.TransRightTurn, mom.Trans)
		assertEqual(t, "Long", v120.LongIncreasingGroundspeed, mom.Long)
		assertEqual(t, "Vert", v120.VertClimb, mom.Vert)
	})

	// I062/136 MeasuredFlightLevel
	assertField(t, rec, "I062/136", func(item asterix.DataItem) {
		fl := item.(*v120.MeasuredFlightLevel)
		assertNearFloat(t, "FlightLevel", 350.0, fl.FlightLevel, 0.25)
	})

	// I062/130 CalculatedTrackGeometricAltitude
	assertField(t, rec, "I062/130", func(item asterix.DataItem) {
		alt := item.(*v120.CalculatedTrackGeometricAltitude)
		assertNearFloat(t, "Altitude", 35000.0, alt.Altitude, 6.25)
	})

	// I062/135 CalculatedTrackBarometricAltitude
	assertField(t, rec, "I062/135", func(item asterix.DataItem) {
		alt := item.(*v120.CalculatedTrackBarometricAltitude)
		assertNearFloat(t, "Altitude", 350.0, alt.Altitude, 0.25)
		assertEqual(t, "QNH", false, alt.QNH)
	})

	// I062/220 CalculatedRateOfClimbDescent
	assertField(t, rec, "I062/220", func(item asterix.DataItem) {
		roc := item.(*v120.CalculatedRateOfClimbDescent)
		assertNearFloat(t, "Rate", 1250.0, roc.Rate, 6.25)
	})

	// I062/300 VehicleFleetIdentification
	assertField(t, rec, "I062/300", func(item asterix.DataItem) {
		vfi := item.(*v120.VehicleFleetIdentification)
		assertEqual(t, "VehicleType", v120.UnknownVehicle, vfi.VehicleType)
	})

	// I062/120 TrackMode2Code
	assertField(t, rec, "I062/120", func(item asterix.DataItem) {
		mc := item.(*v120.TrackMode2Code)
		assertEqual(t, "Code", uint16(1234), mc.Code)
	})

	// I062/500 EstimatedAccuracies
	assertField(t, rec, "I062/500", func(item asterix.DataItem) {
		ea := item.(*v120.EstimatedAccuracies)
		if ea.APCX == nil {
			t.Error("EstimatedAccuracies.APCX is nil")
		} else {
			assertEqual(t, "APCX", uint16(100), *ea.APCX)
		}
		if ea.APCY == nil {
			t.Error("EstimatedAccuracies.APCY is nil")
		} else {
			assertEqual(t, "APCY", uint16(200), *ea.APCY)
		}
		if ea.GeometricAltitudeAccuracy == nil {
			t.Error("EstimatedAccuracies.GeometricAltitudeAccuracy is nil")
		} else {
			assertEqual(t, "GeometricAltitudeAccuracy", uint8(8), *ea.GeometricAltitudeAccuracy)
		}
		if ea.VelocityX == nil {
			t.Error("EstimatedAccuracies.VelocityX is nil")
		} else {
			assertEqual(t, "VelocityX", uint8(3), *ea.VelocityX)
		}
		if ea.VelocityY == nil {
			t.Error("EstimatedAccuracies.VelocityY is nil")
		} else {
			assertEqual(t, "VelocityY", uint8(4), *ea.VelocityY)
		}
	})
}

// TestCAT062FullStackMultiRecord encodes two records in one DataBlock and
// verifies that both decode correctly.
func TestCAT062FullStackMultiRecord(t *testing.T) {
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("NewUAP120: %v", err)
	}

	encoder := asterix.NewEncoder()
	encoder.StartBatch(asterix.Cat062, uap062)

	record1 := map[string]asterix.DataItem{
		"I062/010": &common.DataSourceIdentifier{SAC: 100, SIC: 15},
		"I062/070": &v120.TimeOfTrackInformation{Time: 36000.0},
		"I062/040": &v120.TrackNumber{Value: 1001},
		"I062/080": func() asterix.DataItem {
			ts := &v120.TrackStatus{MON: true, CNF: false}
			ts.SetHasExtension()
			return ts
		}(),
		"I062/105": &v120.CalculatedPositionWGS84{Latitude: 55.6761, Longitude: 12.5683}, // Copenhagen
		"I062/185": &v120.CalculatedTrackVelocity{Vx: 200.0, Vy: 50.0},
	}
	record2 := map[string]asterix.DataItem{
		"I062/010": &common.DataSourceIdentifier{SAC: 100, SIC: 15},
		"I062/070": &v120.TimeOfTrackInformation{Time: 36000.125},
		"I062/040": &v120.TrackNumber{Value: 2002},
		"I062/080": func() asterix.DataItem {
			ts := &v120.TrackStatus{MON: false, CNF: true, SPI: true}
			ts.SetHasExtension()
			return ts
		}(),
		"I062/105": &v120.CalculatedPositionWGS84{Latitude: 59.9127, Longitude: 10.7461}, // Oslo
		"I062/185": &v120.CalculatedTrackVelocity{Vx: -150.0, Vy: 80.0},
	}

	if err := encoder.AddToBatch(record1); err != nil {
		t.Fatalf("AddToBatch record1: %v", err)
	}
	if err := encoder.AddToBatch(record2); err != nil {
		t.Fatalf("AddToBatch record2: %v", err)
	}

	encoded, err := encoder.FinishBatch()
	if err != nil {
		t.Fatalf("FinishBatch: %v", err)
	}

	t.Logf("Multi-record encoded %d bytes: %X", len(encoded), encoded)

	// Decode
	decoder := asterix.NewDecoder()
	decoder.RegisterUAP(uap062)

	dataBlocks, err := decoder.DecodeAll(encoded)
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}
	if len(dataBlocks) != 1 {
		t.Fatalf("expected 1 DataBlock, got %d", len(dataBlocks))
	}
	db := dataBlocks[0]
	if db.RecordCount() != 2 {
		t.Fatalf("expected 2 records, got %d", db.RecordCount())
	}

	// Verify record 1
	r1 := db.Records()[0]
	assertField(t, r1, "I062/040", func(item asterix.DataItem) {
		assertEqual(t, "record1 TrackNumber", uint16(1001), item.(*v120.TrackNumber).Value)
	})
	assertField(t, r1, "I062/105", func(item asterix.DataItem) {
		pos := item.(*v120.CalculatedPositionWGS84)
		assertNearFloat(t, "record1 Latitude", 55.6761, pos.Latitude, 0.001)
	})

	// Verify record 2
	r2 := db.Records()[1]
	assertField(t, r2, "I062/040", func(item asterix.DataItem) {
		assertEqual(t, "record2 TrackNumber", uint16(2002), item.(*v120.TrackNumber).Value)
	})
	assertField(t, r2, "I062/080", func(item asterix.DataItem) {
		ts := item.(*v120.TrackStatus)
		assertEqual(t, "record2 SPI", true, ts.SPI)
		assertEqual(t, "record2 CNF", true, ts.CNF)
	})
	assertField(t, r2, "I062/105", func(item asterix.DataItem) {
		pos := item.(*v120.CalculatedPositionWGS84)
		assertNearFloat(t, "record2 Latitude", 59.9127, pos.Latitude, 0.001)
	})
}

// --- Test helpers ---

// assertField retrieves a data item from a record and calls fn with it.
func assertField(t *testing.T, rec *asterix.Record, id string, fn func(asterix.DataItem)) {
	t.Helper()
	item, ok := rec.GetDataItem(id)
	if !ok {
		t.Errorf("field %s not found in decoded record", id)
		return
	}
	fn(item)
}

// assertEqual fails the test if got != want, printing a descriptive message.
func assertEqual[T comparable](t *testing.T, name string, want, got T) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

// assertNearFloat fails if |got - want| > tolerance.
func assertNearFloat(t *testing.T, name string, want, got, tolerance float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("%s: got %f, want %f (diff %f > tolerance %f)", name, got, want, diff, tolerance)
	}
}

// TestCAT062FullStackEncodedBytesAreStable verifies that encoding the same
// record twice produces identical bytes (deterministic output).
func TestCAT062FullStackEncodedBytesAreStable(t *testing.T) {
	uap062, err := uap.NewUAP120()
	if err != nil {
		t.Fatalf("NewUAP120: %v", err)
	}
	encoder := asterix.NewEncoder()
	items := buildFullCAT062Record()

	enc1, err := encoder.EncodeItems(asterix.Cat062, uap062, items)
	if err != nil {
		t.Fatalf("first EncodeItems: %v", err)
	}
	enc2, err := encoder.EncodeItems(asterix.Cat062, uap062, items)
	if err != nil {
		t.Fatalf("second EncodeItems: %v", err)
	}

	if fmt.Sprintf("%X", enc1) != fmt.Sprintf("%X", enc2) {
		t.Errorf("encoding is not deterministic:\n  first:  %X\n  second: %X", enc1, enc2)
	}
}
