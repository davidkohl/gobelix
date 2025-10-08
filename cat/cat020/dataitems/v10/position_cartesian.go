// cat/cat020/dataitems/v10/position_cartesian.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionCartesian represents I020/042 - Position in Cartesian Coordinates
// Fixed length: 6 bytes
// Calculated position in Cartesian Coordinates, in two's complement
type PositionCartesian struct {
	X float64 // X coordinate in meters
	Y float64 // Y coordinate in meters
}

// NewPositionCartesian creates a new Position Cartesian data item
func NewPositionCartesian() *PositionCartesian {
	return &PositionCartesian{}
}

// Decode decodes the Position Cartesian from bytes
func (p *PositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 6 {
		return 0, fmt.Errorf("%w: need 6 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(6)

	// X: 3 bytes, two's complement, LSB = 0.5 m
	xRaw := int32(data[0])<<16 | int32(data[1])<<8 | int32(data[2])
	// Sign extend from 24 bits to 32 bits
	if xRaw&0x800000 != 0 {
		xRaw |= ^0xFFFFFF // Sign extend with all 1s in upper bits
	}
	p.X = float64(xRaw) * 0.5

	// Y: 3 bytes, two's complement, LSB = 0.5 m
	yRaw := int32(data[3])<<16 | int32(data[4])<<8 | int32(data[5])
	// Sign extend from 24 bits to 32 bits
	if yRaw&0x800000 != 0 {
		yRaw |= ^0xFFFFFF // Sign extend with all 1s in upper bits
	}
	p.Y = float64(yRaw) * 0.5

	return 6, nil
}

// Encode encodes the Position Cartesian to bytes
func (p *PositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert X to raw value (LSB = 0.5 m)
	xRaw := int32(p.X / 0.5)
	// Mask to 24 bits
	xRaw &= 0xFFFFFF

	// Convert Y to raw value (LSB = 0.5 m)
	yRaw := int32(p.Y / 0.5)
	// Mask to 24 bits
	yRaw &= 0xFFFFFF

	// Write X (3 bytes)
	data := []byte{
		byte((xRaw >> 16) & 0xFF),
		byte((xRaw >> 8) & 0xFF),
		byte(xRaw & 0xFF),
		byte((yRaw >> 16) & 0xFF),
		byte((yRaw >> 8) & 0xFF),
		byte(yRaw & 0xFF),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing position cartesian: %w", err)
	}

	return 6, nil
}

// Validate validates the Position Cartesian
func (p *PositionCartesian) Validate() error {
	// Max range = +/- 4194.3 km (~2265 NM) with LSB 0.5m
	const maxRange = 4194300.0 // meters
	if p.X < -maxRange || p.X > maxRange {
		return fmt.Errorf("%w: X must be in range [%.1f, %.1f], got %.1f", asterix.ErrInvalidMessage, -maxRange, maxRange, p.X)
	}
	if p.Y < -maxRange || p.Y > maxRange {
		return fmt.Errorf("%w: Y must be in range [%.1f, %.1f], got %.1f", asterix.ErrInvalidMessage, -maxRange, maxRange, p.Y)
	}
	return nil
}

// String returns a string representation
func (p *PositionCartesian) String() string {
	return fmt.Sprintf("X=%.1fm, Y=%.1fm", p.X, p.Y)
}
