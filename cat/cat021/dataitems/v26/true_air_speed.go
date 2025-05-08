// TrueAirSpeed implements I021/151
package v26

import (
	"bytes"
	"fmt"
)

// TrueAirSpeed implements I021/151
type TrueAirSpeed struct {
	RE    bool    // Range Exceeded indicator
	Speed float64 // Speed in knots
}

func (t *TrueAirSpeed) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading true air speed: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for true air speed: got %d bytes, want 2", n)
	}

	// Extract RE bit (bit 16)
	t.RE = (data[0] & 0x80) != 0

	// Extract speed value from remaining bits
	rawSpeed := uint16(data[0]&0x7F)<<8 | uint16(data[1])

	// Speed with LSB = 1 knot
	t.Speed = float64(rawSpeed)

	return n, t.Validate()
}

func (t *TrueAirSpeed) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Convert speed to raw value
	rawSpeed := uint16(t.Speed)

	// Prepare the data bytes
	data := make([]byte, 2)

	// Set the RE bit if needed
	if t.RE {
		data[0] |= 0x80
	}

	// Set the speed bits
	data[0] |= byte(rawSpeed>>8) & 0x7F
	data[1] = byte(rawSpeed)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing true air speed: %w", err)
	}
	return n, nil
}

func (t *TrueAirSpeed) Validate() error {
	// True Air Speed should be non-negative
	// The maximum value is 16383 knots (all 15 bits set)
	if t.Speed < 0 {
		return fmt.Errorf("true air speed cannot be negative: %f", t.Speed)
	}

	// If RE bit is not set, check that the value fits in the range
	if !t.RE && t.Speed > 16383 {
		return fmt.Errorf("true air speed exceeds maximum value without RE bit: %f", t.Speed)
	}

	return nil
}

func (t *TrueAirSpeed) String() string {
	if t.RE {
		return fmt.Sprintf(">%.1f kts", t.Speed)
	}
	return fmt.Sprintf("%.1f kts", t.Speed)
}
