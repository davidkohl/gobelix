// asterix/category_roundtrip_test.go
//
// Full-stack round-trip tests for all supported ASTERIX categories.
// Each test builds a record with representative data items, encodes it
// through the real ASTERIX Encoder+UAP pipeline, decodes it back with
// the real Decoder, and verifies that every field survives the journey.
package asterix_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/davidkohl/gobelix/asterix"

	uap001 "github.com/davidkohl/gobelix/cat/cat001/uap"
	uap002 "github.com/davidkohl/gobelix/cat/cat002/uap"
	uap020 "github.com/davidkohl/gobelix/cat/cat020/uap"
	uap021 "github.com/davidkohl/gobelix/cat/cat021/uap"
	uap023 "github.com/davidkohl/gobelix/cat/cat023/uap"
	uap034 "github.com/davidkohl/gobelix/cat/cat034/uap"
	uap048 "github.com/davidkohl/gobelix/cat/cat048/uap"
	uap062 "github.com/davidkohl/gobelix/cat/cat062/uap"
	uap063 "github.com/davidkohl/gobelix/cat/cat063/uap"

	v12cat001 "github.com/davidkohl/gobelix/cat/cat001/dataitems/v12"
	v10cat002 "github.com/davidkohl/gobelix/cat/cat002/dataitems/v10"
	v15cat020 "github.com/davidkohl/gobelix/cat/cat020/dataitems/v15"
	v26cat021 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
	v13cat023 "github.com/davidkohl/gobelix/cat/cat023/dataitems/v13"
	v129cat034 "github.com/davidkohl/gobelix/cat/cat034/dataitems/v129"
	v132cat048 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v132"
	v120cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	v16cat063 "github.com/davidkohl/gobelix/cat/cat063/dataitems/v16"

	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// encodeDecodeRoundTripCat encodes a map of data items for the given category and UAP,
// decodes the resulting bytes back through the Decoder, and returns the first Record.
func encodeDecodeRoundTripCat(t *testing.T, cat asterix.Category, uap asterix.UAP, items map[string]asterix.DataItem) *asterix.Record {
	t.Helper()

	enc := asterix.NewEncoder()
	encoded, err := enc.EncodeItems(cat, uap, items)
	if err != nil {
		t.Fatalf("EncodeItems failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("EncodeItems produced empty output")
	}

	dec := asterix.NewDecoder()
	dec.RegisterUAP(uap)
	blocks, err := dec.DecodeAll(encoded)
	if err != nil {
		t.Fatalf("DecodeAll failed: %v\nencoded bytes: %X", err, encoded)
	}
	if len(blocks) == 0 {
		t.Fatal("DecodeAll returned no DataBlocks")
	}
	recs := blocks[0].Records()
	if len(recs) == 0 {
		t.Fatal("DataBlock contains no records")
	}
	return recs[0]
}

// assertItem retrieves a data item from a record and calls fn with the concrete type.
// It fails the test if the item is not present.
func assertItem(t *testing.T, rec *asterix.Record, id string, fn func(asterix.DataItem)) {
	t.Helper()
	item, ok := rec.GetDataItem(id)
	if !ok {
		t.Errorf("item %s not found in decoded record", id)
		return
	}
	fn(item)
}

// assertEq is a generic equality check helper.
func assertEq[T comparable](t *testing.T, name string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", name, want, got)
	}
}

// assertNear checks that two floats are within tolerance of each other.
func assertNear(t *testing.T, name string, want, got, tol float64) {
	t.Helper()
	diff := want - got
	if diff < 0 {
		diff = -diff
	}
	if diff > tol {
		t.Errorf("%s: want %.6f, got %.6f (tolerance %.6f)", name, want, got, tol)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT001 v1.2 – Monoradar Target Messages
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT001RoundTrip(t *testing.T) {
	uap, err := uap001.NewUAP12()
	if err != nil {
		t.Fatalf("NewUAP12: %v", err)
	}

	items := map[string]asterix.DataItem{
		"I001/010": &common.DataSourceIdentifier{SAC: 10, SIC: 20},
		"I001/020": &v12cat001.TargetReportDescriptor{
			TYP: 2,
			SIM: false,
			SSR: true,
			PSR: false,
			ANT: false,
			SPI: true,
			RAB: false,
			TST: false,
		},
		"I001/040": &v12cat001.PositionPolar{
			RHO:   50.5,
			THETA: 90.0,
		},
		"I001/070": &v12cat001.Mode3ACode{
			V:    true,
			G:    false,
			L:    false,
			Mode: 0o1234, // octal 1234 = decimal 668
		},
		// TruncatedTimeOfDay is a 2-byte (16-bit) field with LSB=1/128s,
		// so the maximum representable value is 65535/128 ≈ 511.99 seconds.
		"I001/141": &v12cat001.TruncatedTimeOfDay{
			Time: 300.0,
		},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat001, uap, items)

	assertItem(t, rec, "I001/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I001/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(10), dsi.SAC)
		assertEq(t, "SIC", uint8(20), dsi.SIC)
	})

	assertItem(t, rec, "I001/020", func(item asterix.DataItem) {
		trd, ok := item.(*v12cat001.TargetReportDescriptor)
		if !ok {
			t.Fatalf("I001/020: unexpected type %T", item)
		}
		assertEq(t, "TYP", uint8(2), trd.TYP)
		assertEq(t, "SSR", true, trd.SSR)
		assertEq(t, "SPI", true, trd.SPI)
	})

	assertItem(t, rec, "I001/070", func(item asterix.DataItem) {
		mc, ok := item.(*v12cat001.Mode3ACode)
		if !ok {
			t.Fatalf("I001/070: unexpected type %T", item)
		}
		assertEq(t, "V", true, mc.V)
		assertEq(t, "Mode", uint16(0o1234), mc.Mode)
	})

	assertItem(t, rec, "I001/141", func(item asterix.DataItem) {
		tod, ok := item.(*v12cat001.TruncatedTimeOfDay)
		if !ok {
			t.Fatalf("I001/141: unexpected type %T", item)
		}
		// LSB = 1/128 s; round-trip tolerance is one LSB
		assertNear(t, "Time", 300.0, tod.Time, 1.0/128.0)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT002 v1.0 – Monoradar Service Messages
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT002RoundTrip(t *testing.T) {
	uap, err := uap002.NewUAP10()
	if err != nil {
		t.Fatalf("NewUAP10: %v", err)
	}

	items := map[string]asterix.DataItem{
		"I002/010": &common.DataSourceIdentifier{SAC: 5, SIC: 7},
		"I002/000": &v10cat002.MessageType{MessageType: 1}, // North marker
		"I002/020": &common.SectorNumber{SectorNumber: 45.0},
		"I002/030": &common.TimeOfDay{Time: 36000.0}, // 10:00:00 UTC
		"I002/041": &v10cat002.AntennaRotationSpeed{RotationPeriod: 4.0},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat002, uap, items)

	assertItem(t, rec, "I002/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I002/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(5), dsi.SAC)
		assertEq(t, "SIC", uint8(7), dsi.SIC)
	})

	assertItem(t, rec, "I002/000", func(item asterix.DataItem) {
		mt, ok := item.(*v10cat002.MessageType)
		if !ok {
			t.Fatalf("I002/000: unexpected type %T", item)
		}
		assertEq(t, "MessageType", uint8(1), mt.MessageType)
	})

	assertItem(t, rec, "I002/030", func(item asterix.DataItem) {
		tod, ok := item.(*common.TimeOfDay)
		if !ok {
			t.Fatalf("I002/030: unexpected type %T", item)
		}
		assertNear(t, "Time", 36000.0, tod.Time, 1.0/128.0)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT020 v1.5 – Multilateration Target Reports
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT020RoundTrip(t *testing.T) {
	uap, err := uap020.NewUAP15()
	if err != nil {
		t.Fatalf("NewUAP15: %v", err)
	}

	items := map[string]asterix.DataItem{
		// Mandatory
		"I020/010": &common.DataSourceIdentifier{SAC: 20, SIC: 30},
		"I020/140": &common.TimeOfDay{Time: 43200.0}, // 12:00:00 UTC
		// Optional
		"I020/020": &v15cat020.TargetReportDescriptor{
			SSR: true,
			MS:  true,
		},
		"I020/041": &v15cat020.PositionWGS84{Latitude: 51.5, Longitude: -0.1},
		"I020/161": &common.TrackNumber{Value: 1000},
		"I020/070": &v15cat020.Mode3ACode{
			V:    true,
			Code: 0o0700,
		},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat020, uap, items)

	assertItem(t, rec, "I020/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I020/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(20), dsi.SAC)
		assertEq(t, "SIC", uint8(30), dsi.SIC)
	})

	assertItem(t, rec, "I020/140", func(item asterix.DataItem) {
		tod, ok := item.(*common.TimeOfDay)
		if !ok {
			t.Fatalf("I020/140: unexpected type %T", item)
		}
		assertNear(t, "Time", 43200.0, tod.Time, 1.0/128.0)
	})

	assertItem(t, rec, "I020/041", func(item asterix.DataItem) {
		pos, ok := item.(*v15cat020.PositionWGS84)
		if !ok {
			t.Fatalf("I020/041: unexpected type %T", item)
		}
		// LSB = 180/2^25 ≈ 5.4e-6 degrees
		assertNear(t, "Latitude", 51.5, pos.Latitude, 1e-4)
		assertNear(t, "Longitude", -0.1, pos.Longitude, 1e-4)
	})

	assertItem(t, rec, "I020/161", func(item asterix.DataItem) {
		tn, ok := item.(*common.TrackNumber)
		if !ok {
			t.Fatalf("I020/161: unexpected type %T", item)
		}
		assertEq(t, "Value", uint16(1000), tn.Value)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT021 v2.6 – ADS-B Target Reports
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT021RoundTrip(t *testing.T) {
	uap, err := uap021.NewUAP26()
	if err != nil {
		t.Fatalf("NewUAP26: %v", err)
	}

	items := map[string]asterix.DataItem{
		// Mandatory
		"I021/010": &common.DataSourceIdentifier{SAC: 21, SIC: 1},
		"I021/040": &v26cat021.TargetReportDescriptor{
			ATP:  1,
			ARC:  1,
			RC:   false,
			RAB:  false,
			DCR:  false,
			GBS:  false,
			SIM:  false,
			TST:  false,
			SAA:  true,
			CL:   0,
			IPC:  false,
			NOGO: false,
			CPR:  false,
			LDPJ: false,
			RCF:  false,
		},
		"I021/161": &common.TrackNumber{Value: 1234},
		"I021/080": &common.TargetAddress{Address: 0x3C0000},
		"RE":       &v26cat021.ReservedExpansion{Data: []byte{2, 0xBE, 0xEF}},
		// Optional items that don't require additional dependencies
		"I021/170": &v26cat021.TargetIdentification{
			Ident: "AFR1234",
		},
		// I021/130 (Position) requires I021/090 (QualityIndicators) + I021/071 or I021/073 (time).
		// Include all three together to satisfy validation.
		"I021/090": &v26cat021.QualityIndicators{NUCr_NACv: 3, NUCp_NIC: 5},
		"I021/071": &v26cat021.TimeOfApplicabilityPosition{Time: 43200.0},
		"I021/130": &common.Position{Latitude: 48.8566, Longitude: 2.3522},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat021, uap, items)

	assertItem(t, rec, "I021/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I021/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(21), dsi.SAC)
		assertEq(t, "SIC", uint8(1), dsi.SIC)
	})

	assertItem(t, rec, "I021/080", func(item asterix.DataItem) {
		ta, ok := item.(*common.TargetAddress)
		if !ok {
			t.Fatalf("I021/080: unexpected type %T", item)
		}
		assertEq(t, "Address", uint32(0x3C0000), ta.Address)
	})

	assertItem(t, rec, "I021/170", func(item asterix.DataItem) {
		ti, ok := item.(*v26cat021.TargetIdentification)
		if !ok {
			t.Fatalf("I021/170: unexpected type %T", item)
		}
		assertEq(t, "Ident", "AFR1234", ti.Ident)
	})

	assertItem(t, rec, "I021/161", func(item asterix.DataItem) {
		tn, ok := item.(*common.TrackNumber)
		if !ok {
			t.Fatalf("I021/161: unexpected type %T", item)
		}
		assertEq(t, "Value", uint16(1234), tn.Value)
	})

	assertItem(t, rec, "RE", func(item asterix.DataItem) {
		re, ok := item.(*v26cat021.ReservedExpansion)
		if !ok {
			t.Fatalf("RE: unexpected type %T", item)
		}
		if !bytes.Equal(re.Data, []byte{2, 0xBE, 0xEF}) {
			t.Errorf("RE data: want % X, got % X", []byte{2, 0xBE, 0xEF}, re.Data)
		}
	})

	assertItem(t, rec, "I021/130", func(item asterix.DataItem) {
		pos, ok := item.(*common.Position)
		if !ok {
			t.Fatalf("I021/130: unexpected type %T", item)
		}
		// LSB ≈ 5.4e-6 degrees
		assertNear(t, "Latitude", 48.8566, pos.Latitude, 1e-4)
		assertNear(t, "Longitude", 2.3522, pos.Longitude, 1e-4)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT023 v1.3 – CNS/ATM Ground Station and Service Status Reports
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT023RoundTrip(t *testing.T) {
	uap, err := uap023.NewUAP13()
	if err != nil {
		t.Fatalf("NewUAP13: %v", err)
	}

	// ReportType=2 (Service Status) avoids the additional validation that
	// requires I023/100 only for ReportType=1 (Ground Station Status).
	items := map[string]asterix.DataItem{
		"I023/010": &v13cat023.DataSourceIdentifier{SAC: 23, SIC: 5},
		"I023/000": &v13cat023.ReportType{ReportType: 2},
		"I023/070": &v13cat023.TimeOfDay{Time: 7200.0},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat023, uap, items)

	assertItem(t, rec, "I023/010", func(item asterix.DataItem) {
		dsi, ok := item.(*v13cat023.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I023/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(23), dsi.SAC)
		assertEq(t, "SIC", uint8(5), dsi.SIC)
	})

	assertItem(t, rec, "I023/000", func(item asterix.DataItem) {
		rt, ok := item.(*v13cat023.ReportType)
		if !ok {
			t.Fatalf("I023/000: unexpected type %T", item)
		}
		assertEq(t, "ReportType", uint8(2), rt.ReportType)
	})

	assertItem(t, rec, "I023/070", func(item asterix.DataItem) {
		tod, ok := item.(*v13cat023.TimeOfDay)
		if !ok {
			t.Fatalf("I023/070: unexpected type %T", item)
		}
		assertNear(t, "Time", 7200.0, tod.Time, 1.0/128.0)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT034 v1.29 – Transmission of Monoradar Service Messages
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT034RoundTrip(t *testing.T) {
	uap, err := uap034.NewUAP129()
	if err != nil {
		t.Fatalf("NewUAP129: %v", err)
	}

	items := map[string]asterix.DataItem{
		"I034/010": &common.DataSourceIdentifier{SAC: 34, SIC: 9},
		"I034/000": v129cat034.NewMessageType(),
		"I034/030": &common.TimeOfDay{Time: 54000.0}, // 15:00:00 UTC
		"I034/020": &common.SectorNumber{SectorNumber: 90.0},
	}
	// Set message type to North marker
	items["I034/000"].(*v129cat034.MessageType).MessageType = 1

	rec := encodeDecodeRoundTripCat(t, asterix.Cat034, uap, items)

	assertItem(t, rec, "I034/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I034/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(34), dsi.SAC)
		assertEq(t, "SIC", uint8(9), dsi.SIC)
	})

	assertItem(t, rec, "I034/000", func(item asterix.DataItem) {
		mt, ok := item.(*v129cat034.MessageType)
		if !ok {
			t.Fatalf("I034/000: unexpected type %T", item)
		}
		assertEq(t, "MessageType", uint8(1), mt.MessageType)
	})

	assertItem(t, rec, "I034/030", func(item asterix.DataItem) {
		tod, ok := item.(*common.TimeOfDay)
		if !ok {
			t.Fatalf("I034/030: unexpected type %T", item)
		}
		assertNear(t, "Time", 54000.0, tod.Time, 1.0/128.0)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT048 v1.32 – Monoradar Target Reports
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT048RoundTrip(t *testing.T) {
	uap, err := uap048.NewUAP132()
	if err != nil {
		t.Fatalf("NewUAP132: %v", err)
	}

	items := map[string]asterix.DataItem{
		"I048/010": &common.DataSourceIdentifier{SAC: 48, SIC: 2},
		"I048/140": &v132cat048.TimeOfDay{Time: 43200.0},
		"I048/020": &v132cat048.TargetReportDescriptor{
			TYP: 5, // SSR+PSR combined
			SIM: false,
			RDP: false,
			SPI: false,
			RAB: false,
		},
		"I048/040": &v132cat048.MeasuredPosition{
			RHO:   30.5,
			THETA: 180.0,
		},
		"I048/070": &v132cat048.Mode3ACode{
			V:    true,
			G:    false,
			L:    false,
			Code: 4321,
		},
		"I048/090": &v132cat048.FlightLevel{
			V:     true,
			G:     false,
			Level: 350.0,
		},
		"I048/161": &common.TrackNumber{Value: 2048},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat048, uap, items)

	assertItem(t, rec, "I048/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I048/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(48), dsi.SAC)
		assertEq(t, "SIC", uint8(2), dsi.SIC)
	})

	assertItem(t, rec, "I048/140", func(item asterix.DataItem) {
		tod, ok := item.(*v132cat048.TimeOfDay)
		if !ok {
			t.Fatalf("I048/140: unexpected type %T", item)
		}
		assertNear(t, "Time", 43200.0, tod.Time, 1.0/128.0)
	})

	assertItem(t, rec, "I048/040", func(item asterix.DataItem) {
		mp, ok := item.(*v132cat048.MeasuredPosition)
		if !ok {
			t.Fatalf("I048/040: unexpected type %T", item)
		}
		// RHO LSB = 1/256 NM
		assertNear(t, "RHO", 30.5, mp.RHO, 1.0/256.0)
		// THETA LSB = 360/65536 ≈ 0.0055 degrees
		assertNear(t, "THETA", 180.0, mp.THETA, 0.01)
	})

	assertItem(t, rec, "I048/070", func(item asterix.DataItem) {
		mc, ok := item.(*v132cat048.Mode3ACode)
		if !ok {
			t.Fatalf("I048/070: unexpected type %T", item)
		}
		assertEq(t, "V", true, mc.V)
		assertEq(t, "Code", uint16(4321), mc.Code)
	})

	assertItem(t, rec, "I048/161", func(item asterix.DataItem) {
		tn, ok := item.(*common.TrackNumber)
		if !ok {
			t.Fatalf("I048/161: unexpected type %T", item)
		}
		assertEq(t, "Value", uint16(2048), tn.Value)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT062 v1.20 – System Track Data
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT062RoundTrip(t *testing.T) {
	uap, err := uap062.NewUAP120()
	if err != nil {
		t.Fatalf("NewUAP120: %v", err)
	}

	ts := &v120cat062.TrackStatus{
		MON: true,
		MRH: true,
		SRC: 2,
		CNF: false,
		SIM: false,
		FPC: true,
		KOS: false,
	}
	ts.SetHasExtension()

	items := map[string]asterix.DataItem{
		// Mandatory
		"I062/010": &common.DataSourceIdentifier{SAC: 62, SIC: 1},
		"I062/040": &v120cat062.TrackNumber{Value: 999},
		"I062/070": &v120cat062.TimeOfTrackInformation{Time: 43200.0},
		"I062/080": ts,
		// Optional
		"I062/060": &v120cat062.TrackMode3ACode{
			Validated: true,
			Garbled:   false,
			Changed:   false,
			Code:      1200,
		},
		"I062/105": &v120cat062.CalculatedPositionWGS84{Latitude: 52.3, Longitude: 4.9},
		"I062/185": &v120cat062.CalculatedTrackVelocity{Vx: 200.0, Vy: -100.0},
		"I062/245": &v120cat062.TargetIdentification{
			IdentType: v120cat062.CallsignRegistration,
			Ident:     "KLM456",
		},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat062, uap, items)

	assertItem(t, rec, "I062/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I062/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(62), dsi.SAC)
		assertEq(t, "SIC", uint8(1), dsi.SIC)
	})

	assertItem(t, rec, "I062/040", func(item asterix.DataItem) {
		tn, ok := item.(*v120cat062.TrackNumber)
		if !ok {
			t.Fatalf("I062/040: unexpected type %T", item)
		}
		assertEq(t, "Value", uint16(999), tn.Value)
	})

	assertItem(t, rec, "I062/070", func(item asterix.DataItem) {
		ti, ok := item.(*v120cat062.TimeOfTrackInformation)
		if !ok {
			t.Fatalf("I062/070: unexpected type %T", item)
		}
		assertNear(t, "Time", 43200.0, ti.Time, 1.0/128.0)
	})

	assertItem(t, rec, "I062/105", func(item asterix.DataItem) {
		pos, ok := item.(*v120cat062.CalculatedPositionWGS84)
		if !ok {
			t.Fatalf("I062/105: unexpected type %T", item)
		}
		assertNear(t, "Latitude", 52.3, pos.Latitude, 1e-4)
		assertNear(t, "Longitude", 4.9, pos.Longitude, 1e-4)
	})

	assertItem(t, rec, "I062/245", func(item asterix.DataItem) {
		ti, ok := item.(*v120cat062.TargetIdentification)
		if !ok {
			t.Fatalf("I062/245: unexpected type %T", item)
		}
		assertEq(t, "Ident", "KLM456", ti.Ident)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CAT063 v1.6 – Sensor Status Messages
// ──────────────────────────────────────────────────────────────────────────────

func TestCAT063RoundTrip(t *testing.T) {
	uap, err := uap063.NewUAP063()
	if err != nil {
		t.Fatalf("NewUAP063: %v", err)
	}

	items := map[string]asterix.DataItem{
		// All three mandatory items
		"I063/010": &common.DataSourceIdentifier{SAC: 63, SIC: 3},
		"I063/030": &v16cat063.TimeOfMessage{Time: 28800.0}, // 08:00:00 UTC
		"I063/050": &v16cat063.SensorIdentifier{SAC: 10, SIC: 20},
		// Optional
		"I063/070": &v16cat063.TimeStampingBias{Bias: 50},
	}

	rec := encodeDecodeRoundTripCat(t, asterix.Cat063, uap, items)

	assertItem(t, rec, "I063/010", func(item asterix.DataItem) {
		dsi, ok := item.(*common.DataSourceIdentifier)
		if !ok {
			t.Fatalf("I063/010: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(63), dsi.SAC)
		assertEq(t, "SIC", uint8(3), dsi.SIC)
	})

	assertItem(t, rec, "I063/030", func(item asterix.DataItem) {
		tom, ok := item.(*v16cat063.TimeOfMessage)
		if !ok {
			t.Fatalf("I063/030: unexpected type %T", item)
		}
		assertNear(t, "Time", 28800.0, tom.Time, 1.0/128.0)
	})

	assertItem(t, rec, "I063/050", func(item asterix.DataItem) {
		si, ok := item.(*v16cat063.SensorIdentifier)
		if !ok {
			t.Fatalf("I063/050: unexpected type %T", item)
		}
		assertEq(t, "SAC", uint8(10), si.SAC)
		assertEq(t, "SIC", uint8(20), si.SIC)
	})

	assertItem(t, rec, "I063/070", func(item asterix.DataItem) {
		tsb, ok := item.(*v16cat063.TimeStampingBias)
		if !ok {
			t.Fatalf("I063/070: unexpected type %T", item)
		}
		assertEq(t, "Bias", int16(50), tsb.Bias)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Stability test: encoding the same record twice must produce identical bytes
// ──────────────────────────────────────────────────────────────────────────────

func TestAllCategoriesEncodingIsStable(t *testing.T) {
	type tc struct {
		name string
		fn   func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem)
	}

	tests := []tc{
		{
			name: "CAT001",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap001.NewUAP12()
				return asterix.Cat001, uap, map[string]asterix.DataItem{
					"I001/010": &common.DataSourceIdentifier{SAC: 1, SIC: 2},
					"I001/070": &v12cat001.Mode3ACode{V: true, Mode: 0o1234},
					"I001/141": &v12cat001.TruncatedTimeOfDay{Time: 3600.0},
				}
			},
		},
		{
			name: "CAT002",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap002.NewUAP10()
				return asterix.Cat002, uap, map[string]asterix.DataItem{
					"I002/010": &common.DataSourceIdentifier{SAC: 5, SIC: 7},
					"I002/000": &v10cat002.MessageType{MessageType: 1},
					"I002/030": &common.TimeOfDay{Time: 36000.0},
				}
			},
		},
		{
			name: "CAT020",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap020.NewUAP15()
				return asterix.Cat020, uap, map[string]asterix.DataItem{
					"I020/010": &common.DataSourceIdentifier{SAC: 20, SIC: 30},
					"I020/140": &common.TimeOfDay{Time: 43200.0},
					"I020/161": &common.TrackNumber{Value: 1000},
				}
			},
		},
		{
			name: "CAT048",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap048.NewUAP132()
				return asterix.Cat048, uap, map[string]asterix.DataItem{
					"I048/010": &common.DataSourceIdentifier{SAC: 48, SIC: 2},
					"I048/140": &v132cat048.TimeOfDay{Time: 43200.0},
					"I048/161": &common.TrackNumber{Value: 2048},
				}
			},
		},
		{
			name: "CAT062",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap062.NewUAP120()
				ts := &v120cat062.TrackStatus{MON: true, MRH: true, SRC: 2}
				ts.SetHasExtension()
				return asterix.Cat062, uap, map[string]asterix.DataItem{
					"I062/010": &common.DataSourceIdentifier{SAC: 62, SIC: 1},
					"I062/040": &v120cat062.TrackNumber{Value: 999},
					"I062/070": &v120cat062.TimeOfTrackInformation{Time: 43200.0},
					"I062/080": ts,
				}
			},
		},
		{
			name: "CAT063",
			fn: func() (asterix.Category, asterix.UAP, map[string]asterix.DataItem) {
				uap, _ := uap063.NewUAP063()
				return asterix.Cat063, uap, map[string]asterix.DataItem{
					"I063/010": &common.DataSourceIdentifier{SAC: 63, SIC: 3},
					"I063/030": &v16cat063.TimeOfMessage{Time: 28800.0},
					"I063/050": &v16cat063.SensorIdentifier{SAC: 10, SIC: 20},
				}
			},
		},
	}

	enc := asterix.NewEncoder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, uap, items := tt.fn()

			b1, err := enc.EncodeItems(cat, uap, items)
			if err != nil {
				t.Fatalf("first encode: %v", err)
			}
			b2, err := enc.EncodeItems(cat, uap, items)
			if err != nil {
				t.Fatalf("second encode: %v", err)
			}
			if !bytes.Equal(b1, b2) {
				t.Errorf("encoding not stable:\n  first:  %X\n  second: %X", b1, b2)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Multi-decoder test: register all UAPs in a single Decoder and decode
// concatenated CAT062+CAT048 message stream.
// ──────────────────────────────────────────────────────────────────────────────

func TestMultiCategoryDecoder(t *testing.T) {
	uap62, err := uap062.NewUAP120()
	if err != nil {
		t.Fatalf("NewUAP120: %v", err)
	}
	uap48, err := uap048.NewUAP132()
	if err != nil {
		t.Fatalf("NewUAP132: %v", err)
	}

	enc := asterix.NewEncoder()

	// Build CAT062 message
	ts := &v120cat062.TrackStatus{MON: true, SRC: 1}
	ts.SetHasExtension()
	cat62items := map[string]asterix.DataItem{
		"I062/010": &common.DataSourceIdentifier{SAC: 62, SIC: 1},
		"I062/040": &v120cat062.TrackNumber{Value: 1111},
		"I062/070": &v120cat062.TimeOfTrackInformation{Time: 43200.0},
		"I062/080": ts,
	}
	b62, err := enc.EncodeItems(asterix.Cat062, uap62, cat62items)
	if err != nil {
		t.Fatalf("encode CAT062: %v", err)
	}

	// Build CAT048 message
	cat48items := map[string]asterix.DataItem{
		"I048/010": &common.DataSourceIdentifier{SAC: 48, SIC: 2},
		"I048/140": &v132cat048.TimeOfDay{Time: 50000.0},
		"I048/161": &common.TrackNumber{Value: 2222},
	}
	b48, err := enc.EncodeItems(asterix.Cat048, uap48, cat48items)
	if err != nil {
		t.Fatalf("encode CAT048: %v", err)
	}

	// Concatenate into one stream
	stream := append(b62, b48...)

	dec := asterix.NewDecoder()
	dec.RegisterUAP(uap62)
	dec.RegisterUAP(uap48)

	blocks, err := dec.DecodeAll(stream)
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 DataBlocks, got %d", len(blocks))
	}

	// Verify CAT062 block
	if blocks[0].Category() != asterix.Cat062 {
		t.Errorf("block 0: expected CAT062, got CAT%d", blocks[0].Category())
	}
	recs62 := blocks[0].Records()
	if len(recs62) == 0 {
		t.Fatal("CAT062 block has no records")
	}
	assertItem(t, recs62[0], "I062/040", func(item asterix.DataItem) {
		tn, ok := item.(*v120cat062.TrackNumber)
		if !ok {
			t.Fatalf("I062/040: unexpected type %T", item)
		}
		assertEq(t, "TrackNumber", uint16(1111), tn.Value)
	})

	// Verify CAT048 block
	if blocks[1].Category() != asterix.Cat048 {
		t.Errorf("block 1: expected CAT048, got CAT%d", blocks[1].Category())
	}
	recs48 := blocks[1].Records()
	if len(recs48) == 0 {
		t.Fatal("CAT048 block has no records")
	}
	assertItem(t, recs48[0], "I048/161", func(item asterix.DataItem) {
		tn, ok := item.(*common.TrackNumber)
		if !ok {
			t.Fatalf("I048/161: unexpected type %T", item)
		}
		assertEq(t, "TrackNumber", uint16(2222), tn.Value)
	})

	_ = fmt.Sprintf("verified %d blocks", len(blocks)) // suppress unused import warning
}
