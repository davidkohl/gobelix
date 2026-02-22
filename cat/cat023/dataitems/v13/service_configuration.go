// cat/cat023/dataitems/v13/service_configuration.go
package v13

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceConfiguration represents I023/101 - Service Configuration
// Two-octet extensible field per spec page 25
type ServiceConfiguration struct {
	RP   uint8 // Report Period for Category 021 Reports (8 bits, LSB=0.5s, 0=data driven)
	SC   uint8 // Service Class - 3 bits (0-7)
	SSRP uint8 // Service Status Reporting Period - 7 bits (1-127 seconds, in extension)
}

// Decode decodes the Service Configuration from bytes
func (s *ServiceConfiguration) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Check we have at least 2 bytes for the primary octets
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: reading service configuration", asterix.ErrBufferTooShort)
	}

	// Peek at the second byte to check FX before consuming it
	secondByte := buf.Bytes()[1]
	fx := (secondByte & 0x01) != 0 // bit 1 (FX)

	// If FX is set, validate we have the extension byte available
	if fx && buf.Len() < 3 {
		return 0, fmt.Errorf("%w: reading service configuration extension", asterix.ErrBufferTooShort)
	}

	// Validation passed - now safely read the first octet: RP (bits 8-1) - Report Period, LSB=0.5s
	rp, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading service configuration RP", asterix.ErrBufferTooShort)
	}
	bytesRead++
	s.RP = rp

	// Second octet: SC (bits 8-6), Spare (bits 5-2), FX (bit 1)
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading service configuration SC", asterix.ErrBufferTooShort)
	}
	bytesRead++

	s.SC = (b >> 5) & 0x07 // bits 8-6 (3 bits)
	// fx is already set from the peek validation above

	// First extension octet (if FX is set): SSRP (bits 8-2), FX (bit 1)
	if fx {
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading service configuration SSRP", asterix.ErrBufferTooShort)
		}
		bytesRead++

		s.SSRP = (b >> 1) & 0x7F // bits 8-2 (7 bits)
		fx = (b & 0x01) != 0      // Check FX bit for next extension
	}

	// Read any additional extension octets (spare in v1.3)
	for fx {
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading service configuration extension", asterix.ErrBufferTooShort)
		}
		bytesRead++
		fx = (b & 0x01) != 0 // Check FX bit for next extension
	}

	return bytesRead, nil
}

// Encode encodes the Service Configuration to bytes
func (s *ServiceConfiguration) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// First octet: RP (bits 8-1) - Report Period, LSB=0.5s
	if err := buf.WriteByte(s.RP); err != nil {
		return bytesWritten, fmt.Errorf("writing service configuration RP: %w", err)
	}
	bytesWritten++

	// Determine if we need extension octet (for SSRP)
	needExtension := s.SSRP > 0

	// Second octet: SC (bits 8-6), Spare (bits 5-2), FX (bit 1)
	b := (s.SC & 0x07) << 5 // bits 8-6 (3 bits)
	if needExtension {
		b |= 0x01 // FX bit
	}

	if err := buf.WriteByte(b); err != nil {
		return bytesWritten, fmt.Errorf("writing service configuration SC: %w", err)
	}
	bytesWritten++

	// Extension octet (if needed): SSRP (bits 8-2), FX=0 (bit 1)
	if needExtension {
		b = (s.SSRP & 0x7F) << 1 // bits 8-2 (7 bits)
		// FX bit is 0 (no further extensions)

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing service configuration SSRP: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Service Configuration
func (s *ServiceConfiguration) Validate() error {
	if s.SC > 7 {
		return fmt.Errorf("%w: service class must be 0-7, got %d", asterix.ErrInvalidMessage, s.SC)
	}
	if s.SSRP > 127 {
		return fmt.Errorf("%w: service status reporting period must be 0-127, got %d", asterix.ErrInvalidMessage, s.SSRP)
	}
	return nil
}

// String returns a string representation
func (s *ServiceConfiguration) String() string {
	var parts []string

	if s.RP == 0 {
		parts = append(parts, "RP:Data Driven")
	} else {
		// LSB = 0.5s, so multiply by 0.5
		period := float64(s.RP) * 0.5
		parts = append(parts, fmt.Sprintf("RP:%.1fs", period))
	}

	serviceClasses := map[uint8]string{
		0: "No Information",
		1: "NRA Class",
		2: "Reserved",
		3: "Reserved",
		4: "Reserved",
		5: "Reserved",
		6: "Reserved",
		7: "Reserved",
	}
	parts = append(parts, fmt.Sprintf("SC:%s", serviceClasses[s.SC]))
	parts = append(parts, fmt.Sprintf("SSRP:%ds", s.SSRP))

	return strings.Join(parts, ", ")
}
