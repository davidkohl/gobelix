// dataitems/cat021/service_id.go
package cat021

import (
	"bytes"
	"fmt"
)

// ServiceID implements I021/015
type ServiceID struct {
	Value uint8
}

func (s *ServiceID) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	err := buf.WriteByte(s.Value)
	if err != nil {
		return 0, fmt.Errorf("writing service ID: %w", err)
	}
	return 1, nil
}

func (s *ServiceID) Decode(buf *bytes.Buffer) (int, error) {
	var err error
	s.Value, err = buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading service ID: %w", err)
	}
	return 1, nil
}

func (s *ServiceID) Validate() error {
	return nil // All uint8 values are valid for service ID
}
