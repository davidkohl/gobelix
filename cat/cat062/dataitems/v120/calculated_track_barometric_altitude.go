// dataitems/cat062/calculated_track_barometric_altitude.go
package v120

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CalculatedTrackBarometricAltitude implements I062/135
// Calculated Barometric Altitude of the track
type CalculatedTrackBarometricAltitude struct {
	Altitude float64 // Altitude in flight levels (1 FL = 100 ft)
	QNH      bool    // Whether QNH correction has been applied
}

func (c *CalculatedTrackBarometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading calculated track barometric altitude: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for calculated track barometric altitude, have %d", asterix.ErrBufferTooShort, n)
	}

	// Check QNH bit (bit 16)
	c.QNH = (data[0] & 0x80) != 0

	// Altitude in two's complement form (15 bits: bits 15-1), LSB = 1/4 FL = 25 ft
	// Extract 15-bit value
	raw := int16((uint16(data[0]&0x7F) << 8) | uint16(data[1]))

	// Sign extend from 15 bits to 16 bits
	if (raw & 0x4000) != 0 { // Check bit 14 (sign bit of 15-bit value)
		raw |= ^0x7FFF // Set upper bits for negative
	}
	c.Altitude = float64(raw) * 0.25

	return n, nil
}

func (c *CalculatedTrackBarometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value (15-bit two's complement)
	raw := int16(c.Altitude / 0.25)

	// Mask to 15 bits and set QNH bit
	rawUnsigned := uint16(raw) & 0x7FFF

	data := []byte{
		byte(rawUnsigned >> 8),
		byte(rawUnsigned),
	}
	if c.QNH {
		data[0] |= 0x80 // Set QNH bit (bit 16)
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
