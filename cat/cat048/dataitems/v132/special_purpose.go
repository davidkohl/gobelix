// dataitems/cat048/special_purpose.go
package v132

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
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

	// Length indicator includes itself per ASTERIX spec:
	// "The first octet is a length indicator, giving the length of the
	// Special Purpose Field including the length indicator itself."
	totalLength := int(lenBytes[0])

	// Sanity check: SP fields are typically small (< 64 bytes)
	// Very large values indicate buffer misalignment from earlier items
	const maxReasonableSPLength = 64
	if totalLength > maxReasonableSPLength {
		return n, fmt.Errorf("special purpose length=%d is unreasonably high (max %d), likely buffer misalignment", totalLength, maxReasonableSPLength)
	}

	// Length=0 is invalid (must at least include length byte itself)
	// Length=1 means just the length byte, no data
	if totalLength == 0 {
		return n, fmt.Errorf("special purpose length=0 is invalid (must be at least 1)")
	}

	dataLength := totalLength - 1 // Subtract the length byte itself

	// Check if we have enough data
	if buf.Len() < dataLength {
		return n, fmt.Errorf("%w: need %d bytes for special purpose, have %d", asterix.ErrBufferTooShort, dataLength, buf.Len())
	}

	// Read the data (may be 0 bytes if length=1)
	data := make([]byte, dataLength)
	m, err := buf.Read(data)
	if err != nil {
		return n + m, fmt.Errorf("reading special purpose data: %w", err)
	}

	// Store length byte and data
	s.Data = append(lenBytes, data...)

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
