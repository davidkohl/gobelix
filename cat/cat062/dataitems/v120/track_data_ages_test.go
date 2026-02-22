// dataitems/cat062/v120/track_data_ages_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestTrackDataAges_RoundTrip tests that encoded data can be decoded back correctly
func TestTrackDataAges_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		// Raw encoded bytes - we'll use this as the baseline truth
		encoded []byte
	}{
		{
			name:    "Empty FSPEC (no subfields)",
			encoded: []byte{0x00},
		},
		{
			name:    "Single subfield - bit 8",
			encoded: []byte{0x80, 0x05}, // FSPEC=0x80, subfield=5
		},
		{
			name:    "Single subfield - bit 7",
			encoded: []byte{0x40, 0x0A}, // FSPEC=0x40, subfield=10
		},
		{
			name:    "Two subfields",
			encoded: []byte{0xC0, 0x05, 0x0A}, // FSPEC=0xC0, two subfields
		},
		{
			name:    "Multiple subfields without extension",
			encoded: []byte{0xFE, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, // Bits 8-2 set (7 subfields)
		},
		{
			name:    "Extended FSPEC - single extension octet",
			encoded: []byte{0x01, 0x80, 0x05}, // FX bit set, second octet bit 8, one subfield
		},
		{
			name:    "Extended FSPEC - multiple extension subfields",
			encoded: []byte{0x01, 0xC0, 0x05, 0x0A}, // FX + bits 8+7 in second octet
		},
		{
			name:    "Mixed - primary and extension subfields",
			encoded: []byte{0x81, 0x40, 0x05, 0x0A}, // Bit 8 in primary + bit 7 in extension
		},
		{
			name:    "Two extension octets",
			encoded: []byte{0x01, 0x01, 0x80, 0x05}, // FX in octet 1, FX in octet 2, bit 8 in octet 3
		},
		{
			name:    "Maximum extension (5 octets FSPEC)",
			encoded: []byte{0x01, 0x01, 0x01, 0x01, 0x80, 0x05}, // 5 FSPEC octets, 1 subfield
		},
		{
			name:    "Minimal extension",
			encoded: []byte{0x01, 0x00}, // FX bit set, no subfields in extension
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode the encoded bytes
			decoded := &v120.TrackDataAges{}
			decodeBuf := bytes.NewBuffer(tt.encoded)
			bytesRead, err := decoded.Decode(decodeBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v, encoded = %X", err, tt.encoded)
			}
			if bytesRead != len(tt.encoded) {
				t.Errorf("Decode() read %d bytes, want %d", bytesRead, len(tt.encoded))
			}

			// Re-encode the decoded data
			encodeBuf := new(bytes.Buffer)
			bytesWritten, err := decoded.Encode(encodeBuf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if bytesWritten != len(tt.encoded) {
				t.Errorf("Encode() wrote %d bytes, want %d", bytesWritten, len(tt.encoded))
			}

			// Compare the re-encoded bytes with original
			reencoded := encodeBuf.Bytes()
			if !bytes.Equal(reencoded, tt.encoded) {
				t.Errorf("Round trip failed:\n  original: %X\n  decoded:  %X", tt.encoded, reencoded)
				t.Logf("Decoded.Data length: %d", len(decoded.Data))
			}
		})
	}
}

// TestTrackDataAges_DecodeTruncated tests handling of truncated/incomplete data
func TestTrackDataAges_DecodeTruncated(t *testing.T) {
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
			name:    "FSPEC indicates subfield but data missing",
			input:   []byte{0x80}, // Bit 8 set but no subfield data
			wantErr: true,
		},
		{
			name:    "Extension bit set but no extension octet",
			input:   []byte{0x01}, // FX bit set but no second FSPEC octet
			wantErr: true,
		},
		{
			name:    "Extension FSPEC indicates subfield but data missing",
			input:   []byte{0x01, 0x80}, // Extension + bit 8, but no data
			wantErr: true,
		},
		{
			name:    "Multiple extension bits but incomplete",
			input:   []byte{0x01, 0x01}, // Two FX bits but missing third octet
			wantErr: true,
		},
		{
			name:    "Valid minimal FSPEC",
			input:   []byte{0x00},
			wantErr: false,
		},
		{
			name:    "Valid single subfield",
			input:   []byte{0x80, 0x05},
			wantErr: false,
		},
		{
			name:    "Valid extension",
			input:   []byte{0x01, 0x80, 0x05},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ages := &v120.TrackDataAges{}
			buf := bytes.NewBuffer(tt.input)
			_, err := ages.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v, input = %X", err, tt.wantErr, tt.input)
			}
		})
	}
}

// TestTrackDataAges_EncodeEmpty tests encoding of empty/uninitialized data
func TestTrackDataAges_EncodeEmpty(t *testing.T) {
	ages := &v120.TrackDataAges{}

	buf := new(bytes.Buffer)
	n, err := ages.Encode(buf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if n != 1 {
		t.Errorf("Encode() wrote %d bytes, want 1", n)
	}

	expected := []byte{0x00}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Encode() = %X, want %X", buf.Bytes(), expected)
	}
}

// TestTrackDataAges_MaximumFSPEC tests the maximum allowed FSPEC size (5 octets)
func TestTrackDataAges_MaximumFSPEC(t *testing.T) {
	// Maximum FSPEC is 5 octets with FX bits in first 4 octets
	// 5th octet should NOT have FX bit set (no 6th octet allowed)
	maxFSPEC := []byte{0x01, 0x01, 0x01, 0x01, 0x80, 0x05}

	ages := &v120.TrackDataAges{}
	buf := bytes.NewBuffer(maxFSPEC)
	bytesRead, err := ages.Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if bytesRead != len(maxFSPEC) {
		t.Errorf("Decode() read %d bytes, want %d", bytesRead, len(maxFSPEC))
	}

	// Re-encode
	encodeBuf := new(bytes.Buffer)
	_, err = ages.Encode(encodeBuf)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if !bytes.Equal(encodeBuf.Bytes(), maxFSPEC) {
		t.Errorf("Round trip failed for max FSPEC:\n  original: %X\n  reencoded: %X",
			maxFSPEC, encodeBuf.Bytes())
	}
}
