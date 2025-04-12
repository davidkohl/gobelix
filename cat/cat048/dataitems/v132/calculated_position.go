// dataitems/cat048/calculated_position.go
package v132

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedPosition implements I048/042
// Calculated position of an aircraft in Cartesian co-ordinates.
type CalculatedPosition struct {
	X float64 // X-component (NM)
	Y float64 // Y-component (NM)
}

// Decode implements the DataItem interface
func (c *CalculatedPosition) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated position: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for calculated position: got %d bytes, want 4", n)
	}

	// X-Component (16 bits): in two's complement, LSB = 1/128 NM
	xRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	c.X = float64(xRaw) / 128.0

	// Y-Component (16 bits): in two's complement, LSB = 1/128 NM
	yRaw := int16(uint16(data[2])<<8 | uint16(data[3]))
	c.Y = float64(yRaw) / 128.0

	return n, nil
}

// Encode implements the DataItem interface
func (c *CalculatedPosition) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert X to raw value
	xRaw := int16(math.Round(c.X * 128.0))
	if c.X >= 256.0 {
		xRaw = 32767 // Max positive value
	} else if c.X <= -256.0 {
		xRaw = -32768 // Min negative value
	}

	// Convert Y to raw value
	yRaw := int16(math.Round(c.Y * 128.0))
	if c.Y >= 256.0 {
		yRaw = 32767 // Max positive value
	} else if c.Y <= -256.0 {
		yRaw = -32768 // Min negative value
	}

	data := make([]byte, 4)
	data[0] = byte(xRaw >> 8)
	data[1] = byte(xRaw)
	data[2] = byte(yRaw >> 8)
	data[3] = byte(yRaw)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated position: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (c *CalculatedPosition) Validate() error {
	if c.X < -256.0 || c.X > 256.0 {
		return fmt.Errorf("the X-component out of range [-256,256]: %f", c.X)
	}
	if c.Y < -256.0 || c.Y > 256.0 {
		return fmt.Errorf("the Y-component out of range [-256,256]: %f", c.Y)
	}
	return nil
}

// String returns a human-readable representation
func (c *CalculatedPosition) String() string {
	return fmt.Sprintf("X: %.3f NM, Y: %.3f NM", c.X, c.Y)
}
