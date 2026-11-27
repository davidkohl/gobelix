package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode2Code represents I001/050 - Mode-2 Code in Octal Representation
// (CAT001 ed 1.4 §5.2.6). Two octets:
//
//	V(16) G(15) L(14) spare(13) A4 A2 A1 B4 B2 B1 C4 C2 C1 D4 D2 D1 (12-1)
type Mode2Code struct {
	V    bool   // Code not validated
	G    bool   // Garbled code
	L    bool   // Smoothed Mode-2 code as provided by a local tracker
	Code uint16 // Mode-2 code, 12 bits (four octal digits ABCD)
}

func (m *Mode2Code) Decode(buf *bytes.Buffer) (int, error) {
	var data [2]byte
	n, err := buf.Read(data[:])
	if err != nil || n != 2 {
		return n, fmt.Errorf("%w: need 2 bytes for Mode-2 code, have %d", asterix.ErrBufferTooShort, n)
	}

	m.V = data[0]&0x80 != 0
	m.G = data[0]&0x40 != 0
	m.L = data[0]&0x20 != 0
	m.Code = uint16(data[0]&0x0F)<<8 | uint16(data[1])

	return 2, nil
}

func (m *Mode2Code) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}
	octet1 := uint8(m.Code>>8) & 0x0F
	if m.V {
		octet1 |= 0x80
	}
	if m.G {
		octet1 |= 0x40
	}
	if m.L {
		octet1 |= 0x20
	}

	if err := buf.WriteByte(octet1); err != nil {
		return 0, fmt.Errorf("writing Mode-2 code flags: %w", err)
	}
	if err := buf.WriteByte(uint8(m.Code)); err != nil {
		return 1, fmt.Errorf("writing Mode-2 code value: %w", err)
	}
	return 2, nil
}

func (m *Mode2Code) String() string {
	flags := ""
	if m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	if m.L {
		flags += " SMOOTHED"
	}
	return fmt.Sprintf("%04o%s", m.Code, flags)
}

func (m *Mode2Code) Validate() error {
	if m.Code > 0x0FFF {
		return fmt.Errorf("%w: Mode-2 code is 12 bits, got %#o", asterix.ErrInvalidMessage, m.Code)
	}
	return nil
}
