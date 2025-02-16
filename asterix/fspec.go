// asterix/fspec.go
package asterix

import (
	"bytes"
	"fmt"
)

// FSPEC represents the Field Specification of an ASTERIX record
type FSPEC struct {
	bits []byte
}

// NewFSPEC creates a new empty FSPEC
func NewFSPEC() *FSPEC {
	return &FSPEC{
		bits: make([]byte, 0, 4), // Most FSPECs fit in 4 bytes
	}
}

// SetFRN marks a Field Reference Number as present
func (f *FSPEC) SetFRN(frn uint8) {
	byteIndex := (frn - 1) / 7 // 7 bits per byte (last bit is FX)
	bitPosition := (frn - 1) % 7

	// Extend FSPEC if needed
	for int(byteIndex) >= len(f.bits) {
		// Set FX bit in previous byte if it exists
		if len(f.bits) > 0 {
			f.bits[len(f.bits)-1] |= 0x01
		}
		f.bits = append(f.bits, 0)
	}

	// Set the bit
	f.bits[byteIndex] |= 0x80 >> bitPosition
}

// GetFRN checks if a Field Reference Number is present
func (f *FSPEC) GetFRN(frn uint8) bool {
	byteIndex := (frn - 1) / 7
	bitPosition := (frn - 1) % 7

	if int(byteIndex) >= len(f.bits) {
		return false
	}

	return f.bits[byteIndex]&(0x80>>bitPosition) != 0
}

// Encode writes the FSPEC to a buffer
func (f *FSPEC) Encode(buf *bytes.Buffer) (int, error) {
	n, err := buf.Write(f.bits)
	if err != nil {
		return n, fmt.Errorf("writing FSPEC bits: %w", err)
	}
	return n, nil
}

// Decode reads the FSPEC from a buffer
func (f *FSPEC) Decode(buf *bytes.Buffer) (int, error) {
	f.bits = f.bits[:0] // Reset existing bits
	bytesRead := 0

	for {
		b, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading FSPEC byte: %w", err)
		}
		bytesRead++

		f.bits = append(f.bits, b)

		// Check FX bit
		if b&0x01 == 0 {
			break // No more extensions
		}
	}

	return bytesRead, nil
}

// Size returns the size of the FSPEC in bytes
func (f *FSPEC) Size() int {
	return len(f.bits)
}
