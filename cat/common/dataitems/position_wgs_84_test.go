// dataitems/common/position_wgs_84_test.go
package common_test

import (
	"bytes"
	"math"
	"testing"

	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestPosition_EncodeDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     common.Position
		wantErr   bool
		tolerance float64 // Acceptable difference due to encoding resolution
	}{
		{
			name: "Valid position - London",
			input: common.Position{
				Latitude:  51.5074,
				Longitude: -0.1278,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - New York",
			input: common.Position{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - Sydney",
			input: common.Position{
				Latitude:  -33.8688,
				Longitude: 151.2093,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - North Pole",
			input: common.Position{
				Latitude:  90.0,
				Longitude: 0.0,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - South Pole",
			input: common.Position{
				Latitude:  -90.0,
				Longitude: 0.0,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - Equator Prime Meridian",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 0.0,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - near International Date Line",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 179.9,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Valid position - negative longitude limit",
			input: common.Position{
				Latitude:  0.0,
				Longitude: -180.0,
			},
			wantErr:   false,
			tolerance: common.ResolutionWGS84 * 2,
		},
		{
			name: "Invalid position - latitude too high",
			input: common.Position{
				Latitude:  91.0,
				Longitude: 0.0,
			},
			wantErr:   true,
			tolerance: 0,
		},
		{
			name: "Invalid position - latitude too low",
			input: common.Position{
				Latitude:  -91.0,
				Longitude: 0.0,
			},
			wantErr:   true,
			tolerance: 0,
		},
		{
			name: "Invalid position - longitude too high",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 181.0,
			},
			wantErr:   true,
			tolerance: 0,
		},
		{
			name: "Invalid position - longitude too low",
			input: common.Position{
				Latitude:  0.0,
				Longitude: -181.0,
			},
			wantErr:   true,
			tolerance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Encode
			buf := new(bytes.Buffer)
			n, err := tt.input.Encode(buf)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if n != 6 {
					t.Errorf("Encode() returned n = %v, want 6", n)
				}

				// Test Decode
				decodedItem := &common.Position{}
				encodedBuf := bytes.NewBuffer(buf.Bytes())
				n, err = decodedItem.Decode(encodedBuf)
				if err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if n != 6 {
					t.Errorf("Decode() returned n = %v, want 6", n)
				}

				// Check latitude within tolerance
				latDiff := math.Abs(decodedItem.Latitude - tt.input.Latitude)
				if latDiff > tt.tolerance {
					t.Errorf("Decode() Latitude = %v, want %v (diff: %v, tolerance: %v)",
						decodedItem.Latitude, tt.input.Latitude, latDiff, tt.tolerance)
				}

				// Check longitude within tolerance
				lonDiff := math.Abs(decodedItem.Longitude - tt.input.Longitude)
				if lonDiff > tt.tolerance {
					t.Errorf("Decode() Longitude = %v, want %v (diff: %v, tolerance: %v)",
						decodedItem.Longitude, tt.input.Longitude, lonDiff, tt.tolerance)
				}
			}
		})
	}
}

func TestPosition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   common.Position
		wantErr bool
	}{
		{
			name: "Valid position",
			input: common.Position{
				Latitude:  51.5074,
				Longitude: -0.1278,
			},
			wantErr: false,
		},
		{
			name: "Valid at latitude limits",
			input: common.Position{
				Latitude:  90.0,
				Longitude: 0.0,
			},
			wantErr: false,
		},
		{
			name: "Valid at negative latitude limit",
			input: common.Position{
				Latitude:  -90.0,
				Longitude: 0.0,
			},
			wantErr: false,
		},
		{
			name: "Valid at longitude limits",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 180.0,
			},
			wantErr: false,
		},
		{
			name: "Valid at negative longitude limit",
			input: common.Position{
				Latitude:  0.0,
				Longitude: -180.0,
			},
			wantErr: false,
		},
		{
			name: "Invalid latitude > 90",
			input: common.Position{
				Latitude:  90.1,
				Longitude: 0.0,
			},
			wantErr: true,
		},
		{
			name: "Invalid latitude < -90",
			input: common.Position{
				Latitude:  -90.1,
				Longitude: 0.0,
			},
			wantErr: true,
		},
		{
			name: "Invalid longitude > 180",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 180.1,
			},
			wantErr: true,
		},
		{
			name: "Invalid longitude < -180",
			input: common.Position{
				Latitude:  0.0,
				Longitude: -180.1,
			},
			wantErr: true,
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

func TestPosition_String(t *testing.T) {
	tests := []struct {
		name     string
		input    common.Position
		expected string
	}{
		{
			name: "London",
			input: common.Position{
				Latitude:  51.5074,
				Longitude: -0.1278,
			},
			expected: "51.507400°N -0.127800°E",
		},
		{
			name: "Equator Prime Meridian",
			input: common.Position{
				Latitude:  0.0,
				Longitude: 0.0,
			},
			expected: "0.000000°N 0.000000°E",
		},
		{
			name: "Sydney (Southern Hemisphere)",
			input: common.Position{
				Latitude:  -33.8688,
				Longitude: 151.2093,
			},
			expected: "-33.868800°N 151.209300°E",
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

func TestPosition_DecodeIncomplete(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Incomplete - 0 bytes",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "Incomplete - 1 byte",
			input:   []byte{0x00},
			wantErr: true,
		},
		{
			name:    "Incomplete - 3 bytes",
			input:   []byte{0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "Incomplete - 5 bytes",
			input:   []byte{0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "Complete - 6 bytes",
			input:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := &common.Position{}
			buf := bytes.NewBuffer(tt.input)
			_, err := pos.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPosition_RoundTrip(t *testing.T) {
	// Test that encode followed by decode returns values within tolerance
	testPositions := []common.Position{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 51.5074, Longitude: -0.1278},
		{Latitude: 40.7128, Longitude: -74.0060},
		{Latitude: -33.8688, Longitude: 151.2093},
		{Latitude: 90.0, Longitude: 0.0},
		{Latitude: -90.0, Longitude: 0.0},
		{Latitude: 0.0, Longitude: 179.9},
		{Latitude: 0.0, Longitude: -179.9},
	}

	tolerance := common.ResolutionWGS84 * 2

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
			decoded := &common.Position{}
			_, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Compare within tolerance
			latDiff := math.Abs(decoded.Latitude - original.Latitude)
			lonDiff := math.Abs(decoded.Longitude - original.Longitude)

			if latDiff > tolerance {
				t.Errorf("Round trip latitude failed: got %v, want %v (diff: %v)",
					decoded.Latitude, original.Latitude, latDiff)
			}
			if lonDiff > tolerance {
				t.Errorf("Round trip longitude failed: got %v, want %v (diff: %v)",
					decoded.Longitude, original.Longitude, lonDiff)
			}
		})
	}
}

func TestPosition_EncodingPrecision(t *testing.T) {
	// Test that the encoding resolution is correct
	pos := common.Position{
		Latitude:  common.ResolutionWGS84,
		Longitude: common.ResolutionWGS84,
	}

	buf := new(bytes.Buffer)
	_, err := pos.Encode(buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded := &common.Position{}
	_, err = decoded.Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// The smallest encodable difference should be within 2x resolution
	tolerance := common.ResolutionWGS84 * 2
	latDiff := math.Abs(decoded.Latitude - pos.Latitude)
	lonDiff := math.Abs(decoded.Longitude - pos.Longitude)

	if latDiff > tolerance {
		t.Errorf("Precision test failed for latitude: diff %v exceeds tolerance %v", latDiff, tolerance)
	}
	if lonDiff > tolerance {
		t.Errorf("Precision test failed for longitude: diff %v exceeds tolerance %v", lonDiff, tolerance)
	}
}
