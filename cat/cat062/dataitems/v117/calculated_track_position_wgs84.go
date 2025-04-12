// dataitems/cat062/calculated_position_wgs84.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// CalculatedPositionWGS84 implements I062/105
// Represents the calculated position of a target in WGS-84 coordinates
type CalculatedPositionWGS84 struct {
	// Latitude in degrees (-90° to +90°)
	Latitude float64

	// Longitude in degrees (-180° to +180°)
	Longitude float64
}

// Decode parses an ASTERIX Category 062 I105 data item from the buffer
func (p *CalculatedPositionWGS84) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 8 {
		return 0, fmt.Errorf("buffer too short for WGS-84 position (need 8 bytes)")
	}

	data := make([]byte, 8)
	n, err := buf.Read(data)
	if err != nil || n != 8 {
		return n, fmt.Errorf("reading WGS-84 position: %w", err)
	}

	// Extract latitude (32 bits) as a signed value
	latBits := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])

	// Convert to signed int32 if needed (two's complement)
	var latValue int32
	if (latBits & 0x80000000) != 0 {
		// Negative value (two's complement)
		latValue = -int32(^latBits + 1)
	} else {
		// Positive value
		latValue = int32(latBits)
	}

	// Calculate latitude in degrees
	// Resolution is 180/2^25 degrees per bit
	p.Latitude = float64(latValue) * 180.0 / float64(1<<25)

	// Extract longitude (32 bits) as a signed value
	lonBits := uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])

	// Convert to signed int32 if needed (two's complement)
	var lonValue int32
	if (lonBits & 0x80000000) != 0 {
		// Negative value (two's complement)
		lonValue = -int32(^lonBits + 1)
	} else {
		// Positive value
		lonValue = int32(lonBits)
	}

	// Calculate longitude in degrees
	// Resolution is 180/2^25 degrees per bit
	p.Longitude = float64(lonValue) * 180.0 / float64(1<<25)

	return n, nil
}

// Encode serializes the WGS-84 position into the buffer
func (p *CalculatedPositionWGS84) Encode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 8)

	// Convert latitude from degrees to the binary representation
	// Resolution is 180/2^25 degrees per bit
	latValue := int32(math.Round(p.Latitude * float64(1<<25) / 180.0))

	// Handle two's complement for negative values
	var latBits uint32
	if latValue < 0 {
		latBits = uint32(^(-latValue) + 1) // Two's complement for negative values
	} else {
		latBits = uint32(latValue)
	}

	// Store latitude in first 4 bytes
	data[0] = byte(latBits >> 24)
	data[1] = byte(latBits >> 16)
	data[2] = byte(latBits >> 8)
	data[3] = byte(latBits)

	// Convert longitude from degrees to the binary representation
	// Resolution is 180/2^25 degrees per bit
	lonValue := int32(math.Round(p.Longitude * float64(1<<25) / 180.0))

	// Handle two's complement for negative values
	var lonBits uint32
	if lonValue < 0 {
		lonBits = uint32(^(-lonValue) + 1) // Two's complement for negative values
	} else {
		lonBits = uint32(lonValue)
	}

	// Store longitude in last 4 bytes
	data[4] = byte(lonBits >> 24)
	data[5] = byte(lonBits >> 16)
	data[6] = byte(lonBits >> 8)
	data[7] = byte(lonBits)

	return buf.Write(data)
}

// String returns a human-readable representation of the WGS-84 position
func (p *CalculatedPositionWGS84) String() string {
	// Format for standard aeronautical display:
	// - Latitude: N/S followed by degrees
	// - Longitude: E/W followed by degrees
	latDir := "N"
	if p.Latitude < 0 {
		latDir = "S"
	}

	lonDir := "E"
	if p.Longitude < 0 {
		lonDir = "W"
	}

	return fmt.Sprintf("%s%s %s%s",
		latDir, formatCoordinate(math.Abs(p.Latitude)),
		lonDir, formatCoordinate(math.Abs(p.Longitude)))
}

// Validate performs validation on the WGS-84 position
func (p *CalculatedPositionWGS84) Validate() error {
	// Check latitude is within valid range [-90, 90]
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude out of range [-90,90]: %f", p.Latitude)
	}

	// Check longitude is within valid range [-180, 180)
	if p.Longitude < -180 || p.Longitude >= 180 {
		return fmt.Errorf("longitude out of range [-180,180): %f", p.Longitude)
	}

	return nil
}

// formatCoordinate formats a coordinate value in degrees, minutes, seconds format
// with appropriate precision
func formatCoordinate(value float64) string {
	degrees := int(value)
	minutesFloat := (value - float64(degrees)) * 60.0
	minutes := int(minutesFloat)
	seconds := (minutesFloat - float64(minutes)) * 60.0

	// Use 4 decimals of precision for seconds, which is sufficiently precise
	// for the data item's resolution (approximately 0.5 meters at the equator)
	return fmt.Sprintf("%d°%d'%.4f\"", degrees, minutes, seconds)
}
