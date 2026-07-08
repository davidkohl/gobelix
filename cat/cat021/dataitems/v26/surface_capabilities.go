// dataitems/cat021/surface_capabilities.go
package v26

import (
	"bytes"
	"fmt"
	"strings"
)

// SurfaceCapabilities implements I021/271
// Surface Capabilities and Characteristics (Extended, 1+ octets)
type SurfaceCapabilities struct {
	// First octet
	POA  bool  // Position Offset Applied
	CDTIS bool // Cockpit Display of Traffic Information Surface
	B2Low bool // Class B2 transmit power less than 70 Watts
	RAS  bool  // Receiving ATC Services
	IDENT bool // Setting of IDENT-switch

	// Second octet (extension 1)
	LW       uint8 // Length/Width code (0-15)
	hasExtension bool
}

func (s *SurfaceCapabilities) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// First octet
	var b1 uint8
	if s.POA {
		b1 |= 0x20
	}
	if s.CDTIS {
		b1 |= 0x10
	}
	if s.B2Low {
		b1 |= 0x08
	}
	if s.RAS {
		b1 |= 0x04
	}
	if s.IDENT {
		b1 |= 0x02
	}
	// Bits 3-2 are spare

	// Set FX bit if there's an extension
	if s.hasExtension || s.LW != 0 {
		b1 |= 0x01
	}

	if err := buf.WriteByte(b1); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// Second octet (extension) if needed
	if s.hasExtension || s.LW != 0 {
		var b2 uint8
		b2 = (s.LW & 0x0F) << 4
		// Bits 4-2 are spare
		// Bit 1 is FX (no further extension)

		if err := buf.WriteByte(b2); err != nil {
			return bytesWritten, fmt.Errorf("writing extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (s *SurfaceCapabilities) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Read first octet
	b1, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	s.POA = (b1 & 0x20) != 0
	s.CDTIS = (b1 & 0x10) != 0
	s.B2Low = (b1 & 0x08) != 0
	s.RAS = (b1 & 0x04) != 0
	s.IDENT = (b1 & 0x02) != 0

	fx := (b1 & 0x01) != 0

	// Read extension if present
	if fx {
		s.hasExtension = true
		b2, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading extension: %w", err)
		}
		bytesRead++

		s.LW = (b2 >> 4) & 0x0F
	}

	return bytesRead, s.Validate()
}

func (s *SurfaceCapabilities) Validate() error {
	if s.LW > 15 {
		return fmt.Errorf("invalid length/width code: %d", s.LW)
	}
	return nil
}

func (s *SurfaceCapabilities) String() string {
	var parts []string

	if s.POA {
		parts = append(parts, "POA")
	}
	if s.CDTIS {
		parts = append(parts, "CDTI-S")
	}
	if s.B2Low {
		parts = append(parts, "B2-Low")
	}
	if s.RAS {
		parts = append(parts, "RAS")
	}
	if s.IDENT {
		parts = append(parts, "IDENT")
	}

	if s.LW != 0 {
		parts = append(parts, fmt.Sprintf("L/W: %d", s.LW))
	}

	if len(parts) == 0 {
		return "No capabilities"
	}

	return strings.Join(parts, ", ")
}
