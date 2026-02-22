// cat/cat034/dataitems/v129/data_source_position.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// DataSourcePosition implements I034/120 - 3D-Position of Data Source
// Fixed length: 8 bytes
// 3D-Position of the radar/data source in WGS-84 coordinates
type DataSourcePosition struct {
	Height    float64 // Height above WGS-84 ellipsoid in feet (LSB = 25 ft)
	Latitude  float64 // Latitude in degrees (-90 to +90)
	Longitude float64 // Longitude in degrees (-180 to +180)
}

const (
	// LSB for height is 25 feet
	heightLSB = 25.0
	// LSB for lat/lon is 180/2^23 degrees
	latLonLSB = 180.0 / 8388608.0 // 180 / 2^23
)

// NewDataSourcePosition creates a new Data Source Position data item
func NewDataSourcePosition() *DataSourcePosition {
	return &DataSourcePosition{}
}

// Decode decodes the Data Source Position from bytes
func (d *DataSourcePosition) Decode(buf *bytes.Buffer) (int, error) {
	var data [8]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading data source position: %w", err)
	}
	if n != 8 {
		return n, fmt.Errorf("%w: need 8 bytes for I034/120, have %d", asterix.ErrBufferTooShort, n)
	}

	// Bytes 0-1: Height (16-bit signed)
	rawHeight := int16(uint16(data[0])<<8 | uint16(data[1]))
	d.Height = float64(rawHeight) * heightLSB

	// Bytes 2-4: Latitude (24-bit signed, two's complement)
	rawLat := int32(data[2])<<16 | int32(data[3])<<8 | int32(data[4])
	// Sign extend from 24-bit to 32-bit
	if rawLat&0x800000 != 0 {
		rawLat -= 0x1000000 // Convert from unsigned 24-bit to signed
	}
	d.Latitude = float64(rawLat) * latLonLSB

	// Bytes 5-7: Longitude (24-bit signed, two's complement)
	rawLon := int32(data[5])<<16 | int32(data[6])<<8 | int32(data[7])
	// Sign extend from 24-bit to 32-bit
	if rawLon&0x800000 != 0 {
		rawLon -= 0x1000000 // Convert from unsigned 24-bit to signed
	}
	d.Longitude = float64(rawLon) * latLonLSB

	return 8, nil
}

// Encode encodes the Data Source Position to bytes
func (d *DataSourcePosition) Encode(buf *bytes.Buffer) (int, error) {
	if err := d.Validate(); err != nil {
		return 0, err
	}

	// Convert height to raw value (16-bit signed)
	rawHeight := int16(d.Height / heightLSB)

	// Convert latitude to raw value (24-bit signed)
	rawLat := int32(d.Latitude / latLonLSB)

	// Convert longitude to raw value (24-bit signed)
	rawLon := int32(d.Longitude / latLonLSB)

	// Write 8 bytes
	data := []byte{
		// Height (2 bytes)
		byte(rawHeight >> 8),
		byte(rawHeight),
		// Latitude (3 bytes)
		byte(rawLat >> 16),
		byte(rawLat >> 8),
		byte(rawLat),
		// Longitude (3 bytes)
		byte(rawLon >> 16),
		byte(rawLon >> 8),
		byte(rawLon),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing data source position: %w", err)
	}

	return n, nil
}

// Validate validates the Data Source Position
func (d *DataSourcePosition) Validate() error {
	// Height: 16-bit signed * 25 ft = -819200 to +819175 ft
	maxHeight := float64(32767) * heightLSB
	minHeight := float64(-32768) * heightLSB
	if d.Height < minHeight || d.Height > maxHeight {
		return fmt.Errorf("%w: height must be %.0f to %.0f ft, got %.0f",
			asterix.ErrInvalidMessage, minHeight, maxHeight, d.Height)
	}

	// Latitude must be -90 to +90 degrees
	if d.Latitude < -90 || d.Latitude > 90 {
		return fmt.Errorf("%w: latitude must be -90 to +90 degrees, got %.6f",
			asterix.ErrInvalidMessage, d.Latitude)
	}

	// Longitude must be -180 to +180 degrees
	if d.Longitude < -180 || d.Longitude > 180 {
		return fmt.Errorf("%w: longitude must be -180 to +180 degrees, got %.6f",
			asterix.ErrInvalidMessage, d.Longitude)
	}

	return nil
}

// String returns a string representation
func (d *DataSourcePosition) String() string {
	return fmt.Sprintf("lat: %.6f°, lon: %.6f°, alt: %.0f ft", d.Latitude, d.Longitude, d.Height)
}
