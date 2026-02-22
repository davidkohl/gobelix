package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetReportDescriptor represents I001/020 - Target Report Descriptor
// Extended item with TYP, SIM, SSR/PSR, ANT, SPI, RAB, TST flags
type TargetReportDescriptor struct {
	TYP uint8 // Report Type (2 bits): 0=Plot, 1=Track end, 2-3=spare
	SIM bool  // Simulated target report
	SSR bool  // SSR plot present (derived from 2-bit SSR/PSR field)
	PSR bool  // PSR plot present (derived from 2-bit SSR/PSR field)
	ANT bool  // Antenna number
	SPI bool  // Special Position Identification
	RAB bool  // Report from fixed transponder
	TST bool  // Test target
}

func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: need at least 1 byte for target report descriptor", asterix.ErrBufferTooShort)
	}
	bytesRead++

	// First octet layout (EUROCONTROL ASTERIX Cat 001):
	//   Bit 8 (0x80): TYP high   \
	//   Bit 7 (0x40): TYP low     } TYP: 0=Plot, 1=Track end
	//   Bit 6 (0x20): SSR/PSR hi  \
	//   Bit 5 (0x10): SSR/PSR lo   } 0=none, 1=PSR, 2=SSR, 3=combined
	//   Bit 4 (0x08): SIM
	//   Bit 3 (0x04): ANT
	//   Bit 2 (0x02): SPI
	//   Bit 1 (0x01): FX
	t.TYP = (data >> 6) & 0x03
	ssrpsr := (data >> 4) & 0x03
	t.SSR = ssrpsr == 2 || ssrpsr == 3
	t.PSR = ssrpsr == 1 || ssrpsr == 3
	t.SIM = (data & 0x08) != 0
	t.ANT = (data & 0x04) != 0
	t.SPI = (data & 0x02) != 0

	// Check FX bit for extension
	hasFX := (data & 0x01) != 0

	if hasFX {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
		}
		bytesRead++

		// Second octet: bit 8 = RAB, bit 7 = TST, bits 6-2 = spare, bit 1 = FX
		t.RAB = (data & 0x80) != 0
		t.TST = (data & 0x40) != 0

		// Handle additional extensions if present
		for (data & 0x01) != 0 {
			data, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
			}
			bytesRead++
		}
	}

	return bytesRead, nil
}

func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	// First octet layout:
	//   Bits 8-7: TYP (2 bits)
	//   Bits 6-5: SSR/PSR (2 bits): 0=none, 1=PSR, 2=SSR, 3=combined
	//   Bit 4:    SIM
	//   Bit 3:    ANT
	//   Bit 2:    SPI
	//   Bit 1:    FX
	octet1 := (t.TYP & 0x03) << 6

	// Encode SSR/PSR as 2-bit field at bits 6-5
	var ssrpsr uint8
	if t.SSR && t.PSR {
		ssrpsr = 3 // combined
	} else if t.SSR {
		ssrpsr = 2 // SSR only
	} else if t.PSR {
		ssrpsr = 1 // PSR only
	}
	octet1 |= (ssrpsr & 0x03) << 4

	if t.SIM {
		octet1 |= 0x08
	}
	if t.ANT {
		octet1 |= 0x04
	}
	if t.SPI {
		octet1 |= 0x02
	}

	// Check if we need second octet
	needSecondOctet := t.RAB || t.TST
	if needSecondOctet {
		octet1 |= 0x01 // Set FX bit
	}

	if err := buf.WriteByte(octet1); err != nil {
		return 0, fmt.Errorf("writing target report descriptor byte 1: %w", err)
	}
	bytesWritten := 1

	if needSecondOctet {
		octet2 := uint8(0)
		if t.RAB {
			octet2 |= 0x80
		}
		if t.TST {
			octet2 |= 0x40
		}
		// No FX bit in second octet (no further extensions)
		if err := buf.WriteByte(octet2); err != nil {
			return 1, fmt.Errorf("writing target report descriptor byte 2: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (t *TargetReportDescriptor) String() string {
	reportType := "Unknown"
	switch t.TYP {
	case 0:
		reportType = "Plot"
	case 1:
		reportType = "Track end"
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
	if t.SPI {
		flags += " SPI"
	}
	if t.RAB {
		flags += " RAB"
	}
	if t.TST {
		flags += " TST"
	}

	return fmt.Sprintf("Type=%s%s", reportType, flags)
}

func (t *TargetReportDescriptor) Validate() error {
	return nil
}
