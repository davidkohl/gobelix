// cat/cat020/dataitems/v15/acceleration.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// CalculatedAcceleration implements I020/210 - Calculated Acceleration
type CalculatedAcceleration struct {
	Ax float64 // Acceleration X in m/s², LSB = 0.25 m/s²
	Ay float64 // Acceleration Y in m/s², LSB = 0.25 m/s²
}

func (c *CalculatedAcceleration) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading calculated acceleration: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for calculated acceleration, have %d", asterix.ErrBufferTooShort, n)
	}

	// 1 byte signed for Ax, LSB = 0.25 m/s²
	c.Ax = float64(int8(data[0])) * 0.25

	// 1 byte signed for Ay, LSB = 0.25 m/s²
	c.Ay = float64(int8(data[1])) * 0.25

	return n, nil
}

func (c *CalculatedAcceleration) Encode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	data[0] = byte(int8(c.Ax / 0.25))
	data[1] = byte(int8(c.Ay / 0.25))

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing calculated acceleration: %w", err)
	}
	return n, nil
}

func (c *CalculatedAcceleration) Validate() error {
	return nil
}

func (c *CalculatedAcceleration) String() string {
	return fmt.Sprintf("Ax: %.2f m/s², Ay: %.2f m/s²", c.Ax, c.Ay)
}
