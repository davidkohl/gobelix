// asterix/fspec_test.go
package asterix

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestFSPECNew(t *testing.T) {
	f := NewFSPEC()
	if f == nil {
		t.Fatal("NewFSPEC() returned nil")
	}
	if f.size != 0 {
		t.Errorf("expected size 0, got %d", f.size)
	}
}

func TestFSPECSetGetFRN(t *testing.T) {
	tests := []struct {
		name     string
		frns     []uint8
		checkFRN uint8
		expected bool
	}{
		{"Single bit", []uint8{1}, 1, true},
		{"Multiple bits", []uint8{1, 3, 5, 7}, 3, true},
		{"Missing bit", []uint8{1, 3, 5, 7}, 2, false},
		{"High bit", []uint8{14}, 14, true},
		{"Very high bit", []uint8{63}, 63, true},
		{"Across multiple bytes", []uint8{1, 8, 15}, 15, true},
		{"Byte boundary", []uint8{7, 8}, 8, true},
		{"FX bit positions", []uint8{8, 16, 24}, 16, true},
		{"Large skip", []uint8{1, 42}, 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFSPEC()
			for _, frn := range tt.frns {
				if err := f.SetFRN(frn); err != nil {
					t.Fatalf("SetFRN(%d) failed: %v", frn, err)
				}
			}

			if got := f.GetFRN(tt.checkFRN); got != tt.expected {
				t.Errorf("GetFRN(%d) = %v, want %v", tt.checkFRN, got, tt.expected)
			}

			// Additional check: all set FRNs should be retrievable
			for _, frn := range tt.frns {
				if got := f.GetFRN(frn); !got {
					t.Errorf("GetFRN(%d) = false, want true", frn)
				}
			}
		})
	}
}

func TestFSPECInvalidFRN(t *testing.T) {
	f := NewFSPEC()
	if err := f.SetFRN(0); err == nil {
		t.Error("SetFRN(0) should return error, got nil")
	}
}

func TestFSPECEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		frns []uint8
		size int
	}{
		{"Single byte", []uint8{1, 2, 3}, 1},
		{"Two bytes", []uint8{1, 8, 9}, 2},
		{"Three bytes", []uint8{1, 8, 15, 16}, 3},
		{"Empty FSPEC encode", []uint8{}, 0},
		{"High FRN", []uint8{42}, 6},
		{"Multiple high FRNs", []uint8{7, 14, 21, 28, 35}, 5},
		{"All bits in first byte", []uint8{1, 2, 3, 4, 5, 6, 7}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip empty FSPEC test for encoding
			if len(tc.frns) == 0 {
				return
			}

			// Create and populate the FSPEC
			f1 := NewFSPEC()
			for _, frn := range tc.frns {
				if err := f1.SetFRN(frn); err != nil {
					t.Fatalf("SetFRN(%d) failed: %v", frn, err)
				}
			}

			// Check expected size
			if tc.size > 0 && f1.Size() != tc.size {
				t.Errorf("Size() = %d, want %d", f1.Size(), tc.size)
			}

			// Encode
			buf := new(bytes.Buffer)
			n, err := f1.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if n != f1.Size() {
				t.Errorf("Encode() wrote %d bytes, want %d", n, f1.Size())
			}
			if buf.Len() != f1.Size() {
				t.Errorf("Encoded buffer has length %d, want %d", buf.Len(), f1.Size())
			}

			// Decode into a new FSPEC
			f2 := NewFSPEC()
			n, err = f2.Decode(buf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if n != f1.Size() {
				t.Errorf("Decode() read %d bytes, want %d", n, f1.Size())
			}

			// Verify both FSPECs are identical
			if !reflect.DeepEqual(f1.data(), f2.data()) {
				t.Errorf("Decoded data %v, want %v", f2.data(), f1.data())
			}

			// Verify each FRN that was set is still set
			for _, frn := range tc.frns {
				if !f2.GetFRN(frn) {
					t.Errorf("FRN %d not set after decode", frn)
				}
			}
		})
	}

	// Specific test for encode error when empty
	t.Run("Encode empty FSPEC error", func(t *testing.T) {
		f := NewFSPEC()
		buf := new(bytes.Buffer)
		_, err := f.Encode(buf)
		if err == nil {
			t.Error("Encode() with empty FSPEC should return error")
		}
	})
}

func TestFSPECDecodeInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expectFX bool
	}{
		{"EOF on first byte", []byte{}, false},
		{"EOF on extension", []byte{0x01}, true},
		{"Too many extensions", []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFSPEC()
			r := bytes.NewReader(tc.input)
			_, err := f.Decode(r)

			// Should get an error
			if err == nil {
				t.Error("Expected error, got nil")
			}

			// Check error type
			if len(tc.input) == 0 {
				if !errors.Is(err, io.EOF) {
					t.Errorf("Expected io.EOF (wrapped or unwrapped), got %v", err)
				}
			} else if tc.input[0]&0x01 != 0 && len(tc.input) == 1 {
				// EOF on extension byte - should get EOF error
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					t.Errorf("Expected EOF-related error for FX extension, got %v", err)
				}
			} else if len(tc.input) > 7 {
				// Too many extensions - should get specific error (not EOF)
				// Just verify we got an error, which we already checked above
			}
		})
	}
}

func TestFSPECByteOperations(t *testing.T) {
	// Create a sample FSPEC with some bits set
	f1 := NewFSPEC()
	for _, frn := range []uint8{1, 5, 9, 14} {
		if err := f1.SetFRN(frn); err != nil {
			t.Fatalf("SetFRN(%d) failed: %v", frn, err)
		}
	}

	// Create a buffer to hold the encoded data
	data := make([]byte, 10)

	// Encode to buffer
	n, err := f1.EncodeToBytes(data, 0)
	if err != nil {
		t.Fatalf("EncodeToBytes failed: %v", err)
	}

	// Decode from buffer
	f2 := NewFSPEC()
	n2, err := f2.DecodeFromBytes(data, 0)
	if err != nil {
		t.Fatalf("DecodeFromBytes failed: %v", err)
	}

	// Verify correct number of bytes processed
	if n != n2 {
		t.Errorf("Encoded %d bytes but decoded %d bytes", n, n2)
	}

	// Verify both FSPECs are identical
	if !reflect.DeepEqual(f1.data(), f2.data()) {
		t.Errorf("Decoded data %v, want %v", f2.data(), f1.data())
	}

	// Test buffer too small
	tinyBuffer := make([]byte, 1)
	_, err = f1.EncodeToBytes(tinyBuffer, 0)
	if err == nil && f1.Size() > 1 {
		t.Error("Expected error for buffer too small, got nil")
	}

	// Test offset out of bounds
	_, err = f2.DecodeFromBytes(data, len(data)+1)
	if err == nil {
		t.Error("Expected error for offset out of bounds, got nil")
	}
}

func TestFSPECReset(t *testing.T) {
	f := NewFSPEC()
	for _, frn := range []uint8{1, 3, 7, 14} {
		if err := f.SetFRN(frn); err != nil {
			t.Fatalf("SetFRN(%d) failed: %v", frn, err)
		}
	}

	// Verify size before reset
	initialSize := f.Size()
	if initialSize == 0 {
		t.Error("Initial size should be > 0")
	}

	// Reset and verify
	f.Reset()
	if f.Size() != 0 {
		t.Errorf("Size after reset is %d, want 0", f.Size())
	}

	// Verify no bits are set
	for i := uint8(1); i <= 20; i++ {
		if f.GetFRN(i) {
			t.Errorf("FRN %d is still set after reset", i)
		}
	}
}

func TestFSPECCopy(t *testing.T) {
	original := NewFSPEC()
	for _, frn := range []uint8{2, 5, 11, 17} {
		if err := original.SetFRN(frn); err != nil {
			t.Fatalf("SetFRN(%d) failed: %v", frn, err)
		}
	}

	// Create copy and verify
	copy := original.Copy()

	// Should have same size
	if copy.Size() != original.Size() {
		t.Errorf("Copy size %d, want %d", copy.Size(), original.Size())
	}

	// Should have same data
	if !reflect.DeepEqual(original.data(), copy.data()) {
		t.Errorf("Copy data %v, want %v", copy.data(), original.data())
	}

	// Check all bits match
	for i := uint8(1); i <= 20; i++ {
		if original.GetFRN(i) != copy.GetFRN(i) {
			t.Errorf("FRN %d differs in copy", i)
		}
	}

	// Modify original and verify copy is unaffected
	original.SetFRN(7)
	if copy.GetFRN(7) {
		t.Error("Modifying original affected copy")
	}
}

func TestFSPECBitCounting(t *testing.T) {
	tests := []struct {
		name         string
		frns         []uint8
		expectedBits int
	}{
		{"Empty", []uint8{}, 0},
		{"Single bit", []uint8{1}, 1},
		{"Multiple bits", []uint8{1, 3, 5}, 3},
		{"All bits in byte", []uint8{1, 2, 3, 4, 5, 6, 7}, 7},
		{"Across bytes", []uint8{1, 8, 15}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFSPEC()
			for _, frn := range tt.frns {
				if err := f.SetFRN(frn); err != nil {
					t.Fatalf("SetFRN(%d) failed: %v", frn, err)
				}
			}

			bits := f.FSPECBitCount()
			if bits != tt.expectedBits {
				t.Errorf("FSPECBitCount() = %d, want %d", bits, tt.expectedBits)
			}

			hasData := f.HasDataBits()
			if (tt.expectedBits > 0) != hasData {
				t.Errorf("HasDataBits() = %v, want %v", hasData, tt.expectedBits > 0)
			}
		})
	}
}

func TestFSPECUint64Conversion(t *testing.T) {
	testCases := []struct {
		name string
		bits uint64
		frns []uint8
	}{
		{"Single bit", 1, []uint8{1}},
		{"Multiple bits", 0x45, []uint8{1, 3, 7}}, // 2^0 + 2^2 + 2^6 = 69 (0x45)
		{"High bits", 0x100000000, []uint8{33}},   // 2^32
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create from uint64
			f1 := FSPECFromUint64(tc.bits)

			// Check all expected FRNs are set
			for _, frn := range tc.frns {
				if !f1.GetFRN(frn) {
					t.Errorf("FRN %d not set", frn)
				}
			}

			// Convert back to uint64
			bits := f1.ToUint64()
			if bits != tc.bits {
				t.Errorf("ToUint64() = %x, want %x", bits, tc.bits)
			}

			// For low bit values, manually check each FRN
			if tc.bits < 256 {
				for i := uint8(1); i <= 8; i++ {
					expected := (tc.bits & (1 << (i - 1))) != 0
					if f1.GetFRN(i) != expected {
						t.Errorf("FRN %d = %v, want %v", i, f1.GetFRN(i), expected)
					}
				}
			}
		})
	}
}

func BenchmarkFSPECSetFRN(b *testing.B) {
	frns := []uint8{1, 3, 7, 9, 15, 21, 28, 35}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := NewFSPEC()
		for _, frn := range frns {
			_ = f.SetFRN(frn)
		}
	}
}

func BenchmarkFSPECGetFRN(b *testing.B) {
	f := NewFSPEC()
	for _, frn := range []uint8{1, 3, 7, 9, 15, 21, 28, 35} {
		_ = f.SetFRN(frn)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := uint8(1); j <= 40; j++ {
			_ = f.GetFRN(j)
		}
	}
}

func BenchmarkFSPECEncodeDecode(b *testing.B) {
	f := NewFSPEC()
	for _, frn := range []uint8{1, 3, 7, 9, 15, 21, 28, 35} {
		_ = f.SetFRN(frn)
	}

	buf := new(bytes.Buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = f.Encode(buf)

		f2 := NewFSPEC()
		buf2 := bytes.NewReader(buf.Bytes())
		_, _ = f2.Decode(buf2)
	}
}

func BenchmarkFSPECByteOperations(b *testing.B) {
	f := NewFSPEC()
	for _, frn := range []uint8{1, 3, 7, 9, 15, 21, 28, 35} {
		_ = f.SetFRN(frn)
	}

	data := make([]byte, 16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = f.EncodeToBytes(data, 0)

		f2 := NewFSPEC()
		_, _ = f2.DecodeFromBytes(data, 0)
	}
}
