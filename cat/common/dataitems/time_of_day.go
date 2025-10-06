// cat/common/dataitems/time_of_day.go
package common

import (
	"bytes"
	"fmt"
)

// TimeOfDay represents time of day in ASTERIX format
// Fixed length: 3 bytes
// Absolute time stamping expressed as UTC time
// LSB = 1/128 second
type TimeOfDay struct {
	TimeOfDay float64 // Seconds since midnight UTC
}

// Decode decodes TimeOfDay from bytes
func (t *TimeOfDay) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 3 {
		return 0, fmt.Errorf("need 3 bytes for TimeOfDay, have %d", buf.Len())
	}

	data := buf.Next(3)

	// 3 bytes, LSB = 1/128 second
	raw := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])

	// Convert to seconds
	t.TimeOfDay = float64(raw) / 128.0

	return 3, nil
}

// Encode encodes TimeOfDay to bytes
func (t *TimeOfDay) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Convert seconds to 1/128 second units
	raw := uint32(t.TimeOfDay * 128.0)

	// Write 3 bytes
	data := []byte{
		byte(raw >> 16),
		byte(raw >> 8),
		byte(raw),
	}

	n, err := buf.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing time of day: %w", err)
	}

	return n, nil
}

// Validate validates the TimeOfDay
func (t *TimeOfDay) Validate() error {
	// Time must be between 0 and 86400 seconds (24 hours)
	if t.TimeOfDay < 0 || t.TimeOfDay >= 86400 {
		return fmt.Errorf("time of day must be 0-86400 seconds, got %.3f", t.TimeOfDay)
	}
	return nil
}

// String returns a string representation
func (t *TimeOfDay) String() string {
	hours := int(t.TimeOfDay / 3600)
	minutes := int((t.TimeOfDay - float64(hours*3600)) / 60)
	seconds := t.TimeOfDay - float64(hours*3600) - float64(minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
}
