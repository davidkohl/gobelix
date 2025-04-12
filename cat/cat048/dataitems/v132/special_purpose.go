// dataitems/cat048/special_purpose.go
package v132

import (
	"bytes"
	"fmt"
	"io"
)

// SpecialPurpose implements SP048
// Special Purpose Field
type SpecialPurpose struct {
	Data []byte
}

// Decode implements the DataItem interface
func (s *SpecialPurpose) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	lenBytes := make([]byte, 1)
	n, err := buf.Read(lenBytes)
	if err != nil {
		return n, fmt.Errorf("reading special purpose length: %w", err)
	}

	// Length is in octets
	length := int(lenBytes[0])

	// Read the data
	data := make([]byte, length)
	m, err := buf.Read(data)
	if err != nil && err != io.EOF {
		return n + m, fmt.Errorf("reading special purpose data: %w", err)
	}

	// Store length byte and data
	s.Data = append(lenBytes, data[:m]...)

	return n + m, nil
}

// Encode implements the DataItem interface
func (s *SpecialPurpose) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		// If no data, encode a minimal valid value (zero length)
		return buf.Write([]byte{0})
	}

	return buf.Write(s.Data)
}

// Validate implements the DataItem interface
func (s *SpecialPurpose) Validate() error {
	// Since this is implementation-specific, we don't validate the content
	return nil
}

// String returns a human-readable representation
func (s *SpecialPurpose) String() string {
	if len(s.Data) <= 1 {
		return "SpecialPurpose[empty]"
	}
	return fmt.Sprintf("SpecialPurpose[%d bytes]", len(s.Data)-1)
}
