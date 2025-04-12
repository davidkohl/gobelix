// dataitems/cat048/height3d.go
package v132

import (
	"bytes"
	"fmt"
)

// Height3D implements I048/110
// Height of a target as measured by a 3D radar.
// The height uses mean sea level as the zero reference level.
type Height3D struct {
	Height float64 // Height in feet
}

// Decode implements the DataItem interface
func (h *Height3D) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading 3D height: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for 3D height: got %d bytes, want 2", n)
	}

	// Extract height in two's complement format (bits 14-1)
	// Bits 16-15 are spare

	// Create a 16-bit value (removing the spare bits)
	rawValue := int16(uint16(data[0]&0x3F)<<8 | uint16(data[1]))

	// Sign extension is handled by the int16 cast

	// Convert to feet, LSB = 25 feet
	h.Height = float64(rawValue) * 25.0

	return n, nil
}

// Encode implements the DataItem interface
func (h *Height3D) Encode(buf *bytes.Buffer) (int, error) {
	if err := h.Validate(); err != nil {
		return 0, err
	}

	// Convert height to raw value
	rawHeight := int16(h.Height / 25.0)

	data := make([]byte, 2)

	// Set height bits (bits 14-1), bits 16-15 are spare
	data[0] |= byte((rawHeight >> 8) & 0x3F) // bits 14-9
	data[1] = byte(rawHeight)                // bits 8-1

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing 3D height: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (h *Height3D) Validate() error {
	// Check reasonable limits for aircraft altitude
	// The 14-bit two's complement range with LSB of 25 ft gives approximately ±200,000 ft
	if h.Height < -200000 || h.Height > 200000 {
		return fmt.Errorf("3D height out of reasonable range [-200000,200000]: %f", h.Height)
	}
	return nil
}

// String returns a human-readable representation
func (h *Height3D) String() string {
	return fmt.Sprintf("%.0f ft", h.Height)
}
