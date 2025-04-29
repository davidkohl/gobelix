// Package v26 implements the dataitems for ASTERIX Category 021 Version 2.6
package v26

import (
	"bytes"
	"fmt"
	"math"
)

// MagneticHeading implements I021/152
// This data item represents the magnetic heading of the aircraft in degrees
type MagneticHeading struct {
	Heading float64 // Heading in degrees
}

func (m *MagneticHeading) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading magnetic heading: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for magnetic heading: got %d bytes, want 2", n)
	}

	// Convert the bytes to heading value
	// Heading LSB = 360/2^16 = 0.0054931640625 degrees
	raw := uint16(data[0])<<8 | uint16(data[1])
	m.Heading = float64(raw) * (360.0 / 65536.0)

	return n, m.Validate()
}

func (m *MagneticHeading) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Convert heading to raw value (0 to 65535 representing 0 to 360 degrees)
	// Using the constant 65535.0 instead of 65536.0 to avoid potential overflow issues
	// when the heading is exactly 360.0 degrees
	rawValue := uint16(math.Round(m.Heading * (65535.0 / 360.0)))

	// Write the bytes
	b := make([]byte, 2)
	b[0] = byte(rawValue >> 8)
	b[1] = byte(rawValue)

	n, err := buf.Write(b)
	if err != nil {
		return n, fmt.Errorf("writing magnetic heading: %w", err)
	}
	return n, nil
}

func (m *MagneticHeading) Validate() error {
	// Heading should be between 0 and 360 degrees
	if m.Heading < 0 || m.Heading >= 360 {
		return fmt.Errorf("magnetic heading out of valid range [0,360): %f", m.Heading)
	}
	return nil
}

func (m *MagneticHeading) String() string {
	return fmt.Sprintf("%.2f°", m.Heading)
}
