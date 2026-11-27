package v10

import (
	"bytes"
	"fmt"
)

// StationConfigurationStatus represents I002/050 - Station Configuration Status (CAT002 ed 1.2 §5.2.6).
// Variable length: one-octet parts, each with 7 indicator bits (8-2) and FX
// (bit 1). The meaning of the bits is not standardised.
type StationConfigurationStatus struct {
	// Status is the first part's indicator bits. Kept for compatibility; it
	// equals Parts[0] after Decode. Encode uses it only when Parts is empty.
	Status uint8
	// Parts holds the 7-bit indicators of every part in order.
	Parts []uint8
}

func (s *StationConfigurationStatus) Decode(buf *bytes.Buffer) (int, error) {
	parts, n, err := decodeFXChain(buf, "station configuration status")
	if err != nil {
		return n, err
	}
	s.Parts = parts
	s.Status = parts[0]
	return n, nil
}

func (s *StationConfigurationStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	parts := s.Parts
	if len(parts) == 0 {
		parts = []uint8{s.Status}
	}
	return encodeFXChain(buf, parts, "station configuration status")
}

func (s *StationConfigurationStatus) Validate() error {
	if err := check7Bit([]uint8{s.Status}, "station configuration status"); err != nil {
		return err
	}
	return check7Bit(s.Parts, "station configuration status")
}

func (s *StationConfigurationStatus) String() string {
	if len(s.Parts) > 1 {
		return fmt.Sprintf("Status: % 02X", s.Parts)
	}
	return fmt.Sprintf("Status: %02X", s.Status)
}
