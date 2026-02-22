// asterix/dataitem_test.go
package asterix

import (
	"bytes"
	"io"
	"testing"
)

func TestItemTypeString(t *testing.T) {
	testCases := []struct {
		itemType ItemType
		expected string
	}{
		{Fixed, "Fixed"},
		{Extended, "Extended"},
		{Explicit, "Explicit"},
		{Repetitive, "Repetitive"},
		{Compound, "Compound"},
		{ItemType(99), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.itemType.String()
			if result != tc.expected {
				t.Errorf("String() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestReadFull(t *testing.T) {
	// Test successful read
	t.Run("Success", func(t *testing.T) {
		data := []byte{1, 2, 3, 4, 5}
		r := bytes.NewReader(data)
		buf := make([]byte, 3)

		n, err := ReadFull(r, buf)
		if err != nil {
			t.Errorf("ReadFull() error = %v", err)
		}
		if n != 3 {
			t.Errorf("ReadFull() read %d bytes, want 3", n)
		}
		if !bytes.Equal(buf, []byte{1, 2, 3}) {
			t.Errorf("ReadFull() = %v, want %v", buf, []byte{1, 2, 3})
		}
	})

	// Test EOF
	t.Run("EOF", func(t *testing.T) {
		data := []byte{1, 2}
		r := bytes.NewReader(data)
		buf := make([]byte, 3)

		_, err := ReadFull(r, buf)
		if err != io.ErrUnexpectedEOF {
			t.Errorf("ReadFull() error = %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})
}

func TestReadUint16(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected uint16
		wantErr  bool
	}{
		{"Zero", []byte{0, 0}, 0, false},
		{"One", []byte{0, 1}, 1, false},
		{"Max", []byte{0xFF, 0xFF}, 0xFFFF, false},
		{"Big endian", []byte{0x12, 0x34}, 0x1234, false},
		{"EOF first byte", []byte{}, 0, true},
		{"EOF second byte", []byte{1}, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.data)
			result, err := ReadUint16(r)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ReadUint16() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("ReadUint16() error = %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("ReadUint16() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestReadUint24(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected uint32
		wantErr  bool
	}{
		{"Zero", []byte{0, 0, 0}, 0, false},
		{"One", []byte{0, 0, 1}, 1, false},
		{"Max", []byte{0xFF, 0xFF, 0xFF}, 0xFFFFFF, false},
		{"Big endian", []byte{0x12, 0x34, 0x56}, 0x123456, false},
		{"EOF first byte", []byte{}, 0, true},
		{"EOF second byte", []byte{1}, 0, true},
		{"EOF third byte", []byte{1, 2}, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.data)
			result, err := ReadUint24(r)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ReadUint24() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("ReadUint24() error = %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("ReadUint24() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestReadUint32(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected uint32
		wantErr  bool
	}{
		{"Zero", []byte{0, 0, 0, 0}, 0, false},
		{"One", []byte{0, 0, 0, 1}, 1, false},
		{"Max", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0xFFFFFFFF, false},
		{"Big endian", []byte{0x12, 0x34, 0x56, 0x78}, 0x12345678, false},
		{"EOF first byte", []byte{}, 0, true},
		{"EOF second byte", []byte{1}, 0, true},
		{"EOF third byte", []byte{1, 2}, 0, true},
		{"EOF fourth byte", []byte{1, 2, 3}, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.data)
			result, err := ReadUint32(r)

			if tc.wantErr {
				if err == nil {
					t.Errorf("ReadUint32() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("ReadUint32() error = %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("ReadUint32() = %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestWriteUint16(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{"Zero", 0, []byte{0, 0}},
		{"One", 1, []byte{0, 1}},
		{"Max", 0xFFFF, []byte{0xFF, 0xFF}},
		{"Big endian", 0x1234, []byte{0x12, 0x34}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := WriteUint16(buf, tc.value)

			if err != nil {
				t.Errorf("WriteUint16() error = %v", err)
				return
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("WriteUint16() wrote %v, want %v", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestWriteUint24(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"Zero", 0, []byte{0, 0, 0}},
		{"One", 1, []byte{0, 0, 1}},
		{"Max 24-bit", 0xFFFFFF, []byte{0xFF, 0xFF, 0xFF}},
		{"Big endian", 0x123456, []byte{0x12, 0x34, 0x56}},
		{"Truncate", 0x01234567, []byte{0x23, 0x45, 0x67}}, // Only 24 bits used
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := WriteUint24(buf, tc.value)

			if err != nil {
				t.Errorf("WriteUint24() error = %v", err)
				return
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("WriteUint24() wrote %v, want %v", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestWriteUint32(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"Zero", 0, []byte{0, 0, 0, 0}},
		{"One", 1, []byte{0, 0, 0, 1}},
		{"Max", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"Big endian", 0x12345678, []byte{0x12, 0x34, 0x56, 0x78}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := WriteUint32(buf, tc.value)

			if err != nil {
				t.Errorf("WriteUint32() error = %v", err)
				return
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("WriteUint32() wrote %v, want %v", buf.Bytes(), tc.expected)
			}
		})
	}
}

// Test the base implementations for each item type
func TestBaseFixedLengthItem(t *testing.T) {
	item := &BaseFixedLengthItem{length: 42}
	if item.FixedLength() != 42 {
		t.Errorf("FixedLength() = %d, want 42", item.FixedLength())
	}
}

func TestBaseExtendedLengthItem(t *testing.T) {
	// Test with extension
	item1 := &BaseExtendedLengthItem{hasExtension: true}
	if !item1.HasExtension() {
		t.Errorf("HasExtension() = false, want true")
	}

	// Test without extension
	item2 := &BaseExtendedLengthItem{hasExtension: false}
	if item2.HasExtension() {
		t.Errorf("HasExtension() = true, want false")
	}
}

func TestBaseExplicitLengthItem(t *testing.T) {
	item := &BaseExplicitLengthItem{length: 123}
	if item.LengthIndicator() != 123 {
		t.Errorf("LengthIndicator() = %d, want 123", item.LengthIndicator())
	}
}

func TestBaseRepetitiveItem(t *testing.T) {
	item := &BaseRepetitiveItem{count: 5}
	if item.RepetitionCount() != 5 {
		t.Errorf("RepetitionCount() = %d, want 5", item.RepetitionCount())
	}
}

func TestBaseCompoundItem(t *testing.T) {
	item := &BaseCompoundItem{subitemCount: 7}
	if item.SubitemCount() != 7 {
		t.Errorf("SubitemCount() = %d, want 7", item.SubitemCount())
	}
}
