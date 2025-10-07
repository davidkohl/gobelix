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
func (s *SystemProcessingMode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		// Empty buffer - field indicated but not present (trailing garbage)
		// Return success with 0 bytes read to allow graceful handling
		return 0, nil
	}

	bytesRead := 0

	// Read primary FSPEC
	fspec := buf.Next(1)[0]
	bytesRead++

	// Read COM if present (bit 8 of fspec)
	if fspec&0x80 != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: need 1 byte for COM", asterix.ErrBufferTooShort)
		}
		com := buf.Next(1)[0]
		s.COM = &com
		bytesRead++
	}

	// Read PSR if present (bit 7 of fspec)
	if fspec&0x40 != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: need 1 byte for PSR", asterix.ErrBufferTooShort)
		}
		psr := buf.Next(1)[0]
		s.PSR = &psr
		bytesRead++
	}

	// Read SSR if present (bit 6 of fspec)
	if fspec&0x20 != 0 {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: need 1 byte for SSR", asterix.ErrBufferTooShort)
		}
		ssr := buf.Next(1)[0]
		s.SSR = &ssr
		bytesRead++
	}

	// Read MDS if present (bit 5 of fspec)
	if fspec&0x10 != 0 {
		if buf.Len() < 1 {
			// FSPEC says MDS present but buffer empty - trailing garbage, ignore silently
			return bytesRead, nil
		}
		mds := buf.Next(1)[0]
		s.MDS = &mds
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
		fspec |= 0x80
	}
	if s.PSR != nil {
		fspec |= 0x40
	}
	if s.SSR != nil {
		fspec |= 0x20
	}
	if s.MDS != nil {
		fspec |= 0x10
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
	// At least one subfield should be present
	if s.COM == nil && s.PSR == nil && s.SSR == nil && s.MDS == nil {
		return fmt.Errorf("%w: at least one processing mode must be present", asterix.ErrInvalidMessage)
	}
	return nil
}

// String returns a string representation
func (s *SystemProcessingMode) String() string {
	result := "Processing Mode:"
	if s.COM != nil {
		result += fmt.Sprintf(" COM: %02X", *s.COM)
	}
	if s.PSR != nil {
		result += fmt.Sprintf(" PSR: %02X", *s.PSR)
	}
	if s.SSR != nil {
		result += fmt.Sprintf(" SSR: %02X", *s.SSR)
	}
	if s.MDS != nil {
		result += fmt.Sprintf(" MDS: %02X", *s.MDS)
	}
	return result
}
