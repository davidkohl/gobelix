// dataitems/cat021/acas_resolution_advisory.go
package v26

import (
	"bytes"
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
)

// ACASResolutionAdvisory implements I021/260 (CAT021 v2.6 §5.2.39).
//
// The item is a verbatim seven-octet copy of the 1090 ES TCAS RA Broadcast
// (message type 28, subtype 2 — BDS Register 6,1 content):
//
//	bits 56-52 TYP  Message Type (28 for 1090 ES version 2+)
//	bits 51-49 STYP Message Sub-type (2 for 1090 ES version 2+)
//	bits 48-35 ARA  Active Resolution Advisories (14 bits)
//	bits 34-31 RAC  RA Complement Record (4 bits)
//	bit  30    RAT  RA Terminated
//	bit  29    MTE  Multiple Threat Encounter
//	bits 28-27 TTI  Threat Type Indicator (2 bits)
//	bits 26-1  TID  Threat Identity Data (26 bits)
type ACASResolutionAdvisory struct {
	TYP  uint8  // Message Type, 5 bits (28 for 1090 ES v2+)
	STYP uint8  // Message Sub-type, 3 bits (2 for 1090 ES v2+)
	ARA  uint16 // Active Resolution Advisories, 14 bits
	RAC  uint8  // RA Complement Record, 4 bits
	RAT  bool   // RA Terminated
	MTE  bool   // Multiple Threat Encounter
	TTI  uint8  // Threat Type Indicator, 2 bits
	TID  uint32 // Threat Identity Data, 26 bits
}

func (a *ACASResolutionAdvisory) Encode(buf *bytes.Buffer) (int, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}

	var data [7]byte
	data[0] = (a.TYP&0x1F)<<3 | (a.STYP & 0x07)
	data[1] = byte(a.ARA >> 6)
	data[2] = byte(a.ARA&0x3F)<<2 | (a.RAC>>2)&0x03
	data[3] = (a.RAC & 0x03) << 6
	if a.RAT {
		data[3] |= 0x20
	}
	if a.MTE {
		data[3] |= 0x10
	}
	data[3] |= (a.TTI & 0x03) << 2
	data[3] |= byte(a.TID>>24) & 0x03
	data[4] = byte(a.TID >> 16)
	data[5] = byte(a.TID >> 8)
	data[6] = byte(a.TID)

	n, err := buf.Write(data[:])
	if err != nil {
		return n, fmt.Errorf("writing ACAS resolution advisory: %w", err)
	}
	return n, nil
}

func (a *ACASResolutionAdvisory) Decode(buf *bytes.Buffer) (int, error) {
	var data [7]byte
	n, err := buf.Read(data[:])
	if err != nil {
		return n, fmt.Errorf("reading ACAS resolution advisory: %w", err)
	}
	if n != 7 {
		return n, fmt.Errorf("%w: need 7 bytes for ACAS resolution advisory, have %d", asterix.ErrBufferTooShort, n)
	}

	a.TYP = data[0] >> 3
	a.STYP = data[0] & 0x07
	a.ARA = uint16(data[1])<<6 | uint16(data[2])>>2
	a.RAC = (data[2]&0x03)<<2 | data[3]>>6
	a.RAT = data[3]&0x20 != 0
	a.MTE = data[3]&0x10 != 0
	a.TTI = (data[3] >> 2) & 0x03
	a.TID = uint32(data[3]&0x03)<<24 | uint32(data[4])<<16 | uint32(data[5])<<8 | uint32(data[6])

	return n, a.Validate()
}

func (a *ACASResolutionAdvisory) Validate() error {
	if a.TYP > 31 {
		return fmt.Errorf("message type exceeds 5-bit field: %d", a.TYP)
	}
	if a.ARA > 0x3FFF {
		return fmt.Errorf("ARA exceeds 14-bit field: %d", a.ARA)
	}
	if a.RAC > 15 {
		return fmt.Errorf("RAC exceeds 4-bit field: %d", a.RAC)
	}
	if a.TTI > 3 {
		return fmt.Errorf("TTI exceeds 2-bit field: %d", a.TTI)
	}
	if a.TID > 0x3FFFFFF {
		return fmt.Errorf("TID exceeds 26-bit field: %d", a.TID)
	}
	return nil
}

func (a *ACASResolutionAdvisory) String() string {
	s := fmt.Sprintf("ARA: %#04x, RAC: %#x, TTI: %d", a.ARA, a.RAC, a.TTI)
	if a.RAT {
		s += ", RAT"
	}
	if a.MTE {
		s += ", MTE"
	}
	if a.TTI == 1 {
		s += fmt.Sprintf(", threat ICAO %06X", a.TID>>2)
	}
	return s
}
