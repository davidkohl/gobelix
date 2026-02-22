// dataitems/cat021/airborne_ground_vector.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// AirborneGroundVector implements I021/160
// Airborne Ground Vector (4 octets)
type AirborneGroundVector struct {
	RE           bool    // Range Exceeded Indicator
	GroundSpeed  float64 // Ground speed in knots, LSB = 2^-14 NM/s
	TrackAngle   float64 // Track angle in degrees, LSB = 360/2^16
}

func (a *AirborneGroundVector) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	// Convert ground speed from knots to NM/s, then to raw value
	speedNMS := a.GroundSpeed / 3600.0 // Convert knots to NM/s
	rawSpeed := uint16(math.Round(speedNMS * math.Pow(2, 14)))

	// Mask to 15 bits
	rawSpeed &= 0x7FFF

	// Convert track angle to raw value: LSB = 360/2^16
	rawAngle := uint16(math.Round(a.TrackAngle * math.Pow(2, 16) / 360.0))

	var data [4]byte

	// First 2 octets: Ground Speed
	if a.RE {
		data[0] = 0x80 | byte(rawSpeed>>8)
	} else {
		data[0] = byte(rawSpeed >> 8)
	}
	data[1] = byte(rawSpeed)

	// Last 2 octets: Track Angle
	data[2] = byte(rawAngle >> 8)
	data[3] = byte(rawAngle)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing airborne ground vector: %w", err)
	}
	return n, nil
}

func (a *AirborneGroundVector) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading airborne ground vector: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for airborne ground vector, have %d", asterix.ErrBufferTooShort, n)
	}

	a.RE = (data[0] & 0x80) != 0

	// Extract ground speed (15 bits)
	rawSpeed := uint16(data[0]&0x7F)<<8 | uint16(data[1])

	// Convert to knots
	speedNMS := float64(rawSpeed) * math.Pow(2, -14) // NM/s
	a.GroundSpeed = speedNMS * 3600.0                // Convert to knots

	// Extract track angle (16 bits)
	rawAngle := uint16(data[2])<<8 | uint16(data[3])

	// Convert to degrees
	a.TrackAngle = float64(rawAngle) * 360.0 / math.Pow(2, 16)

	return n, a.Validate()
}

func (a *AirborneGroundVector) Validate() error {
	if a.GroundSpeed < 0 {
		return fmt.Errorf("ground speed cannot be negative: %.2f", a.GroundSpeed)
	}
	if a.TrackAngle < 0 || a.TrackAngle >= 360 {
		return fmt.Errorf("track angle out of range [0,360): %.2f", a.TrackAngle)
	}
	return nil
}

func (a *AirborneGroundVector) String() string {
	if a.RE {
		return fmt.Sprintf("GS: >%.1fkts, Track: %.1f°", a.GroundSpeed, a.TrackAngle)
	}
	return fmt.Sprintf("GS: %.1fkts, Track: %.1f°", a.GroundSpeed, a.TrackAngle)
}
