// dataitems/cat048/reserved_expansion.go
package v132

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
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

	// Length indicator includes itself per ASTERIX spec:
	// "The first octet is a length indicator, giving the length of the
	// Reserved Expansion Field including the length indicator itself."
	totalLength := int(lenBytes[0])

	// Sanity check: RE fields are typically small (< 64 bytes)
	// Very large values indicate buffer misalignment from earlier items
	const maxReasonableRELength = 64
	if totalLength > maxReasonableRELength {
		return n, fmt.Errorf("reserved expansion length=%d is unreasonably high (max %d), likely buffer misalignment", totalLength, maxReasonableRELength)
	}

	// Length=0 is invalid (must at least include length byte itself)
	// Length=1 means just the length byte, no data
	if totalLength == 0 {
		return n, fmt.Errorf("reserved expansion length=0 is invalid (must be at least 1)")
	}

	dataLength := totalLength - 1 // Subtract the length byte itself

	// Check if we have enough data
	if buf.Len() < dataLength {
		return n, fmt.Errorf("%w: need %d bytes for reserved expansion, have %d", asterix.ErrBufferTooShort, dataLength, buf.Len())
	}

	// Read the data (may be 0 bytes if length=1)
	data := make([]byte, dataLength)
	m, err := buf.Read(data)
	if err != nil {
		return n + m, fmt.Errorf("reading reserved expansion data: %w", err)
	}

	// Store length byte and data
	r.Data = append(lenBytes, data...)

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
