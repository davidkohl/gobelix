// dataitems/cat062/target_identification.go
package v120

import (
	"bytes"
	"fmt"
)

// TargetIdentificationType represents the source of target identification
type TargetIdentificationType uint8

const (
	CallsignRegistration TargetIdentificationType = iota
	CallsignNotDownlinked
	RegistrationNotDownlinked
	InvalidIdentification
)

// TargetIdentification implements I062/245
// Target (aircraft or vehicle) identification in 8 characters
type TargetIdentification struct {
	IdentType TargetIdentificationType
	Ident     string // Up to 8 characters of identification
}

// sixBitToASCII implements the ICAO Annex 10 Vol IV character set mapping
// Each 6-bit code maps to a character in this 64-character array
// '#' represents undefined/reserved codes that should not appear in valid data
var sixBitToASCII = []byte("#ABCDEFGHIJKLMNOPQRSTUVWXYZ##### ###############0123456789######")

func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 7)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading target identification: %w", err)
	}
	if n != 7 {
		return n, fmt.Errorf("insufficient data for target identification: got %d bytes, want 7", n)
	}

	// First byte contains STI (bits 56/55) and 6 spare bits
	t.IdentType = TargetIdentificationType((data[0] >> 6) & 0x03)

	// The rest contains 8 characters (6 bits each) across 6 bytes
	var chars [8]byte

	// Extract 6-bit character codes
	chars[0] = (data[1] & 0xFC) >> 2
	chars[1] = ((data[1] & 0x03) << 4) | ((data[2] & 0xF0) >> 4)
	chars[2] = ((data[2] & 0x0F) << 2) | ((data[3] & 0xC0) >> 6)
	chars[3] = data[3] & 0x3F
	chars[4] = (data[4] & 0xFC) >> 2
	chars[5] = ((data[4] & 0x03) << 4) | ((data[5] & 0xF0) >> 4)
	chars[6] = ((data[5] & 0x0F) << 2) | ((data[6] & 0xC0) >> 6)
	chars[7] = data[6] & 0x3F

	// Convert to ASCII and validate
	result := make([]byte, 8)
	for i, code := range chars {
		if int(code) >= len(sixBitToASCII) {
			return n, fmt.Errorf("invalid character code %d at position %d", code, i)
		}
		ch := sixBitToASCII[code]
		if ch == '#' {
			return n, fmt.Errorf("invalid/reserved character code %d at position %d", code, i)
		}
		result[i] = ch
	}

	// Remove any trailing spaces
	t.Ident = string(bytes.TrimRight(result[:], " "))

	return n, nil
}

func (t *TargetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Pad to 8 characters with spaces
	ident := fmt.Sprintf("%-8s", t.Ident)

	// Convert ASCII to 6-bit
	var chars [8]byte
	for i := 0; i < 8; i++ {
		found := false
		for j, ch := range sixBitToASCII {
			if ch == ident[i] {
				chars[i] = byte(j)
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("invalid character '%c' at position %d", ident[i], i)
		}
	}

	// Pack into 7 bytes
	output := make([]byte, 7)

	// First byte contains the STI
	output[0] = byte(t.IdentType) << 6

	// Pack the character codes
	output[1] = (chars[0] << 2) | (chars[1] >> 4)
	output[2] = (chars[1] << 4) | (chars[2] >> 2)
	output[3] = (chars[2] << 6) | chars[3]
	output[4] = (chars[4] << 2) | (chars[5] >> 4)
	output[5] = (chars[5] << 4) | (chars[6] >> 2)
	output[6] = (chars[6] << 6) | chars[7]

	n, err := buf.Write(output)
	if err != nil {
		return n, fmt.Errorf("writing target identification: %w", err)
	}
	return n, nil
}

func (t *TargetIdentification) Validate() error {
	if t.IdentType > InvalidIdentification {
		return fmt.Errorf("invalid identification type: %d", t.IdentType)
	}

	if len(t.Ident) > 8 {
		return fmt.Errorf("ident too long: max 8 characters, got %d", len(t.Ident))
	}

	// Check that each character is in the allowed set
	for i, ch := range t.Ident {
		found := false
		for _, validCh := range sixBitToASCII {
			if validCh == byte(ch) {
				found = true
				break
			}
		}
		if !found || byte(ch) == '#' {
			return fmt.Errorf("invalid character '%c' at position %d", ch, i)
		}
	}

	return nil
}

func (t *TargetIdentification) String() string {
	typeStr := ""
	switch t.IdentType {
	case CallsignRegistration:
		typeStr = "Callsign/Registration"
	case CallsignNotDownlinked:
		typeStr = "Callsign (not downlinked)"
	case RegistrationNotDownlinked:
		typeStr = "Registration (not downlinked)"
	case InvalidIdentification:
		typeStr = "Invalid"
	}

	return fmt.Sprintf("%s: %s", typeStr, t.Ident)
}
