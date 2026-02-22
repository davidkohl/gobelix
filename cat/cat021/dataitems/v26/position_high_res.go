// dataitems/cat021/position_high_res.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// PositionHighRes implements I021/131
// Position in WGS-84 Coordinates, High Resolution (8 octets)
type PositionHighRes struct {
	Latitude  float64 // Latitude in degrees, LSB = 180/2^30
	Longitude float64 // Longitude in degrees, LSB = 180/2^30
}

func (p *PositionHighRes) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert latitude to raw value: LSB = 180/2^30
	latRaw := int32(math.Round(p.Latitude * math.Pow(2, 30) / 180.0))

	// Convert longitude to raw value: LSB = 180/2^30
	lonRaw := int32(math.Round(p.Longitude * math.Pow(2, 30) / 180.0))

	var data [8]byte

	// Latitude (4 octets)
	data[0] = byte(latRaw >> 24)
	data[1] = byte(latRaw >> 16)
	data[2] = byte(latRaw >> 8)
	data[3] = byte(latRaw)

	// Longitude (4 octets)
	data[4] = byte(lonRaw >> 24)
	data[5] = byte(lonRaw >> 16)
	data[6] = byte(lonRaw >> 8)
	data[7] = byte(lonRaw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing high-res position: %w", err)
	}
	return n, nil
}

func (p *PositionHighRes) Decode(buf *bytes.Buffer) (int, error) {
	var data [8]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading high-res position: %w", err)
	}
	if n != 8 {
		return n, fmt.Errorf("%w: need 8 bytes for high-res position, have %d", asterix.ErrBufferTooShort, n)
	}

	// Extract latitude (4 octets, signed)
	latRaw := int32(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))

	// Extract longitude (4 octets, signed)
	lonRaw := int32(uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7]))

	// Convert to degrees
	p.Latitude = float64(latRaw) * 180.0 / math.Pow(2, 30)
	p.Longitude = float64(lonRaw) * 180.0 / math.Pow(2, 30)

	return n, p.Validate()
}

func (p *PositionHighRes) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude out of range [-90,90]: %.8f", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude out of range [-180,180]: %.8f", p.Longitude)
	}
	return nil
}

func (p *PositionHighRes) String() string {
	return fmt.Sprintf("%.8f°, %.8f°", p.Latitude, p.Longitude)
}
