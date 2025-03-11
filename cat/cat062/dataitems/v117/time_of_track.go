// dataitems/cat062/time_of_track_information.go
package v117

import (
	"bytes"
	"fmt"
	"math"
)

// TimeOfTrackInformation implements I062/070
// Absolute time stamping of the information provided in the track message,
// in the form of elapsed time since last midnight, expressed as UTC.
type TimeOfTrackInformation struct {
	Time float64 // Time in seconds since midnight
}

func (t *TimeOfTrackInformation) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 3)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading time of track information: %w", err)
	}
	if n != 3 {
		return n, fmt.Errorf("insufficient data for time of track information: got %d bytes, want 3", n)
	}

	counts := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	t.Time = float64(counts) / 128.0 // LSB = 1/128 seconds = 2^-7 seconds

	return n, nil
}

func (t *TimeOfTrackInformation) Encode(buf *bytes.Buffer) (int, error) {
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

	data := make([]byte, 3)
	data[0] = byte(counts >> 16)
	data[1] = byte(counts >> 8)
	data[2] = byte(counts)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing time of track information: %w", err)
	}
	return n, nil
}

func (t *TimeOfTrackInformation) Validate() error {
	// No validation needed here as we handle any time value during encode
	return nil
}

func (t *TimeOfTrackInformation) String() string {
	// We'll convert to a human-readable time format
	// Note: This doesn't account for time exceeding 24 hours

	// Extract hours, minutes, seconds
	seconds := math.Mod(t.Time, 86400) // Limit to 24 hours for display
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	fraction := seconds - math.Floor(seconds)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, int(fraction*1000))
}
