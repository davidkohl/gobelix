// cat/common/dataitems/track_number.go
package common

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackNumber implements the track number data item
// Used in multiple categories: I001/161, I020/161, I048/161, I062/040
// Fixed length: 2 bytes (first 4 bits spare, 12 bits for track number)
type TrackNumber struct {
	Value uint16 // Track number (0-4095)
}

// Decode decodes the Track Number from bytes
func (t *TrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track number: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for track number, have %d", asterix.ErrBufferTooShort, n)
	}

	// First 4 bits are spare, track number is in the last 12 bits
	t.Value = uint16(data[0]&0x0F)<<8 | uint16(data[1])

	return n, nil
}

// Encode encodes the Track Number to bytes
func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	data := []byte{
		byte((t.Value >> 8) & 0x0F), // Upper 4 bits (spare bits are 0)
		byte(t.Value),               // Lower 8 bits
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track number: %w", err)
	}
	return n, nil
}

// Validate validates the Track Number
func (t *TrackNumber) Validate() error {
	if t.Value > 4095 {
		return fmt.Errorf("%w: track number exceeds valid range [0,4095]: %d", asterix.ErrInvalidMessage, t.Value)
	}
	return nil
}

// String returns a human-readable representation
func (t *TrackNumber) String() string {
	return fmt.Sprintf("%d", t.Value)
}
