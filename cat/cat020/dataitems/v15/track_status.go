// cat/cat020/dataitems/v15/track_status.go
package v15

import (
	"bytes"
	"fmt"
)

// TrackStatus implements I020/170 - Track Status
// Per CAT020 spec v1.11 pages 25-26:
// Octet 1: CNF(8) TRE(7) CST(6) CDM(5-4) MAH(3) STH(2) FX(1)
// First extent (if FX=1): GHO(8) spare(7-2) FX(1)
type TrackStatus struct {
	CNF bool  // bit 8: 0=Confirmed, 1=Track in initiation phase (tentative)
	TRE bool  // bit 7: 0=Default, 1=Last report for track
	CST bool  // bit 6: 0=Not coasted, 1=Coasted
	CDM uint8 // bits 5-4: 00=Maintaining, 01=Climbing, 10=Descending, 11=Invalid
	MAH bool  // bit 3: 0=Default, 1=Horizontal manoeuvre
	STH bool  // bit 2: 0=Measured position, 1=Smoothed position
	GHO bool  // First extent bit 8: 0=Default, 1=Ghost track (fake target suspected)
}

func (t *TrackStatus) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0

	// First octet
	data, err := buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status: %w", err)
	}
	bytesRead++

	// Per spec: CNF(8) TRE(7) CST(6) CDM(5-4) MAH(3) STH(2) FX(1)
	t.CNF = (data & 0x80) != 0 // bit 8
	t.TRE = (data & 0x40) != 0 // bit 7
	t.CST = (data & 0x20) != 0 // bit 6 (single bit!)
	t.CDM = (data >> 3) & 0x03 // bits 5-4
	t.MAH = (data & 0x04) != 0 // bit 3
	t.STH = (data & 0x02) != 0 // bit 2
	fx := (data & 0x01) != 0   // bit 1

	if !fx {
		return bytesRead, nil
	}

	// First extent: GHO(8) spare(7-2) FX(1)
	data, err = buf.ReadByte()
	if err != nil {
		return bytesRead, fmt.Errorf("reading track status first extent: %w", err)
	}
	bytesRead++

	t.GHO = (data & 0x80) != 0 // bit 8
	// bits 7-2 are spare
	fx = (data & 0x01) != 0 // bit 1

	// Skip any further extension octets
	for fx {
		data, err = buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading track status extension: %w", err)
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

	// First octet: CNF(8) TRE(7) CST(6) CDM(5-4) MAH(3) STH(2) FX(1)
	var octet1 uint8
	if t.CNF {
		octet1 |= 0x80 // bit 8
	}
	if t.TRE {
		octet1 |= 0x40 // bit 7
	}
	if t.CST {
		octet1 |= 0x20 // bit 6
	}
	octet1 |= (t.CDM & 0x03) << 3 // bits 5-4
	if t.MAH {
		octet1 |= 0x04 // bit 3
	}
	if t.STH {
		octet1 |= 0x02 // bit 2
	}

	// Check if we need the extension
	needExtent := t.GHO
	if needExtent {
		octet1 |= 0x01 // FX
	}

	err := buf.WriteByte(octet1)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing track status: %w", err)
	}
	bytesWritten++

	if !needExtent {
		return bytesWritten, nil
	}

	// First extent: GHO(8) spare(7-2) FX(1)=0
	var ext1 uint8
	if t.GHO {
		ext1 |= 0x80 // bit 8
	}
	// FX = 0 (no more extensions)

	err = buf.WriteByte(ext1)
	if err != nil {
		return bytesWritten, fmt.Errorf("writing track status first extent: %w", err)
	}
	bytesWritten++

	return bytesWritten, nil
}

func (t *TrackStatus) Validate() error {
	if t.CDM > 3 {
		return fmt.Errorf("CDM out of range: %d", t.CDM)
	}
	return nil
}

func (t *TrackStatus) String() string {
	status := []string{}

	// CNF: 0=Confirmed, 1=Tentative (in initiation phase)
	if t.CNF {
		status = append(status, "TENTATIVE")
	} else {
		status = append(status, "CONFIRMED")
	}

	if t.TRE {
		status = append(status, "LAST-REPORT")
	}

	if t.CST {
		status = append(status, "COASTED")
	}

	switch t.CDM {
	case 0:
		// Maintaining - don't add
	case 1:
		status = append(status, "CLIMBING")
	case 2:
		status = append(status, "DESCENDING")
	case 3:
		status = append(status, "ALT-INVALID")
	}

	if t.MAH {
		status = append(status, "MANEUVER")
	}

	if t.STH {
		status = append(status, "SMOOTHED")
	}

	if t.GHO {
		status = append(status, "GHOST")
	}

	return fmt.Sprintf("Status: %v", status)
}
