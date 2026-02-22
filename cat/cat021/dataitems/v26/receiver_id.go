// dataitems/cat021/receiver_id.go
package v26

import (
	"bytes"
	"fmt"
)

// ReceiverID implements I021/400
// Receiver ID (1 octet)
type ReceiverID struct {
	ID uint8 // Receiver identification number
}

func (r *ReceiverID) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(r.ID); err != nil {
		return 0, fmt.Errorf("writing receiver ID: %w", err)
	}
	return 1, nil
}

func (r *ReceiverID) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading receiver ID: %w", err)
	}

	r.ID = b
	return 1, nil
}

func (r *ReceiverID) Validate() error {
	return nil
}

func (r *ReceiverID) String() string {
	return fmt.Sprintf("Receiver #%d", r.ID)
}
