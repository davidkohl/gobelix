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
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for antenna rotation speed, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	data := buf.Next(2)
	raw := uint16(data[0])<<8 | uint16(data[1])
	a.RotationPeriod = float64(raw) / 128.0
	return 2, nil
}

func (a *AntennaRotationSpeed) Encode(buf *bytes.Buffer) (int, error) {
	raw := uint16(a.RotationPeriod * 128.0)
	buf.WriteByte(byte(raw >> 8))
	buf.WriteByte(byte(raw & 0xFF))
	return 2, nil
}

func (a *AntennaRotationSpeed) Validate() error {
	return nil
}

func (a *AntennaRotationSpeed) String() string {
	return fmt.Sprintf("%.3f s", a.RotationPeriod)
}
