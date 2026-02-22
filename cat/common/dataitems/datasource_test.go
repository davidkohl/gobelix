// dataitems/common/datasource_test.go
package common_test

import (
	"bytes"
	"testing"

	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

func TestDataSourceIdentifier_EncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    common.DataSourceIdentifier
		expected []byte
		wantErr  bool
	}{
		{
			name: "Valid minimum values",
			input: common.DataSourceIdentifier{
				SAC: 1,
				SIC: 1,
			},
			expected: []byte{0x01, 0x01},
			wantErr:  false,
		},
		{
			name: "Valid typical values",
			input: common.DataSourceIdentifier{
				SAC: 25,
				SIC: 100,
			},
			expected: []byte{0x19, 0x64},
			wantErr:  false,
		},
		{
			name: "Valid maximum values",
			input: common.DataSourceIdentifier{
				SAC: 255,
				SIC: 255,
			},
			expected: []byte{0xFF, 0xFF},
			wantErr:  false,
		},
		{
			name: "Invalid SAC zero",
			input: common.DataSourceIdentifier{
				SAC: 0,
				SIC: 100,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid SIC zero",
			input: common.DataSourceIdentifier{
				SAC: 25,
				SIC: 0,
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid both zero",
			input: common.DataSourceIdentifier{
				SAC: 0,
				SIC: 0,
			},
			expected: nil,
			wantErr:  true,
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
				if n != len(tt.expected) {
					t.Errorf("Encode() returned n = %v, want %v", n, len(tt.expected))
				}
				if !bytes.Equal(buf.Bytes(), tt.expected) {
					t.Errorf("Encode() = %X, want %X", buf.Bytes(), tt.expected)
				}

				// Test Decode
				decodedItem := &common.DataSourceIdentifier{}
				encodedBuf := bytes.NewBuffer(tt.expected)
				n, err = decodedItem.Decode(encodedBuf)
				if err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if n != len(tt.expected) {
					t.Errorf("Decode() returned n = %v, want %v", n, len(tt.expected))
				}
				if decodedItem.SAC != tt.input.SAC {
					t.Errorf("Decode() SAC = %v, want %v", decodedItem.SAC, tt.input.SAC)
				}
				if decodedItem.SIC != tt.input.SIC {
					t.Errorf("Decode() SIC = %v, want %v", decodedItem.SIC, tt.input.SIC)
				}
			}
		})
	}
}

func TestDataSourceIdentifier_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   common.DataSourceIdentifier
		wantErr bool
	}{
		{
			name: "Valid values",
			input: common.DataSourceIdentifier{
				SAC: 25,
				SIC: 100,
			},
			wantErr: false,
		},
		{
			name: "Invalid SAC zero",
			input: common.DataSourceIdentifier{
				SAC: 0,
				SIC: 100,
			},
			wantErr: true,
		},
		{
			name: "Invalid SIC zero",
			input: common.DataSourceIdentifier{
				SAC: 25,
				SIC: 0,
			},
			wantErr: true,
		},
		{
			name: "Valid minimum non-zero",
			input: common.DataSourceIdentifier{
				SAC: 1,
				SIC: 1,
			},
			wantErr: false,
		},
		{
			name: "Valid maximum",
			input: common.DataSourceIdentifier{
				SAC: 255,
				SIC: 255,
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

func TestDataSourceIdentifier_String(t *testing.T) {
	tests := []struct {
		name     string
		input    common.DataSourceIdentifier
		expected string
	}{
		{
			name: "Typical values",
			input: common.DataSourceIdentifier{
				SAC: 25,
				SIC: 100,
			},
			expected: "SAC: 25, SIC: 100",
		},
		{
			name: "Minimum values",
			input: common.DataSourceIdentifier{
				SAC: 1,
				SIC: 1,
			},
			expected: "SAC: 1, SIC: 1",
		},
		{
			name: "Maximum values",
			input: common.DataSourceIdentifier{
				SAC: 255,
				SIC: 255,
			},
			expected: "SAC: 255, SIC: 255",
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

func TestDataSourceIdentifier_DecodeIncomplete(t *testing.T) {
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
			input:   []byte{0x19},
			wantErr: true,
		},
		{
			name:    "Complete - 2 bytes",
			input:   []byte{0x19, 0x64},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsi := &common.DataSourceIdentifier{}
			buf := bytes.NewBuffer(tt.input)
			_, err := dsi.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataSourceIdentifier_RoundTrip(t *testing.T) {
	// Test that encode followed by decode returns the original value
	testPairs := []struct {
		SAC uint8
		SIC uint8
	}{
		{1, 1},
		{25, 100},
		{100, 50},
		{255, 255},
		{128, 64},
	}

	for _, pair := range testPairs {
		t.Run("", func(t *testing.T) {
			original := common.DataSourceIdentifier{SAC: pair.SAC, SIC: pair.SIC}

			// Encode
			buf := new(bytes.Buffer)
			_, err := original.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded := &common.DataSourceIdentifier{}
			_, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Compare
			if decoded.SAC != original.SAC || decoded.SIC != original.SIC {
				t.Errorf("Round trip failed: got SAC=%d SIC=%d, want SAC=%d SIC=%d",
					decoded.SAC, decoded.SIC, original.SAC, original.SIC)
			}
		})
	}
}
