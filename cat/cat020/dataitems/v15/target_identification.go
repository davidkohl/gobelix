// cat/cat020/dataitems/v15/target_identification.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// sixBitToASCII implements the ICAO Annex 10 Vol IV character set mapping
// Each 6-bit code maps to a character in this 64-character array
// ' ' (space) represents undefined/reserved codes
var sixBitToASCII = []byte(" ABCDEFGHIJKLMNOPQRSTUVWXYZ      ###############0123456789######")

// TargetIdentification implements I020/245 - Target Identification
type TargetIdentification struct {
	STI      uint8  // Source/Type Indicator (0-3)
	Callsign string // Aircraft callsign (8 characters, 6-bit ICAO)
}

func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading target identification: %w", err)
	}
	if n != 7 {
		return n, fmt.Errorf("%w: need 7 bytes for target identification, have %d", asterix.ErrBufferTooShort, n)
	}

	// First byte: STI (2 bits) + spare (6 bits)
	t.STI = (data[0] >> 6) & 0x03

	// Remaining 6 bytes: 8 characters x 6 bits = 48 bits
	var chars [8]byte
	chars[0] = (data[1] >> 2) & 0x3F
	chars[1] = ((data[1] & 0x03) << 4) | ((data[2] >> 4) & 0x0F)
	chars[2] = ((data[2] & 0x0F) << 2) | ((data[3] >> 6) & 0x03)
	chars[3] = data[3] & 0x3F
	chars[4] = (data[4] >> 2) & 0x3F
	chars[5] = ((data[4] & 0x03) << 4) | ((data[5] >> 4) & 0x0F)
	chars[6] = ((data[5] & 0x0F) << 2) | ((data[6] >> 6) & 0x03)
	chars[7] = data[6] & 0x3F

	// Convert 6-bit ICAO codes to ASCII using lookup table
	var result [8]byte
	for i, code := range chars {
		if int(code) >= len(sixBitToASCII) {
			result[i] = ' ' // Invalid code, replace with space
		} else {
			ch := sixBitToASCII[code]
			if ch == '#' {
				result[i] = ' ' // Reserved code, replace with space
			} else {
				result[i] = ch
			}
		}
	}

	t.Callsign = string(bytes.TrimRight(result[:], " "))

	return n, nil
}

func (t *TargetIdentification) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	var data [7]byte

	// First byte: STI (2 bits) + spare (6 bits)
	data[0] = (t.STI & 0x03) << 6

	// Pad callsign to 8 characters with spaces
	callsign := fmt.Sprintf("%-8s", t.Callsign)
	if len(callsign) > 8 {
		callsign = callsign[:8]
	}

	// Convert ASCII to 6-bit ICAO codes using reverse lookup
	var chars [8]byte
	for i := 0; i < 8; i++ {
		c := callsign[i]
		found := false
		for j, ch := range sixBitToASCII {
			if ch == c {
				chars[i] = byte(j)
				found = true
				break
			}
		}
		if !found {
			chars[i] = 0 // Default to space (code 0)
		}
	}

	// Pack 8 x 6-bit characters into 6 bytes
	data[1] = (chars[0] << 2) | ((chars[1] >> 4) & 0x03)
	data[2] = ((chars[1] & 0x0F) << 4) | ((chars[2] >> 2) & 0x0F)
	data[3] = ((chars[2] & 0x03) << 6) | (chars[3] & 0x3F)
	data[4] = (chars[4] << 2) | ((chars[5] >> 4) & 0x03)
	data[5] = ((chars[5] & 0x0F) << 4) | ((chars[6] >> 2) & 0x0F)
	data[6] = ((chars[6] & 0x03) << 6) | (chars[7] & 0x3F)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing target identification: %w", err)
	}
	return n, nil
}

func (t *TargetIdentification) Validate() error {
	if t.STI > 3 {
		return fmt.Errorf("STI out of range: %d", t.STI)
	}
	return nil
}

func (t *TargetIdentification) String() string {
	stiStr := ""
	switch t.STI {
	case 0:
		stiStr = "Callsign"
	case 1:
		stiStr = "Registration"
	case 2:
		stiStr = "Mode S"
	case 3:
		stiStr = "Reserved"
	}
	return fmt.Sprintf("%s: %s", stiStr, t.Callsign)
}
