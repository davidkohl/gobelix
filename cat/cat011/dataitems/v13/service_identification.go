package v13

import (
	"bytes"
	"fmt"
)

// ServiceIdentification implements I011/015 - Service Identification
// Definition: Identification of the service provided to one or more users.
type ServiceIdentification struct {
	SID uint8 // Service Identification (0-255)
}

func (s *ServiceIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading service identification: %w", err)
	}

	s.SID = data
	return 1, nil
}

func (s *ServiceIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(s.SID); err != nil {
		return 0, fmt.Errorf("writing service identification: %w", err)
	}
	return 1, nil
}

func (s *ServiceIdentification) Validate() error {
	return nil
}

func (s *ServiceIdentification) String() string {
	return fmt.Sprintf("Service ID: %d", s.SID)
}
