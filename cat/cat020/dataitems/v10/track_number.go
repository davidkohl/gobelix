// cat/cat020/dataitems/v10/track_number.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackNumber represents I020/161 - Track Number
// Fixed length: 2 bytes
// An integer value representing a unique reference to a track record
type TrackNumber struct {
	TrackNumber uint16 // Track number (0-4095, bits 12-1)
}

// NewTrackNumber creates a new Track Number data item
func NewTrackNumber() *TrackNumber {
	return &TrackNumber{}
}

// Decode decodes the Track Number from bytes
func (t *TrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)
	// Bits 16-13 are spare (0), bits 12-1 are track number
	t.TrackNumber = binary.BigEndian.Uint16(data) & 0x0FFF

	return 2, nil
}

// Encode encodes the Track Number to bytes
func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Spare bits (16-13) are 0, track number in bits 12-1
	value := t.TrackNumber & 0x0FFF
	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return 0, fmt.Errorf("writing track number: %w", err)
	}

	return 2, nil
}

// Validate validates the Track Number
func (t *TrackNumber) Validate() error {
	if t.TrackNumber > 4095 {
		return fmt.Errorf("%w: track number must be 0-4095, got %d", asterix.ErrInvalidMessage, t.TrackNumber)
	}
	return nil
}

// String returns a string representation
func (t *TrackNumber) String() string {
	return fmt.Sprintf("%d", t.TrackNumber)
}
