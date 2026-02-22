package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// MessageType represents I002/000 - Message Type
type MessageType struct {
	MessageType uint8 // 1=North marker, 2=Sector crossing, 3=South marker, 8=Activation of blind zone, 9=Stop of blind zone
}

func (m *MessageType) Decode(buf *bytes.Buffer) (int, error) {
	var data [1]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading message type: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("%w: need 1 byte for message type, have %d", asterix.ErrBufferTooShort, n)
	}
	m.MessageType = data[0]
	return 1, nil
}

func (m *MessageType) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(m.MessageType); err != nil {
		return 0, fmt.Errorf("writing message type: %w", err)
	}
	return 1, nil
}

func (m *MessageType) Validate() error {
	return nil
}

func (m *MessageType) String() string {
	msgTypes := map[uint8]string{
		1: "North marker",
		2: "Sector crossing",
		3: "South marker",
		8: "Activation of blind zone filtering",
		9: "Stop of blind zone filtering",
	}
	if name, ok := msgTypes[m.MessageType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", m.MessageType)
}
