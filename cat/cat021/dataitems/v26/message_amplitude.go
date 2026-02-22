// dataitems/cat021/message_amplitude.go
package v26

import (
	"bytes"
	"fmt"
)

// MessageAmplitude implements I021/132
// Message Amplitude field (1 octet)
type MessageAmplitude struct {
	Amplitude int8 // Message amplitude in dBm (-127 to +127)
}

func (m *MessageAmplitude) Encode(buf *bytes.Buffer) (int, error) {
	if err := buf.WriteByte(byte(m.Amplitude)); err != nil {
		return 0, fmt.Errorf("writing message amplitude: %w", err)
	}
	return 1, nil
}

func (m *MessageAmplitude) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading message amplitude: %w", err)
	}

	m.Amplitude = int8(b)
	return 1, nil
}

func (m *MessageAmplitude) Validate() error {
	return nil
}

func (m *MessageAmplitude) String() string {
	return fmt.Sprintf("%ddBm", m.Amplitude)
}
