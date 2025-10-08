// cat/cat020/dataitems/v10/measured_height.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MeasuredHeight represents I020/110 - Measured Height (Local Cartesian Coordinates)
// Fixed length: 2 bytes
// Height above local 2D coordinate system, in two's complement form,
// based on a direct measurement not related to barometric pressure
type MeasuredHeight struct {
	Height float64 // Height in feet
}

// NewMeasuredHeight creates a new Measured Height data item
func NewMeasuredHeight() *MeasuredHeight {
	return &MeasuredHeight{}
}

// Decode decodes the Measured Height from bytes
func (m *MeasuredHeight) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// Two's complement, LSB = 6.25 ft
	heightRaw := int16(binary.BigEndian.Uint16(data))
	m.Height = float64(heightRaw) * 6.25

	return 2, nil
}

// Encode encodes the Measured Height to bytes
func (m *MeasuredHeight) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// Convert height to raw value (LSB = 6.25 ft)
	heightRaw := int16(m.Height / 6.25)

	if err := binary.Write(buf, binary.BigEndian, heightRaw); err != nil {
		return 0, fmt.Errorf("writing measured height: %w", err)
	}

	return 2, nil
}

// Validate validates the Measured Height
func (m *MeasuredHeight) Validate() error {
	// Range: ±204,800 ft (int16 range * 6.25)
	if m.Height < -204800.0 || m.Height > 204800.0 {
		return fmt.Errorf("%w: height must be in range [-204800, 204800] ft, got %.2f", asterix.ErrInvalidMessage, m.Height)
	}
	return nil
}

// String returns a string representation
func (m *MeasuredHeight) String() string {
	return fmt.Sprintf("%.2f ft", m.Height)
}
