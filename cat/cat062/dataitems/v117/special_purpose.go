// dataitems/cat062/special_purpose.go
package v117

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// SpecialPurpose implements the Special Purpose (SP) field for Cat 062
// This field is used for implementation-specific or non-standard data
// that doesn't fit within the standard ASTERIX data items.
type SpecialPurpose struct {
	// Data contains the raw bytes of the special purpose field
	Data []byte
}

// Decode parses an ASTERIX Category 062 Special Purpose field from the buffer
func (sp *SpecialPurpose) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for Special Purpose field (need at least 1 byte)")
	}

	// Read the length byte first
	lenByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading Special Purpose length: %w", err)
	}

	// The length byte includes itself, so the actual data length is length-1
	dataLen := int(lenByte) - 1
	if dataLen < 0 {
		return 1, fmt.Errorf("invalid Special Purpose length: %d", lenByte)
	}

	if buf.Len() < dataLen {
		return 1, fmt.Errorf("buffer too short for Special Purpose data: need %d bytes", dataLen)
	}

	// Read the special purpose data
	sp.Data = make([]byte, dataLen)
	n, err := buf.Read(sp.Data)
	if err != nil || n != dataLen {
		return 1 + n, fmt.Errorf("reading Special Purpose data: %w", err)
	}

	return 1 + dataLen, nil
}

// Encode serializes the Special Purpose field into the buffer
func (sp *SpecialPurpose) Encode(buf *bytes.Buffer) (int, error) {
	if sp.Data == nil {
		// Empty SP field - just write a length of 1 (the length byte itself)
		err := buf.WriteByte(1)
		if err != nil {
			return 0, fmt.Errorf("writing empty Special Purpose length: %w", err)
		}
		return 1, nil
	}

	// Calculate total length including the length byte
	totalLen := len(sp.Data) + 1

	// Ensure length fits in one byte
	if totalLen > 255 {
		return 0, fmt.Errorf("Special Purpose data too large: %d bytes (max 254)", len(sp.Data))
	}

	// Write length byte
	err := buf.WriteByte(byte(totalLen))
	if err != nil {
		return 0, fmt.Errorf("writing Special Purpose length: %w", err)
	}

	// Write data
	n, err := buf.Write(sp.Data)
	if err != nil {
		return 1, fmt.Errorf("writing Special Purpose data: %w", err)
	}

	return 1 + n, nil
}

// String returns a human-readable representation of the Special Purpose field
func (sp *SpecialPurpose) String() string {
	if sp.Data == nil || len(sp.Data) == 0 {
		return "SP[empty]"
	}

	// Format as hexadecimal for easier debugging
	return fmt.Sprintf("SP[%d bytes: %s]", len(sp.Data), hex.EncodeToString(sp.Data))
}

// Validate performs basic validation on the Special Purpose field
func (sp *SpecialPurpose) Validate() error {
	// Only check that data doesn't exceed maximum size
	if sp.Data != nil && len(sp.Data) > 254 {
		return fmt.Errorf("Special Purpose data too large: %d bytes (max 254)", len(sp.Data))
	}

	return nil
}
