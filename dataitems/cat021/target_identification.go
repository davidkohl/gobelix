// dataitems/cat021/target_identification.go
package cat021

import (
	"bytes"
	"fmt"
)

// TargetIdentification implements I021/170
type TargetIdentification struct {
	Ident string
}

// sixBitToASCII implements the ICAO Annex 10 Vol IV character set mapping
// Each 6-bit code maps to a character in this 64-character array
// '#' represents undefined/reserved codes that should not appear in valid data
var sixBitToASCII = []byte("#ABCDEFGHIJKLMNOPQRSTUVWXYZ##### ###############0123456789######")

func (t *TargetIdentification) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 6)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading target identification: %w", err)
	}
	if n != 6 {
		return n, fmt.Errorf("insufficient data for target identification: got %d bytes, want 6", n)
	}

	var chars [8]byte

	// First character set - bytes 0,1,2
	chars[0] = (data[0] & 0xFC) >> 2
	chars[1] = ((data[0] & 0x03) << 4) | ((data[1] & 0xF0) >> 4)
	chars[2] = ((data[1] & 0x0F) << 2) | ((data[2] & 0xC0) >> 6)
	chars[3] = data[2] & 0x3F

	// Second character set - bytes 3,4,5
	chars[4] = (data[3] & 0xFC) >> 2
	chars[5] = ((data[3] & 0x03) << 4) | ((data[4] & 0xF0) >> 4)
	chars[6] = ((data[4] & 0x0F) << 2) | ((data[5] & 0xC0) >> 6)
	chars[7] = data[5] & 0x3F

	// Convert to ASCII and validate
	result := make([]byte, 8)
	for i, code := range chars {
		if int(code) >= len(sixBitToASCII) {
			return n, fmt.Errorf("invalid character code %d at position %d (raw bytes: %X)", code, i, data)
		}
		ch := sixBitToASCII[code]
		if ch == '#' {
			return n, fmt.Errorf("invalid/reserved character code %d at position %d (raw bytes: %X)", code, i, data)
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
		idx := bytes.IndexByte(sixBitToASCII, ident[i])
		if idx < 0 {
			return 0, fmt.Errorf("invalid character '%c' at position %d", ident[i], i)
		}
		chars[i] = byte(idx)
	}

	// Pack into 6 bytes
	output := make([]byte, 6)

	// First group
	output[0] = (chars[0] << 2) | (chars[1] >> 4)
	output[1] = (chars[1] << 4) | (chars[2] >> 2)
	output[2] = (chars[2] << 6) | chars[3]

	// Second group
	output[3] = (chars[4] << 2) | (chars[5] >> 4)
	output[4] = (chars[5] << 4) | (chars[6] >> 2)
	output[5] = (chars[6] << 6) | chars[7]

	n, err := buf.Write(output)
	if err != nil {
		return n, fmt.Errorf("writing target identification: %w", err)
	}
	return n, nil
}

func (t *TargetIdentification) Validate() error {
	if len(t.Ident) > 8 {
		return fmt.Errorf("ident too long: max 8 characters")
	}

	for i, c := range t.Ident {
		if bytes.IndexByte(sixBitToASCII, byte(c)) < 0 {
			return fmt.Errorf("invalid character '%c' at position %d", c, i)
		}
	}
	return nil
}

func (t *TargetIdentification) Id() string {
	return "I021/170"
}
