// dataitems/cat021/mops_version.go
// Package v26 implements the data items for ASTERIX Category 021 Version 2.6
package v26

import (
	"bytes"
	"fmt"
)

// MOPSVersion implements I021/210
//
// Field Reference Number (FRN): 18
//
// Description: This data item defines the Mode S Minimum Operational
// Performance Standards (MOPS) version used by the transponder and
// the type of link technology used for ADS-B reporting.
//
// Format: Fixed length Data Item of one octet
//
// The structure is:
// Bit 8: Reserved (0)
// Bit 7: VNS - Version Not Supported flag
// Bits 6-4: VN - Version Number
// Bits 3-1: LTT - Link Technology Type
//
// This data helps the ground system understand the capabilities and
// standards compliance level of the aircraft's transponder, which
// can affect how the data is processed and interpreted.
type MOPSVersion struct {
	VNS bool  // Version Not Supported
	VN  uint8 // Version Number
	LTT uint8 // Link Technology Type
}

// Version Numbers for 1090 ES (LTT=2)
const (
	VN_ED102_DO260   uint8 = 0 // ED102/DO-260
	VN_DO260A        uint8 = 1 // DO-260A
	VN_ED102A_DO260B uint8 = 2 // ED102A/DO-260B
	VN_ED102B_DO260C uint8 = 3 // ED-102B/DO-260C
)

// Link Technology Types
const (
	LTTOther  uint8 = iota // Other
	LTTVAT                 // UAT
	LTT1090ES              // 1090 Extended Squitter
	LTTVDL4                // VDL Mode 4
)

// Decode reads the MOPSVersion data from the buffer
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
	m.VN = (data[0] >> 3) & 0x07  // bits 6-4
	m.LTT = data[0] & 0x07        // bits 3-1

	return n, m.Validate()
}

// Encode writes the MOPSVersion data to the buffer
func (m *MOPSVersion) Encode(buf *bytes.Buffer) (int, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}

	// The LTT value needs to be in bits 3-1
	b := m.LTT & 0x07

	// No need to set VN bits if it's 0 - leave them as 0s
	if m.VN > 0 {
		// Set VN in bits 6-4
		b |= (m.VN & 0x07) << 3
	}

	// Set VNS bit if true
	if m.VNS {
		b |= 0x40
	}

	err := buf.WriteByte(b)
	if err != nil {
		return 0, fmt.Errorf("writing MOPS version: %w", err)
	}
	return 1, nil
}

// Validate checks if the MOPSVersion contains valid data
func (m *MOPSVersion) Validate() error {
	if m.VN > VN_ED102B_DO260C {
		return fmt.Errorf("invalid version number: %d", m.VN)
	}
	if m.LTT > LTTVDL4 {
		return fmt.Errorf("invalid link technology type: %d", m.LTT)
	}
	return nil
}

// String returns a human-readable representation of the MOPSVersion
func (m *MOPSVersion) String() string {
	var ver string
	switch m.VN {
	case VN_ED102_DO260:
		ver = "ED102/DO-260"
	case VN_DO260A:
		ver = "DO-260A"
	case VN_ED102A_DO260B:
		ver = "ED102A/DO-260B"
	case VN_ED102B_DO260C:
		ver = "ED102B/DO-260C"
	default:
		ver = fmt.Sprintf("Unknown(%d)", m.VN)
	}

	link := "Unknown"
	switch m.LTT {
	case LTTOther:
		link = "Other"
	case LTTVAT:
		link = "UAT"
	case LTT1090ES:
		link = "1090ES"
	case LTTVDL4:
		link = "VDL4"
	}

	if m.VNS {
		return fmt.Sprintf("%s[!]/%s", ver, link)
	}
	return fmt.Sprintf("%s/%s", ver, link)
}
