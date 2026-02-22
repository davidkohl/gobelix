// dataitems/cat048/v132/mode3a_code_test.go
package v132_test

import (
	"bytes"
	"testing"

	v132 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v132"
)

func TestMode3ACode_EncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    v132.Mode3ACode
		expected []byte
	}{
		{
			name: "Code 0000 - no flags",
			input: v132.Mode3ACode{
				V:    false,
				G:    false,
				L:    false,
				Code: 0000,
			},
			expected: []byte{0x00, 0x00},
		},
		{
			name: "Code 7777 - all flags set",
			input: v132.Mode3ACode{
				V:    true,
				G:    true,
				L:    true,
				Code: 7777,
			},
			expected: []byte{0xEF, 0xFF}, // V=1,G=1,L=1,spare=0,A=7,B=7,C=7,D=7
		},
		{
			name: "Code 1200 - typical VFR squawk, validated",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: 1200,
			},
			expected: []byte{0x82, 0x80}, // V=1,spare=0,A=1,B=2,C=0,D=0
		},
		{
			name: "Code 7700 - emergency, validated",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: 7700,
			},
			expected: []byte{0x8F, 0xC0}, // V=1,A=7,B=7,C=0,D=0
		},
		{
			name: "Code 7600 - radio failure",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: 7600,
			},
			expected: []byte{0x8F, 0x80}, // V=1,A=7,B=6,C=0,D=0
		},
		{
			name: "Code 7500 - hijack",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: 7500,
			},
			expected: []byte{0x8F, 0x40}, // V=1,A=7,B=5,C=0,D=0
		},
		{
			name: "Code 4321 - mixed digits",
			input: v132.Mode3ACode{
				V:    false,
				G:    false,
				L:    false,
				Code: 4321,
			},
			expected: []byte{0x08, 0xD1}, // A=4,B=3,C=2,D=1
		},
		{
			name: "Code with garbled flag",
			input: v132.Mode3ACode{
				V:    false,
				G:    true,
				L:    false,
				Code: 1234,
			},
			expected: []byte{0x42, 0x9C}, // G=1,A=1,B=2,C=3,D=4
		},
		{
			name: "Code with L flag (derived)",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    true,
				Code: 5555,
			},
			expected: []byte{0xAB, 0x6D}, // V=1,L=1,A=5,B=5,C=5,D=5
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
			if n != 2 {
				t.Errorf("Encode() returned n = %v, want 2", n)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("Encode() = %X, want %X", buf.Bytes(), tt.expected)
			}

			// Test Decode
			decoded := &v132.Mode3ACode{}
			encodedBuf := bytes.NewBuffer(tt.expected)
			n, err = decoded.Decode(encodedBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if n != 2 {
				t.Errorf("Decode() returned n = %v, want 2", n)
			}
			if decoded.V != tt.input.V {
				t.Errorf("Decode() V = %v, want %v", decoded.V, tt.input.V)
			}
			if decoded.G != tt.input.G {
				t.Errorf("Decode() G = %v, want %v", decoded.G, tt.input.G)
			}
			if decoded.L != tt.input.L {
				t.Errorf("Decode() L = %v, want %v", decoded.L, tt.input.L)
			}
			if decoded.Code != tt.input.Code {
				t.Errorf("Decode() Code = %04o, want %04o", decoded.Code, tt.input.Code)
			}
		})
	}
}

func TestMode3ACode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   v132.Mode3ACode
		wantErr bool
	}{
		{
			name: "Valid code 0000",
			input: v132.Mode3ACode{
				V:    true,
				Code: 0000,
			},
			wantErr: false,
		},
		{
			name: "Valid code 7777",
			input: v132.Mode3ACode{
				V:    true,
				Code: 7777,
			},
			wantErr: false,
		},
		{
			name: "Valid code 1200",
			input: v132.Mode3ACode{
				V:    true,
				Code: 1200,
			},
			wantErr: false,
		},
		{
			name: "Invalid code 8000 (digit > 7)",
			input: v132.Mode3ACode{
				V:    true,
				Code: 8000,
			},
			wantErr: true,
		},
		{
			name: "Invalid code 1289 (digit > 7)",
			input: v132.Mode3ACode{
				V:    true,
				Code: 1289,
			},
			wantErr: true,
		},
		{
			name: "Invalid code 7778 (last digit > 7)",
			input: v132.Mode3ACode{
				V:    true,
				Code: 7778,
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

func TestMode3ACode_String(t *testing.T) {
	tests := []struct {
		name     string
		input    v132.Mode3ACode
		expected string
	}{
		{
			name: "Code with no flags",
			input: v132.Mode3ACode{
				V:    false,
				G:    false,
				L:    false,
				Code: 1200,
			},
			expected: "2260", // Octal output format
		},
		{
			name: "Code with V flag",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: 7700,
			},
			expected: "V 17024", // Octal output format
		},
		{
			name: "Code with G flag",
			input: v132.Mode3ACode{
				V:    false,
				G:    true,
				L:    false,
				Code: 1234,
			},
			expected: "G 2322", // Octal output format
		},
		{
			name: "Code with V and L flags",
			input: v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    true,
				Code: 4321,
			},
			expected: "V,L 10341", // Octal output format
		},
		{
			name: "Code with all flags",
			input: v132.Mode3ACode{
				V:    true,
				G:    true,
				L:    true,
				Code: 7777,
			},
			expected: "V,G,L 17141", // Octal output format
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

func TestMode3ACode_RoundTrip(t *testing.T) {
	// Test emergency and special codes
	testCodes := []uint16{
		0000, // No transponder
		1200, // VFR
		7500, // Hijack
		7600, // Radio failure
		7700, // Emergency
		4321, // Arbitrary
		7777, // Maximum
	}

	for _, code := range testCodes {
		t.Run("", func(t *testing.T) {
			original := v132.Mode3ACode{
				V:    true,
				G:    false,
				L:    false,
				Code: code,
			}

			// Encode
			buf := new(bytes.Buffer)
			_, err := original.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded := &v132.Mode3ACode{}
			_, err = decoded.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			// Compare
			if decoded.Code != original.Code {
				t.Errorf("Round trip failed: got %04o, want %04o", decoded.Code, original.Code)
			}
			if decoded.V != original.V {
				t.Errorf("Round trip V failed: got %v, want %v", decoded.V, original.V)
			}
			if decoded.G != original.G {
				t.Errorf("Round trip G failed: got %v, want %v", decoded.G, original.G)
			}
			if decoded.L != original.L {
				t.Errorf("Round trip L failed: got %v, want %v", decoded.L, original.L)
			}
		})
	}
}

func TestMode3ACode_DecodeIncomplete(t *testing.T) {
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
			input:   []byte{0x80},
			wantErr: true,
		},
		{
			name:    "Complete - 2 bytes",
			input:   []byte{0x80, 0x00},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &v132.Mode3ACode{}
			buf := bytes.NewBuffer(tt.input)
			_, err := code.Decode(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
