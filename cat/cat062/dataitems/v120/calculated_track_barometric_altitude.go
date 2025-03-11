// dataitems/cat062/calculated_track_barometric_altitude.go
package v120

import (
	"bytes"
	"fmt"
)

// CalculatedTrackBarometricAltitude implements I062/135
// Calculated Barometric Altitude of the track
type CalculatedTrackBarometricAltitude struct {
	Altitude float64 // Altitude in flight levels (1 FL = 100 ft)
	QNH      bool    // Whether QNH correction has been applied
}

func (c *CalculatedTrackBarometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading calculated track barometric altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for calculated track barometric altitude: got %d bytes, want 2", n)
	}

	// Check QNH bit
	c.QNH = (data[0] & 0x80) != 0

	// Altitude in two's complement form, with QNH bit cleared, LSB = 1/4 FL = 25 ft
	raw := int16((uint16(data[0]&0x7F) << 8) | uint16(data[1]))
	c.Altitude = float64(raw) * 0.25

	return n, nil
}

func (c *CalculatedTrackBarometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value
	raw := int16(c.Altitude / 0.25)

	// Prepare first byte with QNH bit if needed
	firstByte := byte(raw >> 8)
	if c.QNH {
		firstByte |= 0x80
	}

	data := []byte{
		firstByte,
		byte(raw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing calculated track barometric altitude: %w", err)
	}
	return n, nil
}

func (c *CalculatedTrackBarometricAltitude) Validate() error {
	// According to the spec, valid range is -15 FL to 1500 FL
	if c.Altitude < -15 || c.Altitude > 1500 {
		return fmt.Errorf("barometric altitude out of range [-15,1500]: %f", c.Altitude)
	}
	return nil
}

func (c *CalculatedTrackBarometricAltitude) String() string {
	qnhInfo := ""
	if c.QNH {
		qnhInfo = " (QNH corrected)"
	}
	return fmt.Sprintf("Barometric Altitude: FL %.2f%s", c.Altitude, qnhInfo)
}
