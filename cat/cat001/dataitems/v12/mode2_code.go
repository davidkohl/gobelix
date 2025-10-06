package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode2Code represents I001/050 - Mode-2 Code in Octal Representation
type Mode2Code struct {
	V    bool   // Validated
	G    bool   // Garbled
	L    bool   // Late
	Code uint16 // Mode-2 code (12 bits)
}

func (m *Mode2Code) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for Mode-2 code, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// First byte: bits 8-5 = spare, bit 4 = V, bit 3 = G, bit 2 = L, bit 1 = spare
	m.V = (data[0] & 0x08) != 0
	m.G = (data[0] & 0x04) != 0
	m.L = (data[0] & 0x02) != 0

	// Mode-2 code in bits 12-1 of second byte
	m.Code = uint16(data[1]) & 0x0FFF

	return 2, nil
}

func (m *Mode2Code) Encode(buf *bytes.Buffer) (int, error) {
	octet1 := uint8(0)
	if m.V {
		octet1 |= 0x08
	}
	if m.G {
		octet1 |= 0x04
	}
	if m.L {
		octet1 |= 0x02
	}

	octet2 := uint8(m.Code & 0x0FFF)

	buf.WriteByte(octet1)
	buf.WriteByte(octet2)
	return 2, nil
}

func (m *Mode2Code) String() string {
	flags := ""
	if !m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	if m.L {
		flags += " LATE"
	}
	return fmt.Sprintf("%04o%s", m.Code, flags)
}

func (m *Mode2Code) Validate() error {
	return nil
}
