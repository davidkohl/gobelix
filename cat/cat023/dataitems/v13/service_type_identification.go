// cat/cat023/dataitems/v13/service_type_identification.go
package v13

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceTypeIdentification represents I023/015 - Service Type and Identification
// One-octet fixed length Data Item per ASTERIX CAT023 spec (section 5.2.3)
type ServiceTypeIdentification struct {
	SID   uint8 // Service Identification (4 bits, bits 8-5)
	STYPE uint8 // Service Type (4 bits, bits 4-1): 1=ADS-B VDL4, 2=ADS-B Ext Squitter, 3=ADS-B UAT, etc.
}

// Decode decodes the Service Type and Identification from bytes
func (s *ServiceTypeIdentification) Decode(buf *bytes.Buffer) (int, error) {
	// I023/015 is a ONE-octet fixed length Data Item per ASTERIX CAT023 spec (section 5.2.3):
	// - bits 8-5: SID (Service Identification) - 4 bits
	// - bits 4-1: STYP (Type of Service) - 4 bits
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: reading service type identification", asterix.ErrBufferTooShort)
	}

	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading service type identification", asterix.ErrBufferTooShort)
	}

	s.SID = (b >> 4) & 0x0F   // bits 8-5 (upper nibble)
	s.STYPE = b & 0x0F        // bits 4-1 (lower nibble)

	return 1, nil
}

// Encode encodes the Service Type and Identification to bytes
func (s *ServiceTypeIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Single octet: SID (bits 8-5) | STYP (bits 4-1)
	b := ((s.SID & 0x0F) << 4) | (s.STYPE & 0x0F)

	if err := buf.WriteByte(b); err != nil {
		return 0, fmt.Errorf("writing service type identification: %w", err)
	}

	return 1, nil
}

// Validate validates the Service Type and Identification
func (s *ServiceTypeIdentification) Validate() error {
	if s.SID > 15 {
		return fmt.Errorf("%w: service identification must be 0-15, got %d", asterix.ErrInvalidMessage, s.SID)
	}
	if s.STYPE < 1 || s.STYPE > 9 {
		return fmt.Errorf("%w: service type must be 1-9, got %d", asterix.ErrInvalidMessage, s.STYPE)
	}
	return nil
}

// String returns a string representation
func (s *ServiceTypeIdentification) String() string {
	serviceTypes := map[uint8]string{
		1: "ADS-B VDL4",
		2: "ADS-B Ext Squitter",
		3: "ADS-B UAT",
		4: "TIS-B VDL4",
		5: "TIS-B Ext Squitter",
		6: "TIS-B UAT",
		7: "FIS-B VDL4",
		8: "GRAS VDL4",
		9: "MLT",
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("SID:%d", s.SID))
	if name, ok := serviceTypes[s.STYPE]; ok {
		parts = append(parts, fmt.Sprintf("Type:%s", name))
	} else {
		parts = append(parts, fmt.Sprintf("Type:Unknown(%d)", s.STYPE))
	}

	return strings.Join(parts, ", ")
}
