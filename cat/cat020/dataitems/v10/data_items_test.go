// cat/cat020/dataitems/v10/data_items_test.go
package v10

import (
	"bytes"
	"testing"
)

func TestTargetReportDescriptor_EncodeDecode(t *testing.T) {
	original := &TargetReportDescriptor{
		SSR:  true,
		MS:   true,
		RAB:  false,
		SPI:  true,
		CHN:  false,
		GBS:  false,
		CRT:  false,
		SIM:  false,
		TST:  false,
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes, got %d", n)
	}

	// Decode
	decoded := NewTargetReportDescriptor()
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes read, got %d", n)
	}

	// Verify
	if decoded.SSR != original.SSR {
		t.Errorf("SSR mismatch: got %v, want %v", decoded.SSR, original.SSR)
	}
	if decoded.MS != original.MS {
		t.Errorf("MS mismatch: got %v, want %v", decoded.MS, original.MS)
	}
	if decoded.SPI != original.SPI {
		t.Errorf("SPI mismatch: got %v, want %v", decoded.SPI, original.SPI)
	}
}

func TestPositionWGS84_EncodeDecode(t *testing.T) {
	original := &PositionWGS84{
		Latitude:  51.507351,
		Longitude: -0.127758,
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != 8 {
		t.Errorf("Expected 8 bytes, got %d", n)
	}

	// Decode
	decoded := NewPositionWGS84()
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != 8 {
		t.Errorf("Expected 8 bytes read, got %d", n)
	}

	// Verify (allow small tolerance due to LSB precision)
	tolerance := 0.00001
	if abs(decoded.Latitude-original.Latitude) > tolerance {
		t.Errorf("Latitude mismatch: got %.6f, want %.6f", decoded.Latitude, original.Latitude)
	}
	if abs(decoded.Longitude-original.Longitude) > tolerance {
		t.Errorf("Longitude mismatch: got %.6f, want %.6f", decoded.Longitude, original.Longitude)
	}
}

func TestPositionCartesian_EncodeDecode(t *testing.T) {
	original := &PositionCartesian{
		X: 12345.5,
		Y: -6789.0,
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != 6 {
		t.Errorf("Expected 6 bytes, got %d", n)
	}

	// Decode
	decoded := NewPositionCartesian()
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != 6 {
		t.Errorf("Expected 6 bytes read, got %d", n)
	}

	// Verify (LSB = 0.5 m)
	if decoded.X != original.X {
		t.Errorf("X mismatch: got %.1f, want %.1f", decoded.X, original.X)
	}
	if decoded.Y != original.Y {
		t.Errorf("Y mismatch: got %.1f, want %.1f", decoded.Y, original.Y)
	}
}

func TestTrackNumber_EncodeDecode(t *testing.T) {
	original := &TrackNumber{
		TrackNumber: 1234,
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes, got %d", n)
	}

	// Decode
	decoded := NewTrackNumber()
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes read, got %d", n)
	}

	// Verify
	if decoded.TrackNumber != original.TrackNumber {
		t.Errorf("Track number mismatch: got %d, want %d", decoded.TrackNumber, original.TrackNumber)
	}
}

func TestMode3ACode_EncodeDecode(t *testing.T) {
	original := &Mode3ACode{
		V:    false,
		G:    false,
		L:    false,
		Code: 01234, // Octal representation
	}

	// Encode
	buf := new(bytes.Buffer)
	n, err := original.Encode(buf)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes, got %d", n)
	}

	// Decode
	decoded := NewMode3ACode()
	n, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected 2 bytes read, got %d", n)
	}

	// Verify
	if decoded.Code != original.Code {
		t.Errorf("Mode3A mismatch: got %04o, want %04o", decoded.Code, original.Code)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
