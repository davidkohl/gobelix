// cat/cat023/dataitems/v13/reserved_expansion.go
package v13

import (
	"bytes"
	"fmt"
	"io"

	"github.com/davidkohl/gobelix/asterix"
)

// ReservedExpansion represents RE023 - Reserved Expansion Field
// Variable length field reserved for future use
type ReservedExpansion struct {
	Data []byte
}

// Decode decodes the Reserved Expansion from bytes
func (r *ReservedExpansion) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	lenBytes := make([]byte, 1)
	n, err := buf.Read(lenBytes)
	if err != nil {
		return n, fmt.Errorf("%w: reading reserved expansion length", asterix.ErrBufferTooShort)
	}

	// Length is in octets
	length := int(lenBytes[0])

	// Read the data
	data := make([]byte, length)
	m, err := buf.Read(data)
	if err != nil && err != io.EOF {
		return n + m, fmt.Errorf("%w: reading reserved expansion data", asterix.ErrBufferTooShort)
	}

	// Store length byte and data
	r.Data = append(lenBytes, data[:m]...)

	return n + m, nil
}

// Encode encodes the Reserved Expansion to bytes
func (r *ReservedExpansion) Encode(buf *bytes.Buffer) (int, error) {
	if len(r.Data) == 0 {
		// If no data, encode a minimal valid value (zero length)
		return buf.Write([]byte{0})
	}

	return buf.Write(r.Data)
}

// Validate validates the Reserved Expansion
func (r *ReservedExpansion) Validate() error {
	return nil
}

// String returns a string representation
func (r *ReservedExpansion) String() string {
	if len(r.Data) <= 1 {
		return "ReservedExpansion[empty]"
	}
	return fmt.Sprintf("ReservedExpansion[%d bytes]", len(r.Data)-1)
}
