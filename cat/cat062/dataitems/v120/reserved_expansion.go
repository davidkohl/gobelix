// dataitems/cat062/reserved_expansion.go
package v120

import (
	"bytes"
	"fmt"
	"io"
)

// ReservedExpansion implements "RE062"
// Reserved for future expansion
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

	// Length is in octets
	length := int(lenBytes[0])

	// Read the data
	data := make([]byte, length)
	m, err := buf.Read(data)
	if err != nil && err != io.EOF {
		return n + m, fmt.Errorf("reading reserved expansion data: %w", err)
	}

	// Store length byte and data
	r.Data = append(lenBytes, data[:m]...)

	return n + m, nil
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
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(r.Data)-1)
}

func (a *ReservedExpansion) Validate() error {
	return nil
}
