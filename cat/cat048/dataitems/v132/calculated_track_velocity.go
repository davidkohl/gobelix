// dataitems/cat048/calculated_track_velocity.go
package v132

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedTrackVelocity implements I048/200
// Calculated track velocity expressed in polar co-ordinates.
type CalculatedTrackVelocity struct {
	GroundSpeed float64 // Ground speed (NM/s)
	Heading     float64 // Heading (degrees)
}

// Decode implements the DataItem interface
func (c *CalculatedTrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated track velocity: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("insufficient data for calculated track velocity: got %d bytes, want 4", n)
	}

	// Ground speed (16 bits), LSB = 2^-14 NM/s ≈ 0.00006 NM/s ≈ 0.22 kt
	speedRaw := uint16(data[0])<<8 | uint16(data[1])
	c.GroundSpeed = float64(speedRaw) * math.Pow(2, -14)

	// Heading (16 bits), LSB = 360/2^16 degrees ≈ 0.0055 degrees
	headingRaw := uint16(data[2])<<8 | uint16(data[3])
	c.Heading = float64(headingRaw) * (360.0 / 65536.0)

	return n, nil
}

// Encode implements the DataItem interface
func (c *CalculatedTrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert ground speed to raw value
	speedRaw := uint16(math.Round(c.GroundSpeed * math.Pow(2, 14)))

	// Handle potential overflow
	if c.GroundSpeed >= 4.0 { // Max value for 16 bits at this resolution
		speedRaw = 0xFFFF
	}

	// Convert heading to raw value, normalize to [0, 360) degrees
	heading := math.Mod(c.Heading, 360.0)
	if heading < 0 {
		heading += 360.0
	}
	headingRaw := uint16(math.Round(heading * (65536.0 / 360.0)))
	if heading >= 360.0-1e-10 { // Handle potential rounding for 360 degrees
		headingRaw = 0
	}

	data := make([]byte, 4)
	data[0] = byte(speedRaw >> 8)
	data[1] = byte(speedRaw)
	data[2] = byte(headingRaw >> 8)
	data[3] = byte(headingRaw)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated track velocity: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (c *CalculatedTrackVelocity) Validate() error {
	if c.GroundSpeed < 0 {
		return fmt.Errorf("negative ground speed not allowed: %f", c.GroundSpeed)
	}

	// The maximum value for 16 bits at resolution 2^-14 NM/s is (2^16-1)*2^-14 ≈ 4 NM/s ≈ 14400 kt
	if c.GroundSpeed > 4.0 {
		return fmt.Errorf("ground speed too large: %f NM/s (max 4 NM/s)", c.GroundSpeed)
	}

	return nil
}

// String returns a human-readable representation
func (c *CalculatedTrackVelocity) String() string {
	// Convert ground speed to knots for display (1 NM/s = 3600 knots)
	groundSpeedKt := c.GroundSpeed * 3600.0

	return fmt.Sprintf("%.1f kt / %.1f°", groundSpeedKt, c.Heading)
}

// SpeedInKnots returns the ground speed in knots
func (c *CalculatedTrackVelocity) SpeedInKnots() float64 {
	return c.GroundSpeed * 3600.0 // 1 NM/s = 3600 knots
}
