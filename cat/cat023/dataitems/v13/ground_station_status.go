// cat/cat023/dataitems/v13/ground_station_status.go
package v13

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// GroundStationStatus represents I023/100 - Ground Station Status
// This is an EXTENDED data item (1+ octets with FX bit for extension).
//
// Primary octet structure:
// | Bit 8 | Bit 7 | Bit 6 | Bit 5 | Bit 4 | Bit 3 | Bit 2 | Bit 1 |
// | NOGO  | ODP   | OXT   | MSC   | TSV   | SPO   | RN    | FX    |
//
// Extension octet (if FX=1):
// | Bit 8-2: GSSP (7 bits) | Bit 1: FX |
type GroundStationStatus struct {
	// Subfield #1: NOGO - Operational Release Status (bit 8)
	NOGO *NOGOStatus

	// Subfield #2: ODP - Data Processor Overload (bit 7)
	ODP *ODPStatus

	// Subfield #3: OXT - Ground Interface Data Communications Overload (bit 6)
	OXT *OXTStatus

	// Subfield #4: MSC - Monitoring System Connected Status (bit 5)
	MSC *MSCStatus

	// Subfield #5: TSV - Time Source Validity (bit 4)
	TSV *TSVStatus

	// Subfield #6: SPO - Spoofing Detected (bit 3)
	SPO *SPOStatus

	// Subfield #7: RN - Renumbering (bit 2)
	RN *RNStatus

	// Subfield #8: GSSP - Ground Station Status Reporting Period (extension octet)
	GSSP *GSSPValue
}

// NOGOStatus represents the NOGO subfield (1 bit)
type NOGOStatus struct {
	NOGO uint8 // 0=Data is released for operational use, 1=Data must not be used operationally
}

// ODPStatus represents the ODP subfield (1 bit)
type ODPStatus struct {
	ODP uint8 // 0=No overload, 1=Overload
}

// OXTStatus represents the OXT subfield (1 bit)
type OXTStatus struct {
	OXT uint8 // 0=No overload, 1=Overload
}

// MSCStatus represents the MSC subfield (1 bit)
type MSCStatus struct {
	MSC uint8 // 0=Monitoring system not connected or unknown, 1=Monitoring system connected
}

// TSVStatus represents the TSV subfield (1 bit)
type TSVStatus struct {
	TSV uint8 // 0=Valid, 1=Invalid
}

// SPOStatus represents the SPO subfield (1 bit)
type SPOStatus struct {
	SPO uint8 // 0=No spoofing detected, 1=Spoofing detected
}

// RNStatus represents the RN subfield (1 bit)
type RNStatus struct {
	RN uint8 // 0=Default, 1=Renumbering
}

// GSSPValue represents the GSSP subfield (7 bits, 1-127 seconds)
type GSSPValue struct {
	GSSP uint8 // Ground Station Status Reporting Period in seconds (LSB=1s, range 1-127)
}

// Decode decodes the Ground Station Status from bytes
// This is an EXTENDED data item (not compound!)
func (g *GroundStationStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Check we have at least 1 byte for the primary octet
	if buf.Len() < 1 {
		return 0, fmt.Errorf("%w: reading ground station status", asterix.ErrBufferTooShort)
	}

	// Peek at the first byte to check FX before consuming it
	primaryByte := buf.Bytes()[0]
	fx := primaryByte & 0x01 // Bit 1 (FX)

	// If FX is set, validate we have the extension byte available
	if fx != 0 && buf.Len() < 2 {
		return 0, fmt.Errorf("%w: reading ground station status extension", asterix.ErrBufferTooShort)
	}

	// Validation passed - now safely read the first octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("%w: reading ground station status", asterix.ErrBufferTooShort)
	}
	bytesRead++

	// Extract all fields from the first octet
	// Bits are numbered 8..1 from MSB to LSB
	// | Bit 8 | Bit 7 | Bit 6 | Bit 5 | Bit 4 | Bit 3 | Bit 2 | Bit 1 |
	// | NOGO  | ODP   | OXT   | MSC   | TSV   | SPO   | RN    | FX    |

	// Bit 8: NOGO
	nogo := (data >> 7) & 0x01
	g.NOGO = &NOGOStatus{NOGO: nogo}

	// Bit 7: ODP
	odp := (data >> 6) & 0x01
	g.ODP = &ODPStatus{ODP: odp}

	// Bit 6: OXT
	oxt := (data >> 5) & 0x01
	g.OXT = &OXTStatus{OXT: oxt}

	// Bit 5: MSC
	msc := (data >> 4) & 0x01
	g.MSC = &MSCStatus{MSC: msc}

	// Bit 4: TSV
	tsv := (data >> 3) & 0x01
	g.TSV = &TSVStatus{TSV: tsv}

	// Bit 3: SPO
	spo := (data >> 2) & 0x01
	g.SPO = &SPOStatus{SPO: spo}

	// Bit 2: RN
	rn := (data >> 1) & 0x01
	g.RN = &RNStatus{RN: rn}

	// Bit 1: FX (field extension) - already extracted from peek above

	// If FX bit is set, read extension octet(s)
	if fx != 0 {
		extData, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("%w: reading ground station status extension", asterix.ErrBufferTooShort)
		}
		bytesRead++

		// Bits 8-2: GSSP (Ground Station Status Reporting Period)
		gssp := (extData >> 1) & 0x7F
		g.GSSP = &GSSPValue{GSSP: gssp}

		// Bit 1: FX (for further extensions)
		fx = extData & 0x01

		// Consume any additional extension octets (spare in v1.3)
		for fx != 0 {
			extData, err = buf.ReadByte()
			if err != nil {
				return bytesRead, fmt.Errorf("%w: reading ground station status extension", asterix.ErrBufferTooShort)
			}
			bytesRead++
			fx = extData & 0x01
		}
	}

	return bytesRead, nil
}

// Encode encodes the Ground Station Status to bytes
// This is an EXTENDED data item - all status bits are packed into one byte.
func (g *GroundStationStatus) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// Build primary octet with all status bits packed:
	// | Bit 8 | Bit 7 | Bit 6 | Bit 5 | Bit 4 | Bit 3 | Bit 2 | Bit 1 |
	// | NOGO  | ODP   | OXT   | MSC   | TSV   | SPO   | RN    | FX    |
	var primaryByte byte = 0

	if g.NOGO != nil {
		primaryByte |= (g.NOGO.NOGO & 0x01) << 7
	}
	if g.ODP != nil {
		primaryByte |= (g.ODP.ODP & 0x01) << 6
	}
	if g.OXT != nil {
		primaryByte |= (g.OXT.OXT & 0x01) << 5
	}
	if g.MSC != nil {
		primaryByte |= (g.MSC.MSC & 0x01) << 4
	}
	if g.TSV != nil {
		primaryByte |= (g.TSV.TSV & 0x01) << 3
	}
	if g.SPO != nil {
		primaryByte |= (g.SPO.SPO & 0x01) << 2
	}
	if g.RN != nil {
		primaryByte |= (g.RN.RN & 0x01) << 1
	}

	// Set FX bit if GSSP is present
	if g.GSSP != nil {
		primaryByte |= 0x01 // FX = 1
	}

	// Write primary octet
	if err := buf.WriteByte(primaryByte); err != nil {
		return bytesWritten, fmt.Errorf("writing ground station status: %w", err)
	}
	bytesWritten++

	// Write extension octet if GSSP is present
	if g.GSSP != nil {
		// | Bits 8-2: GSSP (7 bits) | Bit 1: FX=0 |
		extByte := (g.GSSP.GSSP & 0x7F) << 1 // GSSP in bits 8-2, FX=0 in bit 1
		if err := buf.WriteByte(extByte); err != nil {
			return bytesWritten, fmt.Errorf("writing ground station status extension: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate validates the Ground Station Status
func (g *GroundStationStatus) Validate() error {
	if g.NOGO != nil && g.NOGO.NOGO > 1 {
		return fmt.Errorf("NOGO value out of range [0,1]: %d", g.NOGO.NOGO)
	}
	if g.ODP != nil && g.ODP.ODP > 1 {
		return fmt.Errorf("ODP value out of range [0,1]: %d", g.ODP.ODP)
	}
	if g.OXT != nil && g.OXT.OXT > 1 {
		return fmt.Errorf("OXT value out of range [0,1]: %d", g.OXT.OXT)
	}
	if g.MSC != nil && g.MSC.MSC > 1 {
		return fmt.Errorf("MSC value out of range [0,1]: %d", g.MSC.MSC)
	}
	if g.TSV != nil && g.TSV.TSV > 1 {
		return fmt.Errorf("TSV value out of range [0,1]: %d", g.TSV.TSV)
	}
	if g.SPO != nil && g.SPO.SPO > 1 {
		return fmt.Errorf("SPO value out of range [0,1]: %d", g.SPO.SPO)
	}
	if g.RN != nil && g.RN.RN > 1 {
		return fmt.Errorf("RN value out of range [0,1]: %d", g.RN.RN)
	}
	if g.GSSP != nil && (g.GSSP.GSSP < 1 || g.GSSP.GSSP > 127) {
		return fmt.Errorf("GSSP value out of range [1,127]: %d", g.GSSP.GSSP)
	}
	return nil
}

// String returns a string representation
func (g *GroundStationStatus) String() string {
	var parts []string

	if g.NOGO != nil {
		nogoDesc := map[uint8]string{
			0: "Released",
			1: "Not Released",
		}
		parts = append(parts, fmt.Sprintf("NOGO:%s", nogoDesc[g.NOGO.NOGO]))
	}
	if g.ODP != nil {
		if g.ODP.ODP == 1 {
			parts = append(parts, "ODP:Overload")
		} else {
			parts = append(parts, "ODP:Normal")
		}
	}
	if g.OXT != nil {
		if g.OXT.OXT == 1 {
			parts = append(parts, "OXT:Overload")
		} else {
			parts = append(parts, "OXT:Normal")
		}
	}
	if g.MSC != nil {
		if g.MSC.MSC == 1 {
			parts = append(parts, "MSC:Connected")
		} else {
			parts = append(parts, "MSC:Disconnected")
		}
	}
	if g.TSV != nil {
		if g.TSV.TSV == 1 {
			parts = append(parts, "TSV:Invalid")
		} else {
			parts = append(parts, "TSV:Valid")
		}
	}
	if g.SPO != nil {
		if g.SPO.SPO == 1 {
			parts = append(parts, "SPO:Detected")
		} else {
			parts = append(parts, "SPO:None")
		}
	}
	if g.RN != nil {
		if g.RN.RN == 1 {
			parts = append(parts, "RN:Renumbering")
		} else {
			parts = append(parts, "RN:Default")
		}
	}
	if g.GSSP != nil {
		parts = append(parts, fmt.Sprintf("GSSP:%ds", g.GSSP.GSSP))
	}

	if len(parts) == 0 {
		return "GroundStationStatus{}"
	}

	return fmt.Sprintf("GroundStationStatus{%s}", strings.Join(parts, ", "))
}
