// cat/cat063/dataitems/v16/special_purpose.go
package v16

import (
	"bytes"
	"fmt"
	"io"
)

// SpecialPurpose implements "SP063"
// Special Purpose Field
type SpecialPurpose struct {
	Data []byte
}

func (s *SpecialPurpose) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	lenBytes := make([]byte, 1)
	n, err := buf.Read(lenBytes)
	if err != nil {
		return n, fmt.Errorf("reading special purpose length: %w", err)
	}

	// Length is in octets, including the length indicator itself
	length := int(lenBytes[0])
	if length < 1 {
		return n, fmt.Errorf("invalid special purpose length: %d", length)
	}

	// Remaining is length - 1 (we've already read the length indicator)
	remaining := length - 1
	if remaining > 0 {
		data := make([]byte, remaining)
		m, err := buf.Read(data)
		if err != nil && err != io.EOF {
			return n + m, fmt.Errorf("reading special purpose data: %w", err)
		}

		// Store length byte and data
		s.Data = append(lenBytes, data[:m]...)
		return n + m, nil
	}

	// Just store the length byte if no additional data
	s.Data = lenBytes
	return n, nil
}

func (s *SpecialPurpose) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		// If no data, encode a minimal valid value (length = 1, just the length byte)
		return buf.Write([]byte{1})
	}

	return buf.Write(s.Data)
}

func (s *SpecialPurpose) Validate() error {
	// Basic validation to ensure the length byte matches the actual length of data
	if len(s.Data) > 0 {
		declaredLen := int(s.Data[0])
		if declaredLen != len(s.Data) {
			return fmt.Errorf("special purpose length mismatch: declared %d, actual %d",
				declaredLen, len(s.Data))
		}
	}
	return nil
}

func (s *SpecialPurpose) String() string {
	if len(s.Data) <= 1 {
		return "SpecialPurpose[empty]"
	}
	return fmt.Sprintf("SpecialPurpose[%d bytes]", len(s.Data)-1)
}
