package v12

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetReportDescriptor represents I001/020 - Target Report Descriptor
// Extended item with TYP, SIM, SSR/PSR, ANT, SPI, RAB, TST flags
type TargetReportDescriptor struct {
	TYP uint8 // Report Type: 0=SSR multilateration, 1=SSR plot, 2=PSR plot, etc.
	SIM bool  // Simulated target report
	SSR bool  // SSR plot present
	PSR bool  // PSR plot present
	ANT bool  // Antenna number
	SPI bool  // Special Position Identification
	RAB bool  // Report from fixed transponder
	TST bool  // Test target
}

func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte for target report descriptor", asterix.ErrBufferTooShort)
	}

	data := buf.Next(1)
	bytesRead++

	// First octet: bits 8-6 = TYP, bit 5 = SIM, bit 4 = SSR/PSR, bit 3 = ANT, bit 2 = SPI, bit 1 = FX
	t.TYP = (data[0] >> 5) & 0x07
	t.SIM = (data[0] & 0x10) != 0
	t.SSR = (data[0] & 0x08) != 0
	t.ANT = (data[0] & 0x04) != 0
	t.SPI = (data[0] & 0x02) != 0

	// Check FX bit for extension
	hasFX := (data[0] & 0x01) != 0

	if hasFX {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
		}
		data = buf.Next(1)
		bytesRead++

		// Second octet: bit 8 = RAB, bit 7 = TST, bits 6-2 = spare, bit 1 = FX
		t.RAB = (data[0] & 0x80) != 0
		t.TST = (data[0] & 0x40) != 0

		// Handle additional extensions if present
		for (data[0] & 0x01) != 0 {
			if buf.Len() < 1 {
				return bytesRead, fmt.Errorf("%w: incomplete target report descriptor extension", asterix.ErrBufferTooShort)
			}
			data = buf.Next(1)
			bytesRead++
		}
	}

	return bytesRead, nil
}

func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	// First octet
	octet1 := (t.TYP & 0x07) << 5
	if t.SIM {
		octet1 |= 0x10
	}
	if t.SSR {
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

	buf.WriteByte(octet1)
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
		buf.WriteByte(octet2)
		bytesWritten++
	}

	return bytesWritten, nil
}

func (t *TargetReportDescriptor) String() string {
	reportType := "Unknown"
	switch t.TYP {
	case 0:
		reportType = "SSR multilateration"
	case 1:
		reportType = "SSR plot"
	case 2:
		reportType = "PSR plot"
	case 3:
		reportType = "Combined PSR+SSR plot"
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
