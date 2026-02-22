// dataitems/cat021/roll_angle.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// RollAngle implements I021/230
// Roll Angle (2 octets)
type RollAngle struct {
	Angle float64 // Roll angle in degrees, LSB = 0.01°, range ±180°
}

func (r *RollAngle) Encode(buf *bytes.Buffer) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}

	// Convert to raw value: LSB = 0.01°
	raw := int16(math.Round(r.Angle * 100))

	var data [2]byte
	data[0] = byte(raw >> 8)
	data[1] = byte(raw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing roll angle: %w", err)
	}
	return n, nil
}

func (r *RollAngle) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading roll angle: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for roll angle, have %d", asterix.ErrBufferTooShort, n)
	}

	// Combine into signed int16
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))

	// Convert to degrees
	r.Angle = float64(raw) / 100.0

	return n, r.Validate()
}

func (r *RollAngle) Validate() error {
	if r.Angle < -180.0 || r.Angle > 180.0 {
		return fmt.Errorf("roll angle out of range [-180,+180]: %.2f", r.Angle)
	}
	return nil
}

func (r *RollAngle) String() string {
	return fmt.Sprintf("%.2f°", r.Angle)
}
