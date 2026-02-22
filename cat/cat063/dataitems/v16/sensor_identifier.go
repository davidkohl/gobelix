// cat/cat063/dataitems/v16/sensor_identifier.go
package v16

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SensorIdentifier implements I063/050
// Identification of the Sensor to which the provided information are related.
type SensorIdentifier struct {
	SAC uint8 // System Area Code
	SIC uint8 // System Identification Code
}

func (s *SensorIdentifier) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading sensor identifier: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for sensor identifier, have %d", asterix.ErrBufferTooShort, n)
	}

	s.SAC = data[0]
	s.SIC = data[1]

	return n, s.Validate()
}

func (s *SensorIdentifier) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	n, err := buf.Write([]byte{s.SAC, s.SIC})
	if err != nil {
		return n, fmt.Errorf("writing sensor identifier: %w", err)
	}
	return n, nil
}

func (s *SensorIdentifier) Validate() error {
	// According to the specs, SAC and SIC are simple uint8 values
	// No specific validation required for this data item
	return nil
}

func (s *SensorIdentifier) String() string {
	return fmt.Sprintf("SAC: %03d, SIC: %03d", s.SAC, s.SIC)
}
