// cat/cat020/dataitems/v15/track_velocity_cartesian.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackVelocityCartesian implements I020/202 - Calculated Track Velocity in Cartesian Coordinates
type TrackVelocityCartesian struct {
	Vx float64 // Velocity X component in m/s, LSB = 0.25 m/s
	Vy float64 // Velocity Y component in m/s, LSB = 0.25 m/s
}

func (t *TrackVelocityCartesian) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track velocity cartesian: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for track velocity, have %d", asterix.ErrBufferTooShort, n)
	}

	// Vx: 2 bytes, signed, LSB = 0.25 m/s
	vxRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	t.Vx = float64(vxRaw) * 0.25

	// Vy: 2 bytes, signed, LSB = 0.25 m/s
	vyRaw := int16(uint16(data[2])<<8 | uint16(data[3]))
	t.Vy = float64(vyRaw) * 0.25

	return n, nil
}

func (t *TrackVelocityCartesian) Encode(buf *bytes.Buffer) (int, error) {
	// Convert to raw values
	vxRaw := int16(t.Vx / 0.25)
	vyRaw := int16(t.Vy / 0.25)

	var data [4]byte

	// Vx
	data[0] = byte(vxRaw >> 8)
	data[1] = byte(vxRaw)

	// Vy
	data[2] = byte(vyRaw >> 8)
	data[3] = byte(vyRaw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track velocity cartesian: %w", err)
	}
	return n, nil
}

func (t *TrackVelocityCartesian) Validate() error {
	return nil
}

func (t *TrackVelocityCartesian) String() string {
	return fmt.Sprintf("Vx: %.2f m/s, Vy: %.2f m/s", t.Vx, t.Vy)
}
