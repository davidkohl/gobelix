// cat/cat034/dataitems/v129/sector_number.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SectorNumber represents I034/020 - Sector Number
// Fixed length: 1 byte
// Antenna azimuth in the horizontal plane in binary representation
type SectorNumber struct {
	SectorNumber float64 // Degrees (360/256 degree resolution)
}

// NewSectorNumber creates a new Sector Number data item
func NewSectorNumber() *SectorNumber {
	return &SectorNumber{}
}

// Decode decodes the Sector Number from bytes
func (s *SectorNumber) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		// Empty buffer - field indicated but not present (trailing garbage)
		// Return success with 0 bytes read to allow graceful handling
		return 0, nil
	}

	data := buf.Next(1)
	// Convert from 0-255 to 0-360 degrees
	s.SectorNumber = float64(data[0]) * (360.0 / 256.0)

	return 1, nil
}

// Encode encodes the Sector Number to bytes
func (s *SectorNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Convert from degrees to 0-255
	value := uint8((s.SectorNumber / 360.0) * 256.0)

	if err := buf.WriteByte(value); err != nil {
		return 0, fmt.Errorf("writing sector number: %w", err)
	}

	return 1, nil
}

// Validate validates the Sector Number
func (s *SectorNumber) Validate() error {
	if s.SectorNumber < 0 || s.SectorNumber >= 360 {
		return fmt.Errorf("%w: sector number must be 0-360 degrees, got %.2f", asterix.ErrInvalidMessage, s.SectorNumber)
	}
	return nil
}

// String returns a string representation
func (s *SectorNumber) String() string {
	return fmt.Sprintf("%.2f°", s.SectorNumber)
}
