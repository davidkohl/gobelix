package v129

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// I034/090 quantisation (CAT034 ed 1.29 §5.2.9).
const (
	// CollimationRangeLSB is the range error LSB: 1/128 NM.
	CollimationRangeLSB = 1.0 / 128.0
	// CollimationAzimuthLSB is the azimuth error LSB: 360°/2^14 ≈ 0.022°.
	CollimationAzimuthLSB = 360.0 / 16384.0
)

// CollimationError represents I034/090 - Collimation Error (CAT034 ed 1.29
// §5.2.9). Two octets, RANGE ERROR (bits 16-9) and AZIMUTH ERROR (bits 8-1),
// two's complement: averaged difference of the primary target position
// relative to the SSR target position.
type CollimationError struct {
	RangeError   float64 // NM, LSB 1/128 NM, -1.0 .. +0.9921875 NM
	AzimuthError float64 // degrees, LSB 360/2^14, -2.8125 .. +2.7905°
}

func (c *CollimationError) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for I034/090, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	d := buf.Next(2)
	c.RangeError = float64(int8(d[0])) * CollimationRangeLSB
	c.AzimuthError = float64(int8(d[1])) * CollimationAzimuthLSB
	return 2, nil
}

func (c *CollimationError) Encode(buf *bytes.Buffer) (int, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}
	r := int8(math.Round(c.RangeError / CollimationRangeLSB))
	a := int8(math.Round(c.AzimuthError / CollimationAzimuthLSB))
	n, err := buf.Write([]byte{byte(r), byte(a)})
	if err != nil {
		return n, fmt.Errorf("writing I034/090: %w", err)
	}
	return n, nil
}

func (c *CollimationError) Validate() error {
	for _, f := range []struct {
		name string
		raw  float64
	}{
		{"range error", c.RangeError / CollimationRangeLSB},
		{"azimuth error", c.AzimuthError / CollimationAzimuthLSB},
	} {
		r := math.Round(f.raw)
		if math.IsNaN(r) || r < math.MinInt8 || r > math.MaxInt8 {
			return fmt.Errorf("%w: I034/090 %s out of the signed 8-bit range", asterix.ErrInvalidMessage, f.name)
		}
	}
	return nil
}

func (c *CollimationError) String() string {
	return fmt.Sprintf("range %.4f NM, azimuth %.4f°", c.RangeError, c.AzimuthError)
}
