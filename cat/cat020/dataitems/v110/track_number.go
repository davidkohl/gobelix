// cat/cat020/dataitems/v110/track_number.go
package v110

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackNumber represents I020/161 - Track Number
// Fixed length: 2 bytes
// An integer value representing a unique reference to a track record
// within a particular track file.
type TrackNumber struct {
	TrackNumber uint16 // Track number (0-65535)
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
	t.TrackNumber = binary.BigEndian.Uint16(data)

	return 2, nil
}

// Encode encodes the Track Number to bytes
func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	if err := binary.Write(buf, binary.BigEndian, t.TrackNumber); err != nil {
		return 0, fmt.Errorf("writing track number: %w", err)
	}

	return 2, nil
}

// Validate validates the Track Number
func (t *TrackNumber) Validate() error {
	// All values 0-65535 are valid
	return nil
}

// String returns a string representation
func (t *TrackNumber) String() string {
	return fmt.Sprintf("%d", t.TrackNumber)
}
