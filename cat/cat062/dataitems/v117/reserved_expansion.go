// dataitems/cat062/reserved_expansion.go
package v117

import (
	"bytes"
	"fmt"
)

// ReservedExpansion implements RE (Reserved Expansion Field) for Cat062 (FRN 34)
// Reserved for future expansion of the data format
type ReservedExpansion struct {
	Data []byte
}

func (r *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	r.Data = nil

	// First byte is length indicator
	if buf.Len() < 1 {
		return 0, fmt.Errorf("buffer too short for reserved expansion length")
	}

	lenByte, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading reserved expansion length: %w", err)
	}
	bytesRead++
	r.Data = append(r.Data, lenByte)

	// Length is in octets
	length := int(lenByte)

	// Validate length to prevent potential buffer overruns
	if length > buf.Len() {
		return bytesRead, fmt.Errorf("buffer too short for reserved expansion data: need %d bytes, have %d",
			length, buf.Len())
	}

	// Only read what's available
	if length > 0 {
		data := make([]byte, length)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead + n, fmt.Errorf("reading reserved expansion data: %w", err)
		}

		if n != length {
			return bytesRead + n, fmt.Errorf("partial data for reserved expansion: expected %d bytes, got %d",
				length, n)
		}

		bytesRead += n
		r.Data = append(r.Data, data...)
	}

	return bytesRead, nil
}

func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if len(r.Data) == 0 {
		// If no data, encode a minimal valid value (zero length)
		return buf.Write([]byte{0})
	}

	return buf.Write(r.Data)
}

func (r *ReservedExpansion) String() string {
	if len(r.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}

	// Data starts after the length indicator
	length := int(r.Data[0])
	if len(r.Data) != length+1 {
		return fmt.Sprintf("ReservedExpansion[malformed: declared %d, actual %d]", length, len(r.Data)-1)
	}

	return fmt.Sprintf("ReservedExpansion[%d bytes]", length)
}

func (r *ReservedExpansion) Validate() error {
	// Basic validation
	if len(r.Data) == 0 {
		return nil // Empty is valid, will be encoded as {0}
	}

	// Check if length byte matches actual data
	lengthByte := r.Data[0]
	if int(lengthByte) != len(r.Data)-1 {
		return fmt.Errorf("inconsistent length: declared %d, actual %d", lengthByte, len(r.Data)-1)
	}

	return nil
}
