// cat/cat020/dataitems/v10/calculated_track_velocity.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CalculatedTrackVelocity represents I020/202 - Calculated Track Velocity in Cartesian Coordinates
// Fixed length: 4 bytes
// Calculated track velocity expressed in Cartesian Coordinates, in two's complement representation
type CalculatedTrackVelocity struct {
	Vx float64 // Velocity component along x-axis in m/s
	Vy float64 // Velocity component along y-axis in m/s
}

// NewCalculatedTrackVelocity creates a new Calculated Track Velocity data item
func NewCalculatedTrackVelocity() *CalculatedTrackVelocity {
	return &CalculatedTrackVelocity{}
}

// Decode decodes the Calculated Track Velocity from bytes
func (c *CalculatedTrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("%w: need 4 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(4)

	// Vx: 2 bytes, two's complement, LSB = 0.25 m/s
	vxRaw := int16(binary.BigEndian.Uint16(data[0:2]))
	c.Vx = float64(vxRaw) * 0.25

	// Vy: 2 bytes, two's complement, LSB = 0.25 m/s
	vyRaw := int16(binary.BigEndian.Uint16(data[2:4]))
	c.Vy = float64(vyRaw) * 0.25

	return 4, nil
}

// Encode encodes the Calculated Track Velocity to bytes
func (c *CalculatedTrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert Vx to raw value (LSB = 0.25 m/s)
	vxRaw := int16(c.Vx / 0.25)

	// Convert Vy to raw value (LSB = 0.25 m/s)
	vyRaw := int16(c.Vy / 0.25)

	// Write Vx (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, vxRaw); err != nil {
		return 0, fmt.Errorf("writing Vx: %w", err)
	}

	// Write Vy (2 bytes)
	if err := binary.Write(buf, binary.BigEndian, vyRaw); err != nil {
		return 2, fmt.Errorf("writing Vy: %w", err)
	}

	return 4, nil
}

// Validate validates the Calculated Track Velocity
func (c *CalculatedTrackVelocity) Validate() error {
	// Max range is ±8192 m/s (int16 max * 0.25)
	if c.Vx < -8192.0 || c.Vx > 8192.0 {
		return fmt.Errorf("%w: Vx must be in range [-8192, 8192] m/s, got %.2f", asterix.ErrInvalidMessage, c.Vx)
	}
	if c.Vy < -8192.0 || c.Vy > 8192.0 {
		return fmt.Errorf("%w: Vy must be in range [-8192, 8192] m/s, got %.2f", asterix.ErrInvalidMessage, c.Vy)
	}
	return nil
}

// String returns a string representation
func (c *CalculatedTrackVelocity) String() string {
	return fmt.Sprintf("Vx=%.2f m/s, Vy=%.2f m/s", c.Vx, c.Vy)
}
