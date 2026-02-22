// dataitems/cat048/mode1_code.go
package v132

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode1Code implements I048/055
// Reply to Mode-1 interrogation.
type Mode1Code struct {
	V    bool  // Code validated
	G    bool  // Garbled code
	L    bool  // Mode-1 code derived/smoothed
	Code uint8 // Mode-1 code in octal (2 digits)
}

// Decode implements the DataItem interface
func (m *Mode1Code) Decode(buf *bytes.Buffer) (int, error) {
	var data [1]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading Mode-1 code: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("%w: need 1 byte for Mode-1 code, have %d", asterix.ErrBufferTooShort, n)
	}

	m.V = (data[0] & 0x80) != 0 // bit 8
	m.G = (data[0] & 0x40) != 0 // bit 7
	m.L = (data[0] & 0x20) != 0 // bit 6

	// Extract octal digits
	a := (data[0] & 0x1C) >> 2 // bits 5-3 (A)
	b := data[0] & 0x03        // bits 2-1 (B)

	// Combine digits into octal representation
	m.Code = uint8(a)*10 + uint8(b)

	return n, m.Validate()
}

// Encode implements the DataItem interface
func (m *Mode1Code) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Extract octal digits
	a := (m.Code / 10) % 10
	b := m.Code % 10

	var data [1]byte

	// Set flag bits
	if m.V {
		data[0] |= 0x80 // bit 8
	}
	if m.G {
		data[0] |= 0x40 // bit 7
	}
	if m.L {
		data[0] |= 0x20 // bit 6
	}

	// Set code bits
	data[0] |= byte(a&0x07) << 2 // bits 5-3 (A)
	data[0] |= byte(b & 0x03)    // bits 2-1 (B)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing Mode-1 code: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *Mode1Code) Validate() error {
	// Mode-1 code: A is 3 bits (0-7), B is 2 bits (0-3)
	// So valid codes are 00-03, 10-13, 20-23, 30-33, 40-43, 50-53, 60-63, 70-73
	a := (m.Code / 10) % 10
	b := m.Code % 10

	if a > 7 {
		return fmt.Errorf("invalid Mode-1 code first digit (must be 0-7): %02d", m.Code)
	}
	if b > 3 {
		return fmt.Errorf("invalid Mode-1 code second digit (must be 0-3): %02d", m.Code)
	}

	return nil
}

// String returns a human-readable representation
func (m *Mode1Code) String() string {
	flags := ""
	if m.V {
		flags += "V,"
	}
	if m.G {
		flags += "G,"
	}
	if m.L {
		flags += "L,"
	}

	if flags != "" {
		flags = flags[:len(flags)-1] + " " // Remove trailing comma
	}

	// Mode-1 is 5 bits (A: 3 bits, B: 2 bits), display as 2-digit code
	return fmt.Sprintf("%s%02d", flags, m.Code)
}
