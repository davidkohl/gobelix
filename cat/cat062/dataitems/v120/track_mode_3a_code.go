// dataitems/cat062/track_mode_3a_code.go
package v120

import (
	"bytes"
	"fmt"
)

// TrackMode3ACode implements I062/060
// Mode-3/A code converted into octal representation
type TrackMode3ACode struct {
	Validated bool   // Code validated
	Garbled   bool   // Garbled code
	Changed   bool   // Change in Mode 3/A
	Code      uint16 // Mode-3/A reply in octal (0-7777)
}

func (t *TrackMode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading Mode 3/A code: %w", err)
	}
	if n != 2 {
		return n, fmt.Errorf("insufficient data for Mode 3/A code: got %d bytes, want 2", n)
	}

	t.Validated = (data[0] & 0x80) == 0 // bit 16, inverted
	t.Garbled = (data[0] & 0x40) != 0   // bit 15
	t.Changed = (data[0] & 0x20) != 0   // bit 14

	// Mode 3/A code is in octal representation
	// Discard bit 13 which is always 0 and extract the code
	octalDigits := []uint16{
		uint16((data[0] & 0x0E) >> 1),                             // A4, A2, A1
		uint16(((data[0] & 0x01) << 2) | ((data[1] & 0xC0) >> 6)), // B4, B2, B1
		uint16((data[1] & 0x38) >> 3),                             // C4, C2, C1
		uint16(data[1] & 0x07),                                    // D4, D2, D1
	}

	// Validate octal digits (should be 0-7)
	for i, digit := range octalDigits {
		if digit > 7 {
			return n, fmt.Errorf("invalid octal digit %d at position %d", digit, i)
		}
	}

	// Compose the octal code
	t.Code = octalDigits[0]*1000 + octalDigits[1]*100 + octalDigits[2]*10 + octalDigits[3]

	return n, nil
}

func (t *TrackMode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Decode the octal digits
	digits := []uint16{
		(t.Code / 1000) % 10,
		(t.Code / 100) % 10,
		(t.Code / 10) % 10,
		t.Code % 10,
	}

	data := make([]byte, 2)

	// Set flag bits
	if !t.Validated {
		data[0] |= 0x80 // bit 16
	}
	if t.Garbled {
		data[0] |= 0x40 // bit 15
	}
	if t.Changed {
		data[0] |= 0x20 // bit 14
	}

	// Encode the octal digits
	data[0] |= byte(digits[0] << 1)          // A4, A2, A1
	data[0] |= byte(digits[1] >> 2)          // Most significant bit of B
	data[1] |= byte((digits[1] & 0x03) << 6) // Least significant bits of B
	data[1] |= byte(digits[2] << 3)          // C4, C2, C1
	data[1] |= byte(digits[3])               // D4, D2, D1

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing Mode 3/A code: %w", err)
	}
	return n, nil
}

func (t *TrackMode3ACode) Validate() error {
	// Check if code is in octal format (each digit 0-7)
	remaining := t.Code
	for i := 0; i < 4; i++ {
		digit := remaining % 10
		if digit > 7 {
			return fmt.Errorf("invalid octal digit %d at position %d", digit, i)
		}
		remaining /= 10
	}

	if remaining > 0 {
		return fmt.Errorf("code exceeds 4 octal digits: %o", t.Code)
	}

	return nil
}

func (t *TrackMode3ACode) String() string {
	flags := ""
	if !t.Validated {
		flags += "Not Validated, "
	}
	if t.Garbled {
		flags += "Garbled, "
	}
	if t.Changed {
		flags += "Changed, "
	}

	if flags != "" {
		flags = flags[:len(flags)-2] + " - " // Remove trailing comma and space
	}

	return fmt.Sprintf("%s%04o", flags, t.Code)
}
