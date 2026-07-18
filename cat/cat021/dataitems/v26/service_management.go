// dataitems/cat021/service_management.go
package v26

import (
	"bytes"
	"fmt"
)

// ServiceManagement implements I021/016
// Service Management field (1 octet)
type ServiceManagement struct {
	RP uint8 // Report Period, LSB = 0.5 s (0 = data driven mode)
}

// Seconds returns the report period in seconds.
func (s *ServiceManagement) Seconds() float64 { return float64(s.RP) * 0.5 }

func (s *ServiceManagement) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	if err := buf.WriteByte(s.RP); err != nil {
		return 0, fmt.Errorf("writing service management: %w", err)
	}
	return 1, nil
}

func (s *ServiceManagement) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading service management: %w", err)
	}

	s.RP = b
	return 1, s.Validate()
}

func (s *ServiceManagement) Validate() error {
	// All octet values are valid: RP has LSB 0.5 s covering 0-127.5 s,
	// with 0 meaning data driven mode. Real-world ground stations send 0.
	return nil
}

func (s *ServiceManagement) String() string {
	if s.RP == 0 {
		return "Report Period: data driven"
	}
	return fmt.Sprintf("Report Period: %.1fs", s.Seconds())
}
