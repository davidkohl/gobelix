// dataitems/cat048/v117/target_report_descriptor.go
package v117

import (
	"bytes"
	"fmt"
	"strings"
)

// TargetReportDescriptor implements I048/020 for version 1.17
// Type and properties of the target report.
// Note: v1.17 has only primary part and first extension (no second-fifth extensions)
type TargetReportDescriptor struct {
	// Primary Part
	TYP uint8 // Type of detection (3 bits)
	SIM bool  // Actual/Simulated target
	RDP bool  // Report from RDP Chain 1/2
	SPI bool  // Absence/Presence of SPI
	RAB bool  // Report from aircraft/field monitor

	// First Extension (only extension in v1.17)
	TST bool  // Real/Test target report
	ME  bool  // No/Yes Military Emergency
	MI  bool  // No/Yes Military Identification
	FOE uint8 // FOE/FRI - Mode 4 interrogation info (2 bits)

	// Track which extensions are present
	extensions uint8
}

// Decode implements the DataItem interface
func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary Part (per ASTERIX CAT048 v1.17 spec)
	// Octet 1: TYP(bits8-6) SIM(bit5) RDP(bit4) SPI(bit3) RAB(bit2) FX(bit1)
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target report descriptor: %w", err)
	}
	bytesRead++

	t.TYP = (b >> 5) & 0x07 // bits 8-6
	t.SIM = (b & 0x10) != 0 // bit 5
	t.RDP = (b & 0x08) != 0 // bit 4
	t.SPI = (b & 0x04) != 0 // bit 3
	t.RAB = (b & 0x02) != 0 // bit 2
	fx := (b & 0x01) != 0   // bit 1 (FX)

	// First Extension (only extension in v1.17)
	// Octet 2: TST(bit8) spare(bits7-6) ME(bit5) MI(bit4) FOE/FRI(bits3-2) FX(bit1)
	if fx {
		t.extensions = 1
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading target report descriptor first extension: %w", err)
		}
		bytesRead++

		t.TST = (b & 0x80) != 0 // bit 8
		// bits 7-6 are spare in v1.17
		t.ME = (b & 0x10) != 0  // bit 5
		t.MI = (b & 0x08) != 0  // bit 4
		t.FOE = (b >> 1) & 0x03 // bits 3-2
		fx = (b & 0x01) != 0    // bit 1 (FX)

		// In v1.17, there are no second+ extensions defined
		// But handle FX bit gracefully by consuming any additional bytes
		for fx {
			t.extensions++
			b, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading target report descriptor extension %d: %w", t.extensions, err)
			}
			bytesRead++
			fx = (b & 0x01) != 0
		}
	}

	return bytesRead, nil
}

// Encode implements the DataItem interface
func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Primary Part
	b := byte(0)
	b |= (t.TYP & 0x07) << 5 // bits 8-6
	if t.SIM {
		b |= 0x10 // bit 5
	}
	if t.RDP {
		b |= 0x08 // bit 4
	}
	if t.SPI {
		b |= 0x04 // bit 3
	}
	if t.RAB {
		b |= 0x02 // bit 2
	}
	if t.extensions > 0 {
		b |= 0x01 // bit 1 (FX)
	}

	err := buf.WriteByte(b)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing target report descriptor: %w", err)
	}
	bytesWritten++

	// First Extension (only extension in v1.17)
	if t.extensions > 0 {
		b = 0
		if t.TST {
			b |= 0x80 // bit 8
		}
		// bits 7-6 are spare
		if t.ME {
			b |= 0x10 // bit 5
		}
		if t.MI {
			b |= 0x08 // bit 4
		}
		b |= (t.FOE & 0x03) << 1 // bits 3-2
		// FX bit (bit 1) is 0 - no second extension in v1.17

		err := buf.WriteByte(b)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing target report descriptor first extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (t *TargetReportDescriptor) Validate() error {
	if t.TYP > 7 {
		return fmt.Errorf("invalid TYP value: %d (must be 0-7)", t.TYP)
	}
	if t.FOE > 3 {
		return fmt.Errorf("invalid FOE value: %d (must be 0-3)", t.FOE)
	}
	return nil
}

// String returns a human-readable representation
func (t *TargetReportDescriptor) String() string {
	var parts []string

	// Detection type
	typStr := map[uint8]string{
		0: "No Detection",
		1: "PSR",
		2: "SSR",
		3: "SSR + PSR",
		4: "ModeS All-Call",
		5: "ModeS Roll-Call",
		6: "ModeS All-Call + PSR",
		7: "ModeS Roll-Call + PSR",
	}
	parts = append(parts, fmt.Sprintf("TYP: %s", typStr[t.TYP]))

	if t.SIM {
		parts = append(parts, "Simulated")
	}
	if t.RDP {
		parts = append(parts, "RDP Chain 2")
	}
	if t.SPI {
		parts = append(parts, "SPI")
	}
	if t.RAB {
		parts = append(parts, "Field Monitor")
	}

	// First extension fields
	if t.extensions > 0 {
		if t.TST {
			parts = append(parts, "Test Target")
		}
		if t.ME {
			parts = append(parts, "Military Emergency")
		}
		if t.MI {
			parts = append(parts, "Military ID")
		}

		foeStr := map[uint8]string{
			0: "No Mode 4",
			1: "Friendly",
			2: "Unknown",
			3: "No Reply",
		}
		parts = append(parts, fmt.Sprintf("FOE: %s", foeStr[t.FOE]))
	}

	return strings.Join(parts, ", ")
}

// SetExtensions sets the extensions count based on whether any extension fields are used
func (t *TargetReportDescriptor) SetExtensions() {
	if t.TST || t.ME || t.MI || t.FOE != 0 {
		t.extensions = 1
	} else {
		t.extensions = 0
	}
}
