// dataitems/cat062/v120/system_track_update_ages_encoder_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestSystemTrackUpdateAges_EncoderBug tests the potential encoder bug
// where the Encode() method just writes s.Data without validation
func TestSystemTrackUpdateAges_EncoderBug(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantDecodeErr bool
		description   string
	}{
		{
			name:          "FSPEC indicates 1 subfield but none provided",
			data:          []byte{0x80}, // Bit 8 set, but no subfield data
			wantDecodeErr: true,
			description:   "Encoder writes incomplete data - FSPEC says data exists but it doesn't",
		},
		{
			name:          "FSPEC indicates extension but incomplete",
			data:          []byte{0x01, 0x80}, // FX bit + bit 8 in extension, but no data
			wantDecodeErr: true,
			description:   "Extension FSPEC promises data that isn't there",
		},
		{
			name:          "2-octet subfield truncated",
			data:          []byte{0x10, 0x00}, // Bit 12 set (2-octet), only 1 byte provided
			wantDecodeErr: true,
			description:   "2-octet subfield only has 1 byte",
		},
		{
			name:          "Valid data",
			data:          []byte{0x80, 0x05},
			wantDecodeErr: false,
			description:   "Properly formed data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an instance with raw data
			ages := &v120.SystemTrackUpdateAges{Data: tt.data}

			// Encode it
			encodeBuf := new(bytes.Buffer)
			_, err := ages.Encode(encodeBuf)
			if err != nil {
				t.Fatalf("Encode() unexpected error = %v", err)
			}

			// Try to decode what we just encoded
			decodedAges := &v120.SystemTrackUpdateAges{}
			decodeBuf := bytes.NewBuffer(encodeBuf.Bytes())
			_, err = decodedAges.Decode(decodeBuf)

			if tt.wantDecodeErr {
				if err == nil {
					t.Errorf("%s: Decode() should have failed for invalid data %X, but succeeded",
						tt.description, tt.data)
				} else {
					t.Logf("Expected decode error (invalid data): %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Decode() unexpected error = %v, data = %X",
						tt.description, err, tt.data)
				}
			}
		})
	}
}

// TestSystemTrackUpdateAges_SetDataCorrectly tests that Data should be set properly
// This documents how Data SHOULD be structured for proper encoding
func TestSystemTrackUpdateAges_SetDataCorrectly(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		canRoundtrip bool
	}{
		{
			name:        "Empty FSPEC",
			data:        []byte{0x00},
			canRoundtrip: true,
		},
		{
			name:        "Single subfield properly formed",
			data:        []byte{0x80, 0x05},
			canRoundtrip: true,
		},
		{
			name:        "FSPEC with no data (INVALID)",
			data:        []byte{0x80}, // Claims subfield exists but no data
			canRoundtrip: false,
		},
		{
			name:        "Extension with no data (INVALID)",
			data:        []byte{0x01, 0x80}, // Extension claims subfield but no data
			canRoundtrip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ages := &v120.SystemTrackUpdateAges{Data: tt.data}
			encodeBuf := new(bytes.Buffer)
			bytesWritten, err := ages.Encode(encodeBuf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if bytesWritten != len(tt.data) {
				t.Errorf("Encode() wrote %d bytes, expected %d", bytesWritten, len(tt.data))
			}

			// Decode
			decoded := &v120.SystemTrackUpdateAges{}
			decodeBuf := bytes.NewBuffer(encodeBuf.Bytes())
			bytesRead, err := decoded.Decode(decodeBuf)

			if tt.canRoundtrip {
				if err != nil {
					t.Errorf("Decode() error = %v (expected success for valid data)", err)
				}
				if bytesRead != bytesWritten {
					t.Errorf("Decode() read %d bytes, Encode() wrote %d bytes", bytesRead, bytesWritten)
				}
				if !bytes.Equal(decoded.Data, tt.data) {
					t.Errorf("Round trip data mismatch: got %X, want %X", decoded.Data, tt.data)
				}
			} else {
				if err == nil {
					t.Errorf("Decode() should fail for invalid data %X, but succeeded", tt.data)
				} else {
					t.Logf("Decode() correctly rejected invalid data: %v", err)
				}
			}
		})
	}
}
