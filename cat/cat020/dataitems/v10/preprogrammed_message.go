// cat/cat020/dataitems/v10/preprogrammed_message.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// PreprogrammedMessage represents I020/310 - Pre-programmed Message
// Fixed length: 1 byte
// Number related to a pre-programmed message that can be transmitted by a vehicle
type PreprogrammedMessage struct {
	TRB bool  // In Trouble indicator
	MSG uint8 // Message number (7 bits)
}

// NewPreprogrammedMessage creates a new Pre-programmed Message data item
func NewPreprogrammedMessage() *PreprogrammedMessage {
	return &PreprogrammedMessage{}
}

// Decode decodes the Pre-programmed Message from bytes
func (p *PreprogrammedMessage) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(1)

	p.TRB = (data[0] & 0x80) != 0
	p.MSG = data[0] & 0x7F

	return 1, nil
}

// Encode encodes the Pre-programmed Message to bytes
func (p *PreprogrammedMessage) Encode(buf *bytes.Buffer) (int, error) {
	if err := p.Validate(); err != nil {
		return 0, err
	}

	var value byte
	if p.TRB {
		value |= 0x80
	}
	value |= p.MSG & 0x7F

	if err := buf.WriteByte(value); err != nil {
		return 0, fmt.Errorf("writing preprogrammed message: %w", err)
	}

	return 1, nil
}

// Validate validates the Pre-programmed Message
func (p *PreprogrammedMessage) Validate() error {
	if p.MSG > 127 {
		return fmt.Errorf("%w: MSG must be 0-127, got %d", asterix.ErrInvalidMessage, p.MSG)
	}
	return nil
}

// String returns a string representation
func (p *PreprogrammedMessage) String() string {
	msgStr := ""
	switch p.MSG {
	case 1:
		msgStr = "Towing aircraft"
	case 2:
		msgStr = "Follow me operation"
	case 3:
		msgStr = "Runway check"
	case 4:
		msgStr = "Emergency operation (fire, medical...)"
	case 5:
		msgStr = "Work in progress (maintenance, birds scarer, sweepers...)"
	default:
		msgStr = fmt.Sprintf("Message %d", p.MSG)
	}

	if p.TRB {
		return fmt.Sprintf("%s (IN TROUBLE)", msgStr)
	}
	return msgStr
}
