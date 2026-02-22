// dataitems/cat062/v120/system_track_update_ages_test.go
package v120_test

import (
	"bytes"
	"testing"

	v120 "github.com/davidkohl/gobelix/cat/cat062/dataitems/v120"
)

// TestSystemTrackUpdateAges_RoundTrip tests that encoded data can be decoded back correctly
func TestSystemTrackUpdateAges_RoundTrip(t *testing.T) {
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
			name:    "Single subfield - PSR age (bit 8)",
			encoded: []byte{0x80, 0x05}, // FSPEC=0x80, PSR age=5
		},
		{
			name:    "Single subfield - PSR Track Age (bit 7)",
			encoded: []byte{0x40, 0x0A}, // FSPEC=0x40, PSR track age=10
		},
		{
			name:    "Two subfields - PSR age + PSR track age",
			encoded: []byte{0xC0, 0x05, 0x0A}, // FSPEC=0xC0, PSR age=5, PSR track age=10
		},
		{
			name:    "Subfield #5 (2 octets) - MDS age",
			encoded: []byte{0x10, 0x00, 0x14}, // FSPEC=0x10 (bit 12), MDS age=0x0014 (20)
		},
		{
			name:    "Multiple subfields without extension",
			encoded: []byte{0xF0, 0x01, 0x02, 0x00, 0x03, 0x04}, // Bits 8,7,6,5 set
		},
		{
			name:    "Extended FSPEC - single extension byte",
			encoded: []byte{0x01, 0x80, 0x05}, // FX bit set, second FSPEC byte=0x80, one subfield
		},
		{
			name:    "Extended FSPEC - multiple extension subfields",
			encoded: []byte{0x01, 0xC0, 0x05, 0x0A}, // FX bit set, bits 8+7 in second FSPEC
		},
		{
			name:    "Mixed - primary and extension subfields",
			encoded: []byte{0x81, 0x40, 0x05, 0x0A}, // Bit 8 in primary + bit 7 in extension
		},
		{
			name:    "All primary subfields set (no extension)",
			// FSPEC=0xFE: bits 8,7,6,5,4,3,2 set (7 subfields)
			// Subfield sizes: 1,1,1,2,1,1,1 = 8 bytes total
			encoded: []byte{0xFE, 0x01, 0x02, 0x03, 0x04, 0x00, 0x05, 0x06, 0x07},
		},
		{
			name:    "Minimal extension - no extension subfields",
			encoded: []byte{0x01, 0x00}, // FX bit set in primary, no subfields set in extension
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode the encoded bytes
			decoded := &v120.SystemTrackUpdateAges{}
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

// TestSystemTrackUpdateAges_DecodeTruncated tests handling of truncated/incomplete data
func TestSystemTrackUpdateAges_DecodeTruncated(t *testing.T) {
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
			name:    "Extension bit set but no extension byte",
			input:   []byte{0x01}, // FX bit set but no second FSPEC byte
			wantErr: true,
		},
		{
			name:    "Extension FSPEC indicates subfield but data missing",
			input:   []byte{0x01, 0x80}, // Extension bit + extension FSPEC bit 8, but no data
			wantErr: true,
		},
		{
			name:    "2-octet subfield truncated (only 1 byte)",
			input:   []byte{0x10, 0x00}, // Bit 12 set (2-octet field), only 1 byte provided
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
			name:    "Valid 2-octet subfield",
			input:   []byte{0x10, 0x00, 0x14},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ages := &v120.SystemTrackUpdateAges{}
			buf := bytes.NewBuffer(tt.input)
			_, err := ages.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v, input = %X", err, tt.wantErr, tt.input)
			}
		})
	}
}

// TestSystemTrackUpdateAges_EncodeEmpty tests encoding of empty/uninitialized data
func TestSystemTrackUpdateAges_EncodeEmpty(t *testing.T) {
	ages := &v120.SystemTrackUpdateAges{}

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

// TestSystemTrackUpdateAges_String tests the string representation
func TestSystemTrackUpdateAges_String(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "Empty",
			data:     []byte{},
			expected: "SystemTrackUpdateAges[0 bytes]",
		},
		{
			name:     "Single byte",
			data:     []byte{0x00},
			expected: "SystemTrackUpdateAges[1 bytes]",
		},
		{
			name:     "Multiple bytes",
			data:     []byte{0xC0, 0x05, 0x0A},
			expected: "SystemTrackUpdateAges[3 bytes]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ages := &v120.SystemTrackUpdateAges{Data: tt.data}
			result := ages.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
