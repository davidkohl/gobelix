// cat/cat023/dataitems/v13/service_status.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceStatus represents I023/110 - Service Status
// Compound data item with optional subfields
type ServiceStatus struct {
	// Subfield #1: STAT - Status of the Service
	STAT *STATValue
}

// STATValue represents the STAT subfield (1 byte)
type STATValue struct {
	STAT uint8 // 2-bit status: 0=Unknown, 1=Failed, 2=Disabled, 3=Degraded, 4=Normal
}

// Decode decodes the Service Status from bytes
func (s *ServiceStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read primary subfield (1 octet with FX extension bit)
	primaryBytes := make([]byte, 0, 1)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading service status primary subfield", asterix.ErrBufferTooShort)
		}
		bytesRead++
		primaryBytes = append(primaryBytes, b)

		// Check if there's an extension
		hasExtension := (b & 0x01) != 0
		if !hasExtension {
			break
		}
	}

	// Now read subfields based on the bits set in the primary subfield
	subfieldIndex := 0
	for byteIdx := 0; byteIdx < len(primaryBytes); byteIdx++ {
		// Process bits 8-2 (bit 1 is FX)
		for bitPos := 7; bitPos >= 1; bitPos-- {
			if (primaryBytes[byteIdx] & (1 << bitPos)) != 0 {
				// This subfield is present
				n, err := s.decodeSubfield(subfieldIndex, buf)
				bytesRead += n
				if err != nil {
					return bytesRead, fmt.Errorf("decoding subfield #%d: %w", subfieldIndex+1, err)
				}
			}
			subfieldIndex++
		}
	}

	return bytesRead, nil
}

// decodeSubfield decodes a specific subfield based on index
func (s *ServiceStatus) decodeSubfield(index int, buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading service status subfield #%d", asterix.ErrBufferTooShort, index+1)
	}

	switch index {
	case 0: // #1: STAT
		s.STAT = &STATValue{STAT: (b >> 5) & 0x07} // bits 8-6 (3 bits for values 0-4)
	default:
		return 1, fmt.Errorf("unknown service status subfield index: %d", index)
	}

	return 1, nil
}

// Encode encodes the Service Status to bytes
func (s *ServiceStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	// Build primary subfield based on which fields are present
	primaryBytes := s.buildPrimarySubfield()
	n, err := buf.Write(primaryBytes)
	if err != nil {
		return n, fmt.Errorf("writing service status primary subfield: %w", err)
	}
	bytesWritten := n

	// Encode STAT subfield if present
	if s.STAT != nil {
		data := []byte{(s.STAT.STAT & 0x07) << 5}
		n, err := buf.Write(data)
		bytesWritten += n
		if err != nil {
			return bytesWritten, fmt.Errorf("writing STAT subfield: %w", err)
		}
	}

	return bytesWritten, nil
}

// buildPrimarySubfield builds the FSPEC for compound data item
func (s *ServiceStatus) buildPrimarySubfield() []byte {
	// Determine which subfields are present
	presence := []bool{
		s.STAT != nil,
	}

	// Build primary subfield byte
	var b byte
	if presence[0] {
		b |= 0x80 // bit 8
	}
	// FX bit is 0 (no further extensions needed for now)

	return []byte{b}
}

// Validate validates the Service Status
func (s *ServiceStatus) Validate() error {
	if s.STAT != nil && s.STAT.STAT > 4 {
		return fmt.Errorf("%w: STAT value must be 0-4, got %d", asterix.ErrInvalidMessage, s.STAT.STAT)
	}
	return nil
}

// String returns a string representation
func (s *ServiceStatus) String() string {
	if s.STAT == nil {
		return "ServiceStatus{}"
	}

	statDesc := map[uint8]string{
		0: "Unknown",
		1: "Failed",
		2: "Disabled",
		3: "Degraded",
		4: "Normal",
	}

	if desc, ok := statDesc[s.STAT.STAT]; ok {
		return fmt.Sprintf("ServiceStatus{STAT:%s}", desc)
	}

	return fmt.Sprintf("ServiceStatus{STAT:Invalid(%d)}", s.STAT.STAT)
}
