// cat/cat020/dataitems/v15/track_number.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackNumber implements I020/161 - Track Number
type TrackNumber struct {
	TrackNumber uint16 // Track number (12 bits, 0-4095)
}

func (t *TrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading track number: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for track number, have %d", asterix.ErrBufferTooShort, n)
	}

	// Bits 16-13: spare
	// Bits 12-1: Track number
	t.TrackNumber = (uint16(data[0]&0x0F) << 8) | uint16(data[1])

	return n, t.Validate()
}

func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	var data [2]byte
	// Bits 16-13: spare (0)
	// Bits 12-1: Track number
	data[0] = byte((t.TrackNumber >> 8) & 0x0F)
	data[1] = byte(t.TrackNumber & 0xFF)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track number: %w", err)
	}
	return n, nil
}

func (t *TrackNumber) Validate() error {
	if t.TrackNumber > 4095 {
		return fmt.Errorf("track number out of range: %d (max 4095)", t.TrackNumber)
	}
	return nil
}

func (t *TrackNumber) String() string {
	return fmt.Sprintf("Track #%d", t.TrackNumber)
}
