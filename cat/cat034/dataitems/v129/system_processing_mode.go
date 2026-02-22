// cat/cat034/dataitems/v129/system_processing_mode.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SystemProcessingMode represents I034/060 - System Processing Mode
// Compound data item
type SystemProcessingMode struct {
	asterix.BaseCompoundItem
	COM   *uint8 // Common Part - optional, 1 byte
	PSR   *uint8 // PSR processing mode - optional, 1 byte
	SSR   *uint8 // SSR processing mode - optional, 1 byte
	MDS   *uint8 // Mode S processing mode - optional, 1 byte
}

// NewSystemProcessingMode creates a new System Processing Mode data item
func NewSystemProcessingMode() *SystemProcessingMode {
	return &SystemProcessingMode{
		BaseCompoundItem: asterix.BaseCompoundItem{},
	}
}

// Decode decodes the System Processing Mode from bytes
// TODO: Stub implementation - reads and discards bytes but doesn't parse them
// This field is not mandatory and full parsing isn't critical for now
func (s *SystemProcessingMode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, nil
	}

	bytesRead := 0

	// Read primary FSPEC byte
	fspec, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading compound FSPEC: %v", asterix.ErrBufferTooShort, err)
	}
	bytesRead++
	firstFspec := fspec

	// Read FSPEC extension bytes if needed (bit 1 = FX)
	for fspec&0x01 != 0 {
		fspec, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading FSPEC extension: %v", asterix.ErrBufferTooShort, err)
		}
		bytesRead++
	}

	// Now read subfields based on FSPEC bits
	// COM - bit 8 of first FSPEC (1 byte)
	if firstFspec&0x80 != 0 {
		_, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading COM: %v", asterix.ErrBufferTooShort, err)
		}
		bytesRead++
	}

	// Spare bits 7, 6

	// PSR - bit 5 of first FSPEC (1 byte)
	if firstFspec&0x10 != 0 {
		_, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading PSR: %v", asterix.ErrBufferTooShort, err)
		}
		bytesRead++
	}

	// SSR - bit 4 of first FSPEC (1 byte)
	if firstFspec&0x08 != 0 {
		_, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading SSR: %v", asterix.ErrBufferTooShort, err)
		}
		bytesRead++
	}

	// MDS - bit 3 of first FSPEC (1 byte)
	if firstFspec&0x04 != 0 {
		_, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading MDS: %v", asterix.ErrBufferTooShort, err)
		}
		bytesRead++
	}

	return bytesRead, nil
}

// Encode encodes the System Processing Mode to bytes
func (s *SystemProcessingMode) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Build FSPEC
	fspec := byte(0)
	if s.COM != nil {
		fspec |= 0x80 // bit 8
	}
	if s.PSR != nil {
		fspec |= 0x10 // bit 5
	}
	if s.SSR != nil {
		fspec |= 0x08 // bit 4
	}
	if s.MDS != nil {
		fspec |= 0x04 // bit 3
	}

	// Write FSPEC
	if err := buf.WriteByte(fspec); err != nil {
		return 0, fmt.Errorf("writing FSPEC: %w", err)
	}
	bytesWritten := 1

	// Write optional fields
	if s.COM != nil {
		if err := buf.WriteByte(*s.COM); err != nil {
			return bytesWritten, fmt.Errorf("writing COM: %w", err)
		}
		bytesWritten++
	}

	if s.PSR != nil {
		if err := buf.WriteByte(*s.PSR); err != nil {
			return bytesWritten, fmt.Errorf("writing PSR: %w", err)
		}
		bytesWritten++
	}

	if s.SSR != nil {
		if err := buf.WriteByte(*s.SSR); err != nil {
			return bytesWritten, fmt.Errorf("writing SSR: %w", err)
		}
		bytesWritten++
	}

	if s.MDS != nil {
		if err := buf.WriteByte(*s.MDS); err != nil {
			return bytesWritten, fmt.Errorf("writing MDS: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the System Processing Mode
func (s *SystemProcessingMode) Validate() error {
	// All fields are optional
	return nil
}

// String returns a string representation
func (s *SystemProcessingMode) String() string {
	result := "Processing Mode:"
	if s.COM != nil {
		result += fmt.Sprintf(" COM=%02X", *s.COM)
	}
	if s.PSR != nil {
		result += fmt.Sprintf(" PSR=%02X", *s.PSR)
	}
	if s.SSR != nil {
		result += fmt.Sprintf(" SSR=%02X", *s.SSR)
	}
	if s.MDS != nil {
		result += fmt.Sprintf(" MDS=%02X", *s.MDS)
	}
	return result
}
