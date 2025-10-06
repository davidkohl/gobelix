// cat/cat034/dataitems/v129/system_config_status.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SystemConfigurationStatus represents I034/050 - System Configuration and Status
// Compound data item
type SystemConfigurationStatus struct {
	asterix.BaseCompoundItem
	COM   uint8  // Common Part - 1 byte
	PSR   *uint8 // PSR status - optional, 1 byte
	SSR   *uint8 // SSR status - optional, 1 byte
	MDS   *uint8 // Mode S status - optional, 1 byte
}

// NewSystemConfigurationStatus creates a new System Configuration and Status data item
func NewSystemConfigurationStatus() *SystemConfigurationStatus {
	return &SystemConfigurationStatus{
		BaseCompoundItem: asterix.BaseCompoundItem{},
	}
}

// Decode decodes the System Configuration and Status from bytes
func (s *SystemConfigurationStatus) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte for FSPEC, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0

	// Read primary FSPEC
	fspec := buf.Next(1)[0]
	bytesRead++

	// Read COM (always present)
	if buf.Len() < 1 {
		return bytesRead, fmt.Errorf("%w: need 1 byte for COM", asterix.ErrBufferTooShort)
	}
	s.COM = buf.Next(1)[0]
	bytesRead++

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
			return bytesRead, fmt.Errorf("%w: need 1 byte for MDS", asterix.ErrBufferTooShort)
		}
		mds := buf.Next(1)[0]
		s.MDS = &mds
		bytesRead++
	}

	return bytesRead, nil
}

// Encode encodes the System Configuration and Status to bytes
func (s *SystemConfigurationStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Build FSPEC
	fspec := byte(0)
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

	// Write COM
	if err := buf.WriteByte(s.COM); err != nil {
		return bytesWritten, fmt.Errorf("writing COM: %w", err)
	}
	bytesWritten++

	// Write optional fields
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

// Validate validates the System Configuration and Status
func (s *SystemConfigurationStatus) Validate() error {
	// COM is always required
	return nil
}

// String returns a string representation
func (s *SystemConfigurationStatus) String() string {
	result := fmt.Sprintf("COM: %02X", s.COM)
	if s.PSR != nil {
		result += fmt.Sprintf(", PSR: %02X", *s.PSR)
	}
	if s.SSR != nil {
		result += fmt.Sprintf(", SSR: %02X", *s.SSR)
	}
	if s.MDS != nil {
		result += fmt.Sprintf(", MDS: %02X", *s.MDS)
	}
	return result
}
