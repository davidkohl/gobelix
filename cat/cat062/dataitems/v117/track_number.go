// dataitems/cat062/track_number.go
package v117

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackNumber implements I062/040
// Identification of a track
type TrackNumber struct {
	Value uint16 // Track number value (0-65535)
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

	t.Value = uint16(data[0])<<8 | uint16(data[1])
	return n, nil
}

func (t *TrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	data := [2]byte{
		byte(t.Value >> 8),
		byte(t.Value),
	}

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing track number: %w", err)
	}
	return n, nil
}

func (t *TrackNumber) String() string {
	return fmt.Sprintf("%d", t.Value)
}

func (t *TrackNumber) Validate() error {
	return nil
}
