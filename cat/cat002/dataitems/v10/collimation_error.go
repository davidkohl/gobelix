package v10

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// I002/090 quantisation (CAT002 ed 1.2 §5.2.10).
const (
	// CollimationRangeLSB is the range error LSB: 1/128 NM.
	CollimationRangeLSB = 1.0 / 128.0
	// CollimationAzimuthLSB is the default azimuth error LSB: 360°/2^(16-f)
	// with f = 2, i.e. 360°/2^14 ≈ 0.022°.
	CollimationAzimuthLSB = 360.0 / 16384.0
)

// CollimationError represents I002/090 - Collimation Error (CAT002 ed 1.2
// §5.2.10). Two octets: RANGE ERROR (bits 16-9) and AZIMUTH ERROR (bits 8-1),
// both two's complement. Averaged difference of the primary target position
// relative to the SSR target position.
type CollimationError struct {
	RangeError   float64 // NM, LSB 1/128 NM, -1.0 .. +0.9921875 NM
	AzimuthError float64 // degrees, LSB 360/2^14 (f = 2), -2.8125 .. +2.7905°
}

func (c *CollimationError) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for collimation error, have %d", asterix.ErrBufferTooShort, buf.Len())
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
		return n, fmt.Errorf("writing collimation error: %w", err)
	}
	return n, nil
}

func (c *CollimationError) Validate() error {
	if err := checkInt8(c.RangeError/CollimationRangeLSB, "range error"); err != nil {
		return err
	}
	return checkInt8(c.AzimuthError/CollimationAzimuthLSB, "azimuth error")
}

func (c *CollimationError) String() string {
	return fmt.Sprintf("range %.4f NM, azimuth %.4f°", c.RangeError, c.AzimuthError)
}

func checkInt8(raw float64, what string) error {
	r := math.Round(raw)
	if math.IsNaN(r) || r < math.MinInt8 || r > math.MaxInt8 {
		return fmt.Errorf("%w: %s out of the signed 8-bit range", asterix.ErrInvalidMessage, what)
	}
	return nil
}
