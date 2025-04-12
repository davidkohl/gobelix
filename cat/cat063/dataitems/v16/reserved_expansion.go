// cat/cat063/dataitems/v16/reserved_expansion.go
package v16

import (
	"bytes"
	"fmt"
	"io"
)

// ReservedExpansion implements "RE063"
// Reserved for future expansion or for specific applications
type ReservedExpansion struct {
	Data []byte
}

func (r *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	lenBytes := make([]byte, 1)
	n, err := buf.Read(lenBytes)
	if err != nil {
		return n, fmt.Errorf("reading reserved expansion length: %w", err)
	}

	// Length is in octets, including the length indicator itself
	length := int(lenBytes[0])
	if length < 1 {
		return n, fmt.Errorf("invalid reserved expansion length: %d", length)
	}

	// Remaining is length - 1 (we've already read the length indicator)
	remaining := length - 1
	if remaining > 0 {
		data := make([]byte, remaining)
		m, err := buf.Read(data)
		if err != nil && err != io.EOF {
			return n + m, fmt.Errorf("reading reserved expansion data: %w", err)
		}

		// Store length byte and data
		r.Data = append(lenBytes, data[:m]...)
		return n + m, nil
	}

	// Just store the length byte if no additional data
	r.Data = lenBytes
	return n, nil
}

func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if len(r.Data) == 0 {
		// If no data, encode a minimal valid value (length = 1, just the length byte)
		return buf.Write([]byte{1})
	}

	return buf.Write(r.Data)
}

func (r *ReservedExpansion) Validate() error {
	// Basic validation to ensure the length byte matches the actual length of data
	if len(r.Data) > 0 {
		declaredLen := int(r.Data[0])
		if declaredLen != len(r.Data) {
			return fmt.Errorf("reserved expansion length mismatch: declared %d, actual %d",
				declaredLen, len(r.Data))
		}
	}
	return nil
}

func (r *ReservedExpansion) String() string {
	if len(r.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(r.Data)-1)
}
