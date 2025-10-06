package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MeasuredRadialDopplerSpeed represents I001/120 - Measured Radial Doppler Speed
type MeasuredRadialDopplerSpeed struct {
	DopplerSpeed float64 // Doppler speed in m/s (LSB = 1 m/s, signed)
}

func (m *MeasuredRadialDopplerSpeed) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte for doppler speed, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)

	// Signed 8-bit value
	speed := int8(data[0])
	m.DopplerSpeed = float64(speed)

	return 1, nil
}

func (m *MeasuredRadialDopplerSpeed) Encode(buf *bytes.Buffer) (int, error) {
	speed := int8(m.DopplerSpeed)
	buf.WriteByte(byte(speed))
	return 1, nil
}

func (m *MeasuredRadialDopplerSpeed) String() string {
	return fmt.Sprintf("%.0f m/s", m.DopplerSpeed)
}

func (m *MeasuredRadialDopplerSpeed) Validate() error {
	return nil
}
