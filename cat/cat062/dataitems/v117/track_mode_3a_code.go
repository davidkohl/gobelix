// dataitems/cat062/track_mode_3a_code.go
package v117

import (
	"bytes"
	"fmt"
)

// TrackMode3ACode implements I062/060
// Mode-3/A code converted into octal representation
type TrackMode3ACode struct {
	// V - Code validated (false) or not validated (true)
	CodeNotValidated bool

	// G - Default (false) or Garbled Code (true)
	GarbledCode bool

	// CH - No Change (false) or Mode 3/A has changed (true)
	Changed bool

	// Mode 3/A code in octal representation (range: 0000-7777)
	// Each digit occupies 3 bits, for a total of 12 bits
	Code uint16

	// Raw data for reporting purposes
	rawData []byte
}

// Decode parses an ASTERIX Category 062 I060 data item from the buffer
func (t *TrackMode3ACode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("buffer too short for Track Mode 3/A Code")
	}

	data := make([]byte, 2)
	n, err := buf.Read(data)
	if err != nil || n != 2 {
		return n, fmt.Errorf("reading Track Mode 3/A Code: %w", err)
	}

	t.rawData = make([]byte, 2)
	copy(t.rawData, data)

	// Parse the flags from the first byte
	t.CodeNotValidated = (data[0] & 0x80) != 0 // bit 16
	t.GarbledCode = (data[0] & 0x40) != 0      // bit 15
	t.Changed = (data[0] & 0x20) != 0          // bit 14

	// Extract the Mode 3/A code (bits 12-1)
	// The code is stored in 12 bits across the two bytes
	// First byte: bits 4-1 (lower 4 bits)
	// Second byte: all 8 bits
	codeValue := uint16(data[0]&0x0F)<<8 | uint16(data[1])

	// Store the raw code value (12 bits)
	t.Code = codeValue

	return n, nil
}

// Encode serializes the Track Mode 3/A Code into the buffer
func (t *TrackMode3ACode) Encode(buf *bytes.Buffer) (int, error) {
	// If we have raw data, just send it back
	if len(t.rawData) == 2 {
		return buf.Write(t.rawData)
	}

	// Prepare the two bytes
	data := [2]byte{}

	// Set the flags in the first byte
	if t.CodeNotValidated {
		data[0] |= 0x80 // bit 16
	}
	if t.GarbledCode {
		data[0] |= 0x40 // bit 15
	}
	if t.Changed {
		data[0] |= 0x20 // bit 14
	}

	// Ensure the code is only 12 bits (0-4095)
	codeValue := t.Code & 0x0FFF

	// Set the Mode 3/A code (bits 12-1)
	data[0] |= byte(codeValue >> 8) // bits 12-9 (high 4 bits of the code)
	data[1] = byte(codeValue)       // bits 8-1 (low 8 bits of the code)

	return buf.Write(data[:])
}

// String returns a human-readable representation of the Track Mode 3/A Code
func (t *TrackMode3ACode) String() string {
	// Extract the 4 octal digits (each 3 bits) from the 12-bit code
	a := (t.Code >> 9) & 0x07 // bits 12-10 (A)
	b := (t.Code >> 6) & 0x07 // bits 9-7 (B)
	c := (t.Code >> 3) & 0x07 // bits 6-4 (C)
	d := t.Code & 0x07        // bits 3-1 (D)

	// Format as 4 octal digits with flags
	codeStr := fmt.Sprintf("%o%o%o%o", a, b, c, d)

	flags := ""
	if t.CodeNotValidated {
		flags += "V"
	}
	if t.GarbledCode {
		flags += "G"
	}
	if t.Changed {
		flags += "CH"
	}

	if flags != "" {
		return fmt.Sprintf("%s[%s]", codeStr, flags)
	}
	return codeStr
}

// Validate performs validation on the Track Mode 3/A Code
func (t *TrackMode3ACode) Validate() error {
	// Ensure the code fits in 12 bits (0-4095)
	if t.Code > 0x0FFF {
		return fmt.Errorf("mode 3/A code exceeds 12-bit limit: %d", t.Code)
	}

	// Check if all digits are octal (0-7)
	// This is redundant since we're using a 12-bit uint16, but it's a good sanity check
	a := (t.Code >> 9) & 0x07 // bits 12-10 (A)
	b := (t.Code >> 6) & 0x07 // bits 9-7 (B)
	c := (t.Code >> 3) & 0x07 // bits 6-4 (C)
	d := t.Code & 0x07        // bits 3-1 (D)

	if a > 7 || b > 7 || c > 7 || d > 7 {
		return fmt.Errorf("invalid octal digit in Mode 3/A code: %04o", t.Code)
	}

	return nil
}
