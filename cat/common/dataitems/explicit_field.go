package common

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ExplicitField is an explicit-length data field whose content is not
// decoded: the Special Purpose field (SP) and Reserved Expansion field (RE)
// of categories that do not define their content.
//
// On the wire the first octet is the length indicator LEN, which counts
// itself (ASTERIX Part 1 §4.3.5), followed by LEN-1 octets of payload.
type ExplicitField struct {
	Payload []byte // Field content without the LEN octet
}

// Decode reads LEN and the payload. LEN = 0 is malformed.
func (e *ExplicitField) Decode(buf *bytes.Buffer) (int, error) {
	l, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading explicit field length", asterix.ErrBufferTooShort)
	}
	if l == 0 {
		return 1, fmt.Errorf("%w: explicit field LEN is 0 (it counts itself, minimum 1)", asterix.ErrInvalidLength)
	}
	n := int(l) - 1
	if buf.Len() < n {
		return 1, fmt.Errorf("%w: explicit field LEN %d needs %d octets, have %d", asterix.ErrBufferTooShort, l, n, buf.Len())
	}
	e.Payload = make([]byte, n)
	copy(e.Payload, buf.Next(n))
	return 1 + n, nil
}

// Encode writes LEN and the payload.
func (e *ExplicitField) Encode(buf *bytes.Buffer) (int, error) {
	if err := e.Validate(); err != nil {
		return 0, err
	}
	if err := buf.WriteByte(byte(len(e.Payload) + 1)); err != nil {
		return 0, fmt.Errorf("writing explicit field length: %w", err)
	}
	n, err := buf.Write(e.Payload)
	if err != nil {
		return 1 + n, fmt.Errorf("writing explicit field payload: %w", err)
	}
	return 1 + n, nil
}

// Validate checks that the payload fits a one-octet length indicator.
func (e *ExplicitField) Validate() error {
	if len(e.Payload) > 254 {
		return fmt.Errorf("%w: explicit field payload %d octets exceeds 254", asterix.ErrInvalidLength, len(e.Payload))
	}
	return nil
}

// String returns the payload in hex.
func (e *ExplicitField) String() string {
	return fmt.Sprintf("Explicit[%d]{% X}", len(e.Payload), e.Payload)
}
