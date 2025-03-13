// dataitems/cat062/calculated_track_velocity.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedTrackVelocity implements I062/185
// Calculated track velocity expressed in Cartesian coordinates (Vx, Vy)
type CalculatedTrackVelocity struct {
	// Vx - Velocity component in x direction (east/west), in meters per second
	// Positive values indicate eastward movement
	Vx float64

	// Vy - Velocity component in y direction (north/south), in meters per second
	// Positive values indicate northward movement
	Vy float64
}

// Decode parses an ASTERIX Category 062 I185 data item from the buffer
func (v *CalculatedTrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("buffer too short for calculated track velocity (need 4 bytes)")
	}

	data := make([]byte, 4)
	n, err := buf.Read(data)
	if err != nil || n != 4 {
		return n, fmt.Errorf("reading calculated track velocity: %w", err)
	}

	// Extract Vx (16 bits) as a signed value
	vxBits := uint16(data[0])<<8 | uint16(data[1])

	// Convert to signed int16 (two's complement)
	var vxValue int16
	if (vxBits & 0x8000) != 0 {
		// Negative value (two's complement)
		vxValue = -int16(^vxBits + 1)
	} else {
		// Positive value
		vxValue = int16(vxBits)
	}

	// Calculate Vx in meters per second
	// LSB = 0.25 m/s
	v.Vx = float64(vxValue) * 0.25

	// Extract Vy (16 bits) as a signed value
	vyBits := uint16(data[2])<<8 | uint16(data[3])

	// Convert to signed int16 (two's complement)
	var vyValue int16
	if (vyBits & 0x8000) != 0 {
		// Negative value (two's complement)
		vyValue = -int16(^vyBits + 1)
	} else {
		// Positive value
		vyValue = int16(vyBits)
	}

	// Calculate Vy in meters per second
	// LSB = 0.25 m/s
	v.Vy = float64(vyValue) * 0.25

	return n, nil
}

// Encode serializes the calculated track velocity into the buffer
func (v *CalculatedTrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 4)

	// Convert Vx from m/s to binary representation
	// LSB = 0.25 m/s
	vxValue := int16(math.Round(v.Vx / 0.25))

	// Handle two's complement for negative values
	var vxBits uint16
	if vxValue < 0 {
		vxBits = uint16(^(-vxValue) + 1) // Two's complement for negative values
	} else {
		vxBits = uint16(vxValue)
	}

	// Store Vx in first 2 bytes
	data[0] = byte(vxBits >> 8)
	data[1] = byte(vxBits)

	// Convert Vy from m/s to binary representation
	// LSB = 0.25 m/s
	vyValue := int16(math.Round(v.Vy / 0.25))

	// Handle two's complement for negative values
	var vyBits uint16
	if vyValue < 0 {
		vyBits = uint16(^(-vyValue) + 1) // Two's complement for negative values
	} else {
		vyBits = uint16(vyValue)
	}

	// Store Vy in last 2 bytes
	data[2] = byte(vyBits >> 8)
	data[3] = byte(vyBits)

	return buf.Write(data)
}

// String returns a human-readable representation of the calculated track velocity
func (v *CalculatedTrackVelocity) String() string {
	return fmt.Sprintf("Vx: %.2f m/s, Vy: %.2f m/s", v.Vx, v.Vy)
}

// Validate performs validation on the calculated track velocity
func (v *CalculatedTrackVelocity) Validate() error {
	// Check Vx is within valid range [-8192, 8191.75] m/s
	if v.Vx < -8192 || v.Vx > 8191.75 {
		return fmt.Errorf("Vx out of range [-8192,8191.75]: %f", v.Vx)
	}

	// Check Vy is within valid range [-8192, 8191.75] m/s
	if v.Vy < -8192 || v.Vy > 8191.75 {
		return fmt.Errorf("Vy out of range [-8192,8191.75]: %f", v.Vy)
	}

	return nil
}
