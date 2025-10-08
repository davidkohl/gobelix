// cat/cat020/dataitems/v10/mode2_code.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode2Code represents I020/050 - Mode-2 Code in Octal Representation
// Fixed length: 2 bytes
// Mode-2 code converted into octal representation
type Mode2Code struct {
	V    bool   // Code validated (false) / Code not validated (true)
	G    bool   // Default (false) / Garbled code (true)
	L    bool   // Mode-2 code derived from transponder (false) / Smoothed code (true)
	Code uint16 // Mode-2 reply in octal representation (12 bits)
}

// NewMode2Code creates a new Mode-2 Code data item
func NewMode2Code() *Mode2Code {
	return &Mode2Code{}
}

// Decode decodes the Mode-2 Code from bytes
func (m *Mode2Code) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)
	value := binary.BigEndian.Uint16(data)

	m.V = (value & 0x8000) != 0
	m.G = (value & 0x4000) != 0
	m.L = (value & 0x2000) != 0
	// Bit 13 is spare
	m.Code = value & 0x0FFF

	return 2, nil
}

// Encode encodes the Mode-2 Code to bytes
func (m *Mode2Code) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	var value uint16
	if m.V {
		value |= 0x8000
	}
	if m.G {
		value |= 0x4000
	}
	if m.L {
		value |= 0x2000
	}
	// Bit 13 is spare (0)
	value |= m.Code & 0x0FFF

	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return 0, fmt.Errorf("writing mode-2 code: %w", err)
	}

	return 2, nil
}

// Validate validates the Mode-2 Code
func (m *Mode2Code) Validate() error {
	if m.Code > 0x0FFF {
		return fmt.Errorf("%w: code must be 0-4095, got %d", asterix.ErrInvalidMessage, m.Code)
	}
	return nil
}

// String returns a string representation
func (m *Mode2Code) String() string {
	status := ""
	if m.V {
		status += "V"
	}
	if m.G {
		status += "G"
	}
	if m.L {
		status += "L"
	}
	if status == "" {
		status = "OK"
	}
	return fmt.Sprintf("%04o (%s)", m.Code, status)
}
