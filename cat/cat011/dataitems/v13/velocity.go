package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TrackVelocityCartesian implements I011/202 - Calculated Track Velocity in Cartesian Co-ordinates
// Definition: Calculated track velocity expressed in Cartesian co-ordinates.
// Format: Four-octet fixed length Data Item
// LSB = 0.25 m/s, max range = ±8192 m/s
type TrackVelocityCartesian struct {
	Vx float64 // X-component velocity in m/s
	Vy float64 // Y-component velocity in m/s
}

const velocityLSB = 0.25 // m/s

func (v *TrackVelocityCartesian) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track velocity: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("track velocity: expected 4 bytes, got %d", n)
	}

	// Vx: bits 32-17 (first 2 bytes), two's complement
	vxRaw := int16(binary.BigEndian.Uint16(data[0:2]))
	v.Vx = float64(vxRaw) * velocityLSB

	// Vy: bits 16-1 (last 2 bytes), two's complement
	vyRaw := int16(binary.BigEndian.Uint16(data[2:4]))
	v.Vy = float64(vyRaw) * velocityLSB

	return 4, nil
}

func (v *TrackVelocityCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	var data [4]byte

	vxRaw := int16(v.Vx / velocityLSB)
	binary.BigEndian.PutUint16(data[0:2], uint16(vxRaw))

	vyRaw := int16(v.Vy / velocityLSB)
	binary.BigEndian.PutUint16(data[2:4], uint16(vyRaw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track velocity: %w", err)
	}
	return n, nil
}

func (v *TrackVelocityCartesian) Validate() error {
	maxVelocity := 8191.75 // (32767 * 0.25)
	if v.Vx < -maxVelocity || v.Vx > maxVelocity {
		return fmt.Errorf("Vx out of range: %f", v.Vx)
	}
	if v.Vy < -maxVelocity || v.Vy > maxVelocity {
		return fmt.Errorf("Vy out of range: %f", v.Vy)
	}
	return nil
}

func (v *TrackVelocityCartesian) String() string {
	return fmt.Sprintf("Velocity: Vx=%.2f m/s, Vy=%.2f m/s", v.Vx, v.Vy)
}
