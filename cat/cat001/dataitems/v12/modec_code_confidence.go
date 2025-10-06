package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeCCodeConfidence represents I001/100 - Mode-C Code and Code Confidence Indicator
type ModeCCodeConfidence struct {
	V       bool    // Validated
	G       bool    // Garbled
	FlightLevel float64 // Flight level
	QC1     bool    // Quality C1
	QA1     bool    // Quality A1
	QC2     bool    // Quality C2
	QA2     bool    // Quality A2
	QC4     bool    // Quality C4
	QA4     bool    // Quality A4
	QB1     bool    // Quality B1
	QD1     bool    // Quality D1
	QB2     bool    // Quality B2
	QD2     bool    // Quality D2
	QB4     bool    // Quality B4
	QD4     bool    // Quality D4
}

func (m *ModeCCodeConfidence) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("%w: need 4 bytes for Mode-C code confidence, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(4)

	// First two bytes: Mode-C code
	m.V = (data[0] & 0x80) != 0
	m.G = (data[0] & 0x40) != 0

	flCode := (uint16(data[0]&0x3F) << 8) | uint16(data[1])
	if flCode >= 0x2000 {
		flCode = flCode - 0x4000
	}
	m.FlightLevel = float64(int16(flCode)) / 4.0

	// Third byte: spare (4 bits), QC1, QA1, QC2, QA2
	m.QC1 = (data[2] & 0x08) != 0
	m.QA1 = (data[2] & 0x04) != 0
	m.QC2 = (data[2] & 0x02) != 0
	m.QA2 = (data[2] & 0x01) != 0

	// Fourth byte: QC4, QA4, QB1, QD1, QB2, QD2, QB4, QD4
	m.QC4 = (data[3] & 0x80) != 0
	m.QA4 = (data[3] & 0x40) != 0
	m.QB1 = (data[3] & 0x20) != 0
	m.QD1 = (data[3] & 0x10) != 0
	m.QB2 = (data[3] & 0x08) != 0
	m.QD2 = (data[3] & 0x04) != 0
	m.QB4 = (data[3] & 0x02) != 0
	m.QD4 = (data[3] & 0x01) != 0

	return 4, nil
}

func (m *ModeCCodeConfidence) Encode(buf *bytes.Buffer) (int, error) {
	// First two bytes: Mode-C code
	flCode := int16(m.FlightLevel * 4.0)
	uflCode := uint16(flCode) & 0x3FFF

	octet1 := uint8((uflCode >> 8) & 0x3F)
	if m.V {
		octet1 |= 0x80
	}
	if m.G {
		octet1 |= 0x40
	}
	octet2 := uint8(uflCode & 0xFF)

	// Third byte
	octet3 := uint8(0)
	if m.QC1 {
		octet3 |= 0x08
	}
	if m.QA1 {
		octet3 |= 0x04
	}
	if m.QC2 {
		octet3 |= 0x02
	}
	if m.QA2 {
		octet3 |= 0x01
	}

	// Fourth byte
	octet4 := uint8(0)
	if m.QC4 {
		octet4 |= 0x80
	}
	if m.QA4 {
		octet4 |= 0x40
	}
	if m.QB1 {
		octet4 |= 0x20
	}
	if m.QD1 {
		octet4 |= 0x10
	}
	if m.QB2 {
		octet4 |= 0x08
	}
	if m.QD2 {
		octet4 |= 0x04
	}
	if m.QB4 {
		octet4 |= 0x02
	}
	if m.QD4 {
		octet4 |= 0x01
	}

	buf.WriteByte(octet1)
	buf.WriteByte(octet2)
	buf.WriteByte(octet3)
	buf.WriteByte(octet4)

	return 4, nil
}

func (m *ModeCCodeConfidence) String() string {
	flags := ""
	if !m.V {
		flags += " NOT_VALIDATED"
	}
	if m.G {
		flags += " GARBLED"
	}
	return fmt.Sprintf("FL%.2f%s", m.FlightLevel, flags)
}

func (m *ModeCCodeConfidence) Validate() error {
	return nil
}
