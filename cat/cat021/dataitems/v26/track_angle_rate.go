// dataitems/cat021/track_angle_rate.go
package v26

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackAngleRate implements I021/165
// Track Angle Rate (Rate of Turn) (2 octets)
type TrackAngleRate struct {
	Rate float64 // Rate of turn in degrees/second, LSB = 1/32 degrees/second
}

func (t *TrackAngleRate) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Convert from degrees/second to raw value: LSB = 1/32 degrees/second
	r := math.Round(t.Rate * 32)
	if r < -512 || r > 511 {
		return 0, fmt.Errorf("track angle rate %f deg/s out of 10-bit range", t.Rate)
	}
	raw := uint16(int16(r)) & 0x03FF // bits 16-11 are spare, must be zero

	var data [2]byte
	data[0] = byte(raw >> 8)
	data[1] = byte(raw)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track angle rate: %w", err)
	}
	return n, nil
}

func (t *TrackAngleRate) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track angle rate: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for track angle rate, have %d", asterix.ErrBufferTooShort, n)
	}

	// Combine into signed int16
	raw := int16(uint16(data[0])<<8 | uint16(data[1]))

	// Convert to degrees/second
	t.Rate = float64(raw) / 32.0

	return n, t.Validate()
}

func (t *TrackAngleRate) Validate() error {
	return nil
}

func (t *TrackAngleRate) String() string {
	return fmt.Sprintf("%.3f°/s", t.Rate)
}
