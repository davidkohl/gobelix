// dataitems/cat048/mode1_code.go
package v132

import (
	"bytes"
	"fmt"
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
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode-1 code: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for Mode-1 code: got %d bytes, want 1", n)
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

	data := make([]byte, 1)

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

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode-1 code: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *Mode1Code) Validate() error {
	// Check that each digit is a valid octal digit (0-7)
	a := (m.Code / 10) % 10
	b := m.Code % 10

	if a > 7 || b > 7 {
		return fmt.Errorf("invalid octal digit in Mode-1 code: %02o", m.Code)
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

	return fmt.Sprintf("%s%02o", flags, m.Code)
}
