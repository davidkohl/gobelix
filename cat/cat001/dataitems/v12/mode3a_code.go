// cat/cat001/dataitems/v12/mode3a_code.go
package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode3ACode represents I001/070 - Mode-3/A Code in Octal Representation
// Fixed length: 2 bytes
type Mode3ACode struct {
	V    bool   // Validated
	G    bool   // Garbled
	L    bool   // Failed
	Mode uint16 // Mode-3/A code in octal (12 bits)
}

// Decode decodes Mode-3/A Code from bytes
func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for mode-3/A code, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// First byte: V G L spare(5 bits)
	m.V = (data[0] & 0x80) != 0
	m.G = (data[0] & 0x40) != 0
	m.L = (data[0] & 0x20) != 0

	// Mode code: 4 bits from first byte + 8 bits from second byte
	m.Mode = (uint16(data[0]&0x0F) << 8) | uint16(data[1])

	return 2, nil
}

// Encode encodes Mode-3/A Code to bytes
func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Build first byte: V G L spare(5 bits)
	byte1 := byte(0)
	if m.V {
		byte1 |= 0x80
	}
	if m.G {
		byte1 |= 0x40
	}
	if m.L {
		byte1 |= 0x20
	}

	// Add 4 most significant bits of mode code
	byte1 |= byte((m.Mode >> 8) & 0x0F)

	// Second byte: 8 least significant bits of mode code
	byte2 := byte(m.Mode & 0xFF)

	data := []byte{byte1, byte2}
	n, err := buf.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing mode-3/A code: %w", err)
	}

	return n, nil
}

// Validate validates the Mode-3/A Code
func (m *Mode3ACode) Validate() error {
	// Mode-3/A code is 12 bits (4 octal digits)
	if m.Mode > 0x0FFF {
		return fmt.Errorf("%w: mode-3/A code exceeds 12 bits: %04X", asterix.ErrInvalidMessage, m.Mode)
	}
	return nil
}

// String returns a string representation
func (m *Mode3ACode) String() string {
	flags := ""
	if m.V {
		flags += "V"
	}
	if m.G {
		flags += "G"
	}
	if m.L {
		flags += "L"
	}
	if flags != "" {
		flags = " [" + flags + "]"
	}
	return fmt.Sprintf("%04o%s", m.Mode, flags)
}
