package v12

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionCartesian represents I001/042 - Calculated Position in Cartesian Coordinates
// Four-octet fixed length data item
type PositionCartesian struct {
	X float64 // X-component in NM
	Y float64 // Y-component in NM
}

func (p *PositionCartesian) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for position in cartesian coordinates, have %d", asterix.ErrBufferTooShort, n)
	}

	// X-component: 16 bits, signed, LSB = 1/64 NM (2^-6)
	xRaw := int16(uint16(data[0])<<8 | uint16(data[1]))
	p.X = float64(xRaw) / 64.0 // LSB = 1/64 NM

	// Y-component: 16 bits, signed, LSB = 1/64 NM (2^-6)
	yRaw := int16(uint16(data[2])<<8 | uint16(data[3]))
	p.Y = float64(yRaw) / 64.0 // LSB = 1/64 NM

	return 4, nil
}

func (p *PositionCartesian) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// X-component: 16 bits, signed, LSB = 1/64 NM
	xRaw := int16(math.Round(p.X * 64.0))
	// Y-component: 16 bits, signed, LSB = 1/64 NM
	yRaw := int16(math.Round(p.Y * 64.0))

	data := []byte{
		byte(xRaw >> 8),
		byte(xRaw),
		byte(yRaw >> 8),
		byte(yRaw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing position in cartesian coordinates: %w", err)
	}
	return n, nil
}

func (p *PositionCartesian) Validate() error {
	// Max range = 2^9 NM = 512 NM (with default f=0)
	maxRange := 512.0
	if math.Abs(p.X) > maxRange {
		return fmt.Errorf("%w: X coordinate out of range [-512,512]: %f", asterix.ErrInvalidMessage, p.X)
	}
	if math.Abs(p.Y) > maxRange {
		return fmt.Errorf("%w: Y coordinate out of range [-512,512]: %f", asterix.ErrInvalidMessage, p.Y)
	}
	return nil
}

func (p *PositionCartesian) String() string {
	return fmt.Sprintf("Position (Cartesian): X=%.3f NM, Y=%.3f NM", p.X, p.Y)
}
