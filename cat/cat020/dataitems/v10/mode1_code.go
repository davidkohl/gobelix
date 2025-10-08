// cat/cat020/dataitems/v10/mode1_code.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode1Code represents I020/055 - Mode-1 Code in Octal Representation
// Fixed length: 1 byte
// Mode-1 code converted into octal representation
type Mode1Code struct {
	V    bool  // Code validated (false) / Code not validated (true)
	G    bool  // Default (false) / Garbled code (true)
	L    bool  // Mode-1 code derived from transponder (false) / Smoothed code (true)
	Code uint8 // Mode-1 code in octal representation (5 bits)
}

// NewMode1Code creates a new Mode-1 Code data item
func NewMode1Code() *Mode1Code {
	return &Mode1Code{}
}

// Decode decodes the Mode-1 Code from bytes
func (m *Mode1Code) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)

	m.V = (data[0] & 0x80) != 0
	m.G = (data[0] & 0x40) != 0
	m.L = (data[0] & 0x20) != 0
	m.Code = data[0] & 0x1F

	return 1, nil
}

// Encode encodes the Mode-1 Code to bytes
func (m *Mode1Code) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	var value byte
	if m.V {
		value |= 0x80
	}
	if m.G {
		value |= 0x40
	}
	if m.L {
		value |= 0x20
	}
	value |= m.Code & 0x1F

	if err := buf.WriteByte(value); err != nil {
		return 0, fmt.Errorf("writing mode-1 code: %w", err)
	}

	return 1, nil
}

// Validate validates the Mode-1 Code
func (m *Mode1Code) Validate() error {
	if m.Code > 31 {
		return fmt.Errorf("%w: code must be 0-31, got %d", asterix.ErrInvalidMessage, m.Code)
	}
	return nil
}

// String returns a string representation
func (m *Mode1Code) String() string {
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
	return fmt.Sprintf("%02o (%s)", m.Code, status)
}
