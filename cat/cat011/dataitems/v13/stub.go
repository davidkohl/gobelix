package v13

import (
	"bytes"
	"fmt"
)

// ExplicitStub is a placeholder for explicit data items (SP, RE)
type ExplicitStub struct {
	ID   string
	Data []byte
}

// NewExplicitStub creates a new explicit stub for the given data item ID
func NewExplicitStub(id string) *ExplicitStub {
	return &ExplicitStub{ID: id}
}

func (e *ExplicitStub) Decode(buf *bytes.Buffer) (int, error) {
	// First byte is length indicator
	length, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading %s length: %w", e.ID, err)
	}

	if length < 1 {
		return 1, nil
	}

	// Read the remaining bytes (length includes the length byte itself)
	dataLen := int(length) - 1
	e.Data = make([]byte, dataLen)
	n, err := buf.Read(e.Data)
	if err != nil {
		return 1 + n, fmt.Errorf("reading %s data: %w", e.ID, err)
	}
	if n != dataLen {
		return 1 + n, fmt.Errorf("%s: expected %d bytes, got %d", e.ID, dataLen, n)
	}

	return 1 + dataLen, nil
}

func (e *ExplicitStub) Encode(buf *bytes.Buffer) (int, error) {
	// Length includes the length byte itself
	length := byte(len(e.Data) + 1)
	if err := buf.WriteByte(length); err != nil {
		return 0, fmt.Errorf("writing %s length: %w", e.ID, err)
	}

	n, err := buf.Write(e.Data)
	if err != nil {
		return 1 + n, fmt.Errorf("writing %s data: %w", e.ID, err)
	}

	return 1 + n, nil
}

func (e *ExplicitStub) Validate() error {
	return nil
}

func (e *ExplicitStub) String() string {
	return fmt.Sprintf("%s: %d bytes", e.ID, len(e.Data))
}
