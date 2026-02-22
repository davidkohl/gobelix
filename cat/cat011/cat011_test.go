package cat011_test

import (
	"bytes"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat011"
	v13 "github.com/davidkohl/gobelix/cat/cat011/dataitems/v13"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestNewUAP(t *testing.T) {
	uap, err := cat011.NewUAP(cat011.Version13)
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	if uap.Category() != asterix.Cat011 {
		t.Errorf("expected category %d, got %d", asterix.Cat011, uap.Category())
	}

	if uap.Version() != "1.3" {
		t.Errorf("expected version 1.3, got %s", uap.Version())
	}
}

func TestLatestVersion(t *testing.T) {
	if cat011.LatestVersion() != cat011.Version13 {
		t.Errorf("expected latest version %s, got %s", cat011.Version13, cat011.LatestVersion())
	}
}

func TestAvailableVersions(t *testing.T) {
	versions := cat011.AvailableVersions()
	if len(versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(versions))
	}
	if versions[0] != cat011.Version13 {
		t.Errorf("expected version %s, got %s", cat011.Version13, versions[0])
	}
}

func TestMessageType(t *testing.T) {
	tests := []struct {
		name    string
		msgType uint8
		want    string
	}{
		{"TargetReport", 1, "Target Report"},
		{"ManualAttachment", 2, "Manual Attachment"},
		{"HoldbarStatus", 7, "Holdbar Status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &v13.MessageType{Type: tt.msgType}

			// Test encode
			buf := &bytes.Buffer{}
			n, err := msg.Encode(buf)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}
			if n != 1 {
				t.Errorf("expected 1 byte, wrote %d", n)
			}

			// Test decode
			decoded := &v13.MessageType{}
			n, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if decoded.Type != tt.msgType {
				t.Errorf("expected type %d, got %d", tt.msgType, decoded.Type)
			}
		})
	}
}

func TestPositionWGS84(t *testing.T) {
	pos := &v13.PositionWGS84{
		Latitude:  52.123456,
		Longitude: 4.567890,
	}

	// Test encode
	buf := &bytes.Buffer{}
	n, err := pos.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 8 {
		t.Errorf("expected 8 bytes, wrote %d", n)
	}

	// Test decode
	decoded := &v13.PositionWGS84{}
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if n != 8 {
		t.Errorf("expected 8 bytes, read %d", n)
	}

	// Check values are close (LSB precision)
	const tolerance = 0.0001
	if diff := pos.Latitude - decoded.Latitude; diff > tolerance || diff < -tolerance {
		t.Errorf("latitude mismatch: expected %f, got %f", pos.Latitude, decoded.Latitude)
	}
	if diff := pos.Longitude - decoded.Longitude; diff > tolerance || diff < -tolerance {
		t.Errorf("longitude mismatch: expected %f, got %f", pos.Longitude, decoded.Longitude)
	}
}

func TestPositionCartesian(t *testing.T) {
	pos := &v13.PositionCartesian{
		X: 1234,
		Y: -5678,
	}

	buf := &bytes.Buffer{}
	n, err := pos.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, wrote %d", n)
	}

	decoded := &v13.PositionCartesian{}
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded.X != pos.X || decoded.Y != pos.Y {
		t.Errorf("mismatch: expected (%f,%f), got (%f,%f)", pos.X, pos.Y, decoded.X, decoded.Y)
	}
}

func TestTrackVelocity(t *testing.T) {
	vel := &v13.TrackVelocityCartesian{
		Vx: 123.5,
		Vy: -45.25,
	}

	buf := &bytes.Buffer{}
	n, err := vel.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, wrote %d", n)
	}

	decoded := &v13.TrackVelocityCartesian{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// LSB = 0.25 m/s
	if decoded.Vx != vel.Vx || decoded.Vy != vel.Vy {
		t.Errorf("mismatch: expected (%f,%f), got (%f,%f)", vel.Vx, vel.Vy, decoded.Vx, decoded.Vy)
	}
}

func TestTrackStatus(t *testing.T) {
	status := &v13.TrackStatus{
		MON: true,
		GBS: true,
		MRH: false,
		SRC: v13.HeightSourceGPS,
		CNF: false,
		SIM: false,
		TSE: true, // This requires first extension
		FPC: true, // This requires second extension
	}

	buf := &bytes.Buffer{}
	n, err := status.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n < 1 {
		t.Errorf("expected at least 1 byte, wrote %d", n)
	}

	decoded := &v13.TrackStatus{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.MON != status.MON {
		t.Errorf("MON mismatch: expected %v, got %v", status.MON, decoded.MON)
	}
	if decoded.TSE != status.TSE {
		t.Errorf("TSE mismatch: expected %v, got %v", status.TSE, decoded.TSE)
	}
	if decoded.FPC != status.FPC {
		t.Errorf("FPC mismatch: expected %v, got %v", status.FPC, decoded.FPC)
	}
}

func TestMode3ACode(t *testing.T) {
	code := &v13.Mode3ACode{Code: 0x1234 & 0x0FFF} // 1234 octal encoded

	buf := &bytes.Buffer{}
	n, err := code.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes, wrote %d", n)
	}

	decoded := &v13.Mode3ACode{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Code != code.Code {
		t.Errorf("code mismatch: expected %04X, got %04X", code.Code, decoded.Code)
	}
}

func TestFlightLevel(t *testing.T) {
	fl := &v13.MeasuredFlightLevel{FlightLevel: 350.5}

	buf := &bytes.Buffer{}
	n, err := fl.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes, wrote %d", n)
	}

	decoded := &v13.MeasuredFlightLevel{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.FlightLevel != fl.FlightLevel {
		t.Errorf("FL mismatch: expected %f, got %f", fl.FlightLevel, decoded.FlightLevel)
	}
}

func TestTargetIdentification(t *testing.T) {
	tid := &v13.TargetIdentification{
		STI:      v13.STIDownlinked,
		Callsign: "DLH1234",
	}

	buf := &bytes.Buffer{}
	n, err := tid.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 7 {
		t.Errorf("expected 7 bytes, wrote %d", n)
	}

	decoded := &v13.TargetIdentification{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.STI != tid.STI {
		t.Errorf("STI mismatch: expected %d, got %d", tid.STI, decoded.STI)
	}
	if decoded.Callsign != tid.Callsign {
		t.Errorf("callsign mismatch: expected %s, got %s", tid.Callsign, decoded.Callsign)
	}
}

func TestVehicleFleetIdentification(t *testing.T) {
	vfi := &v13.VehicleFleetIdentification{VFI: v13.VFIFire}

	buf := &bytes.Buffer{}
	n, err := vfi.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 byte, wrote %d", n)
	}

	decoded := &v13.VehicleFleetIdentification{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.VFI != vfi.VFI {
		t.Errorf("VFI mismatch: expected %d, got %d", vfi.VFI, decoded.VFI)
	}
}

func TestDataSourceIdentifier(t *testing.T) {
	dsi := &common.DataSourceIdentifier{
		SAC: 25,
		SIC: 100,
	}

	buf := &bytes.Buffer{}
	n, err := dsi.Encode(buf)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 bytes, wrote %d", n)
	}

	decoded := &common.DataSourceIdentifier{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.SAC != dsi.SAC || decoded.SIC != dsi.SIC {
		t.Errorf("mismatch: expected (%d,%d), got (%d,%d)", dsi.SAC, dsi.SIC, decoded.SAC, decoded.SIC)
	}
}

func TestCreateDataItem(t *testing.T) {
	uap, err := cat011.NewUAP(cat011.Version13)
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	tests := []string{
		"I011/010", "I011/000", "I011/015", "I011/140", "I011/041",
		"I011/042", "I011/202", "I011/210", "I011/060", "I011/245",
		"I011/380", "I011/161", "I011/170", "I011/290", "I011/430",
		"I011/090", "I011/093", "I011/092", "I011/215", "I011/270",
		"I011/390", "I011/300", "I011/310", "I011/500", "I011/600",
		"I011/605", "I011/610", "SP011", "RE011",
	}

	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			item, err := uap.CreateDataItem(id)
			if err != nil {
				t.Errorf("failed to create data item %s: %v", id, err)
			}
			if item == nil {
				t.Errorf("created nil data item for %s", id)
			}
		})
	}
}

func TestUnknownDataItem(t *testing.T) {
	uap, err := cat011.NewUAP(cat011.Version13)
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	_, err = uap.CreateDataItem("I011/999")
	if err == nil {
		t.Error("expected error for unknown data item")
	}
}
