package v10

import (
	"bytes"
	"fmt"
)

// StationProcessingMode represents I002/060 - Station Processing Mode (CAT002 ed 1.2 §5.2.7).
// Variable length: one-octet parts, each with 7 indicator bits (8-2) and FX
// (bit 1). The meaning of the bits is not standardised.
type StationProcessingMode struct {
	// Mode is the first part's indicator bits. Kept for compatibility; it
	// equals Parts[0] after Decode. Encode uses it only when Parts is empty.
	Mode uint8
	// Parts holds the 7-bit indicators of every part in order.
	Parts []uint8
}

func (s *StationProcessingMode) Decode(buf *bytes.Buffer) (int, error) {
	parts, n, err := decodeFXChain(buf, "station processing mode")
	if err != nil {
		return n, err
	}
	s.Parts = parts
	s.Mode = parts[0]
	return n, nil
}

func (s *StationProcessingMode) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	parts := s.Parts
	if len(parts) == 0 {
		parts = []uint8{s.Mode}
	}
	return encodeFXChain(buf, parts, "station processing mode")
}

func (s *StationProcessingMode) Validate() error {
	if err := check7Bit([]uint8{s.Mode}, "station processing mode"); err != nil {
		return err
	}
	return check7Bit(s.Parts, "station processing mode")
}

func (s *StationProcessingMode) String() string {
	if len(s.Parts) > 1 {
		return fmt.Sprintf("Mode: % 02X", s.Parts)
	}
	return fmt.Sprintf("Mode: %02X", s.Mode)
}
