// cat/cat020/dataitems/v10/target_report_descriptor.go
package v10

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// TargetReportDescriptor represents I020/020 - Target Report Descriptor
// Variable length: 1+ octets
// Type and characteristics of the data as transmitted by a system
type TargetReportDescriptor struct {
	// First octet
	SSR   bool // Non-Mode S 1090MHz multilateration
	MS    bool // Mode-S 1090 MHz multilateration
	HF    bool // HF multilateration
	VDL4  bool // VDL Mode 4 multilateration
	UAT   bool // UAT multilateration
	DME   bool // DME/TACAN multilateration

	// First extent (if present)
	RAB bool // Report from field monitor (fixed transponder)
	SPI bool // Special Position Identification
	CHN bool // Chain 2 (0=Chain 1, 1=Chain 2)
	GBS bool // Transponder Ground bit set
	CRT bool // Corrupted replies in multilateration
	SIM bool // Simulated target report
	TST bool // Test Target
}

// NewTargetReportDescriptor creates a new Target Report Descriptor data item
func NewTargetReportDescriptor() *TargetReportDescriptor {
	return &TargetReportDescriptor{}
}

// Decode decodes the Target Report Descriptor from bytes
func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: need at least 1 byte, have %d", asterix.ErrBufferTooShort, buf.Len())
	}

	bytesRead := 0
	data := buf.Next(1)
	bytesRead++

	// First octet: bits 8-2
	t.SSR = (data[0] & 0x80) != 0
	t.MS = (data[0] & 0x40) != 0
	t.HF = (data[0] & 0x20) != 0
	t.VDL4 = (data[0] & 0x10) != 0
	t.UAT = (data[0] & 0x08) != 0
	t.DME = (data[0] & 0x04) != 0
	// Bit 2 is spare (0)
	fx := (data[0] & 0x01) != 0

	// First extent (if FX bit is set)
	if fx {
		if buf.Len() < 1 {
			return bytesRead, fmt.Errorf("%w: need 1 more byte for extent", asterix.ErrBufferTooShort)
		}
		data = buf.Next(1)
		bytesRead++

		t.RAB = (data[0] & 0x80) != 0
		t.SPI = (data[0] & 0x40) != 0
		t.CHN = (data[0] & 0x20) != 0
		t.GBS = (data[0] & 0x10) != 0
		t.CRT = (data[0] & 0x08) != 0
		t.SIM = (data[0] & 0x04) != 0
		t.TST = (data[0] & 0x02) != 0
		// fx = (data[0] & 0x01) != 0 // Could extend further, but spec only defines 2 octets
	}

	return bytesRead, nil
}

// Encode encodes the Target Report Descriptor to bytes
func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Determine if we need the first extent
	needsExtent := t.RAB || t.SPI || t.CHN || t.GBS || t.CRT || t.SIM || t.TST

	// First octet
	var octet byte
	if t.SSR {
		octet |= 0x80
	}
	if t.MS {
		octet |= 0x40
	}
	if t.HF {
		octet |= 0x20
	}
	if t.VDL4 {
		octet |= 0x10
	}
	if t.UAT {
		octet |= 0x08
	}
	if t.DME {
		octet |= 0x04
	}
	// Bit 2 is spare (0)
	if needsExtent {
		octet |= 0x01 // FX bit
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing first octet: %w", err)
	}
	bytesWritten++

	// First extent (if needed)
	if needsExtent {
		octet = 0
		if t.RAB {
			octet |= 0x80
		}
		if t.SPI {
			octet |= 0x40
		}
		if t.CHN {
			octet |= 0x20
		}
		if t.GBS {
			octet |= 0x10
		}
		if t.CRT {
			octet |= 0x08
		}
		if t.SIM {
			octet |= 0x04
		}
		if t.TST {
			octet |= 0x02
		}
		// FX bit is 0 (no further extension)

		if err := buf.WriteByte(octet); err != nil {
			return bytesWritten, fmt.Errorf("writing first extent: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Target Report Descriptor
func (t *TargetReportDescriptor) Validate() error {
	// At least one multilateration type should be set
	if !t.SSR && !t.MS && !t.HF && !t.VDL4 && !t.UAT && !t.DME {
		return fmt.Errorf("%w: at least one multilateration type must be set", asterix.ErrInvalidMessage)
	}
	return nil
}

// String returns a string representation
func (t *TargetReportDescriptor) String() string {
	s := "TYP:"
	if t.SSR {
		s += " SSR"
	}
	if t.MS {
		s += " MS"
	}
	if t.HF {
		s += " HF"
	}
	if t.VDL4 {
		s += " VDL4"
	}
	if t.UAT {
		s += " UAT"
	}
	if t.DME {
		s += " DME"
	}
	if t.RAB {
		s += " RAB"
	}
	if t.SPI {
		s += " SPI"
	}
	if t.CHN {
		s += " CHN"
	}
	if t.GBS {
		s += " GBS"
	}
	if t.CRT {
		s += " CRT"
	}
	if t.SIM {
		s += " SIM"
	}
	if t.TST {
		s += " TST"
	}
	return s
}
