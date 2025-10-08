// cat/cat020/dataitems/v10/geometric_altitude.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// GeometricAltitude represents I020/105 - Geometric Altitude (WGS-84)
// Fixed length: 2 bytes
// Vertical distance between the target and the projection of its position
// on the earth's ellipsoid, as defined by WGS84, in two's complement form
type GeometricAltitude struct {
	Altitude float64 // Geometric altitude in feet
}

// NewGeometricAltitude creates a new Geometric Altitude data item
func NewGeometricAltitude() *GeometricAltitude {
	return &GeometricAltitude{}
}

// Decode decodes the Geometric Altitude from bytes
func (g *GeometricAltitude) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// Two's complement, LSB = 6.25 ft
	altRaw := int16(binary.BigEndian.Uint16(data))
	g.Altitude = float64(altRaw) * 6.25

	return 2, nil
}

// Encode encodes the Geometric Altitude to bytes
func (g *GeometricAltitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := g.Validate(); err != nil {
		return 0, err
	}

	// Convert altitude to raw value (LSB = 6.25 ft)
	altRaw := int16(g.Altitude / 6.25)

	if err := binary.Write(buf, binary.BigEndian, altRaw); err != nil {
		return 0, fmt.Errorf("writing geometric altitude: %w", err)
	}

	return 2, nil
}

// Validate validates the Geometric Altitude
func (g *GeometricAltitude) Validate() error {
	// Range: ±204,800 ft (int16 range * 6.25)
	if g.Altitude < -204800.0 || g.Altitude > 204800.0 {
		return fmt.Errorf("%w: altitude must be in range [-204800, 204800] ft, got %.2f", asterix.ErrInvalidMessage, g.Altitude)
	}
	return nil
}

// String returns a string representation
func (g *GeometricAltitude) String() string {
	return fmt.Sprintf("%.2f ft", g.Altitude)
}
