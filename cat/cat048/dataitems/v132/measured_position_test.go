// dataitems/cat048/v132/measured_position_test.go
package v132_test

import (
	"bytes"
	"math"
	"testing"

	v132 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v132"
)

func TestMeasuredPosition_EncodeDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     v132.MeasuredPosition
		expected  []byte
		tolerance float64
	}{
		{
			name: "Zero position",
			input: v132.MeasuredPosition{
				RHO:   0.0,
				THETA: 0.0,
			},
			expected:  []byte{0x00, 0x00, 0x00, 0x00},
			tolerance: 1.0 / 256.0, // LSB for RHO
		},
		{
			name: "Typical radar position - 50 NM, 90 degrees",
			input: v132.MeasuredPosition{
				RHO:   50.0,
				THETA: 90.0,
			},
			expected:  []byte{0x32, 0x00, 0x40, 0x00}, // 50*256=12800=0x3200, 90*(65536/360)=16384=0x4000
			tolerance: 1.0 / 256.0,
		},
		{
			name: "Maximum range - close to 256 NM",
			input: v132.MeasuredPosition{
				RHO:   255.99,
				THETA: 0.0,
			},
			expected:  []byte{0xFF, 0xFD, 0x00, 0x00}, // 255.99 * 256 = 65533.44 -> 65533 = 0xFFFD
			tolerance: 1.0 / 256.0 * 2, // Allow 2xLSB tolerance
		},
		{
			name: "Full circle - 360 degrees (wraps to 0)",
			input: v132.MeasuredPosition{
				RHO:   100.0,
				THETA: 360.0,
			},
			expected:  []byte{0x64, 0x00, 0x00, 0x00}, // 100*256=25600=0x6400, 360 wraps to 0
			tolerance: 1.0 / 256.0,
		},
		{
			name: "North direction - 0 degrees",
			input: v132.MeasuredPosition{
				RHO:   150.5,
				THETA: 0.0,
			},
			expected:  []byte{0x96, 0x80, 0x00, 0x00}, // 150.5*256=38528=0x9680
			tolerance: 1.0 / 256.0,
		},
		{
			name: "South direction - 180 degrees",
			input: v132.MeasuredPosition{
				RHO:   75.25,
				THETA: 180.0,
			},
			expected:  []byte{0x4B, 0x40, 0x80, 0x00}, // 75.25*256=19264=0x4B40, 180*(65536/360)=32768=0x8000
			tolerance: 1.0 / 256.0,
		},
		{
			name: "East direction - 90 degrees",
			input: v132.MeasuredPosition{
				RHO:   200.0,
				THETA: 90.0,
			},
			expected:  []byte{0xC8, 0x00, 0x40, 0x00}, // 200*256=51200=0xC800
			tolerance: 1.0 / 256.0,
		},
		{
			name: "West direction - 270 degrees",
			input: v132.MeasuredPosition{
				RHO:   125.0,
				THETA: 270.0,
			},
			expected:  []byte{0x7D, 0x00, 0xC0, 0x00}, // 125*256=32000=0x7D00, 270*(65536/360)=49152=0xC000
			tolerance: 1.0 / 256.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Encode
			buf := new(bytes.Buffer)
			n, err := tt.input.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if n != 4 {
				t.Errorf("Encode() returned n = %v, want 4", n)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("Encode() = %X, want %X", buf.Bytes(), tt.expected)
			}

			// Test Decode
			decoded := &v132.MeasuredPosition{}
			encodedBuf := bytes.NewBuffer(tt.expected)
			n, err = decoded.Decode(encodedBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if n != 4 {
				t.Errorf("Decode() returned n = %v, want 4", n)
			}

			// Check RHO within tolerance
			rhoDiff := math.Abs(decoded.RHO - tt.input.RHO)
			if rhoDiff > tt.tolerance {
				t.Errorf("Decode() RHO = %v, want %v (diff: %v, tolerance: %v)",
					decoded.RHO, tt.input.RHO, rhoDiff, tt.tolerance)
			}

			// Check THETA (accounting for 360-degree wrap)
			expectedTheta := math.Mod(tt.input.THETA, 360.0)
			if expectedTheta < 0 {
				expectedTheta += 360.0
			}
			thetaDiff := math.Abs(decoded.THETA - expectedTheta)
			// Also check wrap-around case
			if thetaDiff > 180 {
				thetaDiff = 360 - thetaDiff
			}
			if thetaDiff > 360.0/65536.0*2 { // 2x LSB tolerance
				t.Errorf("Decode() THETA = %v, want %v (diff: %v)",
					decoded.THETA, expectedTheta, thetaDiff)
			}
		})
	}
}

func TestMeasuredPosition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   v132.MeasuredPosition
		wantErr bool
	}{
		{
			name: "Valid position",
			input: v132.MeasuredPosition{
				RHO:   100.0,
				THETA: 45.0,
			},
			wantErr: false,
		},
		{
			name: "Valid at maximum range",
			input: v132.MeasuredPosition{
				RHO:   255.99,
				THETA: 0.0,
			},
			wantErr: false,
		},
		{
			name: "Invalid negative range",
			input: v132.MeasuredPosition{
				RHO:   -10.0,
				THETA: 0.0,
			},
			wantErr: true,
		},
		{
			name: "Invalid range >= 256 NM",
			input: v132.MeasuredPosition{
				RHO:   256.0,
				THETA: 0.0,
			},
			wantErr: true,
		},
		{
			name: "Valid with any azimuth (wraps)",
			input: v132.MeasuredPosition{
				RHO:   50.0,
				THETA: 450.0, // Will wrap to 90
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMeasuredPosition_String(t *testing.T) {
	tests := []struct {
		name     string
		input    v132.MeasuredPosition
		expected string
	}{
		{
			name: "Zero position",
			input: v132.MeasuredPosition{
				RHO:   0.0,
				THETA: 0.0,
			},
			expected: "RHO: 0.000 NM, THETA: 0.000°",
		},
		{
			name: "Typical position",
			input: v132.MeasuredPosition{
				RHO:   150.5,
				THETA: 270.25,
			},
			expected: "RHO: 150.500 NM, THETA: 270.250°",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMeasuredPosition_RoundTrip(t *testing.T) {
	// Test that encode followed by decode returns values within tolerance
	testPositions := []v132.MeasuredPosition{
		{RHO: 0.0, THETA: 0.0},
		{RHO: 50.0, THETA: 90.0},
		{RHO: 100.0, THETA: 180.0},
		{RHO: 150.5, THETA: 270.5},
		{RHO: 255.0, THETA: 359.9},
		{RHO: 10.25, THETA: 45.5},
		{RHO: 200.75, THETA: 135.25},
	}

	rhoTolerance := 1.0 / 256.0 * 2    // 2x LSB for RHO
	thetaTolerance := 360.0 / 65536.0 * 2 // 2x LSB for THETA

	for _, pos := range testPositions {
		t.Run("", func(t *testing.T) {
			original := pos

			// Encode
			buf := new(bytes.Buffer)
			_, err := original.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded := &v132.MeasuredPosition{}
			_, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Compare within tolerance
			rhoDiff := math.Abs(decoded.RHO - original.RHO)
			if rhoDiff > rhoTolerance {
				t.Errorf("Round trip RHO failed: got %v, want %v (diff: %v)",
					decoded.RHO, original.RHO, rhoDiff)
			}

			// Normalize expected theta to [0, 360)
			expectedTheta := math.Mod(original.THETA, 360.0)
			if expectedTheta < 0 {
				expectedTheta += 360.0
			}

			thetaDiff := math.Abs(decoded.THETA - expectedTheta)
			// Handle wrap-around
			if thetaDiff > 180 {
				thetaDiff = 360 - thetaDiff
			}
			if thetaDiff > thetaTolerance {
				t.Errorf("Round trip THETA failed: got %v, want %v (diff: %v)",
					decoded.THETA, expectedTheta, thetaDiff)
			}
		})
	}
}

func TestMeasuredPosition_Resolution(t *testing.T) {
	// Test that LSB resolution is correct
	rhoLSB := 1.0 / 256.0 // ~0.00390625 NM
	thetaLSB := 360.0 / 65536.0 // ~0.0055 degrees

	// Test RHO LSB
	pos := v132.MeasuredPosition{
		RHO:   rhoLSB,
		THETA: 0.0,
	}

	buf := new(bytes.Buffer)
	_, err := pos.Encode(buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded := &v132.MeasuredPosition{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if math.Abs(decoded.RHO-rhoLSB) > rhoLSB {
		t.Errorf("RHO LSB test failed: got %v, want %v", decoded.RHO, rhoLSB)
	}

	// Test THETA LSB
	pos = v132.MeasuredPosition{
		RHO:   10.0,
		THETA: thetaLSB,
	}

	buf = new(bytes.Buffer)
	_, err = pos.Encode(buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded = &v132.MeasuredPosition{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if math.Abs(decoded.THETA-thetaLSB) > thetaLSB*2 {
		t.Errorf("THETA LSB test failed: got %v, want %v", decoded.THETA, thetaLSB)
	}
}
