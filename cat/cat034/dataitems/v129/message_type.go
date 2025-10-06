// cat/cat034/dataitems/v129/message_type.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MessageType represents I034/000 - Message Type
// Fixed length: 1 byte
type MessageType struct {
	MessageType uint8 // 1=North marker, 2=Sector crossing, 3=South marker, 4=New sector
}

// NewMessageType creates a new Message Type data item
func NewMessageType() *MessageType {
	return &MessageType{}
}

// Decode decodes the Message Type from bytes
func (m *MessageType) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)
	m.MessageType = data[0]

	return 1, nil
}

// Encode encodes the Message Type to bytes
func (m *MessageType) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	if err := buf.WriteByte(m.MessageType); err != nil {
		return 0, fmt.Errorf("writing message type: %w", err)
	}

	return 1, nil
}

// Validate validates the Message Type
func (m *MessageType) Validate() error {
	if m.MessageType < 1 || m.MessageType > 4 {
		return fmt.Errorf("%w: message type must be 1-4, got %d", asterix.ErrInvalidMessage, m.MessageType)
	}
	return nil
}

// String returns a string representation
func (m *MessageType) String() string {
	types := map[uint8]string{
		1: "North Marker",
		2: "Sector Crossing",
		3: "South Marker",
		4: "New Sector",
	}
	if name, ok := types[m.MessageType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", m.MessageType)
}
