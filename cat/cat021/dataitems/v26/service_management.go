// dataitems/cat021/service_management.go
package v26

import (
	"bytes"
	"fmt"
)

// ServiceManagement implements I021/016
// Service Management field (1 octet)
//
// RP is the raw Report Period with LSB = 0.5 seconds (CAT021 v2.6 §5.2.5).
// RP = 0 means data-driven mode (ED-129B REQ 90); periodic mode encodes the
// configured period, e.g. a 4-second period is RP = 8.
type ServiceManagement struct {
	RP uint8 // Report Period, LSB = 0.5 s; 0 = data-driven
}

func (s *ServiceManagement) Encode(buf *bytes.Buffer) (int, error) {
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
	return 1, nil
}

// Validate checks if the ServiceManagement contains valid data. Any raw byte
// is a legal Report Period (0 = data-driven mode), so there is nothing to
// reject.
func (s *ServiceManagement) Validate() error {
	return nil
}

func (s *ServiceManagement) String() string {
	if s.RP == 0 {
		return "Data-driven"
	}
	return fmt.Sprintf("Reporting Period: %.1fs", float64(s.RP)*0.5)
}
