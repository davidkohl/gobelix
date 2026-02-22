// cat/cat023/dataitems/v13/special_purpose.go
package v13

import (
	"bytes"
	"fmt"
	"io"

	"github.com/davidkohl/gobelix/asterix"
)

// SpecialPurpose represents SP023 - Special Purpose Field
// Variable length field for special purpose use
type SpecialPurpose struct {
	Data []byte
}

// Decode decodes the Special Purpose from bytes
func (s *SpecialPurpose) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	lenBytes := make([]byte, 1)
	n, err := buf.Read(lenBytes)
	if err != nil {
		return n, fmt.Errorf("%w: reading special purpose length", asterix.ErrBufferTooShort)
	}

	// Length is in octets
	length := int(lenBytes[0])

	// Read the data
	data := make([]byte, length)
	m, err := buf.Read(data)
	if err != nil && err != io.EOF {
		return n + m, fmt.Errorf("%w: reading special purpose data", asterix.ErrBufferTooShort)
	}

	// Store length byte and data
	s.Data = append(lenBytes, data[:m]...)

	return n + m, nil
}

// Encode encodes the Special Purpose to bytes
func (s *SpecialPurpose) Encode(buf *bytes.Buffer) (int, error) {
	if len(s.Data) == 0 {
		// If no data, encode a minimal valid value (zero length)
		return buf.Write([]byte{0})
	}

	return buf.Write(s.Data)
}

// Validate validates the Special Purpose
func (s *SpecialPurpose) Validate() error {
	return nil
}

// String returns a string representation
func (s *SpecialPurpose) String() string {
	if len(s.Data) <= 1 {
		return "SpecialPurpose[empty]"
	}
	return fmt.Sprintf("SpecialPurpose[%d bytes]", len(s.Data)-1)
}
