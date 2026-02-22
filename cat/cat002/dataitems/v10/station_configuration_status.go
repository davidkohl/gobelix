package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// StationConfigurationStatus represents I002/050 - Station Configuration Status
type StationConfigurationStatus struct {
	Status uint8 // Configuration status bits
}

func (s *StationConfigurationStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	var data [1]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading station configuration status: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("%w: need at least 1 byte for station configuration status", asterix.ErrBufferTooShort)
	}
	bytesRead++

	// First octet: status bits (bits 8-2), FX (bit 1)
	s.Status = (data[0] >> 1) & 0x7F

	// Check FX bit for extension
	hasFX := (data[0] & 0x01) != 0

	// Handle extensions if present
	for hasFX {
		n, err = buf.Read(data[:])
		if err != nil {
			return bytesRead, fmt.Errorf("reading station configuration status extension: %w", err)
		}
		if n != 1 {
			return bytesRead, fmt.Errorf("%w: incomplete station configuration status extension", asterix.ErrBufferTooShort)
		}
		bytesRead++
		hasFX = (data[0] & 0x01) != 0
	}

	return bytesRead, nil
}

func (s *StationConfigurationStatus) Encode(buf *bytes.Buffer) (int, error) {
	// First octet: status in bits 8-2, no FX
	octet := (s.Status & 0x7F) << 1
	if err := buf.WriteByte(octet); err != nil {
		return 0, fmt.Errorf("writing station configuration status: %w", err)
	}
	return 1, nil
}

func (s *StationConfigurationStatus) Validate() error {
	return nil
}

func (s *StationConfigurationStatus) String() string {
	return fmt.Sprintf("Status: %02X", s.Status)
}
