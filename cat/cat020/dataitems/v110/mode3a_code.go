// cat/cat020/dataitems/v110/mode3a_code.go
package v110

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode3ACode represents I020/070 - Mode-3/A Code in Octal Representation
// Fixed length: 2 bytes
// Mode-3/A code converted into octal representation
type Mode3ACode struct {
	V        bool   // Validated (0 = Code validated, 1 = Code not validated)
	G        bool   // Garbled (0 = Default, 1 = Garbled code)
	L        bool   // Smoothed (0 = Mode-3/A code derived from transponder, 1 = Smoothed)
	Mode3A   uint16 // Mode-3/A reply in octal representation (12 bits, 0-7777 octal)
}

// NewMode3ACode creates a new Mode-3/A Code data item
func NewMode3ACode() *Mode3ACode {
	return &Mode3ACode{}
}

// Decode decodes the Mode-3/A Code from bytes
func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// Bit 16: V (Validated)
	m.V = (data[0] & 0x80) != 0

	// Bit 15: G (Garbled)
	m.G = (data[0] & 0x40) != 0

	// Bit 14: L (Smoothed)
	m.L = (data[0] & 0x20) != 0

	// Bit 13: Spare (should be 0)

	// Bits 12-1: Mode-3/A code in octal (12 bits)
	m.Mode3A = (uint16(data[0]&0x0F) << 8) | uint16(data[1])

	return 2, nil
}

// Encode encodes the Mode-3/A Code to bytes
func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	data := make([]byte, 2)

	// Bit 16: V
	if m.V {
		data[0] |= 0x80
	}

	// Bit 15: G
	if m.G {
		data[0] |= 0x40
	}

	// Bit 14: L
	if m.L {
		data[0] |= 0x20
	}

	// Bit 13: Spare (0)

	// Bits 12-1: Mode-3/A code
	data[0] |= byte((m.Mode3A >> 8) & 0x0F)
	data[1] = byte(m.Mode3A & 0xFF)

	if _, err := buf.Write(data); err != nil {
		return 0, fmt.Errorf("writing Mode-3/A code: %w", err)
	}

	return 2, nil
}

// Validate validates the Mode-3/A Code
func (m *Mode3ACode) Validate() error {
	// Mode-3/A code should be 12 bits (0-4095 decimal, 0-7777 octal)
	if m.Mode3A > 0x0FFF {
		return fmt.Errorf("%w: Mode-3/A code must be 0-4095, got %d", asterix.ErrInvalidMessage, m.Mode3A)
	}
	return nil
}

// String returns a string representation
func (m *Mode3ACode) String() string {
	flags := ""
	if m.V {
		flags += " V"
	}
	if m.G {
		flags += " G"
	}
	if m.L {
		flags += " L"
	}
	// Display in octal as per ASTERIX spec
	return fmt.Sprintf("%04o%s", m.Mode3A, flags)
}
