// dataitems/cat048/time_of_day.go
package v132

import (
	"bytes"
	"fmt"
	"math"
	"time"
)

// TimeOfDay implements I048/140
// Absolute time stamping expressed as Co-ordinated Universal Time (UTC).
type TimeOfDay struct {
	Time float64 // Time in seconds since midnight
}

// Decode implements the DataItem interface
func (t *TimeOfDay) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 3)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading time of day: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("insufficient data for time of day: got %d bytes, want 3", n)
	}

	// Time is a 24-bit value, LSB = 1/128 seconds
	counts := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	t.Time = float64(counts) / 128.0 // Convert to seconds

	return n, t.Validate()
}

// Encode implements the DataItem interface
func (t *TimeOfDay) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Convert time to 1/128 second counts
	counts := uint32(math.Round(t.Time * 128.0))

	// Ensure it fits in 24 bits
	if counts >= 1<<24 {
		counts = (1 << 24) - 1 // Cap at maximum value
	}

	data := make([]byte, 3)
	data[0] = byte(counts >> 16) // Most significant 8 bits
	data[1] = byte(counts >> 8)  // Middle 8 bits
	data[2] = byte(counts)       // Least significant 8 bits

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing time of day: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (t *TimeOfDay) Validate() error {
	// Time should be within 0 to 24 hours (86400 seconds)
	if t.Time < 0 || t.Time >= 86400 {
		return fmt.Errorf("time of day out of valid range [0,86400): %f", t.Time)
	}
	return nil
}

// String returns a human-readable representation
func (t *TimeOfDay) String() string {
	// Convert to hours, minutes, seconds
	hours := int(t.Time) / 3600
	minutes := (int(t.Time) % 3600) / 60
	seconds := int(t.Time) % 60
	fraction := t.Time - math.Floor(t.Time)

	// Format as HH:MM:SS.mmm
	return fmt.Sprintf("%02d:%02d:%02d.%03d",
		hours, minutes, seconds, int(fraction*1000))
}

// FromTime sets the TimeOfDay from a time.Time value
func (t *TimeOfDay) FromTime(tm time.Time) {
	// Extract hours, minutes, seconds, nanoseconds
	hours := tm.Hour()
	minutes := tm.Minute()
	seconds := tm.Second()
	nanoseconds := tm.Nanosecond()

	// Calculate total seconds since midnight
	t.Time = float64(hours*3600+minutes*60+seconds) + float64(nanoseconds)/1e9
}
