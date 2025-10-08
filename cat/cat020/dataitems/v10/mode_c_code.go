// cat/cat020/dataitems/v10/mode_c_code.go
package v10

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ModeCCode represents I020/100 - Mode-C Code
// Fixed length: 4 bytes
// Mode-C height in Gray notation as received from the transponder together with
// the confidence level for each reply bit
type ModeCCode struct {
	V    bool   // Code validated (false) / Code not validated (true)
	G    bool   // Default (false) / Garbled code (true)
	Code uint16 // Mode-C reply in Gray notation (12 bits: C1,A1,C2,A2,C4,A4,B1,D1,B2,D2,B4,D4)
	QC1  bool   // Quality pulse C1
	QA1  bool   // Quality pulse A1
	QC2  bool   // Quality pulse C2
	QA2  bool   // Quality pulse A2
	QC4  bool   // Quality pulse C4
	QA4  bool   // Quality pulse A4
	QB1  bool   // Quality pulse B1
	QD1  bool   // Quality pulse D1 (also Q bit)
	QB2  bool   // Quality pulse B2
	QD2  bool   // Quality pulse D2
	QB4  bool   // Quality pulse B4
	QD4  bool   // Quality pulse D4
}

// NewModeCCode creates a new Mode-C Code data item
func NewModeCCode() *ModeCCode {
	return &ModeCCode{}
}

// Decode decodes the Mode-C Code from bytes
func (m *ModeCCode) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 4 {
		return 0, fmt.Errorf("%w: need 4 bytes, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(4)

	// First 2 bytes: V, G, spare, Code
	value1 := binary.BigEndian.Uint16(data[0:2])
	m.V = (value1 & 0x8000) != 0
	m.G = (value1 & 0x4000) != 0
	// Bits 14-13 are spare
	m.Code = value1 & 0x0FFF

	// Last 2 bytes: spare, Quality bits
	value2 := binary.BigEndian.Uint16(data[2:4])
	// Bits 16-13 are spare
	m.QC1 = (value2 & 0x0800) != 0
	m.QA1 = (value2 & 0x0400) != 0
	m.QC2 = (value2 & 0x0200) != 0
	m.QA2 = (value2 & 0x0100) != 0
	m.QC4 = (value2 & 0x0080) != 0
	m.QA4 = (value2 & 0x0040) != 0
	m.QB1 = (value2 & 0x0020) != 0
	m.QD1 = (value2 & 0x0010) != 0
	m.QB2 = (value2 & 0x0008) != 0
	m.QD2 = (value2 & 0x0004) != 0
	m.QB4 = (value2 & 0x0002) != 0
	m.QD4 = (value2 & 0x0001) != 0

	return 4, nil
}

// Encode encodes the Mode-C Code to bytes
func (m *ModeCCode) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// First 2 bytes
	var value1 uint16
	if m.V {
		value1 |= 0x8000
	}
	if m.G {
		value1 |= 0x4000
	}
	// Bits 14-13 are spare (0)
	value1 |= m.Code & 0x0FFF

	if err := binary.Write(buf, binary.BigEndian, value1); err != nil {
		return 0, fmt.Errorf("writing mode-C code: %w", err)
	}

	// Last 2 bytes
	var value2 uint16
	// Bits 16-13 are spare (0)
	if m.QC1 {
		value2 |= 0x0800
	}
	if m.QA1 {
		value2 |= 0x0400
	}
	if m.QC2 {
		value2 |= 0x0200
	}
	if m.QA2 {
		value2 |= 0x0100
	}
	if m.QC4 {
		value2 |= 0x0080
	}
	if m.QA4 {
		value2 |= 0x0040
	}
	if m.QB1 {
		value2 |= 0x0020
	}
	if m.QD1 {
		value2 |= 0x0010
	}
	if m.QB2 {
		value2 |= 0x0008
	}
	if m.QD2 {
		value2 |= 0x0004
	}
	if m.QB4 {
		value2 |= 0x0002
	}
	if m.QD4 {
		value2 |= 0x0001
	}

	if err := binary.Write(buf, binary.BigEndian, value2); err != nil {
		return 2, fmt.Errorf("writing quality bits: %w", err)
	}

	return 4, nil
}

// Validate validates the Mode-C Code
func (m *ModeCCode) Validate() error {
	if m.Code > 0x0FFF {
		return fmt.Errorf("%w: code must be 0-4095, got %d", asterix.ErrInvalidMessage, m.Code)
	}
	return nil
}

// String returns a string representation
func (m *ModeCCode) String() string {
	status := ""
	if m.V {
		status += "V"
	}
	if m.G {
		status += "G"
	}
	if status == "" {
		status = "OK"
	}
	return fmt.Sprintf("Code=%03X (%s)", m.Code, status)
}
