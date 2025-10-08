// dataitems/cat048/aircraft_identification.go
package v132

import (
	"bytes"
	"fmt"
)

// AircraftIdentification implements I048/240
// Aircraft identification (in 8 characters) obtained from an aircraft
// equipped with a Mode S transponder.
type AircraftIdentification struct {
	Ident string // 8-character aircraft identification
}

// Table for mapping 6-bit characters to ASCII
// Each 6-bit code maps to a character in this 64-character array
// '#' represents undefined/reserved codes that should not appear in valid data
var sixBitToASCII = []byte("#ABCDEFGHIJKLMNOPQRSTUVWXYZ##### ###############0123456789######")

// Decode implements the DataItem interface
func (a *AircraftIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading aircraft identification: %w", err)
	}
	if n != 6 {
		return n, fmt.Errorf("insufficient data for aircraft identification: got %d bytes, want 6", n)
	}

	// 8 characters encoded in 6 bytes (each character uses 6 bits)
	chars := make([]byte, 8)

	// Extract first 4 characters
	chars[0] = (data[0] & 0xFC) >> 2
	chars[1] = ((data[0] & 0x03) << 4) | ((data[1] & 0xF0) >> 4)
	chars[2] = ((data[1] & 0x0F) << 2) | ((data[2] & 0xC0) >> 6)
	chars[3] = data[2] & 0x3F

	// Extract last 4 characters
	chars[4] = (data[3] & 0xFC) >> 2
	chars[5] = ((data[3] & 0x03) << 4) | ((data[4] & 0xF0) >> 4)
	chars[6] = ((data[4] & 0x0F) << 2) | ((data[5] & 0xC0) >> 6)
	chars[7] = data[5] & 0x3F

	// Convert 6-bit codes to ASCII characters
	asciiChars := make([]byte, 8)
	for i, code := range chars {
		if int(code) >= len(sixBitToASCII) {
			return n, fmt.Errorf("invalid character code %d at position %d", code, i)
		}
		ch := sixBitToASCII[code]
		if ch == '#' {
			// Reserved/undefined code - replace with space for lenient decoding
			// This handles version differences or corrupted data gracefully
			ch = ' '
		}
		asciiChars[i] = ch
	}

	// Convert to string and trim trailing spaces
	a.Ident = string(bytes.TrimRight(asciiChars, " "))

	return n, nil
}

// Encode implements the DataItem interface
func (a *AircraftIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	// Pad with spaces to 8 characters
	ident := fmt.Sprintf("%-8s", a.Ident)

	// Convert ASCII to 6-bit codes
	var codes [8]byte
	for i := 0; i < 8; i++ {
		found := false
		for j, ch := range sixBitToASCII {
			if ch == ident[i] {
				codes[i] = byte(j)
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("invalid character '%c' at position %d", ident[i], i)
		}
	}

	// Pack 8 6-bit codes into 6 bytes
	data := make([]byte, 6)

	// First 4 characters -> first 3 bytes
	data[0] = (codes[0] << 2) | (codes[1] >> 4)
	data[1] = ((codes[1] & 0x0F) << 4) | (codes[2] >> 2)
	data[2] = ((codes[2] & 0x03) << 6) | codes[3]

	// Last 4 characters -> last 3 bytes
	data[3] = (codes[4] << 2) | (codes[5] >> 4)
	data[4] = ((codes[5] & 0x0F) << 4) | (codes[6] >> 2)
	data[5] = ((codes[6] & 0x03) << 6) | codes[7]

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing aircraft identification: %w", err)
	}
	return n, nil
}

// Validate implements the DataItem interface
func (a *AircraftIdentification) Validate() error {
	if len(a.Ident) > 8 {
		return fmt.Errorf("aircraft identification too long: %d characters (max 8)", len(a.Ident))
	}

	// Check that each character is in the valid set
	for i, ch := range a.Ident {
		found := false
		for _, validCh := range sixBitToASCII {
			if validCh != '#' && validCh == byte(ch) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid character '%c' at position %d", ch, i)
		}
	}

	return nil
}

// String returns a human-readable representation
func (a *AircraftIdentification) String() string {
	return a.Ident
}
