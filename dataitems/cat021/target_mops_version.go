// dataitems/cat021/mops_version.go
package cat021

import (
	"bytes"
	"fmt"
)

// MOPSVersion implements I021/210
type MOPSVersion struct {
	VNS bool  // Version Not Supported
	VN  uint8 // Version Number
	LTT uint8 // Link Technology Type
}

// Version Numbers for 1090 ES (LTT=2)
const (
	VN_ED102_DO260   uint8 = iota // ED102/DO-260
	VN_DO260A                     // DO-260A
	VN_ED102A_DO260B              // ED102A/DO-260B
	VN_ED102B_DO260C              // ED-102B/DO-260C
)

// Link Technology Types
const (
	LTTOther uint8 = iota
	LTTVAT
	LTT1090ES
	LTTVDL4
)

func (m *MOPSVersion) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading MOPS version: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for MOPS version: got %d bytes, want 1", n)
	}

	// Skip reserved bit 8
	m.VNS = (data[0] & 0x40) != 0 // bit 7
	m.VN = (data[0] >> 4) & 0x07  // bits 6-4
	m.LTT = data[0] & 0x07        // bits 3-1

	return n, m.Validate()
}

func (m *MOPSVersion) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	var b byte
	if m.VNS {
		b |= 0x40
	}
	b |= (m.VN & 0x07) << 4
	b |= m.LTT & 0x07

	err := buf.WriteByte(b)
	if err != nil {
		return 0, fmt.Errorf("writing MOPS version: %w", err)
	}
	return 1, nil
}

func (m *MOPSVersion) Validate() error {
	if m.VN > VN_ED102B_DO260C {
		return fmt.Errorf("invalid version number: %d", m.VN)
	}
	if m.LTT > LTTVDL4 {
		return fmt.Errorf("invalid link technology type: %d", m.LTT)
	}
	return nil
}
