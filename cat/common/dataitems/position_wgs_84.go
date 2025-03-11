// dataitems/common/position.go
package common

import (
	"bytes"
	"fmt"
	"math"
)

type Position struct {
	Latitude  float64 // -90° to +90°
	Longitude float64 // -180° to +180°
}

const ResolutionWGS84 = 180.0 / (1 << 23) // ≈ 2.14576721191406 × 10^-5 degrees

func (p *Position) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw values
	latRaw := int32(math.Round(p.Latitude / ResolutionWGS84))
	lonRaw := int32(math.Round(p.Longitude / ResolutionWGS84))

	// Ensure values fit in 24 bits and handle negative numbers
	latRaw &= 0xFFFFFF
	lonRaw &= 0xFFFFFF

	bytesWritten := 0
	// Write latitude (3 bytes)
	if err := buf.WriteByte(byte(latRaw >> 16)); err != nil {
		return bytesWritten, fmt.Errorf("writing latitude byte 1: %w", err)
	}
	bytesWritten++
	if err := buf.WriteByte(byte(latRaw >> 8)); err != nil {
		return bytesWritten, fmt.Errorf("writing latitude byte 2: %w", err)
	}
	bytesWritten++
	if err := buf.WriteByte(byte(latRaw)); err != nil {
		return bytesWritten, fmt.Errorf("writing latitude byte 3: %w", err)
	}
	bytesWritten++

	// Write longitude (3 bytes)
	if err := buf.WriteByte(byte(lonRaw >> 16)); err != nil {
		return bytesWritten, fmt.Errorf("writing longitude byte 1: %w", err)
	}
	bytesWritten++
	if err := buf.WriteByte(byte(lonRaw >> 8)); err != nil {
		return bytesWritten, fmt.Errorf("writing longitude byte 2: %w", err)
	}
	bytesWritten++
	if err := buf.WriteByte(byte(lonRaw)); err != nil {
		return bytesWritten, fmt.Errorf("writing longitude byte 3: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (p *Position) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read 6 bytes (48 bits)
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil {
		return bytesRead, fmt.Errorf("reading position data: %w", err)
	}
	bytesRead = n

	if n != 6 {
		return bytesRead, fmt.Errorf("expected 6 bytes, got %d", n)
	}

	// Extract 24-bit values and convert from two's complement
	latRaw := int32(uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2]))
	lonRaw := int32(uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5]))

	// Handle negative numbers
	if latRaw&0x800000 != 0 {
		latRaw = -1 * (((^latRaw) + 1) & 0xFFFFFF)
	}
	if lonRaw&0x800000 != 0 {
		lonRaw = -1 * (((^lonRaw) + 1) & 0xFFFFFF)
	}

	// Convert to degrees
	p.Latitude = float64(latRaw) * ResolutionWGS84
	p.Longitude = float64(lonRaw) * ResolutionWGS84

	return bytesRead, p.Validate()
}

func (p *Position) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude %f outside valid range [-90,+90]", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude %f outside valid range [-180,+180]", p.Longitude)
	}
	return nil
}

func (p *Position) String() string {
	return fmt.Sprintf("%.6f°N %.6f°E", p.Latitude, p.Longitude)
}
