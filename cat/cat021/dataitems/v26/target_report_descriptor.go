// dataitems/cat021/target_report_descriptor.go
package v26

import (
	"bytes"
	"fmt"
	"strings"
)

type TargetReportDescriptor struct {
	// Primary field
	ATP uint8 // Address Type 0-7
	ARC uint8 // Altitude Reporting Capability 0-3
	RC  bool  // Range Check
	RAB bool  // Report Type

	// First extension
	DCR bool  // Differential Correction
	GBS bool  // Ground Bit Setting
	SIM bool  // Simulated Target
	TST bool  // Test Target
	SAA bool  // Selected Altitude Available
	CL  uint8 // Confidence Level 0-3

	// Second extension
	IPC  bool // Independent Position Check
	NOGO bool // NOGO-bit set
	CPR  bool // Compact Position Reporting
	LDPJ bool // Local Decoding Position Jump
	RCF  bool // Range Check Failed

	// Third extension
	TYP  uint8 // Report Type 0-3
	STYP uint8 // Subtype 0-7
	ARA  bool  // Active Resolution Advisory
	SPI  bool  // Special Position Identification

	// Fourth extension
	TBC uint8 // Total Bits Corrected 0-63
	MBC uint8 // Maximum Bits Corrected 0-63

	hasExtensions uint8 // Tracks which extensions are present (0-4)
}

func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// First byte
	b := (t.ATP << 5) | (t.ARC << 3)
	if t.RC {
		b |= 0x04
	}
	if t.RAB {
		b |= 0x02
	}
	if t.hasExtensions > 0 {
		b |= 0x01 // FX bit
	}

	if err := buf.WriteByte(b); err != nil {
		return bytesWritten, fmt.Errorf("writing primary field: %w", err)
	}
	bytesWritten++

	// First extension if present
	if t.hasExtensions > 0 {
		b = 0
		if t.DCR {
			b |= 0x80
		}
		if t.GBS {
			b |= 0x40
		}
		if t.SIM {
			b |= 0x20
		}
		if t.TST {
			b |= 0x10
		}
		if t.SAA {
			b |= 0x08
		}
		b |= (t.CL & 0x03) << 1

		if t.hasExtensions > 1 {
			b |= 0x01 // FX bit
		}

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing first extension: %w", err)
		}
		bytesWritten++
	}

	// Second extension if present
	if t.hasExtensions > 1 {
		b = 0
		if t.IPC {
			b |= 0x80
		}
		if t.NOGO {
			b |= 0x40
		}
		if t.CPR {
			b |= 0x20
		}
		if t.LDPJ {
			b |= 0x10
		}
		if t.RCF {
			b |= 0x08
		}

		if t.hasExtensions > 2 {
			b |= 0x01 // FX bit
		}

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing second extension: %w", err)
		}
		bytesWritten++
	}

	// Third extension if present
	if t.hasExtensions > 2 {
		b = 0
		b |= (t.TYP & 0x03) << 6
		b |= (t.STYP & 0x07) << 3
		if t.ARA {
			b |= 0x04
		}
		if t.SPI {
			b |= 0x02
		}

		if t.hasExtensions > 3 {
			b |= 0x01 // FX bit
		}

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing third extension: %w", err)
		}
		bytesWritten++
	}

	// Fourth extension if present
	if t.hasExtensions > 3 {
		b = 0
		b |= (t.TBC & 0x3F) << 2
		b |= (t.MBC & 0x3F)

		if err := buf.WriteByte(b); err != nil {
			return bytesWritten, fmt.Errorf("writing fourth extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First byte
	b, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading primary field: %w", err)
	}
	bytesRead++

	t.ATP = (b >> 5) & 0x07
	t.ARC = (b >> 3) & 0x03
	t.RC = (b & 0x04) != 0
	t.RAB = (b & 0x02) != 0
	fx := (b & 0x01) != 0

	// First extension
	if fx {
		t.hasExtensions = 1
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading first extension: %w", err)
		}
		bytesRead++

		t.DCR = (b & 0x80) != 0
		t.GBS = (b & 0x40) != 0
		t.SIM = (b & 0x20) != 0
		t.TST = (b & 0x10) != 0
		t.SAA = (b & 0x08) != 0
		t.CL = (b >> 1) & 0x03
		fx = (b & 0x01) != 0

		// Second extension
		if fx {
			t.hasExtensions = 2
			b, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading second extension: %w", err)
			}
			bytesRead++

			t.IPC = (b & 0x80) != 0
			t.NOGO = (b & 0x40) != 0
			t.CPR = (b & 0x20) != 0
			t.LDPJ = (b & 0x10) != 0
			t.RCF = (b & 0x08) != 0
			fx = (b & 0x01) != 0

			// Third extension
			if fx {
				t.hasExtensions = 3
				b, err = buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading third extension: %w", err)
				}
				bytesRead++

				t.TYP = (b >> 6) & 0x03
				t.STYP = (b >> 3) & 0x07
				t.ARA = (b & 0x04) != 0
				t.SPI = (b & 0x02) != 0
				fx = (b & 0x01) != 0

				// Fourth extension
				if fx {
					t.hasExtensions = 4
					b, err = buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading fourth extension: %w", err)
					}
					bytesRead++

					t.TBC = (b >> 2) & 0x3F
					t.MBC = b & 0x3F
				}
			}
		}
	}

	return bytesRead, t.Validate()
}

func (t *TargetReportDescriptor) Validate() error {
	if t.ATP > 7 {
		return fmt.Errorf("invalid ATP value: %d", t.ATP)
	}
	if t.ARC > 3 {
		return fmt.Errorf("invalid ARC value: %d", t.ARC)
	}
	if t.CL > 3 {
		return fmt.Errorf("invalid confidence level: %d", t.CL)
	}
	if t.TYP > 3 {
		return fmt.Errorf("invalid report type: %d", t.TYP)
	}
	if t.STYP > 7 {
		return fmt.Errorf("invalid subtype: %d", t.STYP)
	}
	if t.TBC > 63 {
		return fmt.Errorf("invalid total bits corrected: %d", t.TBC)
	}
	if t.MBC > 63 {
		return fmt.Errorf("invalid maximum bits corrected: %d", t.MBC)
	}
	return nil
}

func (t *TargetReportDescriptor) String() string {
	var details []string

	// Add ATP (Address Type)
	atpDesc := "Unknown"
	switch t.ATP {
	case 0:
		atpDesc = "24-Bit ICAO address"
	case 1:
		atpDesc = "Duplicate address"
	case 2:
		atpDesc = "Surface vehicle address"
	case 3:
		atpDesc = "Anonymous address"
	}
	details = append(details, fmt.Sprintf("ATP: %s", atpDesc))

	// Add ARC (Altitude Reporting Capability)
	arcDesc := "Unknown"
	switch t.ARC {
	case 0:
		arcDesc = "25ft"
	case 1:
		arcDesc = "100ft"
	case 2:
		arcDesc = "Unknown"
	case 3:
		arcDesc = "Invalid"
	}
	details = append(details, fmt.Sprintf("ARC: %s", arcDesc))

	// Add RC and RAB
	if t.RC {
		details = append(details, "RC: Range Check Passed")
	}
	if t.RAB {
		details = append(details, "RAB: Field Monitor")
	}

	// Include important status indicators from extensions
	if t.DCR {
		details = append(details, "DCR: Differential Correction")
	}
	if t.GBS {
		details = append(details, "GBS: Ground Bit Set")
	}
	if t.SIM {
		details = append(details, "SIM: Simulated Target")
	}
	if t.TST {
		details = append(details, "TST: Test Target")
	}

	return strings.Join(details, ", ")
}
