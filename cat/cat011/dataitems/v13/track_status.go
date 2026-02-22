package v13

import (
	"bytes"
	"fmt"
)

// TrackStatus implements I011/170 - Track Status
// Definition: Status of a track.
// Format: Variable length data item comprising a first part of one Octet,
// followed by 1-Octet extents as necessary.
type TrackStatus struct {
	// Primary part (octet 1)
	MON bool  // bit 8: 0=Multisensor track, 1=Monosensor track
	GBS bool  // bit 7: 0=Transponder Ground bit not set/unknown, 1=Ground bit set
	MRH bool  // bit 6: Most Reliable Height: 0=Barometric, 1=Geometric
	SRC uint8 // bits 5-3: Source of height (0-7)
	CNF bool  // bit 2: 0=Confirmed track, 1=Tentative track

	// First extension
	SIM    bool  // bit 8: 0=Actual track, 1=Simulated track
	TSE    bool  // bit 7: 0=Default, 1=Track service end
	TSB    bool  // bit 6: 0=Default, 1=Track service begin
	FRIFOE uint8 // bits 5-4: Friend/Foe identification
	ME     bool  // bit 3: 0=Default, 1=Military Emergency
	MI     bool  // bit 2: 0=Default, 1=Military Identification

	// Second extension
	AMA bool // bit 8: 0=Not from amalgamation, 1=From amalgamation
	SPI bool // bit 7: 0=Default, 1=SPI present
	CST bool // bit 6: 0=Default, 1=Coasting
	FPC bool // bit 5: 0=Not flight-plan correlated, 1=Flight plan correlated
	AFF bool // bit 4: 0=Default, 1=ADS-B data inconsistent

	// Third extension
	PSR bool // bit 7: 0=Default, 1=PSR age > threshold
	SSR bool // bit 6: 0=Default, 1=SSR age > threshold
	MDS bool // bit 5: 0=Default, 1=Mode S age > threshold
	ADS bool // bit 4: 0=Default, 1=ADS age > threshold
	SUC bool // bit 3: 0=Default, 1=Special Used Code
	AAC bool // bit 2: 0=Default, 1=Assigned Mode A Code Conflict

	extensionCount int // Track how many extensions were decoded
}

// Height source constants
const (
	HeightSourceNone          uint8 = 0 // No source
	HeightSourceGPS           uint8 = 1 // GPS
	HeightSource3DRadar       uint8 = 2 // 3D radar
	HeightSourceTriangulation uint8 = 3 // Triangulation
	HeightSourceCoverage      uint8 = 4 // Height from coverage
	HeightSourceLookupTable   uint8 = 5 // Speed look-up table
	HeightSourceDefault       uint8 = 6 // Default height
	HeightSourceMultilat      uint8 = 7 // Multilateration
)

// Friend/Foe constants
const (
	FriFoeNoMode4     uint8 = 0 // No Mode 4 interrogation
	FriFoeFriendly    uint8 = 1 // Friendly target
	FriFoeUnknown     uint8 = 2 // Unknown target
	FriFoeNoReply     uint8 = 3 // No reply
)

func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// Primary octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status: %w", err)
	}
	bytesRead++

	t.MON = (data & 0x80) != 0 // bit 8
	t.GBS = (data & 0x40) != 0 // bit 7
	t.MRH = (data & 0x20) != 0 // bit 6
	t.SRC = (data >> 2) & 0x07 // bits 5-3
	t.CNF = (data & 0x02) != 0 // bit 2
	fx := (data & 0x01) != 0   // bit 1

	if !fx {
		return bytesRead, nil
	}

	// First extension
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status ext1: %w", err)
	}
	bytesRead++
	t.extensionCount = 1

	t.SIM = (data & 0x80) != 0    // bit 8
	t.TSE = (data & 0x40) != 0    // bit 7
	t.TSB = (data & 0x20) != 0    // bit 6
	t.FRIFOE = (data >> 3) & 0x03 // bits 5-4
	t.ME = (data & 0x04) != 0     // bit 3
	t.MI = (data & 0x02) != 0     // bit 2
	fx = (data & 0x01) != 0       // bit 1

	if !fx {
		return bytesRead, nil
	}

	// Second extension
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status ext2: %w", err)
	}
	bytesRead++
	t.extensionCount = 2

	t.AMA = (data & 0x80) != 0 // bit 8
	t.SPI = (data & 0x40) != 0 // bit 7
	t.CST = (data & 0x20) != 0 // bit 6
	t.FPC = (data & 0x10) != 0 // bit 5
	t.AFF = (data & 0x08) != 0 // bit 4
	// bits 3-2 are spare
	fx = (data & 0x01) != 0 // bit 1

	if !fx {
		return bytesRead, nil
	}

	// Third extension
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status ext3: %w", err)
	}
	bytesRead++
	t.extensionCount = 3

	// bit 8 is spare
	t.PSR = (data & 0x40) != 0 // bit 7
	t.SSR = (data & 0x20) != 0 // bit 6
	t.MDS = (data & 0x10) != 0 // bit 5
	t.ADS = (data & 0x08) != 0 // bit 4
	t.SUC = (data & 0x04) != 0 // bit 3
	t.AAC = (data & 0x02) != 0 // bit 2
	fx = (data & 0x01) != 0    // bit 1

	// Skip any additional extensions we don't understand
	for fx {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading track status ext: %w", err)
		}
		bytesRead++
		fx = (data & 0x01) != 0
	}

	return bytesRead, nil
}

func (t *TrackStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Determine how many extensions we need
	needExt3 := t.PSR || t.SSR || t.MDS || t.ADS || t.SUC || t.AAC
	needExt2 := needExt3 || t.AMA || t.SPI || t.CST || t.FPC || t.AFF
	needExt1 := needExt2 || t.SIM || t.TSE || t.TSB || t.FRIFOE != 0 || t.ME || t.MI

	// Primary octet
	var octet uint8
	if t.MON {
		octet |= 0x80
	}
	if t.GBS {
		octet |= 0x40
	}
	if t.MRH {
		octet |= 0x20
	}
	octet |= (t.SRC & 0x07) << 2
	if t.CNF {
		octet |= 0x02
	}
	if needExt1 {
		octet |= 0x01 // FX
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing track status: %w", err)
	}
	bytesWritten++

	if !needExt1 {
		return bytesWritten, nil
	}

	// First extension
	octet = 0
	if t.SIM {
		octet |= 0x80
	}
	if t.TSE {
		octet |= 0x40
	}
	if t.TSB {
		octet |= 0x20
	}
	octet |= (t.FRIFOE & 0x03) << 3
	if t.ME {
		octet |= 0x04
	}
	if t.MI {
		octet |= 0x02
	}
	if needExt2 {
		octet |= 0x01 // FX
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing track status ext1: %w", err)
	}
	bytesWritten++

	if !needExt2 {
		return bytesWritten, nil
	}

	// Second extension
	octet = 0
	if t.AMA {
		octet |= 0x80
	}
	if t.SPI {
		octet |= 0x40
	}
	if t.CST {
		octet |= 0x20
	}
	if t.FPC {
		octet |= 0x10
	}
	if t.AFF {
		octet |= 0x08
	}
	// bits 3-2 are spare
	if needExt3 {
		octet |= 0x01 // FX
	}

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing track status ext2: %w", err)
	}
	bytesWritten++

	if !needExt3 {
		return bytesWritten, nil
	}

	// Third extension
	octet = 0
	// bit 8 is spare
	if t.PSR {
		octet |= 0x40
	}
	if t.SSR {
		octet |= 0x20
	}
	if t.MDS {
		octet |= 0x10
	}
	if t.ADS {
		octet |= 0x08
	}
	if t.SUC {
		octet |= 0x04
	}
	if t.AAC {
		octet |= 0x02
	}
	// FX = 0 (no more extensions)

	if err := buf.WriteByte(octet); err != nil {
		return bytesWritten, fmt.Errorf("writing track status ext3: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (t *TrackStatus) Validate() error {
	if t.SRC > 7 {
		return fmt.Errorf("SRC out of range: %d", t.SRC)
	}
	if t.FRIFOE > 3 {
		return fmt.Errorf("FRI/FOE out of range: %d", t.FRIFOE)
	}
	return nil
}

func (t *TrackStatus) String() string {
	status := []string{}

	if t.MON {
		status = append(status, "MONOSENSOR")
	} else {
		status = append(status, "MULTISENSOR")
	}

	if t.GBS {
		status = append(status, "GROUND")
	}

	if t.MRH {
		status = append(status, "GEO-HEIGHT")
	}

	srcStr := ""
	switch t.SRC {
	case HeightSourceNone:
		srcStr = "NO-SRC"
	case HeightSourceGPS:
		srcStr = "GPS"
	case HeightSource3DRadar:
		srcStr = "3D-RADAR"
	case HeightSourceTriangulation:
		srcStr = "TRIANGULATION"
	case HeightSourceCoverage:
		srcStr = "COVERAGE"
	case HeightSourceLookupTable:
		srcStr = "LOOKUP"
	case HeightSourceDefault:
		srcStr = "DEFAULT"
	case HeightSourceMultilat:
		srcStr = "MLAT"
	}
	status = append(status, srcStr)

	if t.CNF {
		status = append(status, "TENTATIVE")
	} else {
		status = append(status, "CONFIRMED")
	}

	if t.SIM {
		status = append(status, "SIMULATED")
	}
	if t.TSE {
		status = append(status, "TRACK-END")
	}
	if t.TSB {
		status = append(status, "TRACK-BEGIN")
	}
	if t.ME {
		status = append(status, "MIL-EMERGENCY")
	}
	if t.MI {
		status = append(status, "MIL-ID")
	}
	if t.AMA {
		status = append(status, "AMALGAMATED")
	}
	if t.SPI {
		status = append(status, "SPI")
	}
	if t.CST {
		status = append(status, "COASTED")
	}
	if t.FPC {
		status = append(status, "FP-CORRELATED")
	}
	if t.AFF {
		status = append(status, "ADS-INCONSISTENT")
	}
	if t.SUC {
		status = append(status, "SPECIAL-CODE")
	}
	if t.AAC {
		status = append(status, "MODE-A-CONFLICT")
	}

	return fmt.Sprintf("Track Status: %v", status)
}
