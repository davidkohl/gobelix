// dataitems/cat062/calculated_acceleration.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedAcceleration implements I062/210
// Calculated Acceleration of the target expressed in Cartesian co-ordinates
type CalculatedAcceleration struct {
	Ax float64 // X component of acceleration in m/s²
	Ay float64 // Y component of acceleration in m/s²
}

func (c *CalculatedAcceleration) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated acceleration: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for calculated acceleration: got %d bytes, want 2", n)
	}

	// Ax: top byte, LSB = 0.25 m/s²
	ax := int8(data[0])
	c.Ax = float64(ax) * 0.25

	// Ay: bottom byte, LSB = 0.25 m/s²
	ay := int8(data[1])
	c.Ay = float64(ay) * 0.25

	return n, nil
}

func (c *CalculatedAcceleration) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw values, ensuring they fit in int8
	axRaw := int8(math.Round(c.Ax / 0.25))
	ayRaw := int8(math.Round(c.Ay / 0.25))

	data := []byte{
		byte(axRaw),
		byte(ayRaw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated acceleration: %w", err)
	}
	return n, nil
}

func (c *CalculatedAcceleration) Validate() error {
	// int8 range with LSB of 0.25 gives range of -32 to 31.75 m/s²
	if c.Ax < -32 || c.Ax > 31.75 {
		return fmt.Errorf("Ax component out of range [-32,31.75]: %f", c.Ax)
	}
	if c.Ay < -32 || c.Ay > 31.75 {
		return fmt.Errorf("Ay component out of range [-32,31.75]: %f", c.Ay)
	}
	return nil
}

func (c *CalculatedAcceleration) String() string {
	magnitude := math.Sqrt(c.Ax*c.Ax + c.Ay*c.Ay)
	direction := math.Atan2(c.Ax, c.Ay) * 180 / math.Pi
	if direction < 0 {
		direction += 360
	}

	return fmt.Sprintf("Acceleration: %.2f m/s² at %.1f° (Ax: %.2f m/s², Ay: %.2f m/s²)",
		magnitude, direction, c.Ax, c.Ay)
}
