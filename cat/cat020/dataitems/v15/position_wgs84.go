// cat/cat020/dataitems/v15/position_wgs84.go
package v15

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionWGS84 implements I020/041 - Position in WGS-84 Coordinates
type PositionWGS84 struct {
	Latitude  float64 // Latitude in degrees, LSB = 180/2^25
	Longitude float64 // Longitude in degrees, LSB = 180/2^25
}

func (p *PositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	var data [8]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading WGS-84 position: %w", err)
	}
	if n != 8 {
		return n, fmt.Errorf("%w: need 8 bytes for WGS-84 position, have %d", asterix.ErrBufferTooShort, n)
	}

	// Latitude: 4 bytes, signed, LSB = 180/2^25
	latRaw := int32(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
	p.Latitude = float64(latRaw) * 180.0 / math.Pow(2, 25)

	// Longitude: 4 bytes, signed, LSB = 180/2^25
	lonRaw := int32(uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7]))
	p.Longitude = float64(lonRaw) * 180.0 / math.Pow(2, 25)

	return n, p.Validate()
}

func (p *PositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw values
	latRaw := int32(math.Round(p.Latitude * math.Pow(2, 25) / 180.0))
	lonRaw := int32(math.Round(p.Longitude * math.Pow(2, 25) / 180.0))

	var data [8]byte

	// Latitude
	data[0] = byte(latRaw >> 24)
	data[1] = byte(latRaw >> 16)
	data[2] = byte(latRaw >> 8)
	data[3] = byte(latRaw)

	// Longitude
	data[4] = byte(lonRaw >> 24)
	data[5] = byte(lonRaw >> 16)
	data[6] = byte(lonRaw >> 8)
	data[7] = byte(lonRaw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing WGS-84 position: %w", err)
	}
	return n, nil
}

func (p *PositionWGS84) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude out of range: %.6f", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude out of range: %.6f", p.Longitude)
	}
	return nil
}

func (p *PositionWGS84) String() string {
	return fmt.Sprintf("Lat: %.6f°, Lon: %.6f°", p.Latitude, p.Longitude)
}
