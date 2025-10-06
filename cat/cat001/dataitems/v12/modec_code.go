package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeCCode represents I001/090 - Mode-C Code in Binary Representation
type ModeCCode struct {
	V       bool    // Validated
	G       bool    // Garbled
	FlightLevel float64 // Flight level in 1/4 FL (LSB = 1/4 FL)
}

func (m *ModeCCode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for Mode-C code, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// First byte: bit 8 = V, bit 7 = G, bits 6-1 = high bits of flight level
	m.V = (data[0] & 0x80) != 0
	m.G = (data[0] & 0x40) != 0

	// Flight level is 14 bits (bits 6-1 of first byte + all 8 bits of second byte)
	flCode := (uint16(data[0]&0x3F) << 8) | uint16(data[1])

	// Convert to signed if necessary (14-bit two's complement)
	if flCode >= 0x2000 {
		flCode = flCode - 0x4000
	}

	m.FlightLevel = float64(int16(flCode)) / 4.0

	return 2, nil
}

func (m *ModeCCode) Encode(buf *bytes.Buffer) (int, error) {
	// Convert flight level to 1/4 FL units
	flCode := int16(m.FlightLevel * 4.0)

	// Handle negative values (14-bit two's complement)
	uflCode := uint16(flCode) & 0x3FFF

	octet1 := uint8((uflCode >> 8) & 0x3F)
	if m.V {
		octet1 |= 0x80
	}
	if m.G {
		octet1 |= 0x40
	}

	octet2 := uint8(uflCode & 0xFF)

	buf.WriteByte(octet1)
	buf.WriteByte(octet2)
	return 2, nil
}

func (m *ModeCCode) String() string {
	flags := ""
	if !m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	return fmt.Sprintf("FL%.2f%s", m.FlightLevel, flags)
}

func (m *ModeCCode) Validate() error {
	return nil
}
