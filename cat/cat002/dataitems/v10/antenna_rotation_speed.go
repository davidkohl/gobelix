package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// AntennaRotationSpeed represents I002/041 - Antenna Rotation Speed
type AntennaRotationSpeed struct {
	RotationPeriod float64 // Rotation period in seconds (LSB = 1/128 seconds)
}

func (a *AntennaRotationSpeed) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading antenna rotation speed: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for antenna rotation speed, have %d", asterix.ErrBufferTooShort, n)
	}
	raw := uint16(data[0])<<8 | uint16(data[1])
	a.RotationPeriod = float64(raw) / 128.0
	return 2, nil
}

func (a *AntennaRotationSpeed) Encode(buf *bytes.Buffer) (int, error) {
	raw := uint16(a.RotationPeriod * 128.0)
	if err := buf.WriteByte(byte(raw >> 8)); err != nil {
		return 0, fmt.Errorf("writing antenna rotation speed MSB: %w", err)
	}
	if err := buf.WriteByte(byte(raw & 0xFF)); err != nil {
		return 1, fmt.Errorf("writing antenna rotation speed LSB: %w", err)
	}
	return 2, nil
}

func (a *AntennaRotationSpeed) Validate() error {
	return nil
}

func (a *AntennaRotationSpeed) String() string {
	return fmt.Sprintf("%.3f s", a.RotationPeriod)
}
