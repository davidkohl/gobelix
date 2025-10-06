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
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte for message type, have %d", asterix.ErrBufferTooShort, buf.Len())
	}
	data := buf.Next(1)
	m.MessageType = data[0]
	return 1, nil
}

func (m *MessageType) Encode(buf *bytes.Buffer) (int, error) {
	buf.WriteByte(m.MessageType)
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
