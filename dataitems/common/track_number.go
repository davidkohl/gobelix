// cat/cat021/items/track_number.go
package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TrackNumber struct {
	Value uint16 // 0 to 4095 (12 bits used)
}

func (t *TrackNumber) Encode(buf *bytes.Buffer) error {
	if err := t.Validate(); err != nil {
		return err
	}

	data := t.Value & 0x0FFF // Ensure only 12 bits are used
	return binary.Write(buf, binary.BigEndian, data)
}

func (t *TrackNumber) Decode(buf *bytes.Buffer) error {
	var value uint16
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return fmt.Errorf("reading track number: %w", err)
	}

	t.Value = value & 0x0FFF // Extract only the 12 bits used for track number
	return nil
}

func (t *TrackNumber) Validate() error {
	if t.Value > 0x0FFF {
		return fmt.Errorf("track number exceeds maximum value (4095): %d", t.Value)
	}
	return nil
}
