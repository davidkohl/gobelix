package v13

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PositionCartesian implements I011/042 - Calculated Position in Cartesian Co-ordinates
// Definition: Calculated position of a target in Cartesian co-ordinates (two's complement form).
// Format: Four-octet fixed length Data Item
// LSB = 1 meter, max range = ±32768m (approx. ±17.7 NM)
type PositionCartesian struct {
	X float64 // X-component in meters
	Y float64 // Y-component in meters
}

func (p *PositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading position cartesian: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("position cartesian: expected 4 bytes, got %d", n)
	}

	// X-component: bits 32-17 (first 2 bytes), two's complement, LSB = 1m
	xRaw := int16(binary.BigEndian.Uint16(data[0:2]))
	p.X = float64(xRaw)

	// Y-component: bits 16-1 (last 2 bytes), two's complement, LSB = 1m
	yRaw := int16(binary.BigEndian.Uint16(data[2:4]))
	p.Y = float64(yRaw)

	return 4, nil
}

func (p *PositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	var data [4]byte

	xRaw := int16(p.X)
	binary.BigEndian.PutUint16(data[0:2], uint16(xRaw))

	yRaw := int16(p.Y)
	binary.BigEndian.PutUint16(data[2:4], uint16(yRaw))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing position cartesian: %w", err)
	}
	return n, nil
}

func (p *PositionCartesian) Validate() error {
	if p.X < -32768 || p.X > 32767 {
		return fmt.Errorf("X out of range: %f (expected -32768 to 32767)", p.X)
	}
	if p.Y < -32768 || p.Y > 32767 {
		return fmt.Errorf("Y out of range: %f (expected -32768 to 32767)", p.Y)
	}
	return nil
}

func (p *PositionCartesian) String() string {
	return fmt.Sprintf("Position Cartesian: X=%.0fm, Y=%.0fm", p.X, p.Y)
}
