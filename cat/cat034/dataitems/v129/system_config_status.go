// cat/cat034/dataitems/v129/system_config_status.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SystemConfigurationStatus represents I034/050 - System Configuration and
// Status (CAT034 ed 1.29 §5.2.6). Compound item: primary subfield
// COM(8) 0 0 PSR(5) SSR(4) MDS(3) 0 FX, then the present subfields in that
// order: COM 1 octet, PSR 1 octet, SSR 1 octet, MDS 2 octets.
//
// The subfields are kept as raw octets; Common, PSRStatus, SSRStatus and
// MDSStatus decode them into their bit fields.
type SystemConfigurationStatus struct {
	asterix.BaseCompoundItem
	COM uint8   // Common part (present unless OmitCOM)
	PSR *uint8  // PSR sensor status, optional
	SSR *uint8  // SSR sensor status, optional
	MDS *uint16 // Mode S sensor status, optional, two octets

	// OmitCOM is true when the COM subfield is absent. Its zero value keeps
	// the earlier behaviour of always encoding COM.
	OmitCOM bool
}

// SystemStatusCOM is the COM subfield of I034/050.
type SystemStatusCOM struct {
	NOGO   bool // bit 8: operational use of the system is inhibited
	RDPC   bool // bit 7: RDPC-2 selected (false: RDPC-1)
	RDPR   bool // bit 6: reset/restart of the selected RDPC
	OVLRDP bool // bit 5: radar data processor overload
	OVLXMT bool // bit 4: transmission subsystem overload
	MSC    bool // bit 3: monitoring system disconnected
	TSV    bool // bit 2: time source invalid
}

// SensorStatus is the PSR or SSR subfield of I034/050.
type SensorStatus struct {
	ANT  bool  // bit 8: antenna 2 selected (false: antenna 1)
	CHAB uint8 // bits 7-6: 0 none, 1 channel A, 2 channel B, 3 diversity (PSR) / invalid (SSR)
	OVL  bool  // bit 5: overload
	MSC  bool  // bit 4: monitoring system disconnected
}

// ModeSStatus is the MDS subfield of I034/050.
type ModeSStatus struct {
	ANT    bool  // bit 16: antenna 2 selected
	CHAB   uint8 // bits 15-14: surveillance channel, 0 none, 1 A, 2 B, 3 illegal
	OVLSUR bool  // bit 13: surveillance overload
	MSC    bool  // bit 12: monitoring system disconnected
	SCF    bool  // bit 11: SCF channel B in use (false: A)
	DLF    bool  // bit 10: DLF channel B in use (false: A)
	OVLSCF bool  // bit 9: overload in surveillance co-ordination function
	OVLDLF bool  // bit 8: overload in data link function
}

func bit(b uint8, mask uint8) bool { return b&mask != 0 }

func setBit(b *uint8, mask uint8, on bool) {
	if on {
		*b |= mask
	}
}

// ParseSystemStatusCOM decodes a COM octet.
func ParseSystemStatusCOM(b uint8) SystemStatusCOM {
	return SystemStatusCOM{
		NOGO: bit(b, 0x80), RDPC: bit(b, 0x40), RDPR: bit(b, 0x20), OVLRDP: bit(b, 0x10),
		OVLXMT: bit(b, 0x08), MSC: bit(b, 0x04), TSV: bit(b, 0x02),
	}
}

// Octet encodes the COM subfield.
func (c SystemStatusCOM) Octet() uint8 {
	var b uint8
	setBit(&b, 0x80, c.NOGO)
	setBit(&b, 0x40, c.RDPC)
	setBit(&b, 0x20, c.RDPR)
	setBit(&b, 0x10, c.OVLRDP)
	setBit(&b, 0x08, c.OVLXMT)
	setBit(&b, 0x04, c.MSC)
	setBit(&b, 0x02, c.TSV)
	return b
}

// ParseSensorStatus decodes a PSR or SSR status octet.
func ParseSensorStatus(b uint8) SensorStatus {
	return SensorStatus{ANT: bit(b, 0x80), CHAB: (b >> 5) & 0x03, OVL: bit(b, 0x10), MSC: bit(b, 0x08)}
}

// Octet encodes a PSR or SSR status subfield.
func (s SensorStatus) Octet() uint8 {
	b := (s.CHAB & 0x03) << 5
	setBit(&b, 0x80, s.ANT)
	setBit(&b, 0x10, s.OVL)
	setBit(&b, 0x08, s.MSC)
	return b
}

// ParseModeSStatus decodes the two MDS status octets.
func ParseModeSStatus(v uint16) ModeSStatus {
	hi, lo := uint8(v>>8), uint8(v)
	return ModeSStatus{
		ANT: bit(hi, 0x80), CHAB: (hi >> 5) & 0x03, OVLSUR: bit(hi, 0x10), MSC: bit(hi, 0x08),
		SCF: bit(hi, 0x04), DLF: bit(hi, 0x02), OVLSCF: bit(hi, 0x01), OVLDLF: bit(lo, 0x80),
	}
}

// Value encodes the MDS status subfield.
func (m ModeSStatus) Value() uint16 {
	hi := (m.CHAB & 0x03) << 5
	setBit(&hi, 0x80, m.ANT)
	setBit(&hi, 0x10, m.OVLSUR)
	setBit(&hi, 0x08, m.MSC)
	setBit(&hi, 0x04, m.SCF)
	setBit(&hi, 0x02, m.DLF)
	setBit(&hi, 0x01, m.OVLSCF)
	var lo uint8
	setBit(&lo, 0x80, m.OVLDLF)
	return uint16(hi)<<8 | uint16(lo)
}

// Common returns the COM subfield and whether it is present.
func (s *SystemConfigurationStatus) Common() (SystemStatusCOM, bool) {
	return ParseSystemStatusCOM(s.COM), !s.OmitCOM
}

// PSRStatus returns the PSR subfield and whether it is present.
func (s *SystemConfigurationStatus) PSRStatus() (SensorStatus, bool) {
	if s.PSR == nil {
		return SensorStatus{}, false
	}
	return ParseSensorStatus(*s.PSR), true
}

// SSRStatus returns the SSR subfield and whether it is present.
func (s *SystemConfigurationStatus) SSRStatus() (SensorStatus, bool) {
	if s.SSR == nil {
		return SensorStatus{}, false
	}
	return ParseSensorStatus(*s.SSR), true
}

// MDSStatus returns the MDS subfield and whether it is present.
func (s *SystemConfigurationStatus) MDSStatus() (ModeSStatus, bool) {
	if s.MDS == nil {
		return ModeSStatus{}, false
	}
	return ParseModeSStatus(*s.MDS), true
}

// NewSystemConfigurationStatus creates a new System Configuration and Status data item
func NewSystemConfigurationStatus() *SystemConfigurationStatus {
	return &SystemConfigurationStatus{
		BaseCompoundItem: asterix.BaseCompoundItem{},
	}
}

// Decode decodes the System Configuration and Status from bytes
func (s *SystemConfigurationStatus) Decode(buf *bytes.Buffer) (int, error) {
	const item = "I034/050"
	*s = SystemConfigurationStatus{}
	fspec, n, err := readPrimary(buf, item)
	if err != nil {
		return n, err
	}

	s.OmitCOM = fspec&primaryCOM == 0
	if !s.OmitCOM {
		if s.COM, err = readOctet(buf, item, "COM"); err != nil {
			return n, err
		}
		n++
	}
	if fspec&primaryPSR != 0 {
		b, err := readOctet(buf, item, "PSR")
		if err != nil {
			return n, err
		}
		n++
		s.PSR = &b
	}
	if fspec&primarySSR != 0 {
		b, err := readOctet(buf, item, "SSR")
		if err != nil {
			return n, err
		}
		n++
		s.SSR = &b
	}
	if fspec&primaryMDS != 0 {
		if buf.Len() < 2 {
			return n, fmt.Errorf("%w: reading %s MDS subfield: need 2 bytes, have %d", asterix.ErrBufferTooShort, item, buf.Len())
		}
		d := buf.Next(2)
		n += 2
		v := uint16(d[0])<<8 | uint16(d[1])
		s.MDS = &v
	}
	return n, nil
}

// Encode encodes the System Configuration and Status to bytes
func (s *SystemConfigurationStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	var fspec byte
	out := []byte{0}
	if !s.OmitCOM {
		fspec |= primaryCOM
		out = append(out, s.COM)
	}
	if s.PSR != nil {
		fspec |= primaryPSR
		out = append(out, *s.PSR)
	}
	if s.SSR != nil {
		fspec |= primarySSR
		out = append(out, *s.SSR)
	}
	if s.MDS != nil {
		fspec |= primaryMDS
		out = append(out, byte(*s.MDS>>8), byte(*s.MDS))
	}
	out[0] = fspec

	n, err := buf.Write(out)
	if err != nil {
		return n, fmt.Errorf("writing I034/050: %w", err)
	}
	return n, nil
}

// Validate validates the System Configuration and Status
func (s *SystemConfigurationStatus) Validate() error {
	// Every combination of subfields, including none, is encodable.
	return nil
}

// String returns a string representation
func (s *SystemConfigurationStatus) String() string {
	result := ""
	if !s.OmitCOM {
		result = fmt.Sprintf("COM: %+v", ParseSystemStatusCOM(s.COM))
	}
	if s.PSR != nil {
		result += fmt.Sprintf(" PSR: %+v", ParseSensorStatus(*s.PSR))
	}
	if s.SSR != nil {
		result += fmt.Sprintf(" SSR: %+v", ParseSensorStatus(*s.SSR))
	}
	if s.MDS != nil {
		result += fmt.Sprintf(" MDS: %+v", ParseModeSStatus(*s.MDS))
	}
	return result
}
