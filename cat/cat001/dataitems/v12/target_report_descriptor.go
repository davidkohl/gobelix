package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetReportDescriptor represents I001/020 - Target Report Descriptor
// (CAT001 ed 1.4 §5.2.2). Variable length: first part plus one-octet extents.
//
// First part:   TYP(8) SIM(7) SSR/PSR(6-5) ANT(4) SPI(3) RAB(2) FX(1)
// First extent: TST(8) DS1/DS2(7-6) ME(5) MI(4) spare(3-2) FX(1)
//
// TYP selects the UAP used for the rest of the record (plot or track).
type TargetReportDescriptor struct {
	TYP uint8 // Report type (1 bit): 0 = plot, 1 = track
	SIM bool  // Simulated plot or track
	SSR bool  // Secondary detection in last scan (SSR/PSR = 10 or 11)
	PSR bool  // Primary detection in last scan (SSR/PSR = 01 or 11)
	ANT bool  // Target report from antenna 2
	SPI bool  // Special Position Identification
	RAB bool  // Plot or track from a fixed transponder

	// First extent
	TST    bool  // Test target indicator
	DS1DS2 uint8 // 0 default, 1 unlawful interference (7500), 2 radio-comm failure (7600), 3 emergency (7700)
	ME     bool  // Military emergency
	MI     bool  // Military identification
}

// Report types carried in TYP.
const (
	TYPPlot  uint8 = 0
	TYPTrack uint8 = 1
)

func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	*t = TargetReportDescriptor{}
	bytesRead := 0

	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: need at least 1 byte for target report descriptor", asterix.ErrBufferTooShort)
	}
	bytesRead++

	t.TYP = data >> 7
	t.SIM = data&0x40 != 0
	ssrpsr := (data >> 4) & 0x03
	t.PSR = ssrpsr == 1 || ssrpsr == 3
	t.SSR = ssrpsr == 2 || ssrpsr == 3
	t.ANT = data&0x08 != 0
	t.SPI = data&0x04 != 0
	t.RAB = data&0x02 != 0

	if data&0x01 == 0 {
		return bytesRead, nil
	}

	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
	}
	bytesRead++
	t.TST = data&0x80 != 0
	t.DS1DS2 = (data >> 5) & 0x03
	t.ME = data&0x10 != 0
	t.MI = data&0x08 != 0

	// Further extents are not defined in ed 1.4; skip them.
	for data&0x01 != 0 {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
		}
		bytesRead++
	}

	return bytesRead, nil
}

func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	octet1 := (t.TYP & 0x01) << 7
	if t.SIM {
		octet1 |= 0x40
	}
	var ssrpsr uint8
	if t.PSR {
		ssrpsr |= 0x01
	}
	if t.SSR {
		ssrpsr |= 0x02
	}
	octet1 |= ssrpsr << 4
	if t.ANT {
		octet1 |= 0x08
	}
	if t.SPI {
		octet1 |= 0x04
	}
	if t.RAB {
		octet1 |= 0x02
	}

	needExtent := t.TST || t.DS1DS2 != 0 || t.ME || t.MI
	if needExtent {
		octet1 |= 0x01
	}
	if err := buf.WriteByte(octet1); err != nil {
		return 0, fmt.Errorf("writing target report descriptor byte 1: %w", err)
	}
	if !needExtent {
		return 1, nil
	}

	var octet2 uint8
	if t.TST {
		octet2 |= 0x80
	}
	octet2 |= (t.DS1DS2 & 0x03) << 5
	if t.ME {
		octet2 |= 0x10
	}
	if t.MI {
		octet2 |= 0x08
	}
	if err := buf.WriteByte(octet2); err != nil {
		return 1, fmt.Errorf("writing target report descriptor byte 2: %w", err)
	}
	return 2, nil
}

func (t *TargetReportDescriptor) String() string {
	reportType := "Plot"
	if t.TYP == TYPTrack {
		reportType = "Track"
	}

	flags := ""
	if t.SIM {
		flags += " SIM"
	}
	if t.SSR {
		flags += " SSR"
	}
	if t.PSR {
		flags += " PSR"
	}
	if t.ANT {
		flags += " ANT2"
	}
	if t.SPI {
		flags += " SPI"
	}
	if t.RAB {
		flags += " RAB"
	}
	if t.TST {
		flags += " TST"
	}
	switch t.DS1DS2 {
	case 1:
		flags += " 7500"
	case 2:
		flags += " 7600"
	case 3:
		flags += " 7700"
	}
	if t.ME {
		flags += " ME"
	}
	if t.MI {
		flags += " MI"
	}

	return fmt.Sprintf("Type=%s%s", reportType, flags)
}

func (t *TargetReportDescriptor) Validate() error {
	if t.TYP > 1 {
		return fmt.Errorf("%w: I001/020 TYP is one bit, got %d", asterix.ErrInvalidMessage, t.TYP)
	}
	if t.DS1DS2 > 3 {
		return fmt.Errorf("%w: I001/020 DS1/DS2 is two bits, got %d", asterix.ErrInvalidMessage, t.DS1DS2)
	}
	return nil
}
