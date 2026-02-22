// cat/cat023/dataitems/v13/operational_range.go
package v13

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// OperationalRange represents I023/200 - Operational Range
// Fixed length: 1 byte
// Current range of service
type OperationalRange struct {
	Range uint8 // Range in nautical miles (LSB = 1 NM)
}

// Decode decodes the Operational Range from bytes
func (o *OperationalRange) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for operational range", asterix.ErrBufferTooShort)
	}

	o.Range = data

	return 1, nil
}

// Encode encodes the Operational Range to bytes
func (o *OperationalRange) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(o.Range); err != nil {
		return 0, fmt.Errorf("writing operational range: %w", err)
	}

	return 1, nil
}

// Validate validates the Operational Range
func (o *OperationalRange) Validate() error {
	// No specific validation needed - all uint8 values are valid
	return nil
}

// String returns a string representation
func (o *OperationalRange) String() string {
	return fmt.Sprintf("%d NM", o.Range)
}
