// dataitems/cat048/reserved_expansion.go
package v132

import (
	"bytes"
	"fmt"
	"io"
)

// ReservedExpansion implements RE048
// Reserved Expansion Field
type ReservedExpansion struct {
	Data []byte
}

// Decode implements the DataItem interface
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

// Encode implements the DataItem interface
func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if len(r.Data) == 0 {
		// If no data, encode a minimal valid value (zero length)
		return buf.Write([]byte{0})
	}

	return buf.Write(r.Data)
}

// Validate implements the DataItem interface
func (r *ReservedExpansion) Validate() error {
	// Since this is implementation-specific, we don't validate the content
	return nil
}

// String returns a human-readable representation
func (r *ReservedExpansion) String() string {
	if len(r.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(r.Data)-1)
}
