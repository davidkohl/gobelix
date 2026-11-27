// cat/cat034/dataitems/v129/system_processing_mode.go
package v129

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// SystemProcessingMode represents I034/060 - System Processing Mode
// (CAT034 ed 1.29 §5.2.7). Compound item: primary subfield
// COM(8) 0 0 PSR(5) SSR(4) MDS(3) 0 FX, then the present one-octet
// subfields COM, PSR, SSR, MDS in that order.
//
// The subfields are kept as raw octets; Common, PSRMode, SSRMode and
// MDSMode decode them into their bit fields.
type SystemProcessingMode struct {
	asterix.BaseCompoundItem
	COM *uint8 // Common part, optional
	PSR *uint8 // PSR processing mode, optional
	SSR *uint8 // SSR processing mode, optional
	MDS *uint8 // Mode S processing mode, optional
}

// ProcessingCOM is the COM subfield of I034/060.
type ProcessingCOM struct {
	REDRDP uint8 // bits 7-5: reduction step in use for an RDP overload (0 = none)
	REDXMT uint8 // bits 4-2: reduction step in use for a transmission overload
}

// PSRProcessing is the PSR subfield of I034/060.
type PSRProcessing struct {
	POL    bool  // bit 8: circular polarisation (false: linear)
	REDRAD uint8 // bits 7-5: reduction step in use in the PSR subsystem
	STC    uint8 // bits 4-3: STC map in use, 0 = map 1 .. 3 = map 4
}

// SSRProcessing is the SSR subfield of I034/060.
type SSRProcessing struct {
	REDRAD uint8 // bits 8-6: reduction step in use in the SSR subsystem
}

// MDSProcessing is the MDS subfield of I034/060.
type MDSProcessing struct {
	REDRAD uint8 // bits 8-6: reduction step in use in the Mode S subsystem
	CLU    bool  // bit 5: cluster state not autonomous
}

// ParseProcessingCOM decodes a COM octet.
func ParseProcessingCOM(b uint8) ProcessingCOM {
	return ProcessingCOM{REDRDP: (b >> 4) & 0x07, REDXMT: (b >> 1) & 0x07}
}

// Octet encodes the COM subfield.
func (c ProcessingCOM) Octet() uint8 { return (c.REDRDP&0x07)<<4 | (c.REDXMT&0x07)<<1 }

// ParsePSRProcessing decodes a PSR octet.
func ParsePSRProcessing(b uint8) PSRProcessing {
	return PSRProcessing{POL: bit(b, 0x80), REDRAD: (b >> 4) & 0x07, STC: (b >> 2) & 0x03}
}

// Octet encodes the PSR subfield.
func (p PSRProcessing) Octet() uint8 {
	b := (p.REDRAD&0x07)<<4 | (p.STC&0x03)<<2
	setBit(&b, 0x80, p.POL)
	return b
}

// ParseSSRProcessing decodes an SSR octet.
func ParseSSRProcessing(b uint8) SSRProcessing { return SSRProcessing{REDRAD: b >> 5} }

// Octet encodes the SSR subfield.
func (p SSRProcessing) Octet() uint8 { return (p.REDRAD & 0x07) << 5 }

// ParseMDSProcessing decodes an MDS octet.
func ParseMDSProcessing(b uint8) MDSProcessing {
	return MDSProcessing{REDRAD: b >> 5, CLU: bit(b, 0x10)}
}

// Octet encodes the MDS subfield.
func (p MDSProcessing) Octet() uint8 {
	b := (p.REDRAD & 0x07) << 5
	setBit(&b, 0x10, p.CLU)
	return b
}

// Common returns the COM subfield and whether it is present.
func (s *SystemProcessingMode) Common() (ProcessingCOM, bool) {
	if s.COM == nil {
		return ProcessingCOM{}, false
	}
	return ParseProcessingCOM(*s.COM), true
}

// PSRMode returns the PSR subfield and whether it is present.
func (s *SystemProcessingMode) PSRMode() (PSRProcessing, bool) {
	if s.PSR == nil {
		return PSRProcessing{}, false
	}
	return ParsePSRProcessing(*s.PSR), true
}

// SSRMode returns the SSR subfield and whether it is present.
func (s *SystemProcessingMode) SSRMode() (SSRProcessing, bool) {
	if s.SSR == nil {
		return SSRProcessing{}, false
	}
	return ParseSSRProcessing(*s.SSR), true
}

// MDSMode returns the MDS subfield and whether it is present.
func (s *SystemProcessingMode) MDSMode() (MDSProcessing, bool) {
	if s.MDS == nil {
		return MDSProcessing{}, false
	}
	return ParseMDSProcessing(*s.MDS), true
}

// NewSystemProcessingMode creates a new System Processing Mode data item
func NewSystemProcessingMode() *SystemProcessingMode {
	return &SystemProcessingMode{
		BaseCompoundItem: asterix.BaseCompoundItem{},
	}
}

// Decode decodes the System Processing Mode from bytes
func (s *SystemProcessingMode) Decode(buf *bytes.Buffer) (int, error) {
	const item = "I034/060"
	*s = SystemProcessingMode{}
	fspec, n, err := readPrimary(buf, item)
	if err != nil {
		return n, err
	}
	for _, sf := range []struct {
		mask byte
		name string
		dst  **uint8
	}{
		{primaryCOM, "COM", &s.COM},
		{primaryPSR, "PSR", &s.PSR},
		{primarySSR, "SSR", &s.SSR},
		{primaryMDS, "MDS", &s.MDS},
	} {
		if fspec&sf.mask == 0 {
			continue
		}
		b, err := readOctet(buf, item, sf.name)
		if err != nil {
			return n, err
		}
		n++
		*sf.dst = &b
	}
	return n, nil
}

// Encode encodes the System Processing Mode to bytes
func (s *SystemProcessingMode) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	out := []byte{0}
	for _, sf := range []struct {
		mask byte
		v    *uint8
	}{
		{primaryCOM, s.COM}, {primaryPSR, s.PSR}, {primarySSR, s.SSR}, {primaryMDS, s.MDS},
	} {
		if sf.v != nil {
			out[0] |= sf.mask
			out = append(out, *sf.v)
		}
	}
	n, err := buf.Write(out)
	if err != nil {
		return n, fmt.Errorf("writing I034/060: %w", err)
	}
	return n, nil
}

// Validate validates the System Processing Mode
func (s *SystemProcessingMode) Validate() error {
	// Every combination of subfields, including none, is encodable.
	return nil
}

// String returns a string representation
func (s *SystemProcessingMode) String() string {
	result := "Processing Mode:"
	if c, ok := s.Common(); ok {
		result += fmt.Sprintf(" COM=%+v", c)
	}
	if p, ok := s.PSRMode(); ok {
		result += fmt.Sprintf(" PSR=%+v", p)
	}
	if p, ok := s.SSRMode(); ok {
		result += fmt.Sprintf(" SSR=%+v", p)
	}
	if p, ok := s.MDSMode(); ok {
		result += fmt.Sprintf(" MDS=%+v", p)
	}
	return result
}
