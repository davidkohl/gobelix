// cat/cat020/dataitems/v15/target_report_descriptor.go
package v15

import (
	"bytes"
	"fmt"
)

// TargetReportDescriptor implements I020/020 - Target Report Descriptor
type TargetReportDescriptor struct {
	SSR  bool // SSR detection
	MS   bool // Mode-S detection
	HF   bool // High Frequency Multilateration
	VDL4 bool // VDL Mode 4 Multilateration
	UAT  bool // UAT Multilateration
	DME  bool // DME
	OT   bool // Other Technology
	RAB  bool // Report from target transponder
	SPI  bool // Special Position Identification
	CHN  bool // Chain (1 or 2)
	GBS  bool // Transponder Ground bit set
	CRT  bool // Corrupted reply in multilateration
	SIM  bool // Simulated target
	TST  bool // Test target
	SAA  bool // Selected altitude available
	CL   uint8 // Confidence level (0-3)
	LLC  bool // List Lookup Check
	IPC  bool // Independent Position Check
	NOGO bool // NOGO bit status
	CPR  bool // Compact Position Reporting
	LDPJ bool // Local Decoding Position Jump
	RCF  bool // Range Check Failure
}

func (t *TargetReportDescriptor) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target report descriptor: %w", err)
	}
	bytesRead++

	t.SSR = (data & 0x80) != 0
	t.MS = (data & 0x40) != 0
	t.HF = (data & 0x20) != 0
	t.VDL4 = (data & 0x10) != 0
	t.UAT = (data & 0x08) != 0
	t.DME = (data & 0x04) != 0
	t.OT = (data & 0x02) != 0
	fx1 := (data & 0x01) != 0

	if !fx1 {
		return bytesRead, nil
	}

	// Second octet
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target report descriptor octet 2: %w", err)
	}
	bytesRead++

	t.RAB = (data & 0x80) != 0
	t.SPI = (data & 0x40) != 0
	t.CHN = (data & 0x20) != 0
	t.GBS = (data & 0x10) != 0
	t.CRT = (data & 0x08) != 0
	t.SIM = (data & 0x04) != 0
	t.TST = (data & 0x02) != 0
	fx2 := (data & 0x01) != 0

	if !fx2 {
		return bytesRead, nil
	}

	// Third octet
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target report descriptor octet 3: %w", err)
	}
	bytesRead++

	t.SAA = (data & 0x80) != 0
	t.CL = (data >> 5) & 0x03
	// Bit 4: spare
	t.LLC = (data & 0x08) != 0
	t.IPC = (data & 0x04) != 0
	t.NOGO = (data & 0x02) != 0
	fx3 := (data & 0x01) != 0

	if !fx3 {
		return bytesRead, nil
	}

	// Fourth octet
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading target report descriptor octet 4: %w", err)
	}
	bytesRead++

	t.CPR = (data & 0x80) != 0
	t.LDPJ = (data & 0x40) != 0
	t.RCF = (data & 0x20) != 0
	// Bits 5-2: spare
	fx4 := (data & 0x01) != 0

	// Read remaining extension octets if any
	for fx4 {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading target report descriptor extension: %w", err)
		}
		bytesRead++
		fx4 = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (t *TargetReportDescriptor) Encode(buf *bytes.Buffer) (int, error) {
	bytesWritten := 0

	// First octet
	var octet1 uint8
	if t.SSR {
		octet1 |= 0x80
	}
	if t.MS {
		octet1 |= 0x40
	}
	if t.HF {
		octet1 |= 0x20
	}
	if t.VDL4 {
		octet1 |= 0x10
	}
	if t.UAT {
		octet1 |= 0x08
	}
	if t.DME {
		octet1 |= 0x04
	}
	if t.OT {
		octet1 |= 0x02
	}

	// Check if we need more octets
	needOctet2 := t.RAB || t.SPI || t.CHN || t.GBS || t.CRT || t.SIM || t.TST
	needOctet3 := t.SAA || t.CL > 0 || t.LLC || t.IPC || t.NOGO
	needOctet4 := t.CPR || t.LDPJ || t.RCF

	if needOctet2 || needOctet3 || needOctet4 {
		octet1 |= 0x01 // FX
	}

	err := buf.WriteByte(octet1)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing target report descriptor: %w", err)
	}
	bytesWritten++

	if !needOctet2 && !needOctet3 && !needOctet4 {
		return bytesWritten, nil
	}

	// Second octet
	var octet2 uint8
	if t.RAB {
		octet2 |= 0x80
	}
	if t.SPI {
		octet2 |= 0x40
	}
	if t.CHN {
		octet2 |= 0x20
	}
	if t.GBS {
		octet2 |= 0x10
	}
	if t.CRT {
		octet2 |= 0x08
	}
	if t.SIM {
		octet2 |= 0x04
	}
	if t.TST {
		octet2 |= 0x02
	}

	if needOctet3 || needOctet4 {
		octet2 |= 0x01 // FX
	}

	err = buf.WriteByte(octet2)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing target report descriptor octet 2: %w", err)
	}
	bytesWritten++

	if !needOctet3 && !needOctet4 {
		return bytesWritten, nil
	}

	// Third octet
	var octet3 uint8
	if t.SAA {
		octet3 |= 0x80
	}
	octet3 |= (t.CL & 0x03) << 5
	if t.LLC {
		octet3 |= 0x08
	}
	if t.IPC {
		octet3 |= 0x04
	}
	if t.NOGO {
		octet3 |= 0x02
	}

	if needOctet4 {
		octet3 |= 0x01 // FX
	}

	err = buf.WriteByte(octet3)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing target report descriptor octet 3: %w", err)
	}
	bytesWritten++

	if !needOctet4 {
		return bytesWritten, nil
	}

	// Fourth octet
	var octet4 uint8
	if t.CPR {
		octet4 |= 0x80
	}
	if t.LDPJ {
		octet4 |= 0x40
	}
	if t.RCF {
		octet4 |= 0x20
	}
	// No more extensions, FX = 0

	err = buf.WriteByte(octet4)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing target report descriptor octet 4: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (t *TargetReportDescriptor) Validate() error {
	if t.CL > 3 {
		return fmt.Errorf("confidence level out of range: %d", t.CL)
	}
	return nil
}

func (t *TargetReportDescriptor) String() string {
	sources := []string{}
	if t.SSR {
		sources = append(sources, "SSR")
	}
	if t.MS {
		sources = append(sources, "Mode-S")
	}
	if t.HF {
		sources = append(sources, "HF-MLAT")
	}
	if t.VDL4 {
		sources = append(sources, "VDL4")
	}
	if t.UAT {
		sources = append(sources, "UAT")
	}
	if t.DME {
		sources = append(sources, "DME")
	}
	if t.OT {
		sources = append(sources, "Other")
	}

	flags := []string{}
	if t.SPI {
		flags = append(flags, "SPI")
	}
	if t.SIM {
		flags = append(flags, "SIM")
	}
	if t.TST {
		flags = append(flags, "TST")
	}

	result := "TRD:"
	if len(sources) > 0 {
		result += fmt.Sprintf(" %v", sources)
	}
	if len(flags) > 0 {
		result += fmt.Sprintf(" [%v]", flags)
	}
	return result
}
