// dataitems/cat062/v120/estimated_accuracies_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestEstimatedAccuracies_RoundTrip tests encode/decode roundtrip
func TestEstimatedAccuracies_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		item    *v120.EstimatedAccuracies
		encoded []byte
	}{
		{
			name: "Empty (no subfields)",
			item: &v120.EstimatedAccuracies{},
			encoded: []byte{0x00}, // FSPEC with no bits set
		},
		{
			name: "APC only (bit 8 set)",
			item: &v120.EstimatedAccuracies{
				APCX: uint16Ptr(100), // 100 * 0.5m = 50m
				APCY: uint16Ptr(200), // 200 * 0.5m = 100m
			},
			encoded: []byte{0x80, 0x00, 0x64, 0x00, 0xC8}, // FSPEC=0x80, X=100, Y=200
		},
		{
			name: "COV only (bit 7 set)",
			item: &v120.EstimatedAccuracies{
				XYCovariance: int16Ptr(150),
			},
			encoded: []byte{0x40, 0x00, 0x96}, // FSPEC=0x40, COV=150
		},
		{
			name: "APW only (bit 6 set)",
			item: &v120.EstimatedAccuracies{
				APWLatitude:  uint16Ptr(50),
				APWLongitude: uint16Ptr(75),
			},
			encoded: []byte{0x20, 0x00, 0x32, 0x00, 0x4B}, // FSPEC=0x20, LAT=50, LON=75
		},
		{
			name: "AGA only (bit 5 set)",
			item: &v120.EstimatedAccuracies{
				GeometricAltitudeAccuracy: uint8Ptr(10),
			},
			encoded: []byte{0x10, 0x0A}, // FSPEC=0x10, AGA=10
		},
		{
			name: "ABA only (bit 4 set)",
			item: &v120.EstimatedAccuracies{
				BarometricAltitudeAccuracy: uint8Ptr(20),
			},
			encoded: []byte{0x08, 0x14}, // FSPEC=0x08, ABA=20
		},
		{
			name: "ATV only (bit 3 set)",
			item: &v120.EstimatedAccuracies{
				VelocityX: uint8Ptr(5),
				VelocityY: uint8Ptr(8),
			},
			encoded: []byte{0x04, 0x05, 0x08}, // FSPEC=0x04, Vx=5, Vy=8
		},
		{
			name: "AA only (bit 2 set)",
			item: &v120.EstimatedAccuracies{
				AccelerationX: uint8Ptr(2),
				AccelerationY: uint8Ptr(3),
			},
			encoded: []byte{0x02, 0x02, 0x03}, // FSPEC=0x02, Ax=2, Ay=3
		},
		{
			name: "ARC only (bit 8 of second byte, requires FX)",
			item: &v120.EstimatedAccuracies{
				RateOfClimbDescentAccuracy: uint8Ptr(15),
			},
			encoded: []byte{0x01, 0x80, 0x0F}, // FSPEC byte 1=0x01 (FX set), byte 2=0x80, ARC=15
		},
		{
			name: "Multiple subfields (APC + ATV)",
			item: &v120.EstimatedAccuracies{
				APCX:      uint16Ptr(100),
				APCY:      uint16Ptr(200),
				VelocityX: uint8Ptr(5),
				VelocityY: uint8Ptr(8),
			},
			encoded: []byte{0x84, 0x00, 0x64, 0x00, 0xC8, 0x05, 0x08}, // FSPEC=0x84 (bits 8+3)
		},
		{
			name: "All first-byte subfields",
			item: &v120.EstimatedAccuracies{
				APCX:                       uint16Ptr(100),
				APCY:                       uint16Ptr(200),
				XYCovariance:               int16Ptr(150),
				APWLatitude:                uint16Ptr(50),
				APWLongitude:               uint16Ptr(75),
				GeometricAltitudeAccuracy:  uint8Ptr(10),
				BarometricAltitudeAccuracy: uint8Ptr(20),
				VelocityX:                  uint8Ptr(5),
				VelocityY:                  uint8Ptr(8),
				AccelerationX:              uint8Ptr(2),
				AccelerationY:              uint8Ptr(3),
			},
			// FSPEC=0xFE (bits 8-2 all set)
			// APC(4) + COV(2) + APW(4) + AGA(1) + ABA(1) + ATV(2) + AA(2) = 16 bytes
			encoded: []byte{0xFE, 0x00, 0x64, 0x00, 0xC8, 0x00, 0x96, 0x00, 0x32, 0x00, 0x4B, 0x0A, 0x14, 0x05, 0x08, 0x02, 0x03},
		},
		{
			name: "First byte + extension (APC + ARC)",
			item: &v120.EstimatedAccuracies{
				APCX:                       uint16Ptr(100),
				APCY:                       uint16Ptr(200),
				RateOfClimbDescentAccuracy: uint8Ptr(15),
			},
			encoded: []byte{0x81, 0x80, 0x00, 0x64, 0x00, 0xC8, 0x0F}, // FSPEC=0x81,0x80, APC(4), ARC(1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encodeBuf := new(bytes.Buffer)
			bytesWritten, err := tt.item.Encode(encodeBuf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if bytesWritten != len(tt.encoded) {
				t.Errorf("Encode() wrote %d bytes, want %d", bytesWritten, len(tt.encoded))
			}
			if !bytes.Equal(encodeBuf.Bytes(), tt.encoded) {
				t.Errorf("Encode() = %X, want %X", encodeBuf.Bytes(), tt.encoded)
			}

			// Decode
			decoded := &v120.EstimatedAccuracies{}
			decodeBuf := bytes.NewBuffer(tt.encoded)
			bytesRead, err := decoded.Decode(decodeBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v, encoded = %X", err, tt.encoded)
			}
			if bytesRead != len(tt.encoded) {
				t.Errorf("Decode() read %d bytes, want %d", bytesRead, len(tt.encoded))
			}

			// Re-encode to verify roundtrip
			reEncodeBuf := new(bytes.Buffer)
			_, err = decoded.Encode(reEncodeBuf)
			if err != nil {
				t.Fatalf("Re-encode failed: %v", err)
			}
			if !bytes.Equal(reEncodeBuf.Bytes(), tt.encoded) {
				t.Errorf("Round trip failed:\n  original: %X\n  re-encoded: %X", tt.encoded, reEncodeBuf.Bytes())
			}
		})
	}
}

// TestEstimatedAccuracies_DecodeTruncated tests handling of incomplete data
func TestEstimatedAccuracies_DecodeTruncated(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty buffer",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "APC bit set but no data",
			input:   []byte{0x80}, // Bit 8 set (APC) but no 4-byte data
			wantErr: true,
			errMsg:  "APC",
		},
		{
			name:    "APC incomplete (only 2 bytes)",
			input:   []byte{0x80, 0x00, 0x64}, // APC needs 4 bytes, only 2 provided
			wantErr: true,
			errMsg:  "APC",
		},
		{
			name:    "ATV bit set but no data",
			input:   []byte{0x04}, // Bit 3 set (ATV) but no 2-byte data
			wantErr: true,
			errMsg:  "ATV",
		},
		{
			name:    "ATV incomplete (only 1 byte)",
			input:   []byte{0x04, 0x05}, // ATV needs 2 bytes, only 1 provided
			wantErr: true,
			errMsg:  "ATV",
		},
		{
			name:    "FX bit set but no extension byte",
			input:   []byte{0x01}, // FX set but no second byte
			wantErr: true,
		},
		{
			name:    "Second FX bit set (invalid)",
			input:   []byte{0x01, 0x01}, // Second FX bit should not be set
			wantErr: true,
			errMsg:  "unexpected second extension bit",
		},
		{
			name:    "Valid empty FSPEC",
			input:   []byte{0x00},
			wantErr: false,
		},
		{
			name:    "Valid APC",
			input:   []byte{0x80, 0x00, 0x64, 0x00, 0xC8},
			wantErr: false,
		},
		{
			name:    "Valid extension",
			input:   []byte{0x01, 0x80, 0x0F},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &v120.EstimatedAccuracies{}
			buf := bytes.NewBuffer(tt.input)
			_, err := item.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v, input = %X", err, tt.wantErr, tt.input)
			}
			if err != nil && tt.errMsg != "" {
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errMsg)) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
			}
		})
	}
}

// Helper functions
func uint8Ptr(v uint8) *uint8   { return &v }
func uint16Ptr(v uint16) *uint16 { return &v }
func int16Ptr(v int16) *int16   { return &v }
