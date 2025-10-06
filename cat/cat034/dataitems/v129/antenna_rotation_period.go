// cat/cat034/dataitems/v129/antenna_rotation_period.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// AntennaRotationPeriod represents I034/041 - Antenna Rotation Period
// Fixed length: 2 bytes
// Antenna rotation period expressed as a multiple of 1/128 seconds
type AntennaRotationPeriod struct {
	Period float64 // Seconds
}

// NewAntennaRotationPeriod creates a new Antenna Rotation Period data item
func NewAntennaRotationPeriod() *AntennaRotationPeriod {
	return &AntennaRotationPeriod{}
}

// Decode decodes the Antenna Rotation Period from bytes
func (a *AntennaRotationPeriod) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)
	raw := uint16(data[0])<<8 | uint16(data[1])

	// LSB = 1/128 seconds
	a.Period = float64(raw) / 128.0

	return 2, nil
}

// Encode encodes the Antenna Rotation Period to bytes
func (a *AntennaRotationPeriod) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	// Convert seconds to 1/128 second units
	value := uint16(a.Period * 128.0)

	data := []byte{
		byte(value >> 8),
		byte(value & 0xFF),
	}

	n, err := buf.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing antenna rotation period: %w", err)
	}

	return n, nil
}

// Validate validates the Antenna Rotation Period
func (a *AntennaRotationPeriod) Validate() error {
	if a.Period < 0 || a.Period > 512 {
		return fmt.Errorf("%w: antenna rotation period out of range: %.3f", asterix.ErrInvalidMessage, a.Period)
	}
	return nil
}

// String returns a string representation
func (a *AntennaRotationPeriod) String() string {
	return fmt.Sprintf("%.3f s (%.1f RPM)", a.Period, 60.0/a.Period)
}
