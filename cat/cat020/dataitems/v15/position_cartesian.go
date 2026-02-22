// cat/cat020/dataitems/v15/position_cartesian.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionCartesian implements I020/042 - Position in Cartesian Coordinates
// Per ASTERIX CAT020 spec: X and Y are each 24-bit signed values, LSB = 0.5m
// Total: 6 bytes (48 bits)
type PositionCartesian struct {
	X float64 // X coordinate in meters, LSB = 0.5 m, range ±4,194,300 m
	Y float64 // Y coordinate in meters, LSB = 0.5 m, range ±4,194,300 m
}

func (p *PositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	var data [6]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading cartesian position: %w", err)
	}
	if n != 6 {
		return n, fmt.Errorf("%w: need 6 bytes for cartesian position, have %d", asterix.ErrBufferTooShort, n)
	}

	// X: 3 bytes (24 bits), signed two's complement, LSB = 0.5 m
	xRaw := int32(data[0])<<16 | int32(data[1])<<8 | int32(data[2])
	// Sign extend from 24 bits to 32 bits
	if xRaw&0x800000 != 0 {
		xRaw |= ^0xFFFFFF // Sign extend
	}
	p.X = float64(xRaw) * 0.5

	// Y: 3 bytes (24 bits), signed two's complement, LSB = 0.5 m
	yRaw := int32(data[3])<<16 | int32(data[4])<<8 | int32(data[5])
	// Sign extend from 24 bits to 32 bits
	if yRaw&0x800000 != 0 {
		yRaw |= ^0xFFFFFF // Sign extend
	}
	p.Y = float64(yRaw) * 0.5

	return n, nil
}

func (p *PositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw 24-bit values
	xRaw := int32(p.X / 0.5)
	yRaw := int32(p.Y / 0.5)

	var data [6]byte

	// X: 3 bytes
	data[0] = byte(xRaw >> 16)
	data[1] = byte(xRaw >> 8)
	data[2] = byte(xRaw)

	// Y: 3 bytes
	data[3] = byte(yRaw >> 16)
	data[4] = byte(yRaw >> 8)
	data[5] = byte(yRaw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing cartesian position: %w", err)
	}
	return n, nil
}

func (p *PositionCartesian) Validate() error {
	// Valid range: ±4,194,300 m (based on 24-bit signed with LSB = 0.5m)
	const maxRange = 4194300.0
	if p.X < -maxRange || p.X > maxRange {
		return fmt.Errorf("X coordinate out of range [%.0f, %.0f]: %f", -maxRange, maxRange, p.X)
	}
	if p.Y < -maxRange || p.Y > maxRange {
		return fmt.Errorf("Y coordinate out of range [%.0f, %.0f]: %f", -maxRange, maxRange, p.Y)
	}
	return nil
}

func (p *PositionCartesian) String() string {
	return fmt.Sprintf("X: %.1fm, Y: %.1fm", p.X, p.Y)
}
