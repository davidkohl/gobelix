package cat062_test

import (
	"bytes"
	"testing"

	cat062 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// TestEncoderByteCounts verifies that each encoder returns the correct byte count.
// If Encode() returns N bytes, it MUST have written exactly N bytes to the buffer.
// A mismatch causes the LENGTH field to be wrong, breaking decoding.
func TestEncoderByteCounts(t *testing.T) {
	tests := []struct {
		name string
		item interface {
			Encode(*bytes.Buffer) (int, error)
		}
	}{
		{"I062/010", &common.DataSourceIdentifier{SAC: 6, SIC: 100}},
		{"I062/070", &cat062.TimeOfTrackInformation{Time: 12345.678}},
		{"I062/105", &cat062.CalculatedPositionWGS84{Latitude: 51.0, Longitude: 10.0}},
		{"I062/100", &cat062.CalculatedTrackPositionCartesian{X: 1000.0, Y: 2000.0}},
		{"I062/185", &cat062.CalculatedTrackVelocity{Vx: 100.0, Vy: 50.0}},
		{"I062/060", &cat062.TrackMode3ACode{Validated: true, Code: 1000}},
		{"I062/245", &cat062.TargetIdentification{IdentType: 0, Ident: "TEST1234"}},
		{"I062/040", &cat062.TrackNumber{Value: 12345}},
		{"I062/080", &cat062.TrackStatus{}},
		{"I062/200", &cat062.ModeOfMovement{}},
		{"I062/136", &cat062.MeasuredFlightLevel{FlightLevel: 350.0}},
		{"I062/135", &cat062.CalculatedTrackBarometricAltitude{QNH: true, Altitude: 350.0}},
		{"I062/220", &cat062.CalculatedRateOfClimbDescent{Rate: 1500.0}},
		
		// Compound items - most likely to have bugs
		{"I062/380", func() *cat062.AircraftDerivedData {
			addr := uint32(0x123456)
			com := uint8(1)
			stat := uint8(2)
			return &cat062.AircraftDerivedData{
				TargetAddress: &addr,
				CommunicationsCOM: &com,
				CommunicationsSTAT: &stat,
			}
		}()},
		
		{"I062/290", &cat062.SystemTrackUpdateAges{Data: []byte{0x80, 0x00, 0x05}}},
		{"I062/295", &cat062.TrackDataAges{Data: []byte{0x00}}},
		
		{"I062/500", func() *cat062.EstimatedAccuracies {
			apcX := uint16(100)
			apcY := uint16(120)
			vx := uint8(20)
			vy := uint8(24)
			return &cat062.EstimatedAccuracies{
				APCX: &apcX,
				APCY: &apcY,
				VelocityX: &vx,
				VelocityY: &vy,
			}
		}()},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			startLen := buf.Len()
			
			n, err := tt.item.Encode(&buf)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			
			actualBytes := buf.Len() - startLen
			
			if n != actualBytes {
				t.Errorf("❌ BYTE COUNT MISMATCH: Encode() returned %d but wrote %d bytes (diff: %+d)",
					n, actualBytes, actualBytes-n)
			} else {
				t.Logf("✓ Encode returned %d, wrote %d bytes", n, actualBytes)
			}
		})
	}
}
