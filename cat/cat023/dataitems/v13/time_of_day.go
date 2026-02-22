// cat/cat023/dataitems/v13/time_of_day.go
package v13

import (
	"bytes"
	"fmt"
	"math"

	"github.com/davidkohl/gobelix/asterix"
)

// TimeOfDay represents I023/070 - Time of Day
// Fixed length: 3 bytes
// Absolute time stamping expressed as UTC time
type TimeOfDay struct {
	Time float64 // Time in seconds since midnight
}

// Decode decodes the Time of Day from bytes
func (t *TimeOfDay) Decode(buf *bytes.Buffer) (int, error) {
	var data [3]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("%w: reading time of day", asterix.ErrBufferTooShort)
	}
	if n != 3 {
		return n, fmt.Errorf("%w: need 3 bytes for time of day, got %d", asterix.ErrBufferTooShort, n)
	}

	counts := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	t.Time = float64(counts) / 128.0 // LSB = 1/128 seconds = 2^-7 seconds

	return n, nil
}

// Encode encodes the Time of Day to bytes
func (t *TimeOfDay) Encode(buf *bytes.Buffer) (int, error) {
	// Handle time wraparound to ensure it fits in 3 bytes
	adjustedTime := t.Time

	// The maximum value representable in 3 bytes (24 bits) at 1/128 second resolution
	// would be (2^24 - 1) / 128 seconds = 131071.99219 seconds ≈ 36.4 hours
	maxTime := (1<<24 - 1) / 128.0

	// Ensure the time fits in the available 3 bytes
	if adjustedTime < 0 {
		return 0, fmt.Errorf("negative time not allowed: %f", adjustedTime)
	}

	// If time exceeds maximum representable value, wrap around
	if adjustedTime > maxTime {
		adjustedTime = math.Mod(adjustedTime, maxTime)
	}

	counts := uint32(math.Round(adjustedTime * 128.0))

	var data [3]byte
	data[0] = byte(counts >> 16)
	data[1] = byte(counts >> 8)
	data[2] = byte(counts)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing time of day: %w", err)
	}
	return n, nil
}

// Validate validates the Time of Day
func (t *TimeOfDay) Validate() error {
	// No validation needed here as we handle any time value during encode
	return nil
}

// String returns a string representation
func (t *TimeOfDay) String() string {
	// Convert to a human-readable time format
	// Note: This doesn't account for time exceeding 24 hours

	// Extract hours, minutes, seconds
	seconds := math.Mod(t.Time, 86400) // Limit to 24 hours for display
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	fraction := seconds - math.Floor(seconds)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, int(fraction*1000))
}
