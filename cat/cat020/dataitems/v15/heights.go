// cat/cat020/dataitems/v15/heights.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MeasuredHeight implements I020/110 - Measured Height (3D Radar)
type MeasuredHeight struct {
	Height int16 // Height in feet, LSB = 25 ft, signed
}

func (m *MeasuredHeight) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading measured height: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for measured height, have %d", asterix.ErrBufferTooShort, n)
	}

	// 14-bit signed value, LSB = 25 ft
	rawHeight := int16((uint16(data[0]&0x3F) << 8) | uint16(data[1]))

	// Sign extend from 14 bits
	if (rawHeight & 0x2000) != 0 {
		rawHeight |= ^0x3FFF
	}

	m.Height = rawHeight * 25

	return n, nil
}

func (m *MeasuredHeight) Encode(buf *bytes.Buffer) (int, error) {
	rawHeight := m.Height / 25
	rawUnsigned := uint16(rawHeight) & 0x3FFF

	var data [2]byte
	data[0] = byte((rawUnsigned >> 8) & 0x3F)
	data[1] = byte(rawUnsigned & 0xFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing measured height: %w", err)
	}
	return n, nil
}

func (m *MeasuredHeight) Validate() error {
	return nil
}

func (m *MeasuredHeight) String() string {
	return fmt.Sprintf("%d ft", m.Height)
}

// GeometricHeight implements I020/105 - Geometric Height (WGS-84)
type GeometricHeight struct {
	Height float64 // Height in feet, LSB = 6.25 ft, signed
}

func (g *GeometricHeight) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading geometric height: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for geometric height, have %d", asterix.ErrBufferTooShort, n)
	}

	// 16-bit signed value, LSB = 6.25 ft
	// Range: -32768 * 6.25 = -204,800 ft to 32767 * 6.25 = 204,793.75 ft
	rawHeight := int16(uint16(data[0])<<8 | uint16(data[1]))
	g.Height = float64(rawHeight) * 6.25

	return n, nil
}

func (g *GeometricHeight) Encode(buf *bytes.Buffer) (int, error) {
	rawHeight := int16(g.Height / 6.25)

	var data [2]byte
	data[0] = byte(rawHeight >> 8)
	data[1] = byte(rawHeight)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing geometric height: %w", err)
	}
	return n, nil
}

func (g *GeometricHeight) Validate() error {
	// Valid range based on 16-bit signed value with LSB = 6.25 ft
	if g.Height < -204800 || g.Height > 204793.75 {
		return fmt.Errorf("geometric height out of range [-204800, 204793.75]: %f", g.Height)
	}
	return nil
}

func (g *GeometricHeight) String() string {
	return fmt.Sprintf("%.0f ft (WGS-84)", g.Height)
}
