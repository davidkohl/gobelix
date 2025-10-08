// cat/cat020/dataitems/v10/calculated_acceleration.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CalculatedAcceleration represents I020/210 - Calculated Acceleration
// Fixed length: 2 bytes
// Calculated Acceleration of the target, in two's complement form
type CalculatedAcceleration struct {
	Ax float64 // Acceleration component along x-axis in m/s²
	Ay float64 // Acceleration component along y-axis in m/s²
}

// NewCalculatedAcceleration creates a new Calculated Acceleration data item
func NewCalculatedAcceleration() *CalculatedAcceleration {
	return &CalculatedAcceleration{}
}

// Decode decodes the Calculated Acceleration from bytes
func (c *CalculatedAcceleration) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// Ax: 1 byte, two's complement, LSB = 0.25 m/s²
	axRaw := int8(data[0])
	c.Ax = float64(axRaw) * 0.25

	// Ay: 1 byte, two's complement, LSB = 0.25 m/s²
	ayRaw := int8(data[1])
	c.Ay = float64(ayRaw) * 0.25

	return 2, nil
}

// Encode encodes the Calculated Acceleration to bytes
func (c *CalculatedAcceleration) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert Ax to raw value (LSB = 0.25 m/s²)
	axRaw := int8(c.Ax / 0.25)

	// Convert Ay to raw value (LSB = 0.25 m/s²)
	ayRaw := int8(c.Ay / 0.25)

	// Write Ax (1 byte)
	if err := buf.WriteByte(byte(axRaw)); err != nil {
		return 0, fmt.Errorf("writing Ax: %w", err)
	}

	// Write Ay (1 byte)
	if err := buf.WriteByte(byte(ayRaw)); err != nil {
		return 1, fmt.Errorf("writing Ay: %w", err)
	}

	return 2, nil
}

// Validate validates the Calculated Acceleration
func (c *CalculatedAcceleration) Validate() error {
	// Max range is ±31 m/s² (int8 range * 0.25)
	if c.Ax < -31.0 || c.Ax > 31.0 {
		return fmt.Errorf("%w: Ax must be in range [-31, 31] m/s², got %.2f", asterix.ErrInvalidMessage, c.Ax)
	}
	if c.Ay < -31.0 || c.Ay > 31.0 {
		return fmt.Errorf("%w: Ay must be in range [-31, 31] m/s², got %.2f", asterix.ErrInvalidMessage, c.Ay)
	}
	return nil
}

// String returns a string representation
func (c *CalculatedAcceleration) String() string {
	return fmt.Sprintf("Ax=%.2f m/s², Ay=%.2f m/s²", c.Ax, c.Ay)
}
