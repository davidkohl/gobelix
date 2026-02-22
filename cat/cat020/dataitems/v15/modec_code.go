// cat/cat020/dataitems/v15/modec_code.go
package v15

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeCCode implements I020/100 - Mode-C Code in Binary Representation
type ModeCCode struct {
	V           bool    // Validated
	G           bool    // Garbled
	FlightLevel float64 // Flight level in 1/4 FL (LSB = 1/4 FL)
	QC1         bool    // Quality C1
	QA1         bool    // Quality A1
	QC2         bool    // Quality C2
	QA2         bool    // Quality A2
	QC4         bool    // Quality C4
	QA4         bool    // Quality A4
	QB1         bool    // Quality B1
	QD1         bool    // Quality D1
	QB2         bool    // Quality B2
	QD2         bool    // Quality D2
	QB4         bool    // Quality B4
	QD4         bool    // Quality D4
}

func (m *ModeCCode) Decode(buf *bytes.Buffer) (int, error) {
	var data [4]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading mode C code: %w", err)
	}
	if n != 4 {
		return n, fmt.Errorf("%w: need 4 bytes for mode C code, have %d", asterix.ErrBufferTooShort, n)
	}

	// First two bytes: Mode-C code
	m.V = (data[0] & 0x80) != 0
	m.G = (data[0] & 0x40) != 0

	// 14-bit signed flight level
	flCode := int16((uint16(data[0]&0x3F) << 8) | uint16(data[1]))
	// Sign extend from 14 bits
	if (flCode & 0x2000) != 0 {
		flCode |= ^0x3FFF
	}
	m.FlightLevel = float64(flCode) / 4.0

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

	return n, nil
}

func (m *ModeCCode) Encode(buf *bytes.Buffer) (int, error) {
	flCode := int16(m.FlightLevel * 4.0)
	uflCode := uint16(flCode) & 0x3FFF

	var data [4]byte

	// First two bytes
	if m.V {
		data[0] |= 0x80
	}
	if m.G {
		data[0] |= 0x40
	}
	data[0] |= byte((uflCode >> 8) & 0x3F)
	data[1] = byte(uflCode & 0xFF)

	// Third byte
	if m.QC1 {
		data[2] |= 0x08
	}
	if m.QA1 {
		data[2] |= 0x04
	}
	if m.QC2 {
		data[2] |= 0x02
	}
	if m.QA2 {
		data[2] |= 0x01
	}

	// Fourth byte
	if m.QC4 {
		data[3] |= 0x80
	}
	if m.QA4 {
		data[3] |= 0x40
	}
	if m.QB1 {
		data[3] |= 0x20
	}
	if m.QD1 {
		data[3] |= 0x10
	}
	if m.QB2 {
		data[3] |= 0x08
	}
	if m.QD2 {
		data[3] |= 0x04
	}
	if m.QB4 {
		data[3] |= 0x02
	}
	if m.QD4 {
		data[3] |= 0x01
	}

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing mode C code: %w", err)
	}
	return n, nil
}

func (m *ModeCCode) Validate() error {
	return nil
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
