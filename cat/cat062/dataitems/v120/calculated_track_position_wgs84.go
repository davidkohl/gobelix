// dataitems/cat062/calculated_position_wgs84.go
package v120

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedPositionWGS84 implements I062/105
// Calculated Position in WGS-84 Co-ordinates with high resolution
type CalculatedPositionWGS84 struct {
	Latitude  float64 // In degrees, positive = north
	Longitude float64 // In degrees, positive = east
}

func (p *CalculatedPositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 8)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading WGS-84 position: %w", err)
	}
	if n != 8 {
		return n, fmt.Errorf("insufficient data for WGS-84 position: got %d bytes, want 8", n)
	}

	// Extract latitude (32 bits, two's complement)
	rawLat := int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3])
	p.Latitude = float64(rawLat) * 180.0 / (1 << 25) // LSB = 180/(2^25) degrees

	// Extract longitude (32 bits, two's complement)
	rawLon := int32(data[4])<<24 | int32(data[5])<<16 | int32(data[6])<<8 | int32(data[7])
	p.Longitude = float64(rawLon) * 180.0 / (1 << 25) // LSB = 180/(2^25) degrees

	return n, p.Validate()
}

func (p *CalculatedPositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert latitude to raw value
	rawLat := int32(math.Round(p.Latitude * (1 << 25) / 180.0))

	// Convert longitude to raw value
	rawLon := int32(math.Round(p.Longitude * (1 << 25) / 180.0))

	data := make([]byte, 8)
	// Encode latitude
	data[0] = byte(rawLat >> 24)
	data[1] = byte(rawLat >> 16)
	data[2] = byte(rawLat >> 8)
	data[3] = byte(rawLat)

	// Encode longitude
	data[4] = byte(rawLon >> 24)
	data[5] = byte(rawLon >> 16)
	data[6] = byte(rawLon >> 8)
	data[7] = byte(rawLon)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing WGS-84 position: %w", err)
	}
	return n, nil
}

func (p *CalculatedPositionWGS84) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude out of valid range [-90,90]: %f", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude >= 180 {
		return fmt.Errorf("longitude out of valid range [-180,180): %f", p.Longitude)
	}
	return nil
}

func (p *CalculatedPositionWGS84) String() string {
	latDir := "N"
	if p.Latitude < 0 {
		latDir = "S"
	}
	lonDir := "E"
	if p.Longitude < 0 {
		lonDir = "W"
	}
	return fmt.Sprintf("%.6f°%s %.6f°%s", math.Abs(p.Latitude), latDir, math.Abs(p.Longitude), lonDir)
}
