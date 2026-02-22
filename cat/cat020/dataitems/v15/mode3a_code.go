// cat/cat020/dataitems/v15/mode3a_code.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode3ACode implements I020/070 - Mode-3/A Code in Octal Representation
type Mode3ACode struct {
	V    bool   // Validated
	G    bool   // Garbled
	L    bool   // Failed (L bit)
	Code uint16 // Mode-3/A code in octal (12 bits)
}

func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading mode 3/A code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for mode 3/A code, have %d", asterix.ErrBufferTooShort, n)
	}

	// Per CAT020 spec v1.11 page 20:
	// Bit 16: V (validated) - 0=validated, 1=not validated
	// Bit 15: G (garbled) - 0=default, 1=garbled
	// Bit 14: L (failed) - 0=derived from transponder, 1=not extracted
	// Bit 13: spare (set to 0)
	// Bits 12-1: Mode-3/A code in octal
	m.V = (data[0] & 0x80) != 0 // bit 16
	m.G = (data[0] & 0x40) != 0 // bit 15
	m.L = (data[0] & 0x20) != 0 // bit 14
	// bit 13 is spare
	m.Code = (uint16(data[0]&0x0F) << 8) | uint16(data[1]) // bits 12-1

	return n, nil
}

func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	var data [2]byte

	// Per CAT020 spec v1.11 page 20:
	// Bit 16: V, Bit 15: G, Bit 14: L, Bit 13: spare, Bits 12-1: code
	if m.V {
		data[0] |= 0x80 // bit 16
	}
	if m.G {
		data[0] |= 0x40 // bit 15
	}
	if m.L {
		data[0] |= 0x20 // bit 14
	}
	// bit 13 is spare (0)
	// Bits 12-1: Mode-3/A code
	data[0] |= byte((m.Code >> 8) & 0x0F)
	data[1] = byte(m.Code & 0xFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing mode 3/A code: %w", err)
	}
	return n, nil
}

func (m *Mode3ACode) Validate() error {
	// Mode-3/A code is 12 bits max
	if m.Code > 0xFFF {
		return fmt.Errorf("mode 3/A code out of range: 0x%X", m.Code)
	}
	return nil
}

func (m *Mode3ACode) String() string {
	flags := ""
	if !m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	if m.L {
		flags += " FAILED"
	}
	return fmt.Sprintf("Mode3A: %04o%s", m.Code, flags)
}
