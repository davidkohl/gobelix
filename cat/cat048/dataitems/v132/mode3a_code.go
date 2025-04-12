// dataitems/cat048/mode3a_code.go
package v132

import (
	"bytes"
	"fmt"
)

// Mode3ACode implements I048/070
// Mode-3/A code converted into octal representation.
type Mode3ACode struct {
	V    bool   // Code validated
	G    bool   // Garbled code
	L    bool   // Mode-3/A code derived/not extracted
	Code uint16 // Mode-3/A reply in octal representation
}

// Decode implements the DataItem interface
func (m *Mode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode-3/A code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for Mode-3/A code: got %d bytes, want 2", n)
	}

	m.V = (data[0] & 0x80) != 0 // bit 16
	m.G = (data[0] & 0x40) != 0 // bit 15
	m.L = (data[0] & 0x20) != 0 // bit 14
	// bit 13 is spare

	// Extract octal digits
	a := (data[0] & 0x0E) >> 1                             // bits 12-10 (A)
	b := ((data[0] & 0x01) << 2) | ((data[1] & 0xC0) >> 6) // bits 9-7 (B)
	c := (data[1] & 0x38) >> 3                             // bits 6-4 (C)
	d := data[1] & 0x07                                    // bits 3-1 (D)

	// Combine digits into octal representation
	m.Code = uint16(a)*1000 + uint16(b)*100 + uint16(c)*10 + uint16(d)

	return n, m.Validate()
}

// Encode implements the DataItem interface
func (m *Mode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Extract octal digits
	a := (m.Code / 1000) % 10
	b := (m.Code / 100) % 10
	c := (m.Code / 10) % 10
	d := m.Code % 10

	data := make([]byte, 2)

	// Set flag bits
	if m.V {
		data[0] |= 0x80 // bit 16
	}
	if m.G {
		data[0] |= 0x40 // bit 15
	}
	if m.L {
		data[0] |= 0x20 // bit 14
	}
	// bit 13 is spare (set to 0)

	// Set code bits
	data[0] |= byte(a) << 1      // bits 12-10 (A)
	data[0] |= byte(b>>2) & 0x01 // bit 9 (part of B)
	data[1] |= byte(b&0x03) << 6 // bits 8-7 (rest of B)
	data[1] |= byte(c&0x07) << 3 // bits 6-4 (C)
	data[1] |= byte(d & 0x07)    // bits 3-1 (D)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode-3/A code: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (m *Mode3ACode) Validate() error {
	// Check that each digit is a valid octal digit (0-7)
	a := (m.Code / 1000) % 10
	b := (m.Code / 100) % 10
	c := (m.Code / 10) % 10
	d := m.Code % 10

	if a > 7 || b > 7 || c > 7 || d > 7 {
		return fmt.Errorf("invalid octal digit in Mode-3/A code: %04o", m.Code)
	}

	return nil
}

// String returns a human-readable representation
func (m *Mode3ACode) String() string {
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

	return fmt.Sprintf("%s%04o", flags, m.Code)
}
