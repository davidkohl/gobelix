// cat/cat020/dataitems/v10/target_identification.go
package v10

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetIdentification represents I020/245 - Target Identification
// Fixed length: 7 bytes
// Target (aircraft or vehicle) identification in 8 characters
type TargetIdentification struct {
	STI          uint8  // Source/Type Indicator (0=Callsign/registration, 1=Registration, 2=Callsign, 3=Not defined)
	Callsign     string // Target identification (8 characters)
}

// NewTargetIdentification creates a new Target Identification data item
func NewTargetIdentification() *TargetIdentification {
	return &TargetIdentification{}
}

// Decode decodes the Target Identification from bytes
func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 7 {
		return 0, fmt.Errorf("%w: need 7 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(7)

	// First byte: STI (bits 8-7), spare bits (6-1)
	t.STI = (data[0] >> 6) & 0x03

	// Characters 1-8 (6 bits each) - ICAO Annex 10 encoding
	// Extract 48 bits from bytes 1-6
	chars := make([]byte, 8)

	chars[0] = (data[1] >> 2) & 0x3F
	chars[1] = ((data[1] & 0x03) << 4) | ((data[2] >> 4) & 0x0F)
	chars[2] = ((data[2] & 0x0F) << 2) | ((data[3] >> 6) & 0x03)
	chars[3] = data[3] & 0x3F
	chars[4] = (data[4] >> 2) & 0x3F
	chars[5] = ((data[4] & 0x03) << 4) | ((data[5] >> 4) & 0x0F)
	chars[6] = ((data[5] & 0x0F) << 2) | ((data[6] >> 6) & 0x03)
	chars[7] = data[6] & 0x3F

	// Convert 6-bit codes to ASCII
	callsign := make([]byte, 8)
	for i, c := range chars {
		callsign[i] = decode6BitChar(c)
	}

	t.Callsign = strings.TrimRight(string(callsign), " ")

	return 7, nil
}

// Encode encodes the Target Identification to bytes
func (t *TargetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Pad callsign to 8 characters
	callsign := t.Callsign
	for len(callsign) < 8 {
		callsign += " "
	}

	// First byte: STI and spare bits
	data := make([]byte, 7)
	data[0] = (t.STI & 0x03) << 6

	// Encode 8 characters as 6-bit values
	chars := make([]byte, 8)
	for i := 0; i < 8; i++ {
		chars[i] = encode6BitChar(callsign[i])
	}

	// Pack 6-bit characters into bytes 1-6
	data[1] = (chars[0] << 2) | ((chars[1] >> 4) & 0x03)
	data[2] = ((chars[1] & 0x0F) << 4) | ((chars[2] >> 2) & 0x0F)
	data[3] = ((chars[2] & 0x03) << 6) | (chars[3] & 0x3F)
	data[4] = (chars[4] << 2) | ((chars[5] >> 4) & 0x03)
	data[5] = ((chars[5] & 0x0F) << 4) | ((chars[6] >> 2) & 0x0F)
	data[6] = ((chars[6] & 0x03) << 6) | (chars[7] & 0x3F)

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing target identification: %w", err)
	}

	return 7, nil
}

// Validate validates the Target Identification
func (t *TargetIdentification) Validate() error {
	if t.STI > 3 {
		return fmt.Errorf("%w: STI must be 0-3, got %d", asterix.ErrInvalidMessage, t.STI)
	}
	if len(t.Callsign) > 8 {
		return fmt.Errorf("%w: callsign must be max 8 characters, got %d", asterix.ErrInvalidMessage, len(t.Callsign))
	}
	return nil
}

// String returns a string representation
func (t *TargetIdentification) String() string {
	stiStr := ""
	switch t.STI {
	case 0:
		stiStr = "Callsign/Reg"
	case 1:
		stiStr = "Registration"
	case 2:
		stiStr = "Callsign"
	case 3:
		stiStr = "Not defined"
	}
	return fmt.Sprintf("%s (%s)", t.Callsign, stiStr)
}

// decode6BitChar converts a 6-bit value to ASCII character (ICAO Annex 10)
func decode6BitChar(c byte) byte {
	if c == 0 {
		return ' '
	} else if c >= 1 && c <= 26 {
		return 'A' + c - 1
	} else if c >= 48 && c <= 57 {
		return '0' + c - 48
	} else if c == 32 {
		return ' '
	}
	return '?'
}

// encode6BitChar converts an ASCII character to 6-bit value (ICAO Annex 10)
func encode6BitChar(c byte) byte {
	if c == ' ' {
		return 32
	} else if c >= 'A' && c <= 'Z' {
		return c - 'A' + 1
	} else if c >= '0' && c <= '9' {
		return c - '0' + 48
	}
	return 32 // Space for unknown characters
}
