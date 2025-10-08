// cat/cat020/dataitems/v10/position_wgs84.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionWGS84 represents I020/041 - Position in WGS-84 Coordinates
// Fixed length: 8 bytes
// Position of a target in WGS-84 Coordinates
type PositionWGS84 struct {
	Latitude  float64 // Latitude in degrees, range -90 to +90
	Longitude float64 // Longitude in degrees, range -180 to +180
}

// NewPositionWGS84 creates a new Position WGS-84 data item
func NewPositionWGS84() *PositionWGS84 {
	return &PositionWGS84{}
}

// Decode decodes the Position WGS-84 from bytes
func (p *PositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 8 {
		return 0, fmt.Errorf("%w: need 8 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(8)

	// Latitude: 4 bytes, two's complement, LSB = 180/2^25 degrees
	latRaw := int32(binary.BigEndian.Uint32(data[0:4]))
	p.Latitude = float64(latRaw) * 180.0 / math.Pow(2, 25)

	// Longitude: 4 bytes, two's complement, LSB = 180/2^25 degrees
	lonRaw := int32(binary.BigEndian.Uint32(data[4:8]))
	p.Longitude = float64(lonRaw) * 180.0 / math.Pow(2, 25)

	return 8, nil
}

// Encode encodes the Position WGS-84 to bytes
func (p *PositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert latitude to raw value
	latRaw := int32(math.Round(p.Latitude * math.Pow(2, 25) / 180.0))

	// Convert longitude to raw value
	lonRaw := int32(math.Round(p.Longitude * math.Pow(2, 25) / 180.0))

	// Write latitude (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, latRaw); err != nil {
		return 0, fmt.Errorf("writing latitude: %w", err)
	}

	// Write longitude (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, lonRaw); err != nil {
		return 4, fmt.Errorf("writing longitude: %w", err)
	}

	return 8, nil
}

// Validate validates the Position WGS-84
func (p *PositionWGS84) Validate() error {
	if p.Latitude < -90.0 || p.Latitude > 90.0 {
		return fmt.Errorf("%w: latitude must be in range [-90, 90], got %.6f", asterix.ErrInvalidMessage, p.Latitude)
	}
	if p.Longitude < -180.0 || p.Longitude >= 180.0 {
		return fmt.Errorf("%w: longitude must be in range [-180, 180), got %.6f", asterix.ErrInvalidMessage, p.Longitude)
	}
	return nil
}

// String returns a string representation
func (p *PositionWGS84) String() string {
	return fmt.Sprintf("Lat=%.6f°, Lon=%.6f°", p.Latitude, p.Longitude)
}
