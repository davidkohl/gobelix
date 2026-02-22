// dataitems/cat021/service_management.go
package v26

import (
	"bytes"
	"fmt"
)

// ServiceManagement implements I021/016
// Service Management field (1 octet)
type ServiceManagement struct {
	RP uint8 // Reporting Period in seconds (1-8)
}

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
	if s.RP < 1 || s.RP > 8 {
		return fmt.Errorf("invalid reporting period: %d (must be 1-8)", s.RP)
	}
	return nil
}

func (s *ServiceManagement) String() string {
	return fmt.Sprintf("Reporting Period: %ds", s.RP)
}
