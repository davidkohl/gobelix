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
	// MessageType per ed 1.29 §5.2.1: 1 North marker, 2 sector crossing,
	// 3 geographical filtering, 4 jamming strobe, 5 solar storm,
	// 6 SSR jamming strobe, 7 Mode S jamming strobe.
	MessageType uint8
}

// NewMessageType creates a new Message Type data item
func NewMessageType() *MessageType {
	return &MessageType{}
}

// Decode decodes the Message Type from bytes
func (m *MessageType) Decode(buf *bytes.Buffer) (int, error) {
	data, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: need 1 byte for message type", asterix.ErrBufferTooShort)
	}

	m.MessageType = data

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
	if m.MessageType < 1 || m.MessageType > 7 {
		return fmt.Errorf("%w: message type must be 1-7, got %d", asterix.ErrInvalidMessage, m.MessageType)
	}
	return nil
}

// String returns a string representation
func (m *MessageType) String() string {
	types := map[uint8]string{
		1: "North Marker",
		2: "Sector Crossing",
		3: "Geographical Filtering",
		4: "Jamming Strobe",
		5: "Solar Storm",
		6: "SSR Jamming Strobe",
		7: "Mode S Jamming Strobe",
	}
	if name, ok := types[m.MessageType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", m.MessageType)
}
