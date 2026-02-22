package v12

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackVelocity represents I001/200 - Calculated Track Velocity in Polar Coordinates
// Four-octet fixed length data item
type TrackVelocity struct {
	GroundSpeed float64 // Calculated ground speed in knots
	Heading     float64 // Calculated heading in degrees
}

func (t *TrackVelocity) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for track velocity, have %d", asterix.ErrBufferTooShort, n)
	}

	// Ground speed: 16 bits, unsigned, LSB = 2^-14 NM/s = 0.22 kt
	gsRaw := uint16(data[0])<<8 | uint16(data[1])
	t.GroundSpeed = float64(gsRaw) * 0.00006103515625 * 3600.0 // 2^-14 NM/s to kt

	// Heading: 16 bits, unsigned, LSB = 360°/2^16 = 0.0055°
	hdgRaw := uint16(data[2])<<8 | uint16(data[3])
	t.Heading = float64(hdgRaw) * 360.0 / 65536.0 // 360°/2^16

	return 4, nil
}

func (t *TrackVelocity) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Ground speed: 16 bits, unsigned, LSB = 2^-14 NM/s = 0.22 kt
	gsRaw := uint16(math.Round(t.GroundSpeed / (0.00006103515625 * 3600.0)))

	// Heading: 16 bits, unsigned, LSB = 360°/2^16
	hdgRaw := uint16(math.Round(t.Heading * 65536.0 / 360.0))

	data := []byte{
		byte(gsRaw >> 8),
		byte(gsRaw),
		byte(hdgRaw >> 8),
		byte(hdgRaw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track velocity: %w", err)
	}
	return n, nil
}

func (t *TrackVelocity) Validate() error {
	if t.GroundSpeed < 0 || t.GroundSpeed > 7200 {
		return fmt.Errorf("%w: ground speed out of range [0,7200]: %f", asterix.ErrInvalidMessage, t.GroundSpeed)
	}
	if t.Heading < 0 || t.Heading >= 360 {
		return fmt.Errorf("%w: heading out of range [0,360): %f", asterix.ErrInvalidMessage, t.Heading)
	}
	return nil
}

func (t *TrackVelocity) String() string {
	return fmt.Sprintf("Track Velocity: %.1f kt, Heading: %.1f°", t.GroundSpeed, t.Heading)
}
