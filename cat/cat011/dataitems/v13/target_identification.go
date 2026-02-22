package v13

import (
	"bytes"
	"fmt"
	"strings"
)

// TargetIdentification implements I011/245 - Target Identification
// Definition: Target (aircraft or vehicle) identification in 8 characters.
// Format: Seven-octet fixed length Data Item
// Bits 56-55: STI (Source of Target Identification)
// Bits 54-49: spare
// Bits 48-1: 8 characters (6 bits each)
type TargetIdentification struct {
	STI        uint8  // Source: 0=downlinked, 1=callsign not downlinked, 2=registration not downlinked
	Callsign   string // 8-character callsign
}

// STI constants
const (
	STIDownlinked          uint8 = 0 // Callsign or registration downlinked from transponder
	STICallsignNotDownlink uint8 = 1 // Callsign not downlinked from transponder
	STIRegNotDownlink      uint8 = 2 // Registration not downlinked from transponder
)

// icaoCharToCode converts an ASCII character to ICAO 6-bit code
// ICAO encoding: 1-26 = A-Z, 32 = space, 48-57 = 0-9
func icaoCharToCode(c byte) byte {
	switch {
	case c >= 'A' && c <= 'Z':
		return c - 'A' + 1
	case c >= '0' && c <= '9':
		return c - '0' + 48
	case c == ' ':
		return 32
	default:
		return 32 // Default to space for unknown characters
	}
}

// icaoCodeToChar converts an ICAO 6-bit code to ASCII character
func icaoCodeToChar(code byte) byte {
	code &= 0x3F // 6 bits only
	switch {
	case code >= 1 && code <= 26:
		return 'A' + code - 1
	case code >= 48 && code <= 57:
		return '0' + code - 48
	case code == 32:
		return ' '
	default:
		return ' '
	}
}

func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading target identification: %w", err)
	}
	if n != 7 {
		return n, fmt.Errorf("target identification: expected 7 bytes, got %d", n)
	}

	// First byte: STI (bits 7-6), spare (bits 5-0)
	t.STI = (data[0] >> 6) & 0x03

	// Extract 8 characters, 6 bits each, from bits 48-1
	// Characters are packed starting from byte 1
	chars := make([]byte, 8)

	// Byte 1 bits 7-2 = char1
	chars[0] = icaoCodeToChar(data[1] >> 2)
	// Byte 1 bits 1-0 + Byte 2 bits 7-4 = char2
	chars[1] = icaoCodeToChar(((data[1] & 0x03) << 4) | (data[2] >> 4))
	// Byte 2 bits 3-0 + Byte 3 bits 7-6 = char3
	chars[2] = icaoCodeToChar(((data[2] & 0x0F) << 2) | (data[3] >> 6))
	// Byte 3 bits 5-0 = char4
	chars[3] = icaoCodeToChar(data[3] & 0x3F)
	// Byte 4 bits 7-2 = char5
	chars[4] = icaoCodeToChar(data[4] >> 2)
	// Byte 4 bits 1-0 + Byte 5 bits 7-4 = char6
	chars[5] = icaoCodeToChar(((data[4] & 0x03) << 4) | (data[5] >> 4))
	// Byte 5 bits 3-0 + Byte 6 bits 7-6 = char7
	chars[6] = icaoCodeToChar(((data[5] & 0x0F) << 2) | (data[6] >> 6))
	// Byte 6 bits 5-0 = char8
	chars[7] = icaoCodeToChar(data[6] & 0x3F)

	t.Callsign = strings.TrimRight(string(chars), " ")

	return 7, nil
}

func (t *TargetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	var data [7]byte

	// Pad callsign to 8 characters
	callsign := t.Callsign
	for len(callsign) < 8 {
		callsign += " "
	}

	// First byte: STI (bits 7-6), spare (bits 5-0)
	data[0] = (t.STI & 0x03) << 6

	// Encode 8 characters
	c := make([]byte, 8)
	for i := 0; i < 8; i++ {
		c[i] = icaoCharToCode(callsign[i])
	}

	// Pack characters
	data[1] = (c[0] << 2) | (c[1] >> 4)
	data[2] = (c[1] << 4) | (c[2] >> 2)
	data[3] = (c[2] << 6) | c[3]
	data[4] = (c[4] << 2) | (c[5] >> 4)
	data[5] = (c[5] << 4) | (c[6] >> 2)
	data[6] = (c[6] << 6) | c[7]

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing target identification: %w", err)
	}
	return n, nil
}

func (t *TargetIdentification) Validate() error {
	if t.STI > 2 {
		return fmt.Errorf("STI out of range: %d (expected 0-2)", t.STI)
	}
	if len(t.Callsign) > 8 {
		return fmt.Errorf("callsign too long: %d characters (max 8)", len(t.Callsign))
	}
	return nil
}

func (t *TargetIdentification) String() string {
	stiStr := ""
	switch t.STI {
	case STIDownlinked:
		stiStr = "downlinked"
	case STICallsignNotDownlink:
		stiStr = "callsign"
	case STIRegNotDownlink:
		stiStr = "registration"
	}
	return fmt.Sprintf("Target ID: %s (%s)", t.Callsign, stiStr)
}
