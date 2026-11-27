package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// decodeFXChain reads a variable-length item made of one-octet parts, each
// carrying a 7-bit value in bits 8-2 and FX in bit 1. It returns one value
// per octet and always at least one on success.
func decodeFXChain(buf *bytes.Buffer, what string) ([]uint8, int, error) {
	var values []uint8
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return nil, len(values), fmt.Errorf("%w: incomplete %s", asterix.ErrBufferTooShort, what)
		}
		values = append(values, b>>1)
		if b&0x01 == 0 {
			return values, len(values), nil
		}
	}
}

// encodeFXChain writes values as one-octet parts with FX set on all but the
// last.
func encodeFXChain(buf *bytes.Buffer, values []uint8, what string) (int, error) {
	for i, v := range values {
		if v > 0x7F {
			return i, fmt.Errorf("%w: %s value %d is 7 bits, got %d", asterix.ErrInvalidMessage, what, i, v)
		}
		b := v << 1
		if i < len(values)-1 {
			b |= 0x01
		}
		if err := buf.WriteByte(b); err != nil {
			return i, fmt.Errorf("writing %s: %w", what, err)
		}
	}
	return len(values), nil
}

func check7Bit(values []uint8, what string) error {
	for i, v := range values {
		if v > 0x7F {
			return fmt.Errorf("%w: %s value %d is 7 bits, got %d", asterix.ErrInvalidMessage, what, i, v)
		}
	}
	return nil
}

var errEmpty = fmt.Errorf("%w: empty item", asterix.ErrInvalidMessage)
