// dataitems/cat048/track_number.go
package v132

import (
	"bytes"
	"fmt"
)

// TrackNumber implements I048/161
// An integer value representing a unique reference to a track record
// within a particular track file.
type TrackNumber struct {
	Value uint16 // Track number (0-4095)
}

// Decode implements the DataItem interface
func (t *TrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading track number: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for track number: got %d bytes, want 2", n)
	}

	// First 4 bits are spare, track number is in the last 12 bits
	t.Value = uint16(data[0]&0x0F)<<8 | uint16(data[1])

	return n, nil
}

// Encode implements the DataItem interface
func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	data := make([]byte, 2)
	// First 4 bits are spare, set to 0
	data[0] = byte((t.Value >> 8) & 0x0F) // Upper 4 bits of track number
	data[1] = byte(t.Value)               // Lower 8 bits of track number

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track number: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (t *TrackNumber) Validate() error {
	if t.Value > 4095 { // 2^12 - 1
		return fmt.Errorf("track number exceeds valid range [0,4095]: %d", t.Value)
	}
	return nil
}

// String returns a human-readable representation
func (t *TrackNumber) String() string {
	return fmt.Sprintf("%d", t.Value)
}
