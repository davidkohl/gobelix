// dataitems/cat062/calculated_track_geometric_altitude.go
package v117

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CalculatedTrackGeometricAltitude implements I062/130
// Vertical distance between the target and the projection of its position on the earth's ellipsoid
type CalculatedTrackGeometricAltitude struct {
	Altitude float64 // Altitude in feet
}

func (c *CalculatedTrackGeometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading calculated track geometric altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for calculated track geometric altitude, have %d", asterix.ErrBufferTooShort, n)
	}

	// Altitude in two's complement form, LSB = 6.25 feet
	// Convert to signed 16-bit value
	raw := int16(data[0])<<8 | int16(data[1])
	c.Altitude = float64(raw) * 6.25

	return n, nil
}

func (c *CalculatedTrackGeometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value
	raw := int16(c.Altitude / 6.25)

	data := []byte{
		byte(raw >> 8),
		byte(raw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated track geometric altitude: %w", err)
	}
	return n, nil
}

func (c *CalculatedTrackGeometricAltitude) Validate() error {
	// According to the spec, valid range is -1500 ft to 150000 ft
	if c.Altitude < -1500 || c.Altitude > 150000 {
		return fmt.Errorf("geometric altitude out of range [-1500,150000]: %f", c.Altitude)
	}
	return nil
}

func (c *CalculatedTrackGeometricAltitude) String() string {
	return fmt.Sprintf("Geometric Altitude: %.2f ft", c.Altitude)
}
