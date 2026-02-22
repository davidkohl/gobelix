// cat/common/dataitems/sector_number.go
package common

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SectorNumber implements the sector number data item
// Used in multiple categories: I002/020, I034/020
// Fixed length: 1 byte
// Antenna azimuth in degrees, LSB = 360/256 degrees (~1.406°)
type SectorNumber struct {
	SectorNumber float64 // Azimuth in degrees (0-360)
}

// Decode decodes the Sector Number from bytes
func (s *SectorNumber) Decode(buf *bytes.Buffer) (int, error) {
	var data [1]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading sector number: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("%w: need 1 byte for sector number, have %d", asterix.ErrBufferTooShort, n)
	}

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
