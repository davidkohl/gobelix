package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TrackPlotNumber represents I001/161 - Track/Plot Number
// Two-octet fixed length data item
type TrackPlotNumber struct {
	Number uint16 // Track/Plot Number (0-65535)
}

func (t *TrackPlotNumber) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for track/plot number, have %d", asterix.ErrBufferTooShort, n)
	}

	t.Number = uint16(data[0])<<8 | uint16(data[1])
	return 2, nil
}

func (t *TrackPlotNumber) Encode(buf *bytes.Buffer) (int, error) {
	data := []byte{
		byte(t.Number >> 8),
		byte(t.Number),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track/plot number: %w", err)
	}
	return n, nil
}

func (t *TrackPlotNumber) Validate() error {
	// All values 0-65535 are valid
	return nil
}

func (t *TrackPlotNumber) String() string {
	return fmt.Sprintf("Track/Plot Number: %d", t.Number)
}
