// cat/cat020/dataitems/v15/mode_codes.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode1Code implements I020/055 - Mode-1 Code in Octal Representation
type Mode1Code struct {
	V    bool  // Validated
	G    bool  // Garbled
	L    bool  // Failed (L bit)
	Code uint8 // Mode-1 code (5 bits, 0-31)
}

func (m *Mode1Code) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading mode 1 code: %w", err)
	}

	// Bit 8: V (validated)
	// Bit 7: G (garbled)
	// Bit 6: L (failed)
	// Bits 5-1: Mode-1 code
	m.V = (data & 0x80) != 0
	m.G = (data & 0x40) != 0
	m.L = (data & 0x20) != 0
	m.Code = data & 0x1F

	return 1, nil
}

func (m *Mode1Code) Encode(buf *bytes.Buffer) (int, error) {
	data := m.Code & 0x1F
	if m.V {
		data |= 0x80
	}
	if m.G {
		data |= 0x40
	}
	if m.L {
		data |= 0x20
	}

	err := buf.WriteByte(data)
	if err != nil {
		return 0, fmt.Errorf("writing mode 1 code: %w", err)
	}
	return 1, nil
}

func (m *Mode1Code) Validate() error {
	if m.Code > 31 {
		return fmt.Errorf("mode 1 code out of range: %d", m.Code)
	}
	return nil
}

func (m *Mode1Code) String() string {
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
	return fmt.Sprintf("Mode1: %02o%s", m.Code, flags)
}

// Mode2Code implements I020/050 - Mode-2 Code in Octal Representation
type Mode2Code struct {
	V    bool   // Validated
	G    bool   // Garbled
	L    bool   // Failed (L bit)
	Code uint16 // Mode-2 code (12 bits, 0-4095)
}

func (m *Mode2Code) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading mode 2 code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for mode 2 code, have %d", asterix.ErrBufferTooShort, n)
	}

	// Bit 16: spare
	// Bit 15: V (validated)
	// Bit 14: G (garbled)
	// Bit 13: L (failed)
	// Bits 12-1: Mode-2 code
	m.V = (data[0] & 0x40) != 0
	m.G = (data[0] & 0x20) != 0
	m.L = (data[0] & 0x10) != 0
	m.Code = (uint16(data[0]&0x0F) << 8) | uint16(data[1])

	return n, nil
}

func (m *Mode2Code) Encode(buf *bytes.Buffer) (int, error) {
	var data [2]byte

	// Bit 16: spare (0)
	if m.V {
		data[0] |= 0x40
	}
	if m.G {
		data[0] |= 0x20
	}
	if m.L {
		data[0] |= 0x10
	}
	data[0] |= byte((m.Code >> 8) & 0x0F)
	data[1] = byte(m.Code & 0xFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing mode 2 code: %w", err)
	}
	return n, nil
}

func (m *Mode2Code) Validate() error {
	if m.Code > 4095 {
		return fmt.Errorf("mode 2 code out of range: %d", m.Code)
	}
	return nil
}

func (m *Mode2Code) String() string {
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
	return fmt.Sprintf("Mode2: %04o%s", m.Code, flags)
}
