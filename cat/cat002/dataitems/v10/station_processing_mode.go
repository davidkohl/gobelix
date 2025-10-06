package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// StationProcessingMode represents I002/060 - Station Processing Mode
type StationProcessingMode struct {
	Mode uint8 // Processing mode bits
}

func (s *StationProcessingMode) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte for station processing mode", asterix.ErrBufferTooShort)
	}

	data := buf.Next(1)
	bytesRead++

	// First octet: mode bits (bits 8-2), FX (bit 1)
	s.Mode = (data[0] >> 1) & 0x7F

	// Check FX bit for extension
	hasFX := (data[0] & 0x01) != 0

	// Handle extensions if present
	for hasFX {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: incomplete station processing mode extension", asterix.ErrBufferTooShort)
		}
		data = buf.Next(1)
		bytesRead++
		hasFX = (data[0] & 0x01) != 0
	}

	return bytesRead, nil
}

func (s *StationProcessingMode) Encode(buf *bytes.Buffer) (int, error) {
	// First octet: mode in bits 8-2, no FX
	octet := (s.Mode & 0x7F) << 1
	buf.WriteByte(octet)
	return 1, nil
}

func (s *StationProcessingMode) Validate() error {
	return nil
}

func (s *StationProcessingMode) String() string {
	return fmt.Sprintf("Mode: %02X", s.Mode)
}
