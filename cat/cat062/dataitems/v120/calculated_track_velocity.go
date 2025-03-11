// dataitems/cat062/calculated_track_velocity.go
package v120

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedTrackVelocity implements I062/185
// Calculated track velocity expressed in Cartesian co-ordinates
type CalculatedTrackVelocity struct {
	Vx float64 // X component of velocity in m/s
	Vy float64 // Y component of velocity in m/s
}

func (c *CalculatedTrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated track velocity: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for calculated track velocity: got %d bytes, want 4", n)
	}

	// Vx component: bytes 0-1, two's complement, LSB = 0.25 m/s
	vxRaw := int16(data[0])<<8 | int16(data[1])
	c.Vx = float64(vxRaw) * 0.25

	// Vy component: bytes 2-3, two's complement, LSB = 0.25 m/s
	vyRaw := int16(data[2])<<8 | int16(data[3])
	c.Vy = float64(vyRaw) * 0.25

	return n, nil
}

func (c *CalculatedTrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw values
	vxRaw := int16(math.Round(c.Vx / 0.25))
	vyRaw := int16(math.Round(c.Vy / 0.25))

	data := []byte{
		byte(vxRaw >> 8),
		byte(vxRaw),
		byte(vyRaw >> 8),
		byte(vyRaw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated track velocity: %w", err)
	}
	return n, nil
}

func (c *CalculatedTrackVelocity) Validate() error {
	// According to the spec, valid range is -8192 m/s to 8191.75 m/s
	if c.Vx < -8192 || c.Vx > 8191.75 {
		return fmt.Errorf("Vx component out of range [-8192,8191.75]: %f", c.Vx)
	}
	if c.Vy < -8192 || c.Vy > 8191.75 {
		return fmt.Errorf("Vy component out of range [-8192,8191.75]: %f", c.Vy)
	}
	return nil
}

func (c *CalculatedTrackVelocity) String() string {
	// Calculate ground speed and track angle
	groundSpeed := math.Sqrt(c.Vx*c.Vx + c.Vy*c.Vy)
	trackAngle := math.Atan2(c.Vx, c.Vy) * 180 / math.Pi
	if trackAngle < 0 {
		trackAngle += 360
	}

	// Convert to knots for display (1 m/s = 1.94384 knots)
	groundSpeedKt := groundSpeed * 1.94384

	return fmt.Sprintf("Velocity: %.1f kt / %.1f° (Vx: %.2f m/s, Vy: %.2f m/s)",
		groundSpeedKt, trackAngle, c.Vx, c.Vy)
}
