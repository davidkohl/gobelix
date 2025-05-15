// cat/cat021/dataitems/v26/position.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// Position implements I021/130 (Position in WGS-84 coordinates)
// This data item represents the aircraft position in WGS-84 coordinates.
type Position struct {
	Latitude  float64 // Latitude in degrees (-90° to +90°)
	Longitude float64 // Longitude in degrees (-180° to +180°)
}

// Decode reads the position data from the buffer
func (p *Position) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading position: %w", err)
	}
	if n != 6 {
		return n, fmt.Errorf("insufficient data for position: got %d bytes, want 6", n)
	}

	// Extract latitude value (2's complement, LSB = 180/2^23 degrees)
	latRaw := int32(data[0])<<16 | int32(data[1])<<8 | int32(data[2])
	if latRaw > 0x7FFFFF {
		// Handle negative values (2's complement)
		latRaw = latRaw - 0x1000000
	}
	p.Latitude = float64(latRaw) * 180.0 / 8388608.0 // 2^23 = 8388608

	// Extract longitude value (2's complement, LSB = 180/2^23 degrees)
	lonRaw := int32(data[3])<<16 | int32(data[4])<<8 | int32(data[5])
	if lonRaw > 0x7FFFFF {
		// Handle negative values (2's complement)
		lonRaw = lonRaw - 0x1000000
	}
	p.Longitude = float64(lonRaw) * 180.0 / 8388608.0 // 2^23 = 8388608

	return n, p.Validate()
}

// Encode writes the position data to the buffer
func (p *Position) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, fmt.Errorf("validating position: %w", err)
	}

	// Convert latitude to raw value (2's complement, LSB = 180/2^23 degrees)
	latRaw := int32(math.Round(p.Latitude * 8388608.0 / 180.0))
	if latRaw < 0 {
		latRaw += 0x1000000 // Convert negative value to 2's complement
	}

	// Convert longitude to raw value (2's complement, LSB = 180/2^23 degrees)
	lonRaw := int32(math.Round(p.Longitude * 8388608.0 / 180.0))
	if lonRaw < 0 {
		lonRaw += 0x1000000 // Convert negative value to 2's complement
	}

	// Write latitude bytes
	data := make([]byte, 6)
	data[0] = byte(latRaw >> 16)
	data[1] = byte(latRaw >> 8)
	data[2] = byte(latRaw)

	// Write longitude bytes
	data[3] = byte(lonRaw >> 16)
	data[4] = byte(lonRaw >> 8)
	data[5] = byte(lonRaw)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing position: %w", err)
	}

	return n, nil
}

// Validate checks if the position values are valid
func (p *Position) Validate() error {
	// Use the validation utilities to check latitude and longitude
	if err := asterix.ValidateLatitude("Position.Latitude", p.Latitude); err != nil {
		return err
	}
	if err := asterix.ValidateLongitude("Position.Longitude", p.Longitude); err != nil {
		return err
	}
	return nil
}

// String returns a string representation of the position
func (p *Position) String() string {
	return fmt.Sprintf("Lat: %.6f°, Lon: %.6f°", p.Latitude, p.Longitude)
}
