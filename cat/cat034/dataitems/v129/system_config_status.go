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
	COM   uint8   // Common Part - 1 byte
	PSR   *uint8  // PSR status - optional, 1 byte
	SSR   *uint8  // SSR status - optional, 1 byte
	MDS   *uint16 // Mode S status - optional, 2 bytes (per spec page 31)
}

// NewSystemConfigurationStatus creates a new System Configuration and Status data item
func NewSystemConfigurationStatus() *SystemConfigurationStatus {
	return &SystemConfigurationStatus{
		BaseCompoundItem: asterix.BaseCompoundItem{},
	}
}

// Decode decodes the System Configuration and Status from bytes
// TODO: Stub implementation - reads and discards bytes but doesn't parse them
// This field is not mandatory and full parsing isn't critical for now
func (s *SystemConfigurationStatus) Decode(buf *bytes.Buffer) (int, error) {
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

	// MDS - bit 3 of first FSPEC (2 bytes per spec page 31)
	if firstFspec&0x04 != 0 {
		var data [2]byte
		n, err := buf.Read(data[:])
		if err != nil || n != 2 {
			return bytesRead + n, fmt.Errorf("%w: reading MDS: need 2 bytes, have %d", asterix.ErrBufferTooShort, n)
		}
		bytesRead += 2
	}

	return bytesRead, nil
}

// Encode encodes the System Configuration and Status to bytes
func (s *SystemConfigurationStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Build FSPEC
	fspec := byte(0x80) // COM is always present in encode
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
		// MDS is 2 bytes
		var data [2]byte
		data[0] = byte(*s.MDS >> 8)   // High byte
		data[1] = byte(*s.MDS & 0xFF) // Low byte
		n, err := buf.Write(data[:])
		if err != nil {
			return bytesWritten + n, fmt.Errorf("writing MDS: %w", err)
		}
		bytesWritten += 2
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
		result += fmt.Sprintf(", MDS: %04X", *s.MDS) // 2 bytes = 4 hex digits
	}
	return result
}
