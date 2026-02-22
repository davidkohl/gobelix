// dataitems/cat062/v120/calculated_rate_of_climb_descent_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestCalculatedRateOfClimbDescent_RoundTrip tests encode/decode roundtrip
func TestCalculatedRateOfClimbDescent_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		rate    float64
		encoded []byte
	}{
		{
			name:    "Zero rate (level flight)",
			rate:    0,
			encoded: []byte{0x00, 0x00},
		},
		{
			name:    "Positive climb +1250 ft/min",
			rate:    1250,
			encoded: []byte{0x00, 0xC8}, // 1250/6.25 = 200 = 0x00C8
		},
		{
			name:    "Positive climb +6.25 ft/min (1 LSB)",
			rate:    6.25,
			encoded: []byte{0x00, 0x01},
		},
		{
			name:    "Negative descent -1250 ft/min",
			rate:    -1250,
			encoded: []byte{0xFF, 0x38}, // -200 in two's complement = 0xFF38
		},
		{
			name:    "Negative descent -6.25 ft/min (1 LSB)",
			rate:    -6.25,
			encoded: []byte{0xFF, 0xFF}, // -1 in two's complement
		},
		{
			name:    "Max positive climb +32000 ft/min",
			rate:    32000,
			encoded: []byte{0x14, 0x00}, // 32000/6.25 = 5120 = 0x1400
		},
		{
			name:    "Max negative descent -32000 ft/min",
			rate:    -32000,
			encoded: []byte{0xEC, 0x00}, // -5120 in two's complement = 0xEC00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			original := &v120.CalculatedRateOfClimbDescent{Rate: tt.rate}
			encodeBuf := new(bytes.Buffer)
			bytesWritten, err := original.Encode(encodeBuf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if bytesWritten != 2 {
				t.Errorf("Encode() wrote %d bytes, want 2", bytesWritten)
			}
			if !bytes.Equal(encodeBuf.Bytes(), tt.encoded) {
				t.Errorf("Encode() = %X, want %X", encodeBuf.Bytes(), tt.encoded)
			}

			// Decode
			decoded := &v120.CalculatedRateOfClimbDescent{}
			decodeBuf := bytes.NewBuffer(tt.encoded)
			bytesRead, err := decoded.Decode(decodeBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if bytesRead != 2 {
				t.Errorf("Decode() read %d bytes, want 2", bytesRead)
			}

			// Compare (allow small rounding error due to LSB quantization)
			if diff := decoded.Rate - tt.rate; diff < -3.125 || diff > 3.125 {
				t.Errorf("Decode() rate = %f, want %f (diff = %f)", decoded.Rate, tt.rate, diff)
			}
		})
	}
}

// TestCalculatedRateOfClimbDescent_DecodeTruncated tests handling of incomplete data
func TestCalculatedRateOfClimbDescent_DecodeTruncated(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Empty buffer",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "Only 1 byte (incomplete)",
			input:   []byte{0x00},
			wantErr: true,
		},
		{
			name:    "Complete 2 bytes",
			input:   []byte{0x00, 0xC8},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := &v120.CalculatedRateOfClimbDescent{}
			buf := bytes.NewBuffer(tt.input)
			_, err := rate.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCalculatedRateOfClimbDescent_Validate tests validation
func TestCalculatedRateOfClimbDescent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rate    float64
		wantErr bool
	}{
		{
			name:    "Valid zero",
			rate:    0,
			wantErr: false,
		},
		{
			name:    "Valid positive",
			rate:    1250,
			wantErr: false,
		},
		{
			name:    "Valid negative",
			rate:    -1250,
			wantErr: false,
		},
		{
			name:    "Valid max positive",
			rate:    32000,
			wantErr: false,
		},
		{
			name:    "Valid max negative",
			rate:    -32000,
			wantErr: false,
		},
		{
			name:    "Invalid too positive",
			rate:    40000,
			wantErr: true,
		},
		{
			name:    "Invalid too negative",
			rate:    -40000,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := &v120.CalculatedRateOfClimbDescent{Rate: tt.rate}
			err := rate.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
