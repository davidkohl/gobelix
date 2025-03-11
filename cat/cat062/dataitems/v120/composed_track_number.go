// dataitems/cat062/composed_track_number.go
package v120

import (
	"bytes"
	"fmt"
)

// ComposedTrackNumber implements I062/510
// Identification of a system track, extendible data item
type ComposedTrackNumber struct {
	Data []byte
}

func (c *ComposedTrackNumber) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	c.Data = nil

	// First part is 3 octets (Master Track Number)
	masterTrack := make([]byte, 3)
	n, err := buf.Read(masterTrack)
	if err != nil {
		return n, fmt.Errorf("reading composed track number master track: %w", err)
	}
	bytesRead += n
	c.Data = append(c.Data, masterTrack...)

	hasExtension := (masterTrack[2] & 0x01) != 0

	// Read additional 3-octet slave track numbers as long as FX bit is set
	for hasExtension {
		slaveTrack := make([]byte, 3)
		n, err := buf.Read(slaveTrack)
		if err != nil {
			return bytesRead, fmt.Errorf("reading composed track number slave track: %w", err)
		}
		bytesRead += n
		c.Data = append(c.Data, slaveTrack...)

		// Check if further extension exists
		hasExtension = (slaveTrack[2] & 0x01) != 0
	}

	return bytesRead, nil
}

func (c *ComposedTrackNumber) Encode(buf *bytes.Buffer) (int, error) {
	if len(c.Data) == 0 {
		// If no data, encode a minimal valid value (just system unit identification and track number)
		return buf.Write([]byte{0, 0, 0})
	}
	return buf.Write(c.Data)
}

func (c *ComposedTrackNumber) String() string {
	return fmt.Sprintf("ComposedTrackNumber[%d bytes]", len(c.Data))
}

func (a *ComposedTrackNumber) Validate() error {
	return nil
}
