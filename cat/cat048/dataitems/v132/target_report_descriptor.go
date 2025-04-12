// dataitems/cat048/target_report_descriptor.go
package v132

import (
	"bytes"
	"fmt"
	"strings"
)

// TargetReportDescriptor implements I048/020
// Type and properties of the target report and target capabilities.
type TargetReportDescriptor struct {
	// Primary Part
	TYP uint8 // Type of detection
	SIM bool  // Actual/Simulated target
	RDP bool  // Report from RDP Chain 1/2
	SPI bool  // Absence/Presence of SPI
	RAB bool  // Report from aircraft/field monitor

	// First Extension
	TST bool  // Real/Test target report
	ERR bool  // No/Yes Extended Range
	XPP bool  // No/Yes X-Pulse
	ME  bool  // No/Yes Military Emergency
	MI  bool  // No/Yes Military Identification
	FOE uint8 // FOE/FRI - Mode 4 interrogation info

	// Second Extension
	ADSB_EP  bool // ADSB Element Populated
	ADSB_VAL bool // ADSB Information Available
	SCN_EP   bool // SCN Element Populated
	SCN_VAL  bool // SCN Information Available
	PAI_EP   bool // PAI Element Populated
	PAI_VAL  bool // PAI Information Available

	// Third Extension
	ACASXV_EP  bool  // ACASXV Element Populated
	ACASXV_VAL uint8 // ACAS Extended Version
	POXPR_EP   bool  // POXPR Element Populated
	POXPR_VAL  bool  // PO Transponder Capability

	// Fourth Extension
	POACT_EP   bool // POACT Element Populated
	POACT_VAL  bool // PO active for current plot
	DTFXPR_EP  bool // DTFXPR Element Populated
	DTFXPR_VAL bool // Basic Dataflash Transponder Capability
	DTFACT_EP  bool // DTFACT Element Populated
	DTFACT_VAL bool // Basic Dataflash active for current plot

	// Fifth Extension
	IRMXPR_EP  bool // IRMXPR Element Populated
	IRMXPR_VAL bool // Transponder IRM Capability
	IRMACT_EP  bool // IRMACT Element Populated
	IRMACT_VAL bool // IRM active for current plot

	// Track which extensions are present
	extensions uint8
}

// Decode implements the DataItem interface
func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary Part
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

	// First Extension
	if fx {
		t.extensions = 1
		b, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading target report descriptor first extension: %w", err)
		}
		bytesRead++

		t.TST = (b & 0x80) != 0 // bit 8
		t.ERR = (b & 0x40) != 0 // bit 7
		t.XPP = (b & 0x20) != 0 // bit 6
		t.ME = (b & 0x10) != 0  // bit 5
		t.MI = (b & 0x08) != 0  // bit 4
		t.FOE = (b >> 2) & 0x03 // bits 3-2
		fx = (b & 0x01) != 0    // bit 1 (FX)

		// Second Extension
		if fx {
			t.extensions = 2
			b, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("reading target report descriptor second extension: %w", err)
			}
			bytesRead++

			t.ADSB_EP = (b & 0x80) != 0  // bit 8
			t.ADSB_VAL = (b & 0x40) != 0 // bit 7
			t.SCN_EP = (b & 0x20) != 0   // bit 6
			t.SCN_VAL = (b & 0x10) != 0  // bit 5
			t.PAI_EP = (b & 0x08) != 0   // bit 4
			t.PAI_VAL = (b & 0x04) != 0  // bit 3
			// bit 2 is spare
			fx = (b & 0x01) != 0 // bit 1 (FX)

			// Third Extension
			if fx {
				t.extensions = 3
				b, err = buf.ReadByte()
				if err != nil {
					return bytesRead, fmt.Errorf("reading target report descriptor third extension: %w", err)
				}
				bytesRead++

				t.ACASXV_EP = (b & 0x80) != 0  // bit 8
				t.ACASXV_VAL = (b >> 4) & 0x0F // bits 7-4
				t.POXPR_EP = (b & 0x08) != 0   // bit 3
				t.POXPR_VAL = (b & 0x04) != 0  // bit 2
				fx = (b & 0x01) != 0           // bit 1 (FX)

				// Fourth Extension
				if fx {
					t.extensions = 4
					b, err = buf.ReadByte()
					if err != nil {
						return bytesRead, fmt.Errorf("reading target report descriptor fourth extension: %w", err)
					}
					bytesRead++

					t.POACT_EP = (b & 0x80) != 0   // bit 8
					t.POACT_VAL = (b & 0x40) != 0  // bit 7
					t.DTFXPR_EP = (b & 0x20) != 0  // bit 6
					t.DTFXPR_VAL = (b & 0x10) != 0 // bit 5
					t.DTFACT_EP = (b & 0x08) != 0  // bit 4
					t.DTFACT_VAL = (b & 0x04) != 0 // bit 3
					// bit 2 is spare
					fx = (b & 0x01) != 0 // bit 1 (FX)

					// Fifth Extension
					if fx {
						t.extensions = 5
						b, err = buf.ReadByte()
						if err != nil {
							return bytesRead, fmt.Errorf("reading target report descriptor fifth extension: %w", err)
						}
						bytesRead++

						t.IRMXPR_EP = (b & 0x80) != 0  // bit 8
						t.IRMXPR_VAL = (b & 0x40) != 0 // bit 7
						t.IRMACT_EP = (b & 0x20) != 0  // bit 6
						t.IRMACT_VAL = (b & 0x10) != 0 // bit 5
						// bits 4-2 are spare
						// No FX in this extension, ignore bit 1
					}
				}
			}
		}
	}

	return bytesRead, t.Validate()
}

// Encode implements the DataItem interface
func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Primary Part
	b := (t.TYP & 0x07) << 5 // bits 8-6
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

	// First Extension
	if t.extensions > 0 {
		b = 0
		if t.TST {
			b |= 0x80 // bit 8
		}
		if t.ERR {
			b |= 0x40 // bit 7
		}
		if t.XPP {
			b |= 0x20 // bit 6
		}
		if t.ME {
			b |= 0x10 // bit 5
		}
		if t.MI {
			b |= 0x08 // bit 4
		}
		b |= (t.FOE & 0x03) << 2 // bits 3-2
		if t.extensions > 1 {
			b |= 0x01 // bit 1 (FX)
		}

		err := buf.WriteByte(b)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing target report descriptor first extension: %w", err)
		}
		bytesWritten++

		// Second Extension
		if t.extensions > 1 {
			b = 0
			if t.ADSB_EP {
				b |= 0x80 // bit 8
			}
			if t.ADSB_VAL {
				b |= 0x40 // bit 7
			}
			if t.SCN_EP {
				b |= 0x20 // bit 6
			}
			if t.SCN_VAL {
				b |= 0x10 // bit 5
			}
			if t.PAI_EP {
				b |= 0x08 // bit 4
			}
			if t.PAI_VAL {
				b |= 0x04 // bit 3
			}
			// bit 2 is spare
			if t.extensions > 2 {
				b |= 0x01 // bit 1 (FX)
			}

			err := buf.WriteByte(b)
			if err != nil {
				return bytesWritten, fmt.Errorf("writing target report descriptor second extension: %w", err)
			}
			bytesWritten++

			// Third Extension
			if t.extensions > 2 {
				b = 0
				if t.ACASXV_EP {
					b |= 0x80 // bit 8
				}
				b |= (t.ACASXV_VAL & 0x0F) << 4 // bits 7-4
				if t.POXPR_EP {
					b |= 0x08 // bit 3
				}
				if t.POXPR_VAL {
					b |= 0x04 // bit 2
				}
				if t.extensions > 3 {
					b |= 0x01 // bit 1 (FX)
				}

				err := buf.WriteByte(b)
				if err != nil {
					return bytesWritten, fmt.Errorf("writing target report descriptor third extension: %w", err)
				}
				bytesWritten++

				// Fourth Extension
				if t.extensions > 3 {
					b = 0
					if t.POACT_EP {
						b |= 0x80 // bit 8
					}
					if t.POACT_VAL {
						b |= 0x40 // bit 7
					}
					if t.DTFXPR_EP {
						b |= 0x20 // bit 6
					}
					if t.DTFXPR_VAL {
						b |= 0x10 // bit 5
					}
					if t.DTFACT_EP {
						b |= 0x08 // bit 4
					}
					if t.DTFACT_VAL {
						b |= 0x04 // bit 3
					}
					// bit 2 is spare
					if t.extensions > 4 {
						b |= 0x01 // bit 1 (FX)
					}

					err := buf.WriteByte(b)
					if err != nil {
						return bytesWritten, fmt.Errorf("writing target report descriptor fourth extension: %w", err)
					}
					bytesWritten++

					// Fifth Extension
					if t.extensions > 4 {
						b = 0
						if t.IRMXPR_EP {
							b |= 0x80 // bit 8
						}
						if t.IRMXPR_VAL {
							b |= 0x40 // bit 7
						}
						if t.IRMACT_EP {
							b |= 0x20 // bit 6
						}
						if t.IRMACT_VAL {
							b |= 0x10 // bit 5
						}
						// bits 4-2 are spare
						// bit 1 is FX, but we don't set it in the last extension

						err := buf.WriteByte(b)
						if err != nil {
							return bytesWritten, fmt.Errorf("writing target report descriptor fifth extension: %w", err)
						}
						bytesWritten++
					}
				}
			}
		}
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (t *TargetReportDescriptor) Validate() error {
	if t.TYP > 7 {
		return fmt.Errorf("invalid TYP value: %d", t.TYP)
	}
	if t.FOE > 3 {
		return fmt.Errorf("invalid FOE value: %d", t.FOE)
	}
	if t.ACASXV_VAL > 15 {
		return fmt.Errorf("invalid ACASXV_VAL value: %d", t.ACASXV_VAL)
	}
	return nil
}

// String returns a human-readable representation
func (t *TargetReportDescriptor) String() string {
	var parts []string

	// Interpret TYP field
	typDesc := "Unknown"
	switch t.TYP {
	case 0:
		typDesc = "No detection"
	case 1:
		typDesc = "Single PSR detection"
	case 2:
		typDesc = "Single SSR detection"
	case 3:
		typDesc = "SSR + PSR detection"
	case 4:
		typDesc = "Single ModeS All-Call"
	case 5:
		typDesc = "Single ModeS Roll-Call"
	case 6:
		typDesc = "ModeS All-Call + PSR"
	case 7:
		typDesc = "ModeS Roll-Call + PSR"
	}
	parts = append(parts, fmt.Sprintf("TYP: %s", typDesc))

	// Add other significant fields
	if t.SIM {
		parts = append(parts, "Simulated")
	}
	if t.SPI {
		parts = append(parts, "SPI")
	}
	if t.RAB {
		parts = append(parts, "Field Monitor")
	}

	// Extensions
	if t.extensions > 0 {
		if t.TST {
			parts = append(parts, "Test")
		}
		if t.ERR {
			parts = append(parts, "Extended Range")
		}
		if t.XPP {
			parts = append(parts, "X-Pulse")
		}
		if t.ME {
			parts = append(parts, "Military Emergency")
		}
		if t.MI {
			parts = append(parts, "Military ID")
		}

		// FOE
		switch t.FOE {
		case 1:
			parts = append(parts, "Mode4: Friendly")
		case 2:
			parts = append(parts, "Mode4: Unknown")
		case 3:
			parts = append(parts, "Mode4: No Reply")
		}
	}

	return strings.Join(parts, ", ")
}

// SetExtensions sets the appropriate extension flag based on which fields are used
func (t *TargetReportDescriptor) SetExtensions() {
	// Check if fifth extension fields are set
	if t.IRMXPR_EP || t.IRMXPR_VAL || t.IRMACT_EP || t.IRMACT_VAL {
		t.extensions = 5
		return
	}

	// Check if fourth extension fields are set
	if t.POACT_EP || t.POACT_VAL || t.DTFXPR_EP || t.DTFXPR_VAL || t.DTFACT_EP || t.DTFACT_VAL {
		t.extensions = 4
		return
	}

	// Check if third extension fields are set
	if t.ACASXV_EP || t.ACASXV_VAL > 0 || t.POXPR_EP || t.POXPR_VAL {
		t.extensions = 3
		return
	}

	// Check if second extension fields are set
	if t.ADSB_EP || t.ADSB_VAL || t.SCN_EP || t.SCN_VAL || t.PAI_EP || t.PAI_VAL {
		t.extensions = 2
		return
	}

	// Check if first extension fields are set
	if t.TST || t.ERR || t.XPP || t.ME || t.MI || t.FOE > 0 {
		t.extensions = 1
		return
	}

	t.extensions = 0
}
