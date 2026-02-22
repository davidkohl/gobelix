package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Mode3ACode implements I011/060 - Mode-3/A Code in Octal Representation
// Definition: Track Mode-3/A code converted into octal representation.
// Format: Two-octet fixed length Data Item
// Bits 16-13: spare (set to 0)
// Bits 12-1: Mode-3/A reply in octal representation (A4A2A1 B4B2B1 C4C2C1 D4D2D1)
type Mode3ACode struct {
	Code uint16 // Octal code (0-4095, representing 0000-7777 octal)
}

func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading mode-3/A code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("mode-3/A code: expected 2 bytes, got %d", n)
	}

	raw := binary.BigEndian.Uint16(data[:])
	// Bits 12-1 contain the code (bits 16-13 are spare)
	m.Code = raw & 0x0FFF

	return 2, nil
}

func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	// Bits 16-13 are spare (0), bits 12-1 contain the code
	binary.BigEndian.PutUint16(data[:], m.Code&0x0FFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing mode-3/A code: %w", err)
	}
	return n, nil
}

func (m *Mode3ACode) Validate() error {
	if m.Code > 0x0FFF {
		return fmt.Errorf("mode-3/A code out of range: %d (max 4095)", m.Code)
	}
	// Check each octal digit is valid (0-7)
	a := (m.Code >> 9) & 0x07
	b := (m.Code >> 6) & 0x07
	c := (m.Code >> 3) & 0x07
	d := m.Code & 0x07

	if a > 7 || b > 7 || c > 7 || d > 7 {
		return fmt.Errorf("mode-3/A code contains invalid octal digit")
	}
	return nil
}

// OctalString returns the code as a 4-digit octal string
func (m *Mode3ACode) OctalString() string {
	a := (m.Code >> 9) & 0x07
	b := (m.Code >> 6) & 0x07
	c := (m.Code >> 3) & 0x07
	d := m.Code & 0x07
	return fmt.Sprintf("%d%d%d%d", a, b, c, d)
}

func (m *Mode3ACode) String() string {
	return fmt.Sprintf("Mode-3/A: %s", m.OctalString())
}
