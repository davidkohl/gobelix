// dataitems/cat062/track_mode2_code.go
package v120

import (
	"bytes"
	"fmt"
)

// TrackMode2Code implements I062/120
// Mode 2 code associated to the track
type TrackMode2Code struct {
	Code uint16 // Mode 2 code in octal format
}

func (t *TrackMode2Code) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading track mode 2 code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for track mode 2 code: got %d bytes, want 2", n)
	}

	// According to spec, top 4 bits are spare, bottom 12 bits contain the code
	t.Code = uint16(data[0]&0x0F)<<8 | uint16(data[1])

	return n, nil
}

func (t *TrackMode2Code) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Ensure top 4 bits are clear as they're spare bits
	data := []byte{
		byte((t.Code >> 8) & 0x0F), // Clear top 4 bits
		byte(t.Code),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing track mode 2 code: %w", err)
	}
	return n, nil
}

func (t *TrackMode2Code) Validate() error {
	// Mode 2 code should fit in 12 bits (octal 0 to 7777)
	if t.Code > 0x0FFF {
		return fmt.Errorf("mode 2 code exceeds 12 bits: %04X", t.Code)
	}
	return nil
}

func (t *TrackMode2Code) String() string {
	return fmt.Sprintf("Mode 2: %04o", t.Code)
}
