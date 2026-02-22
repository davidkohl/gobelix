// dataitems/cat021/v26/target_address_test.go
package v26_test

import (
	"bytes"
	"testing"

	v26 "github.com/davidkohl/gobelix/cat/cat021/dataitems/v26"
)

func TestTargetAddress_EncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    v26.TargetAddress
		expected []byte
		wantErr  bool
	}{
		{
			name: "Valid minimum address",
			input: v26.TargetAddress{
				Address: 0x000000,
			},
			expected: []byte{0x00, 0x00, 0x00},
			wantErr:  false,
		},
		{
			name: "Valid typical address",
			input: v26.TargetAddress{
				Address: 0xABCDEF,
			},
			expected: []byte{0xAB, 0xCD, 0xEF},
			wantErr:  false,
		},
		{
			name: "Valid maximum 24-bit address",
			input: v26.TargetAddress{
				Address: 0xFFFFFF,
			},
			expected: []byte{0xFF, 0xFF, 0xFF},
			wantErr:  false,
		},
		{
			name: "Valid mid-range address",
			input: v26.TargetAddress{
				Address: 0x123456,
			},
			expected: []byte{0x12, 0x34, 0x56},
			wantErr:  false,
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
				decodedItem := &v26.TargetAddress{}
				encodedBuf := bytes.NewBuffer(tt.expected)
				n, err = decodedItem.Decode(encodedBuf)
				if err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if n != len(tt.expected) {
					t.Errorf("Decode() returned n = %v, want %v", n, len(tt.expected))
				}
				if decodedItem.Address != tt.input.Address {
					t.Errorf("Decode() Address = %X, want %X", decodedItem.Address, tt.input.Address)
				}
			}
		})
	}
}

func TestTargetAddress_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   v26.TargetAddress
		wantErr bool
	}{
		{
			name: "Valid address within 24 bits",
			input: v26.TargetAddress{
				Address: 0xFFFFFF,
			},
			wantErr: false,
		},
		{
			name: "Invalid address exceeds 24 bits",
			input: v26.TargetAddress{
				Address: 0x1000000, // 25 bits
			},
			wantErr: true,
		},
		{
			name: "Valid zero address",
			input: v26.TargetAddress{
				Address: 0x000000,
			},
			wantErr: false,
		},
		{
			name: "Invalid large address",
			input: v26.TargetAddress{
				Address: 0xFFFFFFFF, // 32 bits
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

func TestTargetAddress_String(t *testing.T) {
	tests := []struct {
		name     string
		input    v26.TargetAddress
		expected string
	}{
		{
			name: "Zero address",
			input: v26.TargetAddress{
				Address: 0x000000,
			},
			expected: "000000",
		},
		{
			name: "Typical address",
			input: v26.TargetAddress{
				Address: 0xABCDEF,
			},
			expected: "ABCDEF",
		},
		{
			name: "Maximum address",
			input: v26.TargetAddress{
				Address: 0xFFFFFF,
			},
			expected: "FFFFFF",
		},
		{
			name: "Small address with leading zeros",
			input: v26.TargetAddress{
				Address: 0x000123,
			},
			expected: "000123",
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

func TestTargetAddress_FromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
		wantErr  bool
	}{
		{
			name:     "Valid hex string uppercase",
			input:    "ABCDEF",
			expected: 0xABCDEF,
			wantErr:  false,
		},
		{
			name:     "Valid hex string lowercase",
			input:    "abcdef",
			expected: 0xABCDEF,
			wantErr:  false,
		},
		{
			name:     "Valid hex string with leading zeros",
			input:    "000123",
			expected: 0x000123,
			wantErr:  false,
		},
		{
			name:     "Invalid hex string - non-hex characters",
			input:    "GHIJKL",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Invalid hex string - too large",
			input:    "1ABCDEF",
			expected: 0x1ABCDEF,
			wantErr:  true, // Exceeds 24 bits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := &v26.TargetAddress{}
			err := ta.FromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && ta.Address != tt.expected {
				t.Errorf("FromString() Address = %X, want %X", ta.Address, tt.expected)
			}
		})
	}
}

func TestTargetAddress_DecodeIncomplete(t *testing.T) {
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
			input:   []byte{0xAB},
			wantErr: true,
		},
		{
			name:    "Incomplete - 2 bytes",
			input:   []byte{0xAB, 0xCD},
			wantErr: true,
		},
		{
			name:    "Complete - 3 bytes",
			input:   []byte{0xAB, 0xCD, 0xEF},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := &v26.TargetAddress{}
			buf := bytes.NewBuffer(tt.input)
			_, err := ta.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTargetAddress_RoundTrip(t *testing.T) {
	// Test that encode followed by decode returns the original value
	testAddresses := []uint32{
		0x000000,
		0x000001,
		0x123456,
		0xABCDEF,
		0x7FFFFF,
		0xFFFFFF,
	}

	for _, addr := range testAddresses {
		t.Run("", func(t *testing.T) {
			original := v26.TargetAddress{Address: addr}

			// Encode
			buf := new(bytes.Buffer)
			_, err := original.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded := &v26.TargetAddress{}
			_, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Compare
			if decoded.Address != original.Address {
				t.Errorf("Round trip failed: got %X, want %X", decoded.Address, original.Address)
			}
		})
	}
}
