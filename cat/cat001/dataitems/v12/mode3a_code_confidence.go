package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// Mode3ACodeConfidence represents I001/080 - Mode-3/A Code Confidence Indicator
type Mode3ACodeConfidence struct {
	QA4 bool // Quality pulse A4
	QA2 bool // Quality pulse A2
	QA1 bool // Quality pulse A1
	QB4 bool // Quality pulse B4
	QB2 bool // Quality pulse B2
	QB1 bool // Quality pulse B1
	QC4 bool // Quality pulse C4
	QC2 bool // Quality pulse C2
	QC1 bool // Quality pulse C1
	QD4 bool // Quality pulse D4
	QD2 bool // Quality pulse D2
	QD1 bool // Quality pulse D1
}

func (m *Mode3ACodeConfidence) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 2 {
		return 0, fmt.Errorf("%w: need 2 bytes for Mode-3/A code confidence, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	data := buf.Next(2)

	// First byte: spare (bit 8-5), QA4, QA2, QA1, QB4
	m.QA4 = (data[0] & 0x08) != 0
	m.QA2 = (data[0] & 0x04) != 0
	m.QA1 = (data[0] & 0x02) != 0
	m.QB4 = (data[0] & 0x01) != 0

	// Second byte: QB2, QB1, QC4, QC2, QC1, QD4, QD2, QD1
	m.QB2 = (data[1] & 0x80) != 0
	m.QB1 = (data[1] & 0x40) != 0
	m.QC4 = (data[1] & 0x20) != 0
	m.QC2 = (data[1] & 0x10) != 0
	m.QC1 = (data[1] & 0x08) != 0
	m.QD4 = (data[1] & 0x04) != 0
	m.QD2 = (data[1] & 0x02) != 0
	m.QD1 = (data[1] & 0x01) != 0

	return 2, nil
}

func (m *Mode3ACodeConfidence) Encode(buf *bytes.Buffer) (int, error) {
	octet1 := uint8(0)
	if m.QA4 {
		octet1 |= 0x08
	}
	if m.QA2 {
		octet1 |= 0x04
	}
	if m.QA1 {
		octet1 |= 0x02
	}
	if m.QB4 {
		octet1 |= 0x01
	}

	octet2 := uint8(0)
	if m.QB2 {
		octet2 |= 0x80
	}
	if m.QB1 {
		octet2 |= 0x40
	}
	if m.QC4 {
		octet2 |= 0x20
	}
	if m.QC2 {
		octet2 |= 0x10
	}
	if m.QC1 {
		octet2 |= 0x08
	}
	if m.QD4 {
		octet2 |= 0x04
	}
	if m.QD2 {
		octet2 |= 0x02
	}
	if m.QD1 {
		octet2 |= 0x01
	}

	buf.WriteByte(octet1)
	buf.WriteByte(octet2)
	return 2, nil
}

func (m *Mode3ACodeConfidence) String() string {
	return fmt.Sprintf("Mode-3/A Confidence Indicator")
}

func (m *Mode3ACodeConfidence) Validate() error {
	return nil
}
