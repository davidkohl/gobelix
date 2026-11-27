package v12

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// DopplerSpeedLSB is the default quantisation of I001/120 in knots:
// 2^(-14+f) NM/s with the default scaling factor f = 6, i.e. 2^-8 NM/s
// = 14.0625 kt (CAT001 ed 1.4 §5.2.12, note 1).
const DopplerSpeedLSB = 3600.0 / 256.0

// MeasuredRadialDopplerSpeed represents I001/120 - Measured Radial Doppler
// Speed (CAT001 ed 1.4 §5.2.12). One octet, two's complement.
type MeasuredRadialDopplerSpeed struct {
	// DopplerSpeedKt is the speed in knots, using the default scaling f = 6
	// (LSB 14.0625 kt, range -1800 .. +1785.9 kt). Other scaling factors
	// can be applied by the caller: speed_kt = int8(raw) * 3600 * 2^(-14+f).
	// (Named with its unit since the field changed from m/s to knots.)
	DopplerSpeedKt float64
}

func (m *MeasuredRadialDopplerSpeed) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for doppler speed", asterix.ErrBufferTooShort)
	}
	m.DopplerSpeedKt = float64(int8(data)) * DopplerSpeedLSB
	return 1, nil
}

func (m *MeasuredRadialDopplerSpeed) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}
	raw := int8(math.Round(m.DopplerSpeedKt / DopplerSpeedLSB))
	if err := buf.WriteByte(byte(raw)); err != nil {
		return 0, fmt.Errorf("writing doppler speed: %w", err)
	}
	return 1, nil
}

func (m *MeasuredRadialDopplerSpeed) String() string {
	return fmt.Sprintf("%.1f kt", m.DopplerSpeedKt)
}

func (m *MeasuredRadialDopplerSpeed) Validate() error {
	raw := math.Round(m.DopplerSpeedKt / DopplerSpeedLSB)
	if math.IsNaN(raw) || raw < -128 || raw > 127 {
		return fmt.Errorf("%w: doppler speed %.1f kt outside -1800..1785.9 kt", asterix.ErrInvalidMessage, m.DopplerSpeedKt)
	}
	return nil
}
