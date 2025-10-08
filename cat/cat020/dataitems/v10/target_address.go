// cat/cat020/dataitems/v10/target_address.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetAddress represents I020/220 - Target Address
// Fixed length: 3 bytes
// Target address (ICAO 24-bit address) assigned uniquely to each Target
type TargetAddress struct {
	Address uint32 // 24-bit ICAO address
}

// NewTargetAddress creates a new Target Address data item
func NewTargetAddress() *TargetAddress {
	return &TargetAddress{}
}

// Decode decodes the Target Address from bytes
func (t *TargetAddress) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 3 {
		return 0, fmt.Errorf("%w: need 3 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(3)

	// 24-bit address stored in 3 bytes
	t.Address = uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])

	return 3, nil
}

// Encode encodes the Target Address to bytes
func (t *TargetAddress) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	// Write 24-bit address as 3 bytes
	data := []byte{
		byte((t.Address >> 16) & 0xFF),
		byte((t.Address >> 8) & 0xFF),
		byte(t.Address & 0xFF),
	}

	n, err := buf.Write(data)
	if err != nil {
		return n, fmt.Errorf("writing target address: %w", err)
	}

	return 3, nil
}

// Validate validates the Target Address
func (t *TargetAddress) Validate() error {
	if t.Address > 0xFFFFFF {
		return fmt.Errorf("%w: address must be 24-bit (0-16777215), got %d", asterix.ErrInvalidMessage, t.Address)
	}
	return nil
}

// String returns a string representation
func (t *TargetAddress) String() string {
	return fmt.Sprintf("%06X", t.Address)
}
